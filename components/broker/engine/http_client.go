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
		return log.ErrorN("%w: error converting event to http request: %v", kubefox.ErrEventInvalid, err)
	}

	if a := req.TargetAdapter; a != nil {
		if u, err := url.Parse(a.URL.StringVal); err == nil { // success
			u = u.JoinPath(httpReq.URL.EscapedPath())

			httpReq.URL.Scheme = u.Scheme
			httpReq.URL.Host = u.Host
			httpReq.URL.User = u.User
			httpReq.URL.Path = u.Path
			httpReq.URL.RawPath = u.RawPath

			if u.Fragment != "" {
				httpReq.URL.Fragment = u.Fragment
				httpReq.URL.RawFragment = u.RawFragment
			}

			httpQuery := httpReq.URL.Query()
			for k, v := range u.Query() {
				httpQuery[k] = v
			}
			httpReq.URL.RawQuery = httpQuery.Encode()

		} else if a.URL.StringVal != "" {
			cancel()
			return fmt.Errorf("error parsing adapter url: %v", err)
		}

		for k, v := range a.Headers {
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
