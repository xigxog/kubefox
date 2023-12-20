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

type VirtualEnvSnapshotSpec struct {
	// +kubebuilder:validation:Required

	// VirtualEnv that this snapshot is of. Note, ClusterVirtualEnvs cannot be
	// snapshotted.
	Source VirtualEnvSource `json:"source"`
}

type VirtualEnvSource struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1

	// Name of the VirtualEnv this snapshot is of. Note, ClusterVirtualEnvs
	// cannot be snapshotted.
	Name string `json:"name"`

	// ResourceVersion of the VirtualEnv this snapshot is of. If data is
	// provided at creation of the VirtualEnvSnapshot then resourceVersion must
	// match the current resourceVersion of the VirtualEnv. If data is not
	// provided at creation time resourceVersion will be populated
	// automatically.
	ResourceVersion string `json:"resourceVersion,omitempty"`

	// DataChecksum is the hash of the VirtualEnv's data this snapshot is of. If
	// data is provided at creation of the VirtualEnvSnapshot then dataChecksum
	// must match the current dataChecksum of the VirtualEnv. If data is not
	// provided at creation time dataChecksum will be populated automatically.
	DataChecksum uint64 `json:"dataChecksum,omitempty"`
}

// +kubebuilder:object:root=true
type VirtualEnvSnapshot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec VirtualEnvSnapshotSpec `json:"spec,omitempty"`
	// Data is a copy of the source VirtualEnv's data object. If provided at
	// creation time then the source VirtualEnv's resourceVersion and current
	// dataChecksum must also be provided. If set to nil at creation time then
	// the current data object, resourceVersion, and dataChecksum of the source
	// VirtualEnv will automatically be copied.
	Data    *VirtualEnvData   `json:"data,omitempty"`
	Details VirtualEnvDetails `json:"details,omitempty"`
}

// +kubebuilder:object:root=true
type VirtualEnvSnapshotList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []VirtualEnvSnapshot `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VirtualEnvSnapshot{}, &VirtualEnvSnapshotList{})
}
