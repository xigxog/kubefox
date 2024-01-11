/*
Copyright © 2023 XigXog

This Source Code Form is subject to the terms of the Mozilla Public License,
v2.0. If a copy of the MPL was not distributed with this file, You can obtain
one at https://mozilla.org/MPL/2.0/.
*/

// +kubebuilder:object:generate=true
package api

// +kubebuilder:object:generate=false
type Adapter interface {
	GetName() string
	GetComponentType() ComponentType
	Validate(data *Data) Problems
}

type EnvVarSchema map[string]*EnvVarDefinition

type EnvSchema struct {
	Vars    EnvVarSchema `json:"vars,omitempty"`
	Secrets EnvVarSchema `json:"secrets,omitempty"`
}

type EnvVarDefinition struct {
	// +kubebuilder:validation:Enum=Array;Boolean;Number;String
	Type EnvVarType `json:"type,omitempty"`
	// +kubebuilder:default=false
	Required bool `json:"required"`
}

type ComponentDefinition struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=DBAdapter;KubeFox;HTTPAdapter
	Type           ComponentType          `json:"type"`
	Routes         []RouteSpec            `json:"routes,omitempty"`
	DefaultHandler bool                   `json:"defaultHandler,omitempty"`
	EnvVarSchema   EnvVarSchema           `json:"envVarSchema,omitempty"`
	Dependencies   map[string]*Dependency `json:"dependencies,omitempty"`
}

type RouteSpec struct {
	// +kubebuilder:validation:Required
	Id int `json:"id"`
	// +kubebuilder:validation:Required
	Rule         string       `json:"rule"`
	Priority     int          `json:"priority,omitempty"`
	EnvVarSchema EnvVarSchema `json:"envVarSchema,omitempty"`
}

type Dependency struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=DBAdapter;KubeFox;HTTPAdapter
	Type ComponentType `json:"type"`
}

type Details struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

type Problems []Problem

type Problem struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=AdapterNotFound;AppDeploymentFailed;DependencyInvalid;DependencyNotFound;ParseError;PolicyViolation;RouteConflict;VarNotFound;VarWrongType

	Type ProblemType `json:"type"`

	// +kubebuilder:validation:Required

	Message string `json:"message,omitempty"`
	// Resources and attributes causing problem.
	Causes []ProblemSource `json:"causes,omitempty"`
}

type ProblemSource struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=AppDeployment;Component;HTTPAdapter;Release;ReleaseManifest;VirtualEnvironment;

	Kind ProblemSourceKind `json:"kind"`
	Name string            `json:"name,omitempty"`
	// ObservedGeneration represents the .metadata.generation of the
	// ProblemSource that the problem was generated from. For instance, if the
	// ProblemSource .metadata.generation is currently 12, but the
	// observedGeneration is 9, the problem is out of date with respect to the
	// current state of the instance.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Path of source object attribute causing problem.
	Path string `json:"path,omitempty"`
	// Value causing problem. Pointer is used to distinguish between not set and
	// empty string.
	Value *string `json:"value,omitempty"`
}
