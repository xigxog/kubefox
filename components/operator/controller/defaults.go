package controller

import (
	common "github.com/xigxog/kubefox/api/kubernetes"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	NATSDefaults = common.ContainerSpec{
		Resources: &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				"memory": resource.MustParse("64Mi"),
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

	BrokerDefaults = common.ContainerSpec{
		Resources: &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				"memory": resource.MustParse("64Mi"), // 90% of limit, used to set GOMEMLIMIT
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

	HTTPSrvDefaults = common.ContainerSpec{
		Resources: &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				"memory": resource.MustParse("64Mi"),
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

	ComponentDefaults = common.ContainerSpec{
		// Resources: &v1.ResourceRequirements{
		// 	Requests: v1.ResourceList{
		// 		"memory": resource.MustParse("144Mi"),
		// 		"cpu":    resource.MustParse("250m"),
		// 	},
		// 	Limits: v1.ResourceList{
		// 		"memory": resource.MustParse("160Mi"),
		// 		"cpu":    resource.MustParse("2"),
		// 	},
		// },
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

func SetDefaults(cur *common.ContainerSpec, defs *common.ContainerSpec) {
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
