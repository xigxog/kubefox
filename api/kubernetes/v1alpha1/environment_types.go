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

type EnvData struct {
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Vars     map[string]*api.Val    `json:"vars,omitempty"`
	Adapters map[string]*EnvAdapter `json:"adapters,omitempty"`
}

type EnvAdapter struct {
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

// EnvDetails defines additional details of VirtualEnv
type EnvDetails struct {
	api.Details `json:",inline"`

	Vars     map[string]api.Details `json:"vars,omitempty"`
	Adapters map[string]api.Details `json:"adapters,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:data
//+kubebuilder:subresource:details

// VirtualEnv is the Schema for the VirtualEnv API
type VirtualEnv struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Data    EnvData    `json:"data,omitempty"`
	Details EnvDetails `json:"details,omitempty"`
}

//+kubebuilder:object:root=true

// VirtualEnvList contains a list of Environments
type VirtualEnvList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []VirtualEnv `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:data
// +kubebuilder:subresource:details
// +kubebuilder:resource:path=environments,scope=Cluster
type ClusterVirtualEnv struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Data    EnvData    `json:"data,omitempty"`
	Details EnvDetails `json:"details,omitempty"`
}

// +kubebuilder:object:root=true
type ClusterVirtualEnvList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []VirtualEnv `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VirtualEnv{}, &VirtualEnvList{})
	SchemeBuilder.Register(&ClusterVirtualEnv{}, &ClusterVirtualEnvList{})
}
