package engine

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
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
	ctx, cancel := context.WithTimeout(context.Background(), req.TTL())
	log := c.log.WithEvent(req.Event)

	httpReq, err := req.Event.HTTPRequest(ctx)
	if err != nil {
		cancel()
		return log.ErrorN("%w: error converting event to http request: %v", kubefox.ErrInvalid, err)
	}

	if adapter := req.TargetAdapter; adapter != nil {
		if adapterURL, err := url.Parse(adapter.URL.StringVal); err == nil { // success
			adapterURL = adapterURL.JoinPath(httpReq.URL.EscapedPath())

			httpReq.URL.Scheme = adapterURL.Scheme
			httpReq.URL.Host = adapterURL.Host
			httpReq.URL.User = adapterURL.User
			httpReq.URL.Path = adapterURL.Path
			httpReq.URL.RawPath = adapterURL.RawPath

			if adapterURL.Fragment != "" {
				httpReq.URL.Fragment = adapterURL.Fragment
				httpReq.URL.RawFragment = adapterURL.RawFragment
			}

			httpQuery := httpReq.URL.Query()
			for k, v := range adapterURL.Query() {
				httpQuery[k] = v
			}
			httpReq.URL.RawQuery = httpQuery.Encode()

		} else if adapter.URL.StringVal != "" {
			cancel()
			return fmt.Errorf("error parsing adapter url: %v", err)
		}

		for k, v := range adapter.Headers {
			httpReq.Header.Set(k, v.StringVal)
		}
	}

	resp := kubefox.NewResp(kubefox.EventOpts{
		Parent: req.Event,
		Target: req.Source,
		Source: &kubefox.Component{
			Name:     req.Target.Name,
			Commit:   c.brk.Component().Commit,
			Id:       c.brk.Component().Id,
			BrokerId: c.brk.Component().Id,
		},
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

		evt := &LiveEvent{
			Event:      resp,
			Receiver:   ReceiverHTTPClient,
			ReceivedAt: time.Now(),
		}
		if err := c.brk.RecvEvent(evt); err != nil {
			log.Error(err)
		}
	}()

	return nil
}
