/*
Copyright © 2023 XigXog

This Source Code Form is subject to the terms of the Mozilla Public License,
v2.0. If a copy of the MPL was not distributed with this file, You can obtain
one at https://mozilla.org/MPL/2.0/.
*/

package main

import (
	"flag"
	"fmt"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/go-logr/zapr"
	"k8s.io/apimachinery/pkg/runtime"
	urt "k8s.io/apimachinery/pkg/util/runtime"
	kscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/xigxog/kubefox/components/operator/controller"
	"github.com/xigxog/kubefox/libs/core/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/libs/core/logkf"
	"github.com/xigxog/kubefox/libs/core/utils"
)

// Injected at build time.
var (
	GitRef    string
	GitCommit string
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	controller.GitRef = GitRef
	controller.GitCommit = GitCommit

	urt.Must(kscheme.AddToScheme(scheme))
	urt.Must(v1alpha1.AddToScheme(scheme))
}

func main() {
	var instance, namespace, vaultAddr, metricsAddr, probeAddr, logFormat, logLevel string
	var leaderElection bool
	flag.StringVar(&instance, "instance", "", "The name of the KubeFox instance. (required)")
	flag.StringVar(&namespace, "namespace", "", "The Kubernetes Namespace of the Operator. (required)")
	flag.StringVar(&vaultAddr, "vault-addr", "", "The host and port of Vault server.")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&leaderElection, "leader-elect", false, "Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&logFormat, "log-format", "console", `Log format; one of ["json", "console"].`)
	flag.StringVar(&logLevel, "log-level", "debug", `Log level; one of ["debug", "info", "warn", "error"].`)
	flag.Parse()

	utils.CheckRequiredFlag("instance", instance)
	utils.CheckRequiredFlag("namespace", namespace)

	log, err := logkf.BuildLogger(logFormat, logLevel)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	logkf.Global = log
	ctrl.SetLogger(zapr.NewLogger(log.Unwrap().Desugar()))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		WebhookServer: webhook.NewServer(webhook.Options{
			CertDir: "/kubefox/operator/",
		}),
		LeaderElection:   leaderElection,
		LeaderElectionID: "a2e4163f.xigxog.io",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		log.Fatal("unable to start manager", err)
	}

	if err = (&controller.PlatformReconciler{
		Client:    &controller.Client{Client: mgr.GetClient()},
		Instance:  instance,
		Namespace: namespace,
		VaultAddr: vaultAddr,
		Scheme:    mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		log.Fatalf("unable to create platform controller", err)
	}
	if err = (&controller.DeploymentReconciler{
		Client:   &controller.Client{Client: mgr.GetClient()},
		Instance: instance,
		Scheme:   mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		log.Fatalf("unable to create deployment controller", err)
	}
	if err = (&controller.ReleaseReconciler{
		Client:   &controller.Client{Client: mgr.GetClient()},
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
