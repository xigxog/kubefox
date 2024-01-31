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
	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/components/operator/vault"
	"github.com/xigxog/kubefox/k8s"
	"github.com/xigxog/kubefox/logkf"
)

// ReleaseManifestReconciler reconciles a ReleaseManifest object
type ReleaseManifestReconciler struct {
	*Client

	log *logkf.Logger
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReleaseManifestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.log = logkf.Global.With(logkf.KeyController, "ReleaseManifest")
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.ReleaseManifest{}).
		Watches(
			&v1alpha1.VirtualEnvironment{},
			handler.EnqueueRequestsFromMapFunc(r.watchVE),
		).
		Complete(r)
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ReleaseManifestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.log.With(
		"name", req.Name,
	)

	manifest := &v1alpha1.ReleaseManifest{}
	if err := r.Get(ctx, k8s.Key(req.Namespace, req.Name), manifest); err != nil {
		return ctrl.Result{}, k8s.IgnoreNotFound(err)
	}

	log.Debugf("reconciling '%s'", k8s.ToString(manifest))
	defer log.Debugf("reconciling '%s' complete", k8s.ToString(manifest))

	ve := &v1alpha1.VirtualEnvironment{}
	veKey := k8s.Key(manifest.Namespace, manifest.Spec.VirtualEnvironment.Name)
	if err := r.Get(ctx, veKey, ve); k8s.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, err
	}

	inUse := ve.UsesReleaseManifest(manifest.Name)

	if manifest.DeletionTimestamp.IsZero() {
		if err := r.AddFinalizer(ctx, manifest, api.FinalizerReleaseProtection); err != nil {
			return RetryConflictWebhookErr(k8s.IgnoreNotFound(err))
		}

	} else if !inUse && k8s.ContainsFinalizer(manifest, api.FinalizerReleaseProtection) {
		vaultCli, err := vault.GetClient(ctx)
		if err != nil {
			return ctrl.Result{}, err
		}
		if err := vaultCli.DeleteData(ctx, manifest.GetDataKey()); err != nil {
			return ctrl.Result{}, err
		}

		if err := r.RemoveFinalizer(ctx, manifest, api.FinalizerReleaseProtection); err != nil {
			return RetryConflictWebhookErr(k8s.IgnoreNotFound(err))
		}
	}

	return ctrl.Result{}, nil
}

func (r *ReleaseManifestReconciler) watchVE(ctx context.Context, obj client.Object) []reconcile.Request {
	manifestList := &v1alpha1.ReleaseManifestList{}
	err := r.List(ctx, manifestList,
		client.InNamespace(obj.GetNamespace()),
		client.MatchingLabels{api.LabelK8sVirtualEnvironment: obj.GetName()})
	if err != nil {
		r.log.Error(err)
		return []reconcile.Request{}
	}

	reqs := []reconcile.Request{}
	for _, m := range manifestList.Items {
		if !m.DeletionTimestamp.IsZero() && k8s.ContainsFinalizer(&m, api.FinalizerReleaseProtection) {
			reqs = append(reqs, reconcile.Request{
				NamespacedName: k8s.Key(m.Namespace, m.Name),
			})
		}
	}

	return reqs
}
