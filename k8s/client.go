package k8s

import (
	"context"

	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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
	// Hang on to TypeMeta as it is erased by create.
	t := obj.GetObjectKind()
	opts := []client.CreateOption{c.FieldOwner}
	if dryRun {
		opts = append(opts, client.DryRunAll)
	}
	err := c.Create(ctx, obj, opts...)
	if apierrors.IsAlreadyExists(err) {
		copy := obj.DeepCopyObject().(client.Object)
		if err := c.Get(ctx, client.ObjectKey{Namespace: obj.GetNamespace(), Name: obj.GetName()}, copy); err != nil {
			return err
		}

		obj.SetResourceVersion(copy.GetResourceVersion())

		opts := []client.UpdateOption{c.FieldOwner}
		if dryRun {
			opts = append(opts, client.DryRunAll)
		}
		err = c.Update(ctx, obj, opts...)
	}
	// Restore TypeMeta.
	obj.GetObjectKind().SetGroupVersionKind(t.GroupVersionKind())

	return err
}

func (c *Client) Apply(ctx context.Context, obj client.Object, opts ...client.PatchOption) error {
	obj.SetManagedFields(nil)
	obj.SetResourceVersion("")
	opts = append(opts, c.FieldOwner, client.ForceOwnership)
	return c.Patch(ctx, obj, client.Apply, opts...)
}

func (c *Client) Merge(ctx context.Context, obj client.Object, opts ...client.PatchOption) error {
	key := client.ObjectKeyFromObject(obj)
	og := obj.DeepCopyObject().(client.Object)
	if err := c.Get(ctx, key, og); err != nil {
		return err
	}

	return c.Patch(ctx, obj, client.MergeFrom(og), opts...)
}

func (c *Client) ApplyStatus(ctx context.Context, obj client.Object, opts ...client.SubResourcePatchOption) error {
	obj.SetManagedFields(nil)
	obj.SetResourceVersion("")
	opts = append(opts, c.FieldOwner, client.ForceOwnership)
	return c.Status().Patch(ctx, obj, client.Apply, opts...)
}

func (r *Client) GetVirtualEnvObj(ctx context.Context, namespace, envName, snapshotName string) (v1alpha1.VirtualEnvObject, error) {
	if snapshotName != "" {
		env := &v1alpha1.VirtualEnvSnapshot{}
		return env, r.Get(ctx, Key(namespace, envName), env)
	}

	env := &v1alpha1.VirtualEnv{}
	err := r.Get(ctx, Key(namespace, envName), env)
	switch {
	case err == nil:
		return env, nil

	case apierrors.IsNotFound(err):
		env := &v1alpha1.ClusterVirtualEnv{}
		return env, r.Get(ctx, Key("", envName), env)

	default:
		return nil, err
	}
}
