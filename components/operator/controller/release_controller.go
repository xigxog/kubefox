/*
Copyright © 2023 XigXog

This Source Code Form is subject to the terms of the Mozilla Public License,
v2.0. If a copy of the MPL was not distributed with this file, You can obtain
one at https://mozilla.org/MPL/2.0/.
*/

package controller

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/xigxog/kubefox/libs/core/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/libs/core/logkf"
)

// ReleaseReconciler reconciles a Release object
type ReleaseReconciler struct {
	*Client

	Instance string
	Scheme   *runtime.Scheme

	cm  *ComponentManager
	log *logkf.Logger
}

//+kubebuilder:rbac:groups=kubefox.xigxog.io,resources=releases,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubefox.xigxog.io,resources=releases/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubefox.xigxog.io,resources=releases/finalizers,verbs=update

// SetupWithManager sets up the controller with the Manager.
func (r *ReleaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.log = logkf.Global.With("controller", "release")
	r.cm = &ComponentManager{
		Instance: r.Instance,
		Client:   r.Client,
		Log:      r.log,
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Release{}).
		Complete(r)
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Release object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *ReleaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.log.With(
		"namespace", req.Namespace,
		"name", req.Name,
	)
	log.Debug("reconciling release")

	rel := &v1alpha1.Release{}
	if err := r.Get(ctx, req.NamespacedName, rel); client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, err

	} else if !apierrors.IsNotFound(err) {
		// TODO move to mutating/admission webhook
		relEnv := &rel.Spec.Environment
		if relEnv.Vars == nil {
			env := &v1alpha1.Environment{}
			if err := r.Get(ctx, nn("", relEnv.Name), env); err != nil {
				return ctrl.Result{}, err
			}
			if relEnv.UID != "" && relEnv.UID != env.UID {
				return ctrl.Result{}, fmt.Errorf("release environment UID does not match environment")
			}
			if relEnv.ResourceVersion != "" && relEnv.ResourceVersion != env.ResourceVersion {
				return ctrl.Result{}, fmt.Errorf("release environment resource version does not match environment")
			}

			relEnv.UID = env.UID
			relEnv.ResourceVersion = env.ResourceVersion
			relEnv.Vars = env.Spec.Vars
			if err := r.Update(ctx, rel); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	r.cm.ReconcileComponents(ctx, req.Namespace)

	log.Debug("release reconciled")
	return ctrl.Result{}, nil
}