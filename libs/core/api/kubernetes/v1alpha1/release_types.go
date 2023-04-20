package v1alpha1

import (
	"github.com/xigxog/kubefox/libs/core/api/common"

	"github.com/xigxog/kubefox/libs/core/api/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ReleaseSpec struct {
	common.EnvironmentProp   `json:",inline"`
	common.EnvironmentIdProp `json:",inline"`
	common.SystemProp        `json:",inline"`
	common.SystemIdProp      `json:",inline"`

	Components []*ReleaseComponent `json:"components,omitempty" validate:"dive"`
}

type ReleaseComponent struct {
	common.ComponentProps `json:",inline"`

	App    string          `json:"app"`
	Routes []*common.Route `json:"routes,omitempty" validate:"dive"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type Release struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReleaseSpec          `json:"spec,omitempty"`
	Status common.ReleaseStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type ReleaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Release `json:"items" validate:"dive"`
}

func (obj *Release) GetSpec() any {
	return &obj.Spec
}

func init() {
	SchemeBuilder.Register(&Release{}, &ReleaseList{})
}

func (r *Release) String() string {
	return kubernetes.FullKey(r)
}
