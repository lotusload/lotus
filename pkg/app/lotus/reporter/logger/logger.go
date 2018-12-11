package logger

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/nghialv/lotus/pkg/app/lotus/config"
	"github.com/nghialv/lotus/pkg/app/lotus/model"
	"github.com/nghialv/lotus/pkg/app/lotus/reporter"
)

type builder struct {
}

func NewBuilder() reporter.Builder {
	return &builder{}
}

func (b *builder) Build(r *config.Receiver, opts reporter.BuildOptions) (reporter.Reporter, error) {
	return &logger{
		logger: opts.NamedLogger("logger-reporter"),
	}, nil
}

type logger struct {
	logger *zap.Logger
}

func (p *logger) Report(ctx context.Context, result *model.Result) error {
	out, err := result.Render(model.RenderFormatText)
	if err != nil {
		return err
	}
	fmt.Println("=========== TEST RESULT ==========")
	fmt.Println(string(out))
	fmt.Println("===========     END     ==========")
	return nil
}
