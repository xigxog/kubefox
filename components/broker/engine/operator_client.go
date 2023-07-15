package engine

import (
	"context"
	"fmt"
	"os"

	"github.com/xigxog/kubefox/libs/core/grpc"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/platform"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	ktyps "k8s.io/apimachinery/pkg/types"
)

type OperatorClient struct {
	Broker

	client grpc.RuntimeServiceClient
}

func NewOperatorClient(brk Broker) *OperatorClient {
	creds, err := platform.NewGRPCClientCreds(brk.Config().CACertPath, ktyps.NamespacedName{
		Namespace: brk.Config().Namespace,
		Name:      fmt.Sprintf("%s-%s", brk.Config().Platform, platform.RootCASecret),
	})
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

	conn, err := gogrpc.Dial(brk.Config().OperatorAddr,
		gogrpc.WithTransportCredentials(creds),
		gogrpc.WithDefaultServiceConfig(platform.GRPCServiceCfg),
		gogrpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		gogrpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)
	if err != nil {
		brk.Log().Errorf("unable to connect to operator: %v", err)
		os.Exit(kubefox.RpcServerErrorCode)
	}

	return &OperatorClient{
		Broker: brk,
		client: grpc.NewRuntimeServiceClient(conn),
	}
}

func (cl *OperatorClient) SendEvent(ctx context.Context, req kubefox.DataEvent) (resp kubefox.DataEvent) {
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
