package simplegrpc

import (
	"context"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	helloworldproto "github.com/nghialv/lotus/pkg/app/example/helloworld"
	"github.com/nghialv/lotus/pkg/cli"
	"github.com/nghialv/lotus/pkg/metrics"
	"github.com/nghialv/lotus/pkg/metrics/grpcmetrics"
)

type scenario struct {
	helloWorldGRPCAddress string
}

func NewCommand() *cobra.Command {
	s := &scenario{}
	cmd := &cobra.Command{
		Use:   "simple-grpc-scenario",
		Short: "Start running simple grpc scenario",
		RunE:  cli.WithContext(s.run),
	}
	cmd.Flags().StringVar(&s.helloWorldGRPCAddress, "helloworld-grpc-address", s.helloWorldGRPCAddress, "The grpc address to helloword service")
	cmd.MarkFlagRequired("helloworld-grpc-address")
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

	// Just send a grpc rpc to helloworld server
	conn, err := grpc.Dial(
		s.helloWorldGRPCAddress,
		grpc.WithStatsHandler(&grpcmetrics.ClientHandler{}),
		grpc.WithInsecure(),
	)
	if err != nil {
		logger.Error("failed to connect to helloworld server", zap.Error(err))
		return err
	}
	defer conn.Close()

	client := helloworldproto.NewGreeterClient(conn)
	_, err = client.SayHello(ctx, &helloworldproto.HelloRequest{
		Name: "lotus",
	})
	if err != nil {
		logger.Error("failed to say hello to server", zap.Error(err))
	}

	// Wait until we got a signal
	<-ctx.Done()
	return nil
}
