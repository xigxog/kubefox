/*
Copyright © 2023 XigXog

This Source Code Form is subject to the terms of the Mozilla Public License,
v2.0. If a copy of the MPL was not distributed with this file, You can obtain
one at https://mozilla.org/MPL/2.0/.
*/

package controller

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	vapi "github.com/hashicorp/vault/api"
	vauth "github.com/hashicorp/vault/api/auth/kubernetes"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/build"
	"github.com/xigxog/kubefox/components/operator/defaults"
	"github.com/xigxog/kubefox/components/operator/templates"
	"github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/k8s"
	"github.com/xigxog/kubefox/logkf"
)

const (
	TenYears     string = "87600h"
	HundredYears string = "876000h"
)

// PlatformReconciler reconciles a Platform object
type PlatformReconciler struct {
	*Client

	Instance  string
	Namespace string
	VaultURL  string

	LogLevel  string
	LogFormat string

	CompMgr *ComponentManager

	setupMap map[string]bool

	vClient *vapi.Client
	log     *logkf.Logger

	mutex sync.Mutex
}

// SetupWithManager sets up the controller with the Manager.
func (r *PlatformReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.log = logkf.Global.With(logkf.KeyController, "Platform")
	r.setupMap = make(map[string]bool)
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Platform{}).
		Owns(&appsv1.DaemonSet{}).
		Owns(&appsv1.Deployment{}).
		Owns(&appsv1.StatefulSet{}).
		Complete(r)
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *PlatformReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.log.With(
		"namespace", req.Namespace,
		"name", req.Name,
	)
	log.Debugf("reconciling Platform '%s/%s'", req.Namespace, req.Name)

	pKey := fmt.Sprintf("%s/%s", req.Namespace, req.Name)

	ns := &v1.Namespace{}
	if err := r.Get(ctx, k8s.Key("", req.Namespace), ns); err != nil {
		r.setSetup(pKey, false)
		return ctrl.Result{}, k8s.IgnoreNotFound(err)
	}
	if ns.Status.Phase == v1.NamespaceTerminating {
		log.Debug("Namespace is terminating")
		r.setSetup(pKey, false)
		return ctrl.Result{}, nil
	}

	platform := &v1alpha1.Platform{}
	if err := r.Get(ctx, req.NamespacedName, platform); err != nil {
		r.setSetup(pKey, false)
		return ctrl.Result{}, k8s.IgnoreNotFound(err)
	}

	err := r.reconcile(ctx, platform, log)
	if err != nil {
		platform.Status.Conditions = k8s.UpdateConditions(metav1.Now(), platform.Status.Conditions, &metav1.Condition{
			Type:               api.ConditionTypeAvailable,
			Status:             metav1.ConditionUnknown,
			ObservedGeneration: platform.ObjectMeta.Generation,
			Reason:             api.ConditionReasonReconcileFailed,
			Message:            err.Error(),
		})
	}

	if err == nil {
		if _, err := r.CompMgr.ReconcileApps(ctx, req.Namespace); err != nil {
			r.log.Error(err)
		}
	}

	if err := r.updateComponentsStatus(ctx, platform); err != nil {
		r.log.Error(err)
	}

	log.Debugf("updating Platform '%s/%s' status", req.Namespace, req.Name)
	if err := r.ApplyStatus(ctx, platform); err != nil {
		r.log.Error(err)
	}

	log.Debugf("reconciling Platform '%s/%s' done", req.Namespace, req.Name)

	return ctrl.Result{}, err
}

