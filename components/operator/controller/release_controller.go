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
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/logkf"
)

// TODO add "with" to logs

// ReleaseReconciler reconciles a Release object
type ReleaseReconciler struct {
	*Client

	CompMgr *ComponentManager

	platformEvtHandler     handler.Funcs
	lastPlatformUpdateTime time.Time

	log *logkf.Logger
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReleaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.log = logkf.Global.With(logkf.KeyController, "release")
	r.platformEvtHandler.UpdateFunc = r.platformUpdate
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Release{}).
		Watches(&v1alpha1.Platform{}, r.platformEvtHandler). // TODO change to enqueue release reconcile, see plat ctrl
		Complete(r)
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ReleaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.log.With(
		"namespace", req.Namespace,
		"name", req.Name,
	)
	log.Debugf("reconciling Release '%s.'%s'", req.Name, req.Namespace)

	result, err := r.reconcile(ctx, req, log)
	if IsFailedWebhookErr(err) {
		log.Debug("reconcile failed because of webhook, retryin in 15 seconds")
		return ctrl.Result{RequeueAfter: time.Second * 15}, nil
	}

	log.Debugf("reconciling Release '%s.'%s' done", req.Name, req.Namespace)
	if _, err := r.CompMgr.ReconcileApps(ctx, req.Namespace); err != nil {
		log.Error(err)
	}

	return result, err
}

func (r *ReleaseReconciler) reconcile(ctx context.Context, req ctrl.Request, log *logkf.Logger) (ctrl.Result, error) {

	rel := &v1alpha1.Release{}
	err := r.Get(ctx, req.NamespacedName, rel)
	if IgnoreNotFound(err) != nil {
		return ctrl.Result{}, err
	}
	if apierrors.IsNotFound(err) {
		platform, err := r.GetPlatform(ctx, req.Namespace)
		if err != nil {
			return ctrl.Result{}, IgnoreNotFound(err)
		}
		if rel := platform.FindRelease(req.Name); rel != nil {
			return ctrl.Result{}, r.updatePlatformRelease(ctx, platform, rel)
		}
		return ctrl.Result{}, nil
	}
	if rel.DeletionTimestamp != nil {
		return ctrl.Result{}, r.releaseDeleted(ctx, rel)
	}

	if rel.Status.FailureTime != nil {
		log.Debug("release failed,  updating conditions")
		return ctrl.Result{}, r.updateRelease(ctx, rel)
	}
	if rel.Status.PendingTime != nil {
		log.Debug("release already pending, updating conditions")
		return ctrl.Result{}, r.updateRelease(ctx, rel)
	}

	now := metav1.Now()
	rel.Status.CreationTime = rel.CreationTimestamp
	rel.Status.PendingTime = &now
	rel.Status.LastTransitionTime = now

	// Ensure AppDeployment exists and versions match.
	appDep := &v1alpha1.AppDeployment{}
	if err := r.Get(ctx, Key(rel.Namespace, rel.Spec.AppDeployment.Name), &v1alpha1.AppDeployment{}); IgnoreNotFound(err) != nil {
		return ctrl.Result{}, err
	} else if apierrors.IsNotFound(err) {
		rel.Status.FailureTime = &now
		rel.Status.FailureMessage = err.Error()
		return ctrl.Result{}, r.updateRelease(ctx, rel)
	}
	if rel.Spec.Version != appDep.Spec.Version {
		rel.Status.FailureTime = &now
		rel.Status.FailureMessage = fmt.Sprintf(
			"version conflict: Release version '%s' does not match AppDeployment version '%s'",
			rel.Spec.Version, appDep.Spec.Version,
		)
		return ctrl.Result{}, r.updateRelease(ctx, rel)
	}

	// Ensure ResolvedEnvironment exists.
	if _, err := r.GetResolvedEnvironment(ctx, rel); IgnoreNotFound(err) != nil {
		return ctrl.Result{}, err
	} else if apierrors.IsNotFound(err) {
		rel.Status.FailureTime = &now
		rel.Status.FailureMessage = err.Error()
		return ctrl.Result{}, r.updateRelease(ctx, rel)
	}

	// TODO requeue in max release age
	return ctrl.Result{}, r.updateRelease(ctx, rel)
}

