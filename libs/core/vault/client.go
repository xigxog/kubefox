package vault

import (
	"fmt"
	"log"
	"net/http"
	"os"

	vaultapi "github.com/hashicorp/vault/api"
	vaultauth "github.com/hashicorp/vault/api/auth/kubernetes"
	"github.com/xigxog/kubefox/libs/core/api/admin"
	"github.com/xigxog/kubefox/libs/core/api/uri"
	"github.com/xigxog/kubefox/libs/core/kubefox"
)

type VaultClient struct {
	client *vaultapi.Client

	kit kubefox.Kit
}

// TODO get svc account tk and login
func NewClient(vaultURL, svcAccToken string, kit kubefox.Kit) (*VaultClient, error) {
	// init vault client
	if kit.DevMode() {
		os.Setenv(vaultapi.EnvVaultSkipVerify, "true")
	}
	cfg := vaultapi.DefaultConfig()
	cfg.Address = vaultURL
	cfg.MaxRetries = 3

	vCl, err := vaultapi.NewClient(cfg)
	if err != nil {
		log.Fatalf("unable to initialize vault client: %v", err)
	}

	k8sAuth, err := vaultauth.NewKubernetesAuth(BrkRole, vaultauth.WithServiceAccountToken(svcAccToken))
	if err != nil {
		return nil, err
	}
	authInfo, err := vCl.Auth().Login(kit.Ctx(), k8sAuth)
	if err != nil {
		return nil, err
	}
	if authInfo == nil {
		return nil, fmt.Errorf("error logging in with kubernetes auth: no auth info was returned")
	}

	return &VaultClient{
		client: vCl,
		kit:    kit,
	}, nil
}

func (c *VaultClient) Get(u uri.URI, obj admin.Object) error {
	c.kit.Log().Debugf("getting %s", u)

	p := u.Path()
	if u.SubKind() == uri.None {
		p = u.HeadPath()
	}

	if err := c.get(p, obj); err != nil {
		return err
	}

	if err := c.get(u.MetadataPath(), obj.GetMetadata()); err != nil {
		c.kit.Log().Debugf("could not get metadata %s, returning obj without it: %v", u.MetadataPath(), err)
	}

	obj.SetName(u.Name())

	return nil
}

func (c *VaultClient) get(path string, obj any) error {
	path = fmt.Sprintf("%s/data/%s", MountPath(c.kit.Platform()), path)

	secret := &Secret{
		Data: &Data{
			Data: &Object{
				Object: obj,
			},
		},
	}

	resp, err := c.client.Logical().ReadRawWithDataWithContext(c.kit.Ctx(), path, nil)
	if resp != nil {
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			err = nil
		}
	}
	if err != nil {
		return err
	}

	if err := resp.DecodeJSON(secret); err != nil {
		return err
	}

	return nil
}
