package engine

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/xigxog/kubefox/components/broker/config"
	"github.com/xigxog/kubefox/libs/core/grpc"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/logkf"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	grpc.UnimplementedBrokerServer

	wrapped  *gogrpc.Server
	brk      Broker
	routesKV nats.KeyValue

	log *logkf.Logger
}

type gRPCEvent struct {
	evt *kubefox.Event
	err error
}

func NewGRPCServer(brk Broker) *GRPCServer {
	return &GRPCServer{
		brk: brk,
		log: logkf.Global,
	}
}

func (srv *GRPCServer) Start(ctx context.Context, routesKV nats.KeyValue) error {
	srv.log.Debug("grpc server starting")
	srv.routesKV = routesKV

	// creds, err := kfp.NewGPRCSrvCreds(ctx, kfp.BrokerCertsDir)
	// if err != nil {
	// 	srv.log.Errorf("error reading cert: %v", err)
	// 	os.Exit(model.RpcServerErrorCode)
	// }
	creds, err := credentials.NewServerTLSFromFile(kubefox.PathTLSCert, kubefox.PathTLSKey)
	if err != nil {
		return srv.log.ErrorN("%v", err)
	}
	srv.wrapped = gogrpc.NewServer(
		gogrpc.Creds(creds),
		gogrpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)

	grpc.RegisterBrokerServer(srv.wrapped, srv)
	reflection.Register(srv.wrapped)

	lis, err := net.Listen("tcp", config.GRPCSrvAddr)
	if err != nil {
		return srv.log.ErrorN("%v", err)
	}

	go func() {
		if err = srv.wrapped.Serve(lis); err != nil {
			srv.log.Error(err)
			os.Exit(1)
		}
	}()

	srv.log.Info("grpc server started")
	return nil
}

func (srv *GRPCServer) Shutdown(timeout time.Duration) {
	if srv.wrapped == nil {
		return
	}
	srv.log.Info("grpc server shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	stoppedCh := make(chan struct{})
	go func() {
		srv.wrapped.GracefulStop()
		stoppedCh <- struct{}{}
	}()

	// wait for graceful shutdown or context to timeout
	select {
	case <-stoppedCh:
		srv.log.Debug("grpc server gracefully stopped")
	case <-ctx.Done():
		srv.log.Warn("unable to stop grpc server gracefully, forcing stop")
		srv.wrapped.Stop()
	}
}

func (srv *GRPCServer) Subscribe(stream grpc.Broker_SubscribeServer) error {
	var (
		err       error
		authToken string
		comp      *kubefox.Component
		sub       ReplicaSubscription
		sendMutex sync.Mutex
	)

	if authToken, comp, err = parseMD(stream); err != nil {
		return srv.log.ErrorN("%v", err)
	}
	log := srv.log.WithComponent(comp)

	if err := srv.brk.AuthorizeComponent(stream.Context(), comp, authToken); err != nil {
		log.Error(err)
		return fmt.Errorf("unauthorized")
	}

	// The first event sent should be to register the component.
	regEvt, err := stream.Recv()
	if err != nil {
		return log.ErrorN("component registration failed: %v", err)
	}
	if regEvt.Type != string(kubefox.EventTypeRegister) {
		return log.ErrorN("component registration failed: expected event of type %s but got %s", kubefox.EventTypeRegister, regEvt.Type)
	}
	compReg := &kubefox.ComponentReg{}
	if err := regEvt.Bind(compReg); err != nil {
		return log.ErrorN("component registration failed: %v", err)
	}

	sendEvt := func(mEvt *kubefox.MatchedEvent) error {
		// Protect the stream from being called by multiple threads.
		sendMutex.Lock()
		defer sendMutex.Unlock()

		log.DebugEw("send event", mEvt.Event)

		return stream.Send(mEvt)
	}

	sub, err = srv.brk.Subscribe(stream.Context(), &SubscriptionConf{
		Component:   comp,
		CompReg:     compReg,
		SendFunc:    sendEvt,
		EnableGroup: true,
	})
	if err != nil {
		return log.ErrorN("%v", err)
	}

	log.Info("component subscribed")
	defer func() {
		sub.Cancel(err)
		log.Info("component unsubscribed")
	}()

	// This simply receives events from the gRPC stream and places them on a
	// channel. This makes checking for the context to be done easier by using a
	// select in the next code block.
	recvCh := make(chan *gRPCEvent)
	go func() {
		for {
			evt, err := stream.Recv()
			recvCh <- &gRPCEvent{evt: evt, err: err}
		}
	}()

	for {
		select {
		case gRPCEvt := <-recvCh:
			err := gRPCEvt.err
			evt := gRPCEvt.evt
			if err != nil {
				status, _ := status.FromError(err)
				switch {
				case err == io.EOF || evt == nil:
					log.Debug("send stream closed")
				case status.Code() == codes.Canceled:
					log.Debug("context canceled")
				default:
					log.Error(err)
				}
				return err
			}

			log.DebugEw("receive event", evt)
			err = srv.brk.RecvEvent(&ReceivedEvent{
				Event:    evt,
				Receiver: EventReceiverGRPC,
			})
			if err != nil {
				log.DebugEw("receive event failed", evt, err)
			}

		case <-sub.Context().Done():
			if sub.Err() != nil {
				log.Error(sub.Err())
			}

			return sub.Err()
		}
	}
}

func parseMD(stream grpc.Broker_SubscribeServer) (authToken string, comp *kubefox.Component, err error) {
	md, found := metadata.FromIncomingContext(stream.Context())
	if !found {
		err = fmt.Errorf("gRPC metadata missing")
		return
	}

	var compId, compCommit, compName string
	compId, err = getMD(md, "componentId")
	if err != nil {
		return
	}
	compCommit, err = getMD(md, "componentCommit")
	if err != nil {
		return
	}
	compName, err = getMD(md, "componentName")
	if err != nil {
		return
	}
	authToken, err = getMD(md, "authToken")
	if err != nil {
		return
	}

	comp = &kubefox.Component{
		Id:     compId,
		Commit: compCommit,
		Name:   compName,
	}

	return
}

func getMD(md metadata.MD, key string) (string, error) {
	arr := md.Get(key)
	switch len(arr) {
	case 1:
		v := arr[0]
		if v == "" {
			return "", fmt.Errorf("%s not provided", key)
		}
		return v, nil
	case 0:
		return "", fmt.Errorf("%s not provided", key)
	default:
		return "", fmt.Errorf("more than one %s provided", key)
	}
}
