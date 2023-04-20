package kubefox

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/xigxog/kubefox/libs/core/grpc"
	"github.com/xigxog/kubefox/libs/core/logger"
	"github.com/xigxog/kubefox/libs/core/platform"
	"github.com/xigxog/kubefox/libs/core/utils"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	KitContextKey ContextKey = "KitContext"
)

var (
	argRegexp     = regexp.MustCompile(`{\w*}`)
	brokerTimeout = 5 * time.Second
)

type KitSvc interface {
	Start()
	Http(string, Entrypoint)
	Kubernetes(string, Entrypoint)
	MatchEvent(string, Entrypoint)
	DefaultEntrypoint(Entrypoint)

	// Organization() string
	Platform() string
	Namespace() string
	DevMode() bool

	Fatal(err error)
	Log() *logger.Log
}

type kitSvc struct {
	brk grpc.ComponentServiceClient

	cfg       *grpc.ComponentConfig
	namespace string

	defEntrypoint Entrypoint
	entrypoints   []*EntrypointMatcher

	log *logger.Log
}

func New() KitSvc {
	var namespace, brokerAddr string
	var devMode, help bool
	flag.StringVar(&namespace, "namespace", "", "Kubernetes namespace of KubeFox Platform. Environment variable 'KUBEFOX_NAMESPACE' (default 'kubefox-system')")
	flag.StringVar(&brokerAddr, "broker", "127.0.0.1:7070", "Address of the broker's gRPC server. Environment variable 'KUBEFOX_BROKER_ADDR' (default '127.0.0.1:7070')")
	flag.BoolVar(&devMode, "dev", false, "Run component in dev mode. Environment variable 'KUBEFOX_DEV'")
	flag.BoolVar(&help, "help", false, "Show usage for component")
	flag.Parse()

	if help {
		fmt.Fprintf(flag.CommandLine.Output(), `
Flags can be set using names below or the environment variable listed.

Flags:
`)
		flag.PrintDefaults()
		os.Exit(0)
	}

	namespace = utils.ResolveFlag(namespace, "KUBEFOX_NAMESPACE", "kubefox-system")
	brokerAddr = utils.ResolveFlag(brokerAddr, "KUBEFOX_BROKER_ADDR", "127.0.0.1:7070")
	devMode = utils.ResolveFlagBool(devMode, "KUBEFOX_DEV", false)

	var log *logger.Log
	if devMode {
		log = logger.DevLogger()
		log.Warn("dev mode enabled")
	} else {
		log = logger.ProdLogger()
	}

	creds, err := platform.NewGRPCClientCreds(namespace)
	if err != nil {
		if devMode {
			log.Warnf("error reading certificate: %v", err)
			log.Warn("dev mode enabled, using insecure connection")
			creds = insecure.NewCredentials()
		} else {
			log.Errorf("error reading certificate: %v", err)
			os.Exit(RpcServerErrorCode)
		}
	}

	conn, err := gogrpc.Dial(brokerAddr, gogrpc.WithTransportCredentials(creds))
	if err != nil {
		log.Errorf("unable to connect to broker: %v", err)
		os.Exit(RpcServerErrorCode)

	}
	broker := grpc.NewComponentServiceClient(conn)
	log.Debugf("connected to broker at %s", brokerAddr)

	ctx, cancel := context.WithTimeout(context.Background(), brokerTimeout)
	defer cancel()

	cfg, err := broker.GetConfig(ctx, &grpc.ConfigRequest{})
	if err != nil {
		log.Errorf("unable to get config from broker: %v", err)
		os.Exit(RpcServerErrorCode)
	}

	if cfg.DevMode && !devMode {
		log.Errorf("dev mode active on broker but not component")
		os.Exit(RpcServerErrorCode)
	} else if !cfg.DevMode && devMode {
		log.Errorf("dev mode active on component but not broker")
		os.Exit(RpcServerErrorCode)
	}

	if devMode {
		log = log.Named(cfg.Component.Name)
	} else {
		log = log.
			// WithOrganization(cfg.Organization).
			WithPlatform(cfg.Platform).
			WithComponent(cfg.Component)
	}

	log.Info("kit service created 👍")
	log.DebugInterface(cfg, "component config:")

	return &kitSvc{
		brk:       broker,
		namespace: namespace,
		cfg:       cfg,
		log:       log,
	}
}

