package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
	auth "github.com/hashicorp/vault/api/auth/kubernetes"

	"github.com/xigxog/kubefox/libs/core/api/admin"
	"github.com/xigxog/kubefox/libs/core/api/uri"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/logger"
	"github.com/xigxog/kubefox/libs/core/platform"
	"github.com/xigxog/kubefox/libs/core/utils"
	"github.com/xigxog/kubefox/libs/core/vault"
	"golang.org/x/sync/semaphore"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type vaultClient struct {
	vaultClient *vaultapi.Client
	k8sClient   k8sclient.Client
	kitSvc      kubefox.KitSvc

	initSem *semaphore.Weighted
	log     *logger.Log
}

func (c *vaultClient) Login() error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	jwt, err := utils.GetSvcAccountToken(c.kitSvc.Namespace(), platform.OprtrSvcAccount)
	if err != nil {
		return err
	}

	k8sAuth, err := auth.NewKubernetesAuth(vault.PlatformRole, auth.WithServiceAccountToken(jwt))
	if err != nil {
		return err
	}

	c.vaultClient.SetToken("")
	authInfo, err := c.vaultClient.Auth().Login(ctx, k8sAuth)
	if err != nil {
		return err
	}
	if authInfo == nil {
		return fmt.Errorf("error logging in with kubernetes auth: no auth info was returned")
	}

	w, err := c.vaultClient.NewLifetimeWatcher(&vaultapi.LifetimeWatcherInput{Secret: authInfo})
	if err != nil {
		return fmt.Errorf("error starting vault token renewer: %w", err)
	}
	go w.Start()
	c.log.Debug("vault token renewer started")

	return nil
}

func (c *vaultClient) List(kit kubefox.Kit, u uri.URI) ([]string, error) {
	path := fmt.Sprintf("%s/metadata/%s", vault.MountPath(kit.Platform()), u.Path())
	kit.Log().Debugf("listing vault path %s", path)

	data, err := c.vaultClient.Logical().ListWithContext(kit.Ctx(), path)
	if err != nil {
		return nil, err
	}

	if data == nil || data.Data == nil {
		return []string{}, nil
	}

	var list []string
	if k, ok := data.Data["keys"].([]interface{}); ok {
		list = make([]string, len(k))
		for i, e := range k {
			list[i] = strings.TrimSuffix(e.(string), "/")
		}
	}

	return list, nil
}

func (c *vaultClient) Create(kit kubefox.Kit, u uri.URI, obj admin.Object) error {
	return c.put(kit, u, obj, Create, 0)
}

func (c *vaultClient) Put(kit kubefox.Kit, u uri.URI, obj admin.Object) error {
	return c.put(kit, u, obj, Put, 0)
}

func (c *vaultClient) Patch(kit kubefox.Kit, u uri.URI, obj admin.Object) error {
	return c.put(kit, u, obj, Patch, 0)
}

