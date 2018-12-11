package reporter

import (
	"context"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/nghialv/lotus/pkg/app/lotus/config"
	"github.com/nghialv/lotus/pkg/app/lotus/model"
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
