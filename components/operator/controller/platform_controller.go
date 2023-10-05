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
	"strings"
	"time"

	vapi "github.com/hashicorp/vault/api"
	vauth "github.com/hashicorp/vault/api/auth/kubernetes"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/xigxog/kubefox/components/operator/templates"
	"github.com/xigxog/kubefox/libs/core/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/libs/core/logkf"
	"github.com/xigxog/kubefox/libs/core/utils"
)

// PlatformReconciler reconciles a Platform object
type PlatformReconciler struct {
	*Client

	Instance  string
	Namespace string
	VaultAddr string

	Scheme *runtime.Scheme

	cm  *ComponentManager
	log *logkf.Logger
}

// SetupWithManager sets up the controller with the Manager.
func (r *PlatformReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.log = logkf.Global.With("controller", "platform")
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

//+kubebuilder:rbac:groups=kubefox.xigxog.io,resources=platforms,verbs=get;list;watch
//+kubebuilder:rbac:groups=kubefox.xigxog.io,resources=platforms/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubefox.xigxog.io,resources=platforms/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *PlatformReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	log := r.log.With(
		"namespace", req.Namespace,
		"name", req.Name,
	)
	log.Debug("reconciling platform")

	ns := &v1.Namespace{}
	if err := r.Get(ctx, nn("", req.Namespace), ns); err != nil {
		return ctrl.Result{}, log.ErrorN("unable to fetch namespace: %w", err)
	}

	p := &v1alpha1.Platform{}
	if err := r.Get(ctx, req.NamespacedName, p); client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, log.ErrorN("unable to fetch platform: %w", err)

	} else if apierrors.IsNotFound(err) {
		log.Debug("platform was deleted, removing namespace label")
		delete(ns.Labels, LabelPlatform)
		if err := r.Update(ctx, ns); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	platformReady := false
	defer func() {
		p.Status.Ready = platformReady
		r.Status().Update(ctx, p)
	}()

	// TODO move to admission webhook
	if lbl, found := ns.Labels[LabelPlatform]; found && lbl != req.Name {
		return ctrl.Result{}, log.ErrorN("namespace belongs to platform '%s'", lbl)

	} else if !found {
		ns.Labels[LabelPlatform] = p.Name
		if err := r.Update(ctx, ns); err != nil {
			return ctrl.Result{}, err
		}
	}

	td := &TemplateData{
		Data: templates.Data{
			Instance: templates.Instance{
				Name:      r.Instance,
				Namespace: r.Namespace,
			},
			Component: templates.Component{
				Name:  "vault",
				Image: "ghcr.io/xigxog/vault:1.14.1-v0.0.1", // TODO move image to arg or const
			},
		},
		Obj:      &appsv1.StatefulSet{},
		Template: "vault",
	}
	if rdy, err := r.cm.SetupComponent(ctx, td); !rdy || err != nil {
		return ctrl.Result{RequeueAfter: time.Second * 15}, err
	}
	vaultURL := fmt.Sprintf("https://%s.%s:8200", td.ComponentName(), td.Instance.Namespace)
	if r.VaultAddr != "" {
		vaultURL = fmt.Sprintf("https://%s", r.VaultAddr)
	}

	cm := &v1.ConfigMap{}
	if err := r.Get(ctx, nn(r.Namespace, r.Instance+"-root-ca"), cm); err != nil {
		return ctrl.Result{}, log.ErrorN("unable to fetch root CA configmap: %w", err)
	}

	td = &TemplateData{
		Data: templates.Data{
			Instance: templates.Instance{
				Name:      r.Instance,
				Namespace: r.Namespace,
				RootCA:    string(cm.Data["ca.crt"]),
			},
			Platform: templates.Platform{
				Name:      p.Name,
				Namespace: p.Namespace,
			},
			Owner: []*metav1.OwnerReference{metav1.NewControllerRef(p, p.GroupVersionKind())},
		},
	}
	if err := r.setupVault(ctx, td, vaultURL); err != nil {
		return ctrl.Result{}, log.ErrorN("problem setting up vault: %w", err)
	}

	if err := r.ApplyTemplate(ctx, "platform", &td.Data); err != nil {
		return ctrl.Result{}, log.ErrorN("problem setting up platform: %w", err)
	}

	td.Template = "nats"
	td.Obj = &appsv1.StatefulSet{}
	td.Component = templates.Component{
		Name:  "nats",
		Image: "nats:2.9.21-alpine",
	}
	if rdy, err := r.cm.SetupComponent(ctx, td); !rdy || err != nil {
		return ctrl.Result{}, err
	}

	td.Template = "broker"
	td.Obj = &appsv1.DaemonSet{}
	td.Component = templates.Component{
		Name:  "broker",
		Image: "ghcr.io/xigxog/kubefox/broker:v0.0.1",
	}
	if rdy, err := r.cm.SetupComponent(ctx, td); !rdy || err != nil {
		return ctrl.Result{}, err
	}

	platformReady = true

	if rdy, err := r.cm.ReconcileComponents(ctx, req.Namespace); err != nil {
		return ctrl.Result{}, err
	} else if !rdy {
		return ctrl.Result{RequeueAfter: time.Second * 3}, nil
	}

	log.Debug("platform reconciled")

	return ctrl.Result{}, nil
}

