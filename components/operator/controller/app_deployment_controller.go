// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package controller

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/logkf"
)

// AppDeploymentReconciler reconciles a AppDeployment object
type AppDeploymentReconciler struct {
	*Client

	CompMgr *ComponentManager

	log *logkf.Logger
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.log = logkf.Global.With(logkf.KeyController, "AppDeployment")
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.AppDeployment{}).
		Complete(r)
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *AppDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.log.With(
		"namespace", req.Namespace,
		"name", req.Name,
	)
	log.Debugf("reconciling AppDeployment '%s/%s'", req.Namespace, req.Name)

	if _, err := r.CompMgr.ReconcileApps(ctx, req.Namespace); err != nil {
		return ctrl.Result{}, err
	}

	log.Debugf("reconciling AppDeployment '%s/%s' done", req.Namespace, req.Name)

	return ctrl.Result{}, nil
}
