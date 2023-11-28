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
	// Hang on to TypeMeta as it is erased by create.
	t := obj.GetObjectKind()
	opts := []client.CreateOption{c.FieldOwner}
	if dryRun {
		opts = append(opts, client.DryRunAll)
	}
	err := c.Create(ctx, obj, opts...)
	if IsAlreadyExists(err) {
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
	envSnapshot.Data.Source = v1alpha1.EnvSource{
		Kind:            kind,
		Name:            envName,
		ResourceVersion: envObj.GetResourceVersion(),
	}

	return envSnapshot, nil
}
