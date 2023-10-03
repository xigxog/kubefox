package templates

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Data struct {
	Instance  Instance
	Platform  Platform
	App       App
	Component Component

	Labels map[string]string
	Owner  *metav1.OwnerReference

	Values map[string]any
}

type Instance struct {
	Name      string
	Namespace string
	RootCA    string
}

type Platform struct {
	Name      string
	Namespace string
}

type App struct {
	Name            string
	Commit          string
	GitRef          string
	Registry        string
	ImagePullSecret string
}

type Component struct {
	Name            string
	Commit          string
	GitRef          string
	Image           string
	ImagePullPolicy string
	Resources       corev1.ResourceRequirements
}

type ResourceList struct {
	Items []*unstructured.Unstructured `json:"items,omitempty"`
}

func (d Data) Namespace() string {
	if d.Platform.Namespace != "" {
		return d.Platform.Namespace
	}
	return d.Instance.Namespace
}

func (d Data) InstanceName() string {
	return d.Instance.Name
}

func (d Data) PlatformName() string {
	if d.Platform.Name == "" {
		return d.InstanceName()
	}
	if strings.HasPrefix(d.Platform.Name, d.InstanceName()) {
		return d.Platform.Name
	}
	return fmt.Sprintf("%s-%s", d.InstanceName(), d.Platform.Name)
}

func (d Data) AppName() string {
	if d.App.Name == "" {
		return d.PlatformName()
	}
	if strings.HasPrefix(d.App.Name, d.PlatformName()) {
		return d.App.Name
	}
	return fmt.Sprintf("%s-%s", d.PlatformName(), d.App.Name)
}

func (d Data) ComponentName() string {
	n := fmt.Sprintf("%s-%s-%s", d.AppName(), d.Component.Name, d.Component.Commit)
	return strings.TrimSuffix(n, "-")
}
