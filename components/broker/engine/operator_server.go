package engine

import (
	"context"
	"net"
	"os"

	"github.com/xigxog/kubefox/libs/core/grpc"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/platform"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

type OperatorServer struct {
	Broker
	grpc.UnimplementedRuntimeServiceServer

	server *gogrpc.Server
}

func NewOperatorServer(ctx context.Context, brk Broker) *OperatorServer {
	creds, err := platform.NewGPRCSrvCreds(ctx, platform.OperatorCertsDir)
	if err != nil {
		if brk.Config().IsDevMode {
			brk.Log().Warnf("error reading cert: %v", err)
			brk.Log().Warn("dev mode enabled, using insecure connection")
			creds = insecure.NewCredentials()
		} else {
			brk.Log().Errorf("error reading cert: %v", err)
			os.Exit(kubefox.RpcServerErrorCode)
		}
	}

	return &OperatorServer{
		Broker: brk,
		server: gogrpc.NewServer(gogrpc.Creds(creds),
			gogrpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
			gogrpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
		),
	}
}

func (srv *OperatorServer) Start() {
	grpc.RegisterRuntimeServiceServer(srv.server, srv)
	go srv.serve(srv.server, srv.Config().OperatorAddr)
	srv.Log().Infof("operator gRPC server started on %s", srv.Config().OperatorAddr)
}

func (srv *OperatorServer) serve(server *gogrpc.Server, addr string) {
	reflection.Register(server)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		srv.Log().Error(err)
		os.Exit(kubefox.RpcServerErrorCode)

	}
	defer lis.Close()

	err = server.Serve(lis)
	if err != nil {
		srv.Log().Error(err)
		os.Exit(kubefox.RpcServerErrorCode)
	}
}

func (srv *OperatorServer) Shutdown() {
	srv.Log().Info("operator shutting down")
	srv.server.GracefulStop()
}

func (srv *OperatorServer) Invoke(ctx context.Context, req *grpc.Data) (*grpc.Data, error) {
	resp := srv.InvokeLocalComponent(ctx, kubefox.EventFromData(req))
	return resp.GetData(), resp.GetError()
}
