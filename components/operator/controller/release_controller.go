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

	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/k8s"
	"github.com/xigxog/kubefox/logkf"
	"github.com/xigxog/kubefox/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ReleaseReconciler reconciles a Release object
type ReleaseReconciler struct {
	*Client

	log *logkf.Logger
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReleaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.log = logkf.Global.With(logkf.KeyController, "release")
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Release{}).
		Watches(
			&v1alpha1.AppDeployment{},
			handler.EnqueueRequestsFromMapFunc(r.watchAppDeployment),
		).
		Watches(
			&v1alpha1.ClusterVirtualEnv{},
			handler.EnqueueRequestsFromMapFunc(r.watchClusterVirtualEnv),
		).
		Watches(
			&v1alpha1.VirtualEnv{},
			handler.EnqueueRequestsFromMapFunc(r.watchVirtualEnv),
		).
		Watches(
			&v1alpha1.VirtualEnvSnapshot{},
			handler.EnqueueRequestsFromMapFunc(r.watchVirtualEnvSnapshot),
		).
		Complete(r)
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ReleaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.log.With(
		"namespace", req.Namespace,
		"name", req.Name,
	)
	log.Debugf("reconciling Release '%s/%s'", req.Namespace, req.Name)

	result, err := r.reconcile(ctx, req, log)
	if IsFailedWebhookErr(err) {
		log.Debug("reconcile failed because of webhook, retrying in 15 seconds")
		return ctrl.Result{RequeueAfter: time.Second * 15}, nil
	}

	log.Debugf("reconciling Release '%s/%s' done", req.Namespace, req.Name)

	return result, err
}

