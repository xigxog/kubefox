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
	"path/filepath"
	"strings"
	"time"

	vapi "github.com/hashicorp/vault/api"
	vauth "github.com/hashicorp/vault/api/auth/kubernetes"
	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/utils"
)

type Client struct {
	*vapi.Client
	ClientOptions
}

type ClientOptions struct {
	Instance string
	Role     string
	URL      string
	// CACert is the path to a PEM-encoded CA cert file to use to verify the
	// Vault server SSL certificate. It takes precedence over CACertBytes.
	CACert string
	// CACertBytes is a PEM-encoded certificate or bundle.
	CACertBytes []byte
	// AutoRenew indicates that the Vault token should automatically be renewed
	// to ensure it does not expire.
	AutoRenew bool
	TokenPath string
}

type Key struct {
	Instance  string
	Namespace string
	Component string
}

type VaultSecret struct {
	Data *VaultData `json:"data"`
}

type VaultData struct {
	Data any `json:"data"`
}

func New(opts ClientOptions) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if opts.TokenPath == "" {
		opts.TokenPath = api.PathSvcAccToken
	}

	cfg := vapi.DefaultConfig()
	cfg.Address = opts.URL
	cfg.MaxRetries = 3
	cfg.HttpClient.Timeout = time.Second * 5
	cfg.ConfigureTLS(&vapi.TLSConfig{
		CACert:      opts.CACert,
		CACertBytes: opts.CACertBytes,
	})

	vaultCli, err := vapi.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(opts.TokenPath)
	if err != nil {
		return nil, err
	}
	token := vauth.WithServiceAccountToken(string(b))
	auth, err := vauth.NewKubernetesAuth(opts.Role, token)
	if err != nil {
		return nil, err
	}
	authInfo, err := vaultCli.Auth().Login(ctx, auth)
	if err != nil {
		return nil, err
	}
	if authInfo == nil {
		return nil, fmt.Errorf("error logging in with Kubernetes auth: no auth info was returned")
	}

	if opts.AutoRenew {
		watcher, err := vaultCli.NewLifetimeWatcher(&vapi.LifetimeWatcherInput{Secret: authInfo})
		if err != nil {
			return nil, fmt.Errorf("error starting Vault token renewer: %w", err)
		}
		go watcher.Start()
	}

	return &Client{
		Client:        vaultCli,
		ClientOptions: opts,
	}, nil
}

func (c *Client) CreateDataStore(ctx context.Context, namespace string) error {
	path := DataPath(api.DataKey{
		Instance:  c.Instance,
		Namespace: namespace,
	})
	scope := "Namespace"
	if namespace == "" {
		scope = "Cluster"
	}

	// Check if store already exists.
	if cfg, _ := c.Sys().MountConfigWithContext(ctx, path); cfg != nil {
		return nil
	}

	return c.Sys().MountWithContext(ctx, path, &vapi.MountInput{
		Type:        "kv",
		Description: scope + " scoped KubeFox Environment Data store.",
		Options: map[string]string{
			"version": "2", // Supports versioning and optimistic locking.
		},
	})
}

func (c *Client) GetData(ctx context.Context, key api.DataKey, data *api.Data) error {
	if key.Instance == "" {
		key.Instance = c.Instance
	}

	resp, err := c.Logical().ReadRawWithDataWithContext(ctx, DataSubPath(key, "data"), nil)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusNotFound {
		return core.ErrNotFound()
	}

	secret := &VaultSecret{
		Data: &VaultData{
			Data: data,
		},
	}
	if err := resp.DecodeJSON(&secret); err != nil {
		return err
	}

	return nil
}

func (c *Client) PutData(ctx context.Context, key api.DataKey, data *api.Data) error {
	if key.Instance == "" {
		key.Instance = c.Instance
	}

	b, err := json.Marshal(&VaultData{Data: data})
	if err != nil {
		return err
	}

	if _, err := c.Logical().WriteBytesWithContext(ctx, DataSubPath(key, "data"), b); err != nil {
		return err
	}

	return nil
}

func (c *Client) DeleteData(ctx context.Context, key api.DataKey) error {
	if key.Instance == "" {
		key.Instance = c.Instance
	}

	if _, err := c.Logical().DeleteWithContext(ctx, DataSubPath(key, "metadata")); err != nil {
		return err
	}

	return nil
}

func RoleName(key Key) string {
	return PolicyName(key, "")
}

func PolicyName(key Key, policy string) string {
	key.Namespace = strings.TrimPrefix(key.Namespace, key.Instance)
	key.Namespace = strings.TrimPrefix(key.Namespace, "-")
	if !strings.HasPrefix(key.Instance, "kubefox") {
		key.Instance = "kubefox-" + key.Instance
	}

	return utils.Join("-", key.Instance, key.Namespace, key.Component, policy)
}

func KubernetesRolePath(key Key) string {
	return fmt.Sprintf("auth/kubernetes/role/%s", RoleName(key))
}

func PKIPath(key Key) string {
	if key.Namespace == "" {
		return fmt.Sprintf("kubefox/pki/instance/%s/root", key.Instance)
	} else {
		return fmt.Sprintf("kubefox/pki/instance/%s/namespace/%s", key.Instance, key.Namespace)
	}
}

func PKISubPath(key Key, subPath string) string {
	return filepath.Join(PKIPath(key), subPath)
}

func DataPath(key api.DataKey) string {
	if key.Namespace == "" {
		return fmt.Sprintf("kubefox/kv/instance/%s/cluster", key.Instance)
	} else {
		return fmt.Sprintf("kubefox/kv/instance/%s/namespace/%s", key.Instance, key.Namespace)
	}
}

func DataSubPath(key api.DataKey, subPath string) string {
	return filepath.Join(DataPath(key), subPath, key.Kind, key.Name)
}
