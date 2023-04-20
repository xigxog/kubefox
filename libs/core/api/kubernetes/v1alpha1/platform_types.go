package v1alpha1

import (
	"github.com/xigxog/kubefox/libs/core/api/common"
	"github.com/xigxog/kubefox/libs/core/api/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=platforms,scope=Cluster

type Platform struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   common.PlatformSpec   `json:"spec,omitempty"`
	Status common.PlatformStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type PlatformList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Platform `json:"items"`
}

func (obj *Platform) GetSpec() any {
	return &obj.Spec
}

func init() {
	SchemeBuilder.Register(&Platform{}, &PlatformList{})
}

func (r *Platform) String() string {
	return kubernetes.FullKey(r)
}
