package server

import (
	"github.com/xigxog/kubefox/components/api-server/client"
	"github.com/xigxog/kubefox/libs/core/api/uri"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/logger"
	"github.com/xigxog/kubefox/libs/core/utils"
)

type server struct {
	client *client.Client

	log *logger.Log
}

func New(client *client.Client, log *logger.Log) *server {
	return &server{
		client: client,
		log:    log,
	}
}

func checkPlatformName(kit kubefox.Kit) error {
	if kit.Platform() != nameArg(kit) {
		return kubefox.ErrResourceNotFound
	}

	return nil
}

func platformURI(kit kubefox.Kit) (uri.URI, error) {
	return uri.New(uri.Authority, uri.Platform, nameArg(kit))
}

func deploymentURI(kit kubefox.Kit) (uri.URI, error) {
	// if refTypeArg(kit) == uri.None {
	// 	return uri.New(uri.Authority, uri.Platform, nameArg(kit),
	// 		uri.Deployment)
	// } else {
	return uri.New(uri.Authority, uri.Platform, nameArg(kit),
		uri.Deployment, sysNameArg(kit), refTypeArg(kit), refArg(kit))
	// }
}

func releaseURI(kit kubefox.Kit) (uri.URI, error) {
	return uri.New(uri.Authority, uri.Platform, nameArg(kit),
		uri.Release, sysNameArg(kit), envNameArg(kit))
}

func systemURI(kit kubefox.Kit) (uri.URI, error) {
	return uri.New(uri.Authority, uri.System, sysNameArg(kit), refTypeArg(kit), refArg(kit))
}

func objURI(kit kubefox.Kit) (uri.URI, error) {
	return uri.New(uri.Authority, kindArg(kit), nameArg(kit), refTypeArg(kit), refArg(kit))
}

func subObjURI(kit kubefox.Kit, subKind uri.SubKind) (uri.URI, error) {
	return uri.New(uri.Authority, kindArg(kit), nameArg(kit), subKind, refArg(kit))
}

func systemNamespace(kit kubefox.Kit, sys string) string {
	return utils.SystemNamespace(kit.Platform(), sys)
}

func kindArg(kit kubefox.Kit) uri.Kind {
	return uri.KindFromString(kit.Request().GetArg("kind"))
}

func nameArg(kit kubefox.Kit) string {
	return kit.Request().GetArg("name")
}

func refTypeArg(kit kubefox.Kit) uri.SubKind {
	return uri.SubKindFromString(kit.Request().GetArg("refType"))
}

func refArg(kit kubefox.Kit) string {
	return kit.Request().GetArg("ref")
}

func sysNameArg(kit kubefox.Kit) string {
	return kit.Request().GetArg("sysName")
}

func envNameArg(kit kubefox.Kit) string {
	return kit.Request().GetArg("envName")
}
