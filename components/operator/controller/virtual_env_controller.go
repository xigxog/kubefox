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

	requeueAfter, err := r.reconcile(ctx, env, log)
	if IgnoreFailedWebhookErr(err) != nil {
		return ctrl.Result{}, err

	} else if IsFailedWebhookErr(err) {
		log.Debug("reconcile failed because of webhook, retrying in 15 seconds")
		return ctrl.Result{RequeueAfter: time.Second * 15}, nil
	}

	// Prune history if limit is exceeded.
	count := int(env.Spec.ReleasePolicy.HistoryLimits.Count)
	if count <= 0 {
		count = api.DefaultReleaseHistoryLimitCount
	}
	if len(env.Status.ReleaseHistory) > count {
		env.Status.ReleaseHistory = env.Status.ReleaseHistory[:count]
	}

	// Update status if changed.
	if !k8s.DeepEqual(&env.Status, &origEnv.Status) {
		log.Debug("VirtualEnv status modified, updating")
		if err := r.MergeStatus(ctx, env, origEnv); k8s.IgnoreNotFound(err) != nil {
			return ctrl.Result{}, err
		}
	}

	// Update spec if Release was "rolled back".
	if !k8s.DeepEqual(&env.Spec, &origEnv.Spec) {
		log.Debug("VirtualEnv spec modified, updating")
		if err := r.Merge(ctx, env, origEnv); k8s.IgnoreNotFound(err) != nil {
			return ctrl.Result{}, err
		}
	}

	if requeueAfter > 0 {
		log.Debugf("Release pending, requeuing after %s to check deadline", requeueAfter)
	}

	return ctrl.Result{RequeueAfter: requeueAfter}, nil
}