func (r *PlatformReconciler) reconcile(ctx context.Context, platform *v1alpha1.Platform, log *logkf.Logger) error {
	pKey := fmt.Sprintf("%s/%s", platform.Namespace, platform.Name)

	cm := &v1.ConfigMap{}
	if err := r.Get(ctx, k8s.Key(r.Namespace, r.Instance+"-root-ca"), cm); err != nil {
		r.setSetup(pKey, false)
		return log.ErrorN("unable to fetch root CA configmap: %w", err)
	}

	maxEventSize := platform.Spec.Events.MaxSize.Value()
	if platform.Spec.Events.MaxSize.IsZero() {
		maxEventSize = api.DefaultMaxEventSizeBytes
	}
	baseTD := &TemplateData{
		Data: templates.Data{
			Instance: templates.Instance{
				Name:           r.Instance,
				Namespace:      r.Namespace,
				RootCA:         cm.Data["ca.crt"],
				BootstrapImage: BootstrapImage,
			},
			Platform: templates.Platform{
				Name:      platform.Name,
				Namespace: platform.Namespace,
			},
			Owner: []*metav1.OwnerReference{
				metav1.NewControllerRef(platform, platform.GroupVersionKind()),
			},
			BuildInfo: build.Info,
			Logger:    platform.Spec.Logger,
			Values: map[string]any{
				api.ValKeyMaxEventSize: maxEventSize,
				api.ValKeyVaultURL:     r.VaultURL,
			},
		},
	}

	setup := r.isSetup(pKey)
	if setup {
		r.log.Debugf("Platform '%s' already setup ", pKey)

	} else {
		// Ensure there are valid commits for Platform components.
		if !api.RegexpCommit.MatchString(build.Info.BrokerCommit) ||
			!api.RegexpCommit.MatchString(build.Info.HTTPSrvCommit) {
			log.Error("broker or httpsrv commit from build info is invalid")
			return nil
		}
		if err := r.setupVault(ctx, baseTD); err != nil {
			return log.ErrorN("problem setting up vault: %w", err)
		}
		if err := r.ApplyTemplate(ctx, "platform", &baseTD.Data, log); err != nil {
			return log.ErrorN("problem setting up Platform: %w", err)
		}

		r.setSetup(pKey, true)
	}

	td := baseTD.ForComponent(api.PlatformComponentNATS, &appsv1.StatefulSet{}, &defaults.NATS, templates.Component{
		Component: &core.Component{
			Type: string(api.ComponentTypeNATS),
			Name: api.PlatformComponentNATS,
		},
		Image:               NATSImage,
		PodSpec:             platform.Spec.NATS.PodSpec,
		ContainerSpec:       platform.Spec.NATS.ContainerSpec,
		IsPlatformComponent: true,
	})
	if err := r.setupPKI(ctx, td, ""); err != nil {
		return err
	}
	if rdy, err := r.CompMgr.SetupComponent(ctx, td); !rdy || err != nil {
		platform.Status.Conditions = k8s.UpdateConditions(metav1.Now(), platform.Status.Conditions, &metav1.Condition{
			Type:               api.ConditionTypeAvailable,
			Status:             metav1.ConditionFalse,
			ObservedGeneration: platform.ObjectMeta.Generation,
			Reason:             api.ConditionReasonNATSUnavailable,
			Message:            fmt.Sprintf(`NATS StatefulSet "%s" is unavailable.`, td.Obj.GetName()),
		})
		return chill(err)
	}

	td = baseTD.ForComponent(api.PlatformComponentBroker, &appsv1.DaemonSet{}, &defaults.Broker, templates.Component{
		Component: &core.Component{
			Type:   string(api.ComponentTypeBroker),
			Name:   api.PlatformComponentBroker,
			Commit: build.Info.BrokerCommit,
		},
		Image:               BrokerImage,
		PodSpec:             platform.Spec.Broker.PodSpec,
		ContainerSpec:       platform.Spec.Broker.ContainerSpec,
		IsPlatformComponent: true,
	})
	if err := r.setupPKI(ctx, td, ""); err != nil {
		return err
	}
	if rdy, err := r.CompMgr.SetupComponent(ctx, td); !rdy || err != nil {
		platform.Status.Conditions = k8s.UpdateConditions(metav1.Now(), platform.Status.Conditions, &metav1.Condition{
			Type:               api.ConditionTypeAvailable,
			Status:             metav1.ConditionFalse,
			ObservedGeneration: platform.ObjectMeta.Generation,
			Reason:             api.ConditionReasonBrokerUnavailable,
			Message:            fmt.Sprintf(`Broker DaemonSet "%s" is unavailable.`, td.Obj.GetName()),
		})
		return chill(err)
	}

	td = baseTD.ForComponent(api.PlatformComponentHTTPSrv, &appsv1.Deployment{}, &defaults.HTTPSrv, templates.Component{
		Component: &core.Component{
			Type:   string(api.ComponentTypeHTTPAdapter),
			Name:   api.PlatformComponentHTTPSrv,
			Commit: build.Info.HTTPSrvCommit,
		},
		Image:               HTTPSrvImage,
		PodSpec:             platform.Spec.HTTPSrv.PodSpec,
		ContainerSpec:       platform.Spec.HTTPSrv.ContainerSpec,
		IsPlatformComponent: true,
	})
	td.Values["serviceLabels"] = platform.Spec.HTTPSrv.Service.Labels
	td.Values["serviceAnnotations"] = platform.Spec.HTTPSrv.Service.Annotations
	td.Values["serviceType"] = platform.Spec.HTTPSrv.Service.Type
	td.Values["httpPort"] = platform.Spec.HTTPSrv.Service.Ports.HTTP
	td.Values["httpsPort"] = platform.Spec.HTTPSrv.Service.Ports.HTTPS
	if err := r.setupPKI(ctx, td, ""); err != nil {
		return err
	}
	if rdy, err := r.CompMgr.SetupComponent(ctx, td); !rdy || err != nil {
		platform.Status.Conditions = k8s.UpdateConditions(metav1.Now(), platform.Status.Conditions, &metav1.Condition{
			Type:               api.ConditionTypeAvailable,
			Status:             metav1.ConditionFalse,
			ObservedGeneration: platform.ObjectMeta.Generation,
			Reason:             api.ConditionReasonHTTPSrvUnavailable,
			Message:            fmt.Sprintf(`HTTPSrv Deployment "%s" is unavailable.`, td.Obj.GetName()),
		})
		return chill(err)
	}

	platform.Status.Conditions = k8s.UpdateConditions(metav1.Now(), platform.Status.Conditions, &metav1.Condition{
		Type:               api.ConditionTypeAvailable,
		Status:             metav1.ConditionTrue,
		ObservedGeneration: platform.ObjectMeta.Generation,
		Reason:             api.ConditionReasonPlatformComponentsAvailable,
		Message:            "Platform Components are available.",
	})

	return nil
}

