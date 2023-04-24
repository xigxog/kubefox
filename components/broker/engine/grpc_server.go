package engine

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/xigxog/kubefox/libs/core/grpc"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/platform"

	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

type GRPCServer struct {
	Broker
	grpc.UnimplementedComponentServiceServer

	server *gogrpc.Server

	subMap  map[string]*context.CancelFunc
	eventCh chan kubefox.DataEvent
}

func NewGRPCServer(brk Broker) *GRPCServer {
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

	return &GRPCServer{
		Broker:  brk,
		subMap:  make(map[string]*context.CancelFunc),
		eventCh: make(chan kubefox.DataEvent, 512),
		server:  gogrpc.NewServer(gogrpc.Creds(creds)),
	}
}

func (srv *GRPCServer) Start() {
	grpc.RegisterComponentServiceServer(srv.server, srv)
	go srv.serve(srv.server, srv.Config().GRPCSrvAddr)
	srv.Log().Infof("component gRPC server started on %s", srv.Config().GRPCSrvAddr)
}

func (srv *GRPCServer) serve(server *gogrpc.Server, addr string) {
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

func (srv *GRPCServer) Shutdown() {
	srv.Log().Info("component gRPC server shutting down")
	for id := range srv.subMap {
		srv.Unsubscribe(context.Background(), &grpc.SubscribeRequest{Id: id})
	}
	close(srv.eventCh)

	srv.server.GracefulStop()
}

func (srv *GRPCServer) SendEvent(ctx context.Context, req kubefox.DataEvent) (resp kubefox.DataEvent) {
	if len(srv.subMap) == 0 {
		resp = req.ChildErrorEvent(fmt.Errorf("request failed: no components subscribed"))
		srv.Log().Error(resp.GetError())
		return
	}

	ing, err := srv.Blocker().NewRespListener(ctx, req.GetId())
	if err != nil {
		resp = req.ChildErrorEvent(fmt.Errorf("request failed: no components subscribed"))
		srv.Log().Error(resp.GetError())
		return
	}

	// Send request to component.
	select {
	case srv.eventCh <- req:
		break
	case <-ctx.Done():
		resp = req.ChildErrorEvent(ctx.Err())
		srv.Log().Error(resp.GetError())
		return
	}

	// Wait for the response.
	resp, err = ing.Wait()
	if err != nil {
		resp = req.ChildErrorEvent(err)
		srv.Log().Error(resp.GetError())
		return
	}
	resp.SetParent(req)

	return
}

func (srv *GRPCServer) InvokeTarget(ctx context.Context, req *grpc.Data) (*grpc.Data, error) {
	resp := srv.InvokeRemoteComponent(ctx, kubefox.EventFromData(req))
	return resp.GetData(), resp.GetError()
}

func (srv *GRPCServer) SendResponse(ctx context.Context, resp *grpc.Data) (*grpc.Ack, error) {
	err := srv.Blocker().SendResponse(resp.ParentId, kubefox.EventFromData(resp))
	if err != nil {
		return nil, err
	}

	return &grpc.Ack{}, nil
}

func (srv *GRPCServer) Subscribe(subReq *grpc.SubscribeRequest, stream grpc.ComponentService_SubscribeServer) error {
	if subReq.Id == "" {
		return kubefox.ErrMissingId
	}

	srv.Log().Infof("component subscribed; subscription: %s", subReq.Id)

	ctx, cancel := context.WithCancel(stream.Context())
	srv.subMap[subReq.Id] = &cancel

	for {
		select {
		case req := <-srv.eventCh:
			select {
			case <-ctx.Done():
				srv.Log().Warnf("ignoring canceled request %s for subscription %s", req.GetId(), subReq.Id)
				continue
			default:
				l := srv.Log()
				if req.GetTraceId() != "" {
					l = srv.Log().With("traceId", req.GetTraceId())
				}
				l.Debugf("sending request %s to subscription %s", req.GetId(), subReq.Id)

				stream.Send(req.GetData())
			}

		case <-ctx.Done():
			delete(srv.subMap, subReq.Id)
			srv.Log().Infof("subscription %s closed", subReq.Id)
			return nil
		}
	}
}

func (srv *GRPCServer) Unsubscribe(ctx context.Context, req *grpc.SubscribeRequest) (*grpc.Ack, error) {
	if req.Id == "" {
		return nil, kubefox.ErrMissingId
	}

	cancel := srv.subMap[req.Id]
	delete(srv.subMap, req.Id)
	if cancel != nil {
		(*cancel)()
	}

	srv.Log().Infof("component unsubscribed; subscription: %s", req.Id)

	return &grpc.Ack{}, nil
}

func (srv *GRPCServer) GetConfig(ctx context.Context, req *grpc.ConfigRequest) (*grpc.ComponentConfig, error) {
	return &grpc.ComponentConfig{
		// Organization: srv.Config().Organization,
		Platform: srv.Config().Platform,
		DevMode:  srv.Config().IsDevMode,
		Component: &grpc.Component{
			Id:      srv.Component().GetId(),
			GitHash: srv.Component().GetGitHash(),
			Name:    srv.Component().GetName(),
		},
	}, nil
}

// Healthy determines if the gRPC server on the broker and the client on the
// component are still communicating by sending a health request and waiting for
// response.
func (srv *GRPCServer) Healthy(ctx context.Context) bool {
	// this purposefully bypasses the broker to avoid attaching traces
	resp := srv.SendEvent(ctx, kubefox.NewDataEvent(kubefox.HealthRequestType))
	if resp.GetError() != nil {
		srv.Log().Errorf("error checking component health: %v", resp.GetError())
		return false
	}

	return resp.GetType() == kubefox.HealthResponseType
}

func (srv *GRPCServer) Name() string {
	return "component-grpc-server"
}