func (r *PlatformReconciler) setupVault(ctx context.Context, td *TemplateData, url string) error {
	vault, err := r.vaultClient(ctx, url, []byte(td.Instance.RootCA))
	if err != nil {
		return err
	}

	uName := fmt.Sprintf("%s-%s", td.PlatformFullName(), td.Platform.Namespace)
	if !strings.HasPrefix(uName, "kubefox") {
		uName = "kubefox-" + uName
	}
	pkiPath := "pki/int/platform/" + uName
	natsPath := "nats/platform/" + uName
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
			"issuing_certificates":    url + "/v1/" + pkiPath + "/ca",
			"crl_distribution_points": url + "/v1/" + pkiPath + "/crl",
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
		"allow_localhost":    true,
		"allowed_domains":    td.PlatformFullName() + "-nats," + td.PlatformFullName() + "-nats." + td.Namespace(),
		"allow_bare_domains": true,
		"max_ttl":            TenYears,
	})
	if err != nil {
		return err
	}
	_, err = vault.Logical().Write(pkiPath+"/roles/broker", map[string]interface{}{
		"issuer_ref":         "default",
		"allow_localhost":    true,
		"allowed_domains":    "localhost",
		"allow_bare_domains": true,
		"max_ttl":            TenYears,
	})
	if err != nil {
		return err
	}

	if cfg, _ := vault.Sys().MountConfig(natsPath); cfg == nil {
		err = vault.Sys().Mount(natsPath, &vapi.MountInput{
			Type: "nats",
		})
		if err != nil {
			return err
		}
		_, err := vault.Logical().Write(natsPath+"/config", map[string]interface{}{
			"service_url": "nats://" + td.PlatformFullName() + "-nats." + td.Namespace() + ":4222",
		})
		if err != nil {
			return err
		}
	}

	err = vault.Sys().PutPolicyWithContext(ctx, uName+"-broker", `
// issue nats certs
path "`+pkiPath+`/issue/nats" {
	capabilities = ["create", "update"]
}
// issue broker certs
path "`+pkiPath+`/issue/broker" {
	capabilities = ["create", "update"]
}
// issue NATS JWTs
path "`+natsPath+`/config" {
	capabilities = ["read"]
}
path "`+natsPath+`/jwt/*" {
	capabilities = ["create"]
}
	`)
	if err != nil {
		return err
	}

	_, err = vault.Logical().Write("auth/kubernetes/role/"+uName+"-broker", map[string]interface{}{
		"bound_service_account_names":      td.PlatformFullName() + "-broker",
		"bound_service_account_namespaces": td.Namespace(),
		"token_policies":                   uName + "-broker,kv-kubefox-reader",
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *PlatformReconciler) vaultClient(ctx context.Context, url string, caCert []byte) (*vapi.Client, error) {
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

	jwt, err := utils.GetSvcAccountToken(r.Namespace, r.Instance+"-operator")
	if err != nil {
		return nil, err
	}
	auth, err := vauth.NewKubernetesAuth("kubefox-operator", vauth.WithServiceAccountToken(jwt))
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

	watcher, err := vault.NewLifetimeWatcher(&vapi.LifetimeWatcherInput{Secret: authInfo})
	if err != nil {
		return nil, fmt.Errorf("error starting Vault token renewer: %w", err)
	}
	go watcher.Start()

	return vault, nil
}
