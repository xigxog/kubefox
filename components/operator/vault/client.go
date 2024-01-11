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
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	vapi "github.com/hashicorp/vault/api"
	vauth "github.com/hashicorp/vault/api/auth/kubernetes"
	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/k8s"
	v1 "k8s.io/api/core/v1"
)

type client struct {
	*vapi.Client
}

type vaultSecret struct {
	Data *api.Data `json:"data"`
}

var (
	Instance  string
	Namespace string
	URL       string
	K8sClient k8s.Client

	globalClient *client
	mutex        sync.Mutex
)

func Client() (*client, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if globalClient != nil {
		return globalClient, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	cm := &v1.ConfigMap{}
	if err := K8sClient.Get(ctx, k8s.Key(Namespace, Instance+"-root-ca"), cm); err != nil {
		return nil, err
	}

	cfg := vapi.DefaultConfig()
	cfg.Address = URL
	cfg.MaxRetries = 3
	cfg.HttpClient.Timeout = time.Second * 5
	cfg.ConfigureTLS(&vapi.TLSConfig{
		CACertBytes: []byte(cm.Data["ca.crt"]),
	})

	c, err := vapi.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(api.PathSvcAccToken)
	if err != nil {
		return nil, err
	}
	token := vauth.WithServiceAccountToken(string(b))
	auth, err := vauth.NewKubernetesAuth("kubefox-operator", token)
	if err != nil {
		return nil, err
	}
	authInfo, err := c.Auth().Login(ctx, auth)
	if err != nil {
		return nil, err
	}
	if authInfo == nil {
		return nil, fmt.Errorf("error logging in with kubernetes auth: no auth info was returned")
	}

	watcher, err := c.NewLifetimeWatcher(&vapi.LifetimeWatcherInput{Secret: authInfo})
	if err != nil {
		return nil, fmt.Errorf("error starting Vault token renewer: %w", err)
	}
	go watcher.Start()

	globalClient = &client{
		Client: c,
	}

	return globalClient, nil

}

func (c *client) GetData(ctx context.Context, key api.DataKey) (*api.Data, error) {
	secret := &vaultSecret{Data: &api.Data{
		Vars:    map[string]*api.Val{},
		Secrets: map[string]*api.Val{},
	}}

	resp, err := c.Logical().ReadRawWithDataWithContext(ctx, key.Path(Instance), nil)
	if resp != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode == http.StatusNotFound {
		return secret.Data, nil
	}
	if err != nil {
		return nil, err
	}
	if err := resp.DecodeJSON(secret); err != nil {
		return nil, err
	}

	return secret.Data, nil
}

func (c *client) PutData(ctx context.Context, key api.DataKey, data *api.Data) error {
	b, err := json.Marshal(&vaultSecret{Data: data})
	if err != nil {
		return err
	}

	if _, err := c.Logical().WriteBytesWithContext(ctx, key.Path(Instance), b); err != nil {
		return err
	}

	return nil
}
