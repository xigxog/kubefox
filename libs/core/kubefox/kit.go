package kubefox

import (
	"context"

	"github.com/xigxog/kubefox/libs/core/api/common"
	"github.com/xigxog/kubefox/libs/core/grpc"
	"github.com/xigxog/kubefox/libs/core/logger"
)

type Kit interface {
	KitContext

	Env(string) *common.Var

	Request() Event
	Response() Event

	Component(string) *ComponentSvc
}

type kit struct {
	context.Context

	req  DataEvent
	resp DataEvent

	kitSvc KitSvc
	broker kitBroker

	log *logger.Log
}

func (kit *kit) Env(key string) *common.Var {
	v, _ := common.VarFromValue(kit.req.GetFabric().EnvVars[key])
	return v
}

func (kit *kit) Request() Event {
	return kit.req
}

func (kit *kit) Response() Event {
	return kit.resp
}

func (kit *kit) Component(name string) *ComponentSvc {
	return &ComponentSvc{
		comp: &grpc.Component{Name: name},
		kit:  kit,
	}
}

// func (kit *kit) Organization() string {
// 	return kit.kitSvc.Organization()
// }

func (kit *kit) Platform() string {
	return kit.kitSvc.Platform()
}

func (kit *kit) PlatformNamespace() string {
	return kit.kitSvc.PlatformNamespace()
}

func (kit *kit) CACertPath() string {
	return kit.kitSvc.CACertPath()
}

func (kit *kit) DevMode() bool {
	return kit.kitSvc.DevMode()
}

func (kit *kit) Log() *logger.Log {
	return kit.log
}
