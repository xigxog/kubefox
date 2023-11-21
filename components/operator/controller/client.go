package controller

import (
	"context"
	"errors"
	"fmt"
	"strings"

	hash "github.com/mitchellh/hashstructure/v2"
	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/components/operator/templates"
	"github.com/xigxog/kubefox/k8s"
	"github.com/xigxog/kubefox/logkf"
	"github.com/xigxog/kubefox/utils"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	ErrResourceVersionConflict = errors.New("resource version conflict")
)

type Client struct {
	k8s.Client
}

func (c *Client) ApplyTemplate(ctx context.Context, name string, data *templates.Data, log *logkf.Logger) error {
	objs, err := templates.Render(name, data)
	if err != nil {
		return err
	}

	for _, obj := range objs {
		log.Debugf("applying template resource '%s'", ToString(obj))
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
		log.Debugf("deleting template resource '%s'", ToString(obj))
		if err := c.Delete(ctx, obj); err != nil {
			return err
		}
	}

	return nil
}

func (r *Client) GetPlatform(ctx context.Context, namespace string) (*v1alpha1.Platform, error) {
	ns := &v1.Namespace{}
	if err := r.Get(ctx, Key("", namespace), ns); err != nil {
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

// GetResolvedEnvironment returns the ResolvedEnvironment referenced by the
// Release, creating it first if needed.
func (r *Client) GetResolvedEnvironment(ctx context.Context, rel *v1alpha1.Release) (*v1alpha1.ResolvedEnvironment, error) {
	envName := rel.Spec.Environment.Name
	resEnvName := fmt.Sprintf("%s-%s", envName, rel.Spec.Environment.ResourceVersion)

	resEnv := &v1alpha1.ResolvedEnvironment{}
	err := r.Get(ctx, Key(rel.Namespace, resEnvName), resEnv)

	switch {
	case apierrors.IsNotFound(err):
		env := &v1alpha1.VirtualEnv{}
		if err := r.Get(ctx, Key("", envName), env); err != nil {
			return nil, err
		}
		if env.ResourceVersion != rel.Spec.Environment.ResourceVersion {
			return nil, ErrResourceVersionConflict
		}
		resEnv = &v1alpha1.ResolvedEnvironment{
			TypeMeta: metav1.TypeMeta{
				APIVersion: v1alpha1.GroupVersion.Identifier(),
				Kind:       "ResolvedEnvironment",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      resEnvName,
				Namespace: rel.Namespace,
			},
			Data:    env.Data,
			Details: env.Details,
		}
		return resEnv, r.Create(ctx, resEnv)

	case err == nil:
		return resEnv, nil

	default:
		return nil, err
	}
}

func Key(namespace, name string) types.NamespacedName {
	return types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
}

func HasLabel(obj client.Object, key, value string) bool {
	if obj == nil {
		return false
	}

	value = utils.CleanLabel(value)
	if obj.GetLabels() == nil {
		obj.SetLabels(make(map[string]string))
	}

	if curVal, found := obj.GetLabels()[key]; curVal == value {
		return true
	} else if !found && value == "" {
		return true
	}

	return false
}

func UpdateLabel(obj client.Object, key, value string) bool {
	if obj == nil {
		return false
	}

	value = utils.CleanLabel(value)
	if HasLabel(obj, key, value) {
		return false
	}

	if value != "" {
		obj.GetLabels()[key] = value
	} else {
		delete(obj.GetLabels(), key)
	}

	return true
}

func HashesEqual(lhs, rhs any) bool {
	lhsHash, err := hash.Hash(lhs, hash.FormatV2, nil)
	if err != nil {
		return false
	}
	rhsHash, err := hash.Hash(rhs, hash.FormatV2, nil)
	if err != nil {
		return false
	}
	if lhsHash != rhsHash {
		return false
	}

	return true
}

func IgnoreNotFound(err error) error {
	if apierrors.IsNotFound(err) || errors.Is(err, ErrNotFound) {
		return nil
	}
	return err
}

// IsFailedWebhookErr will return true if error indicates it was caused by
// calling a webhook. This is useful during operator startup when the Pod is not
// marked ready which causes the webhooks to fail.
func IsFailedWebhookErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "failed calling webhook")
}

func ToString(obj client.Object) string {
	gvk := obj.GetObjectKind().GroupVersionKind()
	grp := gvk.Group
	if grp == "" {
		grp = "core"
	}
	return fmt.Sprintf("%s/%s/%s/%s.%s", grp, gvk.Version, gvk.Kind, obj.GetName(), obj.GetNamespace())
}
