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
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/mitchellh/hashstructure/v2"
	"github.com/xigxog/kubefox/api"
	common "github.com/xigxog/kubefox/api/kubernetes"
	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/build"
	"github.com/xigxog/kubefox/components/operator/defaults"
	"github.com/xigxog/kubefox/components/operator/templates"
	"github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/k8s"
	"github.com/xigxog/kubefox/logkf"
	"golang.org/x/mod/semver"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ComponentManager struct {
	*Client

	mutex sync.Mutex

	instance string
	log      *logkf.Logger
}

type TemplateData struct {
	templates.Data

	Template string
	Obj      client.Object
}

func NewComponentManager(instance string, cli *Client) *ComponentManager {
	return &ComponentManager{
		Client:   cli,
		instance: instance,
		log:      logkf.Global,
	}
}

func (cm *ComponentManager) SetupComponent(ctx context.Context, td *TemplateData) (bool, error) {
	log := cm.log.
		WithInstance(td.Instance.Name).
		WithPlatform(td.Platform.Name).
		WithComponent(td.Component.Component)

	hash, err := hashstructure.Hash(td.Data, hashstructure.FormatV2, nil)
	if err != nil {
		return false, err
	}
	td.Data.Hash = strconv.Itoa(int(hash))

	log.Debugf("setting up Component '%s'", td.Name())

	name := k8s.Key(td.Platform.Namespace, td.Name())
	if err := cm.Get(ctx, name, td.Obj); client.IgnoreNotFound(err) != nil {
		return false, log.ErrorN("unable to fetch Component workload: %w", err)
	}

	curHash := td.Obj.GetAnnotations()[api.AnnotationTemplateDataHash]
	if curHash != td.Data.Hash {
		log.Infof("change to template data detected, applying template")
		return false, cm.ApplyTemplate(ctx, td.Template, &td.Data, log)
	}
	ver := td.Obj.GetLabels()[api.LabelK8sRuntimeVersion]
	if semver.Compare(ver, build.Info.Version) < 0 {
		log.Infof("version upgrade detected, applying template to upgrade %s->%s", ver, build.Info.Version)
		return false, cm.ApplyTemplate(ctx, td.Template, &td.Data, log)
	}

	var available int32
	switch obj := td.Obj.(type) {
	case *appsv1.StatefulSet:
		available = obj.Status.AvailableReplicas
		if obj.Status.CurrentRevision != obj.Status.UpdateRevision {
			return false, nil // StatefulSet updating
		}

	case *appsv1.Deployment:
		available = obj.Status.AvailableReplicas

	case *appsv1.DaemonSet:
		if obj.Status.CurrentNumberScheduled-obj.Status.DesiredNumberScheduled < 0 {
			available = 0
		} else {
			available = obj.Status.NumberAvailable
		}
	}
	if available <= 0 {
		log.Debug("Component is not available, applying template to ensure correct state")
		return false, cm.ApplyTemplate(ctx, td.Template, &td.Data, log)
	}

	log.Debugf("Component '%s' available", td.Name())
	return true, nil
}

