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
	"sync"
	"time"

	"github.com/xigxog/kubefox/api"
	common "github.com/xigxog/kubefox/api/kubernetes"
	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/logkf"
	"github.com/xigxog/kubefox/telemetry"
	"google.golang.org/protobuf/types/known/structpb"
)

type Receiver int

const (
	ReceiverNATS Receiver = iota
	ReceiverGRPCServer
	ReceiverHTTPServer
	ReceiverHTTPClient
)

type SendEvent func(*BrokerEventContext) error

type BrokerEventContext struct {
	context.Context

	Key string

	Receiver   Receiver
	ReceivedAt time.Time

	Event           *core.Event
	AppDeployment   *v1alpha1.AppDeployment
	ReleaseManifest *v1alpha1.ReleaseManifest
	VirtualEnv      *v1alpha1.VirtualEnvironment

	Data *api.Data

	RouteId int64

	TargetAdapter common.Adapter

	Span   *telemetry.Span
	Log    *logkf.Logger
	Cancel context.CancelCauseFunc

	tick  time.Time
	mutex sync.Mutex
}

func (ctx *BrokerEventContext) TTL() time.Duration {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	if ctx.tick.Equal(time.Time{}) {
		ctx.tick = ctx.ReceivedAt
	}

	ctx.Event.ReduceTTL(ctx.tick)
	ctx.tick = time.Now()

	return ctx.Event.TTL()
}

func (ctx *BrokerEventContext) MatchedEvent() *core.MatchedEvent {
	m := &core.MatchedEvent{
		Event:   ctx.Event,
		RouteId: ctx.RouteId,
	}

	if ctx.Event == nil || ctx.AppDeployment == nil || ctx.Data == nil || ctx.Data.Vars == nil {
		return m
	}

	def, err := ctx.AppDeployment.GetDefinition(ctx.Event.Target)
	if err != nil {
		return m
	}

	// Only include vars that target declared as dependencies.
	m.Env = make(map[string]*structpb.Value, len(def.EnvVarSchema))
	for k := range def.EnvVarSchema {
		m.Env[k] = ctx.Data.Vars[k].Proto()
	}

	return m
}

func (ctx *BrokerEventContext) Value(key any) any {
	return ctx.Context.Value(key)
}

func (ctx *BrokerEventContext) CoreErr() *core.Err {
	err := ctx.Err()
	if err == nil {
		return nil
	}

	coreErr := &core.Err{}
	if ok := errors.As(err, &coreErr); !ok {
		return core.ErrUnexpected(err)
	}

	return coreErr
}

func (ctx *BrokerEventContext) Err() error {
	cause := context.Cause(ctx)

	// Canceled indicates routing completed without issue.
	if errors.Is(cause, context.Canceled) {
		return nil
	}

	return cause
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
