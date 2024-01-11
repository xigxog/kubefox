// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package defaults

import (
	common "github.com/xigxog/kubefox/api/kubernetes"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	NATS = common.ContainerSpec{
		Resources: &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				"memory": resource.MustParse("64Mi"),
				"cpu":    resource.MustParse("100m"),
			},
			Limits: v1.ResourceList{
				"memory": resource.MustParse("128Mi"),
				"cpu":    resource.MustParse("2"),
			},
		},
		LivenessProbe: &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{
					Path: "/healthz?js-enabled-only=true",
					Port: intstr.FromString("monitor"),
				},
			},
			TimeoutSeconds:   3,
			PeriodSeconds:    30,
			FailureThreshold: 3,
		},
		ReadinessProbe: &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{
					Path: "/healthz?js-enabled-only=true",
					Port: intstr.FromString("monitor"),
				},
			},
			TimeoutSeconds:   3,
			PeriodSeconds:    10,
			FailureThreshold: 3,
		},
		StartupProbe: &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{
					Path: "/healthz",
					Port: intstr.FromString("monitor"),
				},
			},
			PeriodSeconds:    5,
			FailureThreshold: 90,
		},
	}

	Broker = common.ContainerSpec{
		Resources: &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				"memory": resource.MustParse("64Mi"), // 90% of limit, used to set GOMEMLIMIT
				"cpu":    resource.MustParse("100m"),
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
	}

	HTTPSrv = common.ContainerSpec{
		Resources: &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				"memory": resource.MustParse("64Mi"),
				"cpu":    resource.MustParse("100m"),
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
	}

	Component = common.ContainerSpec{
		Resources: &v1.ResourceRequirements{
			Requests: v1.ResourceList{},
			Limits:   v1.ResourceList{},
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
	}
)

func Set(cur *common.ContainerSpec, defs *common.ContainerSpec) {
	if cur.Resources == nil {
		cur.Resources = defs.Resources
	}
	if cur.LivenessProbe == nil {
		cur.LivenessProbe = defs.LivenessProbe
	}
	if cur.ReadinessProbe == nil {
		cur.ReadinessProbe = defs.ReadinessProbe
	}
	if cur.StartupProbe == nil {
		cur.StartupProbe = defs.StartupProbe
	}
}
