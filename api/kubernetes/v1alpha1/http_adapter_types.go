/*
Copyright © 2023 XigXog

This Source Code Form is subject to the terms of the Mozilla Public License,
v2.0. If a copy of the MPL was not distributed with this file, You can obtain
one at https://mozilla.org/MPL/2.0/.
*/

package v1alpha1

import (
	"fmt"

	"github.com/xigxog/kubefox/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type HTTPAdapterSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Format=uri

	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`

	// +kubebuilder:default=false

	// InsecureSkipVerify controls whether the Adapter verifies the server's
	// certificate chain and host name. If InsecureSkipVerify is true, any
	// certificate presented by the server and any host name in that certificate
	// is accepted. In this mode, TLS is susceptible to machine-in-the-middle
	// attacks.
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`

	// +kubebuilder:default=Never
	// +kubebuilder:validation:Enum=Never;Always;SameHost

	FollowRedirects api.FollowRedirects `json:"followRedirects,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=httpadapters,shortName=http
// +kubebuilder:printcolumn:name="Title",type=string,JSONPath=`.details.title`,priority=1
// +kubebuilder:printcolumn:name="Description",type=string,JSONPath=`.details.description`,priority=1

type HTTPAdapter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec    HTTPAdapterSpec `json:"spec,omitempty"`
	Details api.Details     `json:"details,omitempty"`
}

// +kubebuilder:object:root=true
type HTTPAdapterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []HTTPAdapter `json:"items"`
}

func (a *HTTPAdapter) GetComponentType() api.ComponentType {
	return api.ComponentTypeHTTPAdapter
}

func (a *HTTPAdapter) Validate(data *api.Data) api.Problems {
	var problems api.Problems

	src := api.ProblemSource{
		Kind:               api.ProblemSourceKindHTTPAdapter,
		Name:               a.Name,
		ObservedGeneration: a.Generation,
		Path:               "$.spec.url",
		Value:              &a.Spec.URL,
	}
	if t, err := api.NewEnvTemplate(a.Spec.URL); err != nil {
		problems = append(problems, api.Problem{
			Type:    api.ProblemTypeParseError,
			Message: fmt.Sprintf(`Error parsing url "%s": %s`, a.Spec.URL, err.Error()),
			Causes:  []api.ProblemSource{src},
		})
	} else {
		problems = append(problems, t.EnvSchema().Vars.Validate(data.Vars, &src)...)
		problems = append(problems, t.EnvSchema().Secrets.Validate(data.ResolvedSecrets, &src)...)
	}

	for header, val := range a.Spec.Headers {
		src := api.ProblemSource{
			Kind:               api.ProblemSourceKindHTTPAdapter,
			Name:               a.Name,
			ObservedGeneration: a.Generation,
			Path:               fmt.Sprintf("$.spec.headers.%s", header),
			Value:              &val,
		}
		if t, err := api.NewEnvTemplate(val); err != nil {
			problems = append(problems, api.Problem{
				Type:    api.ProblemTypeParseError,
				Message: fmt.Sprintf(`Error parsing header "%s": %s`, val, err.Error()),
				Causes:  []api.ProblemSource{src},
			})
		} else {
			problems = append(problems, t.EnvSchema().Vars.Validate(data.Vars, &src)...)
			problems = append(problems, t.EnvSchema().Secrets.Validate(data.ResolvedSecrets, &src)...)
		}
	}

	return problems
}

func init() {
	SchemeBuilder.Register(&HTTPAdapter{}, &HTTPAdapterList{})
}
