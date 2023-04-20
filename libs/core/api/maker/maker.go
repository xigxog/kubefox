package maker

import (
	"reflect"
	"strings"

	"github.com/xigxog/kubefox/libs/core/api/admin"
	"github.com/xigxog/kubefox/libs/core/api/admin/v1alpha1"
	"github.com/xigxog/kubefox/libs/core/api/kubernetes"
	"github.com/xigxog/kubefox/libs/core/api/uri"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	KubernetesGroup = "k8s.kubefox.io"
	AdminAPIVersion = "admin.kubefox.io/v1alpha1"
)

type Props struct {
	Name string

	// Organization string
	Platform string

	Environment    string
	EnvironmentId  string
	EnvironmentRef string
	Config         string
	ConfigId       string
	ConfigRef      string
	System         string
	SystemId       string
	SystemRef      string
	Component      string
	CompHash       string

	Namespace string
	Instance  string

	Group   string
	Version string
	Kind    string
}

func Empty[T any]() *T {
	return New[T](Props{})
}

func New[T any](props Props) *T {
	obj := new(T)
	objI := any(obj)
	typ := reflect.TypeOf(obj).Elem()

	if r, ok := objI.(admin.Object); ok {
		r.SetKind(uri.KindFromString(typ.Name()))
		r.SetAPIVersion(AdminAPIVersion)
		r.SetName(props.Name)
	}

	if r, ok := objI.(admin.SubObject); ok {
		r.SetKind(uri.SubKindFromString(typ.Name()))
		r.SetAPIVersion(AdminAPIVersion)
	}

	if o, ok := objI.(runtime.Object); ok {
		pkg := typ.PkgPath()
		pkgParts := strings.Split(pkg, "/")

		if props.Group == "core" {
			props.Group = ""
		} else if props.Group == "" && strings.Contains(pkg, "kubefox") {
			props.Group = KubernetesGroup
		}
		if props.Version == "" {
			props.Version = pkgParts[len(pkgParts)-1]
		}
		if props.Kind == "" {
			props.Kind = typ.Name()
		}

		o.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{
			Group:   props.Group,
			Version: props.Version,
			Kind:    props.Kind,
		})
	}

	if o, ok := objI.(kubernetes.Object); ok {
		o.SetName(props.Name)
		o.SetNamespace(props.Namespace)
	}

	if o, ok := objI.(metav1.Object); ok {
		o.SetLabels(Labels(props))
	}

	return obj
}

// TODO add recommended k8s labels
// https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/

// Labels generates a map of labels for any non-empty properties.
func Labels(props Props) map[string]string {
	labels := make(map[string]string)
	// addLabel(labels, kubernetes.OrganizationLabel, props.Organization)
	addLabel(labels, kubernetes.PlatformLabel, props.Platform)
	addLabel(labels, kubernetes.EnvironmentLabel, props.Environment)
	addLabel(labels, kubernetes.EnvRefLabel, props.EnvironmentRef)
	addLabel(labels, kubernetes.EnvIdLabel, props.EnvironmentId)
	addLabel(labels, kubernetes.ConfigLabel, props.Config)
	addLabel(labels, kubernetes.ConfigRefLabel, props.ConfigRef)
	addLabel(labels, kubernetes.ConfigIdLabel, props.ConfigId)
	addLabel(labels, kubernetes.SystemLabel, props.System)
	addLabel(labels, kubernetes.SystemRefLabel, props.SystemRef)
	addLabel(labels, kubernetes.SystemIdLabel, props.SystemId)
	addLabel(labels, kubernetes.ComponentLabel, props.Component)
	addLabel(labels, kubernetes.CompHashLabel, props.CompHash)
	addLabel(labels, kubernetes.InstanceLabel, props.Instance)

	return labels
}

func addLabel(labels map[string]string, key string, value string) {
	if value != "" {
		labels[key] = value
	}
}

func ObjectFromURI(u uri.URI) admin.Object {
	var obj admin.Object

	switch u.Kind() {
	case uri.Config:
		obj = New[v1alpha1.Config](Props{Name: u.Name()})
	case uri.Environment:
		obj = New[v1alpha1.Environment](Props{Name: u.Name()})
	case uri.System:
		obj = New[v1alpha1.System](Props{Name: u.Name()})
	case uri.Platform:
		obj = New[v1alpha1.Platform](Props{Name: u.Name()})
	}

	if obj == nil {
		return nil
	}

	if u.SubKind() == uri.Id {
		obj.SetId(u.SubPath())
	}

	return obj
}

func SubObjFromURI(u uri.URI) admin.SubObject {
	switch u.SubKind() {
	case uri.Deployment:
		r := Empty[v1alpha1.Deployment]()
		r.System = u.SubPath()
		return r
	case uri.Release:
		return Empty[v1alpha1.Release]()
	}

	return nil
}