func (r *PlatformReconciler) updateComponentsStatus(ctx context.Context, p *v1alpha1.Platform) error {
	p.Status.Components = nil

	req, err := labels.NewRequirement(api.LabelK8sComponentType,
		selection.NotEquals, []string{string(api.ComponentTypeKubeFox)},
	)
	if err != nil {
		return err
	}
	podList := &v1.PodList{}
	if err := r.List(ctx, podList,
		client.InNamespace(p.Namespace),
		client.MatchingLabelsSelector{Selector: labels.NewSelector().Add(*req)},
	); err != nil {
		return err
	}

	for _, pod := range podList.Items {
		cond := k8s.PodCondition(&pod, v1.PodReady)
		p.Status.Components = append(p.Status.Components, v1alpha1.ComponentStatus{
			Ready:    cond.Status == v1.ConditionTrue,
			Name:     pod.Labels[api.LabelK8sComponent],
			Commit:   pod.Labels[api.LabelK8sComponentCommit],
			Type:     api.ComponentType(pod.Labels[api.LabelK8sComponentType]),
			PodName:  pod.Name,
			PodIP:    pod.Status.PodIP,
			NodeName: pod.Spec.NodeName,
			NodeIP:   pod.Status.HostIP,
		})
	}

	return nil
}

func (r *PlatformReconciler) setupVault(ctx context.Context, td *TemplateData) error {
	r.log.Debugf("setting up vault for Platform '%s'", td.PlatformFullName())

	vault, err := r.vaultClient(ctx, r.VaultURL, []byte(td.Instance.RootCA))
	if err != nil {
		return err
	}

	// Setup KVs for secrets.
	secretPath := fmt.Sprintf("kubefox/instance/%s/namespace/%s",
		td.Instance.Name, td.Platform.Namespace)
	if cfg, _ := vault.Sys().MountConfig(secretPath); cfg != nil {
		r.log.Debugf("namespace kv store at path '%s' exists", secretPath)

	} else {
		r.log.Infof("creating namespace kv store at path '%s'", secretPath)
		err = vault.Sys().Mount(secretPath, &vapi.MountInput{
			Type: "kv",
			Description: fmt.Sprintf("Namespace scoped KubeFox secret data store; instance: %s, namespace: %s",
				td.Instance.Name, td.Platform.Namespace),
			Options: map[string]string{
				"version": "2", // Supports versioning and optimistic locking.
			},
		})
		if err != nil {
			return fmt.Errorf("error creating namespace kv store at path '%s': %w", secretPath, err)
		}
	}

	secretPath = fmt.Sprintf("kubefox/instance/%s/cluster", td.Instance.Name)
	if cfg, _ := vault.Sys().MountConfig(secretPath); cfg != nil {
		r.log.Debugf("cluster kv store at path '%s' exists", secretPath)

	} else {
		r.log.Infof("creating cluster kv store at path '%s'", secretPath)
		err = vault.Sys().Mount(secretPath, &vapi.MountInput{
			Type:        "kv",
			Description: fmt.Sprintf("Cluster scoped KubeFox secret data store; instance: %s", td.Instance.Name),
			Options: map[string]string{
				"version": "2", // Supports versioning and optimistic locking.
			},
		})
		if err != nil {
			return fmt.Errorf("error creating cluster kv store at path '%s': %w", secretPath, err)
		}
	}

	// Setup PKI.
	pkiPath := fmt.Sprintf("pki/int/platform/%s", td.PlatformVaultName())
	if cfg, _ := vault.Sys().MountConfig(pkiPath); cfg == nil {
		err = vault.Sys().Mount(pkiPath, &vapi.MountInput{
			Type: "pki",
			Config: vapi.MountConfigInput{
				MaxLeaseTTL: HundredYears,
			},
		})
		if err != nil {
			return err
		}
		_, err := vault.Logical().Write(pkiPath+"/config/urls", map[string]interface{}{
			"issuing_certificates":    r.VaultURL + "/v1/" + pkiPath + "/ca",
			"crl_distribution_points": r.VaultURL + "/v1/" + pkiPath + "/crl",
		})
		if err != nil {
			return err
		}
		s, err := vault.Logical().Write(pkiPath+"/intermediate/generate/internal", map[string]interface{}{
			"common_name": "KubeFox Platform " + td.PlatformFullName() + " Intermediate CA",
			"issuer_name": td.PlatformFullName() + "-intermediate",
		})
		if err != nil {
			return err
		}
		s, err = vault.Logical().Write("pki/root/root/sign-intermediate", map[string]interface{}{
			"csr":    s.Data["csr"],
			"format": "pem_bundle",
			"ttl":    HundredYears,
		})
		if err != nil {
			return err
		}
		_, err = vault.Logical().Write(pkiPath+"/intermediate/set-signed", map[string]interface{}{
			"certificate": s.Data["certificate"],
		})
		if err != nil {
			return err
		}
	}

	r.log.Debugf("vault successfully setup for Platform '%s'", td.Platform.Name)

	return nil
}

