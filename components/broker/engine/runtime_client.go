package engine

import (
	"context"
	"os"

	"github.com/xigxog/kubefox/libs/core/grpc"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/platform"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RuntimeClient struct {
	Broker

	client grpc.RuntimeServiceClient
}

func NewRuntimeClient(brk Broker) *RuntimeClient {
	creds, err := platform.NewGRPCClientCreds(brk.Config().Namespace)
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

	conn, err := gogrpc.Dial(brk.Config().RuntimeSrvAddr,
		gogrpc.WithTransportCredentials(creds),
		gogrpc.WithTimeout(brk.ConnectTimeout()),
		gogrpc.WithBlock(),
	)
	if err != nil {
		brk.Log().Errorf("unable to connect to runtime server: %v", err)
		os.Exit(kubefox.RpcServerErrorCode)
	}

	return &RuntimeClient{
		Broker: brk,
		client: grpc.NewRuntimeServiceClient(conn),
	}
}

func (cl *RuntimeClient) SendEvent(ctx context.Context, req kubefox.DataEvent) (resp kubefox.DataEvent) {
	data, err := cl.client.Invoke(ctx, req.GetData())
	if err != nil {
		resp = req.ChildErrorEvent(err)
		cl.Log().Error(resp.GetError())
		return
	}

	resp = kubefox.EventFromData(data)
	resp.SetParent(req)

	return
}
