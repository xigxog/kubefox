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
	"sort"
	"time"

	"github.com/xigxog/kubefox/api"
	common "github.com/xigxog/kubefox/api/kubernetes"
	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/core"
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
	now := metav1.Now()
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

	requeueAfter, err := r.reconcile(ctx, env, now, log)
	if IgnoreFailedWebhookErr(err) != nil {
		return ctrl.Result{}, err

	} else if IsFailedWebhookErr(err) {
		log.Debug("reconcile failed because of webhook, retrying in 15 seconds")
		return ctrl.Result{RequeueAfter: time.Second * 15}, nil
	}

	// Sort and prune Release history based on limits.
	{
		// Most recently archived Release should be first.
		sort.SliceStable(env.Status.ReleaseHistory, func(i, j int) bool {
			lhs := env.Status.ReleaseHistory[i]
			if lhs.ArchiveTime == nil {
				lhs.ArchiveTime = &now
			}
			rhs := env.Status.ReleaseHistory[j]
			if rhs.ArchiveTime == nil {
				rhs.ArchiveTime = &now
			}

			return lhs.ArchiveTime.After(env.Status.ReleaseHistory[j].ArchiveTime.Time)
		})

		// Prune history if count limit is exceeded.
		count := int(env.Spec.ReleasePolicy.HistoryLimits.Count)
		if count <= 0 {
			count = api.DefaultReleaseHistoryLimitCount
		}
		if len(env.Status.ReleaseHistory) > count {
			env.Status.ReleaseHistory = env.Status.ReleaseHistory[:count]
		}

		// Prune history if age limit is exceeded.
		if ageLimit := env.Spec.ReleasePolicy.HistoryLimits.AgeDays; ageLimit > 0 {
			for i, s := range env.Status.ReleaseHistory {
				if time.Since(s.ArchiveTime.Time) > (time.Hour * 24 * time.Duration(ageLimit)) {
					env.Status.ReleaseHistory = env.Status.ReleaseHistory[:i+1]
					break
				}
			}
		}
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

func (r *VirtualEnvReconciler) reconcile(ctx context.Context, env *v1alpha1.VirtualEnv, now metav1.Time, log *logkf.Logger) (time.Duration, error) {
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
			Type:               api.ConditionTypeActiveReleaseAvailable,
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
			if err = r.updateProblems(ctx, now, env, env.Status.PendingRelease); err != nil {
				return 0, err
			}

			switch {
			case len(env.Status.PendingRelease.Problems) == 0:
				status = metav1.ConditionFalse
				reason = api.ConditionReasonReleaseActive
				remainingDeadline = 0

				// Set pendingRelease as activeRelease.
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

			case len(env.Status.PendingRelease.Problems) > 0:
				status = metav1.ConditionTrue
				reason = api.ConditionReasonProblemsExist
				msg = "One or more problems exist with Release preventing it from being activated, see `status.pendingRelease` for details."
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
		// Update ActiveReleaseAvailable condition.
		var (
			status = metav1.ConditionFalse
			reason = api.ConditionReasonReleasePending
			msg    = "No active Release, Release is pending activation."
			err    error
		)
		if env.Status.ActiveRelease != nil {
			if err = r.updateProblems(ctx, now, env, env.Status.ActiveRelease); err != nil {
				return 0, err
			}
			if len(env.Status.ActiveRelease.Problems) == 0 {
				status = metav1.ConditionTrue
				reason = api.ConditionReasonAppDeploymentAvailable
				msg = "Release AppDeployment is available, Routes and Adapters are valid and compatible with the VirtualEnv."
			} else {
				status = metav1.ConditionFalse
				reason = api.ConditionReasonProblemsExist
				msg = "One or more problems exist with the active Release causing it to be unavailable, see `status.activeRelease` for details."
			}
		}
		env.Status.Conditions = k8s.UpdateConditions(now, env.Status.Conditions, &metav1.Condition{
			Type:               api.ConditionTypeActiveReleaseAvailable,
			Status:             status,
			ObservedGeneration: env.ObjectMeta.Generation,
			Reason:             reason,
			Message:            msg,
		})
	}

	return remainingDeadline, nil
}

func (r *VirtualEnvReconciler) updateProblems(ctx context.Context, now metav1.Time,
	env *v1alpha1.VirtualEnv, rel *v1alpha1.ReleaseStatus) error {

	rel.AppDeployment.ObservedGeneration = 0
	rel.Problems = nil

	data := env.Data
	if rel.VirtualEnvSnapshot != "" {
		snap := &v1alpha1.VirtualEnvSnapshot{}
		err := r.Get(ctx, k8s.Key(env.Namespace, rel.VirtualEnvSnapshot), snap)
		switch {
		case err == nil:
			data = *snap.Data

			if snap.Spec.Source.Name != env.Name {
				msg := fmt.Sprintf(`VirtualEnvSnapshot "%s" source "%s" is not VirtualEnv "%s".`,
					snap.Name, snap.Spec.Source.Name, env.Name)
				rel.Problems = append(rel.Problems, common.Problem{
					ObservedTime: now,
					Problem: api.Problem{
						Type:    api.ProblemTypeVirtualEnvSnapshotFailed,
						Message: msg,
						Causes: []api.ProblemSource{
							{
								Kind:               api.ProblemSourceKindVirtualEnvSnapshot,
								Name:               snap.Name,
								ObservedGeneration: snap.Generation,
								Path:               "$.spec.source.name",
								Value:              &snap.Spec.Source.Name,
							},
						},
					},
				})
			}

		case k8s.IsNotFound(err):
			err = nil
			msg := fmt.Sprintf(`VirtualEnvSnapshot "%s" not found.`, rel.VirtualEnvSnapshot)
			rel.Problems = append(rel.Problems, common.Problem{
				ObservedTime: now,
				Problem: api.Problem{
					Type:    api.ProblemTypeVirtualEnvSnapshotFailed,
					Message: msg,
					Causes: []api.ProblemSource{
						{
							Kind: api.ProblemSourceKindVirtualEnvSnapshot,
							Name: rel.VirtualEnvSnapshot,
						},
					},
				},
			})

		case k8s.IgnoreNotFound(err) != nil:
			return err
		}
	}

	appDep := &v1alpha1.AppDeployment{}
	err := r.Get(ctx, k8s.Key(env.Namespace, rel.AppDeployment.Name), appDep)
	switch {
	case err == nil:
		rel.AppDeployment.ObservedGeneration = appDep.Generation
		progressing := k8s.Condition(appDep.Status.Conditions, api.ConditionTypeProgressing)
		available := k8s.Condition(appDep.Status.Conditions, api.ConditionTypeAvailable)

		appDepProblems, err := appDep.Validate(&data, func(name string, typ api.ComponentType) (api.Adapter, error) {
			switch typ {
			case api.ComponentTypeHTTPAdapter:
				a := &v1alpha1.HTTPAdapter{}
				if err := r.Get(ctx, k8s.Key(appDep.Namespace, name), a); err != nil {
					return nil, err
				}
				return a, nil

			default:
				return nil, core.ErrNotFound()
			}
		})
		if err != nil {
			return err
		}
		for _, p := range appDepProblems {
			rel.Problems = append(rel.Problems, common.Problem{
				ObservedTime: now,
				Problem:      p,
			})
		}

		if available.Status == metav1.ConditionFalse &&
			available.Reason != api.ConditionReasonProblemsExist {

			msg := fmt.Sprintf(`One or more Component Deployments of AppDeployment "%s" are unavailable.`, appDep.Name)
			value := fmt.Sprintf("%s,%s", available.Status, available.Reason)
			rel.Problems = append(rel.Problems, common.Problem{
				ObservedTime: now,
				Problem: api.Problem{
					Type:    api.ProblemTypeAppDeploymentFailed,
					Message: msg,
					Causes: []api.ProblemSource{
						{
							Kind:               api.ProblemSourceKindAppDeployment,
							Name:               appDep.Name,
							ObservedGeneration: appDep.Generation,
							Path:               "$.status.conditions[?(@.type=='Available')].status,reason",
							Value:              &value,
						},
					},
				},
			})
		}

		if progressing.Status == metav1.ConditionFalse &&
			progressing.Reason != api.ConditionReasonComponentsDeployed &&
			progressing.Reason != api.ConditionReasonProblemsExist {

			msg := fmt.Sprintf(`One or more Component Deployments of AppDeployment "%s" failed.`, appDep.Name)
			value := fmt.Sprintf("%s,%s", progressing.Status, progressing.Reason)
			rel.Problems = append(rel.Problems, common.Problem{
				ObservedTime: now,
				Problem: api.Problem{
					Type:    api.ProblemTypeAppDeploymentFailed,
					Message: msg,
					Causes: []api.ProblemSource{
						{
							Kind:               api.ProblemSourceKindAppDeployment,
							Name:               appDep.Name,
							ObservedGeneration: appDep.Generation,
							Path:               "$.status.conditions[?(@.type=='Progressing')].status,reason",
							Value:              &value,
						},
					},
				},
			})
		}

		if rel.AppDeployment.Version != "" && rel.AppDeployment.Version != appDep.Spec.Version {
			msg := fmt.Sprintf(`AppDeployment "%s" version "%s" does not match Release version "%s".`,
				appDep.Name, appDep.Spec.Version, rel.AppDeployment.Version)
			rel.Problems = append(rel.Problems, common.Problem{
				ObservedTime: now,
				Problem: api.Problem{
					Type:    api.ProblemTypeAppDeploymentFailed,
					Message: msg,
					Causes: []api.ProblemSource{
						{
							Kind:               api.ProblemSourceKindVirtualEnv,
							ObservedGeneration: env.Generation,
							Path:               "$.spec.release.appDeployment.version",
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
				},
			})
		}

	case k8s.IsNotFound(err):
		err = nil
		msg := fmt.Sprintf(`AppDeployment "%s" not found.`, rel.AppDeployment.Name)
		rel.Problems = append(rel.Problems, common.Problem{
			ObservedTime: now,
			Problem: api.Problem{
				Type:    api.ProblemTypeAppDeploymentFailed,
				Message: msg,
				Causes: []api.ProblemSource{
					{
						Kind: api.ProblemSourceKindAppDeployment,
						Name: rel.AppDeployment.Name,
					},
				},
			},
		})

	case k8s.IgnoreNotFound(err) != nil:
		return err
	}

	policy := env.Spec.ReleasePolicy
	if policy.AppDeploymentPolicy == "" {
		policy.AppDeploymentPolicy = api.AppDeploymentPolicyVersionRequired
	}
	if policy.VirtualEnvPolicy == "" {
		policy.VirtualEnvPolicy = api.VirtualEnvPolicySnapshotRequired
	}

	if policy.VirtualEnvPolicy == api.VirtualEnvPolicySnapshotRequired && rel.VirtualEnvSnapshot == "" {
		msg := "VirtualEnvSnapshot is required but not set for Release."
		value := string(policy.VirtualEnvPolicy)
		rel.Problems = append(rel.Problems, common.Problem{
			ObservedTime: now,
			Problem: api.Problem{
				Type:    api.ProblemTypePolicyViolation,
				Message: msg,
				Causes: []api.ProblemSource{
					{
						Kind:               api.ProblemSourceKindVirtualEnv,
						Name:               env.Name,
						ObservedGeneration: env.Generation,
						Path:               "$.spec.releasePolicy.virtualEnvPolicy",
						Value:              &value,
					},
					{
						Kind:               api.ProblemSourceKindVirtualEnv,
						ObservedGeneration: env.Generation,
						Path:               "$.spec.release.virtualEnvSnapshot",
						Value:              &rel.VirtualEnvSnapshot,
					},
				},
			},
		})
	}
	if policy.AppDeploymentPolicy == api.AppDeploymentPolicyVersionRequired && rel.AppDeployment.Version == "" {
		msg := "AppDeployment version is required but not set for Release."
		value := string(policy.AppDeploymentPolicy)
		rel.Problems = append(rel.Problems, common.Problem{
			ObservedTime: now,
			Problem: api.Problem{
				Type:    api.ProblemTypePolicyViolation,
				Message: msg,
				Causes: []api.ProblemSource{
					{
						Kind:               api.ProblemSourceKindVirtualEnv,
						Name:               env.Name,
						ObservedGeneration: env.Generation,
						Path:               "$.spec.releasePolicy.appDeploymentPolicy",
						Value:              &value,
					},
					{
						Kind:               api.ProblemSourceKindAppDeployment,
						ObservedGeneration: appDep.Generation,
						Path:               "$.spec.version",
						Value:              &appDep.Spec.Version,
					},
				},
			},
		})
	}

	return nil
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
