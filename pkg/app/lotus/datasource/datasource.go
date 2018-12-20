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

package datasource

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/lotusload/lotus/pkg/app/lotus/config"
	"github.com/lotusload/lotus/pkg/app/lotus/model"
)

type Builder interface {
	Build(ds *config.DataSource, opts BuildOptions) (DataSource, error)
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

type DataSource interface {
	Querier
	Checker
}

type Querier interface {
	Query(ctx context.Context, query string, ts time.Time) ([]*Sample, error)
	CollectSummary(ctx context.Context, ts time.Time) (*model.MetricsSummary, error)
}

type Sample struct {
	Labels    map[string]string
	Value     float64
	Timestamp time.Time
}

type Checker interface {
	Check(ctx context.Context, checks []Check) (*CheckResult, error)
}

type Check struct {
	Name string
	Expr string
	For  string
}

type CheckResult struct {
	Actives []string
}
