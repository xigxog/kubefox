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

type VirtualEnvReconciler struct {
	*Client

	log *logkf.Logger
}

// SetupWithManager sets up the controller with the Manager.
func (r *VirtualEnvReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.log = logkf.Global.With(logkf.KeyController, "VirtualEnvironment")
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.VirtualEnvironment{}).
		Watches(
			&v1alpha1.AppDeployment{},
			handler.EnqueueRequestsFromMapFunc(r.watchAppDeployment),
		).
		Watches(
			&v1alpha1.Environment{},
			handler.EnqueueRequestsFromMapFunc(r.watchEnvironment),
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
	log.Debugf("reconciling VirtualEnvironment '%s/%s'", req.Namespace, req.Name)
	defer log.Debugf("reconciling VirtualEnvironment '%s/%s' done", req.Namespace, req.Name)

	ve := &v1alpha1.VirtualEnvironment{}
	if err := r.Get(ctx, req.NamespacedName, ve); err != nil {
		return ctrl.Result{}, k8s.IgnoreNotFound(err)
	}
	origVE := ve.DeepCopy()

	requeueAfter, err := r.reconcile(ctx, ve, now, log)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Sort and prune Release history based on limits.
	{
		// Most recently archived Release should be first.
		sort.SliceStable(ve.Status.ReleaseHistory, func(i, j int) bool {
			lhs := ve.Status.ReleaseHistory[i]
			if lhs.ArchiveTime == nil {
				lhs.ArchiveTime = &now
			}
			rhs := ve.Status.ReleaseHistory[j]
			if rhs.ArchiveTime == nil {
				rhs.ArchiveTime = &now
			}

			return lhs.ArchiveTime.After(ve.Status.ReleaseHistory[j].ArchiveTime.Time)
		})

		countLimit := api.DefaultReleaseHistoryCountLimit
		if ve.Spec.ReleasePolicy != nil &&
			ve.Spec.ReleasePolicy.HistoryLimits != nil &&
			ve.Spec.ReleasePolicy.HistoryLimits.Count != nil {

			countLimit = int(*ve.Spec.ReleasePolicy.HistoryLimits.Count)
		}
		ageLimit := api.DefaultReleaseHistoryAgeLimit
		if ve.Spec.ReleasePolicy != nil &&
			ve.Spec.ReleasePolicy.HistoryLimits != nil &&
			ve.Spec.ReleasePolicy.HistoryLimits.AgeDays != nil {

			ageLimit = int(*ve.Spec.ReleasePolicy.HistoryLimits.AgeDays)
		}

		// Prune history if count limit is exceeded.
		if len(ve.Status.ReleaseHistory) > countLimit {
			ve.Status.ReleaseHistory = ve.Status.ReleaseHistory[:countLimit]
		}
		// Prune history if age limit is exceeded.
		if ageLimit > 0 {
			for i, s := range ve.Status.ReleaseHistory {
				if time.Since(s.ArchiveTime.Time) > (time.Hour * 24 * time.Duration(ageLimit)) {
					ve.Status.ReleaseHistory = ve.Status.ReleaseHistory[:i+1]
					break
				}
			}
		}
	}

	// Update status if changed.
	if !k8s.DeepEqual(&ve.Status, &origVE.Status) {
		log.Debug("VirtualEnvironment status modified, updating")
		if err := r.MergeStatus(ctx, ve, origVE); k8s.IgnoreNotFound(err) != nil {
			return ctrl.Result{}, err
		}
	}

	if requeueAfter > 0 {
		log.Debugf("Release pending, requeuing after %s to check deadline", requeueAfter)
	}

	return ctrl.Result{RequeueAfter: requeueAfter}, nil
}

