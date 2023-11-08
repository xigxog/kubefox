package templates

import (
	"fmt"
	"strings"

	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/build"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// TODO replace with types from API
type Data struct {
	Instance  Instance
	Platform  Platform
	App       App
	Component Component

	ExtraLabels map[string]string
	Owner       []*metav1.OwnerReference

	Values map[string]any

	BuildInfo build.BuildInfo
}

type Instance struct {
	Name           string
	Namespace      string
	RootCA         string
	LogLevel       string
	LogFormat      string
	BootstrapImage string
	Version        string
}

type Platform struct {
	Name      string
	Namespace string
	LogLevel  string
	LogFormat string
}

type App struct {
	v1alpha1.PodSpec

	Name            string
	Commit          string
	Branch          string
	Tag             string
	Registry        string
	ImagePullSecret string
	LogLevel        string
	LogFormat       string
}

type Component struct {
	v1alpha1.PodSpec
	v1alpha1.ContainerSpec

	Name            string
	Commit          string
	Image           string
	ImagePullPolicy string
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

func (d Data) AppVersion() string {
	if p := strings.Split(d.App.Tag, "/"); d.App.Tag != "" {
		return p[len(p)-1]
	}
	if p := strings.Split(d.App.Branch, "/"); d.App.Branch != "" {
		return p[len(p)-1]
	}

	return d.App.Commit
}

func (d Data) ComponentFullName() string {
	if d.Component.Name == "" {
		return ""
	}

	commit := d.Component.Commit
	if len(d.Component.Commit) > 7 {
		commit = commit[0:7]
	}

	name := d.App.Name
	if name == "" {
		name = d.Platform.Name
	}
	if name == "" {
		name = d.Instance.Name
	}
	name = fmt.Sprintf("%s-%s-%s", name, d.Component.Name, commit)
	return strings.TrimSuffix(name, "-")
}

func (d Data) ComponentVaultName() string {
	if d.App.Name == "" {
		return fmt.Sprintf("%s-%s", d.PlatformVaultName(), d.Component.Name)
	} else {
		return fmt.Sprintf("%s-%s-%s", d.PlatformVaultName(), d.App.Name, d.Component.Name)
	}
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
