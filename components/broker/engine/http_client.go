package engine

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/logkf"
)

type HTTPClient struct {
	wrapped *http.Client
	brk     Broker
	comp    *kubefox.Component

	log *logkf.Logger
}

func NewHTTPClient(brk Broker) *HTTPClient {
	id, err := os.Hostname()
	if err != nil || id == "" {
		id = uuid.NewString()
	}

	comp := &kubefox.Component{
		Name:   "http-client",
		Commit: kubefox.GitCommit,
		Id:     id,
	}
	return &HTTPClient{
		wrapped: &http.Client{},
		brk:     brk,
		comp:    comp,
		log:     logkf.Global,
	}
}

func (c *HTTPClient) SendEvent(req *LiveEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), req.TTL())
	defer cancel()

	httpReq, err := req.Event.HTTPRequest(ctx)
	if err != nil {
		return err
	}

	httpResp, err := c.wrapped.Do(httpReq)
	if err != nil {
		return err
	}

	resp := kubefox.NewResp(kubefox.EventOpts{
		Parent: req.Event,
		Source: c.comp,
		Target: req.Source,
	})
	if err = resp.SetHTTPResponse(httpResp); err != nil {
		return err
	}

	rEvt := &LiveEvent{
		Event:        resp,
		Receiver:     ReceiverHTTPClient,
		ReceivedAt:   time.Now(),
		Subscription: req.Subscription,
	}
	if err := c.brk.RecvEvent(rEvt); err != nil {
		return err
	}

	return nil
}
