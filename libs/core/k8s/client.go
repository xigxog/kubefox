package k8s

import (
	"fmt"
	"time"

	kfv1a1 "github.com/xigxog/kubefox/libs/core/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"k8s.io/client-go/kubernetes/scheme"
	k8scli "sigs.k8s.io/controller-runtime/pkg/client"
	k8scfg "sigs.k8s.io/controller-runtime/pkg/client/config"
)

type Client struct {
	k8scli.Client

	fOwner k8scli.FieldOwner
}

func New(fOwner k8scli.FieldOwner) (*Client, error) {
	if err := kfv1a1.SchemeBuilder.AddToScheme(scheme.Scheme); err != nil {
		return nil, err
	}
	k8sCfg, err := k8scfg.GetConfig()
	if err != nil {
		return nil, err
	}
	k8sCli, err := k8scli.New(k8sCfg, k8scli.Options{Scheme: scheme.Scheme})
	if err != nil {
		return nil, err
	}

	return &Client{Client: k8sCli, fOwner: fOwner}, nil
}

func (c *Client) Apply(ctx kubefox.KitContext, obj k8scli.Object, retry bool) error {
	s := fmt.Sprintf("%s/%s", obj.GetObjectKind().GroupVersionKind().Kind, obj.GetName())
	ctx.Log().Debugf("applying resource %s", s)

	if err := c.Client.Patch(ctx, obj, k8scli.Apply, c.fOwner, k8scli.ForceOwnership); err != nil {
		if retry {
			ctx.Log().Warnf("error applying resource %s: %v, retrying...", s, err)
			time.Sleep(3 * time.Second)
			return c.Apply(ctx, obj, false)
		} else {
			return err
		}
	}

	return nil
}
