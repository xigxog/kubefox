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

// AppDeploymentSpec defines the desired state of AppDeployment
type AppDeploymentSpec struct {
	App App `json:"app"`
	// +kubebuilder:validation:MinProperties=1
	Components map[string]*Component `json:"components"`
}

type App struct {
	common.App `json:",inline"`

	// +kubebuilder:validation:Pattern="^[a-z0-9]{40}$"
	Commit              string `json:"commit"`
	ImagePullSecretName string `json:"imagePullSecretName,omitempty"`
}

type Component struct {
	common.ComponentSpec `json:",inline"`

	// +kubebuilder:validation:Pattern="^[a-z0-9]{40}$"
	Commit string `json:"commit"`
	Image  string `json:"image,omitempty"`
}

// AppDeploymentStatus defines the observed state of AppDeployment
type AppDeploymentStatus struct {
	// +kubebuilder:validation:Optional
	Ready bool `json:"ready"`
}

// AppDeploymentDetails defines additional details of AppDeployment
type AppDeploymentDetails struct {
	App        AppDetails                `json:"app,omitempty"`
	Components map[string]common.Details `json:"components,omitempty"`
}

type AppDetails struct {
	common.Details `json:",inline"`

	Branch string `json:"branch,omitempty"`
	Tag    string `json:"tag,omitempty"`
	// +kubebuilder:validation:Format=uri
	RepoURL string `json:"repoURL,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:subresource:details

// AppDeployment is the Schema for the AppDeployments API
type AppDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec    AppDeploymentSpec    `json:"spec,omitempty"`
	Status  AppDeploymentStatus  `json:"status,omitempty"`
	Details AppDeploymentDetails `json:"details,omitempty"`
}

//+kubebuilder:object:root=true

// AppDeploymentList contains a list of AppDeployments
type AppDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []AppDeployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AppDeployment{}, &AppDeploymentList{})
}