func (r *ReleaseReconciler) platformUpdate(ctx context.Context, evt event.UpdateEvent, limit workqueue.RateLimitingInterface) {
	go func() {
		platform := evt.ObjectNew.(*v1alpha1.Platform)
		r.log.Debugf("reconciling Releases for Platform '%s'", platform.Name)

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
		defer cancel()

		success := true
		for envName, env := range platform.Spec.Environments { // TODO if no envs, remove all finalizers
			for _, rel := range env.SupersededReleases {
				success = false
				if err := r.updatePlatformRelease(ctx, platform, rel); err != nil {
					r.log.Warnf("error updating superseded Release '%s' from Platform environment '%s': %v",
						rel.Name, envName, err)
				}
			}
			// TODO if no active, might still be what was active with finalizer
			if err := r.updatePlatformRelease(ctx, platform, env.Release); err != nil {
				success = false
				r.log.Warnf("error updating active Release '%s' from Platform '%s' Environment '%s': %v",
					env.Release.Name, platform.Name, envName, err)
			}
		}
		if success {
			r.lastPlatformUpdateTime = time.Now()
		}
		r.log.Debugf("reconciling Releases for Platform '%s' done", platform.Name)
	}()
}

func (r *ReleaseReconciler) releaseDeleted(ctx context.Context, rel *v1alpha1.Release) error {
	platform, err := r.GetPlatform(ctx, rel.Namespace)
	if err != nil {
		return err
	}

	env := platform.Environment(rel.Spec.Environment.Name)
	if env.Release != nil && env.Release.Name == rel.Name {
		r.log.Warnf(
			"active Release '%s' for Platform '%s' in Environment '%s' was deleted, finalizer will not be removed",
			rel.Name, platform.Name, rel.Spec.Environment.Name,
		)
		return nil
	}
	r.log.Infof("Release '%s' for Platform '%s' in Environment '%s' was deleted",
		rel.Name, platform.Name, rel.Spec.Environment.Name)

	// if _, found := env.SupersededReleases[rel.Name]; found {
	// 	env.SupersededReleases[rel.Name] = nil
	// 	needsUpdate = true
	// }
	// if needsUpdate {
	// 	if err := r.Merge(ctx, p); err != nil {
	// 		return err
	// 	}
	// }

	if controllerutil.RemoveFinalizer(rel, api.FinalizerReleaseProtection) {
		r.log.Debugf("removing finalizer for Release '%s' for Platform '%s' in Environment '%s'",
			rel.Name, platform.Name, rel.Spec.Environment.Name)

		if err := r.updateAppDep(ctx, rel, controllerutil.RemoveFinalizer); err != nil {
			return err
		}
		if err := r.updateResolvedEnv(ctx, rel, controllerutil.RemoveFinalizer); err != nil {
			return err
		}

		return r.Merge(ctx, rel)
	}

	return nil
}

func (r *ReleaseReconciler) updatePlatformRelease(ctx context.Context, platform *v1alpha1.Platform, platformRel *v1alpha1.PlatformEnvRelease) error {
	if platformRel == nil {
		return nil
	}

	rel := &v1alpha1.Release{}
	err := r.Get(ctx, Key(platform.Namespace, platformRel.Name), rel)
	if IgnoreNotFound(err) != nil {
		return err
	}

	if apierrors.IsNotFound(err) {
		// Simulate deleted.
		now := metav1.Now()
		rel = &v1alpha1.Release{
			ObjectMeta: metav1.ObjectMeta{
				Name:              platformRel.Name,
				Namespace:         platform.Namespace,
				DeletionTimestamp: &now,
			},
			Spec:   platformRel.ReleaseSpec,
			Status: platformRel.ReleaseStatus,
		}
	}
	if rel.DeletionTimestamp != nil {
		return r.releaseDeleted(ctx, rel)
	}

	if platformRel.LastTransitionTime.Time.Before(r.lastPlatformUpdateTime) {
		// We've already processed, can skip.
		return nil
	}

	rel.Status = platformRel.ReleaseStatus
	r.log.Debugf("updating Release '%s' status from Platform '%s'", rel.Name, platform.Name)

	return r.ApplyStatus(ctx, rel) // will trigger reconcile if diff
}

