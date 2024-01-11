// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package engine

import (
	"fmt"
	"sync"
	"time"

	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/core"
	"google.golang.org/protobuf/types/known/structpb"
)

type Receiver int

const (
	ReceiverNATS Receiver = iota
	ReceiverGRPCServer
	ReceiverHTTPServer
	ReceiverHTTPClient
)

type SendEvent func(*BrokerEvent) error

type Adapters map[string]api.Adapter

type BrokerEvent struct {
	*core.Event

	ContextKey   string
	Data         *api.Data
	DataChecksum string
	AppDep       *v1alpha1.AppDeployment
	RouteId      int64

	TargetAdapter api.Adapter
	Adapters      Adapters

	Receiver   Receiver
	ReceivedAt time.Time
	DoneCh     chan *core.Err

	tick  time.Time
	mutex sync.Mutex
}

func (evt *BrokerEvent) TTL() time.Duration {
	evt.mutex.Lock()
	defer evt.mutex.Unlock()

	if evt.tick.Equal(time.Time{}) {
		evt.tick = evt.ReceivedAt
	}

	evt.ReduceTTL(evt.tick)
	evt.tick = time.Now()

	return evt.Event.TTL()
}

func (evt *BrokerEvent) Done() chan *core.Err {
	return evt.DoneCh
}

func (evt *BrokerEvent) MatchedEvent() *core.MatchedEvent {
	var env map[string]*structpb.Value
	if evt.Data != nil && evt.Data.Vars != nil {
		env = make(map[string]*structpb.Value, len(evt.Data.Vars))
		for k, v := range evt.Data.Vars {
			env[k] = v.Proto()
		}
	}

	return &core.MatchedEvent{
		Event:   evt.Event,
		RouteId: evt.RouteId,
		Env:     env,
	}
}

func (a Adapters) Set(val api.Adapter) {
	if val == nil {
		return
	}

	key := fmt.Sprintf("%s-%s", val.GetName(), val.GetComponentType())
	a[key] = val
}

func (a Adapters) Get(name string, typ api.ComponentType) (api.Adapter, bool) {
	key := fmt.Sprintf("%s-%s", name, typ)
	val, found := a[key]

	return val, found
}

func (a Adapters) GetByComponent(c *core.Component) (api.Adapter, bool) {
	if c == nil {
		return nil, false
	}

	return a.Get(c.Name, api.ComponentType(c.Type))
}

func (r Receiver) String() string {
	switch r {
	case ReceiverNATS:
		return "nats-client"
	case ReceiverGRPCServer:
		return "grpc-server"
	case ReceiverHTTPServer:
		return "http-server"
	case ReceiverHTTPClient:
		return "http-client"
	default:
		return "unknown"
	}
}
