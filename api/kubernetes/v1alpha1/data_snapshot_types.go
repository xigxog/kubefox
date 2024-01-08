/*
Copyright © 2023 XigXog

This Source Code Form is subject to the terms of the Mozilla Public License,
v2.0. If a copy of the MPL was not distributed with this file, You can obtain
one at https://mozilla.org/MPL/2.0/.
*/

package v1alpha1

import (
	"fmt"

	"github.com/mitchellh/hashstructure/v2"
	"github.com/xigxog/kubefox/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DataSnapshotSpec struct {
	// +kubebuilder:validation:Required

	// Source resource that this snapshot is of.
	Source DataSource `json:"source"`
}

type DataSource struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=VirtualEnvironment

	// Kind of the resource the data is sourced from.
	Kind api.DataSourceKind `json:"kind"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1

	// Name of the resource the data is sourced from.
	Name string `json:"name"`

	// ResourceVersion of the source resource. If data is provided at creation of
	// the snapshot then resourceVersion must match the current resourceVersion
	// of the source. If data is not provided at creation time resourceVersion
	// will be populated automatically.
	ResourceVersion string `json:"resourceVersion,omitempty"`

	// DataChecksum is the hash of the source's data. If data is provided at
	// creation of the snapshot then dataChecksum must match the current
	// dataChecksum of the source. If data is not provided at creation time
	// dataChecksum will be populated automatically.
	DataChecksum string `json:"dataChecksum,omitempty"`
}

type DataDetails struct {
	api.Details `json:",inline"`

	Vars    map[string]api.Details `json:"vars,omitempty"`
	Secrets map[string]api.Details `json:"secrets,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=datasnapshots,shortName=data
// +kubebuilder:printcolumn:name="Source",type=string,JSONPath=`.spec.source.name`
// +kubebuilder:printcolumn:name="Created",type=string,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:printcolumn:name="Title",type=string,JSONPath=`.details.title`,priority=1
// +kubebuilder:printcolumn:name="Description",type=string,JSONPath=`.details.description`,priority=1

type DataSnapshot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec DataSnapshotSpec `json:"spec,omitempty"`
	// Data is a copy of the source's data object. If provided at creation time
	// then the source's resourceVersion and current dataChecksum must also be
	// provided. If set to nil at creation time then the current data object,
	// resourceVersion, and dataChecksum of the source will automatically be
	// copied.
	Data    *api.Data   `json:"data,omitempty"`
	Details DataDetails `json:"details,omitempty"`
}

// +kubebuilder:object:root=true
type DataSnapshotList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []DataSnapshot `json:"items"`
}

func (d *DataSnapshot) GetData() *api.Data {
	return d.Data
}

func (d *DataSnapshot) GetDataChecksum() string {
	hash, _ := hashstructure.Hash(d.Data, hashstructure.FormatV2, nil)
	return fmt.Sprint(hash)
}

func init() {
	SchemeBuilder.Register(&DataSnapshot{}, &DataSnapshotList{})
}
