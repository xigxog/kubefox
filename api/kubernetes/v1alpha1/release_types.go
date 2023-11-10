/*
Copyright © 2023 XigXog

This Source Code Form is subject to the terms of the Mozilla Public License,
v2.0. If a copy of the MPL was not distributed with this file, You can obtain
one at https://mozilla.org/MPL/2.0/.
*/

package v1alpha1

import (
	common "github.com/xigxog/kubefox/api/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReleaseSpec defines the desired state of Release
type ReleaseSpec struct {
	// +kubebuilder:validation:MinLength=1
	Version       string            `json:"version,omitempty"`
	Environment   ReleaseEnv        `json:"environment"`
	AppDeployment ReleaseDeployment `json:"appDeployment"`
}

type ReleaseEnv struct {
	common.Ref `json:",inline"`
	EnvSpec    `json:",inline"`
}

type ReleaseDeployment struct {
	common.Ref `json:",inline"`
	App        *App `json:"app,omitempty"`
	// +kubebuilder:validation:MinProperties=1
	Components map[string]*Component `json:"components,omitempty"`
}

// ReleaseStatus defines the observed state of Release
type ReleaseStatus struct {
	// +kubebuilder:validation:Optional
	Ready bool `json:"ready"`
}

// ReleaseDetails defines additional details of Release
type ReleaseDetails struct {
	common.Details `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:subresource:details

// Release is the Schema for the Releases API
type Release struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec    ReleaseSpec    `json:"spec,omitempty"`
	Status  ReleaseStatus  `json:"status,omitempty"`
	Details ReleaseDetails `json:"details,omitempty"`
}

//+kubebuilder:object:root=true

// ReleaseList contains a list of Releases
type ReleaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Release `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Release{}, &ReleaseList{})
}

func (r *Release) AppDeploymentSpec() *AppDeploymentSpec {
	return &AppDeploymentSpec{
		App:        *r.Spec.AppDeployment.App,
		Components: r.Spec.AppDeployment.Components,
	}
}

func (r *Release) SetAppDeploymentSpec(spec *AppDeploymentSpec) {
	r.Spec.AppDeployment.App = &spec.App
	r.Spec.AppDeployment.Components = spec.Components
}
