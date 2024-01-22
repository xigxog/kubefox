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

	CompMgr *ComponentManager

	log *logkf.Logger
}

type VirtualEnvContext struct {
	context.Context
	*v1alpha1.VirtualEnvironment

	Original *v1alpha1.VirtualEnvironment

	Data   *api.Data
	Policy *v1alpha1.ReleasePolicy

	EnvFound bool

	Now metav1.Time
}

// SetupWithManager sets up the controller with the Manager.
func (r *VirtualEnvReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.log = logkf.Global.With(logkf.KeyController, "VirtualEnvironment")
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.VirtualEnvironment{}).
		Watches(
			&v1alpha1.AppDeployment{},
			handler.EnqueueRequestsFromMapFunc(r.watchAppDeployments),
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

	env := &v1alpha1.Environment{}
	if err := r.Get(ctx, k8s.Key("", ve.Spec.Environment), env); k8s.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, err
	}

	veCtx := &VirtualEnvContext{
		Context:            ctx,
		VirtualEnvironment: ve,
		Original:           ve.DeepCopy(),
		Data:               ve.Data.MergeInto(&env.Data),
		Policy:             ve.GetReleasePolicy(env),
		EnvFound:           env.UID != "",
		Now:                metav1.Now(),
	}

	if ve.DeletionTimestamp == nil {
		if err := r.reconcile(veCtx, log); err != nil {
			if IsFailedWebhookErr(err) {
				log.Debug("reconcile failed because of webhook, retrying in 15 seconds")
				return reconcile.Result{RequeueAfter: time.Second * 15}, nil
			}

			return ctrl.Result{}, err
		}

	} else {
		// VE was deleted, clear status to force cleanup.
		ve.Status = v1alpha1.VirtualEnvironmentStatus{}
	}

	// Sort and prune Release history based on limits. Most recently archived
	// Release should be first.
	sort.SliceStable(ve.Status.ReleaseHistory, func(i, j int) bool {
		lhs := ve.Status.ReleaseHistory[i]
		if lhs.ArchiveTime == nil {
			lhs.ArchiveTime = &veCtx.Now
		}
		rhs := ve.Status.ReleaseHistory[j]
		if rhs.ArchiveTime == nil {
			rhs.ArchiveTime = &veCtx.Now
		}

		return lhs.ArchiveTime.After(ve.Status.ReleaseHistory[j].ArchiveTime.Time)
	})

	countLimit := api.DefaultReleaseHistoryCountLimit
	if *veCtx.Policy.HistoryLimits.Count != 0 {
		countLimit = int(*veCtx.Policy.HistoryLimits.Count)
	}
	ageLimit := api.DefaultReleaseHistoryAgeLimit
	if *veCtx.Policy.HistoryLimits.AgeDays != 0 {
		ageLimit = int(*veCtx.Policy.HistoryLimits.AgeDays)
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

	// Update status if changed.
	if !k8s.DeepEqual(&ve.Status, &veCtx.Original.Status) {
		log.Debug("VirtualEnvironment status modified, updating")

		err := r.MergeStatus(ctx, ve, veCtx.Original)
		if IsFailedWebhookErr(err) {
			log.Debug("reconcile failed because of webhook, retrying in 15 seconds")
			return reconcile.Result{RequeueAfter: time.Second * 15}, nil
		}
		if k8s.IgnoreNotFound(err) != nil {
			return ctrl.Result{}, err
		}
	}

	// Delete any unused ReleaseManifests.
	err := r.cleanupManifests(ctx, ve)
	if IsFailedWebhookErr(err) {
		log.Debug("reconcile failed because of webhook, retrying in 15 seconds")
		return reconcile.Result{RequeueAfter: time.Second * 15}, nil
	}
	if err != nil {
		return ctrl.Result{}, err
	}

	if ve.Status.PendingRelease != nil {
		deadline := veCtx.Policy.GetPendingDeadline() - ve.GetReleasePendingDuration()
		log.Debugf("Release pending, requeuing after %s to check deadline", deadline)
		return ctrl.Result{RequeueAfter: deadline}, nil
	}

	return ctrl.Result{}, nil
}

