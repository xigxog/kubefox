/*
Copyright © 2023 XigXog

This Source Code Form is subject to the terms of the Mozilla Public License,
v2.0. If a copy of the MPL was not distributed with this file, You can obtain
one at https://mozilla.org/MPL/2.0/.
*/

package v1alpha1

import (
	common "github.com/xigxog/kubefox/api/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EnvironmentSpec defines the desired state of Environment
type EnvironmentSpec struct {
	EnvSpec `json:",inline"`

	Parent EnvParent `json:"parent,omitempty"`
}

type EnvSpec struct {
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Vars     map[string]*common.Val `json:"vars,omitempty"`
	Adapters map[string]*Adapter    `json:"adapters,omitempty"`
}

type EnvParent struct {
	Name string `json:"name"`
}

type Adapter struct {
	// +kubebuilder:validation:Enum=db;http
	Type common.ComponentType `json:"type"`
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
	FollowRedirects common.FollowRedirects `json:"followRedirects,omitempty"`
}

// EnvironmentStatus defines the observed state of Environment
type EnvironmentStatus struct {
	Parent   common.RefTimestamped   `json:"parent,omitempty"`
	Children []common.RefTimestamped `json:"children,omitempty"`
	Spec     EnvSpecStatus           `json:"spec,omitempty"`
}

type EnvSpecStatus struct {
	Resolved EnvSpec `json:"resolved,omitempty"`
}

// EnvironmentDetails defines additional details of Environment
type EnvironmentDetails struct {
	common.Details `json:",inline"`

	Vars     map[string]common.Details `json:"vars,omitempty"`
	Adapters map[string]common.Details `json:"adapters,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:subresource:details
//+kubebuilder:resource:path=environments,scope=Cluster

// Environment is the Schema for the Environments API
type Environment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec    EnvironmentSpec    `json:"spec,omitempty"`
	Status  EnvironmentStatus  `json:"status,omitempty"`
	Details EnvironmentDetails `json:"details,omitempty"`
}

//+kubebuilder:object:root=true

// EnvironmentList contains a list of Environments
type EnvironmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Environment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Environment{}, &EnvironmentList{})
}
