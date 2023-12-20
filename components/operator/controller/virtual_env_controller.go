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

	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/k8s"
	"github.com/xigxog/kubefox/logkf"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// VirtualEnvReconciler reconciles a Release object
type VirtualEnvReconciler struct {
	*Client

	log *logkf.Logger
}

// SetupWithManager sets up the controller with the Manager.
func (r *VirtualEnvReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.log = logkf.Global.With(logkf.KeyController, "virtualenv")
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.VirtualEnv{}).
		Watches(
			&v1alpha1.AppDeployment{},
			handler.EnqueueRequestsFromMapFunc(r.watchAppDeployment),
		).
		Watches(
			&v1alpha1.ClusterVirtualEnv{},
			handler.EnqueueRequestsFromMapFunc(r.watchClusterVirtualEnv),
		).
		Watches(
			&v1alpha1.VirtualEnvSnapshot{},
			handler.EnqueueRequestsFromMapFunc(r.watchVirtualEnvSnapshot),
		).
		Complete(r)
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *VirtualEnvReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.log.With(
		"namespace", req.Namespace,
		"name", req.Name,
	)
	log.Debugf("reconciling VirtualEnv '%s/%s'", req.Namespace, req.Name)
	defer log.Debugf("reconciling VirtualEnv '%s/%s' done", req.Namespace, req.Name)

	env := &v1alpha1.VirtualEnv{}
	if err := r.Get(ctx, req.NamespacedName, env); err != nil {
		return ctrl.Result{}, k8s.IgnoreNotFound(err)
	}
	origEnv := env.DeepCopy()

	if err := r.reconcile(ctx, env, log); IgnoreFailedWebhookErr(err) != nil {
		return ctrl.Result{}, err

	} else if IsFailedWebhookErr(err) {
		log.Debug("reconcile failed because of webhook, retrying in 15 seconds")
		return ctrl.Result{RequeueAfter: time.Second * 15}, nil
	}

	if !k8s.DeepEqual(&env.Status, &origEnv.Status) {
		log.Debug("VirtualEnv status modified, updating")
		if err := r.MergeStatus(ctx, env, origEnv); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *VirtualEnvReconciler) reconcile(ctx context.Context, env *v1alpha1.VirtualEnv, log *logkf.Logger) error {
	now := metav1.Now()

	if env.Spec.Release == nil {
		env.Status.Conditions = k8s.UpdateConditions(now, env.Status.Conditions, &metav1.Condition{
			Type:               api.ConditionTypeReleaseAvailable,
			Status:             metav1.ConditionFalse,
			ObservedGeneration: env.ObjectMeta.Generation,
			Reason:             api.ConditionReasonReleaseEmpty,
			Message:            "No Release set for VirtualEnv.",
		}, &metav1.Condition{
			Type:               api.ConditionTypeReleasePending,
			Status:             metav1.ConditionFalse,
			ObservedGeneration: env.ObjectMeta.Generation,
			Reason:             api.ConditionReasonReleaseEmpty,
			Message:            "No Release set for VirtualEnv.",
		})
		if env.Status.ActiveRelease != nil {
			env.Status.ActiveRelease.ArchiveTime = &now
			env.Status.ReleaseHistory = append([]v1alpha1.ReleaseStatus{*env.Status.ActiveRelease}, env.Status.ReleaseHistory...)
			env.Status.ActiveRelease = nil
		}
		if env.Status.PendingRelease != nil {
			env.Status.PendingRelease.ArchiveTime = &now
			env.Status.ReleaseHistory = append([]v1alpha1.ReleaseStatus{*env.Status.PendingRelease}, env.Status.ReleaseHistory...)
			env.Status.PendingRelease = nil
		}

		// TODO enforce history limits

		return nil
	}

	isActive := env.Status.ActiveRelease != nil &&
		env.Status.ActiveRelease.AppDeployment.ReleaseAppDeployment == env.Spec.Release.AppDeployment
	isPending := env.Status.PendingRelease != nil &&
		env.Status.PendingRelease.AppDeployment.ReleaseAppDeployment == env.Spec.Release.AppDeployment

	if isActive {
		env.Status.PendingRelease = nil
	}
	if !isActive && !isPending {
		isPending = true
		env.Status.PendingRelease = &v1alpha1.ReleaseStatus{
			AppDeployment: v1alpha1.ReleaseAppDeploymentStatus{
				ReleaseAppDeployment: env.Spec.Release.AppDeployment,
			},
			VirtualEnvSnapshot: env.Spec.Release.VirtualEnvSnapshot,
			RequestTime:        now,
		}
	}

	{
		// Update ReleasePending condition.
		var (
			status = metav1.ConditionFalse
			reason = api.ConditionReasonReleaseActive
			msg    = "Release is active."
			err    error
		)
		if env.Status.PendingRelease != nil {
			status = metav1.ConditionTrue
			reason, msg, err = r.conditionReason(ctx, env, env.Status.PendingRelease)
			if err != nil {
				return err
			}

			if reason == "" {
				reason = api.ConditionReasonReleaseActive
				status = metav1.ConditionFalse

				if env.Status.ActiveRelease != nil {
					env.Status.ActiveRelease.ArchiveTime = &now
					env.Status.ReleaseHistory = append([]v1alpha1.ReleaseStatus{*env.Status.ActiveRelease}, env.Status.ReleaseHistory...)
					// TODO enforce history limits
				}
				env.Status.ActiveRelease = env.Status.PendingRelease
				env.Status.ActiveRelease.ActivationTime = &now
				env.Status.PendingRelease = nil
			}
		}
		env.Status.Conditions = k8s.UpdateConditions(now, env.Status.Conditions, &metav1.Condition{
			Type:               api.ConditionTypeReleasePending,
			Status:             status,
			ObservedGeneration: env.ObjectMeta.Generation,
			Reason:             reason,
			Message:            msg,
		})
	}

	{
		// Update ReleaseAvailable condition.
		var (
			status = metav1.ConditionFalse
			reason = api.ConditionReasonReleasePending
			msg    = "Release is pending activation."
			err    error
		)
		if env.Status.ActiveRelease != nil {
			reason, msg, err = r.conditionReason(ctx, env, env.Status.ActiveRelease)
			if err != nil {
				return err
			}
			if reason == "" {
				status = metav1.ConditionTrue
				reason = api.ConditionReasonAppDeploymentAvailable
				msg = "Release AppDeployment, Routes, and Adapters are available."
			}
		}
		env.Status.Conditions = k8s.UpdateConditions(now, env.Status.Conditions, &metav1.Condition{
			Type:               api.ConditionTypeReleaseAvailable,
			Status:             status,
			ObservedGeneration: env.ObjectMeta.Generation,
			Reason:             reason,
			Message:            msg,
		})
	}

	return nil
}

