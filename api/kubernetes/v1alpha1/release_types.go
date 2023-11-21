/*
Copyright © 2023 XigXog

This Source Code Form is subject to the terms of the Mozilla Public License,
v2.0. If a copy of the MPL was not distributed with this file, You can obtain
one at https://mozilla.org/MPL/2.0/.
*/

package v1alpha1

import (
	"github.com/xigxog/kubefox/api"
	common "github.com/xigxog/kubefox/api/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReleaseSpec defines the desired state of Release
type ReleaseSpec struct {
	// Version of the App being release. Use of semantic versioning is
	// recommended. If set the value is compared to the AppDeployment, if they
	// conflict the release will fail.
	Version string `json:"version,omitempty"`
	// +kubebuilder:validation:Enum=Promotion;Release;Rollback
	Type          api.ReleaseType      `json:"type"`
	Environment   common.Ref           `json:"environment"`
	AppDeployment ReleaseAppDeployment `json:"appDeployment"`
}

type ReleaseAppDeployment struct {
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
}

// ReleaseStatus defines the observed state of Release
type ReleaseStatus struct {
	CreationTime       metav1.Time  `json:"creationTime,omitempty"`
	PendingTime        *metav1.Time `json:"pendingTime,omitempty"`
	ReleaseTime        *metav1.Time `json:"releaseTime,omitempty"`
	SupersededTime     *metav1.Time `json:"supersededTime,omitempty"`
	LastTransitionTime metav1.Time  `json:"lastTransitionTime,omitempty"`
	FailureTime        *metav1.Time `json:"failureTime,omitempty"`
	FailureMessage     string       `json:"failureMessage,omitempty"`
}

// ReleaseDetails defines additional details of Release
type ReleaseDetails struct {
	api.Details `json:",inline"`

	AppDeployment AppDeploymentDetails `json:"appDeployment,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:subresource:details

// Release is the Schema for the Releases API
type Release struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec    ReleaseSpec    `json:"spec,omitempty"`
	Status  ReleaseStatus  `json:"status,omitempty"`
	Details ReleaseDetails `json:"details,omitempty"`
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
