package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/xigxog/kubefox/components/api-server/client"
	srv "github.com/xigxog/kubefox/components/api-server/server"
	"github.com/xigxog/kubefox/libs/core/admin"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/utils"
)

func main() {
	var vaultURL string
	flag.StringVar(&vaultURL, "vault", "", "URL of the Vault server. Environment variable 'KUBEFOX_VAULT_URL' (default 'https://127.0.0.1:8200')")

	// parses flags
	kitSvc := kubefox.New()
	vaultURL = utils.ResolveFlag(vaultURL, "KUBEFOX_VAULT_URL", "https://127.0.0.1:8200")

	cl, err := client.New(kitSvc, vaultURL)
	if err != nil {
		kitSvc.Fatal(err)
	}

	apiSvr := srv.New(cl, kitSvc.Log())

	kitSvc.Http(rule("GET   ", "ping"), pong)

	// OBJECTS
	kitSvc.Http(rule("GET   ", "{kind}"), apiSvr.ListObjs)

	kitSvc.Http(rule("POST  ", "{kind}/{name}"), apiSvr.CreateObj)
	kitSvc.Http(rule("GET   ", "{kind}/{name}"), apiSvr.GetObjHead)
	kitSvc.Http(rule("DELETE", "{kind}/{name}"), nil)

	kitSvc.Http(rule("PUT   ", "{kind}/{name}/metadata"), apiSvr.PutMeta)
	kitSvc.Http(rule("PATCH ", "{kind}/{name}/metadata"), apiSvr.PatchMeta)

	kitSvc.Http(rule("GET   ", "{kind}/{name}/{refType}"), apiSvr.ListObjs)
	kitSvc.Http(rule("POST  ", "{kind}/{name}/{refType}/{ref}"), apiSvr.CreateTag)
	kitSvc.Http(rule("PUT   ", "{kind}/{name}/{refType}/{ref}"), apiSvr.PutBranch)
	kitSvc.Http(rule("GET   ", "{kind}/{name}/{refType}/{ref}"), apiSvr.GetObj)
	kitSvc.Http(rule("DELETE", "{kind}/{name}/{refType}/{ref}"), apiSvr.DeleteRef)

	// PLATFORM
	kitSvc.Http(rule("GET   ", "platform"), apiSvr.ListPlatforms)
	kitSvc.Http(rule("PUT   ", "platform/{name}"), apiSvr.PutPlatform)
	kitSvc.Http(rule("PATCH ", "platform/{name}"), apiSvr.PatchPlatform)
	kitSvc.Http(rule("GET   ", "platform/{name}"), apiSvr.GetPlatform)

	// DEPLOYMENTS
	kitSvc.Http(rule("GET   ", "platform/{name}/deployment"), apiSvr.ListDeployments)
	kitSvc.Http(rule("POST  ", "platform/{name}/deployment"), apiSvr.CreateDeployment)
	kitSvc.Http(rule("GET   ", "platform/{name}/deployment/{sysName}/{refType}/{ref}"), apiSvr.GetDeployment)
	kitSvc.Http(rule("DELETE", "platform/{name}/deployment/{sysName}/{refType}/{ref}"), apiSvr.DeleteDeployment)

	// RELEASES
	kitSvc.Http(rule("GET   ", "platform/{name}/release"), apiSvr.ListReleases)
	kitSvc.Http(rule("POST  ", "platform/{name}/release"), apiSvr.CreateRelease)
	kitSvc.Http(rule("GET   ", "platform/{name}/release/{sysName}/{envName}"), apiSvr.GetRelease)
	kitSvc.Http(rule("DELETE", "platform/{name}/release/{sysName}/{envName}"), apiSvr.DeleteRelease)

	kitSvc.DefaultEntrypoint(defaultEntrypoint)

	kitSvc.Start()
}

func defaultEntrypoint(kit kubefox.Kit) error {
	kit.Log().Debug("no matching entrypoint")
	return srv.NotFound(kit, nil)
}

func pong(kit kubefox.Kit) error {
	return srv.AdminResponse(kit, &admin.Response{
		IsError: false,
		Code:    http.StatusOK,
		Msg:     "Pong!",
	})
}

func rule(method, suffix string) string {
	return fmt.Sprintf("Method(`%s`) && Path(`%s/%s`)",
		strings.TrimSpace(method), admin.APIPath, strings.TrimSpace(suffix))
}
