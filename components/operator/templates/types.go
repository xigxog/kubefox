// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package templates

import (
	"fmt"
	"strings"

	common "github.com/xigxog/kubefox/api/kubernetes"
	"github.com/xigxog/kubefox/build"
	"github.com/xigxog/kubefox/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Data struct {
	Instance  Instance
	Platform  Platform
	Component Component

	Owner []*metav1.OwnerReference

	Values map[string]any

	Logger    common.LoggerSpec
	BuildInfo build.BuildInfo `hash:"ignore"`

	Hash string `hash:"ignore"`
}

type Instance struct {
	Name           string
	Namespace      string
	RootCA         string
	BootstrapImage string
}

type Platform struct {
	Name      string
	Namespace string
}

type Component struct {
	*core.Component      `json:",inline"`
	common.PodSpec       `json:",inline"`
	common.ContainerSpec `json:",inline"`

	Image           string
	ImagePullPolicy string
	ImagePullSecret string

	IsPlatformComponent bool
}

type ResourceList struct {
	Items []*unstructured.Unstructured `json:"items,omitempty"`
}

func (d Data) Namespace() string {
	if d.Platform.Namespace != "" {
		return d.Platform.Namespace
	}
	return d.Instance.Namespace
}

func (d Data) PlatformFullName() string {
	if d.Platform.Name == "" {
		return d.Instance.Name
	}
	if strings.HasPrefix(d.Platform.Name, d.Instance.Name) {
		return d.Platform.Name
	}
	return fmt.Sprintf("%s-%s", d.Instance.Name, d.Platform.Name)
}

func (d Data) ComponentFullName() string {
	if d.Component.Name == "" {
		return ""
	}

	name := d.Component.GroupKey()
	if d.Component.IsPlatformComponent {
		name = d.Component.Name
	}

	return name
}

func (d Data) HomePath() string {
	return "/tmp/kubefox"
}
