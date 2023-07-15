package kubefox

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"

	"github.com/google/uuid"
	"github.com/xigxog/kubefox/libs/core/grpc"
	"github.com/xigxog/kubefox/libs/core/logger"
	"github.com/xigxog/kubefox/libs/core/platform"
	"github.com/xigxog/kubefox/libs/core/utils"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	ktyps "k8s.io/apimachinery/pkg/types"
)

const (
	ReqCtxKey ContextKey = "KitRequestContext"
)

var (
	argRegexp = regexp.MustCompile(`{\w*}`)
)

type KitContext interface {
	context.Context

	// Organization() string
	Platform() string
	PlatformNamespace() string
	CACertPath() string
	DevMode() bool

	Log() *logger.Log
}

type KitSvc interface {
	KitContext

	Start()
	OnStart(StartHandler)

	Http(string, Entrypoint)
	Kubernetes(string, Entrypoint)
	MatchEvent(string, Entrypoint)
	DefaultEntrypoint(Entrypoint)

	Fatal(err error)
}

type kitSvc struct {
	context.Context

	platform   string
	namespace  string
	brokerAddr string
	caCertPath string
	devMode    bool

	cfg *grpc.ComponentConfig

	startHandler StartHandler

	defEntrypoint Entrypoint
	entrypoints   []*EntrypointMatcher

	cancel context.CancelFunc
	log    *logger.Log
}

func New() KitSvc {
	svc := &kitSvc{cfg: &grpc.ComponentConfig{}}

	var help bool
	flag.StringVar(&svc.platform, "platform", "", "Platform instance component runs on; environment variable 'KUBEFOX_PLATFORM' (required)")
	flag.StringVar(&svc.namespace, "platform-namespace", "", "Namespace containing platform instance; environment variable 'KUBEFOX_PLATFORM_NAMESPACE' (required)")
	flag.StringVar(&svc.brokerAddr, "broker", "127.0.0.1:6060", "Address of the broker gRPC server; environment variable 'KUBEFOX_BROKER_ADDR' (default '127.0.0.1:6060')")
	flag.StringVar(&svc.caCertPath, "ca-cert-path", platform.CACertPath, "Path of file containing KubeFox root CA certificate; environment variable 'KUBEFOX_CA_CERT_PATH' (default '"+platform.CACertPath+"')")
	flag.BoolVar(&svc.devMode, "dev", false, "Run component in dev mode; environment variable 'KUBEFOX_DEV'")
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

	svc.platform = utils.ResolveFlag(svc.platform, "KUBEFOX_PLATFORM", "")
	svc.namespace = utils.ResolveFlag(svc.namespace, "KUBEFOX_PLATFORM_NAMESPACE", "")
	svc.brokerAddr = utils.ResolveFlag(svc.brokerAddr, "KUBEFOX_BROKER_ADDR", "127.0.0.1:6060")
	svc.caCertPath = utils.ResolveFlag(svc.caCertPath, "KUBEFOX_CA_CERT_PATH", platform.CACertPath)
	svc.devMode = utils.ResolveFlagBool(svc.devMode, "KUBEFOX_DEV", false)

	if svc.devMode {
		svc.log = logger.DevLogger()
		svc.log.Warn("dev mode enabled")
	} else {
		svc.log = logger.ProdLogger()
	}

	svc.Context, svc.cancel = context.WithTimeout(context.Background(), platform.StartupTimeout)
	svc.log.Info("kit service created 👍")

	return svc
}

func (svc *kitSvc) Fatal(err error) {
	svc.Log().Fatalf("component reported fatal error: %v", err)
}

// func (svc *kitSvc) Organization() string {
// 	return svc.cfg.Organization
// }

func (svc *kitSvc) Platform() string {
	return svc.platform
}

func (svc *kitSvc) PlatformNamespace() string {
	return svc.namespace
}

func (svc *kitSvc) CACertPath() string {
	return svc.caCertPath
}

func (svc *kitSvc) DevMode() bool {
	return svc.devMode
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

func (svc *kitSvc) OnStart(h StartHandler) {
	svc.startHandler = h
}

func (svc *kitSvc) Start() {
	defer svc.cancel()

	// creds, err := creds.NewClientTLSFromFile(svc.caCertFile, "")
	creds, err := platform.NewGRPCClientCreds(svc.caCertPath, ktyps.NamespacedName{
		Namespace: svc.namespace,
		Name:      fmt.Sprintf("%s-%s", svc.Platform(), platform.RootCASecret),
	})
	if err != nil {
		if svc.devMode {
			svc.log.Warnf("error reading certificate: %v", err)
			svc.log.Warn("dev mode enabled, using insecure connection")
			creds = insecure.NewCredentials()
		} else {
			svc.log.Errorf("error reading certificate: %v", err)
			os.Exit(RpcServerErrorCode)
		}
	}

	conn, err := gogrpc.Dial(svc.brokerAddr,
		gogrpc.WithTransportCredentials(creds),
		gogrpc.WithDefaultServiceConfig(platform.GRPCServiceCfg),
	)
	if err != nil {
		svc.log.Errorf("unable to connect to broker: %v", err)
		os.Exit(RpcServerErrorCode)

	}
	brk := grpc.NewComponentServiceClient(conn)
	svc.log.Debugf("connected to broker at %s", svc.brokerAddr)

	cfg, err := brk.GetConfig(svc, &grpc.ConfigRequest{})
	if err != nil {
		svc.log.Errorf("unable to get config from broker: %v", err)
		os.Exit(RpcServerErrorCode)
	}

	if cfg.Platform != svc.platform {
		svc.log.Errorf("broker belongs to an different platform instance")
		os.Exit(RpcServerErrorCode)
	}
	if cfg.DevMode && !svc.devMode {
		svc.log.Errorf("dev mode active on broker but not component")
		os.Exit(RpcServerErrorCode)
	} else if !cfg.DevMode && svc.devMode {
		svc.log.Errorf("dev mode active on component but not broker")
		os.Exit(RpcServerErrorCode)
	}

	if svc.devMode {
		svc.log.SugaredLogger = svc.log.
			Named(cfg.Component.Name).
			SugaredLogger
	} else {
		svc.log.SugaredLogger = svc.log.
			// WithOrganization(cfg.Organization).
			WithPlatform(cfg.Platform).
			WithComponent(cfg.Component).
			SugaredLogger
	}

	subId := uuid.NewString()
	stream, err := brk.Subscribe(context.Background(), &grpc.SubscribeRequest{Id: subId})
	if err != nil {
		svc.log.Fatal(err)
		return
	}
	svc.log.Infof("subscribed to broker; subscription: %s", subId)

	if svc.startHandler != nil {
		svc.log.Debug("invoking registered start handler")
		if err := svc.startHandler(svc); err != nil {
			svc.Fatal(err)
		}
	}

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
				Context: stream.Context(),
				req:     req,
				resp:    resp,
				kitSvc:  svc,
				log:     log,
			}
			kit.broker = kitBroker{kit: kit, broker: brk}

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
				log.DebugInterface(req, "request:")
				resp.SetError(compErr)
			}

			// FIXME should this be using background context? Need timeout from
			// stream context? Add a timeout to event so it can be passed around?
			brk.SendResponse(context.Background(), resp.GetData())
		}()
	}
}

func healthResponse(kit Kit) error {
	kit.Response().SetType(HealthResponseType)

	return nil
}
