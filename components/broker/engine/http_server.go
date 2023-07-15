package engine

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/utils"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/contrib/propagators/jaeger"
	"go.opentelemetry.io/otel/propagation"
)

type HTTPServer struct {
	Broker

	httpSrv    *http.Server
	propagator propagation.TextMapPropagator
}

func NewHTTPServer(brk Broker) *HTTPServer {
	return &HTTPServer{
		Broker:     brk,
		propagator: propagation.NewCompositeTextMapPropagator(jaeger.Jaeger{}, b3.New()),
	}
}

func (srv *HTTPServer) Start() {
	srv.httpSrv = &http.Server{
		Addr:         srv.Config().HTTPSrvAddr,
		WriteTimeout: srv.EventTimeout(),
		ReadTimeout:  srv.EventTimeout(),
		IdleTimeout:  srv.EventTimeout(),
		Handler:      srv,
	}

	go func() {
		err := srv.httpSrv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			srv.Log().Error(err)
			os.Exit(kubefox.HTTPServerErrorCode)
		}
	}()

	srv.Log().Infof("http server started on %s", srv.Config().HTTPSrvAddr)
}

func (srv *HTTPServer) Shutdown(ctx context.Context) {
	srv.Log().Info("HTTP server shutting down")

	if srv.httpSrv != nil {
		if err := srv.httpSrv.Shutdown(ctx); err != nil {
			srv.Log().Error(err)
		}
	}
}

func (srv *HTTPServer) ServeHTTP(resWriter http.ResponseWriter, httpReq *http.Request) {
	ctx := srv.propagator.Extract(httpReq.Context(), propagation.HeaderCarrier(httpReq.Header))
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
