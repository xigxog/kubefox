package controller

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/xigxog/kubefox/components/operator/templates"
	"github.com/xigxog/kubefox/libs/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/libs/core/kubefox"
)

const (
	FieldOwner client.FieldOwner = "kubefox-operator"
)

type Client struct {
	client.Client
}

func (c *Client) ApplyTemplate(ctx context.Context, name string, data *templates.Data) error {
	objs, err := templates.Render(name, data)
	if err != nil {
		return err
	}

	for _, obj := range objs {
		if err := c.Apply(ctx, obj); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) Apply(ctx context.Context, obj client.Object) error {
	return c.Patch(ctx, obj, client.Apply, FieldOwner, client.ForceOwnership)
}

func (c *Client) Merge(ctx context.Context, obj client.Object) error {
	return c.Patch(ctx, obj, client.Merge, FieldOwner)
}

func (r *Client) GetPlatform(ctx context.Context, namespace string) (*v1alpha1.Platform, error) {
	ns := &v1.Namespace{}
	if err := r.Get(ctx, nn("", namespace), ns); err != nil {
		return nil, fmt.Errorf("unable to fetch namespace: %w", err)
	}
	pName, found := ns.Labels[kubefox.LabelK8sPlatform]
	if !found {
		return nil, fmt.Errorf("namespace does not have '%s' label", kubefox.LabelK8sPlatform)
	}
	p := &v1alpha1.Platform{}
	if err := r.Get(ctx, nn(namespace, pName), p); err != nil {
		return nil, fmt.Errorf("unable to fetch platform: %w", err)
	}

	return p, nil
}

func nn(namespace, name string) types.NamespacedName {
	return types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
}
