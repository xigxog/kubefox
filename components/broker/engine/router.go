// Copyright 2024 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package engine

import "github.com/xigxog/kubefox/core"

// TODO break out routing from engine
type Router interface {
	RouteEvent(*core.Event, Receiver) *BrokerEventContext
}

type router struct {
}

func (brk *broker) RouteEvent(evt *core.Event, recv Receiver) *BrokerEventContext {
	return nil
}
