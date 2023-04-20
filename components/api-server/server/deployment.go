package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/xigxog/kubefox/components/api-server/client"
	adminv1a1 "github.com/xigxog/kubefox/libs/core/api/admin/v1alpha1"
	"github.com/xigxog/kubefox/libs/core/api/common"
	k "github.com/xigxog/kubefox/libs/core/api/kubernetes"
	kubev1a1 "github.com/xigxog/kubefox/libs/core/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/libs/core/api/maker"
	"github.com/xigxog/kubefox/libs/core/api/uri"
	"github.com/xigxog/kubefox/libs/core/kubefox"
)

func (srv *server) ListDeployments(kit kubefox.Kit) error {
	u, err := deploymentURI(kit)
	if err != nil {
		return NotFound(kit, nil)
	}

	labels := map[string]string{}
	if sysNameArg(kit) != "" {
		labels[k.SystemLabel] = sysNameArg(kit)
	}

	results := maker.Empty[kubev1a1.ComponentSetList]()
	if err := srv.client.Kube().List(kit, results, labels); err != nil {
		return ListErr(kit, u, err)
	}

	list := []string{}
	for _, compSet := range results.Items {
		for k := range compSet.Spec.Deployments {
			list = append(list, string(k))
		}
	}

	return Listed(kit, u, list)
}

func (srv *server) CreateDeployment(kit kubefox.Kit) error {
	depURI, sysURI, err := srv.getDeploymentURIs(kit)
	if err != nil {
		return BadRequest(kit, http.StatusBadRequest, depURI, err)
	}

	curCompSet, err := srv.getCompSet(kit, sysURI, 0)
	if err != nil {
		return Err(kit, sysURI, err)
	}

	// only branches can be re-deployed
	if sysURI.SubKind() != uri.Branch && curCompSet.Spec.Deployments[sysURI.Key()] != nil {
		u, _ := uri.New(depURI.Authority(), depURI.Kind(), depURI.Name(), uri.Deployment, sysURI.SubPathWithKind())
		return Conflict(kit, u)
	}

	sysObj := maker.Empty[adminv1a1.System]()
	if err := srv.client.Vault().Get(kit, sysURI, sysObj); err != nil {
		return Err(kit, depURI, err)
	}

	// gather Components from Apps, map is used to ensure no dupe Components
	compMap := map[common.ComponentKey]*common.ComponentProps{}
	for _, app := range sysObj.Apps {
		for compName, comp := range app.Components {
			// ensure name is set so Key() is correct
			comp.Name = compName
			if _, ok := compMap[comp.Key()]; !ok {
				compMap[comp.Key()] = &comp.ComponentProps
			}
		}
	}
	compList := []*common.ComponentProps{}
	for _, comp := range compMap {
		compList = append(compList, comp)
	}

	compSet := maker.New[kubev1a1.ComponentSet](maker.Props{
		Name:      curCompSet.Name,
		Namespace: curCompSet.Namespace,
	})
	compSet.Spec.Deployments = map[uri.Key]*kubev1a1.Deployment{
		sysURI.Key(): {
			Components: compList,
		},
	}

	if err := srv.client.Kube().Patch(kit, compSet); err != nil {
		return Err(kit, depURI, err)
	}

	return CreatedResp(kit, depURI, buildDeployment(sysURI, curCompSet, compList))
}

func (srv *server) GetDeployment(kit kubefox.Kit) error {
	depURI, sysURI, err := srv.getDeploymentURIs(kit)
	if err != nil {
		return Err(kit, nil, err)
	}

	compSet, err := srv.getCompSet(kit, sysURI, 0)
	if err != nil {
		return Err(kit, depURI, err)
	}

	sys := compSet.Spec.Deployments[sysURI.Key()]
	if sys == nil {
		return NotFound(kit, depURI)
	}

	return RetrievedResp(kit, depURI, buildDeployment(sysURI, compSet, sys.Components))
}

func (srv *server) DeleteDeployment(kit kubefox.Kit) error {
	depURI, sysURI, err := srv.getDeploymentURIs(kit)
	if err != nil {
		return Err(kit, nil, err)
	}

	compSet, err := srv.getCompSet(kit, sysURI, 0)
	if err != nil {
		return Err(kit, depURI, err)
	}

	sys := compSet.Spec.Deployments[sysURI.Key()]
	if sys == nil {
		return NotFound(kit, depURI)
	}

	// reset for patch
	compSet.Spec.Deployments = map[uri.Key]*kubev1a1.Deployment{
		sysURI.Key(): nil,
	}
	if err := srv.client.Kube().Patch(kit, compSet); err != nil {
		return Err(kit, depURI, err)
	}

	return DeletedResp(kit, depURI)
}

func (srv *server) getDeploymentURIs(kit kubefox.Kit) (depURI uri.URI, sysURI uri.URI, err error) {
	depURI, err = deploymentURI(kit)
	if err != nil {
		return
	}

	if sysNameArg(kit) == "" {
		// no sys name arg means create, parse deployment from body
		d := maker.Empty[adminv1a1.Deployment]()
		if err = kit.Request().UnmarshalStrict(d); err != nil {
			return
		}
		sysURI, err = uri.New(uri.Authority, uri.System, d.System)
		if err != nil {
			return
		}

	} else {
		sysURI, err = systemURI(kit)
		if err != nil {
			return
		}
	}

	return
}

func (srv *server) getCompSet(kit kubefox.Kit, sysURI uri.URI, attempt uint8) (*kubev1a1.ComponentSet, error) {
	compSet := maker.New[kubev1a1.ComponentSet](maker.Props{
		Name:      sysURI.Name(),
		Namespace: systemNamespace(kit, sysURI.Name()),
	})
	if err := srv.client.Kube().Get(kit, compSet); err != nil {
		msg := fmt.Sprintf("resource %s/ComponentSet/%s not found", compSet.Namespace, compSet.Name)
		if client.IsNotFound(err) {
			// TODO change to look at platform and then perform wait on k8s res
			// retry is to wait for operator to setup componentset
			if attempt < 2 {
				srv.log.Warnf("%s, retrying", msg)
				time.Sleep(time.Second)
				return srv.getCompSet(kit, sysURI, attempt+1)
			} else {
				return nil, fmt.Errorf(msg)
			}
		}
		return nil, err
	}

	return compSet, nil
}

func buildDeployment(sysURI uri.URI, compSet *kubev1a1.ComponentSet, compList []*common.ComponentProps) *adminv1a1.Deployment {
	dep := maker.Empty[adminv1a1.Deployment]()
	dep.System = string(sysURI.Key())

	ready := true
	if status := compSet.Status.Deployments[sysURI.Key()]; status != nil {
		ready = status.Ready
	}

	compStatus := map[common.ComponentKey]*common.ComponentStatus{}
	for _, comp := range compList {
		status := &common.ComponentStatus{}
		if curStatus := compSet.Status.Components[comp.Key()]; curStatus != nil {
			status.Ready = curStatus.Ready
		}
		ready = ready && status.Ready
		compStatus[comp.Key()] = status
	}

	dep.Status = adminv1a1.DeploymentStatus{
		Ready:      ready,
		Components: compStatus,
	}

	return dep
}
