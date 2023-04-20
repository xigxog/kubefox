package kubefox

import (
	"context"

	"github.com/xigxog/kubefox/libs/core/api/common"
	"github.com/xigxog/kubefox/libs/core/grpc"
	"github.com/xigxog/kubefox/libs/core/logger"
)

type Kit interface {
	Env(string) *common.Var

	Request() Event
	Response() Event

	Component(string) *ComponentSvc

	// Organization() string
	Platform() string
	Namespace() string
	DevMode() bool

	Ctx() context.Context
	Log() *logger.Log
}

type kit struct {
	req  DataEvent
	resp DataEvent

	kitSvc KitSvc
	broker kitBroker

	ctx context.Context
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

func (kit *kit) Namespace() string {
	return kit.kitSvc.Namespace()
}

func (kit *kit) DevMode() bool {
	return kit.kitSvc.DevMode()
}

func (kit *kit) Ctx() context.Context {
	return kit.ctx
}

func (kit *kit) Log() *logger.Log {
	return kit.log
}