func (cm *ComponentManager) ReconcileApps(ctx context.Context, namespace string) (bool, error) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	platform, err := cm.GetPlatform(ctx, namespace)
	if err != nil {
		return false, k8s.IgnoreNotFound(err)
	}
	if !k8s.IsAvailable(platform.Status.Conditions) {
		cm.log.Debug("Platform not available")
		return false, nil
	}

	log := cm.log.With(
		logkf.KeyInstance, cm.instance,
		logkf.KeyPlatform, platform.Name,
	)

	appDeps := &v1alpha1.AppDeploymentList{}
	if err := cm.List(ctx, appDeps, client.InNamespace(platform.Namespace)); err != nil {
		return false, err
	}

	log.Debugf("found %d AppDeployments", len(appDeps.Items))

	maxEventSize := platform.Spec.Events.MaxSize.Value()
	if platform.Spec.Events.MaxSize.IsZero() {
		maxEventSize = api.DefaultMaxEventSizeBytes
	}
	td := TemplateData{
		Data: templates.Data{
			Instance: templates.Instance{
				Name:           cm.instance,
				BootstrapImage: BootstrapImage,
			},
			Platform: templates.Platform{
				Name:      platform.Name,
				Namespace: platform.Namespace,
			},
			Owner: []*metav1.OwnerReference{
				metav1.NewControllerRef(platform, platform.GroupVersionKind()),
			},

			BuildInfo: build.Info,
		},
	}

	compsTplData := map[string]*TemplateData{}
	for _, appDep := range appDeps.Items {
		for compName, comp := range appDep.Spec.Components {
			image := comp.Image
			if image == "" {
				reg := strings.TrimSuffix(appDep.Spec.ContainerRegistry, "/")
				image = fmt.Sprintf("%s/%s:%s", reg, compName, comp.Commit)
			}

			compTD := td.ForComponent("component", &appsv1.Deployment{}, &defaults.Component, templates.Component{
				Component: core.NewComponent(
					api.ComponentTypeKubeFox,
					appDep.Spec.AppName,
					compName,
					comp.Commit,
				),
				Image:           image,
				ImagePullSecret: appDep.Spec.ImagePullSecretName,
			})
			compTD.Values = map[string]any{api.ValKeyMaxEventSize: maxEventSize}

			compsTplData[compTD.Component.GroupKey()] = compTD
		}
	}
	log.Debugf("found %d unique Components", len(compsTplData))

	allAvailable := true
	depMap := make(map[string]*appsv1.Deployment, len(compsTplData))
	for _, compTD := range compsTplData {
		available, err := cm.SetupComponent(ctx, compTD)
		if err != nil {
			return false, err
		}
		if !available {
			allAvailable = false
		}

		depMap[compTD.Component.GroupKey()] = compTD.Obj.(*appsv1.Deployment)
	}

	now := metav1.Now()
	for _, appDep := range appDeps.Items {
		problems, err := appDep.Validate(nil,
			func(name string, typ api.ComponentType) (common.Adapter, error) {
				switch typ {
				case api.ComponentTypeHTTPAdapter:
					a := &v1alpha1.HTTPAdapter{}
					return a, cm.Get(ctx, k8s.Key(appDep.Namespace, name), a)

				default:
					return nil, core.ErrNotFound()
				}
			})
		if err != nil {
			return false, err
		}

		var available, progressing *metav1.Condition
		if len(problems) > 0 {
			available = &metav1.Condition{
				Type:    api.ConditionTypeAvailable,
				Status:  metav1.ConditionFalse,
				Reason:  api.ConditionReasonProblemsFound,
				Message: "One or more problems found, see `status.problems` for details.",
			}
			progressing = &metav1.Condition{
				Type:    api.ConditionTypeProgressing,
				Status:  metav1.ConditionFalse,
				Reason:  api.ConditionReasonProblemsFound,
				Message: "One or more problems found, see `status.problems` for details.",
			}
		} else {
			c, p := availableCondition(&appDep.Spec, depMap)
			available = c
			problems = append(problems, p...)

			c, p = progressingCondition(&appDep.Spec, depMap)
			progressing = c
			problems = append(problems, p...)
		}

		available.ObservedGeneration = appDep.Generation
		progressing.ObservedGeneration = appDep.Generation

		appDep.Status.Conditions = k8s.UpdateConditions(now, appDep.Status.Conditions, available, progressing)
		appDep.Status.Problems = problems

		if err := cm.Status().Update(ctx, &appDep); err != nil {
			return false, err
		}

		log.Debugf("%s; available: %s, progressing: %s", k8s.ToString(&appDep), available.Status, progressing.Status)
	}

	depList := &appsv1.DeploymentList{}
	if err := cm.List(ctx, depList,
		client.InNamespace(platform.Namespace),
		// filter out Platform Components
		client.MatchingLabels{api.LabelK8sComponentType: string(api.ComponentTypeKubeFox)},
	); err != nil {
		return false, err
	}
	log.Debugf("found %d existing Component Deployments", len(depList.Items))

	for _, d := range depList.Items {
		if _, found := compsTplData[d.Name]; !found {
			log.Debugf("deleting Component Deployment '%s'", d.Name)

			// TODO turn annotation into list of resources created to avoid
			// leaking data.
			tdStr := d.Annotations[api.AnnotationTemplateData]
			data := &templates.Data{}
			if err := json.Unmarshal([]byte(tdStr), data); err != nil {
				return false, err
			}
			if err := cm.DeleteTemplate(ctx, "component", data, log); err != nil {
				return false, err
			}
		}
	}

	log.Debugf("apps reconciled")

	return allAvailable, nil
}

func (td *TemplateData) ForComponent(template string, obj client.Object, defs *common.ContainerSpec, comp templates.Component) *TemplateData {
	newTD := &TemplateData{
		Template: template,
		Obj:      obj,
		Data: templates.Data{
			Instance:  td.Instance,
			Platform:  td.Platform,
			Owner:     td.Owner,
			Logger:    td.Logger,
			BuildInfo: td.BuildInfo,
			Component: comp,
			Values:    make(map[string]any),
		},
	}

	// Copy values.
	for k, v := range td.Values {
		newTD.Values[k] = v
	}

	defaults.Set(&newTD.Component.ContainerSpec, defs)

	if cpu := newTD.Component.Resources.Limits.Cpu(); !cpu.IsZero() {
		newTD.Values["GOMAXPROCS"] = cpu.Value()
	}
	if mem := newTD.Component.Resources.Limits.Memory(); !mem.IsZero() {
		newTD.Values["GOMEMLIMIT"] = int(float64(mem.Value()) * 0.9)
	}

	return newTD
}

