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
	Name           string
	Namespace      string
	RootCA         string
	LogLevel       string
	LogFormat      string
	BootstrapImage string
}

type Platform struct {
	Name      string
	Namespace string
	LogLevel  string
	LogFormat string
}

type App struct {
	Name            string
	Commit          string
	GitRef          string
	Registry        string
	ImagePullSecret string
	Resources       *corev1.ResourceRequirements
	NodeSelector    *corev1.NodeSelector
	Tolerations     []*corev1.Toleration
	Affinity        *corev1.Affinity
	LogLevel        string
	LogFormat       string
}

type Component struct {
	Name            string
	Commit          string
	GitRef          string
	Image           string
	ImagePullPolicy string
	Resources       *corev1.ResourceRequirements
	NodeSelector    *corev1.NodeSelector
	Tolerations     []*corev1.Toleration
	Affinity        *corev1.Affinity
	LogLevel        string
	LogFormat       string
}

type ResourceList struct {
	Items []*unstructured.Unstructured `json:"items,omitempty"`
}

func (d Data) Name() string {
	if d.Platform.Name == "" {
		return d.Instance.Name
	}
	return d.Platform.Name
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

func (d Data) PlatformVaultName() string {
	name := fmt.Sprintf("%s-%s", d.Platform.Namespace, d.Platform.Name)
	if !strings.HasPrefix(name, "kubefox") {
		name = "kubefox-" + name
	}
	return name
}

func (d Data) ComponentFullName() string {
	if d.Component.Name == "" {
		return ""
	}

	name := d.App.Name
	if name == "" {
		name = d.Platform.Name
	}
	if name == "" {
		name = d.Instance.Name
	}
	name = fmt.Sprintf("%s-%s-%s", name, d.Component.Name, d.Component.Commit)
	return strings.TrimSuffix(name, "-")
}

func (d Data) LogLevel() string {
	if d.Component.LogLevel != "" {
		return d.Component.LogLevel
	}
	if d.App.LogLevel != "" {
		return d.App.LogLevel
	}
	if d.Platform.LogLevel != "" {
		return d.Platform.LogLevel
	}
	if d.Instance.LogLevel != "" {
		return d.Instance.LogLevel
	}
	return "info"
}

func (d Data) LogFormat() string {
	if d.Component.LogFormat != "" {
		return d.Component.LogFormat
	}
	if d.App.LogFormat != "" {
		return d.App.LogFormat
	}
	if d.Platform.LogFormat != "" {
		return d.Platform.LogFormat
	}
	if d.Instance.LogFormat != "" {
		return d.Instance.LogFormat
	}
	return "json"
}

func (d Data) HomePath() string {
	return "/tmp/kubefox"
}
