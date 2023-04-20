package server

import (
	"github.com/xigxog/kubefox/libs/core/api/common"
	"github.com/xigxog/kubefox/libs/core/platform"
)

// TODO generate this as yaml during build so it includes correct git hashes

// Platform components
var apiSrvComp = &common.AppComponent{
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
var runtimeSrvComp = &common.AppComponent{
	ComponentProps: common.ComponentProps{
		Name: platform.RuntimeSrvComp.GetName(),
		ComponentTypeProp: common.ComponentTypeProp{
			Type: "kubefox",
		},
		GitHashProp: common.GitHashProp{
			GitHash: platform.RuntimeSrvComp.GetGitHash(),
		},
	},
}
var opComp = &common.AppComponent{
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

// Platform adapters
var httpAdpt = &common.AppComponent{
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
var k8sAdpt = &common.AppComponent{
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

var platformFabric = &common.Fabric{
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
					apiSrvComp.Name:     apiSrvComp,
					runtimeSrvComp.Name: runtimeSrvComp,
					opComp.Name:         opComp,
					httpAdpt.Name:       httpAdpt,
					k8sAdpt.Name:        k8sAdpt,
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
