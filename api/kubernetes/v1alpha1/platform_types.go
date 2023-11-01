/*
Copyright © 2023 XigXog

This Source Code Form is subject to the terms of the Mozilla Public License,
v2.0. If a copy of the MPL was not distributed with this file, You can obtain
one at https://mozilla.org/MPL/2.0/.
*/

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PlatformSpec defines the desired state of Platform
type PlatformSpec struct {
	Broker  Broker  `json:"broker,omitempty"`
	HTTPSrv HTTPSrv `json:"httpsrv,omitempty"`
	NATS    NATS    `json:"nats,omitempty"`
	// +kubebuilder:validation:Minimum=3
	EventTTLSeconds uint              `json:"defaultEventTTLSeconds,omitempty"`
	EventMaxSize    resource.Quantity `json:"eventMaxSize,omitempty"`
}

type NATS struct {
	PodSpec       `json:",inline"`
	ContainerSpec `json:",inline"`
}

type HTTPSrv struct {
	PodSpec       `json:",inline"`
	ContainerSpec `json:",inline"`

	Service HTTPSrvService `json:"service,omitempty"`
}

type Broker struct {
	PodSpec       `json:",inline"`
	ContainerSpec `json:",inline"`
}

type HTTPSrvService struct {
	// +kubebuilder:validation:Enum=ClusterIP;NodePort;LoadBalancer
	Type  string       `json:"type,omitempty"`
	Ports HTTPSrvPorts `json:"ports,omitempty"`
}

type HTTPSrvPorts struct {
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:validation:Minimum=1
	HTTP uint `json:"http,omitempty"`
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:validation:Minimum=1
	HTTPS uint `json:"https,omitempty"`
}

// PlatformStatus defines the observed state of Platform
type PlatformStatus struct {
	// +kubebuilder:validation:Optional
	Ready bool `json:"ready"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Platform is the Schema for the Platforms API
type Platform struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PlatformSpec   `json:"spec,omitempty"`
	Status PlatformStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PlatformList contains a list of Platforms
type PlatformList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Platform `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Platform{}, &PlatformList{})
}