func (svc *kitSvc) Fatal(err error) {
	svc.Log().Fatalf("component reported fatal error: %v", err)
}

// func (svc *kitSvc) Organization() string {
// 	return svc.cfg.Organization
// }

func (svc *kitSvc) Platform() string {
	return svc.cfg.Platform
}

func (svc *kitSvc) Namespace() string {
	return svc.namespace
}

func (svc *kitSvc) DevMode() bool {
	return svc.cfg.DevMode
}

func (svc *kitSvc) Log() *logger.Log {
	return svc.log
}

func (svc *kitSvc) Http(rule string, entrypoint Entrypoint) {
	svc.MatchEvent(fmt.Sprintf("Type(`%s`) && %s", HTTPRequestType, rule), entrypoint)
}

func (svc *kitSvc) Kubernetes(rule string, entrypoint Entrypoint) {
	svc.MatchEvent(fmt.Sprintf("Type(`%s`) && %s", KubernetesRequestType, rule), entrypoint)
}

func (svc *kitSvc) MatchEvent(rule string, entrypoint Entrypoint) {
	matcher, err := NewEventMatcher(rule)
	if err != nil {
		svc.log.Fatal(err)
	}

	svc.entrypoints = append(svc.entrypoints, &EntrypointMatcher{
		entrypoint: entrypoint,
		matcher:    matcher,
	})

	// Sort rules, longest (most specific) rule should be tested first.
	sort.SliceStable(svc.entrypoints, func(i, j int) bool {
		// Normalize path args so they don't affect length.
		l := argRegexp.ReplaceAllString(svc.entrypoints[j].Rule(), "{}")
		r := argRegexp.ReplaceAllString(svc.entrypoints[i].Rule(), "{}")

		return len(l) < len(r)
	})
}

func (svc *kitSvc) DefaultEntrypoint(entrypoint Entrypoint) {
	svc.defEntrypoint = entrypoint
}

func (svc *kitSvc) Start() {
	subId := uuid.NewString()
	stream, err := svc.brk.Subscribe(context.Background(), &grpc.SubscribeRequest{Id: subId})
	if err != nil {
		svc.log.Fatal(err)
		return
	}
	svc.log.Infof("subscribed to broker; subscription: %s", subId)

	for {
		reqData, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			svc.log.Fatal(err)
			return
		}

		go func() {
			req := EventFromData(reqData)
			resp := req.ChildEvent()

			log := svc.log
			if req.GetTraceId() != "" {
				log = svc.log.With("traceId", req.GetTraceId())
			}

			kit := &kit{
				req:    req,
				resp:   resp,
				kitSvc: svc,
				ctx:    stream.Context(),
				log:    log,
			}
			kit.broker = kitBroker{kit: kit, broker: svc.brk}

			var entrypoint Entrypoint
			switch reqData.Type {
			case HealthRequestType:
				entrypoint = healthResponse

			default:
				// TODO move matching to broker
				for _, epm := range svc.entrypoints {
					if epm.Match(kit.req) {
						entrypoint = epm.entrypoint
						break
					}
				}
				if entrypoint == nil {
					entrypoint = svc.defEntrypoint
				}
			}

			var compErr error
			if entrypoint == nil {
				compErr = ErrEntrypointNotFound
			} else {
				compErr = entrypoint(kit)
			}

			if compErr != nil {
				log.Errorf("error calling entrypoint: %v", compErr)
				resp.SetError(compErr)
			}

			// FIXME should this be using background context? Need timeout from
			// stream context? Add a timeout to event so it can be passed around?
			svc.brk.SendResponse(context.Background(), resp.GetData())
		}()
	}
}

func healthResponse(kit Kit) error {
	kit.Response().SetType(HealthResponseType)

	return nil
}
