package operator

import (
	"net/http"

	"github.com/xigxog/kubefox/libs/core/kubefox"
	"k8s.io/apimachinery/pkg/runtime"
)

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
