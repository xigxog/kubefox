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
	"sigs.k8s.io/controller-runtime/pkg/client"

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

	vClient *vapi.Client
	cm      *ComponentManager
	log     *logkf.Logger

	mutex sync.Mutex
}

// SetupWithManager sets up the controller with the Manager.
func (r *PlatformReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.log = logkf.Global.With(logkf.KeyController, "platform")
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
	if err := r.Get(ctx, req.NamespacedName, p); client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, log.ErrorN("unable to fetch platform: %w", err)

	} else if apierrors.IsNotFound(err) {
		log.Debug("platform was deleted, removing namespace label")
		delete(ns.Labels, kubefox.LabelK8sPlatform)
		return ctrl.Result{}, r.Update(ctx, ns)
	}

	ready := false
	defer func() {
		p.Status.Ready = ready
		r.Status().Update(ctx, p)
	}()

	// TODO move to admission webhook
	if lbl, found := ns.Labels[kubefox.LabelK8sPlatform]; found && lbl != req.Name {
		return ctrl.Result{}, log.ErrorN("namespace belongs to platform '%s'", lbl)

	} else if !found {
		ns.Labels[kubefox.LabelK8sPlatform] = p.Name
		if err := r.Update(ctx, ns); err != nil {
			return ctrl.Result{}, err
		}
	}

	cm := &v1.ConfigMap{}
	if err := r.Get(ctx, nn(r.Namespace, r.Instance+"-root-ca"), cm); err != nil {
		return ctrl.Result{}, log.ErrorN("unable to fetch root CA configmap: %w", err)
	}
	td := &TemplateData{
		Data: templates.Data{
			Instance: templates.Instance{
				Name:           r.Instance,
				Namespace:      r.Namespace,
				RootCA:         string(cm.Data["ca.crt"]),
				BootstrapImage: BootstrapImage,
				Version:        build.Info.Version,
			},
			Values: map[string]any{
				"pkiInitImage": VaultImage,
			},
		},
	}
	if err := r.ApplyTemplate(ctx, "instance", &td.Data); err != nil {
		return ctrl.Result{}, log.ErrorN("problem setting up instance: %w", err)
	}

	td.Platform = templates.Platform{
		Name:      p.Name,
		Namespace: p.Namespace,
		LogFormat: r.LogFormat,
		LogLevel:  r.LogLevel,
	}
	td.Owner = []*metav1.OwnerReference{
		metav1.NewControllerRef(p, p.GroupVersionKind()),
	}

	if err := r.setupVault(ctx, td); err != nil {
		return ctrl.Result{}, log.ErrorN("problem setting up vault: %w", err)
	}

	if err := r.ApplyTemplate(ctx, "platform", &td.Data); err != nil {
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
	if rdy, err := r.cm.SetupComponent(ctx, td); !rdy || err != nil {
		return ctrl.Result{}, err
	}

	td.Template = "httpsrv"
	td.Obj = &appsv1.DaemonSet{}
	td.Component = templates.Component{
		Name:          "httpsrv",
		Image:         HTTPSrvImage,
		PodSpec:       p.Spec.HTTPSrv.PodSpec,
		ContainerSpec: p.Spec.HTTPSrv.ContainerSpec,
	}
	td.Values = map[string]any{
		"serviceType": p.Spec.HTTPSrv.Service.Type,
		"httpPort":    p.Spec.HTTPSrv.Service.Ports.HTTP,
		"httpsPort":   p.Spec.HTTPSrv.Service.Ports.HTTPS,
	}
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
	if rdy, err := r.cm.SetupComponent(ctx, td); !rdy || err != nil {
		return ctrl.Result{}, err
	}

	// Used by defer func created above.
	ready = true
	log.Debug("platform reconciled")

	if rdy, err := r.cm.ReconcileComponents(ctx, p.Namespace); !rdy || err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *PlatformReconciler) setupVault(ctx context.Context, td *TemplateData) error {
	vault, err := r.vaultClient(ctx, r.VaultURL, []byte(td.Instance.RootCA))
	if err != nil {
		return err
	}

	vName := td.PlatformVaultName()
	pkiPath := "pki/int/platform/" + vName
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
	_, err = vault.Logical().Write(pkiPath+"/roles/nats", map[string]interface{}{
		"issuer_ref":         "default",
		"allowed_domains":    td.Platform.Name + "-nats." + td.Platform.Namespace,
		"allow_localhost":    true,
		"allow_bare_domains": true,
		"max_ttl":            TenYears,
	})
	if err != nil {
		return err
	}
	_, err = vault.Logical().Write(pkiPath+"/roles/broker", map[string]interface{}{
		"issuer_ref":         "default",
		"allowed_domains":    td.Platform.Name + "-broker." + td.Platform.Namespace,
		"allow_localhost":    true,
		"allow_bare_domains": true,
		"max_ttl":            TenYears,
	})
	if err != nil {
		return err
	}
	_, err = vault.Logical().Write(pkiPath+"/roles/httpsrv", map[string]interface{}{
		"issuer_ref":         "default",
		"allowed_domains":    td.Platform.Name + "-httpsrv." + td.Platform.Namespace,
		"allow_localhost":    true,
		"allow_bare_domains": true,
		"max_ttl":            TenYears,
	})
	if err != nil {
		return err
	}

	err = vault.Sys().PutPolicyWithContext(ctx, vName+"-nats", `
	// issue nats certs
	path "`+pkiPath+`/issue/nats" {
		capabilities = ["create", "update"]
	}
		`)
	if err != nil {
		return err
	}
	err = vault.Sys().PutPolicyWithContext(ctx, vName+"-broker", `
	// issue broker certs
	path "`+pkiPath+`/issue/broker" {
		capabilities = ["create", "update"]
	}
	`)
	if err != nil {
		return err
	}
	err = vault.Sys().PutPolicyWithContext(ctx, vName+"-httpsrv", `
	// issue httpsrv certs
	path "`+pkiPath+`/issue/httpsrv" {
		capabilities = ["create", "update"]
	}
	`)
	if err != nil {
		return err
	}

	_, err = vault.Logical().Write("auth/kubernetes/role/"+vName+"-nats", map[string]interface{}{
		"bound_service_account_names":      td.Platform.Name + "-nats",
		"bound_service_account_namespaces": td.Platform.Namespace,
		"token_policies":                   vName + "-nats",
	})
	if err != nil {
		return err
	}
	_, err = vault.Logical().Write("auth/kubernetes/role/"+vName+"-broker", map[string]interface{}{
		"bound_service_account_names":      td.Platform.Name + "-broker",
		"bound_service_account_namespaces": td.Platform.Namespace,
		"token_policies":                   vName + "-broker,kv-kubefox-reader",
	})
	if err != nil {
		return err
	}
	_, err = vault.Logical().Write("auth/kubernetes/role/"+vName+"-httpsrv", map[string]interface{}{
		"bound_service_account_names":      td.Platform.Name + "-httpsrv",
		"bound_service_account_namespaces": td.Platform.Namespace,
		"token_policies":                   vName + "-httpsrv",
	})
	if err != nil {
		return err
	}

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
	cfg.HttpClient.Timeout = time.Minute
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
