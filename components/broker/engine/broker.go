package engine

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-logr/zapr"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/build"
	"github.com/xigxog/kubefox/components/broker/config"
	"github.com/xigxog/kubefox/components/broker/telemetry"
	kubefox "github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/logkf"
	"github.com/xigxog/kubefox/matcher"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	authv1 "k8s.io/api/authentication/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// OS status codes
const (
	ExitCodeConfiguration = 10
	ExitCodeNATS          = 11
	ExitCodeGRPCServer    = 12
	ExitCodeHTTP          = 13
	ExitCodeTelemetry     = 14
	ExitCodeResourceStore = 15
	ExitCodeKubernetes    = 16
	InterruptCode         = 130
)

var (
	timeout    = 30 * time.Second
	NoopCancel = func(err error) {}
)

type Engine interface {
	Start()
}

type Broker interface {
	AuthorizeComponent(context.Context, *kubefox.Component, string) error
	Subscribe(context.Context, *SubscriptionConf) (ReplicaSubscription, error)
	RecvEvent(*LiveEvent) error
	Component() *kubefox.Component
}

type broker struct {
	comp *kubefox.Component

	grpcSrv *GRPCServer

	natsClient *NATSClient
	httpClient *HTTPClient
	k8sClient  client.Client

	healthSrv *telemetry.HealthServer
	telClient *telemetry.Client

	subMgr SubscriptionMgr
	recvCh chan *LiveEvent

	store *Store

	ctx    context.Context
	cancel context.CancelFunc

	wg sync.WaitGroup

	log *logkf.Logger
}

func New() Engine {
	name, id := kubefox.GenerateNameAndId()
	logkf.Global = logkf.Global.
		With(logkf.KeyBrokerId, id).
		With(logkf.KeyBrokerName, name)
	ctrl.SetLogger(zapr.NewLogger(logkf.Global.Unwrap().Desugar()))

	ctx, cancel := context.WithCancel(context.Background())
	brk := &broker{
		comp: &kubefox.Component{
			Name:     name,
			Commit:   build.Info.Commit,
			Id:       id,
			BrokerId: id,
		},
		healthSrv: telemetry.NewHealthServer(),
		telClient: telemetry.NewClient(),
		subMgr:    NewManager(),
		recvCh:    make(chan *LiveEvent),
		store:     NewStore(config.Namespace),
		ctx:       ctx,
		cancel:    cancel,
		log:       logkf.Global,
	}
	brk.grpcSrv = NewGRPCServer(brk)
	brk.natsClient = NewNATSClient(brk)
	brk.httpClient = NewHTTPClient(brk)

	return brk
}

func (brk *broker) Component() *kubefox.Component {
	return brk.comp
}

func (brk *broker) Start() {
	brk.log.Debugf("broker %s starting", brk.comp.Key())

	ctx, cancel := context.WithTimeout(brk.ctx, timeout)
	defer cancel()

	cfg, err := ctrl.GetConfig()
	if err != nil {
		brk.shutdown(ExitCodeKubernetes, err)
	}
	k8s, err := client.New(cfg, client.Options{})
	if err != nil {
		brk.shutdown(ExitCodeKubernetes, err)
	}
	brk.k8sClient = k8s

	if config.HealthSrvAddr != "false" {
		if err := brk.healthSrv.Start(); err != nil {
			brk.shutdown(ExitCodeTelemetry, err)
		}
	}

	if config.TelemetryAddr != "false" {
		if err := brk.telClient.Start(ctx); err != nil {
			brk.shutdown(ExitCodeTelemetry, err)
		}
	}

	if err := brk.natsClient.Connect(ctx); err != nil {
		brk.shutdown(ExitCodeNATS, err)
	}
	brk.healthSrv.Register(brk.natsClient)

	if err := brk.store.Open(brk.natsClient.ComponentsKV()); err != nil {
		brk.shutdown(ExitCodeResourceStore, err)
	}

	if err := brk.grpcSrv.Start(ctx); err != nil {
		brk.shutdown(ExitCodeGRPCServer, err)
	}

	consumer := fmt.Sprintf("broker-%s", brk.comp.Id)
	subj := brk.comp.BrokerSubject()
	if err := brk.natsClient.ConsumeEvents(brk.ctx, consumer, subj); err != nil {
		brk.shutdown(ExitCodeNATS, err)
	}

	brk.log.Infof("starting %d workers", config.NumWorkers)
	brk.wg.Add(config.NumWorkers + 1) // +1 for reg updater
	go brk.startRegUpdater()
	for i := 0; i < config.NumWorkers; i++ {
		go brk.startWorker(i)
	}

	brk.log.Info("broker started")

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-ch

	brk.shutdown(0, nil)
}

