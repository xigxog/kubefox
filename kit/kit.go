package kit

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/xigxog/kubefox/api"
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
	spec *api.ComponentDetails

	routes     []*route
	defHandler EventHandler

	brk        *grpc.Client
	numWorkers int

	export bool

	log *logkf.Logger
}

func New() Kit {
	svc := &kit{
		routes: make([]*route, 0),
		spec: &api.ComponentDetails{
			ComponentDefinition: api.ComponentDefinition{
				Type:         api.ComponentTypeKubeFox,
				Routes:       make([]api.RouteSpec, 0),
				EnvSchema:    make(map[string]*api.EnvVarSchema),
				Dependencies: make(map[string]*api.Dependency),
			},
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
	flag.IntVar(&svc.numWorkers, "num-workers", runtime.NumCPU(), "Number of worker threads to start, default is number of logical CPUs.")
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

	if !svc.export {
		utils.CheckRequiredFlag("name", comp.Name)
		utils.CheckRequiredFlag("commit", comp.Commit)

		if comp.Commit != build.Info.Commit {
			fmt.Fprintf(os.Stderr, "commit '%s' does not match build info commit '%s'", comp.Commit, build.Info.Commit)
			os.Exit(1)
		}
	} else {
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
		RouteSpec: api.RouteSpec{
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

	schema := &api.EnvVarSchema{}
	for _, o := range opts {
		o(schema)
	}
	if schema.Type == "" {
		schema.Type = api.EnvVarTypeString
	}
	svc.spec.EnvSchema[name] = schema

	return env.NewVar(name, schema.Type)
}

func (svc *kit) Component(name string) Dependency {
	return svc.dependency(name, api.ComponentTypeKubeFox)
}

func (svc *kit) HTTPAdapter(name string) Dependency {
	return svc.dependency(name, api.ComponentTypeHTTP)
}

func (svc *kit) dependency(name string, typ api.ComponentType) Dependency {
	c := &dependency{
		typ:  typ,
		name: name,
	}
	svc.spec.Dependencies[name] = &api.Dependency{Type: typ}

	return c
}

func (svc *kit) Start() {
	if svc.export {
		c, _ := json.MarshalIndent(svc.spec, "", "  ")
		fmt.Println(string(c))
		os.Exit(0)
	}

	var err error
	defer func() {
		if err != nil {
			logkf.Global.Sync()
			os.Exit(1)
		}
	}()

	svc.log.DebugInterface("component spec:", svc.spec)

	if err = svc.brk.StartHealthSrv(); err != nil {
		svc.log.Errorf("error starting health server: %v", err)
		return
	}

	go svc.brk.Start(&svc.spec.ComponentDefinition, maxAttempts)

	var wg sync.WaitGroup
	wg.Add(svc.numWorkers)
	svc.log.Infof("starting %d workers", svc.numWorkers)

	for i := 0; i < svc.numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case req := <-svc.brk.Req():
					svc.recvReq(req)
				case err = <-svc.brk.Err():
					svc.log.Errorf("broker error: %v", err)
					return
				}
			}
		}()
	}

	wg.Wait()
}

func (svc *kit) recvReq(req *grpc.ComponentEvent) {
	req.Event.ReduceTTL(req.ReceivedAt)

	ctx, cancel := context.WithTimeout(context.Background(), req.Event.TTL())
	defer cancel()

	log := svc.log.WithEvent(req.Event)

	ktx := &kontext{
		Event: req.Event,
		kit:   svc,
		env:   req.Env,
		start: time.Now(),
		ctx:   ctx,
		log:   log,
	}

	var err error
	switch {
	case req.RouteId == api.DefaultRouteId:
		if svc.defHandler == nil {
			err = kubefox.ErrNotFound(fmt.Errorf("default handler not found"))
		} else {
			err = svc.defHandler(ktx)
		}

	case req.RouteId >= 0 && req.RouteId < int64(len(svc.routes)):
		err = svc.routes[req.RouteId].handler(ktx)

	default:
		err = kubefox.ErrNotFound(fmt.Errorf("invalid route id %d", req.RouteId))
	}

	if err != nil {
		log.Error(err)

		errEvt := kubefox.NewErr(err, kubefox.EventOpts{})
		if err := ktx.ForwardResp(errEvt).Send(); err != nil {
			log.Error(err)
		}
	}
}
