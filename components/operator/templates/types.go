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
	Owner  []*metav1.OwnerReference

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

func (d Data) PlatformFullName() string {
	if d.Platform.Name == "" {
		return d.Instance.Name
	}
	if strings.HasPrefix(d.Platform.Name, d.Instance.Name) {
		return d.Platform.Name
	}
	return fmt.Sprintf("%s-%s", d.Instance.Name, d.Platform.Name)
}

func (d Data) ComponentName() string {
	prefix := d.App.Name
	if prefix == "" {
		prefix = d.Platform.Name
	}
	if prefix == "" {
		prefix = d.Instance.Name
	}
	name := fmt.Sprintf("%s-%s-%s", prefix, d.Component.Name, d.Component.Commit)
	return strings.TrimSuffix(name, "-")
}
