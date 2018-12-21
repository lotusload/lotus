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

	"github.com/lotusload/lotus/pkg/log"
	"github.com/lotusload/lotus/pkg/version"
)

func WithContext(runner func(context.Context, *zap.Logger) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(ch)
		return runWithContext(cmd, ch, runner)
	}
}

func runWithContext(cmd *cobra.Command, signalCh <-chan os.Signal, runner func(context.Context, *zap.Logger) error) error {
	logger, err := newLogger(cmd)
	if err != nil {
		return err
	}
	defer logger.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		select {
		case s := <-signalCh:
			logger.Info("stopping due to signal", zap.Any("signal", s))
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
