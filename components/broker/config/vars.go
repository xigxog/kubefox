// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package config

import "time"

var (
	Instance  string
	Platform  string
	Namespace string

	MaxEventSize      int64
	NumWorkers        int
	TelemetryInterval time.Duration

	LogFormat string
	LogLevel  string

	GRPCSrvAddr   string
	HealthSrvAddr string

	NATSAddr          string
	VaultURL          string
	TelemetryAddr     string
	TelemetryProtocol string
)
