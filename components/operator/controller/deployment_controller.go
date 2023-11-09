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

	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/logkf"
)

// DeploymentReconciler reconciles a Deployment object
type DeploymentReconciler struct {
	*Client

	Instance string
	Scheme   *runtime.Scheme

	CompMgr *ComponentManager
	log     *logkf.Logger
}

// SetupWithManager sets up the controller with the Manager.
func (r *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.log = logkf.Global.With(logkf.KeyController, "deployment")
	if err := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Deployment{}).
		Complete(r); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Release{}).
		Complete(r)
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.log.With(
		"namespace", req.Namespace,
		"name", req.Name,
	)
	log.Debugf("reconciling kubefox deployment '%s.%s'", req.Name, req.Namespace)

	if rdy, err := r.CompMgr.ReconcileApps(ctx, req.Namespace); !rdy || err != nil {
		log.Debug("platform not ready, platform controller will reconcile")
		return ctrl.Result{}, err
	}

	log.Debugf("kubefox deployment reconciled '%s.%s'", req.Name, req.Namespace)
	return ctrl.Result{}, nil
}
