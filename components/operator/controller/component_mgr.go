package controller

import (
	"context"
	"errors"

	"golang.org/x/mod/semver"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/build"
	"github.com/xigxog/kubefox/components/operator/templates"
	kubefox "github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/logkf"
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
		logkf.KeyInstance, td.Instance.Name,
		logkf.KeyPlatform, td.Platform.Name,
		logkf.KeyComponentName, td.Component.Name,
	)

	log.Debug("setting up component")

	name := nn(td.Namespace(), td.ComponentFullName())
	if err := cm.Client.Get(ctx, name, td.Obj); client.IgnoreNotFound(err) != nil {
		return false, log.ErrorN("unable to fetch component workload: %w", err)
	}

	ver := td.Obj.GetLabels()[kubefox.LabelK8sKubeFoxVersion]

	if semver.Compare(ver, build.Info.Version) < 0 {
		log.Infof("version upgrade detected, applying template to upgrade %s->%s", ver, build.Info.Version)
		return false, cm.Client.ApplyTemplate(ctx, td.Template, &td.Data)
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

func (cm *ComponentManager) ReconcileComponents(ctx context.Context, namespace string) (bool, error) {
	platform, err := cm.Client.GetPlatform(ctx, namespace)
	if err != nil {
		return false, IgnoreNotFound(err)
	}
	if !platform.Status.Ready {
		return false, nil
	}

	log := cm.Log.With(
		logkf.KeyInstance, cm.Instance,
		logkf.KeyPlatform, platform.Name,
	)

	if !platform.Status.Ready {
		log.Debug("platform not ready")
		return false, nil
	}

	relList := &v1alpha1.ReleaseList{}
	if err := cm.Client.List(ctx, relList, client.InNamespace(platform.Namespace)); err != nil {
		return false, err
	}
	depList := &v1alpha1.DeploymentList{}
	if err := cm.Client.List(ctx, depList, client.InNamespace(platform.Namespace)); err != nil {
		return false, err
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
		client.HasLabels{kubefox.LabelK8sComponent},
	); err != nil {
		return false, err
	}

	compMap := make(map[string]TemplateData)
	for _, d := range specs {
		for n, c := range d.Components {
			td := TemplateData{
				Data: templates.Data{
					Instance: templates.Instance{
						Name:           cm.Instance,
						BootstrapImage: BootstrapImage,
						Version:        build.Info.Version,
					},
					Platform: templates.Platform{
						Name:      platform.Name,
						Namespace: platform.Namespace,
					},
					App: templates.App{
						Name:            d.App.Name,
						Commit:          d.App.Commit,
						Branch:          d.App.Branch,
						Tag:             d.App.Tag,
						Registry:        d.App.ContainerRegistry,
						ImagePullSecret: d.App.ImagePullSecret,
					},
					Component: templates.Component{
						Name:   n,
						Commit: c.Commit,
						Image:  c.Image,
					},
					Owner: []*metav1.OwnerReference{metav1.NewControllerRef(platform, platform.GroupVersionKind())},
				},
				Template: "component",
				Obj:      &appsv1.Deployment{},
			}
			compMap[td.ComponentFullName()] = td
		}
	}
	log.Debugf("found %d unique components", len(compMap))

	for _, d := range compDepList.Items {
		if _, found := compMap[d.Name]; !found {
			log.Debugw("deleting component", logkf.KeyComponentName, d.Name)
			if err := cm.Client.Delete(ctx, &d); err != nil {
				return false, err
			}
			// Clean up ServiceAccount.
			if err := cm.Client.Delete(ctx, &v1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: d.Namespace,
					Name:      d.Name,
				},
			}); err != nil {
				log.Debug(err)
			}
		}
	}

	rdy := true
	for _, compTD := range compMap {
		if r, err := cm.SetupComponent(ctx, &compTD); err != nil {
			return false, err
		} else {
			rdy = rdy && r
		}
	}

	log.Debugf("components reconciled")

	return rdy, nil
}

func IgnoreNotFound(err error) error {
	if apierrors.IsNotFound(err) || errors.Is(err, ErrNotFound) {
		return nil
	}
	return err
}
