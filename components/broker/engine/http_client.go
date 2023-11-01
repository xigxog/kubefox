package engine

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	kubefox "github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/logkf"
)

type HTTPClient struct {
	wrapped *http.Client

	brk Broker

	log *logkf.Logger
}

func NewHTTPClient(brk Broker) *HTTPClient {

	// TODO
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	return &HTTPClient{
		wrapped: http.DefaultClient,
		brk:     brk,
		log:     logkf.Global,
	}
}

func (c *HTTPClient) SendEvent(req *LiveEvent) error {
	log := c.log.WithEvent(req.Event)

	ctx, cancel := context.WithTimeout(context.Background(), req.TTL())

	httpReq, err := req.Event.HTTPRequest(ctx)
	if err != nil {
		cancel()
		return log.ErrorN("%w: error converting event to http request: %v", kubefox.ErrEventInvalid, err)
	}

	resp := kubefox.NewResp(kubefox.EventOpts{
		Parent: req.Event,
		Source: req.Target,
		Target: req.Source,
	})

	go func() {
		defer cancel()

		httpResp, err := c.wrapped.Do(httpReq)
		if err != nil {
			log.Error(err)
			return
		}

		if err = resp.SetHTTPResponse(httpResp); err != nil {
			log.Error(err)
			return
		}

		rEvt := &LiveEvent{
			Event:      resp,
			Receiver:   ReceiverHTTPClient,
			ReceivedAt: time.Now(),
		}
		if err := c.brk.RecvEvent(rEvt); err != nil {
			log.Error(err)
			return
		}
	}()

	return nil
}
