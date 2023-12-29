package engine

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/components/broker/config"
	"github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/grpc"
	"github.com/xigxog/kubefox/logkf"

	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	grpc.UnimplementedBrokerServer

	wrapped *gogrpc.Server
	brk     Broker

	log *logkf.Logger
}

func NewGRPCServer(brk Broker) *GRPCServer {
	return &GRPCServer{
		brk: brk,
		log: logkf.Global,
	}
}

func (srv *GRPCServer) Start(ctx context.Context) error {
	srv.log.Debug("grpc server starting")

	creds, err := credentials.NewServerTLSFromFile(api.PathTLSCert, api.PathTLSKey)
	if err != nil {
		return core.ErrUnexpected(err)
	}
	srv.wrapped = gogrpc.NewServer(
		gogrpc.Creds(creds),
		// gogrpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)

	grpc.RegisterBrokerServer(srv.wrapped, srv)
	reflection.Register(srv.wrapped)

	lis, err := net.Listen("tcp", config.GRPCSrvAddr)
	if err != nil {
		return core.ErrPortUnavailable(err)
	}

	go func() {
		if err = srv.wrapped.Serve(lis); err != nil {
			srv.log.Fatal(err)
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
	sub, err := srv.subscribe(stream)

	l := srv.log
	if sub != nil {
		l = l.WithComponent(sub.Component())
	}

	if err != nil {
		status, _ := status.FromError(err)
		switch {
		case err == io.EOF:
			l.Debug("send stream closed by component")
		case status.Code() == codes.Canceled:
			l.Debug("context canceled")
		case status.Code() == codes.PermissionDenied:
			l.Warn(err)
		default:
			l.Error(err)
		}

		if sub != nil {
			sub.Cancel(err)
		}
	}

	l.Info("component unsubscribed")
	return err
}

func (srv *GRPCServer) subscribe(stream grpc.Broker_SubscribeServer) (ReplicaSubscription, error) {
	var (
		err       error
		authToken string
		comp      *core.Component
		sub       ReplicaSubscription
		sendMutex sync.Mutex
	)

	if authToken, comp, err = parseMD(stream); err != nil {
		return nil, core.ErrUnauthorized(err)
	}
	subLog := srv.log.WithComponent(comp)

	if err := srv.brk.AuthorizeComponent(stream.Context(), comp, authToken); err != nil {
		return nil, core.ErrUnauthorized(err)
	}

	sendEvt := func(evt *BrokerEvent) error {
		// Protect the stream from being called by multiple threads.
		sendMutex.Lock()
		defer sendMutex.Unlock()

		subLog.WithEvent(evt.Event).Debug("send event")

		if err := stream.Send(evt.MatchedEvent()); err != nil {
			return core.ErrUnexpected(err)
		}
		return nil
	}

	// The first event sent should be the component spec.
	regEvt, err := stream.Recv()
	if err != nil {
		return nil, core.ErrUnauthorized(err)
	}
	if regEvt.EventType() != api.EventTypeRegister {
		return nil, core.ErrUnauthorized(fmt.Errorf("expected event of type %s but got %s",
			api.EventTypeRegister, regEvt.Type))
	}
	compDef := &api.ComponentDefinition{}
	if err := regEvt.Bind(compDef); err != nil {
		return nil, core.ErrUnauthorized(err)
	}

	sub, err = srv.brk.Subscribe(stream.Context(), &SubscriptionConf{
		Component:    comp,
		ComponentDef: compDef,
		SendFunc:     sendEvt,
		EnableGroup:  true,
	})
	if err != nil {
		return nil, err
	}

	regResp := &BrokerEvent{
		Event: core.NewResp(core.EventOpts{
			Type:   api.EventTypeRegister,
			Parent: regEvt,
			Source: srv.brk.Component(),
			Target: regEvt.Source,
		}),
	}
	if err := sendEvt(regResp); err != nil {
		return sub, err
	}

	subLog.Info("component subscribed")

	// This simply receives events from the gRPC stream and places them on a
	// channel. This makes checking for the context to be done easier by using a
	// select in the next code block.
	recvCh := make(chan *core.Event)
	go func() {
		for {
			if !sub.IsActive() {
				return
			}
			evt, err := stream.Recv()
			if err != nil {
				sub.Cancel(err)
				return
			}
			recvCh <- evt
		}
	}()

	for {
		select {
		case evt := <-recvCh:
			l := subLog.WithEvent(evt)
			l.Debug("receive event")

			if evt.Source == nil {
				evt.Source = comp
			} else if !evt.Source.Equal(comp) {
				return sub, core.ErrUnauthorized(
					fmt.Errorf("event from '%s' claiming to be '%s'", comp.Key(), evt.Source.Key()))
			}
			evt.Source.BrokerId = srv.brk.Component().Id

			brkEvt := srv.brk.RecvEvent(evt, ReceiverGRPCServer)
			if err := <-brkEvt.Done(); err != nil &&
				evt.Category == core.Category_REQUEST &&
				err.Code() != core.CodeTimeout {

				errResp := core.NewErr(err, core.EventOpts{
					Parent: evt,
					Source: srv.brk.Component(),
					Target: evt.Source,
				})

				if err := sendEvt(&BrokerEvent{Event: errResp}); err != nil {
					return sub, err
				}
			}

		case <-sub.Context().Done():
			return sub, sub.Err()
		}
	}
}

func parseMD(stream grpc.Broker_SubscribeServer) (authToken string, comp *core.Component, err error) {
	md, found := metadata.FromIncomingContext(stream.Context())
	if !found {
		err = fmt.Errorf("gRPC metadata missing")
		return
	}

	var compId, compCommit, compName, compType string
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
	compType, err = getMD(md, "componentType")
	if err != nil {
		return
	}
	authToken, err = getMD(md, "authToken")
	if err != nil {
		return
	}

	comp = &core.Component{
		Type:   compType,
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
