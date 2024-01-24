// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package controller

import (
	"fmt"

	"github.com/xigxog/kubefox/build"
)

const (
	NATSImage = "ghcr.io/xigxog/nats:2.9.24"
)

var (
	BrokerImage    = fmt.Sprintf("ghcr.io/xigxog/kubefox/broker:%s", build.Info.Version)
	BootstrapImage = fmt.Sprintf("ghcr.io/xigxog/kubefox/bootstrap:%s", build.Info.Version)
	HTTPSrvImage   = fmt.Sprintf("ghcr.io/xigxog/kubefox/httpsrv:%s", build.Info.Version)
)
