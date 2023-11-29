/*
Copyright © 2023 XigXog

This Source Code Form is subject to the terms of the Mozilla Public License,
v2.0. If a copy of the MPL was not distributed with this file, You can obtain
one at https://mozilla.org/MPL/2.0/.
*/

package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/klog/v2"

	"github.com/go-logr/zapr"
	apiv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	kutil "k8s.io/apimachinery/pkg/util/runtime"
	kscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/yaml"

	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/build"
	"github.com/xigxog/kubefox/components/operator/controller"
	"github.com/xigxog/kubefox/components/operator/templates"
	"github.com/xigxog/kubefox/k8s"
	"github.com/xigxog/kubefox/logkf"
	"github.com/xigxog/kubefox/utils"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	kutil.Must(kscheme.AddToScheme(scheme))
	kutil.Must(v1alpha1.AddToScheme(scheme))
	kutil.Must(apiv1.AddToScheme(scheme))
}

var (
	instance, namespace  string
	vaultURL, healthAddr string
	logFormat, logLevel  string
	leaderElection       bool
)

func main() {
	flag.StringVar(&instance, "instance", "", "The name of the KubeFox instance. (required)")
	flag.StringVar(&namespace, "namespace", "", "The Kubernetes Namespace of the Operator. (required)")
	flag.StringVar(&vaultURL, "vault-url", "", "The URL of Vault server. (required)")
	flag.StringVar(&healthAddr, "health-addr", "0.0.0.0:1111", "Address and port the HTTP health server should bind to.")
	flag.BoolVar(&leaderElection, "leader-elect", true, "Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&logFormat, "log-format", "console", `Log format; one of ["json", "console"].`)
	flag.StringVar(&logLevel, "log-level", "debug", `Log level; one of ["debug", "info", "warn", "error"].`)
	flag.Parse()

	utils.CheckRequiredFlag("instance", instance)
	utils.CheckRequiredFlag("namespace", namespace)
	utils.CheckRequiredFlag("vault-url", vaultURL)

	logkf.Global = logkf.
		BuildLoggerOrDie(logFormat, logLevel).
		WithPlatformComponent(api.PlatformComponentOperator)
	defer logkf.Global.Sync()
	ctrl.SetLogger(zapr.NewLogger(logkf.Global.Unwrap().Desugar()))
	klog.SetLogger(zapr.NewLogger(logkf.Global.Unwrap().Desugar()))

	log := logkf.Global

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: server.Options{
			BindAddress: "0",
		},
		HealthProbeBindAddress: healthAddr,
		WebhookServer: webhook.NewServer(webhook.Options{
			CertDir: api.KubeFoxHome,
		}),
		LeaderElection:          leaderElection,
		LeaderElectionID:        "a2e4163f.kubefox.xigxog.io",
		LeaderElectionNamespace: namespace,
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		log.Fatal("unable to start manager", err)
	}

	ctrlClient := &controller.Client{
		Client: k8s.Client{
			Client:     mgr.GetClient(),
			FieldOwner: client.FieldOwner(fmt.Sprintf("%s-operator", instance)),
		},
	}
	compMgr := controller.NewComponentManager(instance, ctrlClient)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	if err := applyCRDs(ctx, ctrlClient); err != nil {
		log.Fatalf("error creating crds: %v", err)
	}
	if err := setupWebhook(ctx, ctrlClient); err != nil {
		log.Fatalf("error creating webhooks: %v", err)
	}
	cancel()

	if err = (&controller.PlatformReconciler{
		Client:    ctrlClient,
		Instance:  instance,
		Namespace: namespace,
		VaultURL:  vaultURL,
		CompMgr:   compMgr,
		LogLevel:  logLevel,
		LogFormat: logFormat,
	}).SetupWithManager(mgr); err != nil {
		log.Fatalf("unable to create Platform controller: %v", err)
	}
	if err = (&controller.SnapshotReconciler{
		Client: ctrlClient,
	}).SetupWithManager(mgr); err != nil {
		log.Fatalf("unable to create VirtualEnvironmentSnapshot controller: %v", err)
	}
	if err = (&controller.AppDeploymentReconciler{
		Client:  ctrlClient,
		CompMgr: compMgr,
	}).SetupWithManager(mgr); err != nil {
		log.Fatalf("unable to create AppDeployment controller: %v", err)
	}
	if err = (&controller.ReleaseReconciler{
		Client: ctrlClient,
	}).SetupWithManager(mgr); err != nil {
		log.Fatalf("unable to create Release controller: %v", err)
	}
	// TODO  ComponentSpec controller

	mgr.GetWebhookServer().Register("/v1alpha1/platform/validate", &webhook.Admission{
		Handler: &controller.PlatformWebhook{
			Client:  ctrlClient,
			Decoder: admission.NewDecoder(scheme),
		},
	})
	mgr.GetWebhookServer().Register("/immutable/validate", &webhook.Admission{
		Handler: &controller.ImmutableWebhook{
			Client:  ctrlClient,
			Decoder: admission.NewDecoder(scheme),
		},
	})

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Fatalf("unable to set up health check: %v", err)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Fatalf("unable to set up ready check: %v", err)
	}

	log.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Fatalf("problem running manager: %v", err)
	}
}

func applyCRDs(ctx context.Context, c *controller.Client) error {
	log := logkf.Global

	created := false
	err := fs.WalkDir(api.EFS, "crds",
		func(path string, d fs.DirEntry, err error) error {
			if d.IsDir() || err != nil {
				return err
			}

			b, err := api.EFS.ReadFile(path)
			if err != nil {
				log.Errorf("unable to read crd %s: %v", path, err)
				return err
			}

			crd := &apiv1.CustomResourceDefinition{}
			if err := yaml.Unmarshal(b, crd); err != nil {
				log.Errorf("unable to parse crd %s: %v", path, err)
				return err
			}

			err = c.Create(ctx, crd)
			switch {
			case errors.IsAlreadyExists(err):
				log.Debugf("crd %s already exists, applying", crd.Name)
				if err = c.Apply(ctx, crd); err != nil {
					log.Errorf("unable to apply crd %s: %v", path, err)
					return err
				}
				return nil

			case err != nil:
				log.Errorf("unable to create crd %s: %v", path, err)
				return err

			default:
				log.Debugf("created crd %s", crd.Name)
				created = true
				return nil
			}
		},
	)
	if err != nil {
		return err
	}

	if created {
		// Need to let API settle before starting controllers.
		log.Debug("crd was created, sleeping to allow API to settle")
		time.Sleep(time.Second * 5)
	}

	return nil
}

func setupWebhook(ctx context.Context, c *controller.Client) error {
	cname := fmt.Sprintf("%s-operator.%s.svc", instance, namespace)
	pkg, err := utils.GeneratePKI(cname, time.Now().AddDate(10, 0, 0))
	if err != nil {
		return err
	}
	if err := os.WriteFile(api.PathTLSCert, []byte(pkg.Cert), 0600); err != nil {
		return err
	}
	if err := os.WriteFile(api.PathTLSKey, []byte(pkg.CertPrivKey), 0600); err != nil {
		return err
	}

	data := &templates.Data{
		Instance: templates.Instance{
			Name:      instance,
			Namespace: namespace,
		},
		Values: map[string]any{
			"caBundle": base64.StdEncoding.EncodeToString([]byte(pkg.CA)),
		},
		BuildInfo: build.Info,
	}
	if err := c.ApplyTemplate(ctx, "instance", data, logkf.Global); err != nil {
		return err
	}

	return nil
}