func (brk *broker) Subscribe(ctx context.Context, conf *SubscriptionConf) (ReplicaSubscription, error) {
	sub, grpSub, err := brk.subMgr.Create(ctx, conf, brk.recvCh)
	if err != nil {
		return nil, err
	}

	if sub.IsGroupEnabled() {
		if err := brk.store.RegisterComponent(ctx, conf.Component, conf.ComponentSpec); err != nil {
			return nil, err
		}

		comp := sub.Component()
		consumer := comp.GroupKey()
		subj := comp.GroupSubject()
		if err = brk.natsClient.ConsumeEvents(grpSub.Context(), consumer, subj); err != nil {
			return nil, err
		}
	}

	// TODO deal with health checks

	return sub, nil
}

func (brk *broker) AuthorizeComponent(ctx context.Context, comp *kubefox.Component, authToken string) error {
	parsed, err := jwt.ParseString(authToken)
	if err != nil {
		return err
	}
	var svcAccName string
	if k, ok := parsed.PrivateClaims()["kubernetes.io"].(map[string]interface{}); ok {
		if sa, ok := k["serviceaccount"].(map[string]interface{}); ok {
			if n, ok := sa["name"].(string); ok {
				svcAccName = n
			}
		}
	}
	if !strings.HasSuffix(svcAccName, comp.GroupKey()) {
		return fmt.Errorf("service account name does not match component")
	}

	review := authv1.TokenReview{
		ObjectMeta: metav1.ObjectMeta{
			Name: svcAccName,
		},
		Spec: authv1.TokenReviewSpec{
			Token: authToken,
		},
	}
	if err := brk.k8sClient.Create(ctx, &review); err != nil {
		return err
	}
	if !review.Status.Authenticated {
		return fmt.Errorf("unauthorized component: %s", review.Status.Error)
	}

	return nil
}

func (brk *broker) RecvEvent(evt *LiveEvent) error {
	switch {
	case evt == nil || evt.Event == nil:
		return kubefox.ErrEventInvalid

	case evt.TTL() <= 0:
		return kubefox.ErrEventTimeout
	}

	brk.recvCh <- evt

	return nil
}

func (brk *broker) startWorker(id int) {
	log := brk.log.With(logkf.KeyWorker, fmt.Sprintf("worker-%d", id))
	defer func() {
		log.Info("worker stopped")
		brk.wg.Done()
	}()

	for {
		select {
		case evt := <-brk.recvCh:
			if err := brk.routeEvent(log, evt); err != nil {
				_, gRPCErr := status.FromError(err)
				switch {
				case gRPCErr: // gRPC error
					err = fmt.Errorf("%w: %v", kubefox.ErrComponentGone, err)

				case apierrors.IsNotFound(err): // Not found error from K8s API
					err = fmt.Errorf("%w: %v", kubefox.ErrRouteNotFound, err)

				case !errors.Is(err, kubefox.ErrKubeFox): // Unknown
					err = fmt.Errorf("%w: %v", kubefox.ErrUnexpected, err)
				}

				log = log.WithEvent(evt.Event)

				switch {
				case errors.Is(err, kubefox.ErrUnexpected):
					log.Error(err)
				case errors.Is(err, kubefox.ErrBrokerMismatch):
					log.Warn(err)
				default:
					log.Info(err)
				}

				if evt.ErrCh != nil {
					evt.ErrCh <- err
				}

			} else if evt.SentCh != nil {
				close(evt.SentCh)
			}

		case <-brk.ctx.Done():
			return
		}
	}
}

func (brk *broker) routeEvent(log *logkf.Logger, evt *LiveEvent) error {
	log.WithEvent(evt.Event).Debugf("routing event from receiver '%s'", evt.Receiver)

	ctx, cancel := context.WithTimeout(context.Background(), evt.TTL())
	defer cancel()

	if err := brk.checkEvent(evt); err != nil {
		return err
	}

	if err := brk.matchEvent(ctx, evt); err != nil {
		return err
	}

	// Set log attributes after matching.
	log = log.WithEvent(evt.Event)
	log.Debug("matched event to target")

	if err := brk.matchComponents(ctx, evt); err != nil {
		return err
	}

	// TODO add policy checks

	var sub Subscription
	if evt.TargetAdapter == nil {
		switch {
		case evt.Target.Id != "":
			sub, _ = brk.subMgr.ReplicaSubscription(evt.Target)

		case evt.Target.Id == "":
			sub, _ = brk.subMgr.GroupSubscription(evt.Target)
		}
	}

	var sendEvent SendEvent
	switch {
	case evt.TargetAdapter != nil:
		sendEvent = func(evt *LiveEvent) error {
			log.Debug("sending event with http client adapter")
			return brk.httpClient.SendEvent(evt)
		}

	case sub != nil:
		sendEvent = func(evt *LiveEvent) error {
			log.Debug("subscription found, sending event with gRPC")
			return sub.SendEvent(evt)
		}

	default:
		sendEvent = func(evt *LiveEvent) error {
			log.Debug("subscription not found, sending event with nats")
			return brk.natsClient.Publish(evt.Target.Subject(), evt.Event)
		}
	}

	if evt.TTL() <= 0 {
		return fmt.Errorf("%w: event timed out during routing", kubefox.ErrEventTimeout)
	}

	return sendEvent(evt)
}

