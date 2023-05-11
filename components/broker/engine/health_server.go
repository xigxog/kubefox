package engine

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/xigxog/kubefox/libs/core/kubefox"
)

type HealthServer struct {
	Broker
	httpSrv *http.Server
}

func NewHealthServer(brk Broker) *HealthServer {
	return &HealthServer{
		Broker: brk,
	}
}

func (srv *HealthServer) Start() {
	srv.httpSrv = &http.Server{
		Addr:         srv.Config().HealthSrvAddr,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 30,
		IdleTimeout:  time.Second * 60,
		Handler:      srv,
	}

	go func() {
		err := srv.httpSrv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			srv.Log().Error(err)
			os.Exit(kubefox.TelemetryServerErrorCode)
		}
	}()

	srv.Log().Infof("health server started on %s", srv.Config().HealthSrvAddr)
}

func (srv *HealthServer) Shutdown(ctx context.Context) {
	srv.Log().Info("health server shutting down")

	if srv.httpSrv != nil {
		if err := srv.httpSrv.Shutdown(ctx); err != nil {
			srv.Log().Error(err)
		}
	}
}

func (srv *HealthServer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if srv.IsHealthy(req.Context()) {
		resp.WriteHeader(http.StatusOK)
	} else {
		resp.WriteHeader(http.StatusServiceUnavailable)
	}
}
