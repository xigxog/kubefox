/*
Copyright © 2023 XigXog

This Source Code Form is subject to the terms of the Mozilla Public License,
v2.0. If a copy of the MPL was not distributed with this file, You can obtain
one at https://mozilla.org/MPL/2.0/.
*/

package main

import (
	"context"
	"flag"
	"io/fs"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/go-logr/zapr"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	urt "k8s.io/apimachinery/pkg/util/runtime"
	kscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/yaml"

	"github.com/xigxog/kubefox/components/operator/controller"
	"github.com/xigxog/kubefox/libs/api"
	"github.com/xigxog/kubefox/libs/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/logkf"
	"github.com/xigxog/kubefox/libs/core/utils"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	urt.Must(kscheme.AddToScheme(scheme))
	urt.Must(v1alpha1.AddToScheme(scheme))
	urt.Must(v1.AddToScheme(scheme))
}

func main() {
	var (
		instance, namespace   string
		vaultAddr, healthAddr string
		logFormat, logLevel   string
		leaderElection        bool
	)
	flag.StringVar(&instance, "instance", "", "The name of the KubeFox instance. (required)")
	flag.StringVar(&namespace, "namespace", "", "The Kubernetes Namespace of the Operator. (required)")
	flag.StringVar(&vaultAddr, "vault-addr", "", "The host and port of Vault server.")
	flag.StringVar(&healthAddr, "health-addr", "0.0.0.0:1111", "Address and port the HTTP health server should bind to.")
	flag.BoolVar(&leaderElection, "leader-elect", true, "Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&logFormat, "log-format", "console", `Log format; one of ["json", "console"].`)
	flag.StringVar(&logLevel, "log-level", "debug", `Log level; one of ["debug", "info", "warn", "error"].`)
	flag.Parse()

	utils.CheckRequiredFlag("instance", instance)
	utils.CheckRequiredFlag("namespace", namespace)

	logkf.Global = logkf.
		BuildLoggerOrDie(logFormat, logLevel).
		WithService("operator")
	defer logkf.Global.Sync()
	ctrl.SetLogger(zapr.NewLogger(logkf.Global.Unwrap().Desugar()))

	log := logkf.Global
	log.Infof("gitCommit: %s, gitRef: %s", kubefox.GitCommit, kubefox.GitRef)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: server.Options{
			BindAddress: "0",
		},
		HealthProbeBindAddress: healthAddr,
		// WebhookServer: webhook.NewServer(webhook.Options{
		// 	CertDir: "/kubefox/operator/",
		// }),
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

	ctrlClient := &controller.Client{Client: mgr.GetClient()}

	createdCRDs(ctrlClient)

	if err = (&controller.PlatformReconciler{
		Client:    ctrlClient,
		Instance:  instance,
		Namespace: namespace,
		VaultAddr: vaultAddr,
		Scheme:    mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		log.Fatalf("unable to create platform controller", err)
	}
	if err = (&controller.DeploymentReconciler{
		Client:   ctrlClient,
		Instance: instance,
		Scheme:   mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		log.Fatalf("unable to create deployment controller", err)
	}
	if err = (&controller.ReleaseReconciler{
		Client:   ctrlClient,
		Instance: instance,
		Scheme:   mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		log.Fatalf("unable to create release controller", err)
	}

	// if err = (&kubefoxv1alpha1.Platform{}).SetupWebhookWithManager(mgr); err != nil {
	// 	setupLog.Error(err, "unable to create webhook", "webhook", "Platform")
	// 	os.Exit(1)
	// }

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Fatal("unable to set up health check", err)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Fatal("unable to set up ready check", err)
	}

	log.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Fatal("problem running manager", err)
	}
}

func createdCRDs(ctrlClient *controller.Client) {
	log := logkf.Global

	created := false
	fs.WalkDir(api.EFS, "crds",
		func(path string, d fs.DirEntry, err error) error {
			if d.IsDir() || err != nil {
				return err
			}

			b, err := api.EFS.ReadFile(path)
			if err != nil {
				log.Errorf("unable to read crd %s: %v", path, err)
				return err
			}

			crd := &v1.CustomResourceDefinition{}
			if err := yaml.Unmarshal(b, crd); err != nil {
				log.Errorf("unable to parse crd %s: %v", path, err)
				return err
			}

			err = ctrlClient.Create(context.Background(), crd)
			if errors.IsAlreadyExists(err) {
				log.Debugf("crd %s already exists", crd.Name)
				return nil
			}
			if err != nil {
				log.Errorf("unable to create crd %s: %v", path, err)
				return err
			}
			log.Debugf("created crd %s", crd.Name)
			created = true
			return nil
		},
	)

	if created {
		// Need to let API settle before starting controllers.
		log.Debug("crd was created, sleeping to allow API to settle")
		time.Sleep(time.Second * 5)
	}
}
