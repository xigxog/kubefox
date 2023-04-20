package operator

import (
	"context"
	"fmt"
	"log"
	"os"

	vault "github.com/hashicorp/vault/api"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/logger"
	"github.com/xigxog/kubefox/libs/core/platform"
	"github.com/xigxog/kubefox/libs/core/utils"
	"golang.org/x/sync/semaphore"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8styps "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	k8scfg "sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	resetMsg = "PVC for the Vault Pod must be deleted, the Pod restarted, and the initialization process restarted."
)

type vaultOperator struct {
	Config

	vaultClient *vault.Client
	k8sClient   k8s.Client

	kitSvc kubefox.KitSvc

	initSem *semaphore.Weighted
	log     *logger.Log
}

func newVaultOperator(cfg Config, kitSvc kubefox.KitSvc) (*vaultOperator, error) {
	// init kubernetes client
	k8sCfg, err := k8scfg.GetConfig()
	if err != nil {
		return nil, err
	}

	kCl, err := k8sclient.New(k8sCfg, k8sclient.Options{Scheme: scheme.Scheme})
	if err != nil {
		return nil, err
	}

	// init vault client
	if kitSvc.DevMode() {
		os.Setenv(vault.EnvVaultSkipVerify, "true")
	}
	vaultCfg := vault.DefaultConfig()
	vaultCfg.Address = cfg.VaultURL
	vaultCfg.MaxRetries = 3

	vCl, err := vault.NewClient(vaultCfg)
	if err != nil {
		log.Fatalf("unable to initialize vault client: %v", err)
	}

	return &vaultOperator{
		Config:      cfg,
		vaultClient: vCl,
		k8sClient:   kCl,
		kitSvc:      kitSvc,
		initSem:     semaphore.NewWeighted(1),
		log:         kitSvc.Log(),
	}, nil

}

func (op *vaultOperator) Init() error {
	// ensure no concurrent calls
	if !op.initSem.TryAcquire(1) {
		op.log.Debug("init vault is already running")
		return nil
	}
	defer op.initSem.Release(1)

	return op.init()
}

func (op *vaultOperator) init() error {
	op.log.Debug("initializing vault")

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	statusRes, err := op.vaultClient.Sys().SealStatus()
	if err != nil {
		return fmt.Errorf("error getting vault's status: %w", err)
	}

	var key string
	kubeSecKey := k8styps.NamespacedName{
		Name:      "kfp-vault-keys",
		Namespace: op.kitSvc.Namespace(),
	}
	if !statusRes.Initialized {
		op.log.Debug("initializing vault")

		secrets, err := op.vaultClient.Sys().Init(&vault.InitRequest{
			SecretShares:    1,
			SecretThreshold: 1,
		})
		if err != nil {
			return fmt.Errorf("error initializing vault: %w", err)
		}

		key = secrets.Keys[0]
		op.vaultClient.SetToken(secrets.RootToken)

		sec := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      kubeSecKey.Name,
				Namespace: kubeSecKey.Namespace,
			},
			StringData: map[string]string{
				"key":   key,
				"token": secrets.RootToken,
			},
		}
		if err := op.k8sClient.Create(ctx, sec); err != nil {
			return fmt.Errorf("error storing vault secret in k8s: %w\n%s", err,
				"The initialization process for Vault did not finish. The "+resetMsg)
		}

	} else {
		op.log.Debug("vault already initialized")

		sec := &corev1.Secret{}
		if err := op.k8sClient.Get(ctx, kubeSecKey, sec); err != nil {
			return fmt.Errorf("error reading vault secret from k8s: %w\n%s", err,
				"Unable to retrieve Vault unseal key! If the Secret holding "+
					"the key has been deleted the "+resetMsg+" ALL DATA WILL BE LOST!")
		}

		key = string(sec.Data["key"])
		op.vaultClient.SetToken(string(sec.Data["token"]))
	}

	if statusRes.Sealed {
		op.log.Debug("unsealing vault")

		_, err := op.vaultClient.Sys().Unseal(key)
		if err != nil {
			return fmt.Errorf("error unsealing vault: %w", err)
		}

	} else {
		op.log.Debug("vault already unsealed")
	}

	if err := op.configureVault(); err != nil {
		return fmt.Errorf("error configuring vault: %w", err)
	}

	return nil
}

