package server

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"strconv"
	"time"

	kubefox "github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/grpc"
	"github.com/xigxog/kubefox/logkf"
)

const (
	maxAttempts = 5
)

type Server struct {
	wrapped *http.Server

	brk *grpc.Client

	log *logkf.Logger
}

func New() *Server {
	return &Server{
		brk: grpc.NewClient(grpc.ClientOpts{
			Component:     Component,
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
			err := srv.wrapped.ServeTLS(ln, kubefox.PathTLSCert, kubefox.PathTLSKey)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				srv.log.Fatal(err)
			}
		}()
		srv.log.Info("https server started")
	}

	go srv.brk.Start(Spec, maxAttempts)

	return <-srv.brk.Err()
}

func (srv *Server) Shutdown() {
	srv.log.Info("http server shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), EventTTL)
	defer cancel()

	if srv.wrapped != nil {
		if err := srv.wrapped.Shutdown(ctx); err != nil {
			srv.log.Error(err)
		}
	}
}

func (srv *Server) ServeHTTP(resWriter http.ResponseWriter, httpReq *http.Request) {
	ctx, cancel := context.WithTimeoutCause(httpReq.Context(), EventTTL, kubefox.ErrTimeout())
	defer cancel()

	resWriter.Header().Set(kubefox.HeaderAdapter, Component.Key())

	req := kubefox.NewReq(kubefox.EventOpts{
		Source: Component,
	})
	req.SetTTL(EventTTL)
	if err := req.SetHTTPRequest(httpReq); err != nil {
		writeError(resWriter, err, srv.log)
		return
	}

	log := srv.log.WithEvent(req)
	log.Debug("receive request")

	resp, err := srv.brk.SendReq(ctx, req, time.Now())
	switch {
	case err != nil:
		writeError(resWriter, err, log)
		return
	case resp.Err() != nil:
		writeError(resWriter, resp.Err(), log)
		return
	}

	if resp.TraceId() != "" {
		resWriter.Header().Set(kubefox.HeaderTraceId, resp.TraceId())
	}

	httpResp := resp.HTTPResponse()
	for key, val := range httpResp.Header {
		for _, h := range val {
			resWriter.Header().Add(key, h)
		}
	}
	resWriter.Header().Set("Content-Type", resp.ContentType)
	resWriter.Header().Set("Content-Length", strconv.Itoa(len(resp.Content)))
	resWriter.WriteHeader(httpResp.StatusCode)
	resWriter.Write(resp.Content)
}

func writeError(resWriter http.ResponseWriter, err error, log *logkf.Logger) {
	log.Debugf("event failed: %v", err)

	statusCode := http.StatusInternalServerError
	kfErr := &kubefox.Err{}
	if ok := errors.As(err, &kfErr); ok {
		statusCode = kfErr.HTTPCode()
	}

	resWriter.WriteHeader(statusCode)
	resWriter.Write([]byte(err.Error()))
}
