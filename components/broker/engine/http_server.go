package engine

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/xigxog/kubefox/components/broker/config"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/logkf"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/contrib/propagators/jaeger"
	"go.opentelemetry.io/otel/propagation"
)

type HTTPServer struct {
	wrapped *http.Server

	brk    Broker
	comp   *kubefox.Component
	sub    ReplicaSubscription
	reqMap map[string]chan *kubefox.Event

	mutex sync.Mutex

	propagator propagation.TextMapPropagator

	log *logkf.Logger
}

func NewHTTPServer(brk Broker) *HTTPServer {
	comp := &kubefox.Component{
		Name:   "kf-http-srv-adapt",
		Commit: kubefox.GitCommit,
		Id:     uuid.NewString(),
	}
	return &HTTPServer{
		brk:        brk,
		comp:       comp,
		reqMap:     make(map[string]chan *kubefox.Event),
		propagator: propagation.NewCompositeTextMapPropagator(jaeger.Jaeger{}, b3.New()),
		log:        logkf.Global,
	}
}

func (srv *HTTPServer) Start() (err error) {
	if config.HTTPSrvAddr == "false" && config.HTTPSSrvAddr == "false" {
		return nil
	}

	srv.wrapped = &http.Server{
		Handler: srv,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	srv.sub, err = srv.brk.Subscribe(context.Background(), &SubscriptionConf{
		Component:   srv.comp,
		SendFunc:    srv.sendEvent,
		EnableGroup: false,
	})
	if err != nil {
		return srv.log.ErrorN("%v", err)
	}

	// Start listener outside of goroutine to deal with address and port issues.
	if config.HTTPSrvAddr != "false" {
		srv.log.WithComponent(srv.comp).Debug("http server starting")
		ln, err := net.Listen("tcp", config.HTTPSrvAddr)
		if err != nil {
			return srv.log.ErrorN("%v", err)
		}
		go func() {
			err := srv.wrapped.Serve(ln)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				srv.log.Error(err)
				os.Exit(HTTPServerExitCode)
			}
		}()
	}
	if config.HTTPSSrvAddr != "false" {
		srv.log.WithComponent(srv.comp).Debug("https server starting")
		lns, err := net.Listen("tcp", config.HTTPSSrvAddr)
		if err != nil {
			return srv.log.ErrorN("%v", err)
		}
		go func() {
			err := srv.wrapped.ServeTLS(lns, kubefox.TLSCertPath, kubefox.TLSKeyPath)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				srv.log.Error(err)
				os.Exit(HTTPServerExitCode)
			}
		}()
	}

	srv.log.Info("http servers started")
	return nil
}

func (srv *HTTPServer) Shutdown(timeout time.Duration) {
	srv.log.Info("http server shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if srv.wrapped != nil {
		if err := srv.wrapped.Shutdown(ctx); err != nil {
			srv.log.Error(err)
		}
	}

	if srv.sub != nil {
		srv.sub.Cancel(nil)
	}
}

func (srv *HTTPServer) ServeHTTP(resWriter http.ResponseWriter, httpReq *http.Request) {
	// ctx := srv.propagator.Extract(httpReq.Context(), propagation.HeaderCarrier(httpReq.Header))

	ctx, cancel := context.WithTimeoutCause(httpReq.Context(), config.EventTTL, ErrEventTimeout)
	defer func() {
		cancel()
	}()

	req := kubefox.NewEvent()
	req.Ttl = config.EventTTL.Microseconds()
	req.Category = kubefox.Category_CATEGORY_REQUEST
	req.Source = srv.comp
	if err := req.SetHTTPRequest(httpReq); err != nil {
		srv.log.Debugf("error parsing event from http request: %v", err)
		resWriter.WriteHeader(http.StatusBadRequest)
		resWriter.Write([]byte(err.Error()))
		return
	}

	log := srv.log.WithEvent(req)

	srv.mutex.Lock()
	respCh := make(chan *kubefox.Event)
	srv.reqMap[req.Id] = respCh
	srv.mutex.Unlock()

	rEvt := &ReceivedEvent{
		Event:    req,
		Receiver: HTTPSrvSvc,
		ErrCh:    make(chan error),
	}
	if err := srv.brk.RecvEvent(rEvt); err != nil {
		writeError(resWriter, context.Cause(ctx), log)
		return
	}

	var resp *kubefox.Event
	select {
	case resp = <-respCh:
		log.DebugEw("received response", resp)

	case err := <-rEvt.ErrCh:
		writeError(resWriter, err, log)
		return

	case <-ctx.Done():
		writeError(resWriter, context.Cause(ctx), log)
		return
	}

	resWriter.Header().Set("Kf-Adapter", srv.comp.Key())
	if resp.TraceId() != "" {
		resWriter.Header().Set("Kf-Trace-Id", resp.TraceId())
	}

	var statusCode int
	switch {
	case resp.Type == string(kubefox.ErrorEventType):
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

func (srv *HTTPServer) Component() *kubefox.Component {
	return srv.comp
}

func (srv *HTTPServer) Subscription() Subscription {
	return srv.sub
}

func (srv *HTTPServer) sendEvent(mEvt *kubefox.MatchedEvent) error {
	resp := mEvt.Event
	srv.mutex.Lock()
	respCh, found := srv.reqMap[resp.ParentId]
	delete(srv.reqMap, resp.ParentId)
	srv.mutex.Unlock()

	if !found {
		err := fmt.Errorf("matching request not found for response")
		srv.log.DebugEw(err.Error(), resp)
		return err
	}

	respCh <- resp

	return nil
}

func writeError(resWriter http.ResponseWriter, err error, log *logkf.Logger) {
	log.Debugf("sending event failed: %v", err)
	switch {
	case errors.Is(err, ErrEventTimeout):
		resWriter.WriteHeader(http.StatusGatewayTimeout)
	case errors.Is(err, ErrEventInvalid):
		resWriter.WriteHeader(http.StatusBadRequest)
	case errors.Is(err, ErrRouteNotFound):
		resWriter.WriteHeader(http.StatusNotFound)
	default:
		resWriter.WriteHeader(http.StatusInternalServerError)
	}
	resWriter.Write([]byte(err.Error()))
}