func (brk *broker) checkEvent(evt *LiveEvent) error {
	if evt.TTL() <= 0 {
		return fmt.Errorf("%w: event timed out while waiting for worker", kubefox.ErrEventTimeout)
	}

	if evt.Source == nil || !evt.Source.IsFull() {
		return fmt.Errorf("%w: event source is invalid", kubefox.ErrEventInvalid)
	}

	if evt.Category == kubefox.Category_RESPONSE && (evt.Target == nil || !evt.Target.IsFull()) {
		return fmt.Errorf("%w: response target is missing required attribute", kubefox.ErrRouteInvalid)
	}

	switch evt.Receiver {
	case ReceiverNATS:
		if evt.Target != nil &&
			evt.Target.BrokerId != "" &&
			evt.Target.BrokerId != brk.comp.Id {
			return fmt.Errorf("%w: event target broker id is %s", kubefox.ErrBrokerMismatch, evt.Target.BrokerId)
		}

	case ReceiverGRPCServer:
		if evt.Target != nil && !evt.Target.IsFull() && !evt.Target.IsNameOnly() {
			return fmt.Errorf("%w: event target is invalid", kubefox.ErrEventInvalid)
		}

		// If a valid context is not present reject.
		if evt.Context == nil || !evt.Context.IsDeployment() && !evt.Context.IsRelease() {
			return fmt.Errorf("%w: event context is invalid", kubefox.ErrEventInvalid)
		}
	}

	return nil
}

func (brk *broker) matchEvent(ctx context.Context, evt *LiveEvent) error {
	var (
		envVars map[string]*kubefox.Val
		matcher *matcher.EventMatcher
		err     error
	)
	switch {
	case evt.Context.IsRelease():
		if rel, err := brk.store.Release(evt.Context.Release); err != nil {
			return fmt.Errorf("%w: unable to get release: %v", kubefox.ErrUnexpected, err)
		} else {
			envVars = rel.Spec.Environment.Vars
			evt.Adapters = rel.Spec.Environment.Adapters
		}

		matcher, err = brk.store.ReleaseMatcher(ctx)

	case evt.Context.IsDeployment():
		if env, err := brk.store.Environment(evt.Context.Environment); err != nil {
			return fmt.Errorf("%w: unable to get environment: %v", kubefox.ErrUnexpected, err)
		} else {
			envVars = env.Spec.Vars
			evt.Adapters = env.Spec.Adapters
		}

		matcher, err = brk.store.DeploymentMatcher(ctx, evt.Context)

	default:
		return fmt.Errorf("%w: event missing deployment or environment context", kubefox.ErrEventInvalid)
	}
	if err != nil {
		return fmt.Errorf("%w: error finding matchers: %v", kubefox.ErrUnexpected, err)
	}

	if evt.Category == kubefox.Category_RESPONSE {
		evt.MatchedEvent = &kubefox.MatchedEvent{Event: evt.Event}
		return nil
	}

	route, matched := matcher.Match(evt.Event)

	var (
		routeId int
	)
	switch {
	case matched:
		routeId = route.Id
		if evt.Target == nil {
			evt.Target = &kubefox.Component{}
		}
		evt.Target.Name = route.Component.Name
		evt.Target.Commit = route.Component.Commit
		evt.SetContext(route.EventContext)

	case evt.Target != nil && evt.Target.Name != "":
		routeId = kubefox.DefaultRouteId

	default:
		return kubefox.ErrRouteNotFound
	}

	// TODO environment checks
	// - only copy env vars required by component in spec
	// - check for required vars
	// - check for unique vars; expensive, don't do for releases, cache
	// - check for var type
	evt.MatchedEvent = &kubefox.MatchedEvent{
		Env:     make(map[string]*structpb.Value, len(envVars)),
		Event:   evt.Event,
		RouteId: int64(routeId),
	}
	for k, v := range envVars {
		evt.MatchedEvent.Env[k] = v.Proto()
	}

	return nil
}

