package server

import (
	"fmt"

	"github.com/xigxog/kubefox/libs/core/api/admin/v1alpha1"
	"github.com/xigxog/kubefox/libs/core/api/common"
	"github.com/xigxog/kubefox/libs/core/api/maker"
	"github.com/xigxog/kubefox/libs/core/api/uri"
	"github.com/xigxog/kubefox/libs/core/component"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/platform"
)

type server struct {
	vaultURL string
}

func New(vaultURL string) *server {
	return &server{vaultURL: vaultURL}
}

func (srv *server) Bootstrap(kit kubefox.Kit) error {
	kit.Log().Debugf("bootstrap req; id=%s", kit.Request().GetId())

	return nil
}

func (srv *server) Weave(kit kubefox.Kit) error {
	evtCtx := kit.Request().GetContext()
	if evtCtx == nil {
		return fmt.Errorf("event is missing context")
	}
	src := kit.Request().GetSource()
	if src == nil {
		return fmt.Errorf("event is missing source component")
	}
	trg, err := component.ParseURI(kit.Request().GetArg(platform.TargetArg))
	if err != nil {
		return err
	}
	sat := kit.Request().GetArg(platform.SvcAccountTokenArg)
	if sat == "" {
		return fmt.Errorf("event is missing service account token")
	}
	envURI, err := uri.New(uri.Authority, uri.Environment, evtCtx.Environment)
	if err != nil {
		return err
	}
	sysURI, err := uri.New(uri.Authority, uri.System, evtCtx.System)
	if err != nil {
		return err
	}

	if sysURI.Name() == platform.System && evtCtx.App == platform.App {
		kit.Response().Marshal(platformFabric)
		return nil
	}

	vaultClient, err := NewVaultClient(srv.vaultURL, sat, kit)
	if err != nil {
		return err
	}

	env := maker.Empty[v1alpha1.Environment]()
	if err := vaultClient.Get(envURI, env); err != nil {
		return fmt.Errorf("error getting environment '%s': %v", envURI, err)
	}
	sys := maker.Empty[v1alpha1.System]()
	if err := vaultClient.Get(sysURI, sys); err != nil {
		return fmt.Errorf("error getting system '%s': %v", sysURI, err)
	}
	sysIdURI, err := uri.New(uri.Authority, uri.System, sys.Metadata.Name, uri.Id, sys.Id)
	if err != nil {
		return err
	}
	sysApp, ok := sys.Apps[evtCtx.App]
	if !ok {
		return fmt.Errorf("event context app '%s' not found in system '%s'", evtCtx.App, sysURI)
	}

	// TODO only the adapters used by the app should be added
	sysApp.Components[httpAdpt.Name] = httpAdpt
	sysApp.Components[k8sAdpt.Name] = k8sAdpt

	fbr := &common.Fabric{
		System: &common.FabricSystem{
			SystemIdProp: common.SystemIdProp{
				SystemId: string(sysIdURI.Key()),
			},
			SystemProp: common.SystemProp{
				System: string(sysURI.Key()),
			},
			App: common.FabricApp{
				App:  *sysApp,
				Name: evtCtx.App,
			},
		},
		Env: &common.FabricEnv{
			Config:  map[string]*common.Var{},
			Secrets: map[string]*common.Var{},
			EnvVars: env.Vars,
		},
	}
	if err := fbr.CheckComponent(src.GetName(), src.GetGitHash()); err != nil {
		return err
	}
	if err := fbr.CheckComponent(trg.GetName(), ""); err != nil {
		return err
	}
	kit.Response().Marshal(fbr)

	return nil
}
