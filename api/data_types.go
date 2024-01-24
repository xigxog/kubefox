// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package api

// +kubebuilder:object:generate=false
type DataProvider interface {
	Object

	GetData() *Data
	GetDataKey() DataKey
}

type DataKey struct {
	Instance  string
	Namespace string
	Kind      string
	Name      string
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

// POST	/:secret-mount-path/data/:path
// DELETE	/:secret-mount-path/metadata/:path