func (r *VirtualEnvReconciler) reconcile(ctx context.Context, env *v1alpha1.VirtualEnv, log *logkf.Logger) (time.Duration, error) {
	now := metav1.Now()
	pending := k8s.Condition(env.Status.Conditions, api.ConditionTypeReleasePending)

	if pending.Status == metav1.ConditionFalse &&
		pending.Reason == api.ConditionReasonPendingDeadlineExceeded &&
		env.Status.PendingReleaseFailed && !env.Status.ActiveRelease.Equals(env.Spec.Release) {

		// The pending Release failed to activate before deadline was exceeded.
		// Rollback to active Release.
		origEnv := env.DeepCopy()
		if active := env.Status.ActiveRelease; active != nil {
			env.Spec.Release = &v1alpha1.Release{
				AppDeployment:      active.AppDeployment.ReleaseAppDeployment,
				VirtualEnvSnapshot: active.VirtualEnvSnapshot,
			}
		} else {
			env.Spec.Release = nil
		}

		return 0, r.Merge(ctx, env, origEnv)
	}

	// Reset flag to default so rollback does not occur again.
	env.Status.PendingReleaseFailed = false

	if env.Spec.Release == nil {
		pendingReason, pendingMsg := api.ConditionReasonNoRelease, "No Release set for VirtualEnv."
		if pending.Reason == api.ConditionReasonPendingDeadlineExceeded {
			// Retain reason and message if the deadline was exceeded previously.
			pendingReason = pending.Reason
			pendingMsg = pending.Message
		}

		env.Status.Conditions = k8s.UpdateConditions(now, env.Status.Conditions, &metav1.Condition{
			Type:               api.ConditionTypeReleaseAvailable,
			Status:             metav1.ConditionFalse,
			ObservedGeneration: env.ObjectMeta.Generation,
			Reason:             api.ConditionReasonNoRelease,
			Message:            "No Release set for VirtualEnv.",
		}, &metav1.Condition{
			Type:               api.ConditionTypeReleasePending,
			Status:             metav1.ConditionFalse,
			ObservedGeneration: env.ObjectMeta.Generation,
			Reason:             pendingReason,
			Message:            pendingMsg,
		})
		if env.Status.ActiveRelease != nil {
			env.Status.ActiveRelease.ArchiveTime = &now
			env.Status.ActiveRelease.ArchiveReason = api.ArchiveReasonSuperseded
			env.Status.ReleaseHistory = append([]v1alpha1.ReleaseStatus{*env.Status.ActiveRelease}, env.Status.ReleaseHistory...)
			env.Status.ActiveRelease = nil
		}
		if env.Status.PendingRelease != nil {
			env.Status.PendingRelease.ArchiveTime = &now
			env.Status.PendingRelease.ArchiveReason = api.ArchiveReasonSuperseded
			env.Status.ReleaseHistory = append([]v1alpha1.ReleaseStatus{*env.Status.PendingRelease}, env.Status.ReleaseHistory...)
			env.Status.PendingRelease = nil
		}

		return 0, nil
	}

	var remainingDeadline time.Duration
	isActive := env.Status.ActiveRelease != nil &&
		env.Status.ActiveRelease.AppDeployment.ReleaseAppDeployment == env.Spec.Release.AppDeployment

	if isActive {
		env.Status.PendingRelease = nil

	} else if !env.Status.PendingRelease.Equals(env.Spec.Release) {
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
		if pending.Reason == api.ConditionReasonPendingDeadlineExceeded {
			// Retain reason and message if the deadline was exceeded previously.
			reason = pending.Reason
			msg = pending.Message
		}
		if env.Status.PendingRelease != nil {
			status = metav1.ConditionTrue
			remainingDeadline = env.ReleasePendingDeadline() - env.ReleasePendingDuration()
			if reason, msg, err = r.updateProblems(ctx, now, env, env.Status.PendingRelease); err != nil {
				return 0, err
			}

			switch {
			case reason == "":
				status = metav1.ConditionFalse
				reason = api.ConditionReasonReleaseActive
				remainingDeadline = 0

				if env.Status.ActiveRelease != nil {
					env.Status.ActiveRelease.ArchiveTime = &now
					env.Status.ActiveRelease.ArchiveReason = api.ArchiveReasonSuperseded
					env.Status.ReleaseHistory = append([]v1alpha1.ReleaseStatus{*env.Status.ActiveRelease}, env.Status.ReleaseHistory...)
				}
				env.Status.ActiveRelease = env.Status.PendingRelease
				env.Status.ActiveRelease.ActivationTime = &now
				env.Status.PendingRelease = nil

			case remainingDeadline <= 0:
				status = metav1.ConditionFalse
				reason = api.ConditionReasonPendingDeadlineExceeded
				msg = "Deadline exceeded before pending Release could be activated."

				env.Status.PendingRelease.ArchiveTime = &now
				env.Status.PendingRelease.ArchiveReason = api.ArchiveReasonPendingDeadlineExceeded
				env.Status.ReleaseHistory = append([]v1alpha1.ReleaseStatus{*env.Status.PendingRelease}, env.Status.ReleaseHistory...)
				env.Status.PendingRelease = nil
				env.Status.PendingReleaseFailed = true
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
			if reason, msg, err = r.updateProblems(ctx, now, env, env.Status.ActiveRelease); err != nil {
				return 0, err
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

	return remainingDeadline, nil
}

func (r *VirtualEnvReconciler) updateProblems(ctx context.Context, now metav1.Time,
	env *v1alpha1.VirtualEnv, rel *v1alpha1.ReleaseStatus) (reason string, msg string, err error) {

	rel.AppDeployment.ObservedGeneration = 0
	rel.Problems = nil

	appDep := &v1alpha1.AppDeployment{}
	err = r.Get(ctx, k8s.Key(env.Namespace, rel.AppDeployment.Name), appDep)
	switch {
	case err == nil:
		rel.AppDeployment.ObservedGeneration = appDep.Generation
		progressing := k8s.Condition(appDep.Status.Conditions, api.ConditionTypeProgressing)

		if !k8s.IsAvailable(appDep.Status.Conditions) {
			reason, msg = api.ConditionReasonAppDeploymentUnavailable,
				k8s.Condition(appDep.Status.Conditions, api.ConditionTypeAvailable).Message

			value := string(metav1.ConditionFalse)
			rel.Problems = append(rel.Problems, v1alpha1.Problem{
				Type:         api.ProblemTypeAppDeploymentUnavailable,
				ObservedTime: api.UncomparableTime(now),
				Message:      msg,
				Causes: []v1alpha1.ProblemSource{
					{
						Kind:               api.ProblemSourceKindAppDeployment,
						Name:               appDep.Name,
						ObservedGeneration: appDep.Generation,
						Path:               "$.status.conditions[?(@.type=='Available')].status",
						Value:              &value,
					},
				},
			})
		}

		if progressing.Status == metav1.ConditionFalse &&
			progressing.Reason == api.ConditionReasonComponentDeploymentFailed {
			reason, msg = api.ConditionReasonAppDeploymentFailed,
				fmt.Sprintf(`One or more Component Deployments of AppDeployment "%s" failed.`, appDep.Name)

			value := fmt.Sprintf("%s, %s", metav1.ConditionFalse, api.ConditionReasonComponentDeploymentFailed)
			rel.Problems = append(rel.Problems, v1alpha1.Problem{
				Type:         api.ProblemTypeAppDeploymentFailed,
				ObservedTime: api.UncomparableTime(now),
				Message:      msg,
				Causes: []v1alpha1.ProblemSource{
					{
						Kind:               api.ProblemSourceKindAppDeployment,
						Name:               appDep.Name,
						ObservedGeneration: appDep.Generation,
						Path:               "$.status.conditions[?(@.type=='Progressing')].status,reason",
						Value:              &value,
					},
				},
			})
		}

		if rel.AppDeployment.Version != "" && rel.AppDeployment.Version != appDep.Spec.Version {
			reason, msg = api.ConditionReasonAppDeploymentFailed,
				fmt.Sprintf(`AppDeployment "%s" version "%s" does not match Release version "%s".`,
					appDep.Name, appDep.Spec.Version, rel.AppDeployment.Version)

			rel.Problems = append(rel.Problems, v1alpha1.Problem{
				Type:         api.ProblemTypeAppDeploymentFailed,
				ObservedTime: api.UncomparableTime(now),
				Message:      msg,
				Causes: []v1alpha1.ProblemSource{
					{
						Kind:               api.ProblemSourceKindRelease,
						ObservedGeneration: env.Generation,
						Path:               "$.appDeployment.version",
						Value:              &rel.AppDeployment.Version,
					},
					{
						Kind:               api.ProblemSourceKindAppDeployment,
						Name:               appDep.Name,
						ObservedGeneration: appDep.Generation,
						Path:               "$.spec.version",
						Value:              &appDep.Spec.Version,
					},
				},
			})
		}

	case k8s.IsNotFound(err):
		reason, msg = api.ConditionReasonAppDeploymentFailed,
			fmt.Sprintf(`AppDeployment "%s" not found.`, rel.AppDeployment.Name)

		rel.Problems = append(rel.Problems, v1alpha1.Problem{
			Type:         api.ProblemTypeAppDeploymentFailed,
			ObservedTime: api.UncomparableTime(now),
			Message:      msg,
			Causes: []v1alpha1.ProblemSource{
				{
					Kind: api.ProblemSourceKindAppDeployment,
					Name: rel.AppDeployment.Name,
				},
			},
		})

	case k8s.IgnoreNotFound(err) != nil:
		return
	}

	if rel.VirtualEnvSnapshot != "" {
		snap := &v1alpha1.VirtualEnvSnapshot{}
		err = r.Get(ctx, k8s.Key(env.Namespace, rel.VirtualEnvSnapshot), snap)
		switch {
		case err == nil:
			if snap.Spec.Source.Name != env.Name {
				reason, msg = api.ConditionReasonVirtualEnvSnapshotFailed,
					fmt.Sprintf(`VirtualEnvSnapshot "%s" source "%s" is not VirtualEnv "%s".`,
						snap.Name, snap.Spec.Source.Name, env.Name)

				rel.Problems = append(rel.Problems, v1alpha1.Problem{
					Type:         api.ProblemTypeVirtualEnvSnapshotFailed,
					ObservedTime: api.UncomparableTime(now),
					Message:      msg,
					Causes: []v1alpha1.ProblemSource{
						{
							Kind:               api.ProblemSourceKindVirtualEnvSnapshot,
							Name:               snap.Name,
							ObservedGeneration: snap.Generation,
							Path:               "$.spec.source.name",
							Value:              &snap.Spec.Source.Name,
						},
					},
				})
			}
			// TODO validate env

		case k8s.IsNotFound(err):
			reason, msg = api.ConditionReasonVirtualEnvSnapshotFailed,
				fmt.Sprintf(`VirtualEnvSnapshot "%s" not found.`, rel.VirtualEnvSnapshot)

			rel.Problems = append(rel.Problems, v1alpha1.Problem{
				Type:         api.ProblemTypeVirtualEnvSnapshotFailed,
				ObservedTime: api.UncomparableTime(now),
				Message:      msg,
				Causes: []v1alpha1.ProblemSource{
					{
						Kind: api.ProblemSourceKindVirtualEnvSnapshot,
						Name: rel.VirtualEnvSnapshot,
					},
				},
			})

		case k8s.IgnoreNotFound(err) != nil:
			return
		}
	}

	policy := env.Spec.ReleasePolicy
	if policy.AppDeploymentPolicy == "" {
		policy.AppDeploymentPolicy = api.AppDeploymentPolicyVersionRequired
	}
	if policy.VirtualEnvPolicy == "" {
		policy.VirtualEnvPolicy = api.VirtualEnvPolicySnapshotRequired
	}

	if policy.VirtualEnvPolicy == api.VirtualEnvPolicySnapshotRequired && rel.VirtualEnvSnapshot == "" {
		reason, msg = api.ConditionReasonPolicyViolation, "VirtualEnvSnapshot is required but not set."
		value := string(policy.VirtualEnvPolicy)
		rel.Problems = append(rel.Problems, v1alpha1.Problem{
			Type:         api.ProblemTypePolicyViolation,
			ObservedTime: api.UncomparableTime(now),
			Message:      msg,
			Causes: []v1alpha1.ProblemSource{
				{
					Kind:               api.ProblemSourceKindVirtualEnv,
					Name:               env.Name,
					ObservedGeneration: env.Generation,
					Path:               "$.spec.releasePolicy.virtualEnvPolicy",
					Value:              &value,
				},
				{
					Kind:               api.ProblemSourceKindRelease,
					ObservedGeneration: env.Generation,
					Path:               "$.virtualEnvSnapshot",
					Value:              &rel.VirtualEnvSnapshot,
				},
			},
		})
	}
	if policy.AppDeploymentPolicy == api.AppDeploymentPolicyVersionRequired && rel.AppDeployment.Version == "" {
		reason, msg = api.ConditionReasonPolicyViolation, "AppDeployment version is required but not set."
		value := string(policy.AppDeploymentPolicy)
		rel.Problems = append(rel.Problems, v1alpha1.Problem{
			Type:         api.ProblemTypePolicyViolation,
			ObservedTime: api.UncomparableTime(now),
			Message:      msg,
			Causes: []v1alpha1.ProblemSource{
				{
					Kind:               api.ProblemSourceKindVirtualEnv,
					Name:               env.Name,
					ObservedGeneration: env.Generation,
					Path:               "$.spec.releasePolicy.appDeploymentPolicy",
					Value:              &value,
				},
				{
					Kind:               api.ProblemSourceKindRelease,
					ObservedGeneration: env.Generation,
					Path:               "$.appDeployment.version",
					Value:              &rel.AppDeployment.Version,
				},
			},
		})
	}

	// Clear error if it is NotFound.
	err = k8s.IgnoreNotFound(err)

	return
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
