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

// ComponentSpecSpec defines the desired state of ComponentSpec
type ComponentSpecSpec struct {
	Selector      ComponentSpecSelector `json:"selector"`
	PodSpec       common.PodSpec        `json:"podSpec,omitempty"`
	ContainerSpec common.ContainerSpec  `json:"containerSpec,omitempty"`
	Logger        common.LoggerSpec     `json:"logger,omitempty"`
}

type ComponentSpecSelector struct {
	App    string `json:"app,omitempty"`
	Name   string `json:"name,omitempty"`
	Commit string `json:"commit,omitempty"`
}

// ComponentSpecStatus defines the observed state of ComponentSpec
type ComponentSpecStatus struct {
	// +kubebuilder:validation:Optional
	Deployments []common.Ref `json:"components"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ComponentSpec is the Schema for the ComponentSpecs API
type ComponentSpec struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ComponentSpecSpec   `json:"spec,omitempty"`
	Status ComponentSpecStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ComponentSpecList contains a list of ComponentSpecs
type ComponentSpecList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ComponentSpec `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ComponentSpec{}, &ComponentSpecList{})
}
