package operator

import (
	"fmt"

	"github.com/xigxog/kubefox/libs/core/api/common"
	kube "github.com/xigxog/kubefox/libs/core/api/kubernetes"
	kubev1a1 "github.com/xigxog/kubefox/libs/core/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/libs/core/api/maker"
	"github.com/xigxog/kubefox/libs/core/api/uri"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/platform"
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
	kit.Log().Infof("processing %s hook for %s", k.GetHook(), req.GetObject())

	switch k.GetHook() {
	case kubefox.Customize:
		// will send vault pod during sync, if 'not ready' should run init again
		inst := req.GetObject().GetLabels()[kube.InstanceLabel]
		vaultPod := &RelatedResourceRule{
			APIVersion: "v1",
			Resource:   "pods",
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					kube.InstanceLabel: inst,
					kube.NameLabel:     "vault",
				},
			},
		}
		certs := &RelatedResourceRule{
			APIVersion: "v1",
			Resource:   "secrets",
			Namespace:  kit.Namespace(),
			Names:      []string{platform.CertSecret, platform.NATSCertSecret},
		}
		env := &RelatedResourceRule{
			APIVersion: "v1",
			Resource:   "configmaps",
			Namespace:  kit.Namespace(),
			Names:      []string{platform.EnvConfigMap},
		}
		brkSvc := &RelatedResourceRule{
			APIVersion: "v1",
			Resource:   "services",
			Namespace:  kit.Namespace(),
			Names:      []string{platform.BrkService},
		}

		return CustomizeEvent(kit, vaultPod, certs, env, brkSvc)

	case kubefox.Sync:
		// orgName := req.GetObject().GetOrganization()
		inst := req.GetObject().GetLabels()[kube.InstanceLabel]
		status := &common.PlatformStatus{
			Healthy: isPlatformHealthy(kit, req.Related),
			Systems: map[uri.Key]*common.PlatformSystemStatus{},
		}

		if !status.Healthy {
			kit.Log().Warnf("vault pod is not ready")
			go op.vaultOp.Init()
		}

		attachments := []runtime.Object{}
		for sysName, sys := range req.GetObject().Spec.Systems {
			ns := maker.New[corev1.Namespace](maker.Props{
				Group: "core",
				Name:  utils.SystemNamespace(inst, string(sysName)),
				// Organization: orgName,
				Instance: inst,
				System:   string(sysName),
			})
			attachments = append(attachments, ns)

			compSet := maker.New[kubev1a1.ComponentSet](maker.Props{
				Name:      string(sysName),
				Namespace: ns.GetName(),
				// Organization: orgName,
				Instance: inst,
				System:   string(sysName),
			})
			attachments = append(attachments, compSet)

			sec := maker.New[corev1.Secret](maker.Props{
				Group:     "core",
				Name:      platform.ImagePullSecret,
				Namespace: ns.GetName(),
				// Organization: orgName,
				Instance: inst,
				System:   string(sysName),
			})
			sec.Type = "kubernetes.io/dockerconfigjson"
			config := "{}"
			if sys.ImagePullSecret != "" {
				config = fmt.Sprintf(`{"auths":{"ghcr.io":{"auth":"%s"}}}`, sys.ImagePullSecret)
			}
			sec.StringData = map[string]string{".dockerconfigjson": config}
			attachments = append(attachments, sec)

			brkSA := maker.New[corev1.ServiceAccount](maker.Props{
				Group:     "core",
				Name:      platform.BrokerSvcAccount,
				Namespace: ns.GetName(),
				// Organization: orgName,
				Instance: inst,
				System:   string(sysName),
			})
			attachments = append(attachments, brkSA)

			for _, r := range req.Related.Secrets {
				sec := maker.New[corev1.Secret](maker.Props{
					Group:     "core",
					Name:      r.Name,
					Namespace: ns.GetName(),
					// Organization: orgName,
					Instance: inst,
					System:   string(sysName),
				})
				sec.Type = r.Type
				sec.Data = r.Data
				attachments = append(attachments, sec)
			}
			for _, r := range req.Related.ConfigMaps {
				cm := maker.New[corev1.ConfigMap](maker.Props{
					Group:     "core",
					Name:      r.Name,
					Namespace: ns.GetName(),
					// Organization: orgName,
					Instance: inst,
					System:   string(sysName),
				})
				cm.Data = r.Data
				attachments = append(attachments, cm)
			}
			for _, r := range req.Related.Services {
				svc := maker.New[corev1.Service](maker.Props{
					Group:     "core",
					Name:      r.Name,
					Namespace: ns.GetName(),
					// Organization: orgName,
					Instance: inst,
					System:   string(sysName),
				})
				svc.Spec = r.Spec
				attachments = append(attachments, svc)
			}

			status.Systems[sysName] = &common.PlatformSystemStatus{
				Healthy: isSystemHealthy(req.Attachments, ns.GetName(), compSet.GetName()),
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
