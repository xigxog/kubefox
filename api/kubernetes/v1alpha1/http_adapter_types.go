/*
Copyright © 2023 XigXog

This Source Code Form is subject to the terms of the Mozilla Public License,
v2.0. If a copy of the MPL was not distributed with this file, You can obtain
one at https://mozilla.org/MPL/2.0/.
*/

package v1alpha1

import (
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
	return api.ComponentTypeHTTP
}

func init() {
	SchemeBuilder.Register(&HTTPAdapter{}, &HTTPAdapterList{})
}
