package kit

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	// _ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/xigxog/kubefox/libs/core/grpc"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/logkf"
	"github.com/xigxog/kubefox/libs/core/utils"

	"github.com/google/uuid"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

type EventHandlerFunc func(kit Kontext) error

type Kit interface {
	context.Context

	Start()
	Route(string, EventHandlerFunc)
	Log() *logkf.Logger
}

type kit struct {
	context.Context

	comp *kubefox.Component

	brokerAddr string
	caCertPath string
	platform   string
	namespace  string

	routes []*route

	brk    grpc.Broker_SubscribeClient
	reqMap map[string]chan *kubefox.Event

	reqMapMutex sync.Mutex
	sendMutex   sync.Mutex

	cancel context.CancelFunc
	log    *logkf.Logger
}

type route struct {
	kubefox.Route

	handler EventHandlerFunc
}

func New() Kit {
	svc := &kit{
		comp: &kubefox.Component{
			Id: uuid.NewString(),
		},
		routes: make([]*route, 0),
		reqMap: make(map[string]chan *kubefox.Event),
	}

	var help bool
	var logFormat, logLevel string
	//-tls-skip-verify
	flag.StringVar(&svc.comp.Name, "name", "", "Component name; environment variable 'KUBEFOX_COMPONENT_NAME'. (required)")
	flag.StringVar(&svc.comp.Commit, "commit", "", "Commit the Component was built from; environment variable 'KUBEFOX_COMPONENT_COMMIT'. (required)")
	flag.StringVar(&svc.platform, "platform", "", "Platform the Component is part of; environment variable 'KUBEFOX_PLATFORM'.")
	flag.StringVar(&svc.namespace, "namespace", "", "Kubernetes namespace of the Component; environment variable 'KUBEFOX_NAMESPACE'.")
	flag.StringVar(&svc.brokerAddr, "broker-addr", "127.0.0.1:6060", "Address of the Broker gRPC server; environment variable 'KUBEFOX_BROKER_GRPC_ADDR'.")
	flag.StringVar(&svc.caCertPath, "ca-cert-path", kubefox.CACertPath, "Path of file containing KubeFox root CA certificate; environment variable 'KUBEFOX_CA_CERT_PATH'.")
	flag.StringVar(&logFormat, "log-format", "console", "Log format; environment variable 'KUBEFOX_LOG_FORMAT'. [options 'json', 'console']")
	flag.StringVar(&logLevel, "log-level", "debug", "Log level; environment variable 'KUBEFOX_LOG_LEVEL'. [options 'debug', 'info', 'warn', 'error']")
	flag.BoolVar(&help, "help", false, "Show usage for component.")
	flag.Parse()

	if help {
		fmt.Fprintf(flag.CommandLine.Output(), `
Flags can be set using names below or the environment variable listed.

Flags:
`)
		flag.PrintDefaults()
		os.Exit(0)
	}

	svc.comp.Name = utils.ResolveFlag(svc.comp.Name, "KUBEFOX_COMPONENT_NAME", "")
	svc.comp.Commit = utils.ResolveFlag(svc.comp.Commit, "KUBEFOX_COMPONENT_COMMIT", "")
	svc.brokerAddr = utils.ResolveFlag(svc.brokerAddr, "KUBEFOX_BROKER_GRPC_ADDR", "127.0.0.1:6060")
	svc.caCertPath = utils.ResolveFlag(svc.caCertPath, "KUBEFOX_CA_CERT_PATH", kubefox.CACertPath)
	logFormat = utils.ResolveFlag(logFormat, "KUBEFOX_LOG_FORMAT", "console")
	logLevel = utils.ResolveFlag(logLevel, "KUBEFOX_LOG_LEVEL", "debug")

	utils.CheckRequiredFlag("name", svc.comp.Name)
	utils.CheckRequiredFlag("commit", svc.comp.Commit)

	logkf.Global = logkf.
		BuildLoggerOrDie(logFormat, logLevel).
		WithComponent(svc.comp)
	defer logkf.Global.Sync()

	svc.log = logkf.Global
	// ctrl.SetLogger(logr.Logger{})

	svc.Context, svc.cancel = context.WithCancel(context.Background())
	svc.log.Info("kit service created 🦊")

	return svc
}

func (svc *kit) Log() *logkf.Logger {
	return svc.log
}

func (svc *kit) Route(rule string, handler EventHandlerFunc) {
	r := &route{
		Route: kubefox.Route{
			Id:   len(svc.routes),
			Rule: rule,
		},
		handler: handler,
	}
	svc.routes = append(svc.routes, r)
}

