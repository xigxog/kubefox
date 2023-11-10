package controller

import (
	"context"
	"fmt"
	"net/http"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/components/operator/templates"
	"github.com/xigxog/kubefox/logkf"
)

const (
	FieldOwner client.FieldOwner = "kubefox-operator"
)

type Client struct {
	client.Client
}

func (c *Client) ApplyTemplate(ctx context.Context, name string, data *templates.Data, log *logkf.Logger) error {
	objs, err := templates.Render(name, data)
	if err != nil {
		return err
	}

	for _, obj := range objs {
		log.Debugf("applying template resource '%s'", toString(obj))
		if err := c.Apply(ctx, obj); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) DeleteTemplate(ctx context.Context, name string, data *templates.Data, log *logkf.Logger) error {
	objs, err := templates.Render(name, data)
	if err != nil {
		return err
	}

	for _, obj := range objs {
		log.Debugf("deleting template resource '%s'", toString(obj))
		if err := c.Delete(ctx, obj); err != nil {
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
	if err := r.Get(ctx, NN("", namespace), ns); err != nil {
		return nil, fmt.Errorf("unable to fetch namespace: %w", err)
	}
	if ns.Status.Phase == v1.NamespaceTerminating {
		return nil, ErrNotFound
	}

	p := &v1alpha1.Platform{}
	l := &v1alpha1.PlatformList{}
	if err := r.List(ctx, l, client.InNamespace(namespace)); err != nil {
		return nil, fmt.Errorf("unable to fetch platform: %w", err)
	}
	switch c := len(l.Items); c {
	case 0:
		return nil, ErrNotFound
	case 1:
		p = &l.Items[0]
	default:
		return nil, ErrTooManyPlatforms
	}

	return p, nil
}

func (r *Client) GetWebhookPlatform(ctx context.Context, obj client.Object) (*v1alpha1.Platform, admission.Response, error) {
	platform, err := r.GetPlatform(ctx, obj.GetNamespace())
	var resp admission.Response
	switch {
	case err == ErrNotFound:
		resp = admission.Denied(
			fmt.Sprintf(`The %s "%s" not allowed: Platform not found in Namespace "%s"`,
				obj.GetObjectKind(), obj.GetName(), obj.GetNamespace()))
	case err == ErrTooManyPlatforms:
		resp = admission.Denied(
			fmt.Sprintf(`The %s "%s" not allowed: More than one Platform found in Namespace "%s"`,
				obj.GetObjectKind(), obj.GetName(), obj.GetNamespace()))
	case err != nil:
		resp = admission.Errored(http.StatusInternalServerError, err)
	}

	return platform, resp, err
}

func NN(namespace, name string) types.NamespacedName {
	return types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
}

func toString(obj client.Object) string {
	gvk := obj.GetObjectKind().GroupVersionKind()
	grp := gvk.Group
	if grp == "" {
		grp = "core"
	}
	return fmt.Sprintf("%s/%s/%s/%s", grp, gvk.Version, gvk.Kind, obj.GetName())
}
