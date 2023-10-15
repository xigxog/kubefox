package kit

import (
	"net/http"

	"github.com/xigxog/kubefox/libs/core/kubefox"
)

type ContextKey string

const (
	ReqContextKey ContextKey = "KitReqCtx"
)

type ReqContext struct {
	k      *kit
	target *kubefox.Component
}

type KitRoundTripper struct{}

// TODO
func NewHTTPClient() *http.Client {
	return &http.Client{Transport: &KitRoundTripper{}}
}

func (rt *KitRoundTripper) Do(httpReq *http.Request) (*http.Response, error) {
	return rt.RoundTrip(httpReq)
}

func (rt *KitRoundTripper) RoundTrip(httpReq *http.Request) (*http.Response, error) {
	reqCtx := httpReq.Context().Value(ReqContextKey).(*ReqContext)
	k := reqCtx.k

	// req := kit.req.ChildEvent()
	req := kubefox.NewEvent()
	req.Target = reqCtx.target

	if err := req.SetHTTPRequest(httpReq); err != nil {
		k.Log().Error(err)
		return nil, err
	}

	resp, err := k.sendReq(httpReq.Context(), req)
	if err != nil {
		k.Log().Error(err)
		return nil, err
	}

	return resp.HTTPResponse(), nil
}
