package vault

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	vapi "github.com/hashicorp/vault/api"
	vauth "github.com/hashicorp/vault/api/auth/kubernetes"
	"github.com/xigxog/kubefox/libs/core/api/admin"
	"github.com/xigxog/kubefox/libs/core/api/uri"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/logger"
	"github.com/xigxog/kubefox/libs/core/utils"
)

type Client struct {
	*vapi.Client

	platform string
	watcher  *vapi.LifetimeWatcher
	log      *logger.Log
}

func NewClient(ctx kubefox.KitContext, role, url string, caBytes []byte) (*Client, error) {
	cfg := vapi.DefaultConfig()
	cfg.Address = url
	cfg.MaxRetries = 3
	cfg.HttpClient.Timeout = time.Minute
	cfg.ConfigureTLS(&vapi.TLSConfig{
		CACertBytes: caBytes,
	})

	cli, err := vapi.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	jwt, err := utils.GetSvcAccountToken(ctx.PlatformNamespace(), role)
	if err != nil {
		return nil, err
	}
	auth, err := vauth.NewKubernetesAuth(role, vauth.WithServiceAccountToken(jwt))
	if err != nil {
		return nil, err
	}
	authInfo, err := cli.Auth().Login(ctx, auth)
	if err != nil {
		return nil, err
	}
	if authInfo == nil {
		return nil, fmt.Errorf("error logging in with kubernetes auth: no auth info was returned")
	}
	ctx.Log().Debug("successfully logged into Vault using Kubernetes ServiceAccount")

	watcher, err := cli.NewLifetimeWatcher(&vapi.LifetimeWatcherInput{Secret: authInfo})
	if err != nil {
		return nil, fmt.Errorf("error starting Vault token renewer: %w", err)
	}
	go watcher.Start()
	ctx.Log().Debug("automatic renewal of Vault token started")

	ctx.Log().Infof("connected to Vault; role: %s, url: %s", role, url)

	return &Client{
		Client:   cli,
		platform: ctx.Platform(),
		watcher:  watcher,
		log:      ctx.Log(),
	}, nil
}

func (c *Client) Close() {
	if c.watcher != nil {
		c.watcher.Stop()
	}
}
func (c *Client) List(kit kubefox.KitContext, u uri.URI) ([]string, error) {
	path := fmt.Sprintf("%s/metadata/%s", MountPath(kit.Platform()), u.Path())
	kit.Log().Debugf("listing vault path %s", path)

	data, err := c.Client.Logical().ListWithContext(kit, path)
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

func (c *Client) Create(kit kubefox.KitContext, u uri.URI, obj admin.Object) error {
	return c.put(kit, u, obj, Create, 0)
}

func (c *Client) Put(kit kubefox.KitContext, u uri.URI, obj admin.Object) error {
	return c.put(kit, u, obj, Put, 0)
}

func (c *Client) Patch(kit kubefox.KitContext, u uri.URI, obj admin.Object) error {
	return c.put(kit, u, obj, Patch, 0)
}

func (c *Client) put(kit kubefox.KitContext, u uri.URI, obj admin.Object, op UpdateOp, attempt uint) error {
	path := u.Path()
	if u.SubKind() == uri.None {
		path = u.HeadPath()
	}
	kit.Log().Debugf("%s %s", op, u)

	if MountPath(kit.Platform()) == "" {
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
		_, err = c.Client.KVv2(MountPath(kit.Platform())).Put(kit, path, data, vapi.WithCheckAndSet(0))
		// update head
		if err == nil && u.SubKind() == uri.Id {
			_, err = c.Client.KVv2(MountPath(kit.Platform())).Put(kit, u.HeadPath(), data)
		}
		if err := c.get(kit, u.MetadataPath(), obj.GetMetadata()); err != nil {
			kit.Log().Debugf("could not get metadata %s, returning obj without it: %v", u.MetadataPath(), err)
		}

	case Put:
		_, err = c.Client.KVv2(MountPath(kit.Platform())).Put(kit, path, data)

	case Patch:
		if err := c.Exists(kit, u); err != nil {
			_, err = c.Client.KVv2(MountPath(kit.Platform())).Put(kit, path, data)
		} else {
			_, err = c.Client.KVv2(MountPath(kit.Platform())).Patch(kit, path, data)
		}
	}

	var respErr *vapi.ResponseError
	if errors.As(err, &respErr) {
		switch respErr.StatusCode {
		case http.StatusNotFound:
			kit.Log().Debugf("not found error %s %s, creating kv store", op, obj)
			// ensure kv store exists, if no errors creating it try put again
			if err = c.CreateKVStore(MountPath(kit.Platform()), "2"); err == nil {
				return c.put(kit, u, obj, op, attempt+1)
			}
		case http.StatusBadRequest:
			if strings.Contains(err.Error(), "check-and-set parameter did not match") {
				err = kubefox.ErrResourceConflict
			}
		}
	}
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Get(kit kubefox.KitContext, u uri.URI, obj admin.Object) error {
	kit.Log().Debugf("getting %s", u)

	p := u.Path()
	if u.SubKind() == uri.None {
		p = u.HeadPath()
	}

	if err := c.get(kit, p, obj); err != nil {
		return err
	}

	if err := c.get(kit, u.MetadataPath(), obj.GetMetadata()); err != nil {
		kit.Log().Debugf("could not get metadata %s, returning obj without it: %v", u.MetadataPath(), err)
	}

	obj.SetName(u.Name())

	return nil
}

func (c *Client) get(kit kubefox.KitContext, path string, obj any) error {
	path = fmt.Sprintf("%s/data/%s", MountPath(kit.Platform()), path)

	secret := &Secret{
		Data: &Data{
			Data: &Object{
				Object: obj,
			},
		},
	}

	resp, err := c.Client.Logical().ReadRawWithDataWithContext(kit, path, nil)
	if resp != nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusNotFound {
			err = kubefox.ErrResourceNotFound
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

func (c *Client) Delete(kit kubefox.KitContext, u uri.URI) error {
	kit.Log().Debugf("deleting %s", u)

	// deletes all versions
	err := c.Client.KVv2(MountPath(kit.Platform())).DeleteMetadata(kit, u.Path())
	if err != nil {
		var respErr *vapi.ResponseError
		if errors.As(err, &respErr); respErr.StatusCode == http.StatusNotFound {
			err = kubefox.ErrResourceNotFound
		}
	}

	return err
}

func (c *Client) Exists(kit kubefox.KitContext, u uri.URI) error {
	kit.Log().Debugf("checking if %s exists", u)
	p := u.Path()
	if u.SubKind() == uri.None {
		p = u.HeadPath()
	}

	if _, err := c.Client.KVv2(MountPath(kit.Platform())).GetMetadata(kit, p); err != nil {
		kit.Log().Debugf("error checking existence: %v", err)
		return fmt.Errorf("%w: %s", kubefox.ErrResourceNotFound, u)
	}

	return nil
}

func (c *Client) CreateKVStore(path string, version string) error {
	if cfg, _ := c.Client.Sys().MountConfig(path); cfg != nil {
		c.log.Debugf("%s kv store exists", path)
		return nil
	}

	c.log.Infof("creating %s kv store", path)
	err := c.Client.Sys().Mount(path, &vapi.MountInput{
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
