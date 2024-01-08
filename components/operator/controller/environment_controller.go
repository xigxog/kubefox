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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/k8s"
	"github.com/xigxog/kubefox/logkf"
)

// EnvironmentReconciler reconciles a Environment object
type EnvironmentReconciler struct {
	*Client

	log *logkf.Logger
}

// SetupWithManager sets up the controller with the Manager.
func (r *EnvironmentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.log = logkf.Global.With(logkf.KeyController, "Environment")
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Environment{}).
		Watches(
			&v1alpha1.VirtualEnvironment{},
			handler.EnqueueRequestsFromMapFunc(r.watchVE),
		).
		Complete(r)
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *EnvironmentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.log.With(
		"name", req.Name,
	)
	log.Debugf("reconciling Environment '%s'", req.Name)
	defer log.Debugf("reconciling Environment '%s' done", req.Name)

	env := &v1alpha1.Environment{}
	if err := r.Get(ctx, k8s.Key("", req.Name), env); err != nil {
		return ctrl.Result{}, k8s.IgnoreNotFound(err)
	}

	// If Environment was deleted but has 'environment-protection' finalizer
	// check if any VirtualEnvironments are using it. If not remove the
	// finalizer.
	if env.DeletionTimestamp != nil && k8s.ContainsFinalizer(env, api.FinalizerEnvironmentProtection) {
		veList := &v1alpha1.VirtualEnvironmentList{}
		if err := r.List(ctx, veList, client.MatchingLabels{api.LabelK8sEnvironment: req.Name}); err != nil {
			return ctrl.Result{}, err
		}

		if l := len(veList.Items); l == 0 {
			k8s.RemoveFinalizer(env, api.FinalizerEnvironmentProtection)

			if err := r.Update(ctx, env); IgnoreFailedWebhookErr(err) != nil {
				return ctrl.Result{}, err

			} else if IsFailedWebhookErr(err) {
				log.Debug("reconcile failed because of webhook, retrying in 15 seconds")
				return ctrl.Result{RequeueAfter: time.Second * 15}, nil
			}

		} else {
			log.Debugf("Environment '%s' is used by %d VirtualEnvironments", env.Name, l)
		}
	}

	return ctrl.Result{}, nil
}

func (r *EnvironmentReconciler) watchVE(ctx context.Context, obj client.Object) []reconcile.Request {
	ve := obj.(*v1alpha1.VirtualEnvironment)
	return []reconcile.Request{
		{
			NamespacedName: k8s.Key("", ve.Spec.Environment),
		},
	}
}
