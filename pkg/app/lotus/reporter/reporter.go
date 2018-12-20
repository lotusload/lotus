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

package reporter

import (
	"context"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/lotusload/lotus/pkg/app/lotus/config"
	"github.com/lotusload/lotus/pkg/app/lotus/model"
)

type Builder interface {
	Build(r *config.Receiver, opts BuildOptions) (Reporter, error)
}

type BuildOptions struct {
	Logger *zap.Logger
}

func (o BuildOptions) NamedLogger(name string) *zap.Logger {
	if o.Logger != nil {
		return o.Logger.Named(name)
	}
	return zap.NewNop().Named(name)
}

type Reporter interface {
	Report(ctx context.Context, result *model.Result) error
}

type multiReporter struct {
	reporters []Reporter
}

func MultiReporter(reporters ...Reporter) Reporter {
	all := make([]Reporter, 0, len(reporters))
	for _, r := range reporters {
		if mr, ok := r.(*multiReporter); ok {
			all = append(all, mr.reporters...)
		} else {
			all = append(all, r)
		}
	}
	return &multiReporter{
		reporters: all,
	}
}

func (mr *multiReporter) Report(ctx context.Context, result *model.Result) error {
	if len(mr.reporters) == 0 {
		return nil
	}
	if len(mr.reporters) == 1 {
		return mr.reporters[0].Report(ctx, result)
	}
	g, ctx := errgroup.WithContext(ctx)
	for i := range mr.reporters {
		reporter := mr.reporters[i]
		g.Go(func() error {
			return reporter.Report(ctx, result)
		})
	}
	return g.Wait()
}