func (r *VirtualEnvReconciler) reconcile(ctx *VirtualEnvContext, log *logkf.Logger) error {
	pending := k8s.Condition(ctx.Status.Conditions, api.ConditionTypeReleasePending)

	// Check if the pending Release failed to activate before deadline was
	// exceeded. If so, rollback to active Release.
	if pending.Status == metav1.ConditionFalse &&
		pending.Reason == api.ConditionReasonPendingDeadlineExceeded &&
		ctx.Status.PendingReleaseFailed && !releaseEqual(ctx.Status.ActiveRelease, ctx.Spec.Release) {

		if active := ctx.Status.ActiveRelease; active != nil {
			manifest := &v1alpha1.ReleaseManifest{}
			if err := r.Get(ctx, k8s.Key(ctx.Namespace, active.ReleaseManifest), manifest); err != nil {
				return err
			}
			ctx.Spec.Release = &v1alpha1.Release{
				Id:   manifest.Spec.Id,
				Apps: map[string]v1alpha1.ReleaseApp{},
			}
			for appName, app := range manifest.Spec.Apps {
				ctx.Spec.Release.Apps[appName] = v1alpha1.ReleaseApp{
					AppDeployment: app.AppDeployment.Name,
					Version:       app.Version,
				}
			}

		} else {
			ctx.Spec.Release = nil
		}

		if err := r.Merge(ctx, ctx.VirtualEnvironment, ctx.Original); k8s.IgnoreNotFound(err) == nil {
			return err
		}

		return nil
	}

	// Reset flag to default so rollback does not occur again.
	ctx.Status.PendingReleaseFailed = false

	if ctx.Spec.Release == nil {
		pendingReason, pendingMsg := api.ConditionReasonNoRelease, "No Release set for VirtualEnvironment."
		if pending.Reason == api.ConditionReasonPendingDeadlineExceeded {
			// Retain reason and message if the deadline was exceeded previously.
			pendingReason = pending.Reason
			pendingMsg = pending.Message
		}

		ctx.Status.Conditions = k8s.UpdateConditions(ctx.Now, ctx.Status.Conditions, &metav1.Condition{
			Type:               api.ConditionTypeActiveReleaseAvailable,
			Status:             metav1.ConditionFalse,
			ObservedGeneration: ctx.Generation,
			Reason:             api.ConditionReasonNoRelease,
			Message:            "No Release set for VirtualEnvironment.",
		}, &metav1.Condition{
			Type:               api.ConditionTypeReleasePending,
			Status:             metav1.ConditionFalse,
			ObservedGeneration: ctx.Generation,
			Reason:             pendingReason,
			Message:            pendingMsg,
		})
		if ctx.Status.ActiveRelease != nil {
			ctx.Status.ActiveRelease.ArchiveTime = &ctx.Now
			ctx.Status.ActiveRelease.ArchiveReason = api.ArchiveReasonSuperseded
			ctx.Status.ReleaseHistory = append([]v1alpha1.ReleaseStatus{*ctx.Status.ActiveRelease}, ctx.Status.ReleaseHistory...)
			ctx.Status.ActiveRelease = nil
		}
		if ctx.Status.PendingRelease != nil {
			ctx.Status.PendingRelease.ArchiveTime = &ctx.Now
			ctx.Status.PendingRelease.ArchiveReason = api.ArchiveReasonSuperseded
			ctx.Status.ReleaseHistory = append([]v1alpha1.ReleaseStatus{*ctx.Status.PendingRelease}, ctx.Status.ReleaseHistory...)
			ctx.Status.PendingRelease = nil
		}

		return nil
	}

	if releaseEqual(ctx.Status.ActiveRelease, ctx.Spec.Release) {
		// Current release is active, clear pending.
		ctx.Status.PendingRelease = nil

	} else if !releaseEqual(ctx.Status.PendingRelease, ctx.Spec.Release) {
		// Current release updated, set it to pending.
		ctx.Status.PendingRelease = &v1alpha1.ReleaseStatus{
			Release:     *ctx.Spec.Release,
			RequestTime: ctx.Now,
		}
	}

	var (
		pendingStatus     = metav1.ConditionFalse
		activeStatus      = metav1.ConditionFalse
		remainingDeadline time.Duration
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
		if ctx.Status.PendingRelease != nil {
			pendingStatus = metav1.ConditionTrue
			remainingDeadline = ctx.Policy.GetPendingDeadline() - ctx.GetReleasePendingDuration()
			if err = r.updateProblems(ctx, ctx.Status.PendingRelease); err != nil {
				return err
			}

			switch {
			// Do not set ConditionFalse if Environment not found.
			case len(ctx.Status.PendingRelease.Problems) == 0 && ctx.EnvFound:
				pendingStatus = metav1.ConditionFalse
				reason = api.ConditionReasonReleaseActivated

				// Set pendingRelease as activeRelease.
				if ctx.Status.ActiveRelease != nil {
					ctx.Status.ActiveRelease.ArchiveTime = &ctx.Now
					ctx.Status.ActiveRelease.ArchiveReason = api.ArchiveReasonSuperseded
					ctx.Status.ReleaseHistory = append(ctx.Status.ReleaseHistory, *ctx.Status.ActiveRelease)
				}
				ctx.Status.ActiveRelease = ctx.Status.PendingRelease
				ctx.Status.ActiveRelease.ActivationTime = &ctx.Now
				ctx.Status.PendingRelease = nil

				// Create ReleaseManifest for VEs that require versions.
				if *ctx.Policy.VersionRequired {
					manifest, err := r.generateManifest(ctx)
					if err != nil {
						return err
					}
					if err := r.Create(ctx, manifest); err != nil {
						return err
					}
					ctx.Status.ActiveRelease.ReleaseManifest = manifest.Name
				}

			case remainingDeadline <= 0:
				pendingStatus = metav1.ConditionFalse
				reason = api.ConditionReasonPendingDeadlineExceeded
				msg = "Deadline exceeded before pending Release could be activated."

				ctx.Status.PendingRelease.ArchiveTime = &ctx.Now
				ctx.Status.PendingRelease.ArchiveReason = api.ArchiveReasonPendingDeadlineExceeded
				ctx.Status.ReleaseHistory = append(ctx.Status.ReleaseHistory, *ctx.Status.PendingRelease)
				ctx.Status.PendingRelease = nil
				ctx.Status.PendingReleaseFailed = true

			case len(ctx.Status.PendingRelease.Problems) > 0:
				pendingStatus = metav1.ConditionTrue
				reason = api.ConditionReasonProblemsFound
				msg = "One or more problems exist with Release preventing it from being activated, see `status.pendingRelease` for details."
			}
		}
		ctx.Status.Conditions = k8s.UpdateConditions(ctx.Now, ctx.Status.Conditions, &metav1.Condition{
			Type:               api.ConditionTypeReleasePending,
			Status:             pendingStatus,
			ObservedGeneration: ctx.Generation,
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
		if ctx.Status.ActiveRelease != nil {
			if err = r.updateProblems(ctx, ctx.Status.ActiveRelease); err != nil {
				return err
			}

			// Do not set ConditionTrue if Environment not found.
			if len(ctx.Status.ActiveRelease.Problems) == 0 && ctx.EnvFound {
				activeStatus = metav1.ConditionTrue
				reason = api.ConditionReasonContextAvailable
				msg = "Release AppDeployments are available, Routes and Adapters are valid and compatible with the VirtualEnv."
			} else {
				activeStatus = metav1.ConditionFalse
				reason = api.ConditionReasonProblemsFound
				msg = "One or more problems exist with the active Release causing it to be unavailable, see `status.activeRelease` for details."
			}
		}
		ctx.Status.Conditions = k8s.UpdateConditions(ctx.Now, ctx.Status.Conditions, &metav1.Condition{
			Type:               api.ConditionTypeActiveReleaseAvailable,
			Status:             activeStatus,
			ObservedGeneration: ctx.Generation,
			Reason:             reason,
			Message:            msg,
		})
	}

	if !ctx.EnvFound {
		msg := fmt.Sprintf("Environment '%s' not found.", ctx.Spec.Environment)

		if pendingStatus == metav1.ConditionTrue {
			ctx.Status.Conditions = k8s.UpdateConditions(ctx.Now, ctx.Status.Conditions, &metav1.Condition{
				Type:               api.ConditionTypeReleasePending,
				Status:             metav1.ConditionTrue,
				ObservedGeneration: ctx.Generation,
				Reason:             api.ConditionReasonEnvironmentNotFound,
				Message:            msg,
			})
		}
		if ctx.Status.PendingRelease != nil {
			ctx.Status.PendingRelease.Problems = nil
		}

		if activeStatus == metav1.ConditionFalse {
			ctx.Status.Conditions = k8s.UpdateConditions(ctx.Now, ctx.Status.Conditions, &metav1.Condition{
				Type:               api.ConditionTypeActiveReleaseAvailable,
				Status:             metav1.ConditionFalse,
				ObservedGeneration: ctx.Generation,
				Reason:             api.ConditionReasonEnvironmentNotFound,
				Message:            msg,
			})
		}
		if ctx.Status.ActiveRelease != nil {
			ctx.Status.ActiveRelease.Problems = nil
		}
	}

	return nil
}

func (r *VirtualEnvReconciler) generateManifest(ctx *VirtualEnvContext) (*v1alpha1.ReleaseManifest, error) {
	manifest := &v1alpha1.ReleaseManifest{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.GroupVersion.Identifier(),
			Kind:       "ReleaseManifest",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s-%s",
				ctx.Name, ctx.ResourceVersion, ctx.Now.UTC().Format("20060102-150405")),
			Namespace: ctx.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: ctx.APIVersion,
					Kind:       ctx.Kind,
					Name:       ctx.Name,
					UID:        ctx.UID,
				},
			},
			Finalizers: []string{api.FinalizerReleaseProtection},
		},
		Spec: v1alpha1.ReleaseManifestSpec{
			Id: ctx.Spec.Release.Id,
			VirtualEnvironment: v1alpha1.ReleaseManifestEnv{
				UID:             ctx.UID,
				Name:            ctx.Name,
				ResourceVersion: ctx.ResourceVersion,
				Environment:     ctx.Spec.Environment,
			},
			Apps: map[string]*v1alpha1.ReleaseManifestApp{},
		},
		Data:    *ctx.Data,
		Details: ctx.Details,
	}

	for appName, app := range ctx.Spec.Release.Apps {
		appDep := &v1alpha1.AppDeployment{}
		if err := r.Get(ctx, k8s.Key(ctx.Namespace, app.AppDeployment), appDep); err != nil {
			return nil, err
		}

		manifest.Spec.Apps[appName] = &v1alpha1.ReleaseManifestApp{
			Version: app.Version,
			AppDeployment: v1alpha1.ReleaseManifestAppDep{
				UID:             appDep.UID,
				Name:            appDep.Name,
				ResourceVersion: appDep.ResourceVersion,
				Spec:            appDep.Spec,
			},
		}
	}

	return manifest, nil
}

