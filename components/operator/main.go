package main

import (
	"flag"
	"fmt"

	"github.com/xigxog/kubefox/components/operator/operator"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/utils"
)

func main() {
	cfg := operator.Config{}
	flag.StringVar(&cfg.BrokerImage, "broker-image", "", "Broker image to use for deployed components. Environment variable 'KUBEFOX_BROKER_IMAGE' (default 'ghcr.io/xigxog/kubefox/broker:main')")
	flag.StringVar(&cfg.VaultURL, "vault", "", "URL of the Vault server. Environment variable 'KUBEFOX_VAULT_URL' (default 'https://127.0.0.1:8200')")
	flag.StringVar(&cfg.NATSPluginCmd, "nats-plugin-cmd", "", "Name of binary of the the Vault NATS Plugin. Environment variable 'KUBEFOX_NATS_PLUGIN_CMD' (required)")
	flag.StringVar(&cfg.NATSPluginVer, "nats-plugin-version", "", "Version of the Vault NATS Plugin. Environment variable 'KUBEFOX_NATS_PLUGIN_VERSION' (required)")
	flag.StringVar(&cfg.NATSPluginHash, "nats-plugin-hash", "", "SHA256 hash of the Vault NATS Plugin. Environment variable 'KUBEFOX_NATS_PLUGIN_HASH' (required)")

	// parses flags
	kitSvc := kubefox.New()
	cfg.BrokerImage = utils.ResolveFlag(cfg.BrokerImage, "KUBEFOX_BROKER_IMAGE", "ghcr.io/xigxog/kubefox/broker:main")
	cfg.VaultURL = utils.ResolveFlag(cfg.VaultURL, "KUBEFOX_VAULT_URL", "https://127.0.0.1:8200")
	cfg.NATSPluginCmd = utils.ResolveFlag(cfg.NATSPluginCmd, "KUBEFOX_NATS_PLUGIN_CMD", "")
	cfg.NATSPluginVer = utils.ResolveFlag(cfg.NATSPluginVer, "KUBEFOX_NATS_PLUGIN_VERSION", "")
	cfg.NATSPluginHash = utils.ResolveFlag(cfg.NATSPluginHash, "KUBEFOX_NATS_PLUGIN_HASH", "")
	if cfg.NATSPluginCmd == "" || cfg.NATSPluginVer == "" || cfg.NATSPluginHash == "" {
		kitSvc.Fatal(fmt.Errorf("nats-plugin-cmd, nats-plugin-version, and nats-plugin-hash are required"))
	}

	o, err := operator.New(cfg, kitSvc)
	if err != nil {
		kitSvc.Fatal(err)
	}

	kitSvc.Kubernetes("TargetKind(`Platform`)", o.ProcessPlatform)
	kitSvc.Kubernetes("TargetKind(`ComponentSet`)", o.ProcessComponentSet)
	kitSvc.Kubernetes("TargetKind(`Release`)", o.ProcessRelease)

	kitSvc.Start()
}
