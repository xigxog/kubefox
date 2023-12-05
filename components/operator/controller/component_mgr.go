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

	log.Debugf("setting up component '%s'", td.ComponentFullName())

	name := k8s.Key(td.Namespace(), td.ComponentFullName())
	if err := cm.Get(ctx, name, td.Obj); client.IgnoreNotFound(err) != nil {
		return false, log.ErrorN("unable to fetch component workload: %w", err)
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
		available = obj.Status.NumberAvailable
	}
	if available <= 0 {
		log.Debug("component is not available, applying template to ensure correct state")
		return false, cm.ApplyTemplate(ctx, td.Template, &td.Data, log)
	}

	log.Debugf("component '%s' available", td.ComponentFullName())
	return true, nil
}

func (cm *ComponentManager) ReconcileApps(ctx context.Context, namespace string) (bool, error) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	platform, err := cm.GetPlatform(ctx, namespace)
	if err != nil {
		return false, k8s.IgnoreNotFound(err)
	}
	if !platform.Status.Available {
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
		client.HasLabels{api.LabelK8sAppComponent, api.LabelK8sAppName}, // don't want Platform stuff
	); err != nil {
		return false, err
	}

	maxEventSize := platform.Spec.Events.MaxSize.Value()
	if platform.Spec.Events.MaxSize.IsZero() {
		maxEventSize = api.DefaultMaxEventSizeBytes
	}
	compMap := make(map[string]TemplateData)
	for _, appDep := range appDepList.Items {
		for compName, comp := range appDep.Spec.App.Components {
			image := comp.Image
			if image == "" {
				image = fmt.Sprintf("%s/%s:%s", appDep.Spec.App.ContainerRegistry, compName, comp.Commit)
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
					Component: templates.Component{
						Name:            compName,
						Commit:          comp.Commit,
						App:             appDep.Spec.App.Name,
						Image:           image,
						ImagePullPolicy: appDep.Spec.App.ImagePullSecretName,
					},
					Owner: []*metav1.OwnerReference{
						metav1.NewControllerRef(platform, platform.GroupVersionKind()),
					},
					Values: map[string]any{
						"maxEventSize": maxEventSize,
					},
					BuildInfo: build.Info,
				},
				Template: "component",
				Obj:      &appsv1.Deployment{},
			}
			compMap[td.ComponentFullName()] = td
		}
	}
	log.Debugf("found %d unique Components", len(compMap))

	allAvailable := true
	compReadyMap := make(map[string]bool, len(compMap))
	for _, compTD := range compMap {
		SetDefaults(&compTD.Component.ContainerSpec, &ComponentDefaults)
		available, err := cm.SetupComponent(ctx, &compTD)
		if err != nil {
			return false, err
		}

		allAvailable = allAvailable && available
		compReadyMap[CompReadyKey(compTD.Component.Name, compTD.Component.Commit)] = available
	}

	for _, d := range appDepList.Items {
		available := IsAppDeploymentAvailable(&d.Spec, compReadyMap)
		if d.Status.Available != available {
			d.Status.Available = available
			if err := cm.ApplyStatus(ctx, &d); err != nil {
				log.Error(err)
			}
		}

		log.Debugf("AppDeployment '%s/%s'; available: %t", d.Namespace, d.Name, available)
	}

	for _, d := range compDepList.Items {
		if _, found := compMap[d.Name]; !found {
			log.Debugf("deleting Component '%s'", d.Name)

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

func CompReadyKey(name, commit string) string {
	return fmt.Sprintf("%s-%s", name, commit)
}

func IsAppDeploymentAvailable(spec *v1alpha1.AppDeploymentSpec, compReadyMap map[string]bool) bool {
	for name, c := range spec.App.Components {
		key := CompReadyKey(name, c.Commit)
		if found, available := compReadyMap[key]; !found || !available {
			return false
		}
	}
	return true
}