func (r *VirtualEnvReconciler) updateProblems(ctx *VirtualEnvContext, rel *v1alpha1.ReleaseStatus) error {
	if rel == nil {
		return nil
	}
	rel.Problems = nil

	data := ctx.Data
	if rel.ReleaseManifest != "" {
		manifest := &v1alpha1.ReleaseManifest{}
		err := r.Get(ctx, k8s.Key(ctx.Namespace, rel.ReleaseManifest), manifest)

		switch {
		case err == nil:
			data = &manifest.Data

		case k8s.IsNotFound(err):
			msg := fmt.Sprintf(`ReleaseManifest "%s" for active Release "%s" not found.`, rel.ReleaseManifest, rel.Id)
			rel.Problems = append(rel.Problems, common.Problem{
				ObservedTime: ctx.Now,
				Problem: api.Problem{
					Type:    api.ProblemTypeAppDeploymentNotFound,
					Message: msg,
					Causes: []api.ProblemSource{
						{
							Kind:               api.ProblemSourceKindVirtualEnvironment,
							Name:               ctx.Name,
							ObservedGeneration: ctx.Generation,
							Path:               "$.status.activeRelease.releaseManifest",
							Value:              &rel.ReleaseManifest,
						},
						{
							Kind: api.ProblemSourceKindReleaseManifest,
							Name: rel.ReleaseManifest,
						},
					},
				},
			})

		case err != nil:
			return err
		}
	}

	for appName, app := range rel.Apps {
		appDep := &v1alpha1.AppDeployment{}
		err := r.Get(ctx, k8s.Key(ctx.Namespace, app.AppDeployment), appDep)

		switch {
		case err == nil:
			progressing := k8s.Condition(appDep.Status.Conditions, api.ConditionTypeProgressing)
			available := k8s.Condition(appDep.Status.Conditions, api.ConditionTypeAvailable)

			appDepProblems, err := appDep.Validate(data,
				func(name string, typ api.ComponentType) (api.Adapter, error) {
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
					ObservedTime: ctx.Now,
					Problem:      p,
				})
			}

			if available.Status == metav1.ConditionFalse &&
				available.Reason != api.ConditionReasonProblemsFound {

				msg := fmt.Sprintf(`One or more Component Deployments of AppDeployment "%s" are unavailable.`, appDep.Name)
				value := fmt.Sprintf("%s,%s", available.Status, available.Reason)
				rel.Problems = append(rel.Problems, common.Problem{
					ObservedTime: ctx.Now,
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
					ObservedTime: ctx.Now,
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

			if *ctx.Policy.VersionRequired && app.Version == "" {
				msg := fmt.Sprintf(`Version is required but not set for App "%s".`, appName)
				value := fmt.Sprint(*ctx.Policy.VersionRequired)
				rel.Problems = append(rel.Problems, common.Problem{
					ObservedTime: ctx.Now,
					Problem: api.Problem{
						Type:    api.ProblemTypePolicyViolation,
						Message: msg,
						Causes: []api.ProblemSource{
							{
								Kind:               api.ProblemSourceKindVirtualEnvironment,
								Name:               ctx.Name,
								ObservedGeneration: ctx.Generation,
								Path:               "$.spec.releasePolicy.versionRequired",
								Value:              &value,
							},
							{
								Kind:               api.ProblemSourceKindVirtualEnvironment,
								Name:               ctx.Name,
								ObservedGeneration: ctx.Generation,
								Path:               fmt.Sprintf("$.spec.release.apps.%s.version", appName),
								Value:              &app.Version,
							},
						},
					},
				})
			}

			if app.Version != "" && app.Version != appDep.Spec.Version {
				msg := fmt.Sprintf(`AppDeployment "%s" version "%s" does not match App "%s" version "%s".`,
					appDep.Name, appDep.Spec.Version, appName, app.Version)
				rel.Problems = append(rel.Problems, common.Problem{
					ObservedTime: ctx.Now,
					Problem: api.Problem{
						Type:    api.ProblemTypeVersionConflict,
						Message: msg,
						Causes: []api.ProblemSource{
							{
								Kind:               api.ProblemSourceKindVirtualEnvironment,
								Name:               ctx.Name,
								ObservedGeneration: ctx.Generation,
								Path:               fmt.Sprintf("$.spec.release.apps.%s.version", appName),
								Value:              &app.Version,
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
			msg := fmt.Sprintf(`AppDeployment "%s" for App "%s" not found.`, app.AppDeployment, appName)
			rel.Problems = append(rel.Problems, common.Problem{
				ObservedTime: ctx.Now,
				Problem: api.Problem{
					Type:    api.ProblemTypeAppDeploymentNotFound,
					Message: msg,
					Causes: []api.ProblemSource{
						{
							Kind:               api.ProblemSourceKindVirtualEnvironment,
							Name:               ctx.Name,
							ObservedGeneration: ctx.Generation,
							Path:               fmt.Sprintf("$.spec.release.apps.%s.appDeployment", appName),
							Value:              &app.AppDeployment,
						},
						{
							Kind: api.ProblemSourceKindAppDeployment,
							Name: app.AppDeployment,
						},
					},
				},
			})

		case err != nil:
			return err
		}
	}

	return nil
}

func (r *VirtualEnvReconciler) cleanupManifests(ctx context.Context, ve *v1alpha1.VirtualEnvironment) error {
	list := &v1alpha1.ReleaseManifestList{}
	if err := r.List(ctx, list, client.InNamespace(ve.Namespace), client.MatchingLabels{
		api.LabelK8sVirtualEnvironment: ve.Name,
	}); err != nil {
		return err
	}

	r.log.Debugf("found %d ReleaseManifests using VirtualEnvironment '%s'", len(list.Items), ve.Name)

	for _, manifest := range list.Items {
		if ve.Status.ActiveRelease != nil &&
			manifest.Name == ve.Status.ActiveRelease.ReleaseManifest {
			continue
		}

		// Check if ReleaseManifest is present in history.
		var found bool
		for _, rel := range ve.Status.ReleaseHistory {
			if manifest.Name == rel.ReleaseManifest {
				found = true
				break
			}
		}

		if !found {
			r.log.Debugf("deleting unused ReleaseManifest '%s'", manifest.Name)
			if k8s.RemoveFinalizer(&manifest, api.FinalizerReleaseProtection) {
				if err := r.Update(ctx, &manifest); err != nil {
					return err
				}
			}
			if err := r.Delete(ctx, &manifest); k8s.IgnoreNotFound(err) != nil {
				return err
			}
		}
	}

	return nil
}

func (r *VirtualEnvReconciler) watchAppDeployments(ctx context.Context, obj client.Object) []reconcile.Request {
	veList := &v1alpha1.VirtualEnvironmentList{}
	if err := r.List(ctx, veList); err != nil {
		r.log.Error(err)
		return []reconcile.Request{}
	}

	var reqs []reconcile.Request
	for _, ve := range veList.Items {
		reqs = append(reqs, filterRelease(obj.GetName(), &ve, ve.Status.ActiveRelease)...)
		reqs = append(reqs, filterRelease(obj.GetName(), &ve, ve.Status.PendingRelease)...)
	}

	return reqs
}

func filterRelease(appDep string, ve *v1alpha1.VirtualEnvironment, rel *v1alpha1.ReleaseStatus) []reconcile.Request {
	if rel == nil {
		return nil
	}

	var reqs []reconcile.Request
	for _, app := range rel.Apps {
		if app.AppDeployment == appDep {
			reqs = append(reqs, reconcile.Request{
				NamespacedName: k8s.Key(ve.Namespace, ve.Name),
			})
		}
	}

	return reqs
}

func (r *VirtualEnvReconciler) watchEnvironment(ctx context.Context, env client.Object) []reconcile.Request {
	veList := &v1alpha1.VirtualEnvironmentList{}
	if err := r.List(ctx, veList, client.MatchingLabels{api.LabelK8sEnvironment: env.GetName()}); err != nil {
		r.log.Error(err)
		return []reconcile.Request{}
	}

	reqs := make([]reconcile.Request, len(veList.Items))
	for i, ve := range veList.Items {
		reqs[i] = reconcile.Request{
			NamespacedName: k8s.Key(ve.Namespace, ve.Name),
		}
	}

	return reqs
}

func releaseEqual(lhs *v1alpha1.ReleaseStatus, rhs *v1alpha1.Release) bool {
	switch {
	case lhs == nil && rhs == nil:
		return true
	case lhs != nil && rhs == nil:
		return false
	case lhs == nil && rhs != nil:
		return false
	}

	return k8s.DeepEqual(&lhs.Release, rhs)
}
