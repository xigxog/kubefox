package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/lestrrat-go/jwx/jwt"
	"github.com/xigxog/kubefox/components/broker/config"
	"github.com/xigxog/kubefox/components/broker/telemetry"
	"github.com/xigxog/kubefox/libs/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/logkf"
	"github.com/xigxog/kubefox/libs/core/matcher"
	"github.com/xigxog/kubefox/libs/core/utils"
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
	ConfigurationExitCode = 10
	JetStreamExitCode     = 11
	GRPCServerExitCode    = 12
	HTTPServerExitCode    = 13
	TelemetryExitCode     = 14
	ResourceStoreExitCode = 15
	kubernetesExitCode    = 16
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
	RecvEvent(*ReceivedEvent) error
}

type broker struct {
	grpcSrv *GRPCServer
	httpSrv *HTTPServer

	jsClient *JetStreamClient
	// httpClient *HTTPClient
	k8sClient client.Client

	healthSrv *telemetry.HealthServer
	telClient *telemetry.Client

	subMgr     SubscriptionMgr
	recvCh     chan *ReceivedEvent
	archiveCh  chan *ReceivedEvent
	kvUpdateCh chan *IdSeq

	store *Store

	ctx    context.Context
	cancel context.CancelFunc

	wg sync.WaitGroup

	log *logkf.Logger
}

func New() Engine {
	ctx, cancel := context.WithCancel(context.Background())
	brk := &broker{
		healthSrv:  telemetry.NewHealthServer(),
		telClient:  telemetry.NewClient(),
		subMgr:     NewManager(),
		recvCh:     make(chan *ReceivedEvent),
		archiveCh:  make(chan *ReceivedEvent),
		kvUpdateCh: make(chan *IdSeq),
		store:      NewStore(config.Namespace),
		ctx:        ctx,
		cancel:     cancel,
		log:        logkf.Global,
	}
	brk.grpcSrv = NewGRPCServer(brk)
	brk.httpSrv = NewHTTPServer(brk)
	brk.jsClient = NewJetStreamClient(brk)
	// brk.httpClient = NewHTTPClient(brk)

	return brk
}

func (brk *broker) Start() {
	// TODO log config

	brk.log.Debug("broker starting")

	ctx, cancel := context.WithTimeout(brk.ctx, timeout)
	defer cancel()

	cfg, err := ctrl.GetConfig()
	if err != nil {
		brk.log.Errorf("connecting to kubernetes failed: %v", err)
		brk.shutdown(kubernetesExitCode)
	}
	k8s, err := client.New(cfg, client.Options{})
	if err != nil {
		brk.log.Errorf("connecting to kubernetes failed: %v", err)
		brk.shutdown(kubernetesExitCode)
	}
	brk.k8sClient = k8s

	if config.HealthSrvAddr != "false" {
		if err := brk.healthSrv.Start(); err != nil {
			brk.shutdown(TelemetryExitCode)
		}
	}

	if config.TelemetryAddr != "false" {
		if err := brk.telClient.Start(ctx); err != nil {
			brk.shutdown(TelemetryExitCode)
		}
	}

	if err := brk.jsClient.Connect(); err != nil {
		brk.shutdown(JetStreamExitCode)
	}
	brk.healthSrv.Register(brk.jsClient)

	if err := brk.store.Open(brk.jsClient.ComponentsKV()); err != nil {
		brk.shutdown(ResourceStoreExitCode)
	}

	if err := brk.grpcSrv.Start(ctx, brk.jsClient.ComponentsKV()); err != nil {
		brk.shutdown(GRPCServerExitCode)
	}

	if err := brk.httpSrv.Start(); err != nil {
		brk.shutdown(HTTPServerExitCode)
	}

	brk.log.Infof("starting %d workers", config.NumWorkers)
	brk.wg.Add(config.NumWorkers + 3) // +3 for archiver and updaters
	go brk.startArchiver()
	go brk.startEventsKVUpdater()
	go brk.startCompsKVUpdater()
	for i := 0; i < config.NumWorkers; i++ {
		go brk.startWorker(i)
	}

	brk.log.Info("broker started")

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-ch

	brk.shutdown(0)
}

