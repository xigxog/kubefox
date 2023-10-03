package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/xigxog/kubefox/components/operator/templates"
	"github.com/xigxog/kubefox/libs/core/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/libs/core/logkf"
)

type TemplateData struct {
	templates.Data

	Template string
	Obj      client.Object
}

type ComponentManager struct {
	Instance string
	Client   *Client
	Log      *logkf.Logger
}

func (cm *ComponentManager) SetupComponent(ctx context.Context, td *TemplateData) (bool, error) {
	log := cm.Log.With(
		logkf.Instance, td.Instance.Name,
		logkf.Platform, td.Platform.Name,
		logkf.ComponentName, td.Component.Name,
	)

	log.Debug("setting up component")

	name := nn(td.Namespace(), td.ComponentName())
	if err := cm.Client.Get(ctx, name, td.Obj); client.IgnoreNotFound(err) != nil {
		return false, log.ErrorN("unable to fetch component workload: %w", err)
	}

	var available int32
	switch obj := td.Obj.(type) {
	case *appsv1.StatefulSet:
		available = obj.Status.AvailableReplicas

	case *appsv1.Deployment:
		available = obj.Status.AvailableReplicas

	case *appsv1.DaemonSet:
		available = obj.Status.NumberAvailable
	}
	if available <= 0 {
		log.Debug("component is not ready, applying template to ensure correct state")
		return false, cm.Client.ApplyTemplate(ctx, td.Template, &td.Data)
	}

	log.Debug("component ready")
	return true, nil
}

func (cm *ComponentManager) ReconcileComponents(ctx context.Context, namespace string) error {
	platform, err := cm.Client.GetPlatform(ctx, namespace)
	if err != nil {
		return err
	}

	log := cm.Log.With(
		logkf.Instance, cm.Instance,
		logkf.Platform, platform.Name,
	)

	relList := &v1alpha1.ReleaseList{}
	if err := cm.Client.List(ctx, relList, client.InNamespace(platform.Namespace)); err != nil {
		return err
	}
	depList := &v1alpha1.DeploymentList{}
	if err := cm.Client.List(ctx, depList, client.InNamespace(platform.Namespace)); err != nil {
		return err
	}

	specs := make([]v1alpha1.DeploymentSpec, 0, len(relList.Items)+len(depList.Items))
	for _, r := range relList.Items {
		specs = append(specs, r.Spec.Deployment)
	}
	for _, d := range depList.Items {
		specs = append(specs, d.Spec)
	}
	log.Debugf("found %d releases and %d deployments", len(relList.Items), len(depList.Items))

	compDepList := &appsv1.DeploymentList{}
	if err := cm.Client.List(ctx, compDepList,
		client.InNamespace(platform.Namespace),
		client.HasLabels{LabelComponent},
	); err != nil {
		return err
	}

	td := &TemplateData{
		Data: templates.Data{
			Instance: templates.Instance{
				Name: cm.Instance,
			},
			Platform: templates.Platform{
				Name:      platform.Name,
				Namespace: platform.Namespace,
			},
			Owner: metav1.NewControllerRef(platform, platform.GroupVersionKind()),
		},
		Template: "component",
	}
	compMap := make(map[string]TemplateData)
	for _, d := range specs {
		td.App = templates.App{
			Name:     d.App.Name,
			Commit:   d.App.Commit,
			GitRef:   d.App.GitRef,
			Registry: d.App.Registry,
		}
		for n, c := range d.Components {
			td.Obj = &appsv1.Deployment{}
			td.Component = templates.Component{
				Name:   n,
				Commit: c.Commit,
				GitRef: c.GitRef,
				Image:  c.Image,
			}
			compMap[td.ComponentName()] = *td
		}
	}
	log.Debugf("found %d unique components", len(compMap))

	for _, d := range compDepList.Items {
		if _, found := compMap[d.Name]; !found {
			log.Debugw("deleting component", logkf.ComponentName, d.Name)
			if err := cm.Client.Delete(ctx, &d); err != nil {
				return err
			}
		}
	}

	for _, compTD := range compMap {
		if _, err := cm.SetupComponent(ctx, &compTD); err != nil {
			return err
		}
	}

	return nil
}