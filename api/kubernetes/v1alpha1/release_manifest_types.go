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
	common "github.com/xigxog/kubefox/api/kubernetes"
	"github.com/xigxog/kubefox/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ReleaseManifestSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1

	ReleaseId string `json:"releaseId"`

	// +kubebuilder:validation:Required

	Environment EnvironmentManifest `json:"environment"`

	// +kubebuilder:validation:Required

	VirtualEnvironment VirtualEnvironmentManifest `json:"virtualEnvironment"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1

	AppDeployments []AppDeploymentManifest `json:"appDeployments"`

	Adapters *ReleaseManifestAdapters `json:"adapters,omitempty"`
}

type EnvironmentManifest struct {
	metav1.TypeMeta `json:",inline"`

	// +kubebuilder:validation:Required

	common.ObjectRef `json:"metadata"`

	Spec    EnvironmentSpec `json:"spec,omitempty"`
	Data    api.Data        `json:"data,omitempty"`
	Details api.DataDetails `json:"details,omitempty"`
}

type VirtualEnvironmentManifest struct {
	metav1.TypeMeta `json:",inline"`

	// +kubebuilder:validation:Required

	common.ObjectRef `json:"metadata"`

	Spec    VirtualEnvironmentSpec `json:"spec,omitempty"`
	Data    api.Data               `json:"data,omitempty"`
	Details api.DataDetails        `json:"details,omitempty"`
}

type AppDeploymentManifest struct {
	metav1.TypeMeta `json:",inline"`

	// +kubebuilder:validation:Required

	common.ObjectRef `json:"metadata"`

	// +kubebuilder:validation:Required

	Spec    AppDeploymentSpec    `json:"spec"`
	Details AppDeploymentDetails `json:"details,omitempty"`
}

type ReleaseManifestAdapters struct {
	HTTPAdapters []HTTPAdapterManifest `json:"http,omitempty"`
}

type HTTPAdapterManifest struct {
	metav1.TypeMeta `json:",inline"`

	// +kubebuilder:validation:Required

	common.ObjectRef `json:"metadata"`

	// +kubebuilder:validation:Required

	Spec    HTTPAdapterSpec `json:"spec"`
	Details api.Details     `json:"details,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=releasemanifests,shortName=manifest;manifests;rm;rms
// +kubebuilder:printcolumn:name="Release Id",type=string,JSONPath=`.spec.releaseId`
// +kubebuilder:printcolumn:name="Environment",type=string,JSONPath=`.spec.environment.metadata.name`
// +kubebuilder:printcolumn:name="VirtualEnvironment",type=string,JSONPath=`.spec.virtualEnvironment.metadata.name`

type ReleaseManifest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:Required

	Spec ReleaseManifestSpec `json:"spec"`

	// +kubebuilder:validation:Required

	// Data is the merged values of the Environment and VirtualEnvironment Data.
	Data api.Data `json:"data"`
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

func (d *ReleaseManifest) GetAppDeployment(name string) (*AppDeployment, error) {
	if d.Spec.AppDeployments == nil {
		return nil, core.ErrNotFound()
	}

	for _, a := range d.Spec.AppDeployments {
		if a.Name == name {
			return &AppDeployment{
				TypeMeta:   a.TypeMeta,
				ObjectMeta: a.ObjectMeta(),
				Spec:       a.Spec,
				Details:    a.Details,
			}, nil
		}
	}

	return nil, core.ErrNotFound()
}

func (d *ReleaseManifest) AddAppDeployment(appDep *AppDeployment) {
	if cur, _ := d.GetAppDeployment(appDep.Name); cur != nil {
		// Adapter already present.
		return
	}

	d.Spec.AppDeployments = append(d.Spec.AppDeployments, AppDeploymentManifest{
		TypeMeta:  appDep.TypeMeta,
		ObjectRef: common.RefFromMeta(appDep.ObjectMeta),
		Spec:      appDep.Spec,
		Details:   appDep.Details,
	})
}

func (d *ReleaseManifest) GetAdapter(name string, typ api.ComponentType) (common.Adapter, error) {
	if d.Spec.Adapters == nil {
		return nil, core.ErrNotFound()
	}

	switch typ {
	case api.ComponentTypeHTTPAdapter:
		for _, h := range d.Spec.Adapters.HTTPAdapters {
			if h.Name == name {
				return &HTTPAdapter{
					ObjectMeta: h.ObjectMeta(),
					Spec:       h.Spec,
				}, nil
			}
		}
	}

	return nil, core.ErrNotFound()
}

func (d *ReleaseManifest) AddAdapter(adapter common.Adapter) {
	if cur, _ := d.GetAdapter(adapter.GetName(), adapter.GetComponentType()); cur != nil {
		// Adapter already present.
		return
	}

	if d.Spec.Adapters == nil {
		d.Spec.Adapters = &ReleaseManifestAdapters{}
	}

	switch adapter := adapter.(type) {
	case *HTTPAdapter:
		d.Spec.Adapters.HTTPAdapters = append(d.Spec.Adapters.HTTPAdapters, HTTPAdapterManifest{
			TypeMeta:  adapter.TypeMeta,
			ObjectRef: common.RefFromMeta(adapter.ObjectMeta),
			Spec:      adapter.Spec,
			Details:   adapter.Details,
		})
	}
}

func init() {
	SchemeBuilder.Register(&ReleaseManifest{}, &ReleaseManifestList{})
}