func (brk *broker) Subscribe(ctx context.Context, conf *SubscriptionConf) (ReplicaSubscription, error) {
	sub, err := brk.subMgr.Create(ctx, conf, brk.recvCh)
	if err != nil {
		return nil, err
	}

	if err := brk.updateCompReg(conf.Component, conf.CompReg); err != nil {
		return nil, err
	}

	err = brk.jsClient.PullEvents(sub)
	if err != nil {
		return nil, err
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

func (brk *broker) RecvEvent(rEvt *ReceivedEvent) error {
	if rEvt == nil || rEvt.Event == nil {
		return ErrEventInvalid
	}
	if rEvt.Event.Ttl <= 0 {
		return ErrEventTimeout
	}
	if rEvt.Subscription != nil && !rEvt.Subscription.IsActive() {
		return ErrSubCanceled
	}

	brk.recvCh <- rEvt

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
		case rEvt := <-brk.recvCh:
			if err := brk.routeEvt(log, rEvt); err != nil {
				_, gRPCErr := status.FromError(err)
				switch {
				case gRPCErr: // gRPC error
					err = fmt.Errorf("%w: %v", ErrComponentGone, err)

				case apierrors.IsNotFound(err): // Not found error from K8s API
					err = fmt.Errorf("%w: %v", ErrRouteNotFound, err)

				case !errors.Is(err, ErrKubeFox): // Unknown
					err = fmt.Errorf("%w: %v", ErrUnexpected, err)
				}

				log = log.WithEvent(rEvt.Event)
				if errors.Is(err, ErrUnexpected) {
					log.Error(err)
				} else {
					log.Info(err)
				}

				rEvt.Err(err)
			}
		case <-brk.ctx.Done():
			return
		}
	}
}

func (brk *broker) routeEvt(log *logkf.Logger, rEvt *ReceivedEvent) error {
	log = log.
		WithEvent(rEvt.Event).
		WithComponent(rEvt.Source).
		WithTarget(rEvt.Target)

	if rEvt.TTL() <= 0 {
		return fmt.Errorf("%w: event timed out while waiting for worker", ErrEventTimeout)
	}

	log.Debug("routing event from work queue")

	mEvt, err := brk.matchEvent(rEvt.Event)
	if err != nil {
		return err
	}

	log = log.WithTarget(rEvt.Target)
	log.Debug("matched event to target")

	if err := brk.checkMatchedEvent(mEvt); err != nil {
		return err
	}

	// TODO add policy checks

	var sub Subscription
	switch {
	case rEvt.Subscription != nil:
		sub = rEvt.Subscription

	case rEvt.Target.Equal(brk.httpSrv.Component()):
		sub = brk.httpSrv.Subscription()

	case rEvt.Target.Id != "":
		sub, _ = brk.subMgr.ReplicaSubscription(rEvt.Target)

	case rEvt.Target.Id == "":
		sub, _ = brk.subMgr.GroupSubscription(rEvt.Target)
	}

	var sendEvent SendEvent
	switch {
	case sub != nil:
		sendEvent = func(mEvt *kubefox.MatchedEvent) error {
			log.Debug("subscription found, sending event directly")
			err := sub.SendEvent(mEvt)
			brk.archiveCh <- rEvt
			return err
		}

	default:
		// Subscriptions not found, publish to JetStream.
		sendEvent = func(mEvt *kubefox.MatchedEvent) error {
			log.Debug("subscription not found, sending event to JetStream")

			if _, err := brk.jsClient.Publish(rEvt.Target.Subject(), mEvt.Event); err != nil {
				return err
			}

			// ack, err := brk.jsClient.Publish(rEvt.Target.Subject(), mEvt.Event)
			// if err != nil {
			// 	return err
			// }

			// go func() {
			// 	select {
			// 	case pubAck := <-ack.Ok():
			// 		brk.kvUpdateCh <- &IdSeq{EventId: rEvt.Event.Id, Sequence: pubAck.Sequence}
			// 	case err := <-ack.Err():
			// 		rEvt.Err(log.ErrorN("%w: %v", ErrUnexpected, err))
			// 	}
			// }()

			return nil
		}
	}

	if rEvt.TTL() <= 0 {
		return fmt.Errorf("%w: event timed out during routing", ErrEventTimeout)
	}

	rEvt.Flush()
	if err = sendEvent(mEvt); err != nil {
		return fmt.Errorf("%w: error sending event", err)
	}

	return nil
}

func (brk *broker) matchEvent(evt *kubefox.Event) (*kubefox.MatchedEvent, error) {
	if evt.Category == kubefox.Category_RESPONSE {
		if evt.Target == nil || !evt.Target.IsFull() {
			return nil, fmt.Errorf("%w: response target is missing required attribute", ErrRouteInvalid)
		}

		return &kubefox.MatchedEvent{Event: evt}, nil
	}

	var (
		matcher *matcher.EventMatcher
		err     error
	)
	evtCtx := evt.CheckContext()
	switch {
	case evtCtx.IsRelease():
		matcher, err = brk.store.GetRelMatchers()

	case evtCtx.IsDeployment():
		matcher, err = brk.store.GetDepMatcher(evtCtx)

	default:
		err = fmt.Errorf("%w: event missing deployment or environment context", ErrEventInvalid)
	}
	if err != nil {
		return nil, err
	}

	var (
		routeId int
	)
	route, matched := matcher.Match(evt)
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
		return nil, ErrRouteNotFound
	}

	var (
		rel     *v1alpha1.Release
		env     *v1alpha1.Environment
		envVars map[string]*kubefox.Val
	)
	switch {
	case evtCtx.Release != "":
		rel, err = brk.store.GetRelease(evtCtx.Release)
		if rel != nil {
			envVars = rel.Spec.Environment.Vars
		}

	case evtCtx.Environment != "":
		env, err = brk.store.GetEnvironment(evtCtx.Environment)
		if env != nil {
			envVars = env.Spec.Vars
		}
	}
	if err != nil {
		return nil, err
	}

	mEvt := &kubefox.MatchedEvent{
		Env:     make(map[string]*structpb.Value, len(envVars)),
		Event:   evt,
		RouteId: int64(routeId),
	}
	for k, v := range envVars {
		mEvt.Env[k] = v.Proto()
	}

	return mEvt, nil
}

