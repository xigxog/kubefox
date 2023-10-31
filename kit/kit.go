package kit

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xigxog/kubefox/build"
	kubefox "github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/grpc"
	"github.com/xigxog/kubefox/logkf"
	"github.com/xigxog/kubefox/utils"

	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	maxRetry = 5
)

type kit struct {
	conf kubefox.ComponentConf

	routes     []*route
	defHandler EventHandler

	brk    grpc.Broker_SubscribeClient
	reqMap map[string]chan *kubefox.Event

	healthSrv *http.Server
	healthy   atomic.Bool

	reqMapMutex sync.Mutex
	sendMutex   sync.Mutex

	brokerAddr string

	export bool

	log *logkf.Logger
}

type route struct {
	kubefox.Route

	handler EventHandler
}

type envVar struct {
	name string
	typ  EnvVarType
}

type EnvVarSchema struct {
	Type        EnvVarType
	Required    bool
	Title       string
	Description string
}

func New() Kit {
	_, id := kubefox.GenerateNameAndId()
	svc := &kit{
		routes: make([]*route, 0),
		conf: kubefox.ComponentConf{
			Component: &kubefox.Component{
				Id: id,
			},
			Routes:       make([]*kubefox.Route, 0),
			EnvSchema:    make(map[string]kubefox.EnvVarSchema),
			Dependencies: make(map[string]kubefox.Dependency),
		},
		reqMap: make(map[string]chan *kubefox.Event),
	}

	var help bool
	var healthAddr, logFormat, logLevel string
	//-tls-skip-verify
	flag.StringVar(&svc.conf.Component.Name, "name", "", "Component name; environment variable 'KUBEFOX_COMPONENT'. (required)")
	flag.StringVar(&svc.conf.Component.Commit, "commit", "", "Commit the Component was built from; environment variable 'KUBEFOX_COMPONENT_COMMIT'. (required)")
	flag.StringVar(&svc.brokerAddr, "broker-addr", "127.0.0.1:6060", "Address of the Broker gRPC server; environment variable 'KUBEFOX_BROKER_ADDR'.")
	flag.StringVar(&healthAddr, "health-addr", "127.0.0.1:1111", `Address and port the HTTP health server should bind to, set to "false" to disable; environment variable 'KUBEFOX_HEALTH_ADDR'.`)
	flag.StringVar(&logFormat, "log-format", "console", "Log format; environment variable 'KUBEFOX_LOG_FORMAT'. [options 'json', 'console']")
	flag.StringVar(&logLevel, "log-level", "debug", "Log level; environment variable 'KUBEFOX_LOG_LEVEL'. [options 'debug', 'info', 'warn', 'error']")
	flag.BoolVar(&svc.export, "export", false, "Exports component configuration in JSON and exits.")
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

	svc.conf.Component.Name = utils.ResolveFlag(svc.conf.Component.Name, "KUBEFOX_COMPONENT", "")
	svc.conf.Component.Commit = utils.ResolveFlag(svc.conf.Component.Commit, "KUBEFOX_COMPONENT_COMMIT", "")
	svc.brokerAddr = utils.ResolveFlag(svc.brokerAddr, "KUBEFOX_BROKER_ADDR", "127.0.0.1:6060")
	healthAddr = utils.ResolveFlag(healthAddr, "KUBEFOX_HEALTH_ADDR", "127.0.0.1:1111")
	logFormat = utils.ResolveFlag(logFormat, "KUBEFOX_LOG_FORMAT", "console")
	logLevel = utils.ResolveFlag(logLevel, "KUBEFOX_LOG_LEVEL", "debug")

	utils.CheckRequiredFlag("name", svc.conf.Component.Name)
	utils.CheckRequiredFlag("commit", svc.conf.Component.Commit)

	if svc.export {
		logLevel = "error"
	}

	logkf.Global = logkf.
		BuildLoggerOrDie(logFormat, logLevel).
		WithComponent(svc.conf.Component)
	defer logkf.Global.Sync()

	svc.log = logkf.Global
	svc.log.DebugInterface("build info:", build.Info)

	if !svc.export {
		svc.startHealthSrv(healthAddr)
	}

	svc.log.Info("🦊 kit created")

	return svc
}

