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
)

// SnapshotReconciler reconciles a VirtualEnvironmentSnapshot object.
type SnapshotReconciler struct {
	*Client

	log *logkf.Logger
}

// SetupWithManager sets up the controller with the Manager.
func (r *SnapshotReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.log = logkf.Global.With(logkf.KeyController, "Environment")
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.VirtualEnvSnapshot{}).
		Complete(r)
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *SnapshotReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.log.With(
		"namespace", req.Namespace,
		"name", req.Name,
	)
	log.Debugf("reconciling VirtualEnvSnapshot '%s/%s'", req.Namespace, req.Name)

	if err := r.reconcile(ctx, req, log); err != nil {
		if IsFailedWebhookErr(err) {
			log.Debug("reconcile failed because of webhook, retrying in 15 seconds")
			return ctrl.Result{RequeueAfter: time.Second * 15}, nil
		}
		return ctrl.Result{}, err
	}

	log.Debugf("reconciling VirtualEnvSnapshot '%s/%s' done", req.Namespace, req.Name)

	return ctrl.Result{}, nil
}

func (r *SnapshotReconciler) reconcile(ctx context.Context, req ctrl.Request, log *logkf.Logger) error {
	env := &v1alpha1.VirtualEnvSnapshot{}
	if err := r.Get(ctx, req.NamespacedName, env); err != nil {
		return k8s.IgnoreNotFound(err)
	}
	curAppDep := env.DeepCopy()

	k8s.UpdateLabel(env, api.LabelK8sVirtualEnv, env.Data.Source.Name)
	k8s.UpdateLabel(env, api.LabelK8sSourceKind, env.Data.Source.Kind)
	k8s.UpdateLabel(env, api.LabelK8sSourceResourceVersion, env.Data.Source.ResourceVersion)

	if !k8s.DeepEqual(curAppDep.ObjectMeta, env.ObjectMeta) {
		log.Debug("VirtualEnvSnapshot modified, updating")
		return r.Update(ctx, env)
	}

	return nil
}
