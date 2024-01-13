/*
Copyright © 2023 XigXog

This Source Code Form is subject to the terms of the Mozilla Public License,
v2.0. If a copy of the MPL was not distributed with this file, You can obtain
one at https://mozilla.org/MPL/2.0/.
*/

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

	// Name of the Environment this VirtualEnvironment is part of.
	Environment   string                `json:"environment"`
	Release       *Release              `json:"release,omitempty"`
	ReleasePolicy *VirtEnvReleasePolicy `json:"releasePolicy,omitempty"`
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
	ReleaseManifest string                      `json:"releaseManifest,omitempty"`
	Apps            map[string]ReleaseAppStatus `json:"apps,omitempty"`

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

type ReleaseAppStatus struct {
	ReleaseApp `json:",inline"`

	// ObservedGeneration represents the .metadata.generation of the
	// AppDeployment that the status was set based upon. For instance, if the
	// AppDeployment .metadata.generation is currently 12, but the
	// observedGeneration is 9, the status is out of date with respect to the
	// current state of the instance.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
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
// +kubebuilder:printcolumn:name="Manifest",type=string,JSONPath=`.status.activeRelease.releaseManifest`
// +kubebuilder:printcolumn:name="Available",type=string,JSONPath=`.status.conditions[?(@.type=='ActiveReleaseAvailable')].status`
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.conditions[?(@.type=='ActiveReleaseAvailable')].reason`
// +kubebuilder:printcolumn:name="Pending",type=string,JSONPath=`.status.conditions[?(@.type=='ReleasePending')].status`
// +kubebuilder:printcolumn:name="Pending Reason",type=string,JSONPath=`.status.conditions[?(@.type=='ReleasePending')].reason`
// +kubebuilder:printcolumn:name="Title",type=string,JSONPath=`.details.title`,priority=1
// +kubebuilder:printcolumn:name="Description",type=string,JSONPath=`.details.description`,priority=1

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

func (p *VirtEnvReleasePolicy) GetPendingDeadline() time.Duration {
	if p == nil {
		return api.DefaultReleasePendingDeadlineSeconds * time.Second
	}

	secs := p.PendingDeadlineSeconds
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

func (ve *VirtualEnvironment) GetReleasePolicy(env *Environment) *VirtEnvReleasePolicy {
	p := ve.Spec.ReleasePolicy.DeepCopy()
	if p == nil {
		p = &VirtEnvReleasePolicy{}
	}

	if p.PendingDeadlineSeconds == nil {
		p.PendingDeadlineSeconds =
			&env.Spec.ReleasePolicy.PendingDeadlineSeconds
	}
	if p.VersionRequired == nil {
		if env.Spec.ReleasePolicy.VersionRequired == nil {
			p.VersionRequired = api.True
		} else {
			p.VersionRequired =
				env.Spec.ReleasePolicy.VersionRequired
		}
	}

	if p.HistoryLimits == nil {
		p.HistoryLimits = &VirtEnvHistoryLimits{}
	}
	if p.HistoryLimits.Count == nil {
		p.HistoryLimits.Count =
			&env.Spec.ReleasePolicy.HistoryLimits.Count
	}
	if p.HistoryLimits.AgeDays == nil {
		p.HistoryLimits.AgeDays =
			&env.Spec.ReleasePolicy.HistoryLimits.AgeDays
	}

	return p
}

func init() {
	SchemeBuilder.Register(&VirtualEnvironment{}, &VirtualEnvironmentList{})
}
