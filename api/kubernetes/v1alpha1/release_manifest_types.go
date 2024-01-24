// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package v1alpha1

import (
	"github.com/xigxog/kubefox/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type ReleaseManifestSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1

	ReleaseId string `json:"releaseId"`

	// +kubebuilder:validation:Required

	Environment ReleaseManifestRef `json:"environment"`

	// +kubebuilder:validation:Required

	VirtualEnvironment ReleaseManifestRef `json:"virtualEnvironment"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinProperties=1

	Apps map[string]*ReleaseManifestApp `json:"apps"`
}

type ReleaseManifestRef struct {
	// +kubebuilder:validation:Required

	UID types.UID `json:"uid"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1

	Name string `json:"name"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1

	ResourceVersion string `json:"resourceVersion"`
}

type ReleaseManifestApp struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Version string `json:"version"`

	// +kubebuilder:validation:Required

	AppDeployment ReleaseManifestAppDep `json:"appDeployment"`
}

type ReleaseManifestAppDep struct {
	ReleaseManifestRef `json:",inline"`

	// +kubebuilder:validation:Required

	Spec AppDeploymentSpec `json:"spec"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=releasemanifests,shortName=manifest;rm
// +kubebuilder:printcolumn:name="Id",type=string,JSONPath=`.spec.releaseId`
// +kubebuilder:printcolumn:name="Environment",type=string,JSONPath=`.spec.environment.name`
// +kubebuilder:printcolumn:name="VirtualEnvironment",type=string,JSONPath=`.spec.virtualEnvironment.name`

type ReleaseManifest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:Required

	Spec ReleaseManifestSpec `json:"spec"`

	// +kubebuilder:validation:Required

	// Data is the merged values of the Environment and VirtualEnvironment data
	// objects.
	Data    api.Data        `json:"data"`
	Details api.DataDetails `json:"details,omitempty"`
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
