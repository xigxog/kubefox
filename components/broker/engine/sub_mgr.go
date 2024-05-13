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
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/logkf"
)

type SubscriptionMgr interface {
	Create(ctx context.Context, cfg *SubscriptionConf) (ReplicaSubscription, GroupSubscription, error)
	Subscription(comp *core.Component) (Subscription, bool)
	Close()
}

type Subscription interface {
	SendEvent(evt *BrokerEventContext) error
	IsActive() bool
	Context() context.Context
}

type GroupSubscription interface {
	Subscription
}

type ReplicaSubscription interface {
	Subscription

	Component() *core.Component
	ComponentDef() *api.ComponentDefinition
	IsGroupEnabled() bool
	Cancel(err error)
	Err() error
}

type SubscriptionConf struct {
	Component    *core.Component
	ComponentDef *api.ComponentDefinition
	SendFunc     SendEvent
	EnableGroup  bool
}

type subscriptionMgr struct {
	subMap map[string]*subscription
	grpMap map[string]*groupSubscription

	mutex sync.RWMutex

	log *logkf.Logger
}

type groupSubscription struct {
	subMap map[string]bool
	sendCh chan *evtRespCh

	ctx    context.Context
	cancel context.CancelFunc
}

type subscription struct {
	comp    *core.Component
	compDef *api.ComponentDefinition
	mgr     *subscriptionMgr

	sendFunc   SendEvent
	sendCh     chan *evtRespCh
	grpEnabled bool

	ctx      context.Context
	cancel   context.CancelCauseFunc
	canceled atomic.Bool
}

type evtRespCh struct {
	mEvt   *BrokerEventContext
	respCh chan *sendResp
}

type sendResp struct {
	Err error
}

func NewManager() SubscriptionMgr {
	return &subscriptionMgr{
		subMap: make(map[string]*subscription),
		grpMap: make(map[string]*groupSubscription),
		log:    logkf.Global,
	}
}

func (mgr *subscriptionMgr) Create(ctx context.Context, cfg *SubscriptionConf) (ReplicaSubscription, GroupSubscription, error) {
	switch {
	case cfg.Component == nil:
		return nil, nil, fmt.Errorf("component is missing")
	case cfg.Component.Name == "":
		return nil, nil, fmt.Errorf("component is missing name")
	case cfg.Component.Hash == "":
		return nil, nil, fmt.Errorf("component is missing hash")
	case cfg.Component.Id == "":
		return nil, nil, fmt.Errorf("component is missing id")
	}

	if sub, found := mgr.ReplicaSubscription(cfg.Component); found {
		mgr.log.WithComponent(cfg.Component).Warn("subscription for component already exists")
		sub.Cancel(nil)
	}

	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	var grpSub *groupSubscription
	if cfg.EnableGroup {
		s, found := mgr.grpMap[cfg.Component.GroupKey()]
		if !found {
			ctx, cancel := context.WithCancel(context.Background())
			s = &groupSubscription{
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
		compDef:    cfg.ComponentDef,
		mgr:        mgr,
		sendFunc:   cfg.SendFunc,
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

func (mgr *subscriptionMgr) Subscription(comp *core.Component) (Subscription, bool) {
	if comp == nil {
		return nil, false
	}

	if sub, found := mgr.ReplicaSubscription(comp); found {
		return sub, true
	}

	return mgr.GroupSubscription(comp)
}

func (mgr *subscriptionMgr) ReplicaSubscription(comp *core.Component) (ReplicaSubscription, bool) {
	if comp == nil || comp.Id == "" {
		return nil, false
	}

	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()

	sub, found := mgr.subMap[comp.Id]
	if !found || sub == nil || !sub.IsActive() {
		return nil, false
	}

	return sub, true
}

func (mgr *subscriptionMgr) GroupSubscription(comp *core.Component) (GroupSubscription, bool) {
	if comp == nil || comp.Name == "" || comp.Hash == "" {
		return nil, false
	}

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

func (grp *groupSubscription) SendEvent(evt *BrokerEventContext) error {
	respCh := make(chan *sendResp)
	grp.sendCh <- &evtRespCh{mEvt: evt, respCh: respCh}
	resp := <-respCh

	return resp.Err
}

func (grp *groupSubscription) IsActive() bool {
	return len(grp.subMap) > 0
}

func (sub *groupSubscription) Context() context.Context {
	return sub.ctx
}

func (sub *subscription) SendEvent(evt *BrokerEventContext) error {
	if err := sub.sendFunc(evt); err != nil {
		return err
	}

	return nil
}

func (sub *subscription) IsActive() bool {
	return !sub.canceled.Load()
}

func (sub *subscription) Component() *core.Component {
	return sub.comp
}

func (sub *subscription) ComponentDef() *api.ComponentDefinition {
	return sub.compDef
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
