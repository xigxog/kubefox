package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (r *Client) GetVirtualEnvObj(ctx context.Context, namespace, envId string, requireSnapshot bool) (v1alpha1.VirtualEnvObject, error) {
	var (
		obj v1alpha1.VirtualEnvObject
		err error
	)

	obj = &v1alpha1.VirtualEnvSnapshot{}
	if err = r.Get(ctx, Key(namespace, envId), obj); IgnoreNotFound(err) != nil {
		return nil, err
	} else if IsNotFound(err) && requireSnapshot {
		return nil, err
	} else if err == nil {
		return obj, nil
	}

	obj = &v1alpha1.VirtualEnv{}
	if err = r.Get(ctx, Key(namespace, envId), obj); IgnoreNotFound(err) != nil {
		return nil, err
	} else if err == nil {
		return obj, nil
	}

	obj = &v1alpha1.ClusterVirtualEnv{}
	if err = r.Get(ctx, Key("", envId), obj); IgnoreNotFound(err) != nil {
		return nil, err
	} else if err == nil {
		return obj, nil
	}

	return nil, err
}

func (r *Client) SnapshotVirtualEnv(ctx context.Context, namespace, envName string) (*v1alpha1.VirtualEnvSnapshot, error) {
	envObj, err := r.GetVirtualEnvObj(ctx, namespace, envName, false)
	if err != nil {
		return nil, err
	}

	envSnapshot := &v1alpha1.VirtualEnvSnapshot{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.GroupVersion.Identifier(),
			Kind:       "VirtualEnvSnapshot",
		},
	}
	if envObj.GetParent() != "" {
		// Parent must be ClusterVirtualEnv so clear namespace.
		parent := &v1alpha1.ClusterVirtualEnv{}
		if err := r.Get(ctx, Key("", envObj.GetParent()), parent); err != nil {
			return nil, err
		}
		v1alpha1.MergeVirtualEnvironment(envSnapshot, parent)
	}

	v1alpha1.MergeVirtualEnvironment(envSnapshot, envObj)

	var kind string
	switch envObj.(type) {
	case *v1alpha1.VirtualEnvSnapshot:
		kind = "VirtualEnvSnapshot"
	case *v1alpha1.VirtualEnv:
		kind = "VirtualEnv"
	case *v1alpha1.ClusterVirtualEnv:
		kind = "ClusterVirtualEnv"
	}

	now := time.Now()
	envSnapshot.Data.SnapshotTime = metav1.NewTime(now)
	envSnapshot.Name = fmt.Sprintf("%s-%s-%s", envName, envObj.GetResourceVersion(), now.UTC().Format("20060102-150405"))
	envSnapshot.Namespace = namespace
	envSnapshot.Data.Source = v1alpha1.VirtualEnvSource{
		Kind:            kind,
		Name:            envName,
		ResourceVersion: envObj.GetResourceVersion(),
	}

	return envSnapshot, nil
}
