package main

import (
	"flag"

	"github.com/xigxog/kubefox/components/operator/operator"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/utils"
)

// Injected at build time
var (
	GitRef  string
	GitHash string
)

func main() {
	flag.StringVar(&operator.ContainerRegistry, "cr", "", "Container registry to use for Platform components. Environment variable 'KUBEFOX_CR' (default 'ghcr.io/xigxog/kubefox')")
	// parses flags
	kitSvc := kubefox.New()

	operator.ContainerRegistry = utils.ResolveFlag(operator.ContainerRegistry, "KUBEFOX_CR", "ghcr.io/xigxog/kubefox")
	operator.GitRef = GitRef
	operator.GitHash = GitHash

	op, err := operator.New(kitSvc)
	if err != nil {
		kitSvc.Fatal(err)
	}

	kitSvc.MatchEvent("Type(`"+kubefox.BootstrapRequestType+"`)", op.Bootstrap)
	kitSvc.MatchEvent("Type(`"+kubefox.FabricRequestType+"`)", op.Weave)

	kitSvc.Kubernetes("TargetKind(`Platform`)", op.ProcessPlatform)
	kitSvc.Kubernetes("TargetKind(`ComponentSet`)", op.ProcessComponentSet)
	kitSvc.Kubernetes("TargetKind(`Release`)", op.ProcessRelease)

	kitSvc.Start()
}
