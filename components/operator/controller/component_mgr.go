package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"github.com/mitchellh/hashstructure/v2"
	"github.com/xigxog/kubefox/api"
	common "github.com/xigxog/kubefox/api/kubernetes"
	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/build"
	"github.com/xigxog/kubefox/components/operator/templates"
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
	log := cm.log.With(
		logkf.KeyInstance, td.Instance.Name,
		logkf.KeyPlatform, td.Platform.Name,
		logkf.KeyComponentName, td.ComponentFullName(),
	)

	hash, err := hashstructure.Hash(td.Data, hashstructure.FormatV2, nil)
	if err != nil {
		return false, err
	}
	td.Data.Hash = strconv.Itoa(int(hash))

	log.Debugf("setting up Component '%s'", td.ComponentFullName())

	name := k8s.Key(td.Namespace(), td.ComponentFullName())
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

	log.Debugf("Component '%s' available", td.ComponentFullName())
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

	appDepList := &v1alpha1.AppDeploymentList{}
	if err := cm.List(ctx, appDepList, client.InNamespace(platform.Namespace)); err != nil {
		return false, err
	}
	log.Debugf("found %d AppDeployments", len(appDepList.Items))

	compDepList := &appsv1.DeploymentList{}
	if err := cm.List(ctx, compDepList,
		client.InNamespace(platform.Namespace),
		client.HasLabels{api.LabelK8sAppComponent, api.LabelK8sAppName}, // filter out Platform Components
	); err != nil {
		return false, err
	}

	maxEventSize := platform.Spec.Events.MaxSize.Value()
	if platform.Spec.Events.MaxSize.IsZero() {
		maxEventSize = api.DefaultMaxEventSizeBytes
	}
	compMap := make(map[string]*TemplateData)
	for _, appDep := range appDepList.Items {
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

		for compName, comp := range appDep.Spec.Components {
			image := comp.Image
			if image == "" {
				image = fmt.Sprintf("%s/%s:%s", appDep.Spec.ContainerRegistry, compName, comp.Commit)
			}

			compTd := td.ForComponent("component", &appsv1.Deployment{}, &ComponentDefaults, templates.Component{
				Name:            compName,
				Commit:          comp.Commit,
				App:             appDep.Spec.AppName,
				Image:           image,
				ImagePullPolicy: appDep.Spec.ImagePullSecretName,
			})
			compTd.Values = map[string]any{api.ValKeyMaxEventSize: maxEventSize}

			compMap[compTd.ComponentFullName()] = compTd
		}
	}
	log.Debugf("found %d unique Components", len(compMap))

	allAvailable := true
	depMap := make(map[string]*appsv1.Deployment, len(compMap))
	for _, compTD := range compMap {
		available, err := cm.SetupComponent(ctx, compTD)
		if err != nil {
			return false, err
		}
		allAvailable = allAvailable && available

		key := CompDepKey(compTD.Component.Name, compTD.Component.Commit)
		depMap[key] = compTD.Obj.(*appsv1.Deployment)
	}

	now := metav1.Now()
	for _, appDep := range appDepList.Items {
		curStatus := appDep.Status.DeepCopy()

		available := AvailableCondition(&appDep, depMap)
		progressing := ProgressingCondition(&appDep, depMap)
		appDep.Status.Conditions = k8s.UpdateConditions(now, appDep.Status.Conditions, available, progressing)

		if !k8s.DeepEqual(&appDep.Status, curStatus) {
			if err := cm.ApplyStatus(ctx, &appDep); err != nil {
				log.Error(err)
			}
		}

		log.Debugf("AppDeployment '%s/%s'; available: %s, progressing: %s",
			appDep.Namespace, appDep.Name, available.Status, progressing.Status)
	}

	for _, d := range compDepList.Items {
		if _, found := compMap[d.Name]; !found {
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

	SetDefaults(&newTD.Component.ContainerSpec, defs)

	if cpu := newTD.Component.Resources.Limits.Cpu(); !cpu.IsZero() {
		newTD.Values["GOMAXPROCS"] = cpu.Value()
	}
	if mem := newTD.Component.Resources.Limits.Memory(); !mem.IsZero() {
		newTD.Values["GOMEMLIMIT"] = int(float64(mem.Value()) * 0.9)
	}

	return newTD
}

func CompDepKey(name, commit string) string {
	return fmt.Sprintf("%s-%s", name, commit)
}

func AvailableCondition(appDep *v1alpha1.AppDeployment, depMap map[string]*appsv1.Deployment) *metav1.Condition {
	available, depName, depCond := isAppDeployment(appsv1.DeploymentAvailable, &appDep.Spec, depMap)
	cond := &metav1.Condition{
		Type:               api.ConditionTypeAvailable,
		Status:             metav1.ConditionTrue,
		ObservedGeneration: appDep.Generation,
		Reason:             api.ConditionReasonComponentsAvailable,
		Message:            "Component Deployments have minimum availability.",
	}
	if !available {
		cond.Status = metav1.ConditionFalse
		cond.Reason = api.ConditionReasonComponentUnavailable
		cond.Message = fmt.Sprintf(`Component Deployment "%s" unavailable; %s`, depName, depCond.Message)
	}

	return cond
}

func ProgressingCondition(appDep *v1alpha1.AppDeployment, depMap map[string]*appsv1.Deployment) *metav1.Condition {
	progressing, _, depCond := isAppDeployment(appsv1.DeploymentProgressing, &appDep.Spec, depMap)
	cond := &metav1.Condition{
		Type:               api.ConditionTypeProgressing,
		Status:             metav1.ConditionTrue,
		ObservedGeneration: appDep.Generation,
		Reason:             api.ConditionReasonComponentDeploymentProgressing,
		Message:            depCond.Message,
	}

	switch {
	case progressing && depCond.Reason == "NewReplicaSetAvailable":
		cond.Status = metav1.ConditionFalse
		cond.Reason = api.ConditionReasonComponentsDeployed
		cond.Message = "Component Deployments have successfully progressed."

	case !progressing:
		cond.Status = metav1.ConditionFalse
		cond.Reason = api.ConditionReasonComponentDeploymentFailed
	}

	return cond
}

func isAppDeployment(condType appsv1.DeploymentConditionType,
	spec *v1alpha1.AppDeploymentSpec, depMap map[string]*appsv1.Deployment) (bool, string, *appsv1.DeploymentCondition) {

	cond := &appsv1.DeploymentCondition{Type: condType, Status: corev1.ConditionUnknown}
	for name, c := range spec.Components {
		key := CompDepKey(name, c.Commit)
		dep, found := depMap[key]
		if !found || dep == nil {
			return false, "", &appsv1.DeploymentCondition{Type: condType, Status: corev1.ConditionUnknown}
		}
		if cond = condition(dep.Status, condType); cond.Status == corev1.ConditionFalse {
			return false, dep.Name, cond
		}
	}

	return true, "", cond
}

func condition(status appsv1.DeploymentStatus, condType appsv1.DeploymentConditionType) *appsv1.DeploymentCondition {
	for _, c := range status.Conditions {
		if c.Type == condType {
			return &c
		}
	}

	return &appsv1.DeploymentCondition{Type: condType, Status: corev1.ConditionUnknown}
}