func (c *vaultClient) put(kit kubefox.Kit, u uri.URI, obj admin.Object, op UpdateOp, attempt uint) error {
	path := u.Path()
	if u.SubKind() == uri.None {
		path = u.HeadPath()
	}
	kit.Log().Debugf("%s %s", op, u)

	if vault.MountPath(kit.Platform()) == "" {
		return fmt.Errorf("error %s %s: no mount path", op, u)
	}
	if u.Name() == "" {
		return fmt.Errorf("error %s %s: no unique path", op, u)
	}

	var data map[string]any
	if u.SubKind() == uri.Metadata {
		data = map[string]any{"object": obj.GetMetadata()}
	} else {
		obj.SetMetadata(nil)
		data = map[string]any{"object": obj}
	}
	obj.SetName(u.Name())

	var err error
	switch op {
	case Create:
		_, err = c.vaultClient.KVv2(vault.MountPath(kit.Platform())).Put(kit.Ctx(), path, data, vaultapi.WithCheckAndSet(0))
		// update head
		if err == nil && u.SubKind() == uri.Id {
			_, err = c.vaultClient.KVv2(vault.MountPath(kit.Platform())).Put(kit.Ctx(), u.HeadPath(), data)
		}
		if err := c.get(kit, u.MetadataPath(), obj.GetMetadata()); err != nil {
			kit.Log().Debugf("could not get metadata '%s', returning obj without it: %v", u.MetadataPath(), err)
		}

	case Put:
		_, err = c.vaultClient.KVv2(vault.MountPath(kit.Platform())).Put(kit.Ctx(), path, data)

	case Patch:
		if err := c.Exists(kit, u); err != nil {
			_, err = c.vaultClient.KVv2(vault.MountPath(kit.Platform())).Put(kit.Ctx(), path, data)
		} else {
			_, err = c.vaultClient.KVv2(vault.MountPath(kit.Platform())).Patch(kit.Ctx(), path, data)
		}
	}

	var respErr *vaultapi.ResponseError
	if errors.As(err, &respErr) {
		switch respErr.StatusCode {
		case http.StatusNotFound:
			kit.Log().Debugf("not found error %s %s, creating kv store", op, obj)
			// ensure kv store exists, if no errors creating it try put again
			if err = c.CreateKVStore(vault.MountPath(kit.Platform()), "2"); err == nil {
				return c.put(kit, u, obj, op, attempt+1)
			}
		case http.StatusBadRequest:
			if strings.Contains(err.Error(), "check-and-set parameter did not match") {
				err = ErrResourceConflict
			}
		}
	}
	if err != nil {
		return err
	}

	return nil
}

func (c *vaultClient) Get(kit kubefox.Kit, u uri.URI, obj admin.Object) error {
	kit.Log().Debugf("getting %s", u)

	p := u.Path()
	if u.SubKind() == uri.None {
		p = u.HeadPath()
	}

	if err := c.get(kit, p, obj); err != nil {
		return err
	}

	if err := c.get(kit, u.MetadataPath(), obj.GetMetadata()); err != nil {
		kit.Log().Debugf("could not get metadata '%s', returning obj without it: %v", u.MetadataPath(), err)
	}

	obj.SetName(u.Name())

	return nil
}

func (c *vaultClient) get(kit kubefox.Kit, path string, obj any) error {
	path = fmt.Sprintf("%s/data/%s", vault.MountPath(kit.Platform()), path)

	secret := &vault.Secret{
		Data: &vault.Data{
			Data: &vault.Object{
				Object: obj,
			},
		},
	}

	resp, err := c.vaultClient.Logical().ReadRawWithDataWithContext(kit.Ctx(), path, nil)
	if resp != nil {
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			err = ErrResourceNotFound
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

func (c *vaultClient) Delete(kit kubefox.Kit, u uri.URI) error {
	kit.Log().Debugf("deleting %s", u)

	// deletes all versions
	err := c.vaultClient.KVv2(vault.MountPath(kit.Platform())).DeleteMetadata(kit.Ctx(), u.Path())
	if err != nil {
		var respErr *vaultapi.ResponseError
		if errors.As(err, &respErr); respErr.StatusCode == http.StatusNotFound {
			err = ErrResourceNotFound
		}
	}

	return err
}

func (c *vaultClient) Exists(kit kubefox.Kit, u uri.URI) error {
	kit.Log().Debugf("checking if %s exists", u)
	p := u.Path()
	if u.SubKind() == uri.None {
		p = u.HeadPath()
	}

	if _, err := c.vaultClient.KVv2(vault.MountPath(kit.Platform())).GetMetadata(kit.Ctx(), p); err != nil {
		kit.Log().Debugf("error checking existence: %v", err)
		return fmt.Errorf("%w: %s", ErrResourceNotFound, u)
	}

	return nil
}

func (c *vaultClient) CreateKVStore(path string, version string) error {
	if cfg, _ := c.vaultClient.Sys().MountConfig(path); cfg != nil {
		c.log.Debugf("%s kv store exists", path)
		return nil
	}

	c.log.Infof("creating %s kv store", path)
	err := c.vaultClient.Sys().Mount(path, &vaultapi.MountInput{
		Type:        "kv",
		Description: "",
		Options: map[string]string{
			"version": version,
		},
	})
	if err != nil {
		return fmt.Errorf("error creating %s kv store: %w", path, err)
	}

	return nil
}