func (r *VirtualEnvReconciler) conditionReason(ctx context.Context, env *v1alpha1.VirtualEnv, rel *v1alpha1.ReleaseStatus) (string, string, error) {
	relPolicies := env.Spec.ReleasePolicies
	if relPolicies == nil {
		relPolicies = &v1alpha1.ReleasePolicies{
			AppDeploymentPolicy: api.AppDeploymentPolicyVersionRequired,
			VirtualEnvPolicy:    api.VirtualEnvPolicySnapshotRequired,
		}
	}
	if relPolicies.AppDeploymentPolicy == api.AppDeploymentPolicyVersionRequired && rel.AppDeployment.Version == "" {
		return api.ConditionReasonPolicyViolation, "AppDeployment version is required but not set.", nil
	}
	if relPolicies.VirtualEnvPolicy == api.VirtualEnvPolicySnapshotRequired && rel.VirtualEnvSnapshot == "" {
		return api.ConditionReasonPolicyViolation, "VirtualEnvSnapshot is required but not set.", nil
	}

	rel.AppDeployment.ObservedGeneration = 0

	appDep := &v1alpha1.AppDeployment{}
	err := r.Get(ctx, k8s.Key(env.Namespace, rel.AppDeployment.Name), appDep)
	switch {
	case err == nil:
		rel.AppDeployment.ObservedGeneration = appDep.Generation
		progressing := k8s.Condition(appDep.Status.Conditions, api.ConditionTypeProgressing)

		switch {
		case rel.AppDeployment.Version != "" && rel.AppDeployment.Version != appDep.Spec.Version:
			return api.ConditionReasonAppDeploymentFailed,
				fmt.Sprintf(`AppDeployment "%s" version "%s" does not match Release version "%s".`,
					appDep.Name, appDep.Spec.Version, rel.AppDeployment.Version), nil

		case progressing.Status == metav1.ConditionFalse &&
			progressing.Reason == api.ConditionReasonProgressDeadlineExceeded:
			return api.ConditionReasonAppDeploymentFailed,
				fmt.Sprintf(`One or more Component Deployments of AppDeployment "%s" failed.`, appDep.Name), nil

		case !k8s.IsAvailable(appDep.Status.Conditions):
			return api.ConditionReasonAppDeploymentUnavailable,
				k8s.Condition(appDep.Status.Conditions, api.ConditionTypeAvailable).Message, nil
		}

	case k8s.IsNotFound(err):
		return api.ConditionReasonAppDeploymentFailed,
			fmt.Sprintf(`AppDeployment "%s" not found.`, rel.AppDeployment.Name), nil

	case err != nil:
		return "", "", err
	}

	if rel.VirtualEnvSnapshot == "" {
		return "", "", nil
	}

	snap := &v1alpha1.VirtualEnvSnapshot{}
	err = r.Get(ctx, k8s.Key(env.Namespace, rel.VirtualEnvSnapshot), snap)
	switch {
	case err == nil:
		switch {
		case snap.Spec.Source.Name != env.Name:
			return api.ConditionReasonVirtualEnvSnapshotFailed,
				fmt.Sprintf(`VirtualEnvSnapshot "%s" source "%s" is not VirtualEnv "%s".`,
					snap.Name, snap.Spec.Source.Name, env.Name), nil

			// TODO validate env
		}

	case k8s.IsNotFound(err):
		return api.ConditionReasonVirtualEnvSnapshotFailed,
			fmt.Sprintf(`VirtualEnvSnapshot "%s" not found.`, rel.VirtualEnvSnapshot), nil

	case err != nil:
		return "", "", err
	}

	return "", "", nil
}

