package virtualuser

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/lotusload/lotus/pkg/cli"
	"github.com/lotusload/lotus/pkg/metrics"
	"github.com/lotusload/lotus/pkg/virtualuser"
)

type scenario struct {
	hatchRate             int
	numVirtualUsers       int
	helloWorldGRPCAddress string
}

func NewCommand() *cobra.Command {
	s := &scenario{
		hatchRate:       1,
		numVirtualUsers: 10,
	}
	cmd := &cobra.Command{
		Use:   "virtual-user-scenario",
		Short: "Start running virtual user scenario",
		RunE:  cli.WithContext(s.run),
	}
	cmd.Flags().IntVar(&s.hatchRate, "hatch-rate", s.hatchRate, "How many virtual users should be spwawn per second")
	cmd.Flags().IntVar(&s.numVirtualUsers, "num-virtual-users", s.numVirtualUsers, "How many virtual users should be spawn in total")
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

	// Start a group of virtual users
	group := virtualuser.NewGroup(
		s.numVirtualUsers,
		s.hatchRate,
		func() (virtualuser.VirtualUser, error) {
			return newUser(s.helloWorldGRPCAddress, logger)
		},
	)
	defer group.Stop(time.Second)
	go group.Run(ctx)

	// Wait until we got a signal
	<-ctx.Done()
	return nil
}
