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

type ClusterEnvSpec struct {
	ReleasePolicy *EnvReleasePolicy `json:"releasePolicy,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=clustervirtualenvs,scope=Cluster
type ClusterVirtualEnv struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec    ClusterEnvSpec `json:"spec"`
	Data    EnvData        `json:"data,omitempty"`
	Details EnvDetails     `json:"details,omitempty"`
}

// +kubebuilder:object:root=true
type ClusterVirtualEnvList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ClusterVirtualEnv `json:"items"`
}

func (e *ClusterVirtualEnv) GetData() *EnvData {
	return &e.Data
}

func (e *ClusterVirtualEnv) GetDetails() *EnvDetails {
	return &e.Details
}

func (e *ClusterVirtualEnv) GetReleasePolicy() *EnvReleasePolicy {
	return e.Spec.ReleasePolicy
}

func (e *ClusterVirtualEnv) SetReleasePolicy(p *EnvReleasePolicy) {
	e.Spec.ReleasePolicy = p
}

func (e *ClusterVirtualEnv) GetParent() string {
	return ""
}

func (e *ClusterVirtualEnv) GetEnvName() string {
	return e.Name
}

func init() {
	SchemeBuilder.Register(&ClusterVirtualEnv{}, &ClusterVirtualEnvList{})
}
