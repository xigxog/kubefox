package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/xigxog/kubefox/components/api-server/client"
	sdkadmin "github.com/xigxog/kubefox/libs/core/admin"
	"github.com/xigxog/kubefox/libs/core/api/uri"
	"github.com/xigxog/kubefox/libs/core/kubefox"
)

func Err(kit kubefox.Kit, u uri.URI, err error) error {
	if client.IsNotFound(err) {
		return NotFound(kit, u)
	}
	if client.IsConflict(err) {
		return Conflict(kit, u)
	}
	if errors.Is(err, client.ErrBadRequest) {
		return BadRequest(kit, http.StatusBadRequest, u, err)
	}
	if errors.Is(err, client.ErrResourceForbidden) {
		return BadRequest(kit, http.StatusForbidden, u, err)
	}

	return UnknownErr(kit, u, err)
}

func BadRequest(kit kubefox.Kit, status int, u uri.URI, err error) error {
	return AdminResponse(kit, &sdkadmin.Response{
		IsError: true,
		Code:    status,
		Msg:     fmt.Sprintf("bad request for %s: %v", u, err),
	})
}

func NotFound(kit kubefox.Kit, u uri.URI) error {
	msg := "not found"
	if u != nil {
		msg = fmt.Sprintf("%s %s", u, msg)
	}
	return AdminResponse(kit, &sdkadmin.Response{
		IsError: true,
		Code:    http.StatusNotFound,
		Msg:     msg,
	})
}

func Conflict(kit kubefox.Kit, u uri.URI) error {
	return AdminResponse(kit, &sdkadmin.Response{
		IsError: true,
		Code:    http.StatusConflict,
		Msg:     fmt.Sprintf("%s exists", u),
	})
}

func InvalidResource(kit kubefox.Kit, u uri.URI, err any) error {
	kit.Log().DebugInterface(err, "%s is invalid, errors:", u)

	return AdminResponse(kit, &sdkadmin.Response{
		IsError:          true,
		Code:             http.StatusBadRequest,
		Msg:              fmt.Sprintf("%s is invalid", u),
		ValidationErrors: err,
	})
}

func Required(kit kubefox.Kit, u uri.URI, ownerId fmt.Stringer) error {
	return AdminResponse(kit, &sdkadmin.Response{
		IsError: true,
		Code:    http.StatusBadRequest,
		Msg:     fmt.Sprintf("cannot delete %s: %s requires %s", u, ownerId, u),
	})
}

func ListErr(kit kubefox.Kit, u uri.URI, err error) error {
	kit.Log().Errorf("unexpected error: %v", err)

	return AdminResponse(kit, &sdkadmin.Response{
		IsError: true,
		Code:    http.StatusInternalServerError,
		Msg:     fmt.Sprintf("error listing %s: %v", u, err),
	})
}

func UnknownErr(kit kubefox.Kit, u uri.URI, err error) error {
	kit.Log().Errorf("unexpected error: %v", err)

	return AdminResponse(kit, &sdkadmin.Response{
		IsError: true,
		Code:    http.StatusInternalServerError,
		Msg:     fmt.Sprintf("error processing %s: %v", u, err),
	})
}

func CreatedResp(kit kubefox.Kit, u uri.URI, obj any) error {
	return AdminResponse(kit, &sdkadmin.Response{
		IsError: false,
		Code:    http.StatusCreated,
		Msg:     fmt.Sprintf("created %s", u),
		Data:    obj,
	})
}

func PutResp(kit kubefox.Kit, u uri.URI) error {
	return AdminResponse(kit, &sdkadmin.Response{
		IsError: false,
		Code:    http.StatusOK,
		Msg:     fmt.Sprintf("put %s", u),
	})
}

func PatchedResp(kit kubefox.Kit, u uri.URI) error {
	return AdminResponse(kit, &sdkadmin.Response{
		IsError: false,
		Code:    http.StatusOK,
		Msg:     fmt.Sprintf("patched %s", u),
	})
}

func RetrievedResp(kit kubefox.Kit, u uri.URI, obj any) error {
	return AdminResponse(kit, &sdkadmin.Response{
		IsError: false,
		Code:    http.StatusOK,
		Msg:     fmt.Sprintf("got %s", u),
		Data:    obj,
	})
}

func Listed[T any](kit kubefox.Kit, u uri.URI, list []T) error {
	return AdminResponse(kit, &sdkadmin.Response{
		IsError: false,
		Code:    http.StatusOK,
		Msg:     fmt.Sprintf("listed %d %s", len(list), u),
		Data:    list,
	})
}

func DeletedResp(kit kubefox.Kit, u uri.URI) error {
	return AdminResponse(kit, &sdkadmin.Response{
		IsError: false,
		Code:    http.StatusOK,
		Msg:     fmt.Sprintf("deleted %s", u),
	})
}

func AdminResponse(kit kubefox.Kit, content *sdkadmin.Response) error {
	kit.Log().Debug(content.Msg)

	content.TraceId = kit.Request().GetSpan().TraceId

	resp := kit.Response().HTTP()
	resp.SetStatusCode(content.Code)
	return resp.Marshal(content)
}
