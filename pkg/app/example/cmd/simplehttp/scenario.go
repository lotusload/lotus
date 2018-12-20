package simplehttp

import (
	"context"
	"net/http"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/net/context/ctxhttp"

	"github.com/lotusload/lotus/pkg/cli"
	"github.com/lotusload/lotus/pkg/metrics"
	"github.com/lotusload/lotus/pkg/metrics/httpmetrics"
)

type scenario struct {
}

func NewCommand() *cobra.Command {
	s := &scenario{}
	cmd := &cobra.Command{
		Use:   "simple-http-scenario",
		Short: "Start running simple http scenario",
		RunE:  cli.WithContext(s.run),
	}
	return cmd
}

func (s *scenario) run(ctx context.Context, logger *zap.Logger) error {
	// Expose a metrics server
	ms, err := metrics.NewServer(
		8081,
		metrics.WithLogger(logger.Sugar()),
	)
	if err != nil {
		logger.Error("failed to create metrics server", zap.Error(err))
		return err
	}
	defer ms.Stop()
	go ms.Run()

	// Just send a http request
	client := &http.Client{
		Transport: &httpmetrics.Transport{
			UsePathAsRoute: true,
		},
	}
	resp, err := ctxhttp.Get(ctx, client, "http://httpbin.org/user-agent")
	if err != nil {
		return err
	}
	resp.Body.Close()

	// Wait until we got a signal
	<-ctx.Done()
	return nil
}
