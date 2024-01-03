// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package server

import (
	"time"

	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/core"
)

var (
	Platform                  string
	HTTPAddr, HTTPSAddr       string
	BrokerAddr, HealthSrvAddr string
	EventTimeout              time.Duration
	MaxEventSize              int64
)

var (
	Component    = &core.Component{Id: core.GenerateId(), Type: string(api.ComponentTypeHTTPAdapter)}
	ComponentDef = &api.ComponentDefinition{Type: api.ComponentTypeHTTPAdapter}
)
