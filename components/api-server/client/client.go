package client

import (
	"log"
	"os"
	"time"

	vault "github.com/hashicorp/vault/api"
	kubev1a1 "github.com/xigxog/kubefox/libs/core/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/logger"
	"github.com/xigxog/kubefox/libs/core/platform"
	"github.com/xigxog/kubefox/libs/core/validator"
	"golang.org/x/sync/semaphore"
	"k8s.io/client-go/kubernetes/scheme"
	kubeclient "sigs.k8s.io/controller-runtime/pkg/client"
	kubeconfig "sigs.k8s.io/controller-runtime/pkg/client/config"
)

var (
	defaultTimeout = 30 * time.Second
)

type UpdateOp uint8

const (
	Create UpdateOp = iota
	Put
	Patch
)

func (o UpdateOp) String() string {
	switch o {
	case Create:
		return "creating"
	case Put:
		return "putting"
	case Patch:
		return "patching"
	default:
		return "shrugging"
	}
}

type Client struct {
	kitSvc      *kubefox.KitSvc
	kubeClient  *kubeClient
	vaultClient *vaultClient

	*validator.Validator

	log *logger.Log
}

func New(vaultURL string, kitSvc kubefox.KitSvc) (*Client, error) {
	// add KubeFox resources to default schema
	if err := kubev1a1.SchemeBuilder.AddToScheme(scheme.Scheme); err != nil {
		return nil, err
	}

	// init kubernetes client
	kubeCfg, err := kubeconfig.GetConfig()
	if err != nil {
		return nil, err
	}
	kube, err := kubeclient.New(kubeCfg, kubeclient.Options{Scheme: scheme.Scheme})
	if err != nil {
		return nil, err
	}
	patchOpts := &kubeclient.PatchOptions{
		FieldManager: platform.APISrvComp.GetName(),
	}

	// init vault client
	if kitSvc.DevMode() {
		os.Setenv(vault.EnvVaultSkipVerify, "true")
	}
	vaultCfg := vault.DefaultConfig()
	vaultCfg.Address = vaultURL
	vaultCfg.MaxRetries = 3

	vault, err := vault.NewClient(vaultCfg)
	if err != nil {
		log.Fatalf("unable to initialize vault client: %v", err)
	}

	kubeClient := &kubeClient{
		kube:      kube,
		patchOpts: patchOpts,
		log:       kitSvc.Log(),
	}

	vaultClient := &vaultClient{
		vaultClient: vault,
		kitSvc:      kitSvc,
		k8sClient:   kube,
		initSem:     semaphore.NewWeighted(1),
		log:         kitSvc.Log(),
	}
	if err := vaultClient.Login(); err != nil {
		return nil, err
	}

	return &Client{
		kubeClient:  kubeClient,
		vaultClient: vaultClient,
		Validator:   validator.New(kitSvc.Log()),
		log:         kitSvc.Log(),
	}, nil
}

func (client *Client) Kube() *kubeClient {
	return client.kubeClient
}

func (client *Client) Vault() *vaultClient {
	return client.vaultClient
}
