// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package vault

import (
	"context"
	"sync"

	"github.com/xigxog/kubefox/k8s"
	"github.com/xigxog/kubefox/vault"
	v1 "k8s.io/api/core/v1"
)

var (
	Opts      vault.ClientOptions
	Namespace string
	K8sClient k8s.Client

	globalClient *vault.Client
	mutex        sync.Mutex
)

func GetClient(ctx context.Context) (*vault.Client, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if globalClient != nil {
		return globalClient, nil
	}

	cm := &v1.ConfigMap{}
	if err := K8sClient.Get(ctx, k8s.Key(Namespace, Opts.Instance+"-root-ca"), cm); err != nil {
		return nil, err
	}
	Opts.CACertBytes = []byte(cm.Data["ca.crt"])

	var err error
	globalClient, err = vault.New(Opts)

	return globalClient, err
}
