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
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +kubebuilder:object:generate=false
type VirtualEnvObject interface {
	client.Object

	GetData() *VirtualEnvData
	GetDetails() *VirtualEnvDetails
	GetReleasePolicy() *VirtualEnvReleasePolicy
	SetReleasePolicy(*VirtualEnvReleasePolicy)
	GetParent() string
	GetEnvName() string
}

type VirtualEnvSpec struct {
	// Parent ClusterVirtualEnv.
	Parent        string                   `json:"parent,omitempty"`
	ReleasePolicy *VirtualEnvReleasePolicy `json:"releasePolicy,omitempty"`
}

type VirtualEnvReleasePolicy struct {
	// +kubebuilder:validation:Enum=VersionOptional;VersionRequired
	AppDeploymentPolicy api.AppDeploymentPolicy `json:"appDeploymentPolicy,omitempty"`
	// +kubebuilder:validation:Enum=SnapshotOptional;SnapshotRequired
	VirtualEnvPolicy api.VirtualEnvPolicy `json:"virtualEnvPolicy,omitempty"`
}

type VirtualEnvData struct {
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Vars     map[string]*api.Val `json:"vars,omitempty"`
	Adapters map[string]*Adapter `json:"adapters,omitempty"`
}

type Adapter struct {
	// +kubebuilder:validation:Enum=db;http
	Type api.ComponentType `json:"type"`
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:pruning:PreserveUnknownFields
	URL common.StringOrSecret `json:"url,omitempty"`
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Headers map[string]*common.StringOrSecret `json:"headers,omitempty"`
	// InsecureSkipVerify controls whether a client verifies the server's
	// certificate chain and host name. If InsecureSkipVerify is true, any
	// certificate presented by the server and any host name in that certificate
	// is accepted. In this mode, TLS is susceptible to machine-in-the-middle
	// attacks.
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`
	// Defaults to never.
	// +kubebuilder:validation:Enum=Never;Always;SameHost
	FollowRedirects api.FollowRedirects `json:"followRedirects,omitempty"`
}

type VirtualEnvDetails struct {
	api.Details `json:",inline"`

	Vars     map[string]api.Details `json:"vars,omitempty"`
	Adapters map[string]api.Details `json:"adapters,omitempty"`
}

// +kubebuilder:object:root=true
type VirtualEnv struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec    VirtualEnvSpec    `json:"spec,omitempty"`
	Data    VirtualEnvData    `json:"data,omitempty"`
	Details VirtualEnvDetails `json:"details,omitempty"`
}

// +kubebuilder:object:root=true
type VirtualEnvList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []VirtualEnv `json:"items"`
}

func (e *VirtualEnv) GetData() *VirtualEnvData {
	return &e.Data
}

func (e *VirtualEnv) GetDetails() *VirtualEnvDetails {
	return &e.Details
}

func (e *VirtualEnv) GetReleasePolicy() *VirtualEnvReleasePolicy {
	return e.Spec.ReleasePolicy
}

func (e *VirtualEnv) SetReleasePolicy(p *VirtualEnvReleasePolicy) {
	e.Spec.ReleasePolicy = p
}

func (e *VirtualEnv) GetParent() string {
	return e.Spec.Parent
}

func (e *VirtualEnv) GetEnvName() string {
	return e.Name
}

func init() {
	SchemeBuilder.Register(&VirtualEnv{}, &VirtualEnvList{})
}

func MergeVirtualEnvironment(dst, src VirtualEnvObject) {
	if dst.GetData().Vars == nil {
		dst.GetData().Vars = map[string]*api.Val{}
	}
	if src.GetDetails().Vars == nil {
		src.GetDetails().Vars = map[string]api.Details{}
	}
	for k, v := range src.GetData().Vars {
		dst.GetData().Vars[k] = v
		if details, found := src.GetDetails().Vars[k]; found {
			snapshotDetails := dst.GetDetails().Vars[k]
			if details.Title != "" {
				snapshotDetails.Title = details.Title
			}
			if details.Description != "" {
				snapshotDetails.Description = details.Description
			}
			dst.GetDetails().Vars[k] = snapshotDetails
		}
	}

	if dst.GetData().Adapters == nil {
		dst.GetData().Adapters = map[string]*Adapter{}
	}
	if src.GetDetails().Adapters == nil {
		src.GetDetails().Adapters = map[string]api.Details{}
	}
	for k, v := range src.GetData().Adapters {
		dst.GetData().Adapters[k] = v
		if details, found := src.GetDetails().Adapters[k]; found {
			snapshotDetails := dst.GetDetails().Adapters[k]
			if details.Title != "" {
				snapshotDetails.Title = details.Title
			}
			if details.Description != "" {
				snapshotDetails.Description = details.Description
			}
			dst.GetDetails().Adapters[k] = snapshotDetails
		}
	}

	if src.GetReleasePolicy() != nil {
		if dst.GetReleasePolicy() == nil {
			dst.SetReleasePolicy(&VirtualEnvReleasePolicy{})
		}
		if src.GetReleasePolicy().AppDeploymentPolicy != "" {
			dst.GetReleasePolicy().AppDeploymentPolicy = src.GetReleasePolicy().AppDeploymentPolicy
		}
		if src.GetReleasePolicy().VirtualEnvPolicy != "" {
			dst.GetReleasePolicy().VirtualEnvPolicy = src.GetReleasePolicy().VirtualEnvPolicy
		}
	}

	if src.GetDetails().Title != "" {
		dst.GetDetails().Title = src.GetDetails().Title
	}
	if src.GetDetails().Description != "" {
		dst.GetDetails().Description = src.GetDetails().Description
	}
}