func (r *ReleaseReconciler) reconcile(ctx context.Context, req ctrl.Request, log *logkf.Logger) (ctrl.Result, error) {
	rel := &v1alpha1.Release{}
	err := r.Get(ctx, req.NamespacedName, rel)
	if k8s.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, err
	}
	if k8s.IsNotFound(err) || rel.ResourceVersion == "0" {
		log.Debugf("Release '%s/%s does not exists'", req.Namespace, req.Name)
		return ctrl.Result{}, nil
	}
	if rel.DeletionTimestamp != nil {
		return ctrl.Result{}, r.releaseDeleted(ctx, rel, log)
	}

	now := metav1.Now()
	origRel := rel.DeepCopy()

	isCurrent := rel.Status.Current != nil && rel.Status.Current.AppDeployment == rel.Spec.AppDeployment
	isRequested := rel.Status.Requested != nil && rel.Status.Requested.AppDeployment == rel.Spec.AppDeployment
	if !isCurrent && !isRequested {
		isRequested = true
		rel.Status.Requested = &v1alpha1.ReleaseStatusEntry{
			AppDeployment:      rel.Spec.AppDeployment,
			VirtualEnvSnapshot: rel.Spec.VirtualEnvSnapshot,
			RequestTime:        now,
		}
	}

	var myStatus *v1alpha1.ReleaseStatusEntry
	if isCurrent {
		myStatus = rel.Status.Current
	} else {
		myStatus = rel.Status.Requested
	}

	// Need finalizer so Release object is available after deletion providing
	// access to AppDeployment and VirtualEnvs names.
	if k8s.AddFinalizer(rel, api.FinalizerReleaseProtection) {
		return ctrl.Result{}, r.Merge(ctx, rel, origRel)
	}

	k8s.UpdateLabel(rel, api.LabelK8sVirtualEnvSnapshot, rel.Spec.VirtualEnvSnapshot)
	k8s.UpdateLabel(rel, api.LabelK8sVirtualEnv, rel.Name)
	k8s.UpdateLabel(rel, api.LabelK8sAppDeployment, rel.Spec.AppDeployment.Name)
	k8s.UpdateLabel(rel, api.LabelK8sAppVersion, rel.Spec.AppDeployment.Version)

	appDep := &v1alpha1.AppDeployment{}
	err = r.Get(ctx, k8s.Key(rel.Namespace, rel.Spec.AppDeployment.Name), appDep)
	switch {
	case err == nil:
		log.Debugf("found AppDeployment '%s'", appDep.Name)
		origAppDep := appDep.DeepCopy()

		k8s.UpdateLabel(rel, api.LabelK8sAppCommit, appDep.Spec.App.Commit)
		k8s.UpdateLabel(rel, api.LabelK8sAppCommitShort, utils.ShortCommit(appDep.Spec.App.Commit))
		k8s.UpdateLabel(rel, api.LabelK8sAppTag, appDep.Spec.App.Tag)
		k8s.UpdateLabel(rel, api.LabelK8sAppBranch, appDep.Spec.App.Branch)

		if k8s.AddFinalizer(appDep, api.FinalizerReleaseProtection) {
			if err := r.Merge(ctx, appDep, origAppDep); err != nil {
				return ctrl.Result{}, err
			}
		}

		relVer, appDepVer := rel.Spec.AppDeployment.Version, appDep.Spec.Version
		switch {
		case relVer != "" && relVer != appDepVer:
			myStatus.AvailableTime = nil
			// TODO set condition

		case !appDep.Status.Available:
			myStatus.AvailableTime = nil
			// TODO set condition

		case appDep.Status.Available:
			if myStatus.AvailableTime == nil {
				myStatus.AvailableTime = &now
			}
		}

	case k8s.IsNotFound(err):
		myStatus.AvailableTime = nil
		k8s.RemoveLabel(rel, api.LabelK8sAppCommit)
		k8s.RemoveLabel(rel, api.LabelK8sAppCommitShort)
		k8s.RemoveLabel(rel, api.LabelK8sAppTag)
		k8s.RemoveLabel(rel, api.LabelK8sAppBranch)
		// TODO set condition

	case err != nil:
		return ctrl.Result{}, err
	}

	var (
		appDepPolicy api.AppDeploymentPolicy = api.AppDeploymentPolicyVersionRequired
		envPolicy    api.VirtualEnvPolicy    = api.VirtualEnvPolicySnapshotRequired
	)
	envId := utils.First(rel.Spec.VirtualEnvSnapshot, rel.Name)
	envObj, err := r.GetVirtualEnvObj(ctx, rel.Namespace, envId, rel.Spec.VirtualEnvSnapshot != "")
	switch {
	case err == nil:
		if envObj.GetEnvName() != rel.Name {
			myStatus.AvailableTime = nil
			// TODO set condition
			break
		}
		log.Debugf("found VirtualEnvObject '%s' of type '%T'", envObj.GetName(), envObj)

		if _, ok := envObj.(*v1alpha1.VirtualEnvSnapshot); ok {
			origEnv := envObj.DeepCopyObject().(client.Object)
			if k8s.AddFinalizer(envObj, api.FinalizerReleaseProtection) {
				if err := r.Merge(ctx, envObj, origEnv); err != nil {
					return ctrl.Result{}, err
				}
			}
		}

		if envObj.GetParent() != "" {
			parentEnv := &v1alpha1.ClusterVirtualEnv{}
			if err := r.Get(ctx, k8s.Key("", envObj.GetParent()), parentEnv); err != nil {
				myStatus.AvailableTime = nil
				// TODO set condition
				break
			}
			if parentEnv.GetReleasePolicy() != nil {
				if parentEnv.GetReleasePolicy().AppDeploymentPolicy != "" {
					appDepPolicy = parentEnv.GetReleasePolicy().AppDeploymentPolicy
				}
				if parentEnv.GetReleasePolicy().VirtualEnvPolicy != "" {
					envPolicy = parentEnv.GetReleasePolicy().VirtualEnvPolicy
				}
			}
		}

		if envObj.GetReleasePolicy() != nil {
			if envObj.GetReleasePolicy().AppDeploymentPolicy != "" {
				appDepPolicy = envObj.GetReleasePolicy().AppDeploymentPolicy
			}
			if envObj.GetReleasePolicy().VirtualEnvPolicy != "" {
				envPolicy = envObj.GetReleasePolicy().VirtualEnvPolicy
			}
		}

	case k8s.IsNotFound(err):
		myStatus.AvailableTime = nil
		// TODO set condition

	case err != nil:
		return ctrl.Result{}, err
	}

	if appDepPolicy == api.AppDeploymentPolicyVersionRequired && rel.Spec.AppDeployment.Version == "" {
		myStatus.AvailableTime = nil
		// TODO set condition
	}
	if envPolicy == api.VirtualEnvPolicySnapshotRequired && rel.Spec.VirtualEnvSnapshot == "" {
		myStatus.AvailableTime = nil
		// TODO set condition
	}

	if !k8s.DeepEqual(rel.ObjectMeta, origRel.ObjectMeta) {
		log.Debug("Release modified, updating")
		return ctrl.Result{}, r.Apply(ctx, rel)
	}

	if isCurrent {
		rel.Status.Requested = nil

	} else if isRequested && myStatus.AvailableTime != nil {
		if rel.Status.Current != nil {
			rel.Status.Current.ArchiveTime = &now
			rel.Status.History = append([]v1alpha1.ReleaseStatusEntry{*rel.Status.Current}, rel.Status.History...)
			log.Debug("history updated")
		}
		rel.Status.Current = myStatus
		rel.Status.Requested = nil
	}

	if !k8s.DeepEqual(rel.Status, origRel.Status) {
		log.Debug("Release status modified, updating")
		if err := r.MergeStatus(ctx, rel, origRel); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil

}

