package threesteps

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/net/context/ctxhttp"
	"google.golang.org/grpc"

	helloworldproto "github.com/lotusload/lotus/pkg/app/example/helloworld"
	"github.com/lotusload/lotus/pkg/cli"
	"github.com/lotusload/lotus/pkg/metrics"
	"github.com/lotusload/lotus/pkg/metrics/grpcmetrics"
	"github.com/lotusload/lotus/pkg/metrics/httpmetrics"
)

var steps = []string{
	"preparer",
	"worker",
	"cleaner",
}

type scenario struct {
	step                  string
	duration              time.Duration
	helloWorldGRPCAddress string
	helloWorldHTTPAddress string

	httpClient *http.Client
	grpcClient helloworldproto.GreeterClient
}

func NewCommand() *cobra.Command {
	s := &scenario{
		step:     steps[0],
		duration: 10 * time.Second,
	}
	cmd := &cobra.Command{
		Use:   "three-steps-scenario",
		Short: "Start running three steps scenario",
		RunE:  cli.WithContext(s.run),
	}
	cmd.Flags().StringVar(&s.step, "step", s.step, "The step will be run")
	cmd.Flags().DurationVar(&s.duration, "duration", s.duration, "How long this step should be run (Only valid preparer or cleaner)")
	cmd.Flags().StringVar(&s.helloWorldGRPCAddress, "helloworld-grpc-address", s.helloWorldGRPCAddress, "The grpc address to helloword service")
	cmd.MarkFlagRequired("helloworld-grpc-address")
	cmd.Flags().StringVar(&s.helloWorldHTTPAddress, "helloworld-http-address", s.helloWorldHTTPAddress, "The http address to helloword service")
	cmd.MarkFlagRequired("helloworld-http-address")
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

	// For worker step
	if s.step == "worker" {
		return s.runWorker(ctx, logger)
	}

	// For preparer or cleaner step we just wait
	select {
	case <-time.After(s.duration):
		return nil
	case <-ctx.Done():
		return nil
	}
}

func (s *scenario) runWorker(ctx context.Context, logger *zap.Logger) error {
	s.httpClient = &http.Client{
		Transport: &httpmetrics.Transport{
			UsePathAsRoute: true,
		},
	}
	conn, err := grpc.Dial(
		s.helloWorldGRPCAddress,
		grpc.WithStatsHandler(&grpcmetrics.ClientHandler{}),
		grpc.WithInsecure(),
	)
	if err != nil {
		logger.Error("failed to connect to helloworl server", zap.Error(err))
		return err
	}
	defer conn.Close()
	s.grpcClient = helloworldproto.NewGreeterClient(conn)

	// Periodically send the requests
	ticker := time.NewTicker(200 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			x := rand.Int() % 5
			if x == 0 {
				s.sendHTTP(ctx, logger)
				break
			}
			s.sendGRPC(ctx, x+1, logger)
		case <-ctx.Done():
			return nil
		}
	}
}

func (s *scenario) sendHTTP(ctx context.Context, logger *zap.Logger) {
	paths := []string{
		"/",
		"/account",
		"/account/profile",
		"/api/message",
	}
	x := rand.Int() % len(paths)
	address := fmt.Sprintf("%s%s", s.helloWorldHTTPAddress, paths[x])

	//ctx, _ = metrics.ContextWithRoute(ctx, "foo")
	resp, err := ctxhttp.Get(ctx, s.httpClient, address)
	if err != nil {
		return
	}
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()

	b := strings.NewReader(`{"key":"value"}`)
	resp, err = ctxhttp.Post(ctx, s.httpClient, address, "application/json", b)
	if err != nil {
		return
	}
	//io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
}

func (s *scenario) sendGRPC(ctx context.Context, rpcs int, logger *zap.Logger) {
	for i := 0; i < rpcs; i++ {
		name := fmt.Sprintf("name-%d", i)
		var err error
		if i%2 == 0 {
			_, err = s.grpcClient.SayHello(ctx, &helloworldproto.HelloRequest{Name: name})
		} else {
			_, err = s.grpcClient.GetProfile(ctx, &helloworldproto.ProfileRequest{Name: name})
		}
		if err != nil {
			logger.Warn("failed to send grpc rpc", zap.Error(err))
		}
	}
}