func (r *VirtualEnvReconciler) reconcile(ctx context.Context, ve *v1alpha1.VirtualEnvironment, now metav1.Time, log *logkf.Logger) (time.Duration, error) {
	origVE := ve.DeepCopy()
	pending := k8s.Condition(ve.Status.Conditions, api.ConditionTypeReleasePending)
	if pending.Status == metav1.ConditionFalse &&
		pending.Reason == api.ConditionReasonPendingDeadlineExceeded &&
		ve.Status.PendingReleaseFailed && !releaseEqual(ve.Status.ActiveRelease, ve.Spec.Release) {

		// The pending Release failed to activate before deadline was exceeded.
		// Rollback to active Release.
		if active := ve.Status.ActiveRelease; active != nil {
			ve.Spec.Release = &v1alpha1.Release{
				AppDeployments: map[string]v1alpha1.ReleaseAppDeployment{},
			}
			for appDepName, appDep := range active.AppDeployments {
				ve.Spec.Release.AppDeployments[appDepName] = appDep.ReleaseAppDeployment
			}

		} else {
			ve.Spec.Release = nil
		}

		err := r.Merge(ctx, ve, origVE)
		if k8s.IgnoreNotFound(err) == nil {
			return 0, err
		}
		if IsFailedWebhookErr(err) {
			log.Debug("reconcile failed because of webhook, retrying in 15 seconds")
			return time.Second * 15, nil
		}

		return 0, err
	}
	// Reset flag to default so rollback does not occur again.
	ve.Status.PendingReleaseFailed = false

	envFound := true
	env := &v1alpha1.Environment{}
	if err := r.Get(ctx, k8s.Key("", ve.Spec.Environment), env); k8s.IgnoreNotFound(err) != nil {
		return 0, err
	} else if k8s.IsNotFound(err) {
		envFound = false
	}

	if ve.Spec.Release == nil {
		pendingReason, pendingMsg := api.ConditionReasonNoRelease, "No Release set for VirtualEnvironment."
		if pending.Reason == api.ConditionReasonPendingDeadlineExceeded {
			// Retain reason and message if the deadline was exceeded previously.
			pendingReason = pending.Reason
			pendingMsg = pending.Message
		}

		ve.Status.Conditions = k8s.UpdateConditions(now, ve.Status.Conditions, &metav1.Condition{
			Type:               api.ConditionTypeActiveReleaseAvailable,
			Status:             metav1.ConditionFalse,
			ObservedGeneration: ve.ObjectMeta.Generation,
			Reason:             api.ConditionReasonNoRelease,
			Message:            "No Release set for VirtualEnvironment.",
		}, &metav1.Condition{
			Type:               api.ConditionTypeReleasePending,
			Status:             metav1.ConditionFalse,
			ObservedGeneration: ve.ObjectMeta.Generation,
			Reason:             pendingReason,
			Message:            pendingMsg,
		})
		if ve.Status.ActiveRelease != nil {
			ve.Status.ActiveRelease.ArchiveTime = &now
			ve.Status.ActiveRelease.ArchiveReason = api.ArchiveReasonSuperseded
			ve.Status.ReleaseHistory = append([]v1alpha1.ReleaseStatus{*ve.Status.ActiveRelease}, ve.Status.ReleaseHistory...)
			ve.Status.ActiveRelease = nil
		}
		if ve.Status.PendingRelease != nil {
			ve.Status.PendingRelease.ArchiveTime = &now
			ve.Status.PendingRelease.ArchiveReason = api.ArchiveReasonSuperseded
			ve.Status.ReleaseHistory = append([]v1alpha1.ReleaseStatus{*ve.Status.PendingRelease}, ve.Status.ReleaseHistory...)
			ve.Status.PendingRelease = nil
		}

		return 0, nil
	}

	var remainingDeadline time.Duration
	isActive := releaseEqual(ve.Status.ActiveRelease, ve.Spec.Release)

	if isActive {
		ve.Status.PendingRelease = nil

	} else if !releaseEqual(ve.Status.PendingRelease, ve.Spec.Release) {
		ve.Status.PendingRelease = &v1alpha1.ReleaseStatus{
			RequestTime:    now,
			AppDeployments: map[string]v1alpha1.ReleaseAppDeploymentStatus{},
		}
		for appDepName, appDep := range ve.Spec.Release.AppDeployments {
			ve.Status.PendingRelease.AppDeployments[appDepName] = v1alpha1.ReleaseAppDeploymentStatus{
				ReleaseAppDeployment: appDep,
			}
		}
	}

	var (
		pendingStatus = metav1.ConditionFalse
		activeStatus  = metav1.ConditionFalse
	)
	{
		// Update ReleasePending condition.
		var (
			reason = api.ConditionReasonReleaseActivated
			msg    = "Release was activated."
			err    error
		)
		if pending.Reason == api.ConditionReasonPendingDeadlineExceeded {
			// Retain reason and message if the deadline was exceeded previously.
			reason = pending.Reason
			msg = pending.Message
		}
		if ve.Status.PendingRelease != nil {
			pendingStatus = metav1.ConditionTrue
			remainingDeadline = ve.GetReleasePolicy(env).GetPendingDeadline() - ve.GetReleasePendingDuration()
			if err = r.updateProblems(ctx, now, env, ve, ve.Status.PendingRelease); err != nil {
				return 0, err
			}

			switch {
			// Do not set ConditionFalse if Environment not found.
			case len(ve.Status.PendingRelease.Problems) == 0 && envFound:
				pendingStatus = metav1.ConditionFalse
				reason = api.ConditionReasonReleaseActivated
				remainingDeadline = 0

				// Set pendingRelease as activeRelease.
				if ve.Status.ActiveRelease != nil {
					ve.Status.ActiveRelease.ArchiveTime = &now
					ve.Status.ActiveRelease.ArchiveReason = api.ArchiveReasonSuperseded
					ve.Status.ReleaseHistory = append([]v1alpha1.ReleaseStatus{*ve.Status.ActiveRelease}, ve.Status.ReleaseHistory...)
				}
				ve.Status.ActiveRelease = ve.Status.PendingRelease
				ve.Status.ActiveRelease.ActivationTime = &now
				ve.Status.PendingRelease = nil

			case remainingDeadline <= 0:
				pendingStatus = metav1.ConditionFalse
				reason = api.ConditionReasonPendingDeadlineExceeded
				msg = "Deadline exceeded before pending Release could be activated."

				ve.Status.PendingRelease.ArchiveTime = &now
				ve.Status.PendingRelease.ArchiveReason = api.ArchiveReasonPendingDeadlineExceeded
				ve.Status.ReleaseHistory = append([]v1alpha1.ReleaseStatus{*ve.Status.PendingRelease}, ve.Status.ReleaseHistory...)
				ve.Status.PendingRelease = nil
				ve.Status.PendingReleaseFailed = true

			case len(ve.Status.PendingRelease.Problems) > 0:
				pendingStatus = metav1.ConditionTrue
				reason = api.ConditionReasonProblemsFound
				msg = "One or more problems exist with Release preventing it from being activated, see `status.pendingRelease` for details."
			}
		}
		ve.Status.Conditions = k8s.UpdateConditions(now, ve.Status.Conditions, &metav1.Condition{
			Type:               api.ConditionTypeReleasePending,
			Status:             pendingStatus,
			ObservedGeneration: ve.ObjectMeta.Generation,
			Reason:             reason,
			Message:            msg,
		})
	}

	{
		// Update ActiveReleaseAvailable condition.
		var (
			reason = api.ConditionReasonReleasePending
			msg    = "No active Release, Release is pending activation."
			err    error
		)
		if ve.Status.ActiveRelease != nil {
			if err = r.updateProblems(ctx, now, env, ve, ve.Status.ActiveRelease); err != nil {
				return 0, err
			}

			// Do not set ConditionTrue if Environment not found.
			if len(ve.Status.ActiveRelease.Problems) == 0 && envFound {
				activeStatus = metav1.ConditionTrue
				reason = api.ConditionReasonContextAvailable
				msg = "Release AppDeployment is available, Routes and Adapters are valid and compatible with the VirtualEnv."
			} else {
				activeStatus = metav1.ConditionFalse
				reason = api.ConditionReasonProblemsFound
				msg = "One or more problems exist with the active Release causing it to be unavailable, see `status.activeRelease` for details."
			}
		}
		ve.Status.Conditions = k8s.UpdateConditions(now, ve.Status.Conditions, &metav1.Condition{
			Type:               api.ConditionTypeActiveReleaseAvailable,
			Status:             activeStatus,
			ObservedGeneration: ve.ObjectMeta.Generation,
			Reason:             reason,
			Message:            msg,
		})
	}

	if !envFound {
		msg := fmt.Sprintf("Environment '%s' not found.", ve.Spec.Environment)

		if pendingStatus == metav1.ConditionTrue {
			ve.Status.Conditions = k8s.UpdateConditions(now, ve.Status.Conditions, &metav1.Condition{
				Type:               api.ConditionTypeReleasePending,
				Status:             metav1.ConditionTrue,
				ObservedGeneration: ve.ObjectMeta.Generation,
				Reason:             api.ConditionReasonEnvironmentNotFound,
				Message:            msg,
			})
		}
		if ve.Status.PendingRelease != nil {
			ve.Status.PendingRelease.Problems = nil
		}

		if activeStatus == metav1.ConditionFalse {
			ve.Status.Conditions = k8s.UpdateConditions(now, ve.Status.Conditions, &metav1.Condition{
				Type:               api.ConditionTypeActiveReleaseAvailable,
				Status:             metav1.ConditionFalse,
				ObservedGeneration: ve.ObjectMeta.Generation,
				Reason:             api.ConditionReasonEnvironmentNotFound,
				Message:            msg,
			})
		}
		if ve.Status.ActiveRelease != nil {
			ve.Status.ActiveRelease.Problems = nil
		}
	}

	return remainingDeadline, nil
}

