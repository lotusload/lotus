package datasource

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/nghialv/lotus/pkg/app/lotus/config"
	"github.com/nghialv/lotus/pkg/app/lotus/model"
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