func (r *PlatformReconciler) setupPKI(ctx context.Context, td *TemplateData, additionalPolicies string) error {
	setup := r.isSetup(td.ComponentFullName())
	if setup {
		r.log.Debugf("pki already setup for component '%s'", td.ComponentFullName())
		return nil
	}
	r.log.Debugf("setting up pki for component '%s'", td.ComponentFullName())

	vault, err := r.vaultClient(ctx, r.VaultURL, []byte(td.Instance.RootCA))
	if err != nil {
		return err
	}

	pkiPath := fmt.Sprintf("pki/int/platform/%s", td.PlatformVaultName())

	path := fmt.Sprintf("%s/roles/%s", pkiPath, td.ComponentVaultName())
	svcName := fmt.Sprintf("%s.%s", td.ComponentFullName(), td.Platform.Namespace)
	_, err = vault.Logical().Write(path, map[string]interface{}{
		"issuer_ref":         "default",
		"allowed_domains":    fmt.Sprintf("%s,%s.svc", svcName, svcName),
		"allow_localhost":    true,
		"allow_bare_domains": true,
		"max_ttl":            TenYears,
	})
	if err != nil {
		return err
	}

	err = vault.Sys().PutPolicyWithContext(ctx, td.ComponentVaultName(), `
	// issue certs
	path "`+pkiPath+`/issue/`+td.ComponentVaultName()+`" {
		capabilities = ["create", "update"]
	}
		`)
	if err != nil {
		return err
	}

	path = fmt.Sprintf("auth/kubernetes/role/%s", td.ComponentVaultName())
	policy := td.ComponentVaultName()
	if additionalPolicies != "" {
		policy = fmt.Sprintf("%s,%s", policy, additionalPolicies)
	}

	r.log.Debugf("writing role; path: %s, sa: %s, policy: %s", path, td.ComponentFullName(), policy)
	_, err = vault.Logical().Write(path, map[string]interface{}{
		"bound_service_account_names":      td.ComponentFullName(),
		"bound_service_account_namespaces": td.Platform.Namespace,
		"token_policies":                   policy,
	})
	if err != nil {
		return err
	}

	r.setSetup(td.ComponentFullName(), true)
	r.log.Debugf("pki successfully setup for component '%s'", td.ComponentFullName())

	return nil
}

