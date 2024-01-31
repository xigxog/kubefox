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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/xigxog/kubefox/api"
	common "github.com/xigxog/kubefox/api/kubernetes"
	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/k8s"
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
		Watches(
			&v1alpha1.HTTPAdapter{},
			handler.EnqueueRequestsFromMapFunc(r.watchAdapters),
		).
		Watches(
			&v1alpha1.VirtualEnvironment{},
			handler.EnqueueRequestsFromMapFunc(r.watchVE),
		).
		Complete(r)
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *AppDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.log.With(
		"namespace", req.Namespace,
		"name", req.Name,
	)

	appDep := &v1alpha1.AppDeployment{}
	if err := r.Get(ctx, k8s.Key(req.Namespace, req.Name), appDep); err != nil {
		return ctrl.Result{}, k8s.IgnoreNotFound(err)
	}

	log.Debugf("reconciling '%s'", k8s.ToString(appDep))
	defer log.Debugf("reconciling '%s' complete", k8s.ToString(appDep))

	if appDep.DeletionTimestamp.IsZero() {
		if err := r.AddFinalizer(ctx, appDep, api.FinalizerReleaseProtection); err != nil {
			return RetryConflictWebhookErr(k8s.IgnoreNotFound(err))
		}

	} else {
		if k8s.ContainsFinalizer(appDep, api.FinalizerReleaseProtection) {
			veList := &v1alpha1.VirtualEnvironmentList{}
			if err := r.List(ctx, veList); err != nil {
				return ctrl.Result{}, err
			}

			for _, ve := range veList.Items {
				if ve.UsesAppDeployment(appDep.Name) {
					return ctrl.Result{}, nil
				}
			}

			if err := r.RemoveFinalizer(ctx, appDep, api.FinalizerReleaseProtection); err != nil {
				return RetryConflictWebhookErr(k8s.IgnoreNotFound(err))
			}
		}

		return ctrl.Result{}, nil
	}

	if _, err := r.CompMgr.ReconcileApps(ctx, req.Namespace); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *AppDeploymentReconciler) watchAdapters(ctx context.Context, obj client.Object) []reconcile.Request {
	reqs := []reconcile.Request{}

	appDepList := &v1alpha1.AppDeploymentList{}
	if err := r.List(ctx, appDepList, client.InNamespace(obj.GetNamespace())); err != nil {
		r.log.Error(err)
		return reqs
	}

	adapter := obj.(common.Adapter)
	for _, appDep := range appDepList.Items {
		if appDep.HasDependency(adapter.GetName(), adapter.GetComponentType()) {
			reqs = append(reqs, reconcile.Request{
				NamespacedName: k8s.Key(appDep.Namespace, appDep.Name),
			})
		}
	}

	return reqs
}

func (r *AppDeploymentReconciler) watchVE(ctx context.Context, obj client.Object) []reconcile.Request {
	ve := obj.(*v1alpha1.VirtualEnvironment)

	reqs := []reconcile.Request{}
	for _, rel := range ve.Status.ReleaseHistory {
		for _, app := range rel.Apps {
			appDep := &v1alpha1.AppDeployment{}
			if err := r.Get(ctx, k8s.Key(ve.Namespace, app.AppDeployment), appDep); err != nil {
				continue
			}
			if !appDep.DeletionTimestamp.IsZero() && k8s.ContainsFinalizer(appDep, api.FinalizerReleaseProtection) {
				reqs = append(reqs, reconcile.Request{
					NamespacedName: k8s.Key(appDep.Namespace, appDep.Name),
				})
			}
		}
	}

	return reqs
}
