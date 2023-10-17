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
	id, err := os.Hostname()
	if err != nil || id == "" {
		id = uuid.NewString()
	}

	comp := &kubefox.Component{
		Name:   "http-server",
		Commit: kubefox.GitCommit,
		Id:     id,
	}
	return &HTTPServer{
		brk:        brk,
		comp:       comp,
		reqMap:     make(map[string]chan *kubefox.Event),
		propagator: propagation.NewCompositeTextMapPropagator(jaeger.Jaeger{}, b3.New()),
		log:        logkf.Global.WithComponent(comp),
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
		srv.log.Debug("http server starting")
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
		srv.log.Debug("https server starting")
		lns, err := net.Listen("tcp", config.HTTPSSrvAddr)
		if err != nil {
			return srv.log.ErrorN("%v", err)
		}
		go func() {
			err := srv.wrapped.ServeTLS(lns, kubefox.PathTLSCert, kubefox.PathTLSKey)
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
	defer cancel()

	resWriter.Header().Set(kubefox.HeaderAdapter, srv.comp.Key())

	req := kubefox.StartReq(kubefox.EventOpts{
		Source: srv.comp,
	})
	req.Ttl = config.EventTTL.Microseconds()
	if err := req.SetHTTPRequest(httpReq); err != nil {
		err = fmt.Errorf("%w: error parsing event: %v", ErrEventInvalid, err)
		writeError(resWriter, err, srv.log)
		return
	}

	log := srv.log.WithEvent(req.Event)
	log.Debug("received request")

	srv.mutex.Lock()
	respCh := make(chan *kubefox.Event)
	srv.reqMap[req.Id] = respCh
	srv.mutex.Unlock()

	defer func() {
		srv.mutex.Lock()
		delete(srv.reqMap, req.Id)
		srv.mutex.Unlock()
	}()

	rEvt := &ReceivedEvent{
		ActiveEvent: req,
		Receiver:    ReceiverHTTPServer,
		ErrCh:       make(chan error),
	}
	if err := srv.brk.RecvEvent(rEvt); err != nil {
		writeError(resWriter, context.Cause(ctx), log)
		return
	}

	var resp *kubefox.Event
	select {
	case resp = <-respCh:
		// Reset log attributes with response.
		log = srv.log.WithEvent(resp)
		log.Debug("received response")

	case err := <-rEvt.ErrCh:
		writeError(resWriter, err, log)
		return

	case <-ctx.Done():
		writeError(resWriter, context.Cause(ctx), log)
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
	srv.mutex.Unlock()

	if !found {
		return ErrEventRequestGone
	}

	respCh <- resp

	return nil
}

func writeError(resWriter http.ResponseWriter, err error, log *logkf.Logger) {
	log.Debugf("event failed: %v", err)
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
