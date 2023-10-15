package engine

import (
	"context"
	"encoding/json"
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
	"github.com/xigxog/kubefox/libs/core/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/logkf"
	"github.com/xigxog/kubefox/libs/core/utils"

	"google.golang.org/protobuf/types/known/structpb"
	authv1 "k8s.io/api/authentication/v1"
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

	if rEvt.Context == nil {
		rEvt.Context = context.Background()
	}
	rEvt.RecvTime = time.Now().UnixMicro()

	brk.recvCh <- rEvt

	return nil
}

func (brk *broker) startWorker(id int) {
	log := brk.log.With("worker", fmt.Sprintf("worker-%d", id))
	defer func() {
		log.Info("worker stopped")
		brk.wg.Done()
	}()

	for {
		select {
		case evt := <-brk.recvCh:
			brk.routeEvt(log, evt)
		case <-brk.ctx.Done():
			return
		}
	}
}

func (brk *broker) routeEvt(log *logkf.Logger, rEvt *ReceivedEvent) {
	evt := rEvt.Event
	evt.Ttl = evt.Ttl - (time.Now().UnixMicro() - rEvt.RecvTime)
	if rEvt.Event.Ttl <= 0 {
		log.DebugEw("event timed out while waiting for worker", evt)
		rEvt.Err(ErrEventTimeout)
		return
	}

	log = log.WithEvent(evt)
	log.Debug("routing event from work queue")
	start := time.Now().UnixMicro()

	mEvt, err := brk.matchEvt(log, evt)
	if err != nil {
		log.Debug(err)
		rEvt.Err(err)
		return
	}

	log = log.WithComponent(evt.Target)
	log.Debug("matched event")

	// TODO add policy hook

	var sub Subscription
	switch {
	case rEvt.Subscription != nil:
		sub = rEvt.Subscription

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
			ack, err := brk.jsClient.Publish(evt.Target.Subject(), mEvt.Event)
			if err != nil {
				return err
			}

			go func() {
				select {
				case pubAck := <-ack.Ok():
					brk.kvUpdateCh <- &IdSeq{EventId: rEvt.Event.Id, Sequence: pubAck.Sequence}
				case err := <-ack.Err():
					rEvt.Err(log.ErrorN("%w: %v", ErrUnexpected, err))
				}
			}()

			return nil
		}
	}

	evt.Ttl = evt.Ttl - (time.Now().UnixMicro() - start)
	if rEvt.Event.Ttl <= 0 {
		log.Debug("event timed out while routing it")
		rEvt.Err(ErrEventTimeout)
		return
	}

	if err := sendEvent(mEvt); err != nil {
		rEvt.Err(log.ErrorN("%w: %v", ErrUnexpected, err))
	}
}

func (brk *broker) matchEvt(log *logkf.Logger, evt *kubefox.Event) (*kubefox.MatchedEvent, error) {
	if evt.Target != nil {
		if brk.httpSrv.Component().Equal(evt.Target) {
			return &kubefox.MatchedEvent{Event: evt}, nil
		}

		if evt.Category == kubefox.Category_CATEGORY_RESPONSE &&
			evt.Target.Name != "" && evt.Target.Commit != "" && evt.Target.Id != "" {

			return &kubefox.MatchedEvent{Event: evt}, nil
		}
	}

	var (
		matcher Matcher
		err     error
	)
	switch {
	case evt.Deployment == "" && evt.Environment == "" && evt.Release == "":
		matcher, err = brk.store.GetRelMatchers()

	case evt.Deployment != "" && evt.Environment != "" && evt.Release == "":
		matcher, err = brk.store.GetDepMatcher(evt.Deployment, evt.Environment)

	default:
		err = fmt.Errorf("%w: missing deployment or environment", ErrEventInvalid)
	}
	if err != nil {
		return nil, err
	}

	mEvt := matcher.Match(evt)
	if mEvt == nil {
		return nil, ErrRouteNotFound
	}

	var (
		rel     *v1alpha1.Release
		env     *v1alpha1.Environment
		envVars map[string]*kubefox.Val
	)
	switch {
	case evt.Release != "":
		rel, err = brk.store.GetRelease(evt.Release)
		if rel != nil {
			envVars = rel.Spec.Environment.Vars
		}

	case evt.Environment != "":
		env, err = brk.store.GetEnvironment(evt.Environment)
		if env != nil {
			envVars = env.Spec.Vars
		}
	}
	if err != nil {
		return nil, err
	}

	mEvt.Env = make(map[string]*structpb.Value, len(envVars))
	for k, v := range envVars {
		mEvt.Env[k] = v.Proto()
	}

	return mEvt, nil
}

func (brk *broker) startArchiver() {
	log := brk.log.With("worker", "archiver")
	defer func() {
		log.Info("archiver stopped")
		brk.wg.Done()
	}()

	errMsg := "unable to publish event to JetStream, event not archived: %v"
	for {
		select {
		case rEvt := <-brk.archiveCh:
			log.Debug("publishing event to JetStream for archiving")

			ack, err := brk.jsClient.Publish(rEvt.Event.Target.DirectSubject(), rEvt.Event)
			if err != nil {
				// This event will be lost in time, never to be seen again :(
				log.ErrorEw(errMsg, rEvt.Event, err)
			}
			go func() {
				select {
				case pubAck := <-ack.Ok():
					brk.kvUpdateCh <- &IdSeq{EventId: rEvt.Event.Id, Sequence: pubAck.Sequence}
					return
				case err := <-ack.Err():
					// This event will be lost in time, never to be seen again :(
					log.ErrorEw(errMsg, rEvt.Event, err)
				}
			}()

		case <-brk.ctx.Done():
			return
		}
	}
}

func (brk *broker) startEventsKVUpdater() {
	log := brk.log.With("worker", "events-kv-updater")
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
					log.With("eventId", idSeq.EventId).Errorf("unable to update event kv: %v")
				}
			}()

		case <-brk.ctx.Done():
			return
		}
	}
}

func (brk *broker) startCompsKVUpdater() {
	log := brk.log.With("worker", "components-kv-updater")
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
