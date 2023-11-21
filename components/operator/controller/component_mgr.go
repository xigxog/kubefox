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
	"github.com/xigxog/kubefox/logkf"
	"github.com/xigxog/kubefox/utils"
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

type EnvRelPackage struct {
	PlatformEnv      *v1alpha1.PlatformEnv
	Active           *v1alpha1.PlatformEnvRelease
	Latest           *v1alpha1.PlatformEnvRelease
	AppDepSpecActive *v1alpha1.AppDeploymentSpec
	AppDepSpecLatest *v1alpha1.AppDeploymentSpec
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

	name := Key(td.Namespace(), td.ComponentFullName())
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
		return false, IgnoreNotFound(err)
	}
	if !platform.Status.Ready {
		return false, nil
	}

	log := cm.log.With(
		logkf.KeyInstance, cm.instance,
		logkf.KeyPlatform, platform.Name,
	)

	if !platform.Status.Ready {
		log.Debug("platform not available")
		return false, nil
	}

	relList := &v1alpha1.ReleaseList{}
	if err := cm.List(ctx, relList,
		client.InNamespace(namespace),
		client.MatchingLabels{api.LabelK8sReleaseStatus: string(api.ReleaseStatusPending)},
	); err != nil {
		return false, err
	}

	pendingRelMap := map[string][]*v1alpha1.PlatformEnvRelease{}
	for _, rel := range relList.Items {
		l := pendingRelMap[rel.Spec.Environment.Name]
		pendingRelMap[rel.Spec.Environment.Name] = append(l, &v1alpha1.PlatformEnvRelease{
			ReleaseSpec:   rel.Spec,
			ReleaseStatus: rel.Status,
			Name:          rel.Name,
		})
	}

	relCount := 0
	envRelMap := map[string]*EnvRelPackage{}
	for envName, env := range platform.Spec.Environments {
		// Find latest pending release.
		latest := &v1alpha1.PlatformEnvRelease{}
		for _, rel := range pendingRelMap[envName] {
			if rel.CreationTime.After(latest.CreationTime.Time) {
				latest = rel
			}
		}
		if env.Release != nil && env.Release.Name != "" {
			// There is an active release, keep it incase latest isn't available.
			holder := &EnvRelPackage{
				PlatformEnv: env,
				Active:      env.Release,
			}
			relCount++
			if latest.CreationTime.After(env.Release.CreationTime.Time) {
				holder.Latest = latest
				relCount++
			}
			envRelMap[envName] = holder

		} else if !latest.CreationTime.IsZero() {
			envRelMap[envName] = &EnvRelPackage{
				PlatformEnv: env,
				Latest:      latest,
			}
			relCount++
		}
	}

	specs := map[string]*v1alpha1.AppDeploymentSpec{}

	depList := &v1alpha1.AppDeploymentList{}
	if err := cm.List(ctx, depList, client.InNamespace(platform.Namespace)); err != nil {
		return false, err
	}
	for _, d := range depList.Items {
		specs[d.Name] = &d.Spec
	}

	for _, rel := range envRelMap {
		if rel.Active != nil {
			appDep := &v1alpha1.AppDeployment{}
			if err = cm.Get(ctx, Key(namespace, rel.Active.AppDeployment.Name), appDep); err != nil {
				return false, err
			}
			rel.AppDepSpecActive = &appDep.Spec
			specs[appDep.Name] = &appDep.Spec
		}
		if rel.Latest != nil {
			appDep := &v1alpha1.AppDeployment{}
			if err = cm.Get(ctx, Key(namespace, rel.Latest.AppDeployment.Name), appDep); err != nil {
				return false, err
			}
			rel.AppDepSpecLatest = &appDep.Spec
			specs[appDep.Name] = &appDep.Spec
		}
	}

	log.Debugf("found %d releases and %d app deployments", relCount, len(depList.Items))

	compDepList := &appsv1.DeploymentList{}
	if err := cm.List(ctx, compDepList,
		client.InNamespace(platform.Namespace),
		client.HasLabels{api.LabelK8sComponent, utils.CleanLabel(api.LabelK8sAppName)}, // don't want Platform stuff
	); err != nil {
		return false, err
	}

	compMap := make(map[string]TemplateData)
	for _, d := range specs {
		for n, c := range d.Components {
			image := c.Image
			if image == "" {
				image = fmt.Sprintf("%s/%s:%s", d.App.ContainerRegistry, n, c.Commit)
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
						Name:            n,
						App:             d.App.Name,
						Commit:          c.Commit,
						Image:           image,
						ImagePullPolicy: d.App.ImagePullSecretName,
					},
					Owner: []*metav1.OwnerReference{
						metav1.NewControllerRef(platform, platform.GroupVersionKind()),
					},
					BuildInfo: build.Info,
				},
				Template: "component",
				Obj:      &appsv1.Deployment{},
			}
			compMap[td.ComponentFullName()] = td
		}
	}
	log.Debugf("found %d unique app components", len(compMap))

	for _, d := range compDepList.Items {
		if _, found := compMap[d.Name]; !found {
			log.Debugf("deleting app component '%s'", d.Name)

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

	for _, d := range depList.Items {
		available := IsAppDeploymentAvailable(&d.Spec, compReadyMap)
		if d.Status.Ready != available {
			d.Status.Ready = available
			if err := cm.ApplyStatus(ctx, &d); err != nil {
				log.Error(err)
			}
		}

		log.Debugf("app deployment '%s.%s' available: %t", d.Name, d.Namespace, available)
	}

	// TODO clean up releases
	//  [ ] Look for deleted releases
	//  [ ] Look for releases in platform that don't exists and vice versa
	platformUpdate := false
	for envName, envRel := range envRelMap {
		if envRel.Active != nil {
			available := IsAppDeploymentAvailable(envRel.AppDepSpecActive, compReadyMap)
			if envRel.Active.AppDeploymentAvailable != available {
				platformUpdate = true
				envRel.Active.AppDeploymentAvailable = available
			}
			log.Debugf("active release '%s.%s' available: %t", envRel.Active.Name, namespace, available)
		}

		if envRel.Latest != nil {
			envRel.Latest.AppDeploymentAvailable = IsAppDeploymentAvailable(envRel.AppDepSpecLatest, compReadyMap)
			log.Debugf("latest release '%s.%s' available: %t", envRel.Latest.Name, namespace, envRel.Latest.AppDeploymentAvailable)
		}

		if envRel.Latest != nil && envRel.Latest.AppDeploymentAvailable {
			platformUpdate = true
			if envRel.PlatformEnv.SupersededReleases == nil {
				envRel.PlatformEnv.SupersededReleases = map[string]*v1alpha1.PlatformEnvRelease{}
			}

			now := metav1.Now()
			// Move all pending Releases to superseded.
			for _, r := range pendingRelMap[envName] {
				if r == envRel.Latest {
					continue
				}
				r.SupersededTime = &now
				r.AppDeploymentAvailable = false
				r.LastTransitionTime = now
				envRel.PlatformEnv.SupersededReleases[r.Name] = r
			}

			// Move active Releases to superseded.
			if envRel.Active != nil {
				envRel.Active.SupersededTime = &now
				envRel.Active.AppDeploymentAvailable = false
				envRel.Active.LastTransitionTime = now
				envRel.PlatformEnv.SupersededReleases[envRel.Active.Name] = envRel.Active
			}

			envRel.Latest.ReleaseTime = &now
			envRel.Latest.SupersededTime = nil
			envRel.Latest.LastTransitionTime = now
			envRel.PlatformEnv.Release = envRel.Latest
		}
	}
	if platformUpdate {
		if err := cm.Update(ctx, platform); err != nil {
			return false, err
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
	for name, c := range spec.Components {
		key := CompReadyKey(name, c.Commit)
		if found, available := compReadyMap[key]; !found || !available {
			return false
		}
	}
	return true
}
