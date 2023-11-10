package engine

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	common "github.com/xigxog/kubefox/api/kubernetes"
	kubefox "github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/logkf"
)

type SubscriptionMgr interface {
	Create(ctx context.Context, cfg *SubscriptionConf, recvCh chan *BrokerEvent) (ReplicaSubscription, GroupSubscription, error)
	Subscription(comp *kubefox.Component) (Subscription, bool)
	ReplicaSubscription(comp *kubefox.Component) (ReplicaSubscription, bool)
	GroupSubscription(comp *kubefox.Component) (GroupSubscription, bool)
	Subscriptions() []ReplicaSubscription
	Close()
}

type Subscription interface {
	SendEvent(evt *BrokerEvent) error
	IsActive() bool
	Context() context.Context
}

type GroupSubscription interface {
	Subscription
}

type SubscriptionConf struct {
	Component     *kubefox.Component
	ComponentSpec *common.ComponentDefinition
	SendFunc      SendEvent
	EnableGroup   bool
}

type ReplicaSubscription interface {
	Subscription
	Component() *kubefox.Component
	ComponentSpec() *common.ComponentDefinition
	IsGroupEnabled() bool
	Cancel(err error)
	Err() error
}

type subscriptionMgr struct {
	subMap map[string]*subscription
	grpMap map[string]*subscriptionGroup

	mutex sync.RWMutex

	log *logkf.Logger
}

type subscriptionGroup struct {
	subMap map[string]bool
	sendCh chan *evtRespCh

	ctx    context.Context
	cancel context.CancelFunc
}

type subscription struct {
	comp     *kubefox.Component
	compSpec *common.ComponentDefinition
	mgr      *subscriptionMgr

	sendFunc   SendEvent
	recvCh     chan *BrokerEvent
	sendCh     chan *evtRespCh
	grpEnabled bool

	ctx      context.Context
	cancel   context.CancelCauseFunc
	canceled atomic.Bool
}

type evtRespCh struct {
	mEvt   *BrokerEvent
	respCh chan *sendResp
}

type sendResp struct {
	Err error
}

func NewManager() SubscriptionMgr {
	return &subscriptionMgr{
		subMap: make(map[string]*subscription),
		grpMap: make(map[string]*subscriptionGroup),
		log:    logkf.Global,
	}
}

func (mgr *subscriptionMgr) Create(ctx context.Context, cfg *SubscriptionConf, recvCh chan *BrokerEvent) (ReplicaSubscription, GroupSubscription, error) {
	if err := checkComp(cfg.Component); err != nil {
		return nil, nil, err
	}

	if sub, found := mgr.ReplicaSubscription(cfg.Component); found {
		mgr.log.WithComponent(cfg.Component).Warn("subscription for component already exists")
		sub.Cancel(nil)
	}

	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	var grpSub *subscriptionGroup
	if cfg.EnableGroup {
		s, found := mgr.grpMap[cfg.Component.GroupKey()]
		if !found {
			ctx, cancel := context.WithCancel(context.Background())
			s = &subscriptionGroup{
				subMap: make(map[string]bool),
				sendCh: make(chan *evtRespCh),
				ctx:    ctx,
				cancel: cancel,
			}
			mgr.grpMap[cfg.Component.GroupKey()] = s
		}
		s.subMap[cfg.Component.Id] = true
		grpSub = s
	}

	subCtx, subCancel := context.WithCancelCause(ctx)
	sub := &subscription{
		comp:       cfg.Component,
		compSpec:   cfg.ComponentSpec,
		mgr:        mgr,
		sendFunc:   cfg.SendFunc,
		recvCh:     recvCh,
		grpEnabled: cfg.EnableGroup,
		ctx:        subCtx,
		cancel:     subCancel,
	}
	if grpSub != nil {
		sub.sendCh = grpSub.sendCh
		go sub.processSendChan()
	}
	mgr.subMap[cfg.Component.Id] = sub

	return sub, grpSub, nil
}

