package engine

import (
	"net/http"

	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/logkf"
)

type HTTPClient struct {
	wrapped *http.Client
	brk     Broker

	log *logkf.Logger
}

func NewHTTPClient(brk Broker) *HTTPClient {
	return &HTTPClient{
		wrapped: &http.Client{},
		brk:     brk,
		log:     logkf.Global,
	}
}

func (c *HTTPClient) SendEvent(req ReceivedEvent) error {
	httpReq, err := req.Event.HTTPRequest(req.Context)
	if err != nil {
		return err
	}

	httpResp, err := c.wrapped.Do(httpReq)
	if err != nil {
		return err
	}

	resp := kubefox.NewEvent()
	if err = resp.SetHTTPResponse(httpResp); err != nil {
		return err
	}
	resp.SetParent(req.Event)

	rEvt := &ReceivedEvent{
		Event:        resp,
		Receiver:     EventReceiverHTTPClient,
		Subscription: req.Subscription,
	}
	if err := c.brk.RecvEvent(rEvt); err != nil {
		return err
	}

	return nil
}
