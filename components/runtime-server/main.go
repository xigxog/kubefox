package main

import (
	"flag"

	"github.com/xigxog/kubefox/components/runtime-server/server"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/utils"
)

func main() {
	var vaultURL string
	flag.StringVar(&vaultURL, "vault", "", "URL of the Vault server. Environment variable 'KUBEFOX_VAULT_URL' (default 'https://127.0.0.1:8200')")
	// parses flags
	kitSvc := kubefox.New()
	vaultURL = utils.ResolveFlag(vaultURL, "KUBEFOX_VAULT_URL", "https://127.0.0.1:8200")

	fbrSrv := server.New(vaultURL)
	kitSvc.MatchEvent("Type(`"+kubefox.BootstrapRequestType+"`)", fbrSrv.Bootstrap)
	kitSvc.MatchEvent("Type(`"+kubefox.FabricRequestType+"`)", fbrSrv.Weave)

	kitSvc.Start()
}
