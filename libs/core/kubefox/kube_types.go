package kubefox

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Hook uint8

const (
	Unknown Hook = iota
	Customize
	Sync
)

type RequestStub struct {
	Object *ObjectStub `json:"object"` // for sync requests
	Parent *ObjectStub `json:"parent"` // for customize requests
}

type ObjectStub struct {
	Kind string `json:"kind"`
}

type KubeResponse struct {
	Attachments []runtime.Object `json:"attachments"`
	Status      any              `json:"status,omitempty"`
}

type CustomizeResponse struct {
	RelatedResourceRules []*RelatedResourceRule `json:"relatedResources,omitempty"`
}

type RelatedResourceRule struct {
	APIVersion    string                `json:"apiVersion"`
	Resource      string                `json:"resource"`
	Namespace     string                `json:"namespace,omitempty"`
	Names         []string              `json:"names,omitempty"`
	LabelSelector *metav1.LabelSelector `json:"labelSelector,omitempty"`
}

func (h Hook) String() string {
	switch h {
	case Customize:
		return "customize"
	case Sync:
		return "sync"
	}
	return "unknown"
}
