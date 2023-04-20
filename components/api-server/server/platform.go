package server

import (
	"net/http"

	"github.com/xigxog/kubefox/components/api-server/client"
	adminv1a1 "github.com/xigxog/kubefox/libs/core/api/admin/v1alpha1"
	kubev1a1 "github.com/xigxog/kubefox/libs/core/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/libs/core/api/maker"
	"github.com/xigxog/kubefox/libs/core/api/uri"
	"github.com/xigxog/kubefox/libs/core/kubefox"
)

func (srv *server) ListPlatforms(kit kubefox.Kit) error {
	u, err := platformURI(kit)
	if err != nil {
		return NotFound(kit, nil)
	}

	return Listed(kit, u, []string{kit.Platform()})
}

func (srv *server) PutPlatform(kit kubefox.Kit) error {
	return srv.updatePlatform(kit, true)
}

func (srv *server) PatchPlatform(kit kubefox.Kit) error {
	return srv.updatePlatform(kit, false)
}

func (srv *server) GetPlatform(kit kubefox.Kit) error {
	u, obj, kObj, err := srv.getPlatform(kit, true)
	if err != nil {
		return Err(kit, u, err)
	}
	obj.PlatformSpec = kObj.Spec
	obj.Status = kObj.Status

	return RetrievedResp(kit, u, obj)
}

// TODO wait till things are setup before returning
func (srv *server) updatePlatform(kit kubefox.Kit, overwrite bool) error {
	u, obj, kObj, err := srv.getPlatform(kit, false)
	if err != nil {
		return Err(kit, u, err)
	}

	if err := kit.Request().UnmarshalStrict(obj); err != nil {
		return BadRequest(kit, http.StatusBadRequest, u, err)
	}
	if err := srv.client.Validate(obj); err != nil {
		return InvalidResource(kit, u, err)
	}

	kObj.Spec = obj.PlatformSpec
	if overwrite {
		if err := srv.client.Vault().Put(kit, u, obj); err != nil {
			return Err(kit, u, err)
		}
		if err := srv.client.Kube().Put(kit, kObj); err != nil {
			return Err(kit, u, err)
		}
		return PutResp(kit, u)

	} else {
		if err := srv.client.Vault().Patch(kit, u, obj); err != nil {
			return Err(kit, u, err)
		}
		if err := srv.client.Kube().Patch(kit, kObj); err != nil {
			return Err(kit, u, err)
		}
		return PatchedResp(kit, u)
	}
}

func (srv *server) getPlatform(kit kubefox.Kit, withMeta bool) (uri.URI, *adminv1a1.Platform, *kubev1a1.Platform, error) {
	u, err := platformURI(kit)
	if err != nil {
		return nil, nil, nil, err
	}

	if err := checkPlatformName(kit); err != nil {
		return u, nil, nil, err
	}

	kObj := maker.Empty[kubev1a1.Platform]()
	kObj.SetName(kit.Platform())
	if err := srv.client.Kube().Get(kit, kObj); err != nil {
		return u, nil, nil, err
	}

	// if kObj.Organization != kit.Organization() {
	// 	return u, nil, nil, client.ErrResourceForbidden
	// }

	obj := maker.Empty[adminv1a1.Platform]()
	obj.SetName(kit.Platform())
	obj.SetId(string(kObj.UID))
	// obj.SetOrganization(kObj.Organization)
	if withMeta {
		if err := srv.client.Vault().Get(kit, u, obj); err != nil {
			if !client.IsNotFound(err) {
				return u, nil, nil, err
			}
		}
	}

	return u, obj, kObj, nil
}
