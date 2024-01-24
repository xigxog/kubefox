// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package v1alpha1

import (
	"github.com/xigxog/kubefox/api"
	common "github.com/xigxog/kubefox/api/kubernetes"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PlatformSpec defines the desired state of Platform
type PlatformSpec struct {
	Events  EventsSpec        `json:"events,omitempty"`
	Broker  BrokerSpec        `json:"broker,omitempty"`
	HTTPSrv HTTPSrvSpec       `json:"httpsrv,omitempty"`
	NATS    NATSSpec          `json:"nats,omitempty"`
	Logger  common.LoggerSpec `json:"logger,omitempty"`
}

type EventsSpec struct {
	// +kubebuilder:validation:Minimum=3
	// +kubebuilder:default=30
	TimeoutSeconds uint `json:"timeoutSeconds,omitempty"`

	// Large events reduce performance and increase memory usage. Default 5Mi.
	// Maximum 16Mi.
	MaxSize resource.Quantity `json:"maxSize,omitempty"`
}

type NATSSpec struct {
	PodSpec       common.PodSpec       `json:"podSpec,omitempty"`
	ContainerSpec common.ContainerSpec `json:"containerSpec,omitempty"`
}

type HTTPSrvSpec struct {
	PodSpec       common.PodSpec       `json:"podSpec,omitempty"`
	ContainerSpec common.ContainerSpec `json:"containerSpec,omitempty"`
	Service       HTTPSrvService       `json:"service,omitempty"`
}

type BrokerSpec struct {
	PodSpec       common.PodSpec       `json:"podSpec,omitempty"`
	ContainerSpec common.ContainerSpec `json:"containerSpec,omitempty"`
}

type HTTPSrvService struct {
	// +kubebuilder:validation:Enum=ClusterIP;NodePort;LoadBalancer
	// +kubebuilder:default=ClusterIP
	Type  string       `json:"type,omitempty"`
	Ports HTTPSrvPorts `json:"ports,omitempty"`

	// Map of string keys and values that can be used to organize and categorize
	// (scope and select) objects. May match selectors of replication
	// controllers and services. [More
	// info](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels).
	Labels map[string]string `json:"labels,omitempty"`
	// Annotations is an unstructured key value map stored with a resource that
	// may be set by external tools to store and retrieve arbitrary metadata.
	// They are not queryable and should be preserved when modifying objects.
	// [More
	// info](https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations).
	Annotations map[string]string `json:"annotations,omitempty"`
}

type HTTPSrvPorts struct {
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=80
	HTTP uint `json:"http,omitempty"`
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=443
	HTTPS uint `json:"https,omitempty"`
}

// PlatformStatus defines the observed state of Platform
type PlatformStatus struct {
	// +patchStrategy=merge
	// +patchMergeKey=podName
	// +listType=map
	// +listMapKey=podName
	Components []ComponentStatus `json:"components,omitempty" patchStrategy:"merge" patchMergeKey:"podName"`
	// +patchStrategy=merge
	// +patchMergeKey=type
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

type ComponentStatus struct {
	Ready    bool              `json:"ready"`
	Name     string            `json:"name"`
	Commit   string            `json:"commit,omitempty"`
	Type     api.ComponentType `json:"type,omitempty"`
	PodName  string            `json:"podName"`
	PodIP    string            `json:"podIP"`
	NodeName string            `json:"nodeName"`
	NodeIP   string            `json:"nodeIP"`
}

// PlatformDetails defines additional details of Platform
type PlatformDetails struct {
	api.Details `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Available",type=string,JSONPath=`.status.conditions[?(@.type=='Available')].status`
// +kubebuilder:printcolumn:name="Event Timeout",type=integer,JSONPath=`.spec.events.timeoutSeconds`
// +kubebuilder:printcolumn:name="Event Max",type=string,JSONPath=`.spec.events.maxSize`
// +kubebuilder:printcolumn:name="Log Level",type=string,JSONPath=`.spec.logger.level`

// Platform is the Schema for the Platforms API
type Platform struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec    PlatformSpec    `json:"spec,omitempty"`
	Status  PlatformStatus  `json:"status,omitempty"`
	Details PlatformDetails `json:"details,omitempty"`
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
