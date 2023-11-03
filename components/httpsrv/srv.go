package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"

	kubefox "github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/grpc"
	"github.com/xigxog/kubefox/logkf"
)

const (
	maxAttempts = 5
)

type HTTPSrv struct {
	wrapped *http.Server

	brk *grpc.Client

	log *logkf.Logger
}

func NewHTTPSrv() *HTTPSrv {
	return &HTTPSrv{
		brk: grpc.NewClient(grpc.ClientOpts{
			Component:     comp,
			BrokerAddr:    brokerAddr,
			HealthSrvAddr: healthAddr,
		}),

		log: logkf.Global.WithComponent(comp),
	}
}

func (srv *HTTPSrv) Run() error {
	if httpAddr == "false" && httpsAddr == "false" {
		return nil
	}

	if healthAddr != "" && healthAddr != "false" {
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
	if httpAddr != "false" {
		srv.log.Debug("http server starting")
		ln, err := net.Listen("tcp", httpAddr)
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
	if httpsAddr != "false" {
		srv.log.Debug("https server starting")
		ln, err := net.Listen("tcp", httpsAddr)
		if err != nil {
			return srv.log.ErrorN("%v", err)
		}
		go func() {
			err := srv.wrapped.ServeTLS(ln, kubefox.PathTLSCert, kubefox.PathTLSKey)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				srv.log.Fatal(err)
			}
		}()
		srv.log.Info("https server started")
	}

	go srv.brk.Start(spec, maxAttempts)

	return <-srv.brk.Err()
}

func (srv *HTTPSrv) Shutdown() {
	srv.log.Info("http server shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), eventTTL)
	defer cancel()

	if srv.wrapped != nil {
		if err := srv.wrapped.Shutdown(ctx); err != nil {
			srv.log.Error(err)
		}
	}
}

func (srv *HTTPSrv) ServeHTTP(resWriter http.ResponseWriter, httpReq *http.Request) {
	ctx, cancel := context.WithTimeoutCause(httpReq.Context(), eventTTL, kubefox.ErrEventTimeout)
	defer cancel()

	resWriter.Header().Set(kubefox.HeaderAdapter, comp.Key())

	req := kubefox.NewReq(kubefox.EventOpts{
		Source: comp,
	})
	req.Ttl = eventTTL.Microseconds()
	if err := req.SetHTTPRequest(httpReq); err != nil {
		err = fmt.Errorf("%w: error parsing event: %v", kubefox.ErrEventInvalid, err)
		writeError(resWriter, err, srv.log)
		return
	}

	log := srv.log.WithEvent(req)
	log.Debug("receive request")

	// TODO broker needs to return error for route not found
	resp, err := srv.brk.SendReq(ctx, req)
	if err != nil {
		writeError(resWriter, err, log)
		return
	}

	if resp.TraceId() != "" {
		resWriter.Header().Set(kubefox.HeaderTraceId, resp.TraceId())
	}

	var statusCode int
	switch {
	case resp.Type == string(kubefox.EventTypeError):
		statusCode = http.StatusInternalServerError

	default:
		httpResp := resp.HTTPResponse()
		statusCode = httpResp.StatusCode
		for key, val := range httpResp.Header {
			for _, h := range val {
				resWriter.Header().Add(key, h)
			}
		}
	}

	resWriter.Header().Set("Content-Type", resp.ContentType)
	resWriter.Header().Set("Content-Length", strconv.Itoa(len(resp.Content)))
	resWriter.WriteHeader(statusCode)
	resWriter.Write(resp.Content)
}

func writeError(resWriter http.ResponseWriter, err error, log *logkf.Logger) {
	log.Debugf("event failed: %v", err)
	switch {
	case errors.Is(err, kubefox.ErrEventTimeout):
		resWriter.WriteHeader(http.StatusGatewayTimeout)
	case errors.Is(err, kubefox.ErrEventInvalid):
		resWriter.WriteHeader(http.StatusBadRequest)
	case errors.Is(err, kubefox.ErrRouteNotFound):
		resWriter.WriteHeader(http.StatusNotFound)
	default:
		resWriter.WriteHeader(http.StatusInternalServerError)
	}
	resWriter.Write([]byte(err.Error()))
}
