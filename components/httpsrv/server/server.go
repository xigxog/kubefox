// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/grpc"
	"github.com/xigxog/kubefox/logkf"
	"github.com/xigxog/kubefox/telemetry"

	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
)

const (
	maxAttempts = 5
)

type Server struct {
	sync.Mutex

	wrapped *http.Server
	brk     *grpc.Client

	log *logkf.Logger
}

func New(comp *core.Component, pod string) *Server {
	return &Server{
		brk: grpc.NewClient(grpc.ClientOpts{
			Platform:      Platform,
			Component:     comp,
			Pod:           pod,
			BrokerAddr:    BrokerAddr,
			HealthSrvAddr: HealthSrvAddr,
		}),
		log: logkf.Global,
	}
}

func (srv *Server) Run() error {
	if HTTPAddr == "false" && HTTPSAddr == "false" {
		return nil
	}

	if HealthSrvAddr != "" && HealthSrvAddr != "false" {
		if err := srv.brk.StartHealthSrv(); err != nil {
			return err
		}
	}

	srv.wrapped = &http.Server{
		Handler: srv,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	// Start listener outside of goroutine to deal with address and port issues.
	if HTTPAddr != "false" {
		srv.log.Debug("http server starting")
		ln, err := net.Listen("tcp", HTTPAddr)
		if err != nil {
			return srv.log.ErrorN("%v", err)
		}
		go func() {
			err := srv.wrapped.Serve(ln)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				srv.log.Fatal(err)
			}
		}()
		srv.log.Info("http server started")
	}
	if HTTPSAddr != "false" {
		srv.log.Debug("https server starting")
		ln, err := net.Listen("tcp", HTTPSAddr)
		if err != nil {
			return srv.log.ErrorN("%v", err)
		}
		go func() {
			err := srv.wrapped.ServeTLS(ln, api.PathTLSCert, api.PathTLSKey)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				srv.log.Fatal(err)
			}
		}()
		srv.log.Info("https server started")
	}

	go srv.brk.Start(&api.ComponentDefinition{Type: api.ComponentTypeHTTPAdapter}, maxAttempts)

	return <-srv.brk.Err()
}

func (srv *Server) Shutdown() {
	srv.log.Info("http server shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), EventTimeout)
	defer cancel()

	if srv.wrapped != nil {
		if err := srv.wrapped.Shutdown(ctx); err != nil {
			srv.log.Error(err)
		}
	}
}

// IDEA have `kubefox-set-cookie` which takes dynamic context and puts it into
// cookie so do not have to set query params?
func (srv *Server) ServeHTTP(resWriter http.ResponseWriter, httpReq *http.Request) {
	ctx, cancel := context.WithTimeoutCause(httpReq.Context(), EventTimeout, core.ErrTimeout())
	defer cancel()

	setHeader(resWriter, api.HeaderAdapter, srv.brk.Component.Key())

	var parentTrace *core.SpanContext
	_, _, otelSC := otelhttptrace.Extract(ctx, httpReq)
	if otelSC.HasTraceID() {
		parentTrace = telemetry.SpanContextFromOTEL(otelSC)
	}

	var spans []*telemetry.Span
	span := telemetry.StartSpan(fmt.Sprintf("%s %s", httpReq.Method, httpReq.URL), parentTrace)
	spans = append(spans, span)

	defer func() {
		span.End()
		srv.brk.SendSpans(spans...)
	}()

	parseSpan := telemetry.StartSpan("parse http request", span.SpanContext())
	spans = append(spans, parseSpan)

	req := core.NewReq(core.EventOpts{
		Source:      srv.brk.Component,
		Timeout:     EventTimeout,
		TraceParent: span.SpanContext(),
	})
	setHeader(resWriter, api.HeaderEventId, req.Id)

	if err := req.SetHTTPRequest(httpReq, MaxEventSize); err != nil {
		writeError(resWriter, err, srv.log)
		return
	}
	parseSpan.End()

	log := srv.log.WithEvent(req)
	log.Debug("receive request")

	resp, err := srv.brk.SendReq(ctx, req, time.Now())

	respSpan := telemetry.StartSpan("send http response", span.SpanContext())
	spans = append(spans, respSpan)
	defer respSpan.End()

	// Add Event Context to response headers.
	if resp != nil && resp.Context != nil {
		setHeader(resWriter, api.HeaderPlatform, resp.Context.Platform)
		setHeader(resWriter, api.HeaderVirtualEnvironment, resp.Context.VirtualEnvironment)
		setHeader(resWriter, api.HeaderAppDep, resp.Context.AppDeployment)
		setHeader(resWriter, api.HeaderReleaseManifest, resp.Context.ReleaseManifest)
	}

	switch {
	case err != nil:
		writeError(resWriter, err, log)
		return
	case resp.Err() != nil:
		writeError(resWriter, resp.Err(), log)
		return
	}

	httpResp := resp.HTTPResponse()
	for key, val := range httpResp.Header {
		for _, h := range val {
			resWriter.Header().Add(key, h)
		}
	}
	setHeader(resWriter, api.HeaderContentType, resp.ContentType)
	setHeader(resWriter, api.HeaderContentLength, strconv.Itoa(len(resp.Content)))
	resWriter.WriteHeader(httpResp.StatusCode)
	resWriter.Write(resp.Content)
}

// setHeader will set the header on the http.ResponseWriter if the value is not
// empty.
func setHeader(resWriter http.ResponseWriter, key, value string) {
	if value == "" {
		return
	}
	resWriter.Header().Set(key, value)
}

func writeError(resWriter http.ResponseWriter, err error, log *logkf.Logger) {
	log.Debugf("event failed: %v", err)

	statusCode := http.StatusInternalServerError
	kfErr := &core.Err{}
	if ok := errors.As(err, &kfErr); ok {
		statusCode = kfErr.HTTPCode()
	}

	resWriter.WriteHeader(statusCode)
	resWriter.Write([]byte(err.Error()))
}