func (r *VirtualEnvReconciler) updateProblems(ctx context.Context, now metav1.Time,
	env *v1alpha1.Environment, ve *v1alpha1.VirtualEnvironment, rel *v1alpha1.ReleaseStatus) error {

	rel.Problems = nil
	data := ve.Data.MergeInto(&env.Data)
	policy := ve.GetReleasePolicy(env)

	for relAppDepName, relAppDep := range rel.AppDeployments {
		relAppDep.ObservedGeneration = 0

		appDep := &v1alpha1.AppDeployment{}
		err := r.Get(ctx, k8s.Key(ve.Namespace, relAppDepName), appDep)

		switch {
		case err == nil:
			relAppDep.ObservedGeneration = appDep.Generation
			progressing := k8s.Condition(appDep.Status.Conditions, api.ConditionTypeProgressing)
			available := k8s.Condition(appDep.Status.Conditions, api.ConditionTypeAvailable)

			appDepProblems, err := appDep.Validate(data, func(name string, typ api.ComponentType) (api.Adapter, error) {
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
				available.Reason != api.ConditionReasonProblemsFound {

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
				progressing.Reason != api.ConditionReasonProblemsFound {

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

			if *policy.VersionRequired && relAppDep.Version == "" {
				msg := fmt.Sprintf(`Version is required but not set for AppDeployment "%s".`, relAppDepName)
				value := fmt.Sprint(*policy.VersionRequired)
				rel.Problems = append(rel.Problems, common.Problem{
					ObservedTime: now,
					Problem: api.Problem{
						Type:    api.ProblemTypePolicyViolation,
						Message: msg,
						Causes: []api.ProblemSource{
							{
								Kind:               api.ProblemSourceKindVirtualEnvironment,
								Name:               ve.Name,
								ObservedGeneration: ve.Generation,
								Path:               "$.spec.releasePolicy.versionRequired",
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

			if relAppDep.Version != "" && relAppDep.Version != appDep.Spec.Version {
				msg := fmt.Sprintf(`AppDeployment "%s" version "%s" does not match Release version "%s".`,
					appDep.Name, appDep.Spec.Version, relAppDep.Version)
				rel.Problems = append(rel.Problems, common.Problem{
					ObservedTime: now,
					Problem: api.Problem{
						Type:    api.ProblemTypeAppDeploymentFailed,
						Message: msg,
						Causes: []api.ProblemSource{
							{
								Kind:               api.ProblemSourceKindVirtualEnvironment,
								ObservedGeneration: ve.Generation,
								Path:               "$.spec.release.appDeployment.version",
								Value:              &relAppDep.Version,
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
			msg := fmt.Sprintf(`AppDeployment "%s" not found.`, relAppDepName)
			rel.Problems = append(rel.Problems, common.Problem{
				ObservedTime: now,
				Problem: api.Problem{
					Type:    api.ProblemTypeAppDeploymentNotFound,
					Message: msg,
					Causes: []api.ProblemSource{
						{
							Kind: api.ProblemSourceKindAppDeployment,
							Name: relAppDepName,
						},
					},
				},
			})

		case k8s.IgnoreNotFound(err) != nil:
			return err
		}
	}

	return nil
}

// TODO
func (r *VirtualEnvReconciler) watchAppDeployment(ctx context.Context, appDep client.Object) []reconcile.Request {
	return r.findEnvs(ctx,
		client.InNamespace(appDep.GetNamespace()),
		client.MatchingLabels{
			api.LabelK8sAppDeployment: appDep.GetName(),
		},
	)
}

func (r *VirtualEnvReconciler) watchEnvironment(ctx context.Context, env client.Object) []reconcile.Request {
	return r.findEnvs(ctx,
		client.MatchingLabels{
			api.LabelK8sEnvironment: env.GetName(),
		},
	)
}

func (r *VirtualEnvReconciler) findEnvs(ctx context.Context, opts ...client.ListOption) []reconcile.Request {
	veList := &v1alpha1.VirtualEnvironmentList{}
	if err := r.List(ctx, veList, opts...); err != nil {
		r.log.Error(err)
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(veList.Items))
	for i, rel := range veList.Items {
		requests[i] = reconcile.Request{
			NamespacedName: k8s.Key(rel.Namespace, rel.Name),
		}
	}

	return requests
}

func releaseEqual(lhs *v1alpha1.ReleaseStatus, rhs *v1alpha1.Release) bool {
	switch {
	case lhs == nil && rhs == nil:
		return true
	case lhs != nil && rhs == nil:
		return false
	case lhs == nil && rhs != nil:
		return false
	case len(lhs.AppDeployments) != len(rhs.AppDeployments):
		return false
	}

	for _, lhsAppDep := range lhs.AppDeployments {
		found := false
		for _, rhsAppDep := range rhs.AppDeployments {
			if lhsAppDep.ReleaseAppDeployment == rhsAppDep {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}
