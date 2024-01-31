// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package v1alpha1

import (
	"time"

	"github.com/xigxog/kubefox/api"
	common "github.com/xigxog/kubefox/api/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type VirtualEnvironmentSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1

	// Name of the Environment this VirtualEnvironment is part of. This field is
	// immutable.
	Environment   string         `json:"environment"`
	Release       *Release       `json:"release,omitempty"`
	ReleasePolicy *ReleasePolicy `json:"releasePolicy,omitempty"`
}

type Release struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinProperties=1

	Apps map[string]ReleaseApp `json:"apps"`
}

type ReleaseApp struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1

	AppDeployment string `json:"appDeployment"`

	// Version of the App being released. Use of semantic versioning is
	// recommended. If set the value is compared to the AppDeployment version.
	// If the two versions do not match the release will fail.
	Version string `json:"version,omitempty"`
}

type ReleasePolicy struct {
	// +kubebuilder:validation:Enum=Stable;Testing

	Type api.ReleaseType `json:"type,omitempty"`

	// +kubebuilder:validation:Minimum=3

	// If the pending Release cannot be activated before the activation deadline
	// it will be considered failed and the Release will automatically rolled
	// back to the current active Release. Pointer is used to distinguish
	// between not set and false.
	ActivationDeadlineSeconds *uint `json:"activationDeadlineSeconds,omitempty"`

	HistoryLimits *HistoryLimits `json:"historyLimits,omitempty"`
}

