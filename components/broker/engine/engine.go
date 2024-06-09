// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package engine

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-logr/zapr"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/build"
	"github.com/xigxog/kubefox/components/broker/config"
	brktel "github.com/xigxog/kubefox/components/broker/telemetry"
	"github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/logkf"
	"github.com/xigxog/kubefox/telemetry"
	"github.com/xigxog/kubefox/utils"
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
	RecordTelemetry(*core.Component, *core.Telemetry)
	AuthorizeComponent(context.Context, *Metadata) error
	Subscribe(context.Context, *SubscriptionConf) (ReplicaSubscription, error)
	RecvEvent(evt *core.Event, receiver Receiver) *BrokerEventContext
	Component() *core.Component
}

type broker struct {
	comp *core.Component

	grpcSrv *GRPCServer

	natsClient *NATSClient
	k8sClient  client.Client

	healthSrv *brktel.HealthServer
	telClient *brktel.Client

	subMgr SubscriptionMgr
	recvCh chan *BrokerEventContext

	store *store

	ctx    context.Context
	cancel context.CancelFunc

	wg sync.WaitGroup

	log *logkf.Logger
}

func New() Engine {
	comp := core.NewPlatformComponent(api.ComponentTypeBroker, api.PlatformComponentBroker, build.Info.Hash)

	id := core.GenerateId()
	comp.Id, comp.BrokerId = id, id

	logkf.Global = logkf.Global.
		With(logkf.KeyBrokerId, comp.Id).
		With(logkf.KeyBrokerName, comp.Name)
	ctrl.SetLogger(zapr.NewLogger(logkf.Global.Unwrap().Desugar()))

	ctx, cancel := context.WithCancel(context.Background())
	brk := &broker{
		comp:      comp,
		healthSrv: brktel.NewHealthServer(),
		telClient: brktel.NewClient(),
		subMgr:    NewManager(),
		recvCh:    make(chan *BrokerEventContext),
		store:     NewStore(),
		ctx:       ctx,
		cancel:    cancel,
		log:       logkf.Global,
	}
	brk.grpcSrv = NewGRPCServer(brk)
	brk.natsClient = NewNATSClient(brk)

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

func (brk *broker) RecordTelemetry(comp *core.Component, tel *core.Telemetry) {
	brk.telClient.AddTelemetry(comp, tel)
}

func (brk *broker) Subscribe(ctx context.Context, conf *SubscriptionConf) (ReplicaSubscription, error) {
	sub, grpSub, err := brk.subMgr.Create(ctx, conf)
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

func (brk *broker) AuthorizeComponent(ctx context.Context, meta *Metadata) error {
	if meta.Platform != config.Platform {
		return fmt.Errorf("component provided incorrect platform")
	}

	parsed, err := jwt.ParseString(meta.Token)
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

	switch typ := api.ComponentType(meta.Component.Type); typ {
	case api.ComponentTypeKubeFox:
		if svcAccName != meta.Component.GroupKey() {
			return fmt.Errorf("service account name does not match component")
		}

		def, err := brk.store.ComponentDef(ctx, meta.Component)
		if err != nil || typ != def.Type || meta.Component.Hash != def.Hash {
			return fmt.Errorf("component not found")
		}

	// Platform Component.
	default:
		if svcAccName != utils.Join("-", meta.Platform, meta.Component.Name) {
			return fmt.Errorf("service account name does not match component")
		}

		p, err := brk.store.Platform(ctx)
		if err != nil {
			return err
		}

		found := false
		if p.DebugEnabled() && meta.Pod == "debug" {
			found = true
			brk.log.Warnf("Platform debug enabled, accepting connection from component '%s'",
				meta.Component.Name)
		}
		for _, c := range p.Status.Components {
			if c.Name == meta.Component.Name &&
				c.Hash == meta.Component.Hash &&
				c.PodName == meta.Pod {

				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("component not found")
		}
	}

	review := authv1.TokenReview{
		ObjectMeta: metav1.ObjectMeta{
			Name: svcAccName,
		},
		Spec: authv1.TokenReviewSpec{
			Token: meta.Token,
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

func (brk *broker) RecvEvent(evt *core.Event, receiver Receiver) *BrokerEventContext {
	parentCtx, cancel := context.WithCancelCause(context.Background())
	ctx, _ := context.WithTimeoutCause(parentCtx, evt.TTL(), core.ErrTimeout())

	span := telemetry.StartSpan(
		fmt.Sprintf("Route %s from %s", evt.Category, evt.Source.GroupKey()), evt.ParentSpan)

	evtCtx := &BrokerEventContext{
		Context:    ctx,
		Cancel:     cancel,
		Event:      evt,
		Receiver:   receiver,
		ReceivedAt: time.Now(),
		Span:       span,
	}

	go func() {
		brk.recvCh <- evtCtx
	}()

	return evtCtx
}

func (brk *broker) startWorker(id int) {
	log := brk.log.With(logkf.KeyWorker, fmt.Sprintf("worker-%d", id))
	defer func() {
		log.Info("worker stopped")
		brk.wg.Done()
	}()

	for {
		select {
		case ctx := <-brk.recvCh:
			ctx.Log = log

			if err := brk.routeEvent(ctx); err != nil {
				if apierrors.IsNotFound(err) {
					err = core.ErrNotFound(err)
				}

				kfErr := &core.Err{}
				if ok := errors.As(err, &kfErr); !ok {
					kfErr = core.ErrUnexpected(err)
				}

				ctx.Span.RecordErr(kfErr)

				switch kfErr.Code() {
				case core.CodeUnexpected:
					ctx.Log.Error(err)
				case core.CodeBrokerMismatch, core.CodeUnauthorized:
					ctx.Log.Warn(err)
				default:
					ctx.Log.Debug(err)
				}

				go func() {
					ctx.Cancel(kfErr)
				}()

			} else {
				ctx.Cancel(nil)
			}

			ctx.Span.SetEventAttributes(ctx.Event)
			ctx.Span.End()

			brk.telClient.AddSpans(brk.comp, ctx.Span)

		case <-brk.ctx.Done():
			return
		}
	}
}

func (brk *broker) routeEvent(ctx *BrokerEventContext) (err error) {
	ctx.Log.Debugf("routing event from receiver '%s'", ctx.Receiver)

	routeSpan := ctx.Span.StartChildSpan(
		"Send "+ctx.Event.Category.String(),
		telemetry.Attr(telemetry.AttrKeyEventId, ctx.Event.Id),
		telemetry.Attr(telemetry.AttrKeyEventSourceName, ctx.Event.Source.Key()))
	defer routeSpan.End()

	findSpan := routeSpan.StartChildSpan("Find Target")
	if err = brk.validateEvent(ctx); err == nil { //success
		err = brk.findTarget(ctx)
	}
	if err != nil {
		findSpan.End(err)
		return
	}
	findSpan.End()

	// Update log and span attributes after matching.
	routeSpan.Name += " to " + ctx.Event.Target.GroupKey()

	ctx.Log = ctx.Log.WithEvent(ctx.Event)
	ctx.Log.Debugf("matched event to target '%s'", ctx.Event.Target.GroupKey())

	sendSpan := routeSpan.StartChildSpan("Send Event")

	sub, found := brk.subMgr.Subscription(ctx.Event.Target, ctx.TargetAdapter)
	switch {
	case found:
		// Found component subscribed via gRPC.
		sendSpan.Name = "Send gRPC event"
		ctx.Log.Debug("subscription found, sending event with gRPC")
		err = sub.SendEvent(ctx)

	case ctx.Receiver != ReceiverNATS && ctx.Event.Target.BrokerId != brk.comp.Id:
		// Component not found locally, send via NATS.
		sendSpan.Name = "Send NATS event"
		ctx.Log.Debug("subscription not found, sending event with nats")
		err = brk.natsClient.Publish(ctx.Event.Target.Subject(), ctx.Event)

	default:
		err = core.ErrComponentGone()
	}
	sendSpan.End(err)

	return
}

func (brk *broker) validateEvent(ctx *BrokerEventContext) error {
	if ctx.TTL() <= 0 {
		return core.ErrTimeout()
	}

	if ctx.Event.Source == nil || !ctx.Event.Source.IsComplete() {
		return core.ErrInvalid(fmt.Errorf("event source is invalid"))
	}

	if ctx.Event.Category == core.Category_RESPONSE && !ctx.Event.Target.IsComplete() {
		return core.ErrInvalid(fmt.Errorf("response target is missing required attribute"))
	}

	switch ctx.Receiver {
	case ReceiverNATS:
		if ctx.Event.Target != nil &&
			ctx.Event.Target.BrokerId != "" &&
			ctx.Event.Target.BrokerId != brk.comp.Id {

			return core.ErrBrokerMismatch(fmt.Errorf("event target broker id is %s", ctx.Event.Target.BrokerId))
		}

	case ReceiverGRPCServer:
		if ctx.Event.Target != nil && !ctx.Event.Target.IsComplete() && !ctx.Event.Target.IsNameOnly() {
			return core.ErrInvalid(fmt.Errorf("event target is invalid"))
		}

		// If a valid context is not present reject.
		if ctx.Context == nil || ctx.Event.Context.Platform != config.Platform ||
			(ctx.Event.Context.VirtualEnvironment == "" && ctx.Event.Context.AppDeployment != "") ||
			(ctx.Event.Context.VirtualEnvironment != "" && ctx.Event.Context.AppDeployment == "") ||
			(ctx.Event.Context.VirtualEnvironment == "" && ctx.Event.Context.AppDeployment == "" &&
				ctx.Event.Context.ReleaseManifest != "") {

			return core.ErrInvalid(fmt.Errorf("event context is invalid"))
		}
	}

	return nil
}

func (brk *broker) findTarget(ctx *BrokerEventContext) (err error) {
	if ctx.Event.HasContext() {
		if err := brk.store.AttachEventContext(ctx); err != nil {
			return err
		}

		if ctx.Event.Target != nil {
			if ctx.Event.Category == core.Category_RESPONSE && ctx.Event.Target.IsComplete() {
				_, err := ctx.AppDeployment.GetDefinition(ctx.Event.Target)
				if err != nil && !brk.store.IsGenesisAdapter(ctx, ctx.Event.Target) {
					return err
				}

				return nil
			}

			if typ := api.ComponentType(ctx.Event.Target.Type); typ.IsAdapter() {
				if !ctx.AppDeployment.HasDependency(ctx.Event.Target.Name, typ) {
					return core.ErrComponentMismatch(fmt.Errorf("target adapter not declared as dependency"))
				}
				if ctx.TargetAdapter, err = brk.store.Adapter(ctx, ctx.Event.Target.Name, typ); err != nil {
					return err
				}
				// TODO: Data on `ctx.Event.Target` still need to be filled out, however,
				// at this point in time we cannot fill in all fields of data that are
				// required to match to a group subscription.

				spec, err := ctx.TargetAdapter.Resolve(ctx.Data)
				if err != nil {
					return err
				}
				ctx.Event.SetSpec(spec)

				return nil
			}
		}

		matcher, err := brk.store.DeploymentMatcher(ctx)
		if err != nil {
			return err
		}

		route, matched := matcher.Match(ctx.Event)
		switch {
		case matched:
			ctx.RouteId = int64(route.Id)
			ctx.Event.SetRoute(route)

		case ctx.Event.Target != nil && ctx.Event.Target.Type == string(api.ComponentTypeKubeFox):
			if ctx.Event.Target.Hash == "" || ctx.Event.Target.App == "" {
				def, err := ctx.AppDeployment.GetDefinition(ctx.Event.Target)
				if err != nil {
					return err
				}

				ctx.Event.Target.Hash = def.Hash
				ctx.Event.Target.App = ctx.AppDeployment.Spec.AppName
			}

			ctx.RouteId = api.DefaultRouteId

		default:
			return core.ErrRouteNotFound()
		}

	} else {
		// Genesis event for Release.
		if ctx.Event.Target != nil {
			return core.ErrInvalid(fmt.Errorf("genesis event target is set"))
		}
		if !brk.store.IsGenesisAdapter(ctx, ctx.Event.Source) {
			return core.ErrInvalid(fmt.Errorf("genesis event source is not a genesis adapter"))
		}

		matcher, err := brk.store.ReleaseMatcher(ctx)
		if err != nil {
			return err
		}

		route, matched := matcher.Match(ctx.Event)
		if !matched {
			return core.ErrRouteNotFound()
		}

		ctx.RouteId = int64(route.Id)
		ctx.Event.SetRoute(route)
		ctx.Log.DebugInterface("route:", route)
		if err := brk.store.AttachEventContext(ctx); err != nil {
			return err
		}
	}

	_, err = ctx.AppDeployment.GetDefinition(ctx.Event.Target)
	if err != nil && !brk.store.IsGenesisAdapter(ctx, ctx.Event.Target) {
		return err
	}

	_, err = ctx.AppDeployment.GetDefinition(ctx.Event.Source)
	if err != nil && !brk.store.IsGenesisAdapter(ctx, ctx.Event.Source) {
		return err
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