func (svc *kit) Log() *logkf.Logger {
	return svc.log
}

func (svc *kit) Title(title string) {
	svc.conf.Title = title
}

func (svc *kit) Description(description string) {
	svc.conf.Title = description
}

func (svc *kit) Route(rule string, handler EventHandler) {
	r := &route{
		Route: kubefox.Route{
			Id:   len(svc.routes),
			Rule: rule,
		},
		handler: handler,
	}
	svc.routes = append(svc.routes, r)
	svc.conf.Routes = append(svc.conf.Routes, &r.Route)
}

func (svc *kit) Default(handler EventHandler) {
	svc.defHandler = handler
	svc.conf.DefaultHandler = handler != nil
}

func (svc *kit) EnvVar(name string, opts ...EnvVarOption) EnvVar {
	if name == "" {
		svc.log.Fatal("environment variable name is required")
	}

	schema := kubefox.EnvVarSchema{Name: name}
	for _, o := range opts {
		o(&schema)
	}
	if schema.Type == "" {
		schema.Type = kubefox.EnvVarTypeString
	}
	svc.conf.EnvSchema[name] = schema

	return &envVar{
		name: name,
		typ:  EnvVarType(schema.Type),
	}
}

func (svc *kit) Component(name string) Dependency {
	c := kubefox.Dependency{
		Name: name,
		Type: kubefox.ComponentTypeKubeFoxComponent,
	}
	svc.conf.Dependencies[name] = c

	return &c
}

func (svc *kit) HTTPAdapter(name string) Dependency {
	c := kubefox.Dependency{
		Name: name,
		Type: kubefox.ComponentTypeHTTPAdapter,
	}
	svc.conf.Dependencies[name] = c

	return &c
}

func (svc *kit) Start() {
	if svc.export {
		c, _ := json.MarshalIndent(svc.conf, "", "  ")
		fmt.Println(string(c))
		os.Exit(0)
	}

	svc.log.DebugInterface("configuration:", svc.conf)

	var (
		retry int
		err   error
	)
	for retry < maxRetry {
		retry, err = svc.run(retry)
		svc.healthy.Store(false)
		svc.log.Warnf("broker subscription closed, retry %d: %v", retry, err)
		time.Sleep(time.Second * time.Duration(rand.Intn(2)+1))
	}
	svc.log.Fatalf("exceeded max retries connection to broker: %v", err)
}

func (svc *kit) startHealthSrv(addr string) {
	svc.healthSrv = &http.Server{
		WriteTimeout: time.Second * 3,
		ReadTimeout:  time.Second * 3,
		IdleTimeout:  time.Second * 30,
		Handler:      svc,
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		svc.log.Fatal("unable to open tcp socket for health server: %v", err)
	}
	go func() {
		err := svc.healthSrv.Serve(ln)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			svc.log.Error(err)
			os.Exit(1)
		}
	}()
	svc.log.Debug("health server started")
}

func (svc *kit) run(retry int) (int, error) {
	creds, err := credentials.NewClientTLSFromFile(kubefox.PathCACert, "")
	if err != nil {
		svc.log.Fatalf("unable to load root CA certificate: %v", err)
	}
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
		gogrpc.WithTransportCredentials(creds),
		gogrpc.WithDefaultServiceConfig(grpcCfg),
	)
	if err != nil {
		return retry + 1, fmt.Errorf("unable to connect to broker: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			svc.log.Fatal(err)
		}
	}()

	if svc.brk, err = grpc.NewBrokerClient(conn).Subscribe(context.Background()); err != nil {
		return retry + 1, fmt.Errorf("subscribing to broker failed: %v", err)
	}

	regEvt := kubefox.NewMsg(kubefox.EventOpts{
		Type:   kubefox.EventTypeRegister,
		Source: svc.conf.Component,
	})
	if err := regEvt.SetJSON(svc.conf); err != nil {
		svc.log.Fatalf("unable to marshal registration: %v", err)
	}
	if err := svc.sendEvent(regEvt); err != nil {
		return retry + 1, fmt.Errorf("unable to register with broker: %v", err)
	}

	svc.healthy.Store(true)
	svc.log.Info("kit subscribed to broker")

	for {
		mEvt, err := svc.brk.Recv()
		if err != nil {
			// Success connection was made, reset retry.
			return 0, err
		}

		svc.log.WithEvent(mEvt.Event).Debug("received event")

		switch mEvt.Event.Category {
		case kubefox.Category_REQUEST:
			go svc.recvReq(mEvt)

		case kubefox.Category_RESPONSE:
			go svc.recvResp(mEvt.Event)

		default:
			svc.log.Debug("default")
		}
	}
}

