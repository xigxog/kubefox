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
	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/build"
	"github.com/xigxog/kubefox/components/broker/config"
	"github.com/xigxog/kubefox/components/broker/telemetry"
	"github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/k8s"
	"github.com/xigxog/kubefox/logkf"
	"github.com/xigxog/kubefox/matcher"
	authv1 "k8s.io/api/authentication/v1"
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
	AuthorizeComponent(context.Context, *core.Component, string) error
	Subscribe(context.Context, *SubscriptionConf) (ReplicaSubscription, error)
	RecvEvent(evt *core.Event, receiver Receiver) *BrokerEvent
	Component() *core.Component
}

type broker struct {
	comp *core.Component

	grpcSrv *GRPCServer

	natsClient *NATSClient
	httpClient *HTTPClient
	k8sClient  client.Client

	healthSrv *telemetry.HealthServer
	telClient *telemetry.Client

	subMgr SubscriptionMgr
	recvCh chan *BrokerEvent

	store *Store

	ctx    context.Context
	cancel context.CancelFunc

	wg sync.WaitGroup

	log *logkf.Logger
}

func New() Engine {
	name, id := core.GenerateNameAndId()
	logkf.Global = logkf.Global.
		With(logkf.KeyBrokerId, id).
		With(logkf.KeyBrokerName, name)
	ctrl.SetLogger(zapr.NewLogger(logkf.Global.Unwrap().Desugar()))

	ctx, cancel := context.WithCancel(context.Background())
	brk := &broker{
		comp: &core.Component{
			Name:     name,
			Commit:   build.Info.Commit,
			Id:       id,
			BrokerId: id,
		},
		healthSrv: telemetry.NewHealthServer(),
		telClient: telemetry.NewClient(),
		subMgr:    NewManager(),
		recvCh:    make(chan *BrokerEvent),
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

func (brk *broker) Component() *core.Component {
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

	if err := brk.store.Open(); err != nil {
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
	brk.wg.Add(config.NumWorkers)
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

func (brk *broker) AuthorizeComponent(ctx context.Context, comp *core.Component, authToken string) error {
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

	// TODO use platform status to check if component is platform component.
	platformCompName := fmt.Sprintf("%s-%s", config.Platform, comp.Name)
	if !strings.HasSuffix(svcAccName, comp.GroupKey()) && svcAccName != platformCompName {
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

	// TODO make sure component exists in an AppDeployment

	return nil
}

func (brk *broker) RecvEvent(evt *core.Event, receiver Receiver) *BrokerEvent {
	brkEvt := &BrokerEvent{
		Event:      evt,
		Receiver:   receiver,
		ReceivedAt: time.Now(),
		DoneCh:     make(chan *core.Err),
	}

	go func() {
		brk.recvCh <- brkEvt
	}()

	return brkEvt
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
				l := log.WithEvent(evt.Event)

				kfErr := &core.Err{}
				if ok := errors.As(err, &kfErr); !ok {
					kfErr = core.ErrUnexpected(err)
				}

				switch kfErr.Code() {
				case core.CodeUnexpected:
					l.Error(err)
				case core.CodeBrokerMismatch:
					l.Warn(err)
				case core.CodeUnauthorized:
					l.Warn(err)
				default:
					l.Debug(err)
				}

				go func() {
					evt.DoneCh <- kfErr
				}()

			} else {
				close(evt.DoneCh)
			}

		case <-brk.ctx.Done():
			return
		}
	}
}

func (brk *broker) routeEvent(log *logkf.Logger, evt *BrokerEvent) error {
	log.WithEvent(evt.Event).Debugf("routing event from receiver '%s'", evt.Receiver)

	ctx, cancel := context.WithTimeout(context.Background(), evt.TTL())
	defer cancel()

	if err := brk.checkEvent(evt); err != nil {
		return err
	}
	if err := brk.findTarget(ctx, evt); err != nil {
		return err
	}

	// Set log attributes after matching.
	log = log.WithEvent(evt.Event)
	if evt.Target != nil {
		log.Debugf("matched event to target '%s'", evt.Target)
	}

	if err := brk.attachEnv(ctx, evt); err != nil {
		return err
	}
	if err := brk.checkComponents(ctx, evt); err != nil {
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

	switch {
	case evt.TargetAdapter != nil:
		// Target is a local adapter.
		log.Debugf("sending event with adapter of type '%s'", evt.TargetAdapter.Type)
		return brk.httpClient.SendEvent(evt)

	case sub != nil:
		// Found component subscribed via gRPC.
		log.Debug("subscription found, sending event with gRPC")
		return sub.SendEvent(evt)

	case evt.Receiver != ReceiverNATS && evt.Target.BrokerId != brk.comp.Id:
		// Component not found locally, send via NATS.
		log.Debug("subscription not found, sending event with nats")
		return brk.natsClient.Publish(evt.Target.Subject(), evt.Event)

	default:
		return core.ErrComponentGone()
	}
}

func (brk *broker) checkEvent(evt *BrokerEvent) error {
	if evt.TTL() <= 0 {
		return core.ErrTimeout()
	}

	if evt.Source == nil || !evt.Source.IsFull() {
		return core.ErrInvalid(fmt.Errorf("event source is invalid"))
	}

	if evt.Category == core.Category_RESPONSE && (evt.Target == nil || !evt.Target.IsFull()) {
		return core.ErrInvalid(fmt.Errorf("response target is missing required attribute"))
	}

	switch evt.Receiver {
	case ReceiverNATS:
		if evt.Target != nil &&
			evt.Target.BrokerId != "" &&
			evt.Target.BrokerId != brk.comp.Id {

			return core.ErrBrokerMismatch(fmt.Errorf("event target broker id is %s", evt.Target.BrokerId))
		}

	case ReceiverGRPCServer:
		if evt.Target != nil && !evt.Target.IsFull() && !evt.Target.IsNameOnly() {
			return core.ErrInvalid(fmt.Errorf("event target is invalid"))
		}

		// If a valid context is not present reject.
		if evt.Context == nil || (!evt.Context.IsDeployment() && !evt.Context.IsRelease()) {
			return core.ErrInvalid(fmt.Errorf("event context is invalid"))
		}
	}

	return nil
}

func (brk *broker) findTarget(ctx context.Context, evt *BrokerEvent) error {
	if evt.Target != nil && evt.Target.IsFull() {
		return nil
	}

	var (
		matcher *matcher.EventMatcher
		err     error
	)
	switch {
	case evt.Context.IsRelease():
		matcher, err = brk.store.ReleaseMatcher(ctx)

	case evt.Context.IsDeployment():
		matcher, err = brk.store.DeploymentMatcher(ctx, evt.Context)

	default:
		return core.ErrInvalid(fmt.Errorf("event missing deployment or environment context"))
	}
	if err != nil {
		if k8s.IsNotFound(err) {
			return core.ErrNotFound(err)
		}
		return core.ErrUnexpected(err)
	}

	route, matched := matcher.Match(evt.Event)
	switch {
	case matched:
		evt.RouteId = int64(route.Id)
		if evt.Target == nil {
			evt.Target = &core.Component{}
		}
		evt.Target.Name = route.Component.Name
		evt.Target.Commit = route.Component.Commit
		evt.SetContext(route.EventContext)

	case evt.Target != nil && evt.Target.Name != "":
		evt.RouteId = api.DefaultRouteId

	default:
		return core.ErrRouteNotFound()
	}

	return nil
}

func (brk *broker) attachEnv(ctx context.Context, evt *BrokerEvent) error {
	// TODO environment checks
	// - only copy env vars required by component in spec
	// - check for required vars
	// - check for unique vars; expensive, don't do for releases, cache
	// - check for var type

	switch {
	case evt.Context.IsRelease():
		if env, err := brk.store.ReleaseEnv(evt.Context.Release); err != nil {
			return core.ErrNotFound(err)
		} else {
			evt.EnvVars = env.GetData().Vars
			evt.Adapters = env.GetData().Adapters
		}

	case evt.Context.IsDeployment():
		if env, err := brk.store.Environment(evt.Context.Environment); err != nil {
			return core.ErrNotFound(err)
		} else {
			evt.EnvVars = env.GetData().Vars
			evt.Adapters = env.GetData().Adapters
		}
	}

	return nil
}

func (brk *broker) checkComponents(ctx context.Context, evt *BrokerEvent) error {
	if evt.Target == nil || evt.Target.Name == "" || evt.Context == nil {
		return core.ErrComponentMismatch()
	}

	var (
		appDep *v1alpha1.AppDeployment
		err    error
	)
	switch {
	case evt.Context.IsRelease():
		appDep, err = brk.store.ReleaseAppDeployment(evt.Context.Release)
	case evt.Context.IsDeployment():
		appDep, err = brk.store.AppDeployment(evt.Context.Deployment)
	default:
		if k8s.IsNotFound(err) {
			return core.ErrNotFound(err)
		}
		return core.ErrUnexpected()
	}
	if err != nil {
		return core.ErrNotFound(err)
	}

	// Check if target is part of deployment spec.
	var adapter *v1alpha1.Adapter
	if evt.Adapters != nil {
		adapter = evt.Adapters[evt.Target.Name]
	}
	depComp := appDep.Spec.App.Components[evt.Target.Name]
	switch {
	case depComp == nil && adapter == nil:
		if !brk.store.IsGenesisAdapter(ctx, evt.Target) {
			return core.ErrComponentMismatch(fmt.Errorf("target component not part of deployment"))
		}

	case depComp == nil && adapter != nil:
		if adapter.Type != api.ComponentTypeHTTP {
			return core.ErrUnsupportedAdapter(fmt.Errorf("adapter type '%s' is not supported", adapter.Type))
		}
		evt.TargetAdapter = adapter

	case evt.Target.Commit == "" && evt.RouteId == api.DefaultRouteId:
		evt.Target.Commit = depComp.Commit
		reg, found := brk.store.Component(ctx, evt.Target)
		if !found {
			return core.ErrNotFound(fmt.Errorf("target component not found"))
		}
		if !reg.DefaultHandler {
			return core.ErrRouteNotFound(fmt.Errorf("target component does not have default handler"))
		}

	case evt.Target.Commit != depComp.Commit:
		return core.ErrComponentMismatch(fmt.Errorf("target component commit does not match deployment"))
	}

	// Check if source is part of deployment spec.
	if evt.Adapters != nil {
		adapter = evt.Adapters[evt.Source.Name]
	}
	depComp = appDep.Spec.App.Components[evt.Source.Name]
	switch {
	case depComp == nil && adapter == nil:
		if !brk.store.IsGenesisAdapter(ctx, evt.Source) {
			return core.ErrComponentMismatch(fmt.Errorf("source component not part of deployment"))
		}

	case depComp == nil && adapter != nil:
		if evt.Source.BrokerId != brk.comp.BrokerId {
			return core.ErrBrokerMismatch(fmt.Errorf("source component is adapter but does not match broker"))
		}

	case evt.Source.Commit != depComp.Commit:
		return core.ErrComponentMismatch(fmt.Errorf("source component commit does not match deployment"))
	}

	return nil
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
