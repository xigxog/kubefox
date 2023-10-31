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

	brk  Broker
	comp *kubefox.Component

	log *logkf.Logger
}

func NewHTTPClient(brk Broker) *HTTPClient {
	comp := &kubefox.Component{
		Name:     "http-client",
		Commit:   brk.Component().Commit,
		Id:       brk.Component().Id,
		BrokerId: brk.Component().BrokerId,
	}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	return &HTTPClient{
		wrapped: http.DefaultClient,
		brk:     brk,
		comp:    comp,
		log:     logkf.Global,
	}
}

func (c *HTTPClient) Component() *kubefox.Component {
	return c.comp
}

func (c *HTTPClient) SendEvent(req *LiveEvent) error {
	log := c.log.WithEvent(req.Event)

	ctx, cancel := context.WithTimeout(context.Background(), req.TTL())

	httpReq, err := req.Event.HTTPRequest(ctx)
	if err != nil {
		cancel()
		return log.ErrorN("%w: error converting event to http request: %v", ErrEventInvalid, err)
	}

	resp := kubefox.NewResp(kubefox.EventOpts{
		Parent: req.Event,
		Source: c.comp,
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
