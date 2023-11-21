/*
Copyright © 2023 XigXog

This Source Code Form is subject to the terms of the Mozilla Public License,
v2.0. If a copy of the MPL was not distributed with this file, You can obtain
one at https://mozilla.org/MPL/2.0/.
*/

package controller

import (
	"context"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/xigxog/kubefox/api"
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

func (r *AppDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.log.With(
		"namespace", req.Namespace,
		"name", req.Name,
	)
	log.Debugf("reconciling AppDeployment '%s.'%s'", req.Name, req.Namespace)

	result, err := r.reconcile(ctx, req, log)
	if IsFailedWebhookErr(err) {
		log.Debug("reconcile failed because of webhook, retryin in 15 seconds")
		return ctrl.Result{RequeueAfter: time.Second * 15}, nil
	}

	log.Debugf("reconciling AppDeployment '%s.'%s' done", req.Name, req.Namespace)
	if _, err := r.CompMgr.ReconcileApps(ctx, req.Namespace); err != nil {
		log.Error(err)
	}

	return result, err
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *AppDeploymentReconciler) reconcile(ctx context.Context, req ctrl.Request, log *logkf.Logger) (ctrl.Result, error) {
	appDep := &v1alpha1.AppDeployment{}
	err := r.Get(ctx, req.NamespacedName, appDep)
	if IgnoreNotFound(err) != nil {
		return ctrl.Result{}, err
	}

	updated := UpdateLabel(appDep, api.LabelK8sAppVersion, appDep.Spec.Version)
	updated = updated || UpdateLabel(appDep, api.LabelK8sAppCommit, appDep.Spec.App.Commit)
	updated = updated || UpdateLabel(appDep, api.LabelK8sAppTag, appDep.Spec.App.Tag)
	updated = updated || UpdateLabel(appDep, api.LabelK8sAppBranch, appDep.Spec.App.Branch)
	if updated {
		return ctrl.Result{Requeue: true}, r.Update(ctx, appDep)
	}

	// TODO update conditions

	return ctrl.Result{}, nil
}