func (r *ReleaseReconciler) updateRelease(ctx context.Context, rel *v1alpha1.Release) error {
	var status api.ReleaseStatus
	var finalizerFunc func(o client.Object, finalizer string) bool
	switch {
	case rel.Status.FailureTime != nil:
		status = api.ReleaseStatusFailed
		finalizerFunc = controllerutil.RemoveFinalizer

	case rel.Status.SupersededTime != nil:
		status = api.ReleaseStatusSuperseded
		finalizerFunc = controllerutil.RemoveFinalizer

	case rel.Status.ReleaseTime != nil:
		status = api.ReleaseStatusReleased
		finalizerFunc = controllerutil.AddFinalizer

	case rel.Status.PendingTime != nil:
		status = api.ReleaseStatusPending
		finalizerFunc = controllerutil.AddFinalizer
	}
	needsUpdate := finalizerFunc(rel, api.FinalizerReleaseProtection)
	needsUpdate = needsUpdate || UpdateLabel(rel, api.LabelK8sReleaseStatus, string(status))

	if needsUpdate {
		if err := r.updateAppDep(ctx, rel, finalizerFunc); err != nil {
			return err
		}
		if err := r.updateResolvedEnv(ctx, rel, finalizerFunc); err != nil {
			return err
		}
		if err := r.Merge(ctx, rel); err != nil {
			return err
		}
	}

	// TODO update conditions based on timestamps

	if rel.Status.LastTransitionTime.IsZero() {
		rel.Status.LastTransitionTime = metav1.Now()
	}

	return r.ApplyStatus(ctx, rel)
}

func (r *ReleaseReconciler) updateAppDep(ctx context.Context, rel *v1alpha1.Release, finalizerFunc func(o client.Object, finalizer string) bool) error {
	appDep := &v1alpha1.AppDeployment{}
	err := r.Get(ctx, Key(rel.Namespace, rel.Spec.AppDeployment.Name), appDep)
	if IgnoreNotFound(err) != nil {
		return err
	} else if apierrors.IsNotFound(err) {
		return nil
	}

	UpdateLabel(rel, api.LabelK8sAppVersion, appDep.Spec.Version)
	UpdateLabel(rel, api.LabelK8sAppCommit, appDep.Spec.App.Commit)
	UpdateLabel(rel, api.LabelK8sAppTag, appDep.Spec.App.Tag)
	UpdateLabel(rel, api.LabelK8sAppBranch, appDep.Spec.App.Branch)
	UpdateLabel(rel, api.LabelK8sAppDeployment, appDep.Name)

	return r.updateFinalizer(ctx, appDep, api.LabelK8sAppDeployment, finalizerFunc)
}

func (r *ReleaseReconciler) updateResolvedEnv(ctx context.Context, rel *v1alpha1.Release, finalizerFunc func(o client.Object, finalizer string) bool) error {
	resEnv, err := r.GetResolvedEnvironment(ctx, rel)
	if IgnoreNotFound(err) != nil {
		return err
	} else if apierrors.IsNotFound(err) {
		return nil
	}

	UpdateLabel(rel, api.LabelK8sEnvironment, rel.Spec.Environment.Name)
	UpdateLabel(rel, api.LabelK8sResolvedEnvironment, resEnv.Name)

	return r.updateFinalizer(ctx, resEnv, api.LabelK8sResolvedEnvironment, finalizerFunc)
}

func (r *ReleaseReconciler) updateFinalizer(ctx context.Context, obj client.Object, label string, finalizerFunc func(o client.Object, finalizer string) bool) error {
	if finalizerFunc(obj, api.FinalizerReleaseProtection) {
		relList := &v1alpha1.ReleaseList{}
		if err := r.List(ctx, relList, client.MatchingLabels{label: obj.GetName()}); err != nil {
			return err
		}

		// Search all Releases and ensure they all (don't) have the
		// ReleaseProtection. If there is a mismatch don't update the object.
		containsFinalizer := controllerutil.ContainsFinalizer(obj, api.FinalizerReleaseProtection)
		for _, r := range relList.Items {
			if containsFinalizer != controllerutil.ContainsFinalizer(&r, api.FinalizerReleaseProtection) {
				return nil
			}
		}

		if err := r.Merge(ctx, obj); err != nil {
			return err
		}
	}

	return nil
}
