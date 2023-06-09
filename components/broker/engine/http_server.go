package engine

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/utils"
	"go.opentelemetry.io/contrib/propagators/jaeger"
	"go.opentelemetry.io/otel/propagation"
)

type HTTPServer struct {
	Broker

	server *http.Server
	jaeger jaeger.Jaeger
}

func NewHTTPServer(brk Broker) *HTTPServer {
	return &HTTPServer{
		Broker: brk,
		jaeger: jaeger.Jaeger{},
	}
}

func (srv *HTTPServer) Start() {
	srv.server = &http.Server{
		Addr: srv.Config().HTTPSrvAddr,
		// TODO get from config
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 30,
		IdleTimeout:  time.Second * 60,
		Handler:      srv,
	}

	go func() {
		err := srv.server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			srv.Log().Error(err)
			os.Exit(kubefox.HTTPServerErrorCode)
		}
	}()

	srv.Log().Infof("http server started on %s", srv.Config().HTTPSrvAddr)
}

func (srv *HTTPServer) Shutdown(ctx context.Context) {
	srv.Log().Info("HTTP server shutting down")

	err := srv.server.Shutdown(ctx)
	if err != nil {
		srv.Log().Error(err)
	}
}

func (srv *HTTPServer) ServeHTTP(resWriter http.ResponseWriter, httpReq *http.Request) {
	ctx := srv.jaeger.Extract(httpReq.Context(), propagation.HeaderCarrier(httpReq.Header))
	resWriter.Header().Set("Kf-Adapter", srv.Component().GetURI())

	req := kubefox.EmptyDataEvent()
	cfg := srv.Config()
	if cfg.IsDevMode && cfg.Dev.EventContext != nil {
		if h := utils.GetParamOrHeader(httpReq, kubefox.EnvHeader, kubefox.EnvHeaderShort); h == "" {
			httpReq.Header.Add(kubefox.EnvHeaderShort, cfg.Dev.Environment)
		}
		if h := utils.GetParamOrHeader(httpReq, kubefox.SysHeader, kubefox.SysHeaderShort); h == "" {
			httpReq.Header.Add(kubefox.SysHeaderShort, cfg.Dev.System)
		}
	}
	if cfg.IsDevMode && cfg.Dev.Target != nil {
		if h := utils.GetParamOrHeader(httpReq, kubefox.TargetHeader); h == "" {
			httpReq.Header.Add(kubefox.TargetHeader, cfg.Dev.Target.GetKey())
		}
	}

	if err := req.HTTPData().ParseRequest(httpReq); err != nil {
		srv.Log().Errorf("error parsing event from http req: %v", err)
		resWriter.WriteHeader(http.StatusBadRequest)
		resWriter.Write([]byte(err.Error()))
		return
	}

	resp := srv.InvokeRemoteComponent(ctx, req).HTTPData()
	if resp != nil && resp.GetSpan() != nil {
		resWriter.Header().Set("Kf-Trace-Id", resp.GetTraceId())
	}
	if resp.GetError() != nil {
		// TODO check accept header and respond with correct format
		resWriter.WriteHeader(http.StatusInternalServerError)
		resWriter.Write([]byte(resp.GetError().Error()))
		return
	}

	for _, key := range resp.GetHeaderKeys() {
		for _, val := range resp.GetHeaderValues(key) {
			resWriter.Header().Add(key, val)
		}
	}

	resWriter.Header().Set("Content-Type", resp.GetContentType())
	resWriter.Header().Set("Content-Length", strconv.Itoa(len(resp.GetContent())))
	resWriter.WriteHeader(resp.GetStatusCode())
	resWriter.Write(resp.GetContent())
}
