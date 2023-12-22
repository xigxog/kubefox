/*
Copyright © 2023 XigXog

This Source Code Form is subject to the terms of the Mozilla Public License,
v2.0. If a copy of the MPL was not distributed with this file, You can obtain
one at https://mozilla.org/MPL/2.0/.
*/

// +kubebuilder:object:generate=true
package api

import (
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type VirtualEnvVarDefinition struct {
	// +kubebuilder:validation:Enum=array;boolean;number;string
	Type     EnvVarType `json:"type,omitempty"`
	Required bool       `json:"required"`
	// Unique indicates that this environment variable must have a unique value
	// across all environments. If the value is not unique then making a dynamic
	// request or creating a release that utilizes this variable will fail.
	Unique bool `json:"unique"`
}

type ComponentDefinition struct {
	// +kubebuilder:validation:Enum=db;genesis;kubefox;http
	Type             ComponentType                       `json:"type"`
	Routes           []RouteSpec                         `json:"routes,omitempty"`
	DefaultHandler   bool                                `json:"defaultHandler,omitempty"`
	VirtualEnvSchema map[string]*VirtualEnvVarDefinition `json:"virtualEnvSchema,omitempty"`
	Dependencies     map[string]*Dependency              `json:"dependencies,omitempty"`
}

type ComponentDetails struct {
	ComponentDefinition `json:",inline"`
	Details             `json:",inline"`
}

type RouteSpec struct {
	Id               int                                 `json:"id"`
	Rule             string                              `json:"rule"`
	Priority         int                                 `json:"priority,omitempty"`
	VirtualEnvSchema map[string]*VirtualEnvVarDefinition `json:"virtualEnvSchema,omitempty"`
}

type Dependency struct {
	// +kubebuilder:validation:Enum=db;kubefox;http
	Type ComponentType `json:"type"`
}

type Details struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

// +kubebuilder:object:generate=false
type Adapter interface {
	GetName() string
	GetComponentType() ComponentType
}

// +kubebuilder:object:generate=false

// UncomparableTime is a Kubernetes v1.Time object that will also be equal to
// another UncomparableTime object when using equality.Semantic, even if the
// times are different.
type UncomparableTime metav1.Time

// DeepCopyInto creates a deep-copy of the UncomparableTime value.  The
// underlying time.Time type is effectively immutable in the time API, so it is
// safe to copy-by-assign, despite the presence of (unexported) Pointer fields.
func (t *UncomparableTime) DeepCopyInto(out *UncomparableTime) {
	*out = *t
}

func init() {
	equality.Semantic.AddFunc(func(a, b UncomparableTime) bool {
		return true
	})
}
