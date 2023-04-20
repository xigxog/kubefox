package server

import (
	"fmt"
	"net/http"

	adminv1a1 "github.com/xigxog/kubefox/libs/core/api/admin/v1alpha1"
	k "github.com/xigxog/kubefox/libs/core/api/kubernetes"
	kubev1a1 "github.com/xigxog/kubefox/libs/core/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/libs/core/api/maker"
	"github.com/xigxog/kubefox/libs/core/api/uri"
	"github.com/xigxog/kubefox/libs/core/kubefox"
)

func (srv *server) ListReleases(kit kubefox.Kit) error {
	u, err := releaseURI(kit)
	if err != nil {
		return NotFound(kit, nil)
	}

	labels := map[string]string{
		// k.OrganizationLabel: kit.Organization(),
	}
	if sysNameArg(kit) != "" {
		labels[k.SystemLabel] = sysNameArg(kit)
	}
	if envNameArg(kit) != "" {
		labels[k.EnvironmentLabel] = envNameArg(kit)
	}

	relList := maker.Empty[kubev1a1.ReleaseList]()
	if err := srv.client.Kube().List(kit, relList, labels); err != nil {
		return ListErr(kit, u, err)
	}

	list := []string{}
	for _, rel := range relList.Items {
		sys := rel.Labels[k.SystemLabel]
		env := rel.Labels[k.EnvironmentLabel]
		list = append(list, fmt.Sprintf("%s%s%s", sys, uri.PathSeparator, env))
	}

	return Listed(kit, u, list)
}

func (srv *server) CreateRelease(kit kubefox.Kit) error {
	u, err := releaseURI(kit)
	if err != nil {
		return NotFound(kit, nil)
	}

	adminRel := maker.Empty[adminv1a1.Release]()
	if err := kit.Request().UnmarshalStrict(adminRel); err != nil {
		return BadRequest(kit, http.StatusForbidden, u, err)
	}

	if err := srv.client.Validate(adminRel); err != nil {
		return InvalidResource(kit, u, err)
	}

	// can safely ignore errors as paths were validate above
	sysURI, _ := uri.New(uri.Authority, uri.System, adminRel.System)
	envURI, _ := uri.New(uri.Authority, uri.Environment, adminRel.Environment)

	compSet, err := srv.getCompSet(kit, sysURI, 0)
	if err != nil {
		return Err(kit, u, err)
	}
	if err := srv.client.Kube().Get(kit, compSet); err != nil {
		return Err(kit, u, err)
	}
	if _, ok := compSet.Spec.Deployments[sysURI.Key()]; !ok {
		u, _ := uri.New(u.Authority(), uri.Platform, u.Name(), uri.Deployment, sysURI.Key())
		return NotFound(kit, u)
	}

	env := maker.Empty[adminv1a1.Environment]()
	if err := srv.client.Vault().Get(kit, envURI, env); err != nil {
		return Err(kit, envURI, err)
	}

	// TODO uncomment when adapters are added
	// configURI, _ := uri.New(u.Authority(), uri.Config, env.Config)
	// if err != nil {
	// 	return Err(kit, nil, err)
	// }
	// config := maker.Empty[adminv1a1.Config]()
	// if err := srv.client.Vault().Get(kit, configURI, config); err != nil {
	// 	return Err(kit, configURI, err)
	// }

	sys := maker.Empty[adminv1a1.System]()
	if err := srv.client.Vault().Get(kit, sysURI, sys); err != nil {
		return Err(kit, sysURI, err)
	}

	compList := []*kubev1a1.ReleaseComponent{}
	for appName, app := range sys.Apps {
		for compName, comp := range app.Components {
			comp.Name = compName
			compList = append(compList, &kubev1a1.ReleaseComponent{
				ComponentProps: comp.ComponentProps,
				App:            appName,
				Routes:         comp.Routes,
			})
		}
	}

	curRel := maker.New[kubev1a1.Release](maker.Props{
		Name:      env.GetName(),
		Namespace: systemNamespace(kit, sysURI.Name()),
	})
	srv.client.Kube().Get(kit, curRel)

	rel := maker.New[kubev1a1.Release](maker.Props{
		Name:      env.GetName(),
		Namespace: curRel.GetNamespace(),
		// Organization:  kit.Organization(),
		Instance:      kit.Platform(),
		Environment:   envURI.Name(),
		EnvironmentId: env.GetId(),
		// TODO uncomment when adapters are added
		// Config:        configURI.Name(),
		// ConfigId:      config.GetId(),
		System:   sysURI.Name(),
		SystemId: sys.GetId(),
	})
	rel.Spec.System = string(sysURI.Key())
	rel.Spec.SystemId = sys.GetId()
	rel.Spec.Environment = string(envURI.Key())
	rel.Spec.EnvironmentId = env.GetId()
	rel.ResourceVersion = curRel.ResourceVersion
	rel.Spec.Components = compList
	if err := srv.client.Kube().Put(kit, rel); err != nil {
		return Err(kit, u, err)
	}

	return CreatedResp(kit, u, adminRel)
}

func (srv *server) GetRelease(kit kubefox.Kit) error {
	u, err := releaseURI(kit)
	if err != nil {
		return NotFound(kit, nil)
	}

	kRel := maker.New[kubev1a1.Release](maker.Props{
		Name:      envNameArg(kit),
		Namespace: systemNamespace(kit, sysNameArg(kit)),
	})
	if err := srv.client.Kube().Get(kit, kRel); err != nil {
		return Err(kit, u, err)
	}

	aRel := maker.Empty[adminv1a1.Release]()
	aRel.System = kRel.Spec.System
	aRel.Environment = kRel.Spec.Environment
	aRel.Status = kRel.Status

	return RetrievedResp(kit, u, aRel)
}

func (srv *server) DeleteRelease(kit kubefox.Kit) error {
	u, err := releaseURI(kit)
	if err != nil {
		return NotFound(kit, nil)
	}

	kRel := maker.New[kubev1a1.Release](maker.Props{
		Name:      envNameArg(kit),
		Namespace: systemNamespace(kit, sysNameArg(kit)),
	})
	if err := srv.client.Kube().Delete(kit, kRel); err != nil {
		return Err(kit, u, err)
	}

	return DeletedResp(kit, u)
}
