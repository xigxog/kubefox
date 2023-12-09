/*
Copyright © 2023 XigXog

This Source Code Form is subject to the terms of the Mozilla Public License,
v2.0. If a copy of the MPL was not distributed with this file, You can obtain
one at https://mozilla.org/MPL/2.0/.
*/

// +kubebuilder:object:generate=true
package api

type EnvVarSchema struct {
	// +kubebuilder:validation:Enum=array;boolean;number;string
	Type     EnvVarType `json:"type,omitempty"`
	Required bool       `json:"required"`
	// Unique indicates that this environment variable must have a unique value
	// across all environments. If the value is not unique then making a dynamic
	// request or creating a release that utilizes this variable will fail.
	Unique bool `json:"unique"`
}

type App struct {
	Name              string `json:"name"`
	ContainerRegistry string `json:"containerRegistry,omitempty"`
}

type AppDetails struct {
	App     `json:",inline"`
	Details `json:",inline"`
}

type ComponentDefinition struct {
	// +kubebuilder:validation:Enum=db;genesis;kubefox;http
	Type           ComponentType            `json:"type"`
	Routes         []RouteSpec              `json:"routes,omitempty"`
	DefaultHandler bool                     `json:"defaultHandler,omitempty"`
	EnvSchema      map[string]*EnvVarSchema `json:"envSchema,omitempty"`
	Dependencies   map[string]*Dependency   `json:"dependencies,omitempty"`
}

type ComponentDetails struct {
	ComponentDefinition `json:",inline"`
	Details             `json:",inline"`
}

type RouteSpec struct {
	Id        int                      `json:"id"`
	Rule      string                   `json:"rule"`
	Priority  int                      `json:"priority,omitempty"`
	EnvSchema map[string]*EnvVarSchema `json:"envSchema,omitempty"`
}

type Dependency struct {
	// +kubebuilder:validation:Enum=db;kubefox;http
	Type ComponentType `json:"type"`
}

type Details struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}
