package engine

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/xigxog/kubefox/components/broker/blocker"
	"github.com/xigxog/kubefox/components/broker/config"
	"github.com/xigxog/kubefox/components/broker/fabric"
	"github.com/xigxog/kubefox/components/broker/jetstream"
	"github.com/xigxog/kubefox/components/broker/telemetry"
	"github.com/xigxog/kubefox/libs/core/component"
	"github.com/xigxog/kubefox/libs/core/grpc"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/logger"
	"github.com/xigxog/kubefox/libs/core/platform"
	"github.com/xigxog/kubefox/libs/core/utils"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	shutdownTimeout = 30 * time.Second

	platformEvtErr    = fmt.Errorf("event is a platform event but local component is not runtime server")
	localNotTargetErr = fmt.Errorf("local component is not the event target")
	outOfCtxErr       = fmt.Errorf("local component does not exist in the event context")
)

type Broker interface {
	Config() *config.Config
	Component() component.Component
	EventTimeout() time.Duration
	ConnectTimeout() time.Duration

	InvokeLocalComponent(context.Context, kubefox.DataEvent) kubefox.DataEvent
	InvokeRemoteComponent(context.Context, kubefox.DataEvent) kubefox.DataEvent
	InvokeRuntimeServer(context.Context, kubefox.DataEvent) kubefox.DataEvent

	StartComponent()
	StartRuntimeSrv()
	StartHTTPClient()
	StartHTTPSrv()
	Shutdown()

	JetStreamClient() *jetstream.Client
	Blocker() *blocker.Blocker

	IsHealthy(context.Context) bool
	Log() *logger.Log
}

type EventSender interface {
	SendEvent(ctx context.Context, req kubefox.DataEvent) kubefox.DataEvent
}

type broker struct {
	cfg *config.Config

	remoteSender EventSender
	localSender  EventSender

	runtimeClient EventSender

	grpcSrv     *GRPCServer
	runtimeSrv  *RuntimeServer
	httpClient  *HTTPClient
	httpSrv     *HTTPServer
	jetSender   *JetStreamSender
	jetListener *JetStreamListener

	telClient *telemetry.Client
	healthSrv *HealthServer

	jetClient *jetstream.Client
	fabStore  *fabric.Store

	blocker *blocker.Blocker

	log *logger.Log
}

func New(flags config.Flags) *broker {
	comp := component.New(component.Fields{
		Name:    flags.CompName,
		GitHash: flags.CompGitHash,
		Id:      genId(),
	})

	var log *logger.Log
	if flags.IsDevMode {
		log = logger.DevLogger().Named(comp.GetName())
		log.Warn("dev mode enabled")
	} else {
		log = logger.ProdLogger().WithComponent(comp)
	}

	log.Info("broker starting")
	log.DebugInterface(flags, "config:")

	if flags.Namespace == "" {
		flags.Namespace = utils.SystemNamespace(flags.Platform, flags.System)
	}

	return &broker{
		cfg: &config.Config{
			Flags: flags,
			Comp:  comp,
		},
		blocker: blocker.NewBlocker(log),
		log:     log,
	}
}