func (op *vaultOperator) configureVault() error {
	authList, err := op.vaultClient.Sys().ListAuth()
	if err != nil {
		return fmt.Errorf("error listing current vault auth methods: %w", err)
	}
	// Check if Kubernetes auth has been enabled and if not enable it.
	if _, found := authList["kubernetes/"]; found {
		op.log.Debug("kubernetes auth method already enabled")

	} else {
		// Enable K8s auth. This allows Pods to use their Service Account tokens
		// to authenticate with Vault.
		op.log.Info("enabling kubernetes auth method")
		err = op.vaultClient.Sys().EnableAuthWithOptions("kubernetes/", &vault.MountInput{Type: "kubernetes"})
		if err != nil {
			return fmt.Errorf("error enabling kubernetes auth method: %w", err)
		}

		op.log.Info("writing kubernetes auth method config")
		kHost := os.Getenv("KUBERNETES_SERVICE_HOST")
		if kHost == "" {
			kHost = "kubernetes.default.svc.cluster.local"
		}
		kPort := os.Getenv("KUBERNETES_SERVICE_PORT")
		if kPort == "" {
			kPort = "443"
		}
		_, err = op.vaultClient.Logical().Write("auth/kubernetes/config",
			map[string]interface{}{
				"kubernetes_host": fmt.Sprintf("https://%s:%s", kHost, kPort),
			})
		if err != nil {
			return fmt.Errorf("error writing config for kubernetes auth method: %w", err)
		}
	}

	op.log.Info("registering nats plugin")
	err = op.vaultClient.Sys().RegisterPlugin(&vault.RegisterPluginInput{
		Name:    "nats",
		Type:    vault.PluginTypeSecrets,
		Command: op.NATSPluginCmd,
		Version: op.NATSPluginVer,
		SHA256:  op.NATSPluginHash,
	})
	if err != nil {
		return fmt.Errorf("error registering nats plugin: %w", err)
	}

	if cfg, _ := op.vaultClient.Logical().Read("sys/mounts/nats/"); cfg == nil {
		op.log.Info("enabling nats plugin")
		err = op.vaultClient.Sys().Mount("nats", &vault.MountInput{
			Type: "nats",
			Config: vault.MountConfigInput{
				PluginVersion: op.NATSPluginVer,
			},
		})
		if err != nil {
			return fmt.Errorf("error enabling nats plugin: %w", err)
		}

		op.log.Info("configuring nats plugin")
		_, err = op.vaultClient.Logical().Write("nats/config",
			map[string]interface{}{
				"service-url": fmt.Sprintf("nats://nats.%s.svc.cluster.local:4222", op.kitSvc.Namespace()),
			})
		if err != nil {
			return fmt.Errorf("error writing config for kubernetes auth method: %w", err)
		}

	} else if cfg.Data["plugin_version"] != op.NATSPluginVer {
		op.log.Infof("upgrading nats plugin from %s to %s", cfg.Data["plugin_version"], op.NATSPluginVer)
		err := op.vaultClient.Sys().TuneMount("nats/", vault.MountConfigInput{
			PluginVersion: op.NATSPluginVer,
		})
		if err != nil {
			return fmt.Errorf("error upgrading nats plugin: %w", err)
		}
		_, err = op.vaultClient.Sys().ReloadPlugin(&vault.ReloadPluginInput{Plugin: "nats"})
		if err != nil {
			return fmt.Errorf("error upgrading nats plugin: %w", err)
		}

	} else {
		op.log.Debug("nats plugin already enabled")
	}

	// Create policy to allow the KubeFox platform components to manage KV
	// secret engines that store Platform objects and to create Kubernetes roles
	// for brokers as components are deployed.

	// TODO add config/policy to allow opensearch broker access to stored un/pw
	op.log.Info("writing kubefox platform policy")
	err = op.vaultClient.Sys().PutPolicy(platformPolicy, fmt.Sprintf(`
		path "auth/kubernetes/role/*" {
			capabilities = ["create", "read", "update", "patch", "delete", "list"]
		}
		path "nats/*" {
			capabilities = ["create", "read", "update", "patch", "delete", "list"]
		}
		path "%s/*" {
			capabilities = ["create", "read", "update", "patch", "delete", "list"]
		}
		path "sys/mounts/%s/*" {
			capabilities = ["create", "read", "update", "patch", "delete", "list"]
		}
	`, kvPrefix, kvPrefix))
	if err != nil {
		return fmt.Errorf("error creating operator policy: %w", err)
	}
	op.log.Info("writing kubefox broker policy")
	err = op.vaultClient.Sys().PutPolicy(brkPolicy, fmt.Sprintf(`
		path "nats/jwt/*" {
			capabilities = ["create"]
		}
		path "%s/*" {
			capabilities = ["read"]
		}
	`, kvPrefix))
	if err != nil {
		return fmt.Errorf("error creating broker policy: %w", err)
	}

	op.log.Info("writing kubefox operator role")
	_, err = op.vaultClient.Logical().Write("/auth/kubernetes/role/"+platformRole,
		map[string]interface{}{
			"bound_service_account_names": []string{
				platform.APISrvSvcAccount,
				platform.RuntimeSrvSvcAccount,
				platform.OprtrSvcAccount,
			},
			"bound_service_account_namespaces": []string{op.kitSvc.Namespace()},
			"policies":                         []string{"default", platformPolicy},
		})
	if err != nil {
		return fmt.Errorf("error creating operator role: %w", err)
	}
	op.log.Info("writing kubefox broker role")
	sysNSWild := utils.SystemNamespace(op.kitSvc.Platform(), "*")
	_, err = op.vaultClient.Logical().Write("/auth/kubernetes/role/"+brkRole,
		map[string]interface{}{
			"bound_service_account_names":      []string{platform.BrokerSvcAccount},
			"bound_service_account_namespaces": []string{op.kitSvc.Namespace(), sysNSWild},
			"policies":                         []string{"default", brkPolicy},
		})
	if err != nil {
		return fmt.Errorf("error creating broker role: %w", err)
	}

	return nil
}