func (r *VirtualEnvReconciler) watchAppDeployment(ctx context.Context, appDep client.Object) []reconcile.Request {
	return r.findEnvs(ctx,
		client.InNamespace(appDep.GetNamespace()),
		client.MatchingLabels{
			api.LabelK8sAppDeployment: appDep.GetName(),
		},
	)
}

func (r *VirtualEnvReconciler) watchClusterVirtualEnv(ctx context.Context, env client.Object) []reconcile.Request {
	return r.findEnvs(ctx,
		client.MatchingLabels{
			api.LabelK8sVirtualEnvParent: env.GetName(),
		},
	)
}

func (r *VirtualEnvReconciler) watchVirtualEnvSnapshot(ctx context.Context, env client.Object) []reconcile.Request {
	return r.findEnvs(ctx,
		client.InNamespace(env.GetNamespace()),
		client.MatchingLabels{
			api.LabelK8sVirtualEnvSnapshot: env.GetName(),
		},
	)
}

func (r *VirtualEnvReconciler) findEnvs(ctx context.Context, opts ...client.ListOption) []reconcile.Request {
	envList := &v1alpha1.VirtualEnvList{}
	if err := r.List(ctx, envList, opts...); err != nil {
		r.log.Error(err)
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(envList.Items))
	for i, rel := range envList.Items {
		requests[i] = reconcile.Request{
			NamespacedName: k8s.Key(rel.Namespace, rel.Name),
		}
	}

	return requests
}