func (brk *broker) InvokeLocalComponent(ctx context.Context, req kubefox.DataEvent) (resp kubefox.DataEvent) {
	ctx, span := telemetry.NewSpan(ctx, brk.EventTimeout(), req)
	defer span.End(resp)

	brk.Log().Debugf("invoking local component; evtType: %s, traceId: %s", req.GetType(), req.GetSpan().TraceId)

	if err := brk.checkEvent(req); err != nil {
		resp = brk.invokeErr(req, err)
		return
	}
	if kubefox.IsPlatformEvent(req) {
		if !brk.Config().IsRuntimeSrv {
			resp = brk.invokeErr(req, platformEvtErr)
			return
		}

		if !component.Equal(platform.RuntimeSrvComp, req.GetTarget()) {
			resp = brk.invokeErr(req, localNotTargetErr)
			return
		}

	} else {
		fab, err := brk.fabStore.Get(ctx, req)
		if err != nil {
			resp = brk.invokeErr(req, localNotTargetErr)
			return
		}

		fabSrcComp := fab.System.App.Components[brk.Component().GetName()]
		if fabSrcComp == nil || fabSrcComp.GitHash != brk.Component().GetGitHash() {
			resp = brk.invokeErr(req, outOfCtxErr)
			return
		}

		if !component.Equal(brk.Component(), req.GetTarget()) {
			resp = brk.invokeErr(req, localNotTargetErr)
			return
		}

		gFab := &grpc.Fabric{
			Config:  map[string]*structpb.Value{},
			Secrets: map[string]*structpb.Value{},
			EnvVars: map[string]*structpb.Value{},
		}
		for k, v := range fab.Env.Config {
			gFab.Config[k] = v.Value()
		}
		for k, v := range fab.Env.Secrets {
			gFab.Secrets[k] = v.Value()
		}
		for k, v := range fab.Env.EnvVars {
			gFab.EnvVars[k] = v.Value()
		}
		req.SetFabric(gFab)
	}

	resp = brk.localSender.SendEvent(ctx, req)
	resp.SetFabric(nil)
	resp.SetSource(brk.Component(), req.GetContext().App)
	resp.SetTarget(req.GetSource())
	if resp.GetType() == kubefox.ErrorEventType {
		if resp.GetError() == nil {
			resp.SetError(errors.New(resp.GetErrorMsg()))
		}
		brk.Log().Error(resp.GetError())
		return
	}

	if resp.GetType() == "" || resp.GetType() == kubefox.UnknownEventType {
		switch req.GetType() {
		case kubefox.BootstrapRequestType:
			resp.SetType(kubefox.BootstrapResponseType)

		case kubefox.ComponentRequestType:
			resp.SetType(kubefox.ComponentResponseType)

		case kubefox.CronRequestType:
			resp.SetType(kubefox.CronResponseType)

		case kubefox.DaprRequestType:
			resp.SetType(kubefox.DaprResponseType)

		case kubefox.FabricRequestType:
			resp.SetType(kubefox.FabricResponseType)

		case kubefox.HealthRequestType:
			resp.SetType(kubefox.HealthResponseType)

		case kubefox.HTTPRequestType:
			resp.SetType(kubefox.HTTPResponseType)

		case kubefox.KubernetesRequestType:
			resp.SetType(kubefox.KubernetesResponseType)

		case kubefox.MetricsRequestType:
			resp.SetType(kubefox.TelemetryResponseType)

		case kubefox.TelemetryRequestType:
			resp.SetType(kubefox.TelemetryResponseType)
		}
	}

	return resp
}

func (brk *broker) InvokeRemoteComponent(ctx context.Context, req kubefox.DataEvent) (resp kubefox.DataEvent) {
	ctx, span := telemetry.NewSpan(ctx, brk.EventTimeout(), req)
	defer span.End(resp)

	if req.GetData() != nil && req.GetContext() != nil {
		brk.Log().Debugf("invoking remote component; target: %s, evtType: %s, traceId: %s",
			req.GetTarget(), req.GetType(), req.GetSpan().TraceId)

		req.SetSource(brk.Component(), req.GetContext().App)
		// req.GetContext().Organization = brk.Config().Organization
		req.GetContext().Platform = brk.Config().Platform
	}

	if err := brk.checkEvent(req); err != nil {
		resp = brk.invokeErr(req, err)
		return
	}

	fab, err := brk.fabStore.Get(ctx, req)
	if err != nil {
		resp = brk.invokeErr(req, err)
		return
	}

	fabTgtComp := fab.GetAppComponent(req.GetTarget().GetName())
	if fabTgtComp == nil {
		resp = brk.invokeErr(req, outOfCtxErr)
		return
	}
	req.GetTarget().SetGitHash(fabTgtComp.GitHash)

	fabSrcComp := fab.GetAppComponent(req.GetSource().GetName())
	if fabSrcComp == nil || fabSrcComp.GitHash != req.GetSource().GetGitHash() {
		resp = brk.invokeErr(req, outOfCtxErr)
		return
	}

	// ensure fabric not sent to target
	req.SetFabric(nil)

	resp = brk.remoteSender.SendEvent(ctx, req)
	resp.SetFabric(nil)
	if resp.GetType() == kubefox.ErrorEventType {
		if resp.GetError() == nil {
			resp.SetError(errors.New(resp.GetErrorMsg()))
		}
		brk.Log().Error(resp.GetError())
		return
	}

	return
}

func (brk *broker) InvokeRuntimeServer(ctx context.Context, req kubefox.DataEvent) (resp kubefox.DataEvent) {
	ctx, span := telemetry.NewSpan(ctx, brk.EventTimeout(), req)
	defer span.End(resp)

	brk.Log().Debugf("invoking runtime server; evtType: %s", req.GetType())

	req.SetSource(brk.Component(), req.GetContext().App)
	req.SetTarget(platform.RuntimeSrvComp)

	// TODO cache token and only refresh when needed
	token, err := utils.GetSvcAccountToken(brk.Config().Namespace, platform.BrokerSvcAccount)
	if err != nil {
		resp = brk.invokeErr(req, err)
		return
	}
	req.SetArg(platform.SvcAccountTokenArg, token)

	if err := brk.checkEvent(req); err != nil {
		resp = brk.invokeErr(req, err)
		return
	}

	resp = brk.runtimeClient.SendEvent(ctx, req)
	resp.SetFabric(nil)
	if resp.GetType() == kubefox.ErrorEventType {
		if resp.GetError() == nil {
			resp.SetError(errors.New(resp.GetErrorMsg()))
		}
		brk.Log().Error(resp.GetError())
		return
	}

	return
}

