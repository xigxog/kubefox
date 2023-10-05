/*
Copyright © 2023 XigXog

This Source Code Form is subject to the terms of the Mozilla Public License,
v2.0. If a copy of the MPL was not distributed with this file, You can obtain
one at https://mozilla.org/MPL/2.0/.
*/

package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/xigxog/kubefox/libs/core/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/libs/core/logkf"
)

// DeploymentReconciler reconciles a Deployment object
type DeploymentReconciler struct {
	*Client

	Instance string
	Scheme   *runtime.Scheme

	cm  *ComponentManager
	log *logkf.Logger
}

// SetupWithManager sets up the controller with the Manager.
func (r *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.log = logkf.Global.With("controller", "deployment")
	r.cm = &ComponentManager{
		Instance: r.Instance,
		Client:   r.Client,
		Log:      r.log,
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Deployment{}).
		Complete(r)
}

//+kubebuilder:rbac:groups=kubefox.xigxog.io,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubefox.xigxog.io,resources=deployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubefox.xigxog.io,resources=deployments/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Deployment object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.log.With(
		"namespace", req.Namespace,
		"name", req.Name,
	)
	log.Debug("reconciling deployment")

	// No need to check for ready as platform controller will reconcile.
	if _, err := r.cm.ReconcileComponents(ctx, req.Namespace); err != nil {
		return ctrl.Result{}, err
	}

	log.Debug("deployment reconciled")
	return ctrl.Result{}, nil
}
