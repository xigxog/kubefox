package kit

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/xigxog/kubefox/libs/core/grpc"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/logkf"
	"github.com/xigxog/kubefox/libs/core/utils"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
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

	if svc.comp.Name == "" {
		fmt.Fprintf(os.Stderr, "The flag 'name' is required.\n\n")
		flag.Usage()
		os.Exit(1)
	}
	if svc.comp.Commit == "" {
		fmt.Fprintf(os.Stderr, "The flag 'commit' is required.\n\n")
		flag.Usage()
		os.Exit(1)
	}

	l, err := logkf.BuildLogger(logFormat, logLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid log setting: %v\n\n", err)
		flag.Usage()
		os.Exit(1)
	}

	svc.log = l.WithComponent(svc.comp)
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

	conn, err := gogrpc.Dial(svc.brokerAddr,
		gogrpc.WithPerRPCCredentials(svc),
		gogrpc.WithTransportCredentials(insecure.NewCredentials()),
		gogrpc.WithDefaultServiceConfig(grpcCfg),
	)
	if err != nil {
		svc.log.Fatalf("unable to connect to broker: %v", err)
	}

	if svc.brk, err = grpc.NewBrokerClient(conn).Subscribe(svc); err != nil {
		svc.log.Fatalf("subscribing to broker failed: %v", err)
	}

	reg := &kubefox.ComponentRegistration{
		Routes: make([]*kubefox.Route, len(svc.routes)),
	}
	for i := range svc.routes {
		reg.Routes[i] = &svc.routes[i].Route
	}

	regEvt := kubefox.NewEvent()
	regEvt.Type = string(kubefox.RegisterType)
	regEvt.Category = kubefox.Category_CATEGORY_SINGLE
	if err := regEvt.Marshal(reg); err != nil {
		svc.log.Fatalf("unable to connect to broker: %v", err)
	}
	if err := svc.sendEvent(regEvt, 0); err != nil {
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
	start := time.Now().UnixMicro()

	resp := kubefox.NewEvent()
	resp.SetParent(req.Event)
	resp.Category = kubefox.Category_CATEGORY_RESPONSE
	resp.Source = svc.comp
	resp.Target = req.Event.Source

	k := &kontext{
		kitSvc: svc,
		req:    req.Event,
		resp:   resp,
		env:    req.Env,
		start:  start,
		log:    log,
	}

	var err error
	defer func() {
		if err != nil {
			log.Error(err)
			resp.Type = string(kubefox.ErrorEventType)
			resp.ContentType = "text/plain; charset=UTF-8"
			resp.Content = []byte(err.Error())
			if err := svc.sendEvent(resp, start); err != nil {
				log.Error(err)
			}
		}
	}()

	if req.RouteId < 0 || req.RouteId >= int64(len(svc.routes)) {
		err = fmt.Errorf("invalid route id %d", req.RouteId)
		return

	}

	err = svc.routes[req.RouteId].handler(k)
}

func (svc *kit) sendReq(req *kubefox.Event) (*kubefox.Event, error) {
	log := svc.log.WithEvent(req)
	log.Debug("sending request")

	ctx, cancel := context.WithTimeout(svc.Context, time.Second*5)
	defer cancel()

	svc.reqMapMutex.Lock()
	respCh := make(chan *kubefox.Event)
	svc.reqMap[req.Id] = respCh
	svc.reqMapMutex.Unlock()

	// TODO
	if err := svc.sendEvent(req, 0); err != nil {
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

// TODO
func (svc *kit) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"componentId":     svc.comp.Id,
		"componentName":   svc.comp.Name,
		"componentCommit": svc.comp.Commit,
		"accountId":       "392db620-1828-423f-a457-4c5680fb7787",
		"authToken":       "eyJhbGciOiJSUzI1NiIsImtpZCI6IlFZNDJITzZZcGl0c3kzTjQ0Wl9DWkxfR3R0TmJRMkE2SkhZeW9wU3NQcWcifQ.eyJhdWQiOlsiaHR0cHM6Ly9rdWJlcm5ldGVzLmRlZmF1bHQuc3ZjLmNsdXN0ZXIubG9jYWwiXSwiZXhwIjoxNjkzMzI2MDI4LCJpYXQiOjE2OTMzMjI0MjgsImlzcyI6Imh0dHBzOi8va3ViZXJuZXRlcy5kZWZhdWx0LnN2Yy5jbHVzdGVyLmxvY2FsIiwia3ViZXJuZXRlcy5pbyI6eyJuYW1lc3BhY2UiOiJrdWJlZm94LXN5c3RlbSIsInNlcnZpY2VhY2NvdW50Ijp7Im5hbWUiOiJtYWluLXZhdWx0IiwidWlkIjoiMzkyZGI2MjAtMTgyOC00MjNmLWE0NTctNGM1NjgwZmI3Nzg3In19LCJuYmYiOjE2OTMzMjI0MjgsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDprdWJlZm94LXN5c3RlbTptYWluLXZhdWx0In0.ufIurFiQPplxMrprERvU4QMNj7C7tmwZa52JDWCzV4Fz48C3VIV2NwIgs1ygmp_MSthWVOA53CCTEjaZB3VSDKd8yWZJXs-lx1-szuQoAS5y9BMPSA6vHxJBXaVLqw0dlwazJDDq-OvDxKjfIwiBiSq-3DvjZtMTDnvni4SC8ttWbXcDRRSrrOX9XdnzOEffmYnPQxkC8G9WEiPCa4BflKRB1ZvIX04ixzPou09U5qMwcBGvTX3kh4thkm3BVo6nsyLYeRM4HyjsjVKuEcYrzZcKh6jwSMviHkX7sO9Vill3TKxvx5HPvkgNjjctfk1N6eAwvTfn5wVZZegYcxhWFQ",
	}, nil
}

func (svc *kit) RequireTransportSecurity() bool {
	return false
}

func (svc *kit) sendEvent(evt *kubefox.Event, start int64) error {
	// need to protect the stream from being called by multiple threads
	svc.sendMutex.Lock()
	defer svc.sendMutex.Unlock()

	evt.Ttl = evt.Ttl - (time.Now().UnixMicro() - start)

	return svc.brk.Send(evt)
}
