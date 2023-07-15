package server

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/xigxog/kubefox/libs/core/api/admin/v1alpha1"
	"github.com/xigxog/kubefox/libs/core/api/common"
	"github.com/xigxog/kubefox/libs/core/api/maker"
	"github.com/xigxog/kubefox/libs/core/api/uri"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/vault"
)

func (srv *server) ListObjs(kit kubefox.Kit) error {
	u, err := objURI(kit)
	if err != nil {
		return NotFound(kit, nil)
	}
	if u.Name() != "" && (u.SubKind() == uri.None || u.SubKind() == uri.Metadata) {
		// an unknown refType was sent
		return NotFound(kit, nil)
	}

	list, err := srv.client.Vault().List(kit, u)
	if err != nil {
		return ListErr(kit, u, err)
	}

	return Listed(kit, u, list)
}

// TODO do not allow tag of env unless config is set to immutable object
func (srv *server) CreateObj(kit kubefox.Kit) error {
	u, err := objURI(kit)
	if err != nil {
		return NotFound(kit, nil)
	}

	obj := maker.ObjectFromURI(u)
	if obj == nil {
		return NotFound(kit, u)
	}
	if err := kit.Request().UnmarshalStrict(obj); err != nil {
		return BadRequest(kit, http.StatusBadRequest, u, err)
	}

	if sys, ok := obj.(*v1alpha1.System); ok {
		obj.SetId(sys.GitHash)
	} else {
		obj.SetId(uuid.NewString())
	}

	if err := srv.client.Validate(obj); err != nil {
		return InvalidResource(kit, u, err)
	}

	if err := srv.checkRefs(kit, obj); err != nil {
		return BadRequest(kit, http.StatusBadRequest, u, err)
	}

	idURI, err := uri.New(uri.Authority, kindArg(kit), nameArg(kit), uri.Id, obj.GetId())
	if err != nil {
		return Err(kit, u, err)
	}

	if err := srv.client.Vault().Create(kit, idURI, obj); err != nil {
		return Err(kit, idURI, err)
	}

	return CreatedResp(kit, idURI, obj)
}

func (srv *server) GetObjHead(kit kubefox.Kit) error {
	u, err := objURI(kit)
	if err != nil {
		return NotFound(kit, nil)
	}

	obj := maker.ObjectFromURI(u)
	if obj == nil {
		return NotFound(kit, u)
	}

	if err := srv.client.Vault().Get(kit, u, obj); err != nil {
		return Err(kit, u, err)
	}

	return RetrievedResp(kit, u, obj)
}

func (srv *server) PutMeta(kit kubefox.Kit) error {
	return srv.updateMeta(kit, vault.Put)
}

func (srv *server) PatchMeta(kit kubefox.Kit) error {
	return srv.updateMeta(kit, vault.Patch)
}

func (srv *server) updateMeta(kit kubefox.Kit, op vault.UpdateOp) error {
	u, err := subObjURI(kit, uri.Metadata)
	if err != nil {
		return NotFound(kit, nil)
	}

	obj := maker.ObjectFromURI(u)
	if obj == nil {
		return NotFound(kit, u)
	}
	if err := kit.Request().UnmarshalStrict(obj); err != nil {
		return BadRequest(kit, http.StatusBadRequest, u, err)
	}

	meta := obj.GetMetadata()
	if meta.Name != "" && meta.Name != u.Name() {
		return InvalidResource(kit, u, fmt.Errorf("%w: %s", kubefox.ErrBadRequest, "name does not match path"))
	}
	if err := srv.client.Validate(meta); err != nil {
		return InvalidResource(kit, u, err)
	}

	switch op {
	case vault.Put:
		if err := srv.client.Vault().Put(kit, u, obj); err != nil {
			return Err(kit, u, err)
		}
		return PutResp(kit, u)

	case vault.Patch:
		if err := srv.client.Vault().Patch(kit, u, obj); err != nil {
			return Err(kit, u, err)
		}
		return PatchedResp(kit, u)

	default:
		return NotFound(kit, nil)
	}
}

func (srv *server) GetObj(kit kubefox.Kit) error {
	u, err := objURI(kit)
	if err != nil {
		return NotFound(kit, nil)
	}

	if u.SubKind() == uri.None {
		return NotFound(kit, u)
	}

	obj := maker.ObjectFromURI(u)
	if obj == nil {
		return NotFound(kit, u)
	}

	if err := srv.client.Vault().Get(kit, u, obj); err != nil {
		return Err(kit, u, err)
	}

	return RetrievedResp(kit, u, obj)
}

