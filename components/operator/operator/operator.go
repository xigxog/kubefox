package operator

import (
	"net/http"

	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/logger"
	"k8s.io/apimachinery/pkg/runtime"
)

type operator struct {
	Config

	vaultOp *vaultOperator

	log *logger.Log
}

func New(cfg Config, kitSvc kubefox.KitSvc) (*operator, error) {
	v, err := newVaultOperator(cfg, kitSvc)
	if err != nil {
		return nil, err
	}
	if err := v.Init(); err != nil {
		return nil, err
	}

	return &operator{
		Config:  cfg,
		vaultOp: v,
		log:     kitSvc.Log(),
	}, nil
}

func CustomizeEvent(kit kubefox.Kit, rules ...*RelatedResourceRule) error {
	resp := kit.Response().HTTP()
	resp.SetType(kubefox.KubernetesResponseType)
	resp.SetStatusCode(http.StatusOK)

	if rules == nil {
		rules = make([]*RelatedResourceRule, 0)
	}

	custResp := &CustomizeResponse{RelatedResourceRules: rules}
	// kit.Log().DebugInterface(custResp, "%s sync customize:", parent)

	return resp.Marshal(custResp)
}

func SyncEvent(kit kubefox.Kit, status any, attachments ...runtime.Object) error {
	resp := kit.Response().HTTP()
	resp.SetType(kubefox.KubernetesResponseType)
	resp.SetStatusCode(http.StatusOK)

	if attachments == nil {
		attachments = make([]runtime.Object, 0)
	}

	syncResp := &KubeResponse{
		Attachments: attachments,
		Status:      status,
	}
	// kit.Log().DebugInterface(syncResp, "resource sync response:")

	return resp.Marshal(syncResp)
}

func ErrEvent(kit kubefox.Kit, err error) error {
	kit.Log().Error(err)

	resp := kit.Response().HTTP()
	resp.SetType(kubefox.KubernetesResponseType)
	resp.SetStatusCode(http.StatusBadRequest)

	return resp.Marshal(struct{ Err string }{
		Err: err.Error(),
	})
}
