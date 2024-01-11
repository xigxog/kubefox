// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package api

import "fmt"

// +kubebuilder:object:generate=false
type DataProvider interface {
	GetData() *Data
	GetDataKey() DataKey
}

type DataKey struct {
	Kind      string
	Name      string
	Namespace string
}

type Data struct {
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Vars map[string]*Val `json:"vars,omitempty"`
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Secrets map[string]*Val `json:"secrets,omitempty"`

	ResolvedSecrets map[string]*Val `json:"-"`
}

type DataDetails struct {
	Details `json:",inline"`

	Vars    map[string]Details `json:"vars,omitempty"`
	Secrets map[string]Details `json:"secrets,omitempty"`
}

func (lhs Data) MergeInto(rhs *Data) *Data {
	rhs = rhs.DeepCopy()
	if rhs.Vars == nil {
		rhs.Vars = map[string]*Val{}
	}
	if rhs.Secrets == nil {
		rhs.Secrets = map[string]*Val{}
	}

	for k, v := range lhs.Vars {
		rhs.Vars[k] = v
	}
	for k, v := range lhs.Secrets {
		rhs.Secrets[k] = v
	}

	return rhs
}

func (k DataKey) Path(instance string) string {
	if k.Namespace == "" {
		return fmt.Sprintf("kubefox/instance/%s/cluster/data/%s/%s",
			instance, k.Kind, k.Name)
	} else {
		return fmt.Sprintf("kubefox/instance/%s/namespace/%s/data/%s/%s",
			instance, k.Namespace, k.Kind, k.Name)
	}
}
