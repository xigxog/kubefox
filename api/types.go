/*
Copyright © 2023 XigXog

This Source Code Form is subject to the terms of the Mozilla Public License,
v2.0. If a copy of the MPL was not distributed with this file, You can obtain
one at https://mozilla.org/MPL/2.0/.
*/

// +kubebuilder:object:generate=true
package api

type VirtualEnvData struct {
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields

	Vars    map[string]*Val `json:"vars,omitempty"`
	Secrets map[string]*Val `json:"secrets,omitempty"`
}

type EnvVarDefinition struct {
	// +kubebuilder:validation:Enum=array;boolean;number;string
	Type EnvVarType `json:"type,omitempty"`
	// +kubebuilder:default=false
	Required bool `json:"required"`
}

type ComponentDefinition struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=db;genesis;kubefox;http
	Type           ComponentType                `json:"type"`
	Routes         []RouteSpec                  `json:"routes,omitempty"`
	DefaultHandler bool                         `json:"defaultHandler,omitempty"`
	EnvVarSchema   map[string]*EnvVarDefinition `json:"envVarSchema,omitempty"`
	Dependencies   map[string]*Dependency       `json:"dependencies,omitempty"`
}

type ComponentDetails struct {
	ComponentDefinition `json:",inline"`
	Details             `json:",inline"`
}

type RouteSpec struct {
	// +kubebuilder:validation:Required
	Id int `json:"id"`
	// +kubebuilder:validation:Required
	Rule         string                       `json:"rule"`
	Priority     int                          `json:"priority,omitempty"`
	EnvVarSchema map[string]*EnvVarDefinition `json:"envVarSchema,omitempty"`
}

type Dependency struct {
	// +kubebuilder:validation:Required
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
