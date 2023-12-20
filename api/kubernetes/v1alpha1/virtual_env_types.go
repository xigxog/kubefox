/*
Copyright © 2023 XigXog

This Source Code Form is subject to the terms of the Mozilla Public License,
v2.0. If a copy of the MPL was not distributed with this file, You can obtain
one at https://mozilla.org/MPL/2.0/.
*/

package v1alpha1

import (
	"github.com/xigxog/kubefox/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type VirtualEnvSpec struct {
	// Parent ClusterVirtualEnv. Note, only ClusterVirtualEnvs can be used as
	// parents.
	Parent string `json:"parent,omitempty"`

	Release *Release `json:"release,omitempty"`

	ReleasePolicies *ReleasePolicies `json:"releasePolicies,omitempty"`
}

type Release struct {
	// +kubebuilder:validation:Required

	AppDeployment ReleaseAppDeployment `json:"appDeployment"`

	// Name of VirtualEnvSnapshot to use for Release. If set the immutable Data
	// object of the snapshot will be used. The source VirtualEnv of the
	// snapshot must be the same as the VirtualEnv of the Release.
	VirtualEnvSnapshot string `json:"virtualEnvSnapshot,omitempty"`
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

type ReleasePolicies struct {
	// +kubebuilder:validation:Minimum=3

	// If the pending Request cannot be activated before the deadline it will be
	// considered failed. If the Release becomes available for activation after
	// the deadline has been exceeded, it will not be activated.
	PendingDeadlineSeconds uint `json:"pendingDeadlineSeconds,omitempty"`

	// +kubebuilder:validation:Enum=VersionOptional;VersionRequired

	AppDeploymentPolicy api.AppDeploymentPolicy `json:"appDeploymentPolicy,omitempty"`

	// +kubebuilder:validation:Enum=SnapshotOptional;SnapshotRequired

	VirtualEnvPolicy api.VirtualEnvPolicy `json:"virtualEnvPolicy,omitempty"`

	HistoryLimits *ReleaseHistoryLimits `json:"historyLimits,omitempty"`
}

type ReleaseHistoryLimits struct {
	// Maximum number of Releases to keep in history. Once the limit is reached
	// the oldest Release in history will be deleted. Age is based on
	// archiveTime. Default 10.
	Count uint `json:"count,omitempty"`

	// Maximum age of the Release to keep in history. Once the limit is reached
	// the oldest Release in history will be deleted. Age is based on
	// archiveTime.
	AgeDays uint `json:"ageDays,omitempty"`
}

type ReleaseStatus struct {
	AppDeployment      ReleaseAppDeploymentStatus `json:"appDeployment,omitempty"`
	VirtualEnvSnapshot string                     `json:"virtualEnvSnapshot,omitempty"`

	// Time at which the VirtualEnv was updated to use the Release.
	RequestTime metav1.Time `json:"requestTime,omitempty"`
	// Time at which the Release became active. If not set the Release was never
	// active.
	ActivationTime *metav1.Time `json:"activationTime,omitempty"`
	// Time at which the Release was archived to history.
	ArchiveTime *metav1.Time   `json:"archiveTime,omitempty"`
	Errors      []ReleaseError `json:"errors,omitempty"`
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

type ReleaseError struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=ParseError;VarConflict;VarNotFound;VarWrongType

	Type string `json:"type"`
	// JSON path of VirtualEnv attribute causing error.
	Path    string `json:"path,omitempty"`
	Message string `json:"message,omitempty"`
	// Source of error.
	Source *ErrorSource `json:"source,omitempty"`
}

type ErrorSource struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=AppDeployment;HTTPAdapter
	Kind string `json:"kind"`

	// +kubebuilder:validation:Required

	Name string `json:"name"`

	// +kubebuilder:validation:Required

	ResourceVersion string `json:"resourceVersion"`
	// JSON path of source object attribute causing error.
	Path string `json:"path,omitempty"`
	// Value causing error.
	Value string `json:"value,omitempty"`
}

type VirtualEnvData struct {
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields

	Vars    map[string]*api.Val `json:"vars,omitempty"`
	Secrets map[string]*api.Val `json:"secrets,omitempty"`
}

type VirtualEnvDetails struct {
	api.Details `json:",inline"`

	Vars    map[string]api.Details `json:"vars,omitempty"`
	Secrets map[string]api.Details `json:"secrets,omitempty"`
}

type VirtualEnvStatus struct {
	// DataChecksum is a hash value of the Data object. If the VirtualEnv has a
	// parent the parent's Data object is merged before the hash is create. It
	// can be used to check for changes to the Data object.
	DataChecksum uint64 `json:"dataChecksum,omitempty"`

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

type VirtualEnv struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec    VirtualEnvSpec    `json:"spec,omitempty"`
	Data    VirtualEnvData    `json:"data,omitempty"`
	Details VirtualEnvDetails `json:"details,omitempty"`
	Status  VirtualEnvStatus  `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type VirtualEnvList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []VirtualEnv `json:"items"`
}

func (env *VirtualEnv) MergeParent(parent *ClusterVirtualEnv) {
	if env.Data.Vars == nil {
		env.Data.Vars = parent.Data.Vars
		env.Details.Vars = parent.Details.Vars
	} else {
		mergeDataAndDetails(
			env.Data.Vars, env.Details.Vars,
			parent.Data.Vars, parent.Details.Vars)
	}

	if env.Data.Secrets == nil {
		env.Data.Secrets = parent.Data.Secrets
		env.Details.Secrets = parent.Details.Secrets
	} else {
		mergeDataAndDetails(
			env.Data.Secrets, env.Details.Secrets,
			parent.Data.Secrets, parent.Details.Secrets)
	}

	if env.Spec.ReleasePolicies == nil {
		env.Spec.ReleasePolicies = parent.Spec.ReleasePolicies
	} else {
		if env.Spec.ReleasePolicies.AppDeploymentPolicy == "" {
			env.Spec.ReleasePolicies.AppDeploymentPolicy =
				parent.Spec.ReleasePolicies.AppDeploymentPolicy
		}
		if env.Spec.ReleasePolicies.VirtualEnvPolicy == "" {
			env.Spec.ReleasePolicies.VirtualEnvPolicy =
				parent.Spec.ReleasePolicies.VirtualEnvPolicy
		}
		if env.Spec.ReleasePolicies.HistoryLimits == nil {
			env.Spec.ReleasePolicies.HistoryLimits =
				parent.Spec.ReleasePolicies.HistoryLimits
		}
	}

	if env.Details.Title == "" {
		env.Details.Title = parent.Details.Title
	}
	if env.Details.Description == "" {
		env.Details.Description = parent.Details.Description
	}
}

func mergeDataAndDetails(
	dstData map[string]*api.Val, dstDetails map[string]api.Details,
	srcData map[string]*api.Val, srcDetails map[string]api.Details) {

	if srcDetails == nil {
		srcDetails = map[string]api.Details{}
	}

	for k, v := range srcData {
		if d, found := dstData[k]; !found || d.IsNil() {
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
	SchemeBuilder.Register(&VirtualEnv{}, &VirtualEnvList{})
}
