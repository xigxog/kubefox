package operator

import (
	"fmt"

	"github.com/xigxog/kubefox/libs/core/api/common"
	kubev1a1 "github.com/xigxog/kubefox/libs/core/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/libs/core/api/maker"
	"github.com/xigxog/kubefox/libs/core/api/uri"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/platform"
	"github.com/xigxog/kubefox/libs/core/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (opr *operator) ProcessComponentSet(kit kubefox.Kit) error {
	k := kit.Request().Kube()

	req := &Request[kubev1a1.ComponentSet]{}
	if err := k.Unmarshal(req); err != nil {
		return ErrEvent(kit, err)
	}
	kit.Log().Infof("processing %s hook for %s", k.GetHook(), req.GetObject())

	switch k.GetHook() {
	case kubefox.Customize:
		return CustomizeEvent(kit)

	case kubefox.Sync:
		// orgName := kit.Organization()
		platName := kit.Platform()
		sysName := req.GetObject().Name

		attachments := []runtime.Object{}
		status := kubev1a1.ComponentSetStatus{
			Components:  map[common.ComponentKey]*common.ComponentStatus{},
			Deployments: map[uri.Key]*common.DeploymentStatus{},
		}

		for depKey, sysObj := range req.GetObject().Spec.Deployments {
			for _, comp := range sysObj.Components {
				// if Deployment already created
				if compStatus, ok := status.Components[comp.Key()]; ok {
					// ensure SysRef is listed
					if !utils.Contains(compStatus.Deployments, depKey) {
						compStatus.Deployments = append(compStatus.Deployments, depKey)
					}
					opr.updateSysRefStatus(&status, depKey, compStatus.Ready)
					// these aren’t the droids we’re looking for, move on
					continue
				}

				// create Deployment
				attachments = append(attachments, opr.deployment(kit.Namespace(), platName, sysName, comp))

				curDep, found := req.Attachments.Deployments[string(comp.Key())]
				compStatus := &common.ComponentStatus{
					Deployments: []uri.Key{depKey},
					Ready:       found && curDep.Status.AvailableReplicas > 0,
				}
				// add status which also indicates Deployment was created
				status.Components[comp.Key()] = compStatus
				opr.updateSysRefStatus(&status, depKey, compStatus.Ready)
			}
		}

		return SyncEvent(kit, status, attachments...)

	default:
		return ErrEvent(kit, fmt.Errorf("unknown hook type: %s", k.GetHook()))
	}
}

func (opr *operator) updateSysRefStatus(status *kubev1a1.ComponentSetStatus, depKey uri.Key, ready bool) {
	if sysObjStatus, ok := status.Deployments[depKey]; ok {
		sysObjStatus.Ready = sysObjStatus.Ready && ready
	} else {
		status.Deployments[depKey] = &common.DeploymentStatus{
			Ready: ready,
		}
	}
}

func (opr *operator) deployment(platNS, plat, sys string, comp *common.ComponentProps) *appsv1.Deployment {
	defMode := new(int32)
	*defMode = 420

	deploy := maker.New[appsv1.Deployment](maker.Props{
		Group: "apps",
		Name:  string(comp.Key()),
		// Organization: org,
		System:    sys,
		Component: comp.Name,
		CompHash:  comp.ShortHash(),
	})

	deploy.Spec = appsv1.DeploymentSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: deploy.Labels,
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: deploy.Labels,
			},
			Spec: corev1.PodSpec{
				ServiceAccountName: platform.BrokerSvcAccount,
				ImagePullSecrets: []corev1.LocalObjectReference{
					{
						Name: platform.ImagePullSecret,
					},
				},
				Containers: []corev1.Container{
					{
						Name:            "broker",
						ImagePullPolicy: corev1.PullAlways,
						Image:           opr.BrokerImage,
						Args: []string{
							"component",
							"--dev",
							"--system=" + sys,
							"--component=" + comp.Name,
							"--component-hash=" + comp.GitHash,
						},
						EnvFrom: []corev1.EnvFromSource{
							{
								ConfigMapRef: &corev1.ConfigMapEnvSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "kfp-env",
									},
								},
							},
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      platform.NATSCertSecret,
								MountPath: "/kubefox/nats/tls",
								ReadOnly:  true,
							},
							{
								Name:      platform.CertSecret,
								MountPath: platform.CertDir,
								ReadOnly:  true,
							},
						},
					},
					{
						Name:  "component",
						Image: comp.Image,
						Args: []string{
							"--dev",
						},
						EnvFrom: []corev1.EnvFromSource{
							{
								ConfigMapRef: &corev1.ConfigMapEnvSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: platform.EnvConfigMap,
									},
								},
							},
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      platform.CertSecret,
								MountPath: platform.CertDir,
								ReadOnly:  true,
							},
						},
					},
				},
				Volumes: []corev1.Volume{
					{
						Name: platform.NATSCertSecret,
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								DefaultMode: defMode,
								SecretName:  platform.NATSCertSecret,
							},
						},
					},
					{
						Name: platform.CertSecret,
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								DefaultMode: defMode,
								SecretName:  platform.CertSecret,
							},
						},
					},
				},
			},
		},
	}

	return deploy
}

func Contains(s []string, e string) bool {
	for i := range s {
		if s[i] == e {
			return true
		}
	}
	return false
}
