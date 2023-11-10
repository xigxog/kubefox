package controller

import (
	"github.com/xigxog/kubefox/api/kubernetes"
	"github.com/xigxog/kubefox/components/operator/templates"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	NATSDefaults = templates.Component{
		ContainerSpec: kubernetes.ContainerSpec{
			Resources: &v1.ResourceRequirements{
				// TODO calc and set correct values and use those in headers
				Requests: v1.ResourceList{
					"memory": resource.MustParse("115Mi"), // 90% of limit, used to set GOMEMLIMIT
					"cpu":    resource.MustParse("250m"),
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
		},
	}

	BrokerDefaults = templates.Component{
		ContainerSpec: kubernetes.ContainerSpec{
			Resources: &v1.ResourceRequirements{
				// TODO calc and set correct values and use those in headers
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
	}

	HTTPSrvDefaults = templates.Component{
		ContainerSpec: kubernetes.ContainerSpec{
			Resources: &v1.ResourceRequirements{
				// TODO calc and set correct values and use those in headers
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
	}

	ComponentDefaults = templates.Component{
		ContainerSpec: kubernetes.ContainerSpec{
			// Resources: &v1.ResourceRequirements{
			// 	// TODO calc and set correct values and use those in headers
			// 	Requests: v1.ResourceList{
			// 		"memory": resource.MustParse("144Mi"), // 90% of limit, used to set GOMEMLIMIT
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
		},
	}
)
