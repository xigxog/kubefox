package templates

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type BrokerType string

type Data struct {
	DevMode bool

	Platform  Platform
	System    System
	Component Component

	Environment Environment
	Config      Config
	Broker      Broker

	Values map[string]any
	Labels map[string]string
	Owner  *metav1.OwnerReference
}

type Platform struct {
	Name      string
	Version   string
	Namespace string
	RootCA    string
}

type System struct {
	Name              string
	Id                string
	Ref               string
	GitHash           string
	GitRef            string
	Namespace         string
	ContainerRegistry string
	ImagePullSecret   string
}

type Broker struct {
	Type            BrokerType
	Image           string
	ImagePullPolicy string
	Resources       corev1.ResourceRequirements
}

type Component struct {
	Name            string
	GitHash         string
	Image           string
	ImagePullPolicy string
	Resources       corev1.ResourceRequirements
}

type Environment struct {
	Name string
	Id   string
	Ref  string
}

type Config struct {
	Name string
	Id   string
	Ref  string
}

type ResourceList struct {
	Items []*unstructured.Unstructured `json:"items,omitempty"`
}
