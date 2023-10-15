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

// DeploymentSpec defines the desired state of Deployment
type DeploymentSpec struct {
	App        App                   `json:"app"`
	Components map[string]*Component `json:"components"`
	Adapters   map[string]*Adapter   `json:"adapters,omitempty"`
}

type App struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Name        string `json:"name"`
	GitRef      string `json:"gitRef,omitempty"`
	// +kubebuilder:validation:Pattern="^[a-z0-9]{7}$"
	Commit            string `json:"commit"`
	GitRepo           string `json:"gitRepo,omitempty"`
	ContainerRegistry string `json:"containerRegistry"`
	ImagePullSecret   string `json:"imagePullSecret,omitempty"`
}

type Component struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	GitRef      string `json:"gitRef,omitempty"`
	// +kubebuilder:validation:Pattern="^[a-z0-9]{7}$"
	Commit    string    `json:"commit"`
	Image     string    `json:"image,omitempty"`
	EnvSchema EnvSchema `json:"env,omitempty"`
}

type Adapter struct {
	// +kubebuilder:validation:Enum=graphql;http;kv;object
	Type string `json:"type"`
}

// DeploymentStatus defines the observed state of Deployment
type DeploymentStatus struct {
	// Ready bool `json:"ready"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Deployment is the Schema for the deployments API
type Deployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeploymentSpec   `json:"spec,omitempty"`
	Status DeploymentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DeploymentList contains a list of Deployment
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
