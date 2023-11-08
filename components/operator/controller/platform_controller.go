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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/build"
	"github.com/xigxog/kubefox/components/operator/templates"
	kubefox "github.com/xigxog/kubefox/core"
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

	Scheme *runtime.Scheme

	vaultMap map[string]bool

	vClient *vapi.Client
	cm      *ComponentManager
	log     *logkf.Logger

	mutex sync.Mutex
}

// SetupWithManager sets up the controller with the Manager.
func (r *PlatformReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.log = logkf.Global.With(logkf.KeyController, "platform")
	r.vaultMap = make(map[string]bool)
	r.cm = &ComponentManager{
		Instance: r.Instance,
		Client:   r.Client,
		Log:      r.log,
	}
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
	log.Debug("reconciling platform")

	ns := &v1.Namespace{}
	if err := r.Get(ctx, nn("", req.Namespace), ns); err != nil {
		if apierrors.IsNotFound(err) {
			log.Debug("namespace is gone")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, log.ErrorN("unable to fetch namespace: %w", err)
	}
	if ns.Status.Phase == v1.NamespaceTerminating {
		log.Debug("namespace is terminating")
		return ctrl.Result{}, nil
	}

	p := &v1alpha1.Platform{}
	if err := r.Get(ctx, req.NamespacedName, p); err != nil {
		return ctrl.Result{}, err
	}

	ready := false
	defer func() {
		p.Status.Ready = ready
		r.Status().Update(ctx, p)
	}()

	cm := &v1.ConfigMap{}
	if err := r.Get(ctx, nn(r.Namespace, r.Instance+"-root-ca"), cm); err != nil {
		return ctrl.Result{}, log.ErrorN("unable to fetch root CA configmap: %w", err)
	}
	td := &TemplateData{
		Data: templates.Data{
			Instance: templates.Instance{
				Name:           r.Instance,
				Namespace:      r.Namespace,
				RootCA:         cm.Data["ca.crt"],
				BootstrapImage: BootstrapImage,
				Version:        build.Info.Version,
			},
			Platform: templates.Platform{
				Name:      p.Name,
				Namespace: p.Namespace,
				LogFormat: r.LogFormat,
				LogLevel:  r.LogLevel,
			},
			Owner: []*metav1.OwnerReference{
				metav1.NewControllerRef(p, p.GroupVersionKind()),
			},
			BuildInfo: build.Info,
			Values: map[string]any{
				"vaultURL": r.VaultURL,
			},
		},
	}

	if err := r.setupVault(ctx, td); err != nil {
		return ctrl.Result{}, log.ErrorN("problem setting up vault: %w", err)
	}

	if err := r.ApplyTemplate(ctx, "platform", &td.Data, log); err != nil {
		return ctrl.Result{}, log.ErrorN("problem setting up platform: %w", err)
	}

	td.Template = "nats"
	td.Obj = &appsv1.StatefulSet{}
	td.Component = templates.Component{
		Name:          "nats",
		Image:         NATSImage,
		PodSpec:       p.Spec.NATS.PodSpec,
		ContainerSpec: p.Spec.NATS.ContainerSpec,
	}
	if td.Component.Resources == nil {
		td.Component.Resources = &v1.ResourceRequirements{
			// TODO calc and set correct values and use those in headers
			Requests: v1.ResourceList{
				"memory": resource.MustParse("115Mi"), // 90% of limit, used to set GOMEMLIMIT
				"cpu":    resource.MustParse("250m"),
			},
			Limits: v1.ResourceList{
				"memory": resource.MustParse("128Mi"),
				"cpu":    resource.MustParse("2"),
			},
		}
	}
	if td.Component.LivenessProbe == nil {
		td.Component.LivenessProbe = &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{
					Path: "/healthz?js-enabled-only=true",
					Port: intstr.FromString("monitor"),
				},
			},
			TimeoutSeconds:   3,
			PeriodSeconds:    30,
			FailureThreshold: 3,
		}
	}
	if td.Component.ReadinessProbe == nil {
		td.Component.ReadinessProbe = &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{
					Path: "/healthz?js-enabled-only=true",
					Port: intstr.FromString("monitor"),
				},
			},
			TimeoutSeconds:   3,
			PeriodSeconds:    10,
			FailureThreshold: 3,
		}
	}
	if td.Component.StartupProbe == nil {
		td.Component.StartupProbe = &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{
					Path: "/healthz",
					Port: intstr.FromString("monitor"),
				},
			},
			PeriodSeconds:    5,
			FailureThreshold: 90,
		}
	}
	if err := r.setupVaultComponent(ctx, td, ""); err != nil {
		return ctrl.Result{}, err
	}
	if rdy, err := r.cm.SetupComponent(ctx, td); !rdy || err != nil {
		return ctrl.Result{}, err
	}

	td.Template = "broker"
	td.Obj = &appsv1.DaemonSet{}
	td.Component = templates.Component{
		Name:          "broker",
		Image:         BrokerImage,
		PodSpec:       p.Spec.Broker.PodSpec,
		ContainerSpec: p.Spec.Broker.ContainerSpec,
	}
	if td.Component.Resources == nil {
		td.Component.Resources = &v1.ResourceRequirements{
			// TODO calc and set correct values and use those in headers
			Requests: v1.ResourceList{
				"memory": resource.MustParse("144Mi"), // 90% of limit, used to set GOMEMLIMIT
				"cpu":    resource.MustParse("250m"),
			},
			Limits: v1.ResourceList{
				"memory": resource.MustParse("160Mi"),
				"cpu":    resource.MustParse("2"),
			},
		}
	}
	if td.Component.LivenessProbe == nil {
		td.Component.LivenessProbe = &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{
					Port: intstr.FromString("health"),
				},
			},
		}
	}
	if td.Component.ReadinessProbe == nil {
		td.Component.ReadinessProbe = &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{
					Port: intstr.FromString("health"),
				},
			},
		}
	}
	if err := r.setupVaultComponent(ctx, td, ""); err != nil {
		return ctrl.Result{}, err
	}
	if rdy, err := r.cm.SetupComponent(ctx, td); !rdy || err != nil {
		return ctrl.Result{}, err
	}

	td.Template = "httpsrv"
	td.Obj = &appsv1.Deployment{}
	td.Component = templates.Component{
		Name:          "httpsrv",
		Image:         HTTPSrvImage,
		PodSpec:       p.Spec.HTTPSrv.PodSpec,
		ContainerSpec: p.Spec.HTTPSrv.ContainerSpec,
	}
	td.Values["serviceType"] = p.Spec.HTTPSrv.Service.Type
	td.Values["httpPort"] = p.Spec.HTTPSrv.Service.Ports.HTTP
	td.Values["httpsPort"] = p.Spec.HTTPSrv.Service.Ports.HTTPS

	if td.Component.Resources == nil {
		td.Component.Resources = &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				"memory": resource.MustParse("144Mi"), // 90% of limit, used to set GOMEMLIMIT
				"cpu":    resource.MustParse("250m"),
			},
			Limits: v1.ResourceList{
				"memory": resource.MustParse("160Mi"),
				"cpu":    resource.MustParse("2"),
			},
		}
	}
	if td.Component.LivenessProbe == nil {
		td.Component.LivenessProbe = &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{
					Port: intstr.FromString("health"),
				},
			},
		}
	}
	if td.Component.ReadinessProbe == nil {
		td.Component.ReadinessProbe = &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{
					Port: intstr.FromString("health"),
				},
			},
		}
	}
	if err := r.setupVaultComponent(ctx, td, ""); err != nil {
		return ctrl.Result{}, err
	}
	if rdy, err := r.cm.SetupComponent(ctx, td); !rdy || err != nil {
		return ctrl.Result{}, err
	}

	// Used by defer func created above.
	ready = true
	log.Debug("platform reconciled")

	if rdy, err := r.cm.ReconcileApps(ctx, p.Namespace); !rdy || err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// TODO break operations into funcs and move to reusable vault client
