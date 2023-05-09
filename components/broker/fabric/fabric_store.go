package fabric

import (
	"context"
	"fmt"

	"github.com/xigxog/kubefox/libs/core/api/common"
	"github.com/xigxog/kubefox/libs/core/api/uri"
	"github.com/xigxog/kubefox/libs/core/grpc"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/logger"
	"github.com/xigxog/kubefox/libs/core/platform"
	"github.com/xigxog/kubefox/libs/core/utils"
)

var (
	mTTL = int64(30)          // 30 seconds timeout for mutable items
	iTTL = int64(6 * 60 * 60) // 6 hours time for immutable items
)

type Broker interface {
	Log() *logger.Log
	InvokeRuntimeServer(context.Context, kubefox.DataEvent) kubefox.DataEvent
}

type Store struct {
	Broker

	sysCache *utils.Cache[*common.FabricSystem]
	envCache *utils.Cache[*common.FabricEnv]
}

func NewStore(brk Broker) *Store {
	return &Store{
		Broker:   brk,
		sysCache: utils.NewCache[*common.FabricSystem](mTTL, iTTL, brk.Log()),
		envCache: utils.NewCache[*common.FabricEnv](mTTL, iTTL, brk.Log()),
	}
}

func (store *Store) Get(ctx context.Context, evt kubefox.DataEvent) (*common.Fabric, error) {
	sysKey := sysKey(evt.GetContext())
	envKey := envKey(evt.GetContext())

	// evt.GetContext().Organization
	sysURI, err := uri.New(uri.Authority, uri.System, evt.GetContext().System)
	if err != nil {
		return nil, err
	}
	// evt.GetContext().Organization
	envURI, err := uri.New(uri.Authority, uri.Environment, evt.GetContext().Environment)
	if err != nil {
		return nil, err
	}
	sysImm := sysURI.SubKind() == uri.Id || sysURI.SubKind() == uri.Tag
	envImm := envURI.SubKind() == uri.Id || envURI.SubKind() == uri.Tag

	sys := store.sysCache.Get(sysKey)
	env := store.envCache.Get(envKey)
	if sys == nil || env == nil {
		req := evt.ChildEvent()
		req.SetType(kubefox.FabricRequestType)
		req.SetArg(platform.TargetArg, evt.GetTarget().GetURI())

		resp := store.InvokeRuntimeServer(ctx, req)
		if resp.GetError() != nil {
			return nil, resp.GetError()
		}

		fab := &common.Fabric{}
		if err := resp.Unmarshal(fab); err != nil {
			return nil, err
		}

		sys = fab.System
		env = fab.Env

		store.sysCache.Set(sysKey, sys, sysImm)
		store.envCache.Set(envKey, env, envImm)
		store.Log().Debugf("fabric retrieved from runtime server; %s", sys)

	} else {
		store.Log().Debugf("fabric found in cache; %s", sys)
	}

	return &common.Fabric{
		System: sys,
		Env:    env,
	}, nil
}

func sysKey(c *grpc.EventContext) string {
	return fmt.Sprintf("%s:%s", c.System, c.App)
}

func envKey(c *grpc.EventContext) string {
	return c.Environment
}
