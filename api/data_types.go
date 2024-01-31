// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package api

import "github.com/xigxog/kubefox/utils"

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
}

type DataDetails struct {
	Details `json:",inline"`

	Vars    map[string]Details `json:"vars,omitempty"`
	Secrets map[string]Details `json:"secrets,omitempty"`
}

func (d DataKey) String() string {
	return utils.Join("-", d.Instance, d.Namespace, d.Kind, d.Name)
}

// Merge copies the data from the provided Data object into this Data object,
// overriding existing values.
func (d *Data) Merge(rhs *Data) {
	d.merge(rhs, true)
}

// Import copies the data from the provided Data object into this Data object,
// retaining existing values.
func (d *Data) Import(rhs *Data) {
	d.merge(rhs, false)
}

func (lhs *Data) merge(rhs *Data, override bool) {
	if lhs.Vars == nil {
		lhs.Vars = map[string]*Val{}
	}
	if lhs.Secrets == nil {
		lhs.Secrets = map[string]*Val{}
	}

	for k, v := range rhs.Vars {
		if override {
			lhs.Vars[k] = rhs.Vars[k]
		} else if _, found := lhs.Vars[k]; !found {
			lhs.Vars[k] = v
		}
	}
	for k, v := range rhs.Secrets {
		if override {
			lhs.Secrets[k] = rhs.Secrets[k]
		} else if _, found := lhs.Secrets[k]; !found {
			lhs.Secrets[k] = v
		}
	}
}