func (mgr *subscriptionMgr) Subscription(comp *kubefox.Component) (Subscription, bool) {
	if sub, found := mgr.ReplicaSubscription(comp); found {
		return sub, true
	}

	return mgr.GroupSubscription(comp)
}

func (mgr *subscriptionMgr) ReplicaSubscription(comp *kubefox.Component) (ReplicaSubscription, bool) {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()

	sub, found := mgr.subMap[comp.Id]
	if !found || sub == nil || !sub.IsActive() {
		return nil, false
	}

	return sub, true
}

func (mgr *subscriptionMgr) GroupSubscription(comp *kubefox.Component) (GroupSubscription, bool) {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()

	grp, found := mgr.grpMap[comp.GroupKey()]
	if !found || grp == nil || !grp.IsActive() {
		return nil, false
	}

	return grp, true
}

func (mgr *subscriptionMgr) Subscriptions() []ReplicaSubscription {
	list := make([]ReplicaSubscription, 0, len(mgr.subMap))
	for _, s := range mgr.subMap {
		if sub, found := mgr.ReplicaSubscription(s.Component()); found {
			list = append(list, sub)
		}
	}

	return list
}

func (mgr *subscriptionMgr) Close() {
	mgr.log.Info("subscription manager closing")

	for _, sub := range mgr.subMap {
		sub.Cancel(nil)
	}

	mgr.log.Debug("subscription manager closed")
}

func (mgr *subscriptionMgr) cancel(sub *subscription, err error) {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	log := mgr.log.WithComponent(sub.comp)
	log.Debug("canceling component subscription")
	sub.cancel(err)

	grp := mgr.grpMap[sub.comp.GroupKey()]
	if grp != nil {
		delete(grp.subMap, sub.comp.Id)
		if len(grp.subMap) == 0 {
			log.Debug("component group is empty, canceling")
			grp.cancel()
			delete(mgr.grpMap, sub.comp.GroupKey())
		}
	}

	delete(mgr.subMap, sub.comp.Id)
}

func (grp *subscriptionGroup) SendEvent(evt *BrokerEvent) error {
	respCh := make(chan *sendResp)
	grp.sendCh <- &evtRespCh{mEvt: evt, respCh: respCh}
	resp := <-respCh

	return resp.Err
}

func (grp *subscriptionGroup) IsActive() bool {
	return len(grp.subMap) > 0
}

func (sub *subscriptionGroup) Context() context.Context {
	return sub.ctx
}

func (sub *subscription) SendEvent(evt *BrokerEvent) error {
	if err := sub.sendFunc(evt); err != nil {
		return err
	}

	return nil
}

func (sub *subscription) IsActive() bool {
	return !sub.canceled.Load()
}

func (sub *subscription) Component() *kubefox.Component {
	return sub.comp
}

func (sub *subscription) ComponentSpec() *common.ComponentDefinition {
	return sub.compSpec
}

func (sub *subscription) IsGroupEnabled() bool {
	return sub.grpEnabled
}

func (sub *subscription) Context() context.Context {
	return sub.ctx
}

func (sub *subscription) Cancel(err error) {
	if sub.canceled.Swap(true) {
		return
	}

	sub.mgr.cancel(sub, err)
}

func (sub *subscription) Err() error {
	cause := context.Cause(sub.ctx)
	if cause != sub.ctx.Err() {
		return cause
	}

	return nil
}

func (sub *subscription) processSendChan() {
	for {
		select {
		case evtRespCh := <-sub.sendCh:
			err := sub.SendEvent(evtRespCh.mEvt)
			evtRespCh.respCh <- &sendResp{Err: err}

		case <-sub.ctx.Done():
			return
		}
	}
}

func checkComp(comp *kubefox.Component) error {
	if comp.Name == "" {
		return fmt.Errorf("component is missing name")
	}
	if comp.Id == "" {
		return fmt.Errorf("component is missing id")
	}
	if comp.Commit == "" {
		return fmt.Errorf("component is missing commit")
	}

	return nil
}
