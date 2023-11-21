/*
Copyright © 2023 XigXog

This Source Code Form is subject to the terms of the Mozilla Public License,
v2.0. If a copy of the MPL was not distributed with this file, You can obtain
one at https://mozilla.org/MPL/2.0/.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:data
//+kubebuilder:subresource:details

// ResolvedEnvironment is the Schema for the ResolvedEnvironments API
type ResolvedEnvironment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Data    EnvData    `json:"data,omitempty"`
	Details EnvDetails `json:"details,omitempty"`
}

//+kubebuilder:object:root=true

// ResolvedEnvironmentList contains a list of ResolvedEnvironments
type ResolvedEnvironmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ResolvedEnvironment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ResolvedEnvironment{}, &ResolvedEnvironmentList{})
}
