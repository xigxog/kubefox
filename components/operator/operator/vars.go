package operator

import (
	"github.com/xigxog/kubefox/libs/core/api/common"
	"github.com/xigxog/kubefox/libs/core/platform"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	fieldOwner k8sclient.FieldOwner = "kubefox-operator"

	kvPrefix = "kfs"

	brkPolicy      = "kfp-broker-policy"
	brkRole        = "kfp-broker-role"
	platformPolicy = "kfp-platform-policy"
	platformRole   = "kfp-platform-role"

	tenYears  = "87600h"
	sharedDir = "/tmp/shared"
)

// Injected at build time
var (
	GitRef  string
	GitHash string
)

// Command flags
var (
	ContainerRegistry string = "ghcr.io/xigxog/kubefox"
)

var (
	apiSrvComp = &common.AppComponent{
		ComponentProps: common.ComponentProps{
			Name: platform.APISrvComp.GetName(),
			ComponentTypeProp: common.ComponentTypeProp{
				Type: "kubefox",
			},
			GitHashProp: common.GitHashProp{
				GitHash: platform.APISrvComp.GetGitHash(),
			},
		},
	}

	opComp = &common.AppComponent{
		ComponentProps: common.ComponentProps{
			Name: platform.OperatorComp.GetName(),
			ComponentTypeProp: common.ComponentTypeProp{
				Type: "kubefox",
			},
			GitHashProp: common.GitHashProp{
				GitHash: platform.OperatorComp.GetGitHash(),
			},
		},
	}

	httpAdpt = &common.AppComponent{
		ComponentProps: common.ComponentProps{
			Name: platform.HTTPIngressAdapt.GetName(),
			ComponentTypeProp: common.ComponentTypeProp{
				Type: "http",
			},
			GitHashProp: common.GitHashProp{
				GitHash: platform.HTTPIngressAdapt.GetGitHash(),
			},
		},
	}

	k8sAdpt = &common.AppComponent{
		ComponentProps: common.ComponentProps{
			Name: platform.K8sAdapt.GetName(),
			ComponentTypeProp: common.ComponentTypeProp{
				Type: "k8s",
			},
			GitHashProp: common.GitHashProp{
				GitHash: platform.K8sAdapt.GetGitHash(),
			},
		},
	}

	platformFabric = &common.Fabric{
		System: &common.FabricSystem{
			SystemProp: common.SystemProp{
				System: platform.System,
			},
			App: common.FabricApp{
				Name: platform.App,
				App: common.App{
					GitHashProp: common.GitHashProp{
						GitHash: platform.GitHash,
					},
					Components: map[string]*common.AppComponent{
						apiSrvComp.Name: apiSrvComp,
						opComp.Name:     opComp,
						httpAdpt.Name:   httpAdpt,
						k8sAdpt.Name:    k8sAdpt,
					},
				},
			},
		},
		Env: &common.FabricEnv{
			Config:  map[string]*common.Var{},
			Secrets: map[string]*common.Var{},
			EnvVars: map[string]*common.Var{},
		},
	}
)