func (brk *broker) matchComponents(ctx context.Context, evt *LiveEvent) error {
	if evt.Target == nil || evt.Target.Name == "" || evt.Context == nil {
		return kubefox.ErrComponentMismatch
	}

	var depSpec v1alpha1.DeploymentSpec
	switch {
	case evt.Context.IsRelease():
		if rel, err := brk.store.Release(evt.Context.Release); err != nil {
			return fmt.Errorf("%w: unable to get release: %v", kubefox.ErrUnexpected, err)
		} else {
			depSpec = rel.Spec.Deployment
		}

	case evt.Context.IsDeployment():
		if dep, err := brk.store.Deployment(evt.Context.Deployment); err != nil {
			return fmt.Errorf("%w: unable to get deployment: %v", kubefox.ErrUnexpected, err)
		} else {
			depSpec = dep.Spec
		}

	default:
		return kubefox.ErrUnexpected
	}

	// Check if target is part of deployment spec.
	var adapter *v1alpha1.Adapter
	if evt.Adapters != nil {
		adapter = evt.Adapters[evt.Target.Name]
	}
	depComp := depSpec.Components[evt.Target.Name]
	switch {
	case depComp == nil && adapter == nil:
		if !brk.store.IsGenesisAdapter(ctx, evt.Target) {
			return fmt.Errorf("%w: target component not part of deployment", kubefox.ErrComponentMismatch)
		}

	case depComp == nil && adapter != nil:
		if adapter.Type != kubefox.ComponentTypeHTTP {
			return fmt.Errorf("%w: adapter type '%s' is not supported", kubefox.ErrUnsupportedAdapter, adapter.Type)
		}
		evt.TargetAdapter = adapter

	case evt.Target.Commit == "" && evt.MatchedEvent.RouteId == kubefox.DefaultRouteId:
		evt.Target.Commit = depComp.Commit
		reg, err := brk.store.Component(ctx, evt.Target)
		if err != nil {
			return err
		}
		if !reg.DefaultHandler {
			return fmt.Errorf("%w: target component does not have default handler", kubefox.ErrRouteNotFound)
		}

	case evt.Target.Commit != depComp.Commit:
		return fmt.Errorf("%w: target component commit does not match deployment", kubefox.ErrComponentMismatch)
	}

	// Check if source is part of deployment spec.
	if evt.Adapters != nil {
		adapter = evt.Adapters[evt.Source.Name]
	}
	depComp = depSpec.Components[evt.Source.Name]
	switch {
	case depComp == nil && adapter == nil:
		if !brk.store.IsGenesisAdapter(ctx, evt.Source) {
			return fmt.Errorf("%w: source component not part of deployment", kubefox.ErrComponentMismatch)
		}

	case depComp == nil && adapter != nil:
		if evt.Source.BrokerId != brk.comp.BrokerId {
			return fmt.Errorf("%w: source component is adapter but does not match broker", kubefox.ErrBrokerMismatch)
		}

	case evt.Source.Commit != depComp.Commit:
		return fmt.Errorf("%w: source component commit does not match deployment", kubefox.ErrComponentMismatch)
	}

	return nil
}

func (brk *broker) startRegUpdater() {
	log := brk.log.With(logkf.KeyWorker, "registration-updater")
	defer func() {
		log.Info("registration-updater stopped")
		brk.wg.Done()
	}()

	ticker := time.NewTicker(ComponentsTTL / 2)
	for {
		select {
		case <-ticker.C:
			brk.updateReg(log)

		case <-brk.ctx.Done():
			return
		}
	}
}

func (brk *broker) updateReg(log *logkf.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	for _, sub := range brk.subMgr.Subscriptions() {
		if !sub.IsGroupEnabled() {
			continue
		}
		if err := brk.store.RegisterComponent(ctx, sub.Component(), sub.ComponentSpec()); err != nil {
			log.Error("error updating component registration: %v", err)
		}
	}
}

func (brk *broker) shutdown(code int, err error) {
	// TODO deal with inflight events when shutdown occurs

	brk.log.Infof("broker shutting down, exit code %d", code)
	if err != nil {
		brk.log.Error(err)
	}

	brk.healthSrv.Shutdown(timeout)

	brk.subMgr.Close()
	brk.grpcSrv.Shutdown(timeout)

	// Stops workers.
	brk.cancel()
	brk.wg.Wait()

	brk.store.Close()

	brk.natsClient.Close()
	brk.telClient.Shutdown(timeout)

	os.Exit(code)
}
