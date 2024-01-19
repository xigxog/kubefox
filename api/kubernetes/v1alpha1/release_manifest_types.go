/*
Copyright © 2023 XigXog

This Source Code Form is subject to the terms of the Mozilla Public License,
v2.0. If a copy of the MPL was not distributed with this file, You can obtain
one at https://mozilla.org/MPL/2.0/.
*/

package v1alpha1

import (
	"github.com/xigxog/kubefox/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ReleaseManifestSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1

	Id string `json:"id"`

	// +kubebuilder:validation:Required

	VirtualEnvironment ReleaseManifestEnv `json:"virtualEnvironment"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinProperties=1

	Apps map[string]*ReleaseManifestApp `json:"apps"`
}

type ReleaseManifestEnv struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1

	Name string `json:"name"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1

	Environment string `json:"environment"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1

	ResourceVersion string `json:"resourceVersion"`
}

type ReleaseManifestApp struct {
	Version string `json:"version,omitempty"`

	// +kubebuilder:validation:Required

	AppDeployment ReleaseManifestAppDep `json:"appDeployment"`
}

type ReleaseManifestAppDep struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1

	Name string `json:"name"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1

	ResourceVersion string `json:"resourceVersion"`

	// +kubebuilder:validation:Required

	Spec AppDeploymentSpec `json:"spec"`
}

type ReleaseManifestStatus struct {
	Problems api.Problems `json:"problems,omitempty"`

	// +patchStrategy=merge
	// +patchMergeKey=type
	// +listType=map
	// +listMapKey=type

	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=releasemanifests,shortName=manifest;rm
// +kubebuilder:printcolumn:name="Environment",type=string,JSONPath=`.spec.virtualEnvironment.name`
// +kubebuilder:printcolumn:name="VirtualEnvironment",type=string,JSONPath=`.spec.virtualEnvironment.environment`
// +kubebuilder:printcolumn:name="Available",type=string,JSONPath=`.status.conditions[?(@.type=='Available')].status`
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.conditions[?(@.type=='Available')].reason`
// +kubebuilder:printcolumn:name="Progressing",type=string,JSONPath=`.status.conditions[?(@.type=='Progressing')].status`
// +kubebuilder:printcolumn:name="Title",type=string,JSONPath=`.details.title`,priority=1
// +kubebuilder:printcolumn:name="Description",type=string,JSONPath=`.details.description`,priority=1

type ReleaseManifest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:Required

	Spec ReleaseManifestSpec `json:"spec"`

	// +kubebuilder:validation:Required

	// Data is the merged values of the Environment and VirtualEnvironment data
	// objects.
	Data    api.Data               `json:"data"`
	Details api.DataDetails        `json:"details,omitempty"`
	Status  *ReleaseManifestStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type ReleaseManifestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ReleaseManifest `json:"items"`
}

func (d *ReleaseManifest) GetData() *api.Data {
	return &d.Data
}

func (d *ReleaseManifest) GetDataKey() api.DataKey {
	return api.DataKey{
		Kind:      d.Kind,
		Name:      d.Name,
		Namespace: d.Namespace,
	}
}

func init() {
	SchemeBuilder.Register(&ReleaseManifest{}, &ReleaseManifestList{})
}
