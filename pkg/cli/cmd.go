package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/nghialv/lotus/pkg/log"
	"github.com/nghialv/lotus/pkg/version"
)

func WithContext(runner func(context.Context, *zap.Logger) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return runWithContext(cmd, runner)
	}
}

func runWithContext(cmd *cobra.Command, runner func(context.Context, *zap.Logger) error) error {
	logger, err := newLogger(cmd)
	if err != nil {
		return err
	}
	defer logger.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(ch)

	go func() {
		select {
		case s := <-ch:
			logger.Info("stopping due to signal", zap.Stringer("signal", s))
			cancel()
		case <-ctx.Done():
		}
	}()

	logger.Info(fmt.Sprintf("running %s", cmd.CommandPath()))
	return runner(ctx, logger)
}

func newLogger(cmd *cobra.Command) (*zap.Logger, error) {
	configs := log.DefaultConfigs
	configs.ServiceContext = &log.ServiceContext{
		Service: strings.Replace(cmd.CommandPath(), " ", ".", -1),
		Version: version.Get().Version,
	}
	if f := cmd.Flag("log-level"); f != nil {
		configs.Level = f.Value.String()
	}
	if f := cmd.Flag("log-encoding"); f != nil {
		configs.Encoding = f.Value.String()
	}
	return log.NewLogger(configs)
}
