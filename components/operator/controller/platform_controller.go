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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/build"
	"github.com/xigxog/kubefox/components/operator/templates"
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
	r.log = logkf.Global.With(logkf.KeyController, "platform")
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
	log.Debug("reconciling Platform")

	ns := &v1.Namespace{}
	if err := r.Get(ctx, Key("", req.Namespace), ns); err != nil {
		if apierrors.IsNotFound(err) {
			log.Debug("Namespace is gone")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, log.ErrorN("unable to fetch Namespace: %w", err)
	}
	if ns.Status.Phase == v1.NamespaceTerminating {
		log.Debug("Namespace is terminating")
		return ctrl.Result{}, nil
	}

	p := &v1alpha1.Platform{}
	if err := r.Get(ctx, req.NamespacedName, p); err != nil {
		return ctrl.Result{}, err
	}

	ready := false
	defer func() {
		p.Status.Ready = ready
		if err := r.ApplyStatus(ctx, p); err != nil {
			log.Error(err)
		}
	}()

	cm := &v1.ConfigMap{}
	if err := r.Get(ctx, Key(r.Namespace, r.Instance+"-root-ca"), cm); err != nil {
		return ctrl.Result{}, log.ErrorN("unable to fetch root CA configmap: %w", err)
	}

	maxEventSize := p.Spec.Config.Events.MaxSize.Value()
	if p.Spec.Config.Events.MaxSize.IsZero() {
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
				Name:      p.Name,
				Namespace: p.Namespace,
			},
			Owner: []*metav1.OwnerReference{
				metav1.NewControllerRef(p, p.GroupVersionKind()),
			},
			BuildInfo: build.Info,
			Logger:    p.Spec.Config.Logger,
			Values: map[string]any{
				"maxEventSize": maxEventSize,
				"vaultURL":     r.VaultURL,
			},
		},
	}

	r.mutex.Lock()
	setup := r.setupMap[baseTD.PlatformFullName()]
	r.mutex.Unlock()
	if setup {
		r.log.Debugf("Platform '%s' already setup ", baseTD.PlatformFullName())

	} else {
		// Ensure there are valid commits for Platform components.
		if !api.RegexpCommit.MatchString(build.Info.BrokerCommit) ||
			!api.RegexpCommit.MatchString(build.Info.HTTPSrvCommit) {
			return ctrl.Result{}, log.ErrorN("broker or httpsrv commit from build info is invalid")
		}
		if err := r.setupVault(ctx, baseTD); err != nil {
			return ctrl.Result{}, log.ErrorN("problem setting up vault: %w", err)
		}
		if err := r.ApplyTemplate(ctx, "platform", &baseTD.Data, log); err != nil {
			return ctrl.Result{}, log.ErrorN("problem setting up Platform: %w", err)
		}

		r.mutex.Lock()
		r.setupMap[baseTD.PlatformFullName()] = true
		r.mutex.Unlock()
	}

	td := baseTD.ForComponent("nats", &appsv1.StatefulSet{}, &NATSDefaults, templates.Component{
		Name:          "nats",
		Image:         NATSImage,
		PodSpec:       p.Spec.Config.NATS.PodSpec,
		ContainerSpec: p.Spec.Config.NATS.ContainerSpec,
	})
	if err := r.setupVaultComponent(ctx, td, ""); err != nil {
		return ctrl.Result{}, err
	}
	if rdy, err := r.CompMgr.SetupComponent(ctx, td); !rdy || err != nil {
		return chill(), err
	}

	td = baseTD.ForComponent("broker", &appsv1.DaemonSet{}, &BrokerDefaults, templates.Component{
		Name:          "broker",
		Image:         BrokerImage,
		PodSpec:       p.Spec.Config.Broker.PodSpec,
		ContainerSpec: p.Spec.Config.Broker.ContainerSpec,
	})
	if err := r.setupVaultComponent(ctx, td, ""); err != nil {
		return ctrl.Result{}, err
	}
	if rdy, err := r.CompMgr.SetupComponent(ctx, td); !rdy || err != nil {
		return chill(), err
	}

	td = baseTD.ForComponent("httpsrv", &appsv1.Deployment{}, &HTTPSrvDefaults, templates.Component{
		Name:          "httpsrv",
		Image:         HTTPSrvImage,
		PodSpec:       p.Spec.Config.HTTPSrv.PodSpec,
		ContainerSpec: p.Spec.Config.HTTPSrv.ContainerSpec,
	})
	td.Values["serviceType"] = p.Spec.Config.HTTPSrv.Service.Type
	td.Values["httpPort"] = p.Spec.Config.HTTPSrv.Service.Ports.HTTP
	td.Values["httpsPort"] = p.Spec.Config.HTTPSrv.Service.Ports.HTTPS
	if err := r.setupVaultComponent(ctx, td, ""); err != nil {
		return ctrl.Result{}, err
	}
	if rdy, err := r.CompMgr.SetupComponent(ctx, td); !rdy || err != nil {
		return chill(), err
	}

	// Used by defer func created above.
	ready = true
	log.Debug("Platform reconciled")

	if rdy, err := r.CompMgr.ReconcileApps(ctx, p.Namespace); !rdy || err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// TODO break operations into funcs and move to reusable vault client
func (r *PlatformReconciler) setupVault(ctx context.Context, td *TemplateData) error {
	r.log.Debugf("setting up vault for Platform '%s'", td.PlatformFullName())

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

	r.log.Debugf("vault successfully setup for Platform '%s'", td.Platform.Name)

	return nil
}

func (r *PlatformReconciler) setupVaultComponent(ctx context.Context, td *TemplateData, additionalPolicies string) error {
	r.mutex.Lock()
	setup := r.setupMap[td.ComponentFullName()]
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
	r.setupMap[td.ComponentFullName()] = true
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

// chill waits a few seconds for things to chillax.
func chill() ctrl.Result {
	time.Sleep(time.Second * 3)
	return ctrl.Result{}
}
