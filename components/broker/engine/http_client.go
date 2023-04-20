package engine

import (
	"context"
	"net/http"
	"time"

	"github.com/xigxog/kubefox/libs/core/kubefox"
)

type HTTPClient struct {
	Broker

	httpClient *http.Client
}

func NewHTTPClient(brk Broker) *HTTPClient {
	return &HTTPClient{
		Broker: brk,
		httpClient: &http.Client{
			Timeout: time.Second * 3,
		},
	}
}

func (cl *HTTPClient) SendEvent(ctx context.Context, req kubefox.DataEvent) (resp kubefox.DataEvent) {
	httpReq, err := req.HTTPData().GetHTTPRequest(ctx)
	if err != nil {
		resp = req.ChildErrorEvent(err)
		cl.Log().Error(resp.GetError())
		return
	}

	httpResp, err := cl.httpClient.Do(httpReq)
	if err != nil {
		resp = req.ChildErrorEvent(err)
		cl.Log().Error(resp.GetError())
		return
	}

	resp = req.ChildEvent()
	if err = resp.HTTPData().ParseResponse(httpResp); err != nil {
		resp = req.ChildErrorEvent(err)
		cl.Log().Error(resp.GetError())
		return
	}

	return
}