func (srv *server) CreateTag(kit kubefox.Kit) error {
	return srv.updateRef(kit, vault.Create)
}

func (srv *server) PutBranch(kit kubefox.Kit) error {
	return srv.updateRef(kit, vault.Put)
}

func (srv *server) updateRef(kit kubefox.Kit, op vault.UpdateOp) error {
	u, err := objURI(kit)
	if err != nil {
		return NotFound(kit, nil)
	}

	objId := string(kit.Request().GetContent())
	objURI, err := uri.New(uri.Authority, u.Kind(), u.Name(), uri.Id, objId)
	if err != nil {
		return BadRequest(kit, http.StatusBadRequest, u, err)
	}

	obj := maker.ObjectFromURI(objURI)
	if obj == nil {
		return NotFound(kit, objURI)
	}
	if err := srv.client.Vault().Get(kit, objURI, obj); err != nil {
		return Err(kit, objURI, err)
	}

	// TODO uncomment when adapters are added
	// if env, ok := obj.(*v1alpha1.Environment); ok {
	// 	cfgURI, err := uri.New(uri.Authority, uri.Config, env.Config)
	// 	if err != nil {
	// 		return Err(kit, objURI, err)
	// 	}
	// 	if cfgURI.SubKind() != uri.Id && cfgURI.SubKind() != uri.Tag {
	// 		return BadRequest(kit, http.StatusBadRequest, u, fmt.Errorf("environment must reference config by id or tag"))
	// 	}
	// }

	switch {
	case op == vault.Put && u.SubKind() == uri.Branch && u.Kind() == uri.System:
		if err := srv.client.Vault().Put(kit, u, obj); err != nil {
			return Err(kit, u, err)
		}
		return PutResp(kit, u)

	case op == vault.Create && u.SubKind() == uri.Tag:
		if err := srv.client.Vault().Create(kit, u, obj); err != nil {
			return Err(kit, u, err)
		}
		return CreatedResp(kit, u, nil)

	default:
		return NotFound(kit, objURI)
	}
}

func (srv *server) DeleteRef(kit kubefox.Kit) error {
	u, err := objURI(kit)
	if err != nil {
		return NotFound(kit, nil)
	}

	if u.SubKind() != uri.Branch && u.SubKind() != uri.Tag {
		return BadRequest(kit, http.StatusBadRequest, u, fmt.Errorf("cannot delete ref of type %s", u.SubKind()))
	}

	if err := srv.client.Vault().Delete(kit, u); err != nil {
		return Err(kit, u, err)
	}

	return DeletedResp(kit, u)
}

func (srv *server) checkRefs(kit kubefox.Kit, obj any) error {
	var err error
	var u uri.URI
	uris := []uri.URI{}

	if r, ok := obj.(common.ConfigReferrer); ok {
		u, err = uri.New(uri.Authority, uri.Config, r.GetConfig())
		uris = append(uris, u)
	}
	if r, ok := obj.(common.ConfigIdReferrer); ok {
		u, err = uri.New(uri.Authority, uri.Config, r.GetConfigId())
		uris = append(uris, u)
	}
	if r, ok := obj.(common.EnvironmentReferrer); ok {
		u, err = uri.New(uri.Authority, uri.Environment, r.GetEnvironment())
		uris = append(uris, u)
	}
	if r, ok := obj.(common.EnvironmentIdReferrer); ok {
		u, err = uri.New(uri.Authority, uri.Environment, r.GetEnvironmentId())
		uris = append(uris, u)
	}
	if r, ok := obj.(common.SystemReferrer); ok {
		u, err = uri.New(uri.Authority, uri.System, r.GetSystem())
		uris = append(uris, u)
	}
	if r, ok := obj.(common.SystemIdReferrer); ok {
		u, err = uri.New(uri.Authority, uri.System, r.GetSystemId())
		uris = append(uris, u)
	}
	if r, ok := obj.(common.InheritsReferrer); ok {
		if r.GetInherits() != "" {
			u, err = uri.New(uri.Authority, uri.Environment, r.GetInherits())
			uris = append(uris, u)
		}
	}

	if err != nil {
		return err
	}

	for _, uri := range uris {
		if err := srv.client.Vault().Exists(kit, uri); err != nil {
			return err
		}
	}

	return nil
}
