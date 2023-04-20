package client

import (
	"context"
	"fmt"

	"github.com/xigxog/kubefox/libs/core/api/kubernetes"
	kubev1a1 "github.com/xigxog/kubefox/libs/core/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/libs/core/api/maker"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/logger"
	"github.com/xigxog/kubefox/libs/core/utils"
	k "sigs.k8s.io/controller-runtime/pkg/client"
)

type kubeClient struct {
	kube      k.Client
	patchOpts *k.PatchOptions

	log *logger.Log
}

func (c *kubeClient) LookupPlatform(namespace string) (*kubev1a1.Platform, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	list := maker.Empty[kubev1a1.PlatformList]()
	if err := c.kube.List(ctx, list, k.InNamespace(namespace)); err != nil {
		return nil, err
	}

	if len(list.Items) != 1 {
		return nil, fmt.Errorf("KubeFox Platform resource not found in namespace %s", namespace)
	}

	return &list.Items[0], nil
}

func (c *kubeClient) Create(kit kubefox.Kit, obj k.Object) error {
	kit.Log().Debugf("creating %s", kubernetes.FullKey(obj))
	gvk := obj.GetObjectKind().GroupVersionKind()

	err := c.kube.Create(kit.Ctx(), obj)
	obj.GetObjectKind().SetGroupVersionKind(gvk)
	return err
}

func (c *kubeClient) Put(kit kubefox.Kit, obj k.Object) error {
	kit.Log().Debugf("updating %s", kubernetes.FullKey(obj))

	gvk := obj.GetObjectKind().GroupVersionKind()
	if err := c.kube.Update(kit.Ctx(), obj); err != nil {
		obj.GetObjectKind().SetGroupVersionKind(gvk)
		if IsNotFound(err) {
			kit.Log().Debugf("resource %s not found, creating it", kubernetes.FullKey(obj))
			return c.Create(kit, obj)
		} else {
			return err
		}
	}

	return nil
}

func (c *kubeClient) Patch(kit kubefox.Kit, obj k.Object) error {
	kit.Log().Debugf("patching %s", kubernetes.FullKey(obj))
	gvk := obj.GetObjectKind().GroupVersionKind()

	if err := c.kube.Patch(kit.Ctx(), obj, k.Merge, c.patchOpts); err != nil {
		obj.GetObjectKind().SetGroupVersionKind(gvk)
		if IsNotFound(err) {
			kit.Log().Debugf("resource %s not found, creating it", kubernetes.FullKey(obj))
			return c.Create(kit, obj)
		} else {
			return err
		}
	}

	return nil
}

func (c *kubeClient) Get(kit kubefox.Kit, obj k.Object) error {
	kit.Log().Debugf("getting %s", kubernetes.FullKey(obj))
	gvk := obj.GetObjectKind().GroupVersionKind()

	err := c.kube.Get(kit.Ctx(), k.ObjectKeyFromObject(obj), obj)
	obj.GetObjectKind().SetGroupVersionKind(gvk)
	return err
}

func (c *kubeClient) List(kit kubefox.Kit, list k.ObjectList, labels map[string]string) error {
	return c.list(kit, list, "", labels)
}

func (c *kubeClient) ListNamespaced(kit kubefox.Kit, list k.ObjectList, sys string, labels map[string]string) error {
	return c.list(kit, list, sys, labels)
}

func (c *kubeClient) list(kit kubefox.Kit, list k.ObjectList, sys string, labels map[string]string) error {
	kit.Log().Debugf("listing %s", kubernetes.KindKey(list))
	gvk := list.GetObjectKind().GroupVersionKind()
	namespace := utils.SystemNamespace(kit.Platform(), sys)

	err := c.kube.List(kit.Ctx(), list, k.InNamespace(namespace), k.MatchingLabels(labels))
	list.GetObjectKind().SetGroupVersionKind(gvk)
	return err
}

func (c *kubeClient) Delete(kit kubefox.Kit, obj k.Object) error {
	kit.Log().Debugf("deleting %s", kubernetes.FullKey(obj))

	return c.kube.Delete(kit.Ctx(), obj)
}
