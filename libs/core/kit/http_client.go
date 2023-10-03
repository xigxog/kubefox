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

func NewHTTPClient() *http.Client {
	return &http.Client{Transport: &KitRoundTripper{}}
}

func (rt *KitRoundTripper) Do(httpReq *http.Request) (*http.Response, error) {
	return rt.RoundTrip(httpReq)
}

func (rt *KitRoundTripper) RoundTrip(httpReq *http.Request) (*http.Response, error) {
	kitCtx := httpReq.Context().Value(ReqContextKey).(*ReqContext)
	kitSvc := kitCtx.k

	// TODO
	// req := kit.req.ChildEvent()
	req := &kubefox.Event{}
	req.Target = kitCtx.target

	if err := req.ParseHTTPRequest(httpReq); err != nil {
		kitSvc.Log().Error(err)
		return nil, err
	}

	resp, err := kitSvc.sendReq(req)
	if err != nil {
		kitSvc.Log().Error(err)
		return nil, err
	}

	return resp.ToHTTPResponse(), nil
}
