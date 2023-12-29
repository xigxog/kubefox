package templates

import (
	"fmt"
	"strings"

	"github.com/xigxog/kubefox/api"
	common "github.com/xigxog/kubefox/api/kubernetes"
	"github.com/xigxog/kubefox/build"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Data struct {
	Instance  Instance
	Platform  Platform
	Component Component

	Owner []*metav1.OwnerReference

	Values map[string]any

	Logger    common.LoggerSpec
	BuildInfo build.BuildInfo `hash:"ignore"`

	Hash string `hash:"ignore"`
}

type Instance struct {
	Name           string
	Namespace      string
	RootCA         string
	BootstrapImage string
}

type Platform struct {
	Name      string
	Namespace string
}

type Component struct {
	common.PodSpec
	common.ContainerSpec

	Name                string
	Commit              string
	Type                api.ComponentType
	App                 string
	AppCommit           string
	Image               string
	ImagePullPolicy     string
	ImagePullSecret     string
	IsPlatformComponent bool
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

	commit := d.Component.Commit
	if len(d.Component.Commit) > 7 {
		commit = commit[0:7]
	}
	if d.Component.IsPlatformComponent {
		commit = ""
	}

	name := d.Component.App
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
	if d.Component.App == "" {
		return fmt.Sprintf("%s-%s", d.PlatformVaultName(), d.Component.Name)
	} else {
		return fmt.Sprintf("%s-%s-%s", d.PlatformVaultName(), d.Component.App, d.Component.Name)
	}
}

func (d Data) HomePath() string {
	return "/tmp/kubefox"
}
