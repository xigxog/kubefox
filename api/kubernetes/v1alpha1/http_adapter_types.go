// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

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

// +kubebuilder:object:generate=false
type HTTPAdapterTemplate struct {
	URL     *api.EnvTemplate
	Headers map[string]*api.EnvTemplate
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=httpadapters,shortName=http
// +kubebuilder:printcolumn:name="URL",type=string,JSONPath=`.spec.url`
// +kubebuilder:printcolumn:name="Insecure",type=boolean,JSONPath=`.spec.insecureSkipVerify`

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
	tpl := a.getTemplate()
	var problems api.Problems

	src := api.ProblemSource{
		Kind:               api.ProblemSourceKindHTTPAdapter,
		Name:               a.Name,
		ObservedGeneration: a.Generation,
		Path:               "$.spec.url",
		Value:              &a.Spec.URL,
	}
	if t := tpl.URL; t.ParseError() != nil {
		problems = append(problems, api.Problem{
			Type:    api.ProblemTypeParseError,
			Message: fmt.Sprintf(`Error parsing template "%s": %s`, a.Spec.URL, t.ParseError()),
			Causes:  []api.ProblemSource{src},
		})
	} else {
		problems = append(problems, t.EnvSchema().Validate(data, &src, false)...)
	}

	for header, val := range a.Spec.Headers {
		v := val
		src := api.ProblemSource{
			Kind:               api.ProblemSourceKindHTTPAdapter,
			Name:               a.Name,
			ObservedGeneration: a.Generation,
			Path:               fmt.Sprintf("$.spec.headers.%s", header),
			Value:              &v,
		}
		if t := tpl.Headers[header]; t.ParseError() != nil {
			problems = append(problems, api.Problem{
				Type:    api.ProblemTypeParseError,
				Message: fmt.Sprintf(`Error parsing template "%s": %s`, val, t.ParseError()),
				Causes:  []api.ProblemSource{src},
			})
		} else {
			problems = append(problems, t.EnvSchema().Validate(data, &src, false)...)
		}
	}

	return problems
}

func (a *HTTPAdapter) Resolve(data *api.Data) error {
	tpl := a.getTemplate()

	url, err := tpl.URL.Resolve(data, true)
	if err != nil {
		return err
	}

	headers := make(map[string]string, len(tpl.Headers))
	for k, v := range tpl.Headers {
		if headers[k], err = v.Resolve(data, true); err != nil {
			return err
		}
	}

	a.Spec.URL = url
	a.Spec.Headers = headers

	return nil
}

func (a *HTTPAdapter) getTemplate() *HTTPAdapterTemplate {
	tpl := &HTTPAdapterTemplate{
		URL:     api.NewEnvTemplate("url", a.Spec.URL),
		Headers: make(map[string]*api.EnvTemplate, len(a.Spec.Headers)),
	}
	for k, v := range a.Spec.Headers {
		tpl.Headers[k] = api.NewEnvTemplate("header."+k, v)
	}

	return tpl
}

func init() {
	SchemeBuilder.Register(&HTTPAdapter{}, &HTTPAdapterList{})
}