func (brk *broker) checkMatchedEvent(mEvt *kubefox.MatchedEvent) error {
	evt := mEvt.Event
	if evt.Target == nil || evt.Target.Name == "" || evt.Context == nil {
		return ErrComponentMismatch
	}

	// Check if target is adapter.
	switch {
	case evt.Target.Equal(brk.httpSrv.Component()):
		return nil
	}

	var (
		depSpec v1alpha1.DeploymentSpec
	)
	switch {
	case evt.Context.IsRelease():
		if rel, err := brk.store.GetRelease(evt.Context.Release); err != nil {
			return fmt.Errorf("%w: unable to get release: %v", ErrUnexpected, err)
		} else {
			depSpec = rel.Spec.Deployment
		}

	case evt.Context.IsDeployment():
		if dep, err := brk.store.GetDeployment(evt.Context.Deployment); err != nil {
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
		return fmt.Errorf("%w: component not part of deployment", ErrComponentMismatch)

	case evt.Target.Commit == "" && mEvt.RouteId == kubefox.DefaultRouteId:
		evt.Target.Commit = depComp.Commit
		reg, err := brk.store.ComponentReg(evt.Target)
		if err != nil {
			return fmt.Errorf("%w: unable to get component registration: %v", ErrUnexpected, err)
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
		case rEvt := <-brk.archiveCh:
			log := log.WithEvent(rEvt.Event)
			log.Debug("publishing event to JetStream for archiving")

			if _, err := brk.jsClient.Publish(rEvt.Target.DirectSubject(), rEvt.Event); err != nil {
				log.Error(err)
			}

			// ack, err := brk.jsClient.Publish(rEvt.Target.DirectSubject(), rEvt.Event)
			// if err != nil {
			// 	// This event will be lost in time, never to be seen again :(
			// 	log.Warnf("unable to publish event to JetStream, event not archived: %v", err)
			// 	continue
			// }

			// evtId := rEvt.Id
			// go func() {
			// 	select {
			// 	case pubAck := <-ack.Ok():
			// 		brk.kvUpdateCh <- &IdSeq{EventId: evtId, Sequence: pubAck.Sequence}
			// 		return
			// 	case err := <-ack.Err():
			// 		// This event will be lost in time, never to be seen again :(
			// 		log.Warnf("received error from nats, event not archived: %v", err)
			// 	}
			// }()

		case <-brk.ctx.Done():
			return
		}
	}
}

func (brk *broker) startEventsKVUpdater() {
	log := brk.log.With(logkf.KeyWorker, "events-kv-updater")
	defer func() {
		log.Info("events-kv-updater stopped")
		brk.wg.Done()
	}()

	for {
		select {
		case idSeq := <-brk.kvUpdateCh:
			log.Debug("updating event kv with event id and sequence for easy event lookup")

			k := idSeq.EventId
			v := utils.UIntToByteArray(idSeq.Sequence)
			go func() {
				if _, err := brk.jsClient.EventsKV().Put(k, v); err != nil {
					log.With(logkf.KeyEventId, idSeq.EventId).
						Warnf("unable to update event kv, will not be able to lookup event: %v", err)
				}
			}()

		case <-brk.ctx.Done():
			return
		}
	}
}

func (brk *broker) startCompsKVUpdater() {
	log := brk.log.With(logkf.KeyWorker, "components-kv-updater")
	defer func() {
		log.Info("components-kv-updater stopped")
		brk.wg.Done()
	}()

	ticker := time.NewTicker(ComponentsTTL / 2)
	for {
		select {
		case <-ticker.C:
			for _, sub := range brk.subMgr.Subscriptions() {
				if err := brk.updateCompReg(sub.Component(), sub.ComponentReg()); err != nil {
					log.Error("error updating components key/value: %v", err)
				}
			}

		case <-brk.ctx.Done():
			return
		}
	}
}

func (brk *broker) updateCompReg(comp *kubefox.Component, reg *kubefox.ComponentReg) error {
	b, err := json.Marshal(reg)
	if err != nil {
		return err
	}

	if _, err := brk.jsClient.ComponentsKV().Put(comp.GroupKey(), b); err != nil {
		return err
	}

	return nil
}

func (brk *broker) shutdown(code int) {
	// TODO deal with inflight events when shutdown occurs

	brk.log.Info("broker shutting down")

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