func availableCondition(
	spec *v1alpha1.AppDeploymentSpec, deployments map[string]*appsv1.Deployment) (*metav1.Condition, api.Problems) {

	var (
		condition   = &metav1.Condition{Type: api.ConditionTypeAvailable}
		available   = false
		unavailable = false
		problems    api.Problems
	)
	for name, c := range spec.Components {
		comp := core.NewComponent(
			api.ComponentTypeKubeFox,
			spec.AppName,
			name,
			c.Commit,
		)
		dep, found := deployments[comp.GroupKey()]
		if !found || dep == nil {
			problems = append(problems,
				api.Problem{
					Type:    api.ProblemTypeDeploymentNotFound,
					Message: "Component Deployment not found.",
					Causes: []api.ProblemSource{
						{
							Kind: api.ProblemSourceKindComponent,
							Name: name,
						},
					},
				},
			)
			continue
		}

		cond := findCondition(dep.Status, appsv1.DeploymentAvailable)
		switch {
		case cond.Status == corev1.ConditionTrue:
			available = true
		default:
			unavailable = true
			problems = append(problems,
				api.Problem{
					Type:    api.ProblemTypeDeploymentUnavailable,
					Message: cond.Message,
					Causes: []api.ProblemSource{
						{
							Kind:               api.ProblemSourceKindDeployment,
							Name:               dep.Name,
							ObservedGeneration: dep.Generation,
						},
						{
							Kind: api.ProblemSourceKindComponent,
							Name: name,
						},
					},
				},
			)
		}
	}

	switch {
	case unavailable:
		condition.Status = metav1.ConditionFalse
		condition.Reason = api.ConditionReasonComponentUnavailable
		condition.Message = "One or more Component Deployments is unavailable, see `status.problems` for details."
	case available:
		condition.Status = metav1.ConditionTrue
		condition.Reason = api.ConditionReasonComponentsAvailable
		condition.Message = "Component Deployments have minimum required Pods available."
	default:
		condition.Status = metav1.ConditionFalse
		condition.Reason = api.ConditionReasonProblemsFound
		condition.Message = "One or more problems found, see `status.problems` for details."
	}

	return condition, problems
}

func progressingCondition(
	spec *v1alpha1.AppDeploymentSpec, deployments map[string]*appsv1.Deployment) (*metav1.Condition, api.Problems) {

	var (
		condition   = &metav1.Condition{Type: api.ConditionTypeProgressing}
		available   = false
		progressing = false
		failed      = false
		problems    api.Problems
	)
	for name, c := range spec.Components {
		comp := core.NewComponent(
			api.ComponentTypeKubeFox,
			spec.AppName,
			name,
			c.Commit,
		)
		dep, found := deployments[comp.GroupKey()]
		if !found || dep == nil {
			// availableCondition adds the ProblemTypeDeploymentNotFound
			// problem, do not need to add again.
			continue
		}

		cond := findCondition(dep.Status, appsv1.DeploymentProgressing)
		switch {
		case cond.Reason == "NewReplicaSetAvailable":
			available = true
		case cond.Status == corev1.ConditionTrue:
			progressing = true
		default:
			failed = true
			problems = append(problems,
				api.Problem{
					Type:    api.ProblemTypeDeploymentFailed,
					Message: cond.Message,
					Causes: []api.ProblemSource{
						{
							Kind:               api.ProblemSourceKindDeployment,
							Name:               dep.Name,
							ObservedGeneration: dep.Generation,
						},
						{
							Kind: api.ProblemSourceKindComponent,
							Name: name,
						},
					},
				},
			)
		}
	}

	switch {
	case failed:
		condition.Status = metav1.ConditionFalse
		condition.Reason = api.ConditionReasonComponentDeploymentFailed
		condition.Message = "One or more Component Deployments has failed, see `status.problems` for details."
	case progressing:
		condition.Status = metav1.ConditionTrue
		condition.Reason = api.ConditionReasonComponentDeploymentProgressing
		condition.Message = "One or more Component Deployments is progressing."
	case available:
		condition.Status = metav1.ConditionFalse
		condition.Reason = api.ConditionReasonComponentsDeployed
		condition.Message = "Component Deployments completed successfully."
	default:
		condition.Status = metav1.ConditionFalse
		condition.Reason = api.ConditionReasonProblemsFound
		condition.Message = "One or more problems found, see `status.problems` for details."
	}

	return condition, problems
}

func findCondition(status appsv1.DeploymentStatus, condType appsv1.DeploymentConditionType) appsv1.DeploymentCondition {
	for _, c := range status.Conditions {
		if c.Type == condType {
			return c
		}
	}

	return appsv1.DeploymentCondition{Type: condType, Status: corev1.ConditionUnknown}
}

// func componentKey(name, commit string) string {
// 	return fmt.Sprintf("%s-%s", name, commit)
// }
