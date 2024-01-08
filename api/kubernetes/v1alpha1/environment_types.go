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

type EnvironmentSpec struct {
	ReleasePolicy EnvReleasePolicy `json:"releasePolicy,omitempty"`
}

type EnvReleasePolicy struct {
	// +kubebuilder:validation:Minimum=3
	// +kubebuilder:default=300

	// If the pending Request cannot be activated before the deadline it will be
	// considered failed. If the Release becomes available for activation after
	// the deadline has been exceeded, it will not be activated.
	PendingDeadlineSeconds uint `json:"pendingDeadlineSeconds,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true

	// If true '.spec.release.appDeployment.version' is required. Pointer is
	// used to distinguish between not set and false.
	VersionRequired *bool `json:"versionRequired"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true

	// If '.spec.release.dataSnapshot' is required. Pointer is used to
	// distinguish between not set and false.
	DataSnapshotRequired *bool `json:"dataSnapshotRequired"`

	HistoryLimits EnvReleaseHistoryLimits `json:"historyLimits"`
}

type EnvReleaseHistoryLimits struct {
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=10

	// Maximum number of Releases to keep in history. Once the limit is reached
	// the oldest Release in history will be deleted. Age is based on
	// archiveTime.
	Count uint `json:"count,omitempty"`

	// +kubebuilder:validation:Minimum=0

	// Maximum age of the Release to keep in history. Once the limit is reached
	// the oldest Release in history will be deleted. Age is based on
	// archiveTime. Set to 0 to disable.
	AgeDays uint `json:"ageDays,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=environments,scope=Cluster,shortName=env
// +kubebuilder:printcolumn:name="Title",type=string,JSONPath=`.details.title`,priority=1
// +kubebuilder:printcolumn:name="Description",type=string,JSONPath=`.details.description`,priority=1

type Environment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec    EnvironmentSpec `json:"spec,omitempty"`
	Data    api.Data        `json:"data,omitempty"`
	Details DataDetails     `json:"details,omitempty"`
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

func (env *Environment) GetDataChecksum() string {
	hash, _ := hashstructure.Hash(&env.Data, hashstructure.FormatV2, nil)
	return fmt.Sprint(hash)
}

func init() {
	SchemeBuilder.Register(&Environment{}, &EnvironmentList{})
}
