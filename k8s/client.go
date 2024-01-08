// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package k8s

import (
	"context"

	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	cmd "k8s.io/client-go/tools/clientcmd"
	cmdapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type Client struct {
	client.Client

	KubeConfig *cmdapi.Config
	RestConfig *rest.Config
	FieldOwner client.FieldOwner
}

func NewClient(fieldOwner client.FieldOwner) (*Client, error) {
	v1alpha1.SchemeBuilder.AddToScheme(scheme.Scheme)

	loader := cmd.NewDefaultClientConfigLoadingRules()
	kubeCfg, err := loader.Load()
	if err != nil {
		return nil, err
	}

	restCfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	cli, err := client.New(restCfg, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		return nil, err
	}

	return &Client{
		Client:     cli,
		KubeConfig: kubeCfg,
		RestConfig: restCfg,
		FieldOwner: fieldOwner,
	}, nil
}

func (c *Client) Upsert(ctx context.Context, obj client.Object, dryRun bool) error {
	orig := obj.DeepCopyObject().(client.Object)

	opts := []client.CreateOption{c.FieldOwner}
	if dryRun {
		opts = append(opts, client.DryRunAll)
	}
	err := c.Create(ctx, obj, opts...)
	if IsAlreadyExists(err) {
		latest := obj.DeepCopyObject().(client.Object)
		if err := c.Get(ctx, client.ObjectKey{Namespace: obj.GetNamespace(), Name: obj.GetName()}, latest); err != nil {
			return err
		}

		obj.SetResourceVersion(latest.GetResourceVersion())

		opts := []client.PatchOption{c.FieldOwner}
		if dryRun {
			opts = append(opts, client.DryRunAll)
		}

		err = c.Merge(ctx, obj, latest, opts...)
	}
	// Restore TypeMeta.
	obj.GetObjectKind().SetGroupVersionKind(orig.GetObjectKind().GroupVersionKind())

	return err
}

func (c *Client) Apply(ctx context.Context, obj client.Object, opts ...client.PatchOption) error {
	obj.SetManagedFields(nil)
	obj.SetResourceVersion("")
	opts = append(opts, c.FieldOwner, client.ForceOwnership)
	return c.Patch(ctx, obj, client.Apply, opts...)
}

func (c *Client) Merge(ctx context.Context, modified, original client.Object, opts ...client.PatchOption) error {
	p, err := c.merge(ctx, modified, original)
	if err != nil {
		return err
	}

	return c.Patch(ctx, modified, p, opts...)
}

func (c *Client) ApplyStatus(ctx context.Context, obj client.Object, opts ...client.SubResourcePatchOption) error {
	obj.SetManagedFields(nil)
	obj.SetResourceVersion("")
	opts = append(opts, c.FieldOwner, client.ForceOwnership)
	return c.Status().Patch(ctx, obj, client.Apply, opts...)
}

func (c *Client) MergeStatus(ctx context.Context, modified, original client.Object, opts ...client.SubResourcePatchOption) error {
	p, err := c.merge(ctx, modified, original)
	if err != nil {
		return err
	}

	return c.Status().Patch(ctx, modified, p, opts...)
}

func (c *Client) merge(ctx context.Context, modified, original client.Object) (client.Patch, error) {
	key := client.ObjectKeyFromObject(modified)
	if original == nil {
		original = modified.DeepCopyObject().(client.Object)
		if err := c.Get(ctx, key, original); err != nil {
			return nil, err
		}
	}
	modified.SetResourceVersion(original.GetResourceVersion())

	return client.MergeFrom(original), nil
}
