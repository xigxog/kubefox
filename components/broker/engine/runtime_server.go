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

type RuntimeServer struct {
	Broker
	grpc.UnimplementedRuntimeServiceServer

	server *gogrpc.Server
}

func NewRuntimeServer(brk Broker) *RuntimeServer {
	creds, err := platform.NewGPRCSrvCreds(brk.Config().Namespace)
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

	return &RuntimeServer{
		Broker: brk,
		server: gogrpc.NewServer(gogrpc.Creds(creds),
			gogrpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
			gogrpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
		),
	}
}

func (srv *RuntimeServer) Start() {
	grpc.RegisterRuntimeServiceServer(srv.server, srv)
	go srv.serve(srv.server, srv.Config().RuntimeSrvAddr)
	srv.Log().Infof("platform gRPC server started on %s", srv.Config().RuntimeSrvAddr)
}

func (srv *RuntimeServer) serve(server *gogrpc.Server, addr string) {
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

func (srv *RuntimeServer) Shutdown() {
	srv.Log().Info("runtime server shutting down")
	srv.server.GracefulStop()
}

func (srv *RuntimeServer) Invoke(ctx context.Context, req *grpc.Data) (*grpc.Data, error) {
	resp := srv.InvokeLocalComponent(ctx, kubefox.EventFromData(req))
	return resp.GetData(), resp.GetError()
}
