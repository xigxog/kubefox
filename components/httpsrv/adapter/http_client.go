// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package adapter

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

	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/grpc"
	"github.com/xigxog/kubefox/logkf"
)

type HTTPClient struct {
	clients map[string]*http.Client

	secureTransport   *http.Transport
	insecureTransport *http.Transport

	brk *grpc.Client

	log *logkf.Logger
}

func NewHTTPClient(brk *grpc.Client) *HTTPClient {
	clients := make(map[string]*http.Client, 7)

	// TODO support live refresh of root ca
	// https://github.com/breml/rootcerts/blob/master/generate_data.go
	// - run background thread
	// - download latest mozilla certs
	// - update secureTransport.TLSClientConfig.Config.RootCAs
	secureTransport := http.DefaultTransport.(*http.Transport).Clone()
	if certs, err := os.ReadFile("/etc/ssl/certs/mozilla.crt"); err != nil {
		logkf.Global.Warnf("error reading Mozilla root CAs from file: %v", err)
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
	clients[clientKey(api.FollowRedirectsNever, false)] = &http.Client{
		CheckRedirect: followNever,
		Transport:     secureTransport,
	}
	clients[clientKey(api.FollowRedirectsNever, true)] = &http.Client{
		CheckRedirect: followNever,
		Transport:     insecureTransport,
	}
	clients[clientKey(api.FollowRedirectsSameHost, false)] = &http.Client{
		CheckRedirect: followSameHost,
		Transport:     secureTransport,
	}
	clients[clientKey(api.FollowRedirectsSameHost, true)] = &http.Client{
		CheckRedirect: followSameHost,
		Transport:     insecureTransport,
	}
	clients[clientKey(api.FollowRedirectsAlways, false)] = &http.Client{
		Transport: secureTransport,
	}
	clients[clientKey(api.FollowRedirectsAlways, true)] = &http.Client{
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

func (c *HTTPClient) SendEvent(req *grpc.ComponentEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), req.Event.TTL())
	log := c.log.WithEvent(req.Event)

	adapter := &v1alpha1.HTTPAdapter{}

	if err := req.Event.Spec(&adapter.Spec); err != nil {
		cancel()
		return core.ErrInvalid(fmt.Errorf("error parsing adapter spec: %v", err))
	}

	httpReq, err := req.Event.HTTPRequest(ctx)
	if err != nil {
		cancel()
		return core.ErrInvalid(err)
	}
	if adapterURL, err := url.Parse(adapter.Spec.URL); err != nil { // success
		cancel()
		return core.ErrInvalid(fmt.Errorf("error parsing adapter url: %v", err))

	} else {
		if adapterURL.Path == "" {
			adapterURL.Path = "/"
		}
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

	for k, v := range adapter.Spec.Headers {
		if strings.EqualFold(k, api.HeaderHost) {
			httpReq.Host = v
		}
		httpReq.Header.Set(k, v)
	}

	comp := core.NewPlatformComponent(
		api.ComponentTypeHTTPAdapter,
		req.Event.Target.Name,
		c.brk.Component.Hash,
	)
	comp.Id, comp.BrokerId = c.brk.Component.Id, c.brk.Component.BrokerId

	resp := core.NewResp(core.EventOpts{
		Parent: req.Event,
		Target: req.Event.Source,
		Source: comp,
	})

	go func() {
		defer cancel()

		var reqErr error
		if httpResp, err := c.adapterClient(adapter).Do(httpReq); err != nil {
			reqErr = core.ErrUnexpected(fmt.Errorf("http request failed: %v", err))
		} else {
			reqErr = resp.SetHTTPResponse(httpResp, MaxEventSize)
		}
		if reqErr != nil {
			if !errors.Is(reqErr, &core.Err{}) {
				reqErr = core.ErrUnexpected(reqErr)
			}
			resp.Type = string(api.EventTypeError)
			resp.SetJSON(reqErr)

			log.Debug(reqErr)
		}

		c.brk.SendResp(resp, req.ReceivedAt)
	}()

	return nil
}

func (c *HTTPClient) adapterClient(a *v1alpha1.HTTPAdapter) *http.Client {
	key := clientKey(a.Spec.FollowRedirects, a.Spec.InsecureSkipVerify)
	client := c.clients[key]
	if client == nil {
		client = c.clients["default"]
	}
	return client
}

func clientKey(follow api.FollowRedirects, insecure bool) string {
	if follow == "" {
		follow = api.FollowRedirectsNever
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
