// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package v1alpha1

import (
	"errors"
	"fmt"

	"github.com/xigxog/kubefox/api"
	common "github.com/xigxog/kubefox/api/kubernetes"
	"github.com/xigxog/kubefox/core"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AppDeploymentSpec defines the desired state of AppDeployment
type AppDeploymentSpec struct {
	// +kubebuilder:validation:Required

	AppName string `json:"appName"`

	// Version of the defined App. Use of semantic versioning is recommended.
	// Once set the AppDeployment spec becomes immutable.
	Version string `json:"version,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="^[a-z0-9]{40}$"

	Commit string `json:"commit"`

	// +kubebuilder:validation:Required

	CommitTime          metav1.Time `json:"commitTime"`
	Branch              string      `json:"branch,omitempty"`
	Tag                 string      `json:"tag,omitempty"`
	RepoURL             string      `json:"repoURL,omitempty"`
	ContainerRegistry   string      `json:"containerRegistry,omitempty"`
	ImagePullSecretName string      `json:"imagePullSecretName,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinProperties=1

	Components map[string]*api.ComponentDefinition `json:"components"`
}

// AppDeploymentStatus defines the observed state of AppDeployment
type AppDeploymentStatus struct {
	// +patchStrategy=merge
	// +patchMergeKey=type
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
	Problems   api.Problems       `json:"problems,omitempty"`
}

// AppDeploymentDetails defines additional details of AppDeployment
type AppDeploymentDetails struct {
	api.Details `json:",inline"`

	Components map[string]api.Details `json:"components,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=appdeployments,shortName=appdep;app
// +kubebuilder:printcolumn:name="App",type=string,JSONPath=`.spec.appName`
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`
// +kubebuilder:printcolumn:name="Available",type=string,JSONPath=`.status.conditions[?(@.type=='Available')].status`
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.conditions[?(@.type=='Available')].reason`
// +kubebuilder:printcolumn:name="Progressing",type=string,JSONPath=`.status.conditions[?(@.type=='Progressing')].status`

// AppDeployment is the Schema for the AppDeployments API
type AppDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec    AppDeploymentSpec    `json:"spec,omitempty"`
	Status  AppDeploymentStatus  `json:"status,omitempty"`
	Details AppDeploymentDetails `json:"details,omitempty"`
}

//+kubebuilder:object:root=true

// AppDeploymentList contains a list of AppDeployments
type AppDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []AppDeployment `json:"items"`
}

func (d *AppDeployment) Validate(data *api.Data, getAdapter common.GetAdapterFunc) (api.Problems, error) {
	problems := api.Problems{}
	for compName, comp := range d.Spec.Components {
		if data != nil {
			problems = append(problems, comp.EnvVarSchema.Validate("Var", data.Vars, &api.ProblemSource{
				Kind:               api.ProblemSourceKindAppDeployment,
				Name:               d.Name,
				ObservedGeneration: d.Generation,
				Path:               fmt.Sprintf("$.spec.components.%s.envVarSchema", compName),
			}, true)...)

			for i, route := range comp.Routes {
				// All route vars are required.
				for _, d := range route.EnvVarSchema {
					d.Required = true
				}
				problems = append(problems, route.EnvVarSchema.Validate("Var", data.Vars, &api.ProblemSource{
					Kind:               api.ProblemSourceKindAppDeployment,
					Name:               d.Name,
					ObservedGeneration: d.Generation,
					Path:               fmt.Sprintf("$.spec.components.%s.routes[%d]", compName, i),
				}, true)...)
			}
		}

		for depName, dep := range comp.Dependencies {
			found := true
			switch {
			case dep.Type == api.ComponentTypeKubeFox:
				_, found = d.Spec.Components[depName]

			case dep.Type.IsAdapter():
				adapter, err := getAdapter(depName, dep.Type)
				switch {
				case err == nil:
					if data != nil {
						problems = append(problems, adapter.Validate(data)...)
					}

				case apierrors.IsNotFound(err) || errors.Is(err, core.ErrNotFound()):
					found = false

				default:
					return nil, err
				}

			default:
				// Unsupported dependency type.
				problems = append(problems, api.Problem{
					Type: api.ProblemTypeDependencyInvalid,
					Message: fmt.Sprintf(`Component "%s" dependency "%s" has unsupported type "%s".`,
						compName, depName, dep.Type),
					Causes: []api.ProblemSource{
						{
							Kind:               api.ProblemSourceKindAppDeployment,
							Name:               d.Name,
							ObservedGeneration: d.Generation,
							Path: fmt.Sprintf("$.spec.components.%s.dependencies.%s.type",
								compName, depName),
							Value: (*string)(&dep.Type),
						},
					},
				})
			}

			if !found {
				problems = append(problems, api.Problem{
					Type: api.ProblemTypeDependencyNotFound,
					Message: fmt.Sprintf(`Component "%s" dependency "%s" of type "%s" not found.`,
						compName, depName, dep.Type),
					Causes: []api.ProblemSource{
						{
							Kind:               api.ProblemSourceKindAppDeployment,
							Name:               d.Name,
							ObservedGeneration: d.Generation,
							Path: fmt.Sprintf("$.spec.components.%s.dependencies.%s",
								compName, depName),
						},
					},
				})
			}
		}
	}

	return problems, nil
}

func (a *AppDeployment) GetDefinition(comp *core.Component) (*api.ComponentDefinition, error) {
	if comp == nil {
		return nil, core.ErrComponentMismatch(fmt.Errorf("component not part of app"))
	}

	c, found := a.Spec.Components[comp.Name]
	if !found {
		return nil, core.ErrComponentMismatch(fmt.Errorf("component not part of app"))
	}
	if comp.Type != string(c.Type) {
		return nil, core.ErrComponentMismatch(fmt.Errorf("component type does not match app"))
	}
	if comp.Commit != "" && comp.Commit != c.Commit {
		return nil, core.ErrComponentMismatch(fmt.Errorf("component commit does not match app"))
	}

	return c, nil
}

func (a *AppDeployment) HasDependency(name string, typ api.ComponentType) bool {
	for _, c := range a.Spec.Components {
		for n, d := range c.Dependencies {
			if name == n && typ == d.Type {
				return true
			}
		}
	}

	return false
}

func init() {
	SchemeBuilder.Register(&AppDeployment{}, &AppDeploymentList{})
}
