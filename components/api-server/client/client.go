package client

import (
	"time"

	kfv1a1 "github.com/xigxog/kubefox/libs/core/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/logger"
	"github.com/xigxog/kubefox/libs/core/platform"
	"github.com/xigxog/kubefox/libs/core/validator"
	"github.com/xigxog/kubefox/libs/core/vault"
	"k8s.io/client-go/kubernetes/scheme"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
	kconfig "sigs.k8s.io/controller-runtime/pkg/client/config"
)

var (
	defaultTimeout = 30 * time.Second
)

type Client struct {
	k8sClient   *k8sClient
	vaultClient *vault.Client

	*validator.Validator

	log *logger.Log
}

func New(ctx kubefox.KitContext, vaultURL string) (*Client, error) {
	// add KubeFox resources to default schema
	if err := kfv1a1.SchemeBuilder.AddToScheme(scheme.Scheme); err != nil {
		return nil, err
	}
	// create Kubernetes client
	kubeCfg, err := kconfig.GetConfig()
	if err != nil {
		return nil, err
	}
	kube, err := kclient.New(kubeCfg, kclient.Options{Scheme: scheme.Scheme})
	if err != nil {
		return nil, err
	}
	patchOpts := &kclient.PatchOptions{
		FieldManager: platform.APISrvComp.GetName(),
	}
	kubeClient := &k8sClient{
		kube:      kube,
		patchOpts: patchOpts,
		log:       ctx.Log(),
	}

	// create Vault client
	vaultClient, err := vault.NewClient(ctx, vault.APISrvRole, vaultURL)
	if err != nil {
		return nil, err
	}

	return &Client{
		k8sClient:   kubeClient,
		vaultClient: vaultClient,
		Validator:   validator.New(ctx.Log()),
		log:         ctx.Log(),
	}, nil
}

func (client *Client) Kube() *k8sClient {
	return client.k8sClient
}

func (client *Client) Vault() *vault.Client {
	return client.vaultClient
}
