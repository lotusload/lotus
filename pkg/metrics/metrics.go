// Copyright (c) 2018 Lotus Load
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"contrib.go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/stats/view"

	"github.com/lotusload/lotus/pkg/metrics/grpcmetrics"
	"github.com/lotusload/lotus/pkg/metrics/httpmetrics"
	"github.com/lotusload/lotus/pkg/virtualuser"
)

type options struct {
	namespace        string
	path             string
	reportingPeriod  time.Duration
	gracefulPeriod   time.Duration
	grpcViews        []*view.View
	httpViews        []*view.View
	virtualUserViews []*view.View
	customViews      []*view.View
	logger           Logger
}

var defaultOptions = options{
	namespace:       "lotus",
	path:            "/metrics",
	reportingPeriod: time.Second,
	gracefulPeriod:  5 * time.Second,
	grpcViews:       grpcmetrics.DefaultClientViews,
	httpViews:       httpmetrics.DefaultClientViews,
	virtualUserViews: []*view.View{
		virtualuser.UserCountView,
	},
	logger: logger{},
}

type Option func(*options)

func WithNamespace(namespace string) Option {
	return func(opts *options) {
		opts.namespace = namespace
	}
}

func WithPath(path string) Option {
	return func(opts *options) {
		opts.path = path
	}
}

func WithReportingPeriod(period time.Duration) Option {
	return func(opts *options) {
		opts.reportingPeriod = period
	}
}

func WithGracefulPeriod(period time.Duration) Option {
	return func(opts *options) {
		opts.gracefulPeriod = period
	}
}

func WithCustomViews(views ...*view.View) Option {
	return func(opts *options) {
		opts.customViews = views
	}
}

func WithGrpcViews(views ...*view.View) Option {
	return func(opts *options) {
		opts.grpcViews = views
	}
}

func WithHttpViews(views ...*view.View) Option {
	return func(opts *options) {
		opts.httpViews = views
	}
}

func WithLogger(logger Logger) Option {
	return func(opts *options) {
		opts.logger = logger
	}
}

func (o options) Views() []*view.View {
	length := len(o.grpcViews) + len(o.httpViews) + len(o.virtualUserViews) + len(o.customViews)
	views := make([]*view.View, 0, length)
	views = append(views, o.grpcViews...)
	views = append(views, o.httpViews...)
	views = append(views, o.virtualUserViews...)
	views = append(views, o.customViews...)
	return views
}

type server struct {
	server *http.Server
	opts   options
}

func NewServer(port int, opt ...Option) (*server, error) {
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}
	view.SetReportingPeriod(opts.reportingPeriod)
	err := view.Register(opts.Views()...)
	if err != nil {
		opts.logger.Errorf("failed to register views: %v", err)
		return nil, err
	}
	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace: opts.namespace,
	})
	if err != nil {
		opts.logger.Errorf("failed to create prometheus exporter: %v", err)
		return nil, err
	}
	view.RegisterExporter(pe)

	mux := http.NewServeMux()
	mux.Handle(opts.path, pe)
	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
	return &server{
		server: s,
		opts:   opts,
	}, nil
}

func (s *server) Run() error {
	err := s.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		s.opts.logger.Errorf("failed to run metrics server: %v", err)
		return err
	}
	return nil
}

func (s *server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.opts.gracefulPeriod)
	defer cancel()
	err := s.server.Shutdown(ctx)
	if err != nil {
		s.opts.logger.Errorf("failed to shutdown metrics server: %v", err)
	}
	return err
}
