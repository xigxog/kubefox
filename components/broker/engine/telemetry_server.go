package engine

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/logger"
)

const (
	gatherTimeout = 5 * time.Second
)

type HealthProvider interface {
	Healthy(context.Context) bool
	Name() string
}

type TelemetryServer struct {
	Broker

	grpcSrv    *GRPCServer
	providers  []HealthProvider
	httpServer *http.Server

	log *logger.Log
}

func NewTelemetryServer(brk Broker) *TelemetryServer {
	s := &TelemetryServer{
		Broker:    brk,
		providers: make([]HealthProvider, 0),
		log:       brk.Log(),
	}

	return s
}

func (srv *TelemetryServer) EnableComponentMetrics(grpcSrv *GRPCServer) {
	srv.grpcSrv = grpcSrv
}

func (srv *TelemetryServer) AddHealthProvider(provider HealthProvider) {
	srv.providers = append(srv.providers, provider)
}

func (srv *TelemetryServer) Serve() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", srv.handleHealth)

	srv.httpServer = &http.Server{
		Addr:    srv.Config().TelemetrySrvAddr,
		Handler: mux,
	}

	go func() {
		err := srv.httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			srv.log.Error(err)
			os.Exit(kubefox.TelemetryServerErrorCode)
		}
	}()

	srv.log.Infof("telemetry server started on %s", srv.Config().TelemetrySrvAddr)
}

func (srv *TelemetryServer) Shutdown(ctx context.Context) {
	srv.log.Info("telemetry server shutting down")

	if srv.httpServer != nil {
		err := srv.httpServer.Shutdown(ctx)
		if err != nil {
			srv.log.Error(err)
		}
	}
}

func (srv *TelemetryServer) handleHealth(writer http.ResponseWriter, req *http.Request) {
	provHealthMap := make(map[string]bool, len(srv.providers))
	healthy := true
	for _, provider := range srv.providers {
		provHealth := provider.Healthy(req.Context())
		provHealthMap[provider.Name()] = provHealth
		healthy = healthy && provHealth
	}

	data, err := json.Marshal(provHealthMap)
	if err != nil {
		srv.log.Error(err)
	}

	if healthy {
		writer.WriteHeader(http.StatusOK)
	} else {
		writer.WriteHeader(http.StatusBadRequest)
	}

	writer.Header().Add("Content-Type", "application/json")
	writer.Write(data)
}
