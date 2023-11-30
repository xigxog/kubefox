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

type ReleaseSpec struct {
	AppDeployment ReleaseAppDeployment `json:"appDeployment"`
	// +kubebuilder:validation:Optional
	VirtualEnvSnapshot string               `json:"virtualEnvSnapshot"`
	HistoryLimit       *ReleaseHistoryLimit `json:"historyLimit,omitempty"`
}

type ReleaseAppDeployment struct {
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
	// Version of the App being released. Use of semantic versioning is
	// recommended. If set the value is compared to the AppDeployment version.
	// If the two versions do not match the release will fail.
	// +kubebuilder:validation:Optional
	Version string `json:"version"`
}

type ReleaseHistoryLimit struct {
	// Total number of archived Releases to keep. Once the limit is reach the
	// oldest Release will be removed from history. Default 100.
	Count uint `json:"count,omitempty"`
	// Age of the oldest archived Release to keep. Age is based on archiveTime.
	AgeDays uint `json:"ageDays,omitempty"`
}

type ReleaseStatusEntry struct {
	AppDeployment      ReleaseAppDeployment `json:"appDeployment"`
	VirtualEnvSnapshot string               `json:"virtualEnvSnapshot,omitempty"`
	RequestTime        metav1.Time          `json:"requestTime,omitempty"`
	AvailableTime      *metav1.Time         `json:"availableTime,omitempty"`
	ArchiveTime        *metav1.Time         `json:"archiveTime,omitempty"`
}

type ReleaseStatus struct {
	// +kubebuilder:validation:Optional
	Current   *ReleaseStatusEntry `json:"current"`
	Requested *ReleaseStatusEntry `json:"requested,omitempty"`

	History []ReleaseStatusEntry `json:"history,omitempty"`
	// TODO conditions
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type Release struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReleaseSpec   `json:"spec"`
	Status ReleaseStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ReleaseList contains a list of Releases
type ReleaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Release `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Release{}, &ReleaseList{})
}
