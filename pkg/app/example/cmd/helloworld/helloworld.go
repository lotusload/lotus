package helloworld

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	helloworldproto "github.com/lotusload/lotus/pkg/app/example/helloworld"
	"github.com/lotusload/lotus/pkg/cli"
)

type server struct {
	grpcPort int
	httpPort int
}

func NewCommand() *cobra.Command {
	s := &server{
		grpcPort: 8080,
		httpPort: 9090,
	}
	cmd := &cobra.Command{
		Use:   "helloworld",
		Short: "Start running helloworld server",
		RunE:  cli.WithContext(s.run),
	}
	cmd.Flags().IntVar(&s.grpcPort, "grpc-port", s.grpcPort, "Port number used to expose grpc server")
	cmd.Flags().IntVar(&s.httpPort, "http-port", s.httpPort, "Port number used to expose http server")
	return cmd
}

func (s *server) run(ctx context.Context, logger *zap.Logger) error {
	// Start a grpc server to handle grpc calls defined in helloworld.proto
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.grpcPort))
	if err != nil {
		logger.Error("failed to listen on grpc port", zap.Int("port", s.grpcPort))
		return err
	}
	grpcServer := grpc.NewServer()
	defer grpcServer.GracefulStop()
	helloworldproto.RegisterGreeterServer(grpcServer, &api{})
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("failed to server", zap.Error(err))
		}
	}()

	// Start a http server to handle http requests
	mux := http.NewServeMux()
	mux.HandleFunc("/", httpHandler)
	mux.HandleFunc("/account", httpHandler)
	mux.HandleFunc("/account/profile", httpHandler)
	mux.HandleFunc("/api/message", httpHandler)
	hs := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.httpPort),
		Handler: mux,
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		hs.Shutdown(ctx)
	}()
	go func() {
		err := hs.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Error("failed to run metrics server", zap.Error(err))
		}
	}()

	// Wait until we got a signal
	<-ctx.Done()
	return nil
}

func httpHandler(w http.ResponseWriter, req *http.Request) {
	codes := []int{200, 200, 200, 200, 200, 200, 200, 200, 404, 500}
	code := codes[rand.Int()%len(codes)]
	response := fmt.Sprintf("reponse code: %d", code)

	d := time.Duration((rand.Int()%5)+1) * 50 * time.Millisecond
	time.Sleep(d)

	if code == 200 {
		fmt.Fprintln(w, response)
		return
	}
	http.Error(w, response, code)
}

type api struct{}

func (a *api) SayHello(ctx context.Context, in *helloworldproto.HelloRequest) (*helloworldproto.HelloResponse, error) {
	time.Sleep(time.Duration(rand.Float64()) * time.Second)

	if rand.Int()%200 == 0 {
		return nil, status.Error(codes.Internal, "api: internal error")
	}
	return &helloworldproto.HelloResponse{
		Message: "Hello " + in.Name,
	}, nil
}

func (a *api) GetProfile(ctx context.Context, in *helloworldproto.ProfileRequest) (*helloworldproto.ProfileResponse, error) {
	time.Sleep(time.Duration(rand.Float64()) * time.Second)

	if rand.Int()%100 == 0 {
		return nil, status.Error(codes.Internal, "api: internal error")
	}
	return &helloworldproto.ProfileResponse{
		Name:     in.Name,
		Homepage: fmt.Sprintf("http://helloworld/%s", in.Name),
	}, nil
}