func (r *ReleaseReconciler) releaseDeleted(ctx context.Context, rel *v1alpha1.Release, log *logkf.Logger) error {
	log.Debugf("Release deleted, cleaning up finalizers")
	origRel := rel.DeepCopy()

	appDep := &v1alpha1.AppDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: rel.Spec.AppDeployment.Name,
		},
	}
	if err := r.checkFinalizer(ctx, rel, appDep, api.LabelK8sAppDeployment, log); err != nil {
		return err
	}

	env := &v1alpha1.VirtualEnvSnapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name: rel.Spec.VirtualEnvSnapshot,
		},
	}
	if err := r.checkFinalizer(ctx, rel, env, api.LabelK8sVirtualEnvSnapshot, log); err != nil {
		return err
	}

	if k8s.RemoveFinalizer(rel, api.FinalizerReleaseProtection) {
		log.Debugf("removing finalizer from %T '%s'", rel, rel.Name)
		if err := r.Merge(ctx, rel, origRel); k8s.IgnoreNotFound(err) != nil {
			return err
		}
	}

	return nil
}

func (r *ReleaseReconciler) checkFinalizer(ctx context.Context, rel *v1alpha1.Release, obj client.Object, label string, log *logkf.Logger) error {
	if obj.GetName() == "" {
		return nil
	}

	relList := &v1alpha1.ReleaseList{}
	if err := r.List(ctx, relList, client.InNamespace(rel.Namespace), client.MatchingLabels{
		label: obj.GetName(),
	}); err != nil {
		return err
	}
	log.Debugf("found '%d' Releases using %T '%s'", len(relList.Items), obj, obj.GetName())

	if len(relList.Items) == 0 || len(relList.Items) == 1 && relList.Items[0].Name == rel.Name {
		err := r.Get(ctx, k8s.Key(rel.Namespace, obj.GetName()), obj)
		if k8s.IgnoreNotFound(err) != nil {
			return err
		}

		orig := obj.DeepCopyObject().(client.Object)
		if err == nil && k8s.RemoveFinalizer(obj, api.FinalizerReleaseProtection) {
			log.Debugf("removing finalizer from %T '%s'", obj, obj.GetName())
			if err := r.Merge(ctx, obj, orig); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *ReleaseReconciler) watchAppDeployment(ctx context.Context, appDep client.Object) []reconcile.Request {
	return r.findReleases(ctx,
		client.InNamespace(appDep.GetNamespace()),
		client.MatchingLabels{
			api.LabelK8sAppDeployment: appDep.GetName(),
		},
	)
}

func (r *ReleaseReconciler) watchClusterVirtualEnv(ctx context.Context, env client.Object) []reconcile.Request {
	return r.findReleases(ctx,
		client.MatchingLabels{
			api.LabelK8sVirtualEnv: env.GetName(),
		},
	)
}

func (r *ReleaseReconciler) watchVirtualEnv(ctx context.Context, env client.Object) []reconcile.Request {
	return r.findReleases(ctx,
		client.MatchingLabels{
			api.LabelK8sVirtualEnv: env.GetName(),
		},
	)
}

func (r *ReleaseReconciler) watchVirtualEnvSnapshot(ctx context.Context, env client.Object) []reconcile.Request {
	return r.findReleases(ctx,
		client.InNamespace(env.GetNamespace()),
		client.MatchingLabels{
			api.LabelK8sVirtualEnvSnapshot: env.GetName(),
		},
	)
}

func (r *ReleaseReconciler) findReleases(ctx context.Context, opts ...client.ListOption) []reconcile.Request {
	relList := &v1alpha1.ReleaseList{}
	if err := r.List(ctx, relList, opts...); err != nil {
		r.log.Error(err)
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(relList.Items))
	for i, rel := range relList.Items {
		requests[i] = reconcile.Request{
			NamespacedName: k8s.Key(rel.Namespace, rel.Name),
		}
	}

	return requests
}