func (svc *kit) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	status := http.StatusOK
	if !svc.healthy.Load() {
		status = http.StatusServiceUnavailable
	}
	resp.WriteHeader(status)
}

func (svc *kit) recvReq(req *kubefox.MatchedEvent) {
	log := svc.log.WithEvent(req.Event)
	log.Debug("receive request")

	ctx, cancel := context.WithTimeout(context.Background(), req.Event.TTL())
	defer cancel()

	ktx := &kontext{
		Event: req.Event,
		resp: kubefox.NewResp(kubefox.EventOpts{
			Parent: req.Event,
			Source: svc.conf.Component,
			Target: req.Event.Source,
		}),
		kit: svc,
		env: req.Env,
		ctx: ctx,
		log: log,
	}

	var err error
	switch {
	case req.RouteId == kubefox.DefaultRouteId:
		if svc.defHandler == nil {
			err = fmt.Errorf("default handler not found")
		} else {
			err = svc.defHandler(ktx)
		}

	case req.RouteId >= 0 && req.RouteId < int64(len(svc.routes)):
		err = svc.routes[req.RouteId].handler(ktx)

	default:
		err = fmt.Errorf("invalid route id %d", req.RouteId)
	}

	if err != nil {
		ktx.resp.Type = string(kubefox.EventTypeError)

		log.Error(err)
		if err := ktx.Resp().SendStr(err.Error()); err != nil {
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

	defer func() {
		svc.reqMapMutex.Lock()
		delete(svc.reqMap, req.Id)
		svc.reqMapMutex.Unlock()
	}()

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
	log := svc.log.WithEvent(resp)
	log.Debug("receive response")

	svc.reqMapMutex.Lock()
	respCh, found := svc.reqMap[resp.ParentId]
	svc.reqMapMutex.Unlock()

	if !found {
		log.Error("request for response not found")
		return
	}

	respCh <- resp
}

func (svc *kit) sendEvent(evt *kubefox.Event) error {
	// Need to protect the stream from being called by multiple threads.
	svc.sendMutex.Lock()
	defer svc.sendMutex.Unlock()

	svc.log.WithEvent(evt).Debug("send event")

	return svc.brk.Send(evt)
}

func (svc *kit) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	b, err := os.ReadFile(kubefox.PathSvcAccToken)
	if err != nil {
		return nil, err
	}
	token := string(b)

	return map[string]string{
		"componentId":     svc.conf.Component.Id,
		"componentName":   svc.conf.Component.Name,
		"componentCommit": svc.conf.Component.Commit,
		"authToken":       token,
	}, nil
}

func (svc *kit) RequireTransportSecurity() bool {
	return true
}

func (v *envVar) GetName() string {
	return v.name
}

func (v *envVar) GetType() EnvVarType {
	return v.typ
}

func Type(typ EnvVarType) EnvVarOption {
	return func(evs *kubefox.EnvVarSchema) {
		evs.Type = kubefox.EnvVarType(typ)
	}
}

func Required() EnvVarOption {
	return func(evs *kubefox.EnvVarSchema) {
		evs.Required = true
	}
}

func Title(title string) EnvVarOption {
	return func(evs *kubefox.EnvVarSchema) {
		evs.Title = title
	}
}

func Description(description string) EnvVarOption {
	return func(evs *kubefox.EnvVarSchema) {
		evs.Description = description
	}
}
