/*
Copyright © 2023 XigXog

This Source Code Form is subject to the terms of the Mozilla Public License,
v2.0. If a copy of the MPL was not distributed with this file, You can obtain
one at https://mozilla.org/MPL/2.0/.
*/

package v1alpha1

import (
	"fmt"
	"time"

	"github.com/mitchellh/hashstructure/v2"
	"github.com/xigxog/kubefox/api"
	common "github.com/xigxog/kubefox/api/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type VirtualEnvironmentSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1

	// Name of the Environment this VirtualEnvironment is part of.
	Environment   string                `json:"environment"`
	Release       *Release              `json:"release,omitempty"`
	ReleasePolicy *VirtEnvReleasePolicy `json:"releasePolicy,omitempty"`
}

type Release struct {
	// +kubebuilder:validation:Required

	AppDeployment ReleaseAppDeployment `json:"appDeployment"`

	// Name of DataSnapshot to use for Release. If set, the immutable Data
	// object of the snapshot will be used. The source of the snapshot must be
	// this VirtualEnvironment.
	DataSnapshot string `json:"dataSnapshot,omitempty"`
}

type ReleaseAppDeployment struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1

	Name string `json:"name"`
	// Version of the App being released. Use of semantic versioning is
	// recommended. If set the value is compared to the AppDeployment version.
	// If the two versions do not match the release will fail.
	Version string `json:"version,omitempty"`
}

type VirtEnvReleasePolicy struct {
	// +kubebuilder:validation:Minimum=3

	// If the pending Request cannot be activated before the deadline it will be
	// considered failed. If the Release becomes available for activation after
	// the deadline has been exceeded, it will not be activated. Pointer is used
	// to distinguish between not set and false.
	PendingDeadlineSeconds *uint `json:"pendingDeadlineSeconds,omitempty"`

	// If true '.spec.release.appDeployment.version' is required. Pointer is
	// used to distinguish between not set and false.
	VersionRequired *bool `json:"versionRequired,omitempty"`

	// If '.spec.release.dataSnapshot' is required. Pointer is used to
	// distinguish between not set and false.
	DataSnapshotRequired *bool `json:"dataSnapshotRequired,omitempty"`

	HistoryLimits *VirtEnvHistoryLimits `json:"historyLimits,omitempty"`
}

type VirtEnvHistoryLimits struct {
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
	AppDeployment ReleaseAppDeploymentStatus `json:"appDeployment,omitempty"`
	DataSnapshot  string                     `json:"dataSnapshot,omitempty"`

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

	Problems []common.Problem `json:"problems,omitempty"`
}

