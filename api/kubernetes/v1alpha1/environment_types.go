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
)

type EnvironmentSpec struct {
	ReleasePolicy EnvReleasePolicy `json:"releasePolicy,omitempty"`
}

type EnvReleasePolicy struct {
	// +kubebuilder:validation:Enum=Stable;Testing
	// +kubebuilder:default=Stable

	Type api.ReleaseType `json:"type,omitempty"`

	// +kubebuilder:validation:Minimum=3
	// +kubebuilder:default=300

	// If the pending Release cannot be activated before the activation deadline
	// it will be considered failed and the Release will automatically rolled
	// back to the current active Release. Pointer is used to distinguish
	// between not set and false.
	ActivationDeadlineSeconds *uint `json:"activationDeadlineSeconds,omitempty"`

	HistoryLimits EnvHistoryLimits `json:"historyLimits,omitempty"`
}

type EnvHistoryLimits struct {
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=10

	// Maximum number of Releases to keep in history. Once the limit is reached
	// the oldest Release in history will be deleted. Age is based on
	// archiveTime. Pointer is used to distinguish between not set and false.
	Count *uint `json:"count,omitempty"`

	// +kubebuilder:validation:Minimum=0

	// Maximum age of the Release to keep in history. Once the limit is reached
	// the oldest Release in history will be deleted. Age is based on
	// archiveTime. Set to 0 to disable. Pointer is used to distinguish between
	// not set and false.
	AgeDays *uint `json:"ageDays,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=environments,scope=Cluster,shortName=env

type Environment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec    EnvironmentSpec `json:"spec,omitempty"`
	Data    api.Data        `json:"data,omitempty"`
	Details api.DataDetails `json:"details,omitempty"`
}

// +kubebuilder:object:root=true
type EnvironmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Environment `json:"items"`
}

func (env *Environment) GetData() *api.Data {
	return &env.Data
}

func (d *Environment) GetDataKey() api.DataKey {
	return api.DataKey{
		Kind: d.Kind,
		Name: d.Name,
	}
}

func init() {
	SchemeBuilder.Register(&Environment{}, &EnvironmentList{})
}
