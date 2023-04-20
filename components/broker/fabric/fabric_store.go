package fabric

import (
	"context"
	"fmt"
	"time"

	"github.com/xigxog/kubefox/libs/core/api/common"
	"github.com/xigxog/kubefox/libs/core/api/uri"
	"github.com/xigxog/kubefox/libs/core/grpc"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/logger"
	"github.com/xigxog/kubefox/libs/core/platform"
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

	sysCache *cache[*common.FabricSystem]
	envCache *cache[*common.FabricEnv]
}

func NewStore(brk Broker) *Store {
	return &Store{
		Broker:   brk,
		sysCache: NewCache[*common.FabricSystem](mTTL, iTTL, brk.Log()),
		envCache: NewCache[*common.FabricEnv](mTTL, iTTL, brk.Log()),
	}
}

func (store *Store) Get(ctx context.Context, evt kubefox.DataEvent) (*common.Fabric, error) {
	now := time.Now().Unix()
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

	sysIt := store.sysCache.GetItem(sysKey)
	envIt := store.envCache.GetItem(envKey)
	if sysIt == nil || envIt == nil {
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

		sysIt = &item[*common.FabricSystem]{
			key:       sysKey,
			value:     fab.System,
			immutable: sysImm,
			aTime:     now,
			cTime:     now,
		}
		envIt = &item[*common.FabricEnv]{
			key:       envKey,
			value:     fab.Env,
			immutable: envImm,
			aTime:     now,
			cTime:     now,
		}
		store.Log().Debugf("fabric retrieved from platform server; %s", sysIt)

		store.sysCache.SetItem(sysKey, sysIt)
		store.envCache.SetItem(envKey, envIt)

	} else {
		store.Log().Debugf("fabric found in cache; %s", sysIt)
	}
	sysIt.aTime = now
	envIt.aTime = now

	return &common.Fabric{
		System: sysIt.value,
		Env:    envIt.value,
	}, nil
}

func sysKey(c *grpc.EventContext) string {
	return fmt.Sprintf("%s:%s", c.System, c.App)
}

func envKey(c *grpc.EventContext) string {
	return c.Environment
}