func (r *PlatformReconciler) setupVault(ctx context.Context, td *TemplateData) error {
	r.mutex.Lock()
	setup := r.vaultMap[td.PlatformFullName()]
	r.mutex.Unlock()
	if setup {
		r.log.Debugf("vault already setup for platform '%s'", td.PlatformFullName())
		return nil
	}
	r.log.Debugf("setting up vault for platform '%s'", td.PlatformFullName())

	// Ensure there are valid commits for platform components.
	if !kubefox.RegexpCommit.MatchString(build.Info.BrokerCommit) ||
		!kubefox.RegexpCommit.MatchString(build.Info.HTTPSrvCommit) {
		return fmt.Errorf("broker or httpsrv commit from build info is invalid")
	}

	vault, err := r.vaultClient(ctx, r.VaultURL, []byte(td.Instance.RootCA))
	if err != nil {
		return err
	}

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

	r.mutex.Lock()
	r.vaultMap[td.PlatformFullName()] = true
	r.mutex.Unlock()
	r.log.Debugf("vault successfully setup for platform '%s'", td.Platform.Name)

	return nil
}

func (r *PlatformReconciler) setupVaultComponent(ctx context.Context, td *TemplateData, additionalPolicies string) error {
	r.mutex.Lock()
	setup := r.vaultMap[td.ComponentFullName()]
	r.mutex.Unlock()
	if setup {
		r.log.Debugf("vault already setup for component '%s'", td.ComponentFullName())
		return nil
	}
	r.log.Debugf("setting up vault for component '%s'", td.ComponentFullName())

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

	r.mutex.Lock()
	r.vaultMap[td.ComponentFullName()] = true
	r.mutex.Unlock()
	r.log.Debugf("vault successfully setup for component '%s'", td.ComponentFullName())

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

	b, err := os.ReadFile(kubefox.PathSvcAccToken)
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
