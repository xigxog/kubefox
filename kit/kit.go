package kit

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/xigxog/kubefox/build"
	kubefox "github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/grpc"
	"github.com/xigxog/kubefox/kit/env"
	"github.com/xigxog/kubefox/logkf"
	"github.com/xigxog/kubefox/utils"
)

const (
	maxAttempts = 5
)

type kit struct {
	spec *kubefox.ComponentSpec

	routes     []*route
	defHandler EventHandler

	brk *grpc.Client

	export bool

	log *logkf.Logger
}

func New() Kit {
	svc := &kit{
		routes: make([]*route, 0),
		spec: &kubefox.ComponentSpec{
			ComponentTypeVar: kubefox.ComponentTypeVar{
				Type: kubefox.ComponentTypeKubeFox,
			},
			Routes:       make([]kubefox.RouteSpec, 0),
			EnvSchema:    make(map[string]*kubefox.EnvVarSchema),
			Dependencies: make(map[string]*kubefox.ComponentTypeVar),
		},
	}

	comp := &kubefox.Component{Id: kubefox.GenerateId()}

	var help bool
	var brokerAddr, healthAddr, logFormat, logLevel string
	//-tls-skip-verify
	flag.StringVar(&comp.Name, "name", "", "Component name; environment variable 'KUBEFOX_COMPONENT'. (required)")
	flag.StringVar(&comp.Commit, "commit", "", "Commit the Component was built from; environment variable 'KUBEFOX_COMPONENT_COMMIT'. (required)")
	flag.StringVar(&brokerAddr, "broker-addr", "127.0.0.1:6060", "Address of the Broker gRPC server; environment variable 'KUBEFOX_BROKER_ADDR'.")
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

	comp.Name = utils.ResolveFlag(comp.Name, "KUBEFOX_COMPONENT", "")
	comp.Commit = utils.ResolveFlag(comp.Commit, "KUBEFOX_COMPONENT_COMMIT", "")
	brokerAddr = utils.ResolveFlag(brokerAddr, "KUBEFOX_BROKER_ADDR", "127.0.0.1:6060")
	healthAddr = utils.ResolveFlag(healthAddr, "KUBEFOX_HEALTH_ADDR", "127.0.0.1:1111")
	logFormat = utils.ResolveFlag(logFormat, "KUBEFOX_LOG_FORMAT", "console")
	logLevel = utils.ResolveFlag(logLevel, "KUBEFOX_LOG_LEVEL", "debug")

	utils.CheckRequiredFlag("name", comp.Name)
	utils.CheckRequiredFlag("commit", comp.Commit)

	if comp.Commit != build.Info.Commit {
		fmt.Fprintf(os.Stderr, "commit '%s' does not match build info commit '%s'", comp.Commit, build.Info.Commit)
		os.Exit(1)
	}

	if svc.export {
		logLevel = "error"
	}

	logkf.Global = logkf.
		BuildLoggerOrDie(logFormat, logLevel).
		WithComponent(comp)
	defer logkf.Global.Sync()

	svc.log = logkf.Global
	svc.log.DebugInterface("build info:", build.Info)

	svc.brk = grpc.NewClient(grpc.ClientOpts{
		Component:     comp,
		BrokerAddr:    brokerAddr,
		HealthSrvAddr: healthAddr,
	})

	svc.log.Info("🦊 kit created")

	return svc
}

func (svc *kit) Log() *logkf.Logger {
	return svc.log
}

func (svc *kit) Title(title string) {
	svc.spec.Title = title
}

func (svc *kit) Description(description string) {
	svc.spec.Title = description
}

func (svc *kit) Route(rule string, handler EventHandler) {
	r := &route{
		RouteSpec: kubefox.RouteSpec{
			Id:   len(svc.routes),
			Rule: rule,
		},
		handler: handler,
	}
	svc.routes = append(svc.routes, r)
	svc.spec.Routes = append(svc.spec.Routes, r.RouteSpec)
}

func (svc *kit) Default(handler EventHandler) {
	svc.defHandler = handler
	svc.spec.DefaultHandler = handler != nil
}

func (svc *kit) EnvVar(name string, opts ...env.VarOption) EnvVar {
	if name == "" {
		svc.log.Fatal("environment variable name is required")
	}

	schema := &kubefox.EnvVarSchema{}
	for _, o := range opts {
		o(schema)
	}
	if schema.Type == "" {
		schema.Type = kubefox.EnvVarTypeString
	}
	svc.spec.EnvSchema[name] = schema

	return &env.Var{
		Name: name,
		Type: schema.Type,
	}
}

func (svc *kit) Component(name string) Dependency {
	return svc.dependency(name, kubefox.ComponentTypeKubeFox)
}

func (svc *kit) HTTPAdapter(name string) Dependency {
	return svc.dependency(name, kubefox.ComponentTypeHTTP)
}

func (svc *kit) dependency(name string, typ kubefox.ComponentType) Dependency {
	c := &dependency{
		ComponentTypeVar: kubefox.ComponentTypeVar{
			Type: typ,
		},
		Name: name,
	}
	svc.spec.Dependencies[name] = &c.ComponentTypeVar

	return c
}

func (svc *kit) Start() {
	if svc.export {
		c, _ := json.MarshalIndent(svc.spec, "", "  ")
		fmt.Println(string(c))
		os.Exit(0)
	}

	svc.log.DebugInterface("component spec:", svc.spec)

	if err := svc.brk.StartHealthSrv(); err != nil {
		svc.log.Fatalf("error starting health server: %v", err)
	}

	go svc.brk.Start(svc.spec, maxAttempts)

	for {
		select {
		case req := <-svc.brk.Req():
			svc.recvReq(req)
		case err := <-svc.brk.Err():
			svc.log.Fatalf("broker unavailable: %v", err)
		}
	}
}

func (svc *kit) recvReq(req *kubefox.MatchedEvent) {
	ctx, cancel := context.WithTimeout(context.Background(), req.Event.TTL())
	defer cancel()

	log := svc.log.WithEvent(req.Event)

	ktx := &kontext{
		Event: req.Event,
		resp: kubefox.NewResp(kubefox.EventOpts{
			Parent: req.Event,
			Source: svc.brk.Component,
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