type HistoryLimits struct {
	// +kubebuilder:validation:Minimum=0

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

type ReleaseStatus struct {
	Release `json:",inline"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1

	Id string `json:"id"`

	ReleaseManifest string `json:"releaseManifest,omitempty"`
	// Time at which the VirtualEnvironment was updated to use the Release.
	RequestTime metav1.Time `json:"requestTime,omitempty"`
	// Time at which the Release became active. If not set the Release was never
	// activated.
	ActivationTime *metav1.Time `json:"activationTime,omitempty"`
	// Time at which the Release was archived to history.
	ArchiveTime *metav1.Time `json:"archiveTime,omitempty"`

	// +kubebuilder:validation:Enum=PendingDeadlineExceeded;RolledBack;Superseded

	// Reason Release was archived.
	ArchiveReason api.ArchiveReason `json:"archiveReason,omitempty"`
	Problems      []common.Problem  `json:"problems,omitempty"`
}

type VirtualEnvironmentStatus struct {
	// DataChecksum is a hash value of the Data object. The Environment Data
	// object is merged before the hash is created. It can be used to check for
	// changes to the Data object.
	DataChecksum string `json:"dataChecksum,omitempty"`

	PendingReleaseFailed bool `json:"pendingReleaseFailed,omitempty"`

	// +kubebuilder:validation:Optional

	ActiveRelease *ReleaseStatus `json:"activeRelease"`

	PendingRelease *ReleaseStatus `json:"pendingRelease,omitempty"`

	ReleaseHistory []ReleaseStatus `json:"releaseHistory,omitempty"`

	// +patchStrategy=merge
	// +patchMergeKey=type
	// +listType=map
	// +listMapKey=type

	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=virtualenvironments,shortName=virtenv;ve
// +kubebuilder:printcolumn:name="Environment",type=string,JSONPath=`.spec.environment`
// +kubebuilder:printcolumn:name="Manifest",type=string,JSONPath=`.status.activeRelease.releaseManifest`
// +kubebuilder:printcolumn:name="Available",type=string,JSONPath=`.status.conditions[?(@.type=='ActiveReleaseAvailable')].status`
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.conditions[?(@.type=='ActiveReleaseAvailable')].reason`
// +kubebuilder:printcolumn:name="Pending",type=string,JSONPath=`.status.conditions[?(@.type=='ReleasePending')].status`
// +kubebuilder:printcolumn:name="Pending Reason",type=string,JSONPath=`.status.conditions[?(@.type=='ReleasePending')].reason`

type VirtualEnvironment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec    VirtualEnvironmentSpec   `json:"spec,omitempty"`
	Data    api.Data                 `json:"data,omitempty"`
	Details api.DataDetails          `json:"details,omitempty"`
	Status  VirtualEnvironmentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type VirtualEnvironmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []VirtualEnvironment `json:"items"`
}

func (ve *VirtualEnvironment) GetData() *api.Data {
	return &ve.Data
}

func (d *VirtualEnvironment) GetDataKey() api.DataKey {
	return api.DataKey{
		Kind:      d.Kind,
		Name:      d.Name,
		Namespace: d.Namespace,
	}
}

// GetReleasePendingDuration returns the current duration that the Release has
// been pending. If there is no Release pending 0 is returned.
func (ve *VirtualEnvironment) GetReleasePendingDuration() time.Duration {
	if ve.Status.PendingRelease == nil {
		return 0
	}

	return time.Since(ve.Status.PendingRelease.RequestTime.Time)
}

func (ve *VirtualEnvironment) GetReleasePolicy(env *Environment) *ReleasePolicy {
	envPol := env.Spec.ReleasePolicy
	vePol := ve.Spec.ReleasePolicy.DeepCopy()
	if vePol == nil {
		vePol = &ReleasePolicy{}
	}

	if vePol.ActivationDeadlineSeconds == nil {
		if envPol.ActivationDeadlineSeconds == nil ||
			*envPol.ActivationDeadlineSeconds == 0 {
			i := uint(api.DefaultReleaseActivationDeadlineSeconds)
			vePol.ActivationDeadlineSeconds = &i
		} else {
			vePol.ActivationDeadlineSeconds = envPol.ActivationDeadlineSeconds
		}
	}
	if vePol.Type == "" {
		if envPol.Type == "" {
			vePol.Type = api.ReleaseTypeStable
		} else {
			vePol.Type = envPol.Type
		}
	}

	if vePol.HistoryLimits == nil {
		vePol.HistoryLimits = &HistoryLimits{}
	}
	if vePol.HistoryLimits.Count == nil {
		if envPol.HistoryLimits.Count == nil ||
			*envPol.HistoryLimits.Count == 0 {
			i := uint(api.DefaultReleaseHistoryCountLimit)
			vePol.HistoryLimits.Count = &i
		} else {
			vePol.HistoryLimits.Count = envPol.HistoryLimits.Count
		}
	}
	if vePol.HistoryLimits.AgeDays == nil {
		if envPol.HistoryLimits.AgeDays == nil ||
			*envPol.HistoryLimits.AgeDays == 0 {
			i := uint(api.DefaultReleaseHistoryAgeLimit)
			vePol.HistoryLimits.AgeDays = &i
		} else {
			vePol.HistoryLimits.AgeDays = envPol.HistoryLimits.AgeDays
		}
	}

	return vePol
}

func (ve *VirtualEnvironment) UsesAppDeployment(name string) bool {
	if ve.Status.ActiveRelease.ContainsAppDeployment(name) {
		return true
	}
	if ve.Status.PendingRelease.ContainsAppDeployment(name) {
		return true
	}

	return false
}

func (ve *VirtualEnvironment) UsesReleaseManifest(name string) bool {
	if ve.Status.ActiveRelease.ContainsReleaseManifest(name) {
		return true
	}

	for _, rel := range ve.Status.ReleaseHistory {
		if rel.ContainsReleaseManifest(name) {
			return true
		}
	}

	return false
}

func (p *ReleasePolicy) GetPendingDeadline() time.Duration {
	if p == nil {
		return api.DefaultReleaseActivationDeadlineSeconds * time.Second
	}

	secs := p.ActivationDeadlineSeconds
	if secs == nil || *secs == 0 {
		return api.DefaultReleaseActivationDeadlineSeconds * time.Second
	}

	return time.Duration(*secs * uint(time.Second))
}

func (s *ReleaseStatus) ContainsAppDeployment(name string) bool {
	if s == nil {
		return false
	}

	for _, app := range s.Apps {
		if app.AppDeployment == name {
			return true
		}
	}

	return false
}

func (s *ReleaseStatus) ContainsReleaseManifest(name string) bool {
	if s == nil {
		return false
	}

	return s.ReleaseManifest == name
}

func init() {
	SchemeBuilder.Register(&VirtualEnvironment{}, &VirtualEnvironmentList{})
}
