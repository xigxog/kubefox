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
	ExitCodeJetStream     = 11
	ExitCodeGRPCServer    = 12
	ExitCodeHTTPServer    = 13
	ExitCodeTelemetry     = 14
	ExitCodeResourceStore = 15
	ExitCodeKubernetes    = 16
	InterruptCode         = 130
)

const (
	FormatBrokerSubject  = "evt.brk.%s"
	FormatBrokerConsumer = "broker-%s"
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
	httpSrv *HTTPServer

	jsClient *JetStreamClient
	// httpClient *HTTPClient
	k8sClient client.Client

	healthSrv *telemetry.HealthServer
	telClient *telemetry.Client

	subMgr    SubscriptionMgr
	recvCh    chan *LiveEvent
	archiveCh chan *LiveEvent

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
		archiveCh: make(chan *LiveEvent),
		store:     NewStore(config.Namespace),
		ctx:       ctx,
		cancel:    cancel,
		log:       logkf.Global,
	}
	brk.grpcSrv = NewGRPCServer(brk)
	brk.httpSrv = NewHTTPServer(brk)
	brk.jsClient = NewJetStreamClient(brk)
	// brk.httpClient = NewHTTPClient(brk)

	return brk
}

func (brk *broker) Component() *kubefox.Component {
	return brk.comp
}

func (brk *broker) Start() {
	// TODO log config
	brk.log.Debug("broker %s starting", brk.comp.Key())

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

	if err := brk.jsClient.Connect(ctx); err != nil {
		brk.shutdown(ExitCodeJetStream, err)
	}
	brk.healthSrv.Register(brk.jsClient)

	if err := brk.store.Open(brk.jsClient.ComponentsKV()); err != nil {
		brk.shutdown(ExitCodeResourceStore, err)
	}

	if err := brk.grpcSrv.Start(ctx); err != nil {
		brk.shutdown(ExitCodeGRPCServer, err)
	}

	consumer := fmt.Sprintf("broker-%s", brk.comp.Id)
	subj := brk.comp.BrokerSubject()
	descrip := fmt.Sprintf("Consumer for broker; name: %s, commit: %s, id: %s",
		brk.comp.Name, brk.comp.Commit, brk.comp.Id)

	if err := brk.jsClient.ConsumeEvents(brk.ctx, consumer, subj, descrip); err != nil {
		brk.shutdown(ExitCodeJetStream, err)
	}

	if err := brk.httpSrv.Start(); err != nil {
		brk.shutdown(ExitCodeHTTPServer, err)
	}
	if err := brk.store.RegisterAdapter(ctx, brk.httpSrv.Component()); err != nil {
		brk.shutdown(ExitCodeJetStream, err)
	}

	brk.log.Infof("starting %d workers", config.NumWorkers)
	brk.wg.Add(config.NumWorkers + 2) // +2 for archiver and updater
	go brk.startArchiver()
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
		if err := brk.store.RegisterComponent(ctx, conf.Component, conf.CompReg); err != nil {
			return nil, err
		}

		comp := sub.Component()
		consumer := comp.GroupKey()
		subj := comp.GroupSubject()
		descrip := fmt.Sprintf("Consumer for component group; name: %s, commit: %s", comp.Name, comp.Commit)
		if err = brk.jsClient.ConsumeEvents(grpSub.Context(), consumer, subj, descrip); err != nil {
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
		return ErrEventInvalid
	case evt.TTL() <= 0:
		return ErrEventTimeout
	case evt.Subscription != nil && !evt.Subscription.IsActive():
		return ErrSubCanceled
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
					err = fmt.Errorf("%w: %v", ErrComponentGone, err)

				case apierrors.IsNotFound(err): // Not found error from K8s API
					err = fmt.Errorf("%w: %v", ErrRouteNotFound, err)

				case !errors.Is(err, ErrKubeFox): // Unknown
					err = fmt.Errorf("%w: %v", ErrUnexpected, err)
				}

				log = log.WithEvent(evt.Event)

				switch {
				case errors.Is(err, ErrUnexpected):
					log.Error(err)
				case errors.Is(err, ErrBrokerMismatch):
					log.Warn(err)
				default:
					log.Info(err)
				}

				evt.Err(err)
			}
		case <-brk.ctx.Done():
			return
		}
	}
}

func (brk *broker) routeEvent(log *logkf.Logger, evt *LiveEvent) error {
	log.WithEvent(evt.Event).Debugf("routing event from receiver '%s'", evt.Receiver)

	if evt.TTL() <= 0 {
		return fmt.Errorf("%w: event timed out while waiting for worker", ErrEventTimeout)
	}
	if evt.Receiver == ReceiverJetStream &&
		evt.Target != nil &&
		evt.Target.BrokerId != "" &&
		evt.Target.BrokerId != brk.comp.Id {
		return fmt.Errorf("%w: event target broker id is %s", ErrBrokerMismatch, evt.Target.BrokerId)
	}

	// check that source is full
	// if from grpc check target only has name or is full
	// also check context is valid

	ctx, cancel := context.WithTimeout(context.Background(), evt.TTL())
	defer cancel()

	if err := brk.matchEvent(ctx, evt); err != nil {
		return err
	}

	// Set log attributes after matching.
	log = log.WithEvent(evt.Event)
	log.Debug("matched event to target")

	if err := brk.checkMatchedEvent(ctx, evt); err != nil {
		return err
	}

	// TODO add policy checks

	var sub Subscription
	switch {
	case evt.Subscription != nil:
		sub = evt.Subscription

	case evt.Target.Equal(brk.httpSrv.Component()):
		sub = brk.httpSrv.Subscription()

	case evt.Target.Id != "":
		sub, _ = brk.subMgr.ReplicaSubscription(evt.Target)

	case evt.Target.Id == "":
		sub, _ = brk.subMgr.GroupSubscription(evt.Target)
	}

	var sendEvent SendEvent
	switch {
	case sub != nil:
		sendEvent = func(evt *LiveEvent) error {
			log.Debug("subscription found, sending event with gRPC")
			err := sub.SendEvent(evt)
			if evt.Receiver != ReceiverJetStream {
				evt.Target.BrokerId = brk.comp.BrokerId
				brk.archiveCh <- evt
			}
			return err
		}

	default:
		sendEvent = func(evt *LiveEvent) error {
			log.Debug("subscription not found, sending event with JetStream")
			err := brk.jsClient.Publish(evt.Target.Subject(), evt.Event)
			if err != nil {
				return err
			}
			return nil
		}
	}

	if evt.TTL() <= 0 {
		return fmt.Errorf("%w: event timed out during routing", ErrEventTimeout)
	}

	if err := sendEvent(evt); err != nil {
		return fmt.Errorf("%w: unable to send event", err)
	}

	return nil
}