func (svc *kit) Start() {
	var conn *gogrpc.ClientConn

	defer func() {
		svc.cancel()
		if err := conn.Close(); err != nil {
			svc.log.Fatal(err)
		}
	}()

	grpcCfg := `{
		"methodConfig": [{
		  "name": [{"service": "", "method": ""}],
		  "waitForReady": false,
		  "retryPolicy": {
			  "MaxAttempts": 3,
			  "InitialBackoff": "3s",
			  "MaxBackoff": "6s",
			  "BackoffMultiplier": 2.0,
			  "RetryableStatusCodes": [ "UNAVAILABLE" ]
		  }
		}]}`
	creds, err := credentials.NewClientTLSFromFile(kubefox.CACertPath, "")
	if err != nil {
		svc.log.Fatalf("unable to load root CA certificate: %v", err)
	}
	conn, err = gogrpc.Dial(svc.brokerAddr,
		gogrpc.WithPerRPCCredentials(svc),
		gogrpc.WithTransportCredentials(creds),
		gogrpc.WithDefaultServiceConfig(grpcCfg),
	)
	if err != nil {
		svc.log.Fatalf("unable to connect to broker: %v", err)
	}

	if svc.brk, err = grpc.NewBrokerClient(conn).Subscribe(svc); err != nil {
		svc.log.Fatalf("subscribing to broker failed: %v", err)
	}

	reg := &kubefox.ComponentReg{
		Routes: make([]*kubefox.Route, len(svc.routes)),
	}
	for i := range svc.routes {
		reg.Routes[i] = &svc.routes[i].Route
	}

	regEvt := kubefox.NewEvent()
	regEvt.Type = string(kubefox.RegisterType)
	regEvt.Category = kubefox.Category_CATEGORY_SINGLE
	if err := regEvt.SetJSON(reg); err != nil {
		svc.log.Fatalf("unable to connect to broker: %v", err)
	}
	if err := svc.sendEvent(regEvt); err != nil {
		svc.log.Fatalf("unable to connect to broker: %v", err)
	}

	svc.log.Info("kit service started 🏁")

	for {
		mEvt, err := svc.brk.Recv()
		if err != nil {
			status, _ := status.FromError(err)
			switch {
			case err == io.EOF:
				svc.log.Debug("send stream closed")
			case status.Code() == codes.Canceled:
				svc.log.Debug("context canceled")
			default:
				svc.log.Error(err)
			}

			return
		}

		svc.log.DebugEw("received event", mEvt.Event)

		switch mEvt.Event.Category {
		case kubefox.Category_CATEGORY_REQUEST:
			go svc.recvReq(mEvt)

		case kubefox.Category_CATEGORY_RESPONSE:
			go svc.recvResp(mEvt.Event)

		default:
			svc.log.Debug("default")
		}
	}
}

func (svc *kit) recvReq(req *kubefox.MatchedEvent) {
	log := svc.log.WithEvent(req.Event)

	ctx, cancel := context.WithTimeout(context.Background(), req.Event.TTL())
	defer cancel()

	ktx := &kontext{
		Context: ctx,
		EventRW: kubefox.NewEventRW(req.Event),
		resp:    kubefox.NewResp(req.Event, svc.comp),
		kit:     svc,
		env:     req.Env,
		start:   time.Now(),
		log:     log,
	}

	var err error
	if req.RouteId < 0 || req.RouteId >= int64(len(svc.routes)) {
		err = fmt.Errorf("invalid route id %d", req.RouteId)
	} else {
		err = svc.routes[req.RouteId].handler(ktx)
	}

	if err != nil {
		ktx.resp = kubefox.NewResp(req.Event, svc.comp)
		ktx.resp.Type = string(kubefox.ErrorEventType)

		log.Error(err)
		if err := ktx.Resp().SendString(err.Error()); err != nil {
			log.Error(err)
		}
	}
}

func (svc *kit) sendReq(ctx context.Context, req *kubefox.Event) (*kubefox.Event, error) {
	log := svc.log.WithEvent(req)
	log.Debug("send request")

	svc.reqMapMutex.Lock()
	respCh := make(chan *kubefox.Event)
	svc.reqMap[req.Id] = respCh
	svc.reqMapMutex.Unlock()

	if err := svc.sendEvent(req); err != nil {
		return nil, log.ErrorN("%v", err)
	}

	select {
	case resp := <-respCh:
		return resp, nil

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (svc *kit) recvResp(resp *kubefox.Event) {
	svc.log.DebugEw("recv event", resp)

	svc.reqMapMutex.Lock()
	respCh, found := svc.reqMap[resp.ParentId]
	delete(svc.reqMap, resp.ParentId)
	svc.reqMapMutex.Unlock()

	if !found {
		svc.log.ErrorEw("matching request not found for response", resp)
		return
	}

	respCh <- resp
}

func (svc *kit) sendEvent(evt *kubefox.Event) error {
	svc.log.DebugEw("send event", evt)

	// need to protect the stream from being called by multiple threads
	svc.sendMutex.Lock()
	defer svc.sendMutex.Unlock()

	return svc.brk.Send(evt)
}

func (svc *kit) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	b, err := os.ReadFile(kubefox.SvcAccTokenFile)
	if err == nil {
		// Return token from file is it was successfully read.
		return nil, err
	}
	token := string(b)

	return map[string]string{
		"componentId":     svc.comp.Id,
		"componentName":   svc.comp.Name,
		"componentCommit": svc.comp.Commit,
		"authToken":       token,
	}, nil
}

func (svc *kit) RequireTransportSecurity() bool {
	return true
}
