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
	"github.com/xigxog/kubefox/k8s"
	"github.com/xigxog/kubefox/logkf"
	"github.com/xigxog/kubefox/utils"
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

	if err := r.reconcile(ctx, req, log); err != nil {
		if IsFailedWebhookErr(err) {
			log.Debug("reconcile failed because of webhook, retrying in 15 seconds")
			return ctrl.Result{RequeueAfter: time.Second * 15}, nil
		}
		return ctrl.Result{}, err
	}

	if _, err := r.CompMgr.ReconcileApps(ctx, req.Namespace); err != nil {
		return ctrl.Result{}, err
	}

	log.Debugf("reconciling AppDeployment '%s/%s' done", req.Namespace, req.Name)

	return ctrl.Result{}, nil
}

func (r *AppDeploymentReconciler) reconcile(ctx context.Context, req ctrl.Request, log *logkf.Logger) error {
	appDep := &v1alpha1.AppDeployment{}
	if err := r.Get(ctx, req.NamespacedName, appDep); err != nil {
		return k8s.IgnoreNotFound(err)
	}

	if k8s.UpdateLabel(appDep, api.LabelK8sAppVersion, appDep.Spec.Version) ||
		k8s.UpdateLabel(appDep, api.LabelK8sAppCommit, appDep.Spec.App.Commit) ||
		k8s.UpdateLabel(appDep, api.LabelK8sAppCommitShort, utils.ShortCommit(appDep.Spec.App.Commit)) ||
		k8s.UpdateLabel(appDep, api.LabelK8sAppTag, appDep.Spec.App.Tag) ||
		k8s.UpdateLabel(appDep, api.LabelK8sAppBranch, appDep.Spec.App.Branch) {

		log.Debug("AppDeployment modified, updating")
		return r.Apply(ctx, appDep)
	}

	// TODO update conditions

	return nil
}
