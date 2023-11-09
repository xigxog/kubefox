package engine

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	kubefox "github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/logkf"
)

type HTTPClient struct {
	clients map[string]*http.Client

	secureTransport   *http.Transport
	insecureTransport *http.Transport

	brk Broker

	log *logkf.Logger
}

func NewHTTPClient(brk Broker) *HTTPClient {
	clients := make(map[string]*http.Client, 7)

	// TODO support live refresh of root ca
	// https://github.com/breml/rootcerts/blob/master/generate_data.go
	// - run background thread
	// - download latest mozilla certs
	// - update secureTransport.TLSClientConfig.Config.RootCAs
	secureTransport := http.DefaultTransport.(*http.Transport).Clone()
	if certs, err := os.ReadFile("/etc/ssl/certs/mozilla.crt"); err != nil {
		logkf.Global.Errorf("error reading Mozilla root CAs from file: %v", err)
	} else {
		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(certs)
		secureTransport.TLSClientConfig = &tls.Config{RootCAs: certPool}
	}

	insecureTransport := http.DefaultTransport.(*http.Transport).Clone()
	insecureTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	clients["default"] = &http.Client{
		CheckRedirect: followNever,
		Transport:     secureTransport,
	}
	clients[key(kubefox.FollowRedirectsNever, false)] = &http.Client{
		CheckRedirect: followNever,
		Transport:     secureTransport,
	}
	clients[key(kubefox.FollowRedirectsNever, true)] = &http.Client{
		CheckRedirect: followNever,
		Transport:     insecureTransport,
	}
	clients[key(kubefox.FollowRedirectsSameHost, false)] = &http.Client{
		CheckRedirect: followSameHost,
		Transport:     secureTransport,
	}
	clients[key(kubefox.FollowRedirectsSameHost, true)] = &http.Client{
		CheckRedirect: followSameHost,
		Transport:     insecureTransport,
	}
	clients[key(kubefox.FollowRedirectsAlways, false)] = &http.Client{
		Transport: secureTransport,
	}
	clients[key(kubefox.FollowRedirectsAlways, true)] = &http.Client{
		Transport: insecureTransport,
	}

	return &HTTPClient{
		clients:           clients,
		secureTransport:   secureTransport,
		insecureTransport: insecureTransport,
		brk:               brk,
		log:               logkf.Global,
	}
}

func (c *HTTPClient) SendEvent(req *BrokerEvent) error {
	adapter := req.TargetAdapter
	if adapter == nil {
		return kubefox.ErrInvalid(fmt.Errorf("adapter is missing"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), req.TTL())
	log := c.log.WithEvent(req.Event)

	httpReq, err := req.Event.HTTPRequest(ctx)
	if err != nil {
		cancel()
		return kubefox.ErrInvalid(err)
	}
	if adapterURL, err := url.Parse(adapter.URL.StringVal); err != nil { // success
		cancel()
		return kubefox.ErrInvalid(fmt.Errorf("error parsing adapter url: %v", err))

	} else {
		adapterURL = adapterURL.JoinPath(httpReq.URL.EscapedPath())

		httpReq.Host = adapterURL.Host
		httpReq.URL.Host = adapterURL.Host
		httpReq.URL.Scheme = adapterURL.Scheme
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
	}

	for k, v := range adapter.Headers {
		if strings.EqualFold(k, kubefox.HeaderHost) {
			httpReq.Host = v.StringVal
		}
		httpReq.Header.Set(k, v.StringVal)
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

		var reqErr error
		if httpResp, err := c.adapterClient(adapter).Do(httpReq); err != nil {
			reqErr = kubefox.ErrUnexpected(fmt.Errorf("http request failed: %v", err))
		} else {
			reqErr = resp.SetHTTPResponse(httpResp)
		}
		if reqErr != nil {
			if !errors.Is(reqErr, &kubefox.Err{}) {
				reqErr = kubefox.ErrUnexpected(reqErr)
			}
			resp.Type = string(kubefox.EventTypeError)
			resp.SetJSON(reqErr)

			log.Debug(err)
		}

		c.brk.RecvEvent(resp, ReceiverHTTPClient)
	}()

	return nil
}

func (c *HTTPClient) adapterClient(a *v1alpha1.Adapter) *http.Client {
	key := key(a.FollowRedirects, a.InsecureSkipVerify)
	client := c.clients[key]
	if client == nil {
		client = c.clients["default"]
	}
	return client
}

func key(follow kubefox.FollowRedirects, insecure bool) string {
	if follow == "" {
		follow = kubefox.FollowRedirectsNever
	}
	return fmt.Sprintf("%s-%t", follow, insecure)
}

func followNever(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

func followSameHost(req *http.Request, via []*http.Request) error {
	if len(via) == 0 {
		return http.ErrUseLastResponse
	}
	if req.URL.Host != via[0].URL.Host {
		// Different host, do not follow redirect.
		return http.ErrUseLastResponse
	}

	return nil
}
