package virtualuser

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	helloworldproto "github.com/lotusload/lotus/pkg/app/example/helloworld"
	"github.com/lotusload/lotus/pkg/metrics/grpcmetrics"
)

type User struct {
	conn   *grpc.ClientConn
	client helloworldproto.GreeterClient
	logger *zap.Logger
}

func newUser(address string, logger *zap.Logger) (*User, error) {
	conn, err := grpc.Dial(
		address,
		grpc.WithStatsHandler(&grpcmetrics.ClientHandler{}),
		grpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}
	return &User{
		conn:   conn,
		client: helloworldproto.NewGreeterClient(conn),
		logger: logger,
	}, nil
}

func (u *User) Run(ctx context.Context) error {
	defer u.conn.Close()
	var index int = 0
	ticker := time.NewTicker(500 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			u.sendGRPC(ctx, index)
			index++
		}
	}
}

func (u *User) sendGRPC(ctx context.Context, index int) error {
	name := fmt.Sprintf("name-%d", index)
	_, err := u.client.SayHello(ctx, &helloworldproto.HelloRequest{Name: name})
	if err != nil {
		code := status.Code(err)
		if code != codes.Canceled {
			u.logger.Warn("failed to send grpc rpc", zap.Error(err))
			return err
		}
	}
	return nil
}