func (brk *broker) matchEvent(ctx context.Context, evt *LiveEvent) error {
	if evt.Category == kubefox.Category_RESPONSE {
		if evt.Target == nil || !evt.Target.IsFull() {
			return fmt.Errorf("%w: response target is missing required attribute", ErrRouteInvalid)
		}

		evt.MatchedEvent = &kubefox.MatchedEvent{Event: evt.Event}
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
		err = fmt.Errorf("%w: event missing deployment or environment context", ErrEventInvalid)
	}
	if err != nil {
		return err
	}

	var (
		routeId int
	)
	route, matched := matcher.Match(evt.Event)
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
		return ErrRouteNotFound
	}

	var (
		rel     *v1alpha1.Release
		env     *v1alpha1.Environment
		envVars map[string]*kubefox.Val
	)
	switch {
	case evt.Context.Release != "":
		rel, err = brk.store.Release(evt.Context.Release)
		if rel != nil {
			envVars = rel.Spec.Environment.Vars
		}

	case evt.Context.Environment != "":
		env, err = brk.store.Environment(evt.Context.Environment)
		if env != nil {
			envVars = env.Spec.Vars
		}
	}
	if err != nil {
		return err
	}

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

func (brk *broker) checkMatchedEvent(ctx context.Context, evt *LiveEvent) error {
	// TODO check source is full
	// TODO check source is part of context
	if evt.Target == nil || evt.Target.Name == "" || evt.Context == nil {
		return ErrComponentMismatch
	}

	// Check if target is local adapter.
	switch {
	case evt.Target.Equal(brk.httpSrv.Component()):
		return nil
	}

	var (
		depSpec v1alpha1.DeploymentSpec
	)
	switch {
	case evt.Context.IsRelease():
		if rel, err := brk.store.Release(evt.Context.Release); err != nil {
			return fmt.Errorf("%w: unable to get release: %v", ErrUnexpected, err)
		} else {
			depSpec = rel.Spec.Deployment
		}

	case evt.Context.IsDeployment():
		if dep, err := brk.store.Deployment(evt.Context.Deployment); err != nil {
			return fmt.Errorf("%w: unable to get deployment: %v", ErrUnexpected, err)
		} else {
			depSpec = dep.Spec
		}

	default:
		return ErrUnexpected
	}

	depComp, found := depSpec.Components[evt.Target.Name]
	switch {
	case !found:
		if adapter := brk.store.Adapter(ctx, evt.Target); !adapter {
			return fmt.Errorf("%w: component not part of deployment", ErrComponentMismatch)
		}

	case evt.Target.Commit == "" && evt.MatchedEvent.RouteId == kubefox.DefaultRouteId:
		evt.Target.Commit = depComp.Commit
		reg, err := brk.store.Component(ctx, evt.Target)
		if err != nil {
			return err
		}
		if !reg.DefaultHandler {
			return fmt.Errorf("%w: component does not have default handler", ErrRouteNotFound)
		}

	case evt.Target.Commit != depComp.Commit:
		return fmt.Errorf("%w: component commit does not match deployment", ErrComponentMismatch)
	}

	return nil
}

func (brk *broker) startArchiver() {
	log := brk.log.With(logkf.KeyWorker, "archiver")
	defer func() {
		log.Info("archiver stopped")
		brk.wg.Done()
	}()

	for {
		select {
		case evt := <-brk.archiveCh:
			log := log.WithEvent(evt.Event)
			log.Debug("publishing event to JetStream for archiving")

			err := brk.jsClient.Publish(evt.Target.DirectSubject(), evt.Event)
			if err != nil {
				// This event will be lost in time, never to be seen again :(
				log.Warnf("unable to publish event to JetStream, event not archived: %v", err)
			}

		case <-brk.ctx.Done():
			return
		}
	}
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

	if err := brk.store.RegisterAdapter(ctx, brk.httpSrv.Component()); err != nil {
		log.Error("error updating components key/value: %v", err)
	}

	for _, sub := range brk.subMgr.Subscriptions() {
		if !sub.IsGroupEnabled() {
			continue
		}
		if err := brk.store.RegisterComponent(ctx, sub.Component(), sub.ComponentReg()); err != nil {
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
	brk.httpSrv.Shutdown(timeout)

	brk.subMgr.Close()
	brk.grpcSrv.Shutdown(timeout)

	// Stops workers.
	brk.cancel()
	brk.wg.Wait()

	brk.store.Close()

	brk.jsClient.Close()
	brk.telClient.Shutdown(timeout)

	os.Exit(code)
}
