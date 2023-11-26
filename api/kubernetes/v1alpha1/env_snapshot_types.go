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

type EnvSnapshotData struct {
	EnvData `json:",inline"`

	Source       EnvSource   `json:"source"`
	SnapshotTime metav1.Time `json:"snapshotTime"`
}

type EnvSource struct {
	// +kubebuilder:validation:Enum=ClusterVirtualEnv;VirtualEnv
	Kind string `json:"kind"`
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
	// +kubebuilder:validation:MinLength=1
	ResourceVersion string `json:"resourceVersion"`
}

// +kubebuilder:object:root=true
type VirtualEnvSnapshot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec    ClusterEnvSpec  `json:"spec"`
	Data    EnvSnapshotData `json:"data,omitempty"`
	Details EnvDetails      `json:"details,omitempty"`
}

// +kubebuilder:object:root=true
type VirtualEnvSnapshotList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []VirtualEnvSnapshot `json:"items"`
}

func (e *VirtualEnvSnapshot) GetData() *EnvData {
	return &e.Data.EnvData
}

func (e *VirtualEnvSnapshot) GetDetails() *EnvDetails {
	return &e.Details
}

func (e *VirtualEnvSnapshot) GetReleasePolicy() *EnvReleasePolicy {
	return e.Spec.ReleasePolicy
}

func (e *VirtualEnvSnapshot) SetReleasePolicy(p *EnvReleasePolicy) {
	e.Spec.ReleasePolicy = p
}

func (e *VirtualEnvSnapshot) GetParent() string {
	return ""
}

func (e *VirtualEnvSnapshot) GetEnvName() string {
	return e.Data.Source.Name
}

func init() {
	SchemeBuilder.Register(&VirtualEnvSnapshot{}, &VirtualEnvSnapshotList{})
}