func (brk *broker) invokeErr(req kubefox.DataEvent, err error) kubefox.DataEvent {
	resp := req.ChildErrorEvent(err)
	resp.SetSource(brk.Component(), req.GetContext().App)
	resp.SetTarget(req.GetSource())

	brk.Log().Error(resp.GetError())

	return resp
}

// TODO switch to use validation
func (brk *broker) checkEvent(req kubefox.DataEvent) error {
	if req == nil {
		return fmt.Errorf("event empty")
	}
	if req.GetData() == nil {
		return fmt.Errorf("event data empty")
	}
	if req.GetType() == kubefox.ErrorEventType {
		return req.GetError()
	}
	if req.GetType() == "" {
		return fmt.Errorf("event type missing")
	}
	if req.GetId() == "" {
		return fmt.Errorf("event id missing")
	}

	if req.GetContext() == nil {
		return fmt.Errorf("event context missing")
	}
	// if req.GetContext().GetOrganization() != brk.Config().Organization {
	// 	return fmt.Errorf("event organization context invalid")
	// }
	if req.GetContext().GetPlatform() != brk.Config().Platform {
		return fmt.Errorf("event platform context invalid")
	}
	if req.GetContext().GetEnvironment() == "" {
		return fmt.Errorf("event environment context missing")
	}

	if req.GetSource() == nil {
		return fmt.Errorf("event source missing")
	}
	if req.GetSource().GetName() == "" {
		return fmt.Errorf("event source name missing")
	}
	if req.GetSource().GetApp() == "" {
		return fmt.Errorf("event source app missing")
	}
	if req.GetSource().GetGitHash() == "" {
		return fmt.Errorf("event source git hash missing")
	}

	if req.GetTarget() == nil {
		return fmt.Errorf("event target missing")
	}
	if req.GetTarget().GetName() == "" {
		return fmt.Errorf("event target name missing")
	}
	if req.GetTarget().GetApp() == "" {
		return fmt.Errorf("event target app missing")
	}

	return nil
}

func (brk *broker) StartComponent() {
	brk.start()
	defer brk.Shutdown()

	brk.startComponent()

	brk.notify()
}

func (brk *broker) StartRuntimeSrv() {
	brk.Config().IsRuntimeSrv = true
	brk.Config().SkipBootstrap = true

	brk.start()
	brk.Config().Comp.SetApp(platform.App)
	defer brk.Shutdown()

	brk.startComponent()

	brk.runtimeSrv = NewRuntimeServer(brk)
	brk.runtimeSrv.Start()

	brk.notify()
}

func (brk *broker) StartHTTPClient() {
	brk.start()
	brk.Config().Comp.SetApp(platform.App)
	defer brk.Shutdown()

	brk.httpClient = NewHTTPClient(brk)
	brk.localSender = brk.httpClient

	brk.jetListener = NewJetStreamListener(brk)
	brk.jetListener.Start()

	brk.notify()
}

func (brk *broker) StartHTTPSrv() {
	brk.Config().System = platform.System

	brk.start()
	brk.Config().Comp.SetApp(platform.App)
	defer brk.Shutdown()

	brk.jetSender = NewJetStreamSender(brk)
	brk.jetSender.Start()
	brk.remoteSender = brk.jetSender

	brk.httpSrv = NewHTTPServer(brk)
	brk.httpSrv.Start()

	brk.notify()
}

func (brk *broker) start() {
	ctx, cancel := context.WithTimeout(context.Background(), brk.EventTimeout())
	defer cancel()

	brk.fabStore = fabric.NewStore(brk)

	if brk.Config().HealthSrvAddr != "false" {
		brk.healthSrv = NewHealthServer(brk)
		brk.healthSrv.Start()
	}

	if brk.Config().TelemetryAgentAddr != "false" {
		brk.telClient = telemetry.NewClient(brk.Config(), brk.Log())
		brk.telClient.Start(ctx)
	}

	// Only try to connect to Platform Runtime's gRPC server if this broker is
	// not managing the RuntimeServer.
	if !brk.Config().IsRuntimeSrv {
		brk.runtimeClient = NewRuntimeClient(brk)
	}
	brk.bootstrap(ctx)

	brk.jetClient = jetstream.NewClient(brk.Config(), brk.Log())
	if err := brk.jetClient.Connect(); err != nil {
		brk.Log().Errorf("error connecting to nats: %v", err)
		os.Exit(kubefox.JetStreamErrorCode)
	}

	if brk.Config().IsDevMode {
		brk.startDevServices()
	}
}

