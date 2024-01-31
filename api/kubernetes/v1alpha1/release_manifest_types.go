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

	Environment common.ObjectRef `json:"environment"`

	// +kubebuilder:validation:Required

	VirtualEnvironment common.ObjectRef `json:"virtualEnvironment"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinProperties=1

	Apps map[string]*ReleaseManifestApp `json:"apps"`

	Adapters *ReleaseManifestAdapters `json:"adapters,omitempty"`
}

type ReleaseManifestApp struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Version string `json:"version"`

	// +kubebuilder:validation:Required

	AppDeployment ReleaseManifestAppDep `json:"appDeployment"`
}

type ReleaseManifestAppDep struct {
	common.ObjectRef `json:",inline"`

	// +kubebuilder:validation:Required

	Spec AppDeploymentSpec `json:"spec"`
}

type ReleaseManifestAdapters struct {
	HTTPAdapters []ReleaseManifestHTTPAdapter `json:"http,omitempty"`
}

type ReleaseManifestHTTPAdapter struct {
	common.ObjectRef `json:",inline"`

	// +kubebuilder:validation:Required

	Spec HTTPAdapterSpec `json:"spec"`
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
		d.Spec.Adapters.HTTPAdapters = append(d.Spec.Adapters.HTTPAdapters, ReleaseManifestHTTPAdapter{
			ObjectRef: common.ObjectRef{
				UID:             adapter.UID,
				ResourceVersion: adapter.ResourceVersion,
				Generation:      adapter.Generation,
				Name:            adapter.Name,
			},
			Spec: adapter.Spec,
		})
	}
}

func (d *ReleaseManifest) GetAdapter(name string, typ api.ComponentType) (common.Adapter, error) {
	if d.Spec.Adapters == nil {
		return nil, core.ErrNotFound()
	}

	var a common.Adapter
	switch typ {
	case api.ComponentTypeHTTPAdapter:
		for _, h := range d.Spec.Adapters.HTTPAdapters {
			if h.Name == name {
				a = &HTTPAdapter{
					ObjectMeta: h.ObjectMeta(),
					Spec:       h.Spec,
				}
				break
			}
		}
	}

	if a == nil {
		return nil, core.ErrNotFound()
	}

	return a, nil
}

func init() {
	SchemeBuilder.Register(&ReleaseManifest{}, &ReleaseManifestList{})
}
