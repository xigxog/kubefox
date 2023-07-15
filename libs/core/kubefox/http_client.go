package kubefox

import (
	"net/http"
)

type KitRoundTripper struct{}

func NewHTTPClient() *http.Client {
	return &http.Client{Transport: &KitRoundTripper{}}
}

func (rt *KitRoundTripper) RoundTrip(httpReq *http.Request) (*http.Response, error) {
	kitCtx := httpReq.Context().Value(ReqCtxKey).(*reqCtx)
	kit := kitCtx.kit

	req := kit.req.ChildEvent()
	req.SetTarget(kitCtx.target)

	if err := req.HTTPData().ParseRequest(httpReq); err != nil {
		kit.Log().Error(err)
		return nil, err
	}

	resp, err := kit.broker.InvokeTarget(req)
	if err != nil {
		kit.Log().Error(err)
		return nil, err
	}

	return resp.HTTPData().GetHTTPResponse(), nil
}
