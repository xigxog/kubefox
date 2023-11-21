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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PlatformSpec defines the desired state of Platform
type PlatformSpec struct {
	Environments map[string]*PlatformEnv `json:"environments,omitempty"`
	Config       PlatformConfig          `json:"config,omitempty"`
}

type PlatformEnv struct {
	Release            *PlatformEnvRelease            `json:"release,omitempty"`
	SupersededReleases map[string]*PlatformEnvRelease `json:"supersededReleases,omitempty"`
}

type PlatformEnvRelease struct {
	ReleaseSpec   `json:",inline"`
	ReleaseStatus `json:",inline"`

	Name                   string `json:"name"`
	AppDeploymentAvailable bool   `json:"appDeploymentAvailable,omitempty"`
}

type PlatformConfig struct {
	Events   EventsSpec        `json:"events,omitempty"`
	Releases ReleasesSpec      `json:"releases,omitempty"`
	Broker   BrokerSpec        `json:"broker,omitempty"`
	HTTPSrv  HTTPSrvSpec       `json:"httpsrv,omitempty"`
	NATS     NATSSpec          `json:"nats,omitempty"`
	Logger   common.LoggerSpec `json:"logger,omitempty"`
}

type EventsSpec struct {
	// +kubebuilder:validation:Minimum=3
	TimeoutSeconds uint `json:"timeoutSeconds,omitempty"`
	// Large events reduce performance and increase memory usage. Default 5MiB.
	// Maximum 16 MiB.
	MaxSize resource.Quantity `json:"maxSize,omitempty"`
}

type ReleasesSpec struct {
	// Number of seconds after which a newly created Release is marked as failed
	// if it has not become available. Defaults to 300 seconds (5 minutes).
	// +kubebuilder:validation:Minimum=3
	TimeoutSeconds uint `json:"timeoutSeconds,omitempty"`
	// Limits on superseded Requests.
	Limits ReleaseLimits `json:"limits,omitempty"`
}

type ReleaseLimits struct {
	// Total number of superseded Requests to keep. Once the limit is reach the
	// oldest unused Request will be removed. Default 100.
	Count uint `json:"count,omitempty"`
	// Age of the oldest superseded Request to keep. Age is based on when the
	// Release was superseded.
	AgeDays uint `json:"ageDays,omitempty"`
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

// PlatformDetails defines additional details of Platform
type PlatformDetails struct {
	api.Details `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:subresource:details

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

// Environment returns the named PlatformEnv creating it if needed.
func (p *Platform) Environment(name string) *PlatformEnv {
	if p.Spec.Environments == nil {
		p.Spec.Environments = make(map[string]*PlatformEnv)
	}
	env := p.Spec.Environments[name]
	if env == nil {
		env = &PlatformEnv{}
		p.Spec.Environments[name] = env
	}
	if env.SupersededReleases == nil {
		env.SupersededReleases = map[string]*PlatformEnvRelease{}
	}

	return env
}

// FindRelease looks for the named release in all Environments checking active,
// pending, and superseded.
func (p *Platform) FindRelease(name string) *PlatformEnvRelease {
	for _, env := range p.Spec.Environments {
		if env.Release != nil && env.Release.Name == name {
			return env.Release
		}
		for _, r := range env.SupersededReleases {
			if r != nil && r.Name == name {
				return r
			}
		}
	}

	return nil
}
