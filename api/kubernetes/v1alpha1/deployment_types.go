/*
Copyright © 2023 XigXog

This Source Code Form is subject to the terms of the Mozilla Public License,
v2.0. If a copy of the MPL was not distributed with this file, You can obtain
one at https://mozilla.org/MPL/2.0/.
*/

package v1alpha1

import (
	kubefox "github.com/xigxog/kubefox/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeploymentSpec defines the desired state of Deployment
type DeploymentSpec struct {
	App        App                   `json:"app"`
	Components map[string]*Component `json:"components"`
}

type App struct {
	kubefox.App `json:",inline"`

	Branch string `json:"branch,omitempty"`
	Tag    string `json:"tag,omitempty"`
	// +kubebuilder:validation:Pattern="^[a-z0-9]{40}$"
	Commit              string `json:"commit"`
	RepoURL             string `json:"repoURL,omitempty"`
	ImagePullSecretName string `json:"imagePullSecretName,omitempty"`
}

type Component struct {
	kubefox.ComponentSpec `json:",inline"`

	// +kubebuilder:validation:Pattern="^[a-z0-9]{40}$"
	Commit string `json:"commit"`
	Image  string `json:"image,omitempty"`
}

// DeploymentStatus defines the observed state of Deployment
type DeploymentStatus struct {
	// +kubebuilder:validation:Optional
	Ready bool `json:"ready"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Deployment is the Schema for the Deployments API
type Deployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeploymentSpec   `json:"spec,omitempty"`
	Status DeploymentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DeploymentList contains a list of Deployments
type DeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Deployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Deployment{}, &DeploymentList{})
}

func (d *Deployment) GetSpec() DeploymentSpec {
	return d.Spec
}