func (r *PlatformReconciler) vaultClient(ctx context.Context, url string, caCert []byte) (*vapi.Client, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.vClient != nil {
		return r.vClient, nil
	}

	cfg := vapi.DefaultConfig()
	cfg.Address = url
	cfg.MaxRetries = 3
	cfg.HttpClient.Timeout = time.Second * 5
	cfg.ConfigureTLS(&vapi.TLSConfig{
		CACertBytes: caCert,
	})

	vault, err := vapi.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(api.PathSvcAccToken)
	if err != nil {
		return nil, err
	}
	token := vauth.WithServiceAccountToken(string(b))
	auth, err := vauth.NewKubernetesAuth("kubefox-operator", token)
	if err != nil {
		return nil, err
	}
	authInfo, err := vault.Auth().Login(ctx, auth)
	if err != nil {
		return nil, err
	}
	if authInfo == nil {
		return nil, fmt.Errorf("error logging in with kubernetes auth: no auth info was returned")
	}
	r.vClient = vault

	watcher, err := vault.NewLifetimeWatcher(&vapi.LifetimeWatcherInput{Secret: authInfo})
	if err != nil {
		return nil, fmt.Errorf("error starting Vault token renewer: %w", err)
	}
	go watcher.Start()

	return vault, nil
}

func (r *PlatformReconciler) setSetup(key string, val bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.setupMap[key] = val
}

func (r *PlatformReconciler) isSetup(key string) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return r.setupMap[key]
}

// chill waits a few seconds for things to chillax.
func chill(err error) error {
	time.Sleep(time.Second * 3)
	return err
}