type ReleaseAppDeploymentStatus struct {
	ReleaseAppDeployment `json:",inline"`

	// ObservedGeneration represents the .metadata.generation of the
	// AppDeployment that the status was set based upon. For instance, if the
	// AppDeployment .metadata.generation is currently 12, but the
	// observedGeneration is 9, the status is out of date with respect to the
	// current state of the instance.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

type VirtualEnvironmentStatus struct {
	// DataChecksum is a hash value of the Data object. The Environment Data
	// object is merged before the hash is create. It can be used to check for
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
// +kubebuilder:printcolumn:name="Available",type=string,JSONPath=`.status.conditions[?(@.type=='ActiveReleaseAvailable')].status`
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.conditions[?(@.type=='ActiveReleaseAvailable')].reason`
// +kubebuilder:printcolumn:name="Release",type=string,JSONPath=`.status.activeRelease.appDeployment.name`
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.status.activeRelease.appDeployment.version`
// +kubebuilder:printcolumn:name="Snapshot",type=string,JSONPath=`.status.activeRelease.dataSnapshot`,priority=1
// +kubebuilder:printcolumn:name="Pending",type=string,JSONPath=`.status.pendingRelease.appDeployment.name`
// +kubebuilder:printcolumn:name="Pending Version",type=string,JSONPath=`.status.pendingRelease.appDeployment.version`
// +kubebuilder:printcolumn:name="Pending Snapshot",type=string,JSONPath=`.status.pendingRelease.dataSnapshot`,priority=1
// +kubebuilder:printcolumn:name="Pending Reason",type=string,JSONPath=`.status.conditions[?(@.type=='ReleasePending')].reason`,priority=1
// +kubebuilder:printcolumn:name="Title",type=string,JSONPath=`.details.title`,priority=1
// +kubebuilder:printcolumn:name="Description",type=string,JSONPath=`.details.description`,priority=1

type VirtualEnvironment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec    VirtualEnvironmentSpec   `json:"spec,omitempty"`
	Data    api.Data                 `json:"data,omitempty"`
	Details DataDetails              `json:"details,omitempty"`
	Status  VirtualEnvironmentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type VirtualEnvironmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []VirtualEnvironment `json:"items"`
}

func (r *ReleaseStatus) Equals(other *Release) bool {
	switch {
	case r == nil && other == nil:
		return true
	case r == nil && other != nil:
		return false
	case r != nil && other == nil:
		return false
	}

	return r.AppDeployment.ReleaseAppDeployment == other.AppDeployment &&
		r.DataSnapshot == other.DataSnapshot
}

func (ve *VirtualEnvironment) GetData() *api.Data {
	return &ve.Data
}

func (ve *VirtualEnvironment) GetDataChecksum() string {
	hash, _ := hashstructure.Hash(&ve.Data, hashstructure.FormatV2, nil)
	return fmt.Sprint(hash)
}

func (ve *VirtualEnvironment) GetReleasePendingDeadline() time.Duration {
	secs := ve.Spec.ReleasePolicy.PendingDeadlineSeconds
	if secs == nil || *secs == 0 {
		return api.DefaultReleasePendingDeadlineSeconds * time.Second
	}

	return time.Duration(*secs * uint(time.Second))
}

// GetReleasePendingDuration returns the current duration that the Release has
// been pending. If there is no Release pending 0 is returned.
func (ve *VirtualEnvironment) GetReleasePendingDuration() time.Duration {
	if ve.Status.PendingRelease == nil {
		return 0
	}

	return time.Since(ve.Status.PendingRelease.RequestTime.Time)
}

func (ve *VirtualEnvironment) Merge(env *Environment) {
	if ve.Data.Vars == nil {
		ve.Data.Vars = env.Data.Vars
		ve.Details.Vars = env.Details.Vars
	} else {
		mergeDataAndDetails(
			ve.Data.Vars, ve.Details.Vars,
			env.Data.Vars, env.Details.Vars)
	}

	if ve.Data.Secrets == nil {
		ve.Data.Secrets = env.Data.Secrets
		ve.Details.Secrets = env.Details.Secrets
	} else {
		mergeDataAndDetails(
			ve.Data.Secrets, ve.Details.Secrets,
			env.Data.Secrets, env.Details.Secrets)
	}

	if ve.Spec.ReleasePolicy == nil {
		ve.Spec.ReleasePolicy = &VirtEnvReleasePolicy{}
	}
	if ve.Spec.ReleasePolicy.PendingDeadlineSeconds == nil {
		ve.Spec.ReleasePolicy.PendingDeadlineSeconds =
			&env.Spec.ReleasePolicy.PendingDeadlineSeconds
	}
	if ve.Spec.ReleasePolicy.VersionRequired == nil {
		if env.Spec.ReleasePolicy.VersionRequired == nil {
			ve.Spec.ReleasePolicy.VersionRequired = api.True
		} else {
			ve.Spec.ReleasePolicy.VersionRequired =
				env.Spec.ReleasePolicy.VersionRequired
		}
	}
	if ve.Spec.ReleasePolicy.DataSnapshotRequired == nil {
		if env.Spec.ReleasePolicy.DataSnapshotRequired == nil {
			ve.Spec.ReleasePolicy.DataSnapshotRequired = api.True
		} else {
			ve.Spec.ReleasePolicy.DataSnapshotRequired =
				env.Spec.ReleasePolicy.DataSnapshotRequired
		}
	}

	if ve.Spec.ReleasePolicy.HistoryLimits == nil {
		ve.Spec.ReleasePolicy.HistoryLimits = &VirtEnvHistoryLimits{}
	}
	if ve.Spec.ReleasePolicy.HistoryLimits.Count == nil {
		ve.Spec.ReleasePolicy.HistoryLimits.Count =
			&env.Spec.ReleasePolicy.HistoryLimits.Count
	}
	if ve.Spec.ReleasePolicy.HistoryLimits.AgeDays == nil {
		ve.Spec.ReleasePolicy.HistoryLimits.AgeDays =
			&env.Spec.ReleasePolicy.HistoryLimits.AgeDays
	}

	if ve.Details.Title == "" {
		ve.Details.Title = env.Details.Title
	}
	if ve.Details.Description == "" {
		ve.Details.Description = env.Details.Description
	}
}

func mergeDataAndDetails[V string | *api.Val](
	dstData map[string]V, dstDetails map[string]api.Details,
	srcData map[string]V, srcDetails map[string]api.Details) {

	if srcDetails == nil {
		srcDetails = map[string]api.Details{}
	}

	for k, v := range srcData {
		if _, found := dstData[k]; !found {
			dstData[k] = v
		}

		d := dstDetails[k]
		s := srcDetails[k]
		if d.Title == "" {
			d.Title = s.Title
		}
		if d.Description == "" {
			d.Description = s.Description
		}
		if !(d.Title == "" && d.Description == "") {
			dstDetails[k] = d
		}
	}
}

func init() {
	SchemeBuilder.Register(&VirtualEnvironment{}, &VirtualEnvironmentList{})
}
