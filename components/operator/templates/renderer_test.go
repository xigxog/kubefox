// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package templates

import (
	"testing"

	"github.com/xigxog/kubefox/api"
	common "github.com/xigxog/kubefox/api/kubernetes"
	"github.com/xigxog/kubefox/build"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestRenderInstace(t *testing.T) {
	d := &Data{
		Instance: Instance{
			Name:      "kubefox",
			Namespace: "kubefox-system",
		},
		Values: map[string]any{
			"caBundle": "abcdef",
		},
		BuildInfo: build.Info,
	}
	if s, err := renderStr("list.tpl", "instance/*", d); err != nil {
		t.Errorf("%v", err)
	} else {
		t.Logf("\n%s", s)
	}
}
func TestRenderPlatform(t *testing.T) {
	d := &Data{
		Instance: Instance{
			Name:      "kubefox",
			Namespace: "kubefox-system",
		},
		Platform: Platform{
			Name:      "dev",
			Namespace: "kubefox-platform",
		},
		Owner: []*metav1.OwnerReference{
			{
				APIVersion: "kubefox.xigxog.io/v1alpha1",
				Kind:       "Platform",
				UID:        "123",
				Name:       "kubefox-dev",
			},
		},
	}
	if s, err := renderStr("list.tpl", "platform/*", d); err != nil {
		t.Errorf("%v", err)
	} else {
		t.Logf("\n%s", s)
	}
}

func TestRenderNATS(t *testing.T) {
	d := &Data{
		Hash: "123",
		Instance: Instance{
			Name:      "kubefox",
			Namespace: "kubefox-system",
		},
		Platform: Platform{
			Name:      "dev",
			Namespace: "kubefox-platform",
		},
		Component: Component{
			IsPlatformComponent: true,
			Name:                "nats",
			Type:                api.ComponentTypeNATS,
			Image:               "nats:2.9.21-alpine",
			PodSpec: common.PodSpec{
				Labels: map[string]string{
					"test":    "test",
					"badval1": "-aaaa",
					"badval2": "a&&aaa%a%a%-",
					"badval3": "Aaaa-",
				},
				Tolerations: []v1.Toleration{
					{
						Key: "asdf",
					},
				},
				Affinity: &v1.Affinity{
					NodeAffinity: &v1.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
							NodeSelectorTerms: []v1.NodeSelectorTerm{
								{},
							},
						},
					},
				},
			},
			ContainerSpec: common.ContainerSpec{
				Resources: &v1.ResourceRequirements{
					Limits: v1.ResourceList{
						"memory": resource.MustParse("128Mi"),
						"cpu":    resource.MustParse("1000m"),
					},
				},
				LivenessProbe: &v1.Probe{
					TimeoutSeconds: 1,
				},
			},
		},
		Owner: []*metav1.OwnerReference{
			{
				APIVersion: "kubefox.xigxog.io/v1alpha1",
				Kind:       "Platform",
				UID:        "123",
				Name:       "kubefox-dev",
			},
		},
		Logger: common.LoggerSpec{
			Level: "debug",
		},
		Values: map[string]any{
			api.ValKeyMaxEventSize: api.DefaultMaxEventSizeBytes,
		},
	}
	if s, err := renderStr("list.tpl", "nats/*", d); err != nil {
		t.Errorf("%v", err)
	} else {
		t.Logf("\n%s", s)
	}
}

func TestRenderBroker(t *testing.T) {
	d := &Data{
		Instance: Instance{
			Name:      "kubefox",
			Namespace: "kubefox-system",
		},
		Platform: Platform{
			Name:      "dev",
			Namespace: "kubefox-platform",
		},
		Component: Component{
			// IsPlatformComponent: true,
			Name:   "broker",
			Type:   api.ComponentTypeBroker,
			Image:  "ghcr.io/xigxog/kubefox/broker:v0.0.1",
			Commit: "aaaaaaa",
		},
		Owner: []*metav1.OwnerReference{
			{
				APIVersion: "kubefox.xigxog.io/v1alpha1",
				Kind:       "Platform",
				UID:        "123",
				Name:       "kubefox-dev",
			},
		},
	}
	if s, err := renderStr("list.tpl", "broker/*", d); err != nil {
		t.Errorf("%v", err)
	} else {
		t.Logf("\n%s", s)
	}
}

func TestRenderHTTPSrv(t *testing.T) {
	d := &Data{
		Values: map[string]any{
			"serviceLabels": map[string]string{
				"test": "test",
				"bad":  "i'm a bad^labl",
			},
			"serviceAnnotations": map[string]string{
				"test": "test",
			},
			"serviceType": "",
			"httpPort":    0,
			"httpsPort":   0,
		},
		Instance: Instance{
			Name:      "kubefox",
			Namespace: "kubefox-system",
		},
		Platform: Platform{
			Name:      "dev",
			Namespace: "kubefox-platform",
		},
		Component: Component{
			Name:  "httpsrv",
			Type:  api.ComponentTypeHTTPAdapter,
			Image: "ghcr.io/xigxog/kubefox/httpsrv:v0.0.1",
			ContainerSpec: common.ContainerSpec{
				Resources: &v1.ResourceRequirements{
					Requests: v1.ResourceList{
						"memory": resource.MustParse("144Mi"), // 90% of limit, used to set GOMEMLIMIT
						"cpu":    resource.MustParse("250m"),
					},
					Limits: v1.ResourceList{
						"memory": resource.MustParse("160Mi"),
						"cpu":    resource.MustParse("2"),
					},
				},
				LivenessProbe: &v1.Probe{
					ProbeHandler: v1.ProbeHandler{
						HTTPGet: &v1.HTTPGetAction{
							Port: intstr.FromString("health"),
						},
					},
				},
				ReadinessProbe: &v1.Probe{
					ProbeHandler: v1.ProbeHandler{
						HTTPGet: &v1.HTTPGetAction{
							Port: intstr.FromString("health"),
						},
					},
				},
			},
		},
		Owner: []*metav1.OwnerReference{
			{
				APIVersion: "kubefox.xigxog.io/v1alpha1",
				Kind:       "Platform",
				UID:        "123",
				Name:       "kubefox-dev",
			},
		},
	}
	if s, err := renderStr("list.tpl", "httpsrv/*", d); err != nil {
		t.Errorf("%v", err)
	} else {
		t.Logf("\n%s", s)
	}
}

func TestRenderComponent(t *testing.T) {
	d := &Data{
		Instance: Instance{
			Name:      "kubefox",
			Namespace: "kubefox-system",
		},
		Platform: Platform{
			Name:      "dev",
			Namespace: "kubefox-platform",
		},
		Component: Component{
			Name:  "hello",
			Type:  api.ComponentTypeKubeFox,
			Image: "ghcr.io/xigxog/kubefox/hello:v0.0.1",
		},
		Owner: []*metav1.OwnerReference{
			{
				APIVersion: "kubefox.xigxog.io/v1alpha1",
				Kind:       "Platform",
				UID:        "123",
				Name:       "kubefox-dev",
			},
			{
				APIVersion: "kubefox.xigxog.io/v1alpha1",
				Kind:       "Deployment",
				UID:        "123",
				Name:       "kubefox-dev",
			},
		},
	}
	if s, err := renderStr("list.tpl", "component/*", d); err != nil {
		t.Errorf("%v", err)
	} else {
		t.Logf("\n%s", s)
	}
}