func (brk *broker) startComponent() {
	brk.jetSender = NewJetStreamSender(brk)
	brk.jetSender.Start()
	brk.remoteSender = brk.jetSender

	brk.grpcSrv = NewGRPCServer(brk)
	brk.grpcSrv.Start()
	brk.localSender = brk.grpcSrv
	if brk.Config().IsRuntimeSrv {
		brk.runtimeClient = brk.grpcSrv
	}

	brk.jetListener = NewJetStreamListener(brk)
	brk.jetListener.Start()
}

func (brk *broker) bootstrap(ctx context.Context) {
	if brk.cfg.SkipBootstrap {
		return
	}

	brk.Log().Debug("bootstrapping component")

	req := kubefox.NewDataEvent(kubefox.BootstrapRequestType)
	req.SetContext(&grpc.EventContext{
		// Organization: brk.Config().Organization,
		Platform:    brk.Config().Platform,
		System:      brk.Config().System,
		Environment: platform.Env,
		App:         platform.App,
	})

	resp := brk.InvokeRuntimeServer(ctx, req)
	if resp.GetType() == kubefox.ErrorEventType {
		if resp.GetError() == nil {
			resp.SetError(errors.New(resp.GetErrorMsg()))
		}
		brk.Log().Errorf("error bootstrapping: %v", resp.GetError())
		os.Exit(kubefox.RpcServerErrorCode)
	}

	// TODO deal with bootstrap
}

func (brk *broker) startDevServices() {
	if brk.Config().DevHTTPSrvAddr != "" {
		devFlags := brk.Config().Flags // copy
		devFlags.CompName = platform.HTTPIngressAdapt.GetName()
		devFlags.IsRuntimeSrv = false
		devFlags.SkipBootstrap = true
		devFlags.HTTPSrvAddr = brk.Config().DevHTTPSrvAddr
		devFlags.DevHTTPSrvAddr = "" // ensure devBrk doesn't start another dev http server

		brk.Log().Infof("starting dev http ingress adapter %s on %s", devFlags.CompName, devFlags.HTTPSrvAddr)

		devBrk := New(devFlags)
		if brk.Config().DevApp != "" {
			devTarget := component.Copy(brk.Component())
			devTarget.SetApp(devFlags.DevApp)

			devBrk.Config().Dev = config.DevContext{
				Target: devTarget,
				EventContext: &grpc.EventContext{
					// Organization: devFlags.Organization,
					Platform:    devFlags.Platform,
					System:      devFlags.System,
					Environment: devFlags.DevEnv,
					App:         devFlags.DevApp,
				},
			}
		} else {
			// to prevent NPE
			devBrk.Config().Dev = config.DevContext{
				EventContext: &grpc.EventContext{},
				Target:       component.New(component.Fields{}),
			}
		}

		go devBrk.StartHTTPSrv()
	}
}

func (brk *broker) Shutdown() {
	brk.Log().Info("broker shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Order is important here, listeners should almost always be shutdown
	// before servers.

	if brk.jetListener != nil {
		brk.jetListener.Shutdown()
	}

	if brk.runtimeSrv != nil {
		brk.runtimeSrv.Shutdown()
	}

	if brk.httpSrv != nil {
		brk.httpSrv.Shutdown(ctx)
	}

	if brk.grpcSrv != nil {
		brk.grpcSrv.Shutdown()
	}

	if brk.jetSender != nil {
		brk.jetSender.Shutdown()
	}

	if brk.jetClient != nil {
		brk.jetClient.Close()
	}

	if brk.telClient != nil {
		brk.telClient.Shutdown(ctx)
	}

	if brk.healthSrv != nil {
		brk.healthSrv.Shutdown(ctx)
	}

	brk.Log().Sync()
}

func (brk *broker) IsHealthy(ctx context.Context) bool {
	healthy := true
	if brk.jetClient != nil {
		healthy = healthy && brk.jetClient.Healthy(ctx)
	}
	if brk.grpcSrv != nil {
		healthy = healthy && brk.grpcSrv.Healthy(ctx)
	}

	brk.Log().Debugf("health check called; healthy: %t", healthy)

	return healthy
}

func (brk *broker) Log() *logger.Log {
	return brk.log
}

func (brk *broker) EventTimeout() time.Duration {
	return time.Duration(brk.cfg.EventTimeout) * time.Second
}

func (brk *broker) ConnectTimeout() time.Duration {
	if brk.cfg.IsDevMode {
		return 5 * time.Second
	}

	return 2 * time.Minute
}

func (brk *broker) Config() *config.Config {
	return brk.cfg
}

func (brk *broker) Component() component.Component {
	return brk.Config().Comp
}

func (brk *broker) Blocker() *blocker.Blocker {
	return brk.blocker
}

func (brk *broker) JetStreamClient() *jetstream.Client {
	return brk.jetClient
}

func (brk *broker) notify() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
}

func genId() string {
	b := make([]rune, 5)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}
