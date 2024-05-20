// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package templates

import (
	common "github.com/xigxog/kubefox/api/kubernetes"
	"github.com/xigxog/kubefox/build"
	"github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Data struct {
	Instance  Instance
	Platform  Platform
	Component Component

	Owner []*metav1.OwnerReference

	Values map[string]any

	Telemetry common.TelemetrySpec
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
	Name       string
	Namespace  string
	BrokerAddr string
}

type Component struct {
	*core.Component      `json:",inline"`
	common.PodSpec       `json:",inline"`
	common.ContainerSpec `json:",inline"`

	Image           string
	ImagePullPolicy corev1.PullPolicy
	ImagePullSecret string

	IsPlatformComponent bool
}

type ResourceList struct {
	Items []*unstructured.Unstructured `json:"items,omitempty"`
}

func (d Data) Name() string {
	if d.Component.Name == "" {
		return ""
	}
	if d.Component.IsPlatformComponent {
		return utils.Join("-", d.Platform.Name, d.Component.Name)
	}

	return d.Component.GroupKey()
}

func (d Data) HomePath() string {
	return "/tmp/kubefox"
}
