package v1alpha1

import (
	"github.com/xigxog/kubefox/libs/core/api/common"
	"github.com/xigxog/kubefox/libs/core/api/kubernetes"
	"github.com/xigxog/kubefox/libs/core/api/uri"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ComponentSetSpec struct {
	Deployments map[uri.Key]*Deployment `json:"deployments,omitempty" validate:"dive"`
}

type Deployment struct {
	Components []*common.ComponentProps `json:"components,omitempty" validate:"dive"`
}

type ComponentSetStatus struct {
	Components  map[common.ComponentKey]*common.ComponentStatus `json:"components,omitempty" validate:"dive"`
	Deployments map[uri.Key]*common.DeploymentStatus            `json:"deployments,omitempty" validate:"dive"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type ComponentSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ComponentSetSpec   `json:"spec,omitempty"`
	Status ComponentSetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type ComponentSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ComponentSet `json:"items" validate:"dive"`
}

func (obj *ComponentSet) GetSpec() any {
	return &obj.Spec
}

func init() {
	SchemeBuilder.Register(&ComponentSet{}, &ComponentSetList{})
}

func (r *ComponentSet) String() string {
	return kubernetes.FullKey(r)
}

func (r *ComponentSetList) String() string {
	return kubernetes.KindKey(r)
}
