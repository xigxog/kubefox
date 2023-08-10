package operator

import (
	"encoding/base64"
	"fmt"

	"github.com/xigxog/kubefox/components/operator/templates"
	"github.com/xigxog/kubefox/libs/core/api/common"
	kube "github.com/xigxog/kubefox/libs/core/api/kubernetes"
	kubev1a1 "github.com/xigxog/kubefox/libs/core/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/libs/core/api/uri"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (op *operator) ProcessPlatform(kit kubefox.Kit) error {
	k := kit.Request().Kube()

	req := &Request[kubev1a1.Platform]{}
	if err := k.Unmarshal(req); err != nil {
		return ErrEvent(kit, err)
	}
	pName := req.GetObject().GetName()
	pNamespace := req.GetObject().GetNamespace()

	kit.Log().Infof("processing %s hook for %s", k.GetHook(), req.GetObject())

	switch k.GetHook() {
	case kubefox.Customize:
		vaultPod := &RelatedResourceRule{
			APIVersion: "v1",
			Resource:   "pods",
			Namespace:  pNamespace,
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					kube.InstanceLabel: pName,
					kube.NameLabel:     "vault",
				},
			},
		}
		rootCASec := &RelatedResourceRule{
			APIVersion: "v1",
			Resource:   "secrets",
			Namespace:  pNamespace,
			Names:      []string{pName + "-root-ca"},
		}

		return CustomizeEvent(kit, vaultPod, rootCASec)

	case kubefox.Sync:
		rootCA := ""
		rootCASec := req.Attachments.Secrets[fmt.Sprintf("%s/%s-root-ca", pNamespace, pName)]
		if rootCASec != nil {
			rootCABytes := []byte{}
			if _, err := base64.StdEncoding.Decode(rootCABytes, rootCASec.Data["ca.crt"]); err != nil {
				return ErrEvent(kit, err)
			}
			rootCA = string(rootCABytes)
		}

		attachments := []runtime.Object{}
		status := &common.PlatformStatus{
			Healthy: isPlatformHealthy(kit, req.Related),
			Systems: map[uri.Key]*common.PlatformSystemStatus{},
		}

		for sysName, sys := range req.GetObject().Spec.Systems {
			nsName := utils.SystemNamespace(pName, string(sysName))

			data := &templates.Data{
				DevMode: kit.DevMode(),
				Platform: templates.Platform{
					Name:      pName,
					Namespace: pNamespace,
					Version:   GitRef,
					RootCA:    rootCA,
				},
				System: templates.System{
					Name:              string(sysName),
					Namespace:         nsName,
					ContainerRegistry: sys.ContainerRegistry,
					ImagePullSecret:   sys.ImagePullSecret,
				},
			}

			objs, err := templates.Render("system", data)
			if err != nil {
				return ErrEvent(kit, err)
			}
			for _, obj := range objs {
				attachments = append(attachments, obj)
			}

			status.Systems[sysName] = &common.PlatformSystemStatus{
				Healthy: isSystemHealthy(req.Attachments, nsName, string(sysName)),
			}
		}

		return SyncEvent(kit, status, attachments...)

	default:
		return ErrEvent(kit, fmt.Errorf("unknown hook type: %s", k.GetHook()))
	}
}

func isPlatformHealthy(kit kubefox.Kit, related *RequestChildren) bool {
	for _, pod := range related.Pods {
		for _, status := range pod.Status.ContainerStatuses {
			if status.Name == "vault" {
				if !status.Ready {
					kit.Log().Warnf("vault pod is not ready")
					return false
				}
			}
		}
	}

	return true
}

func isSystemHealthy(attachments *RequestChildren, namespace, compSet string) bool {
	curNs := attachments.Namespaces[namespace]
	if curNs == nil || curNs.Status.Phase != corev1.NamespaceActive {
		return false
	}

	// map key includes namespace
	curCompSet := attachments.ComponentSets[fmt.Sprintf("%s/%s", namespace, compSet)]
	if curCompSet == nil {
		return false
	}

	for _, comp := range curCompSet.Status.Components {
		if !comp.Ready {
			return false
		}
	}

	return true
}
