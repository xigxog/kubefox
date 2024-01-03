// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package telemetry

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/xigxog/kubefox/components/broker/config"
	"github.com/xigxog/kubefox/logkf"
)

type HealthProvider interface {
	Name() string
	IsHealthy(ctx context.Context) bool
}

type HealthServer struct {
	httpSrv   *http.Server
	providers []HealthProvider

	mutex sync.Mutex

	log *logkf.Logger
}

func NewHealthServer() *HealthServer {
	return &HealthServer{
		providers: make([]HealthProvider, 0),
		log:       logkf.Global,
	}
}

func (srv *HealthServer) Register(provider HealthProvider) {
	srv.mutex.Lock()
	defer srv.mutex.Unlock()

	srv.providers = append(srv.providers, provider)
}

func (srv *HealthServer) Start() error {
	srv.log.Debug("health server starting")

	srv.httpSrv = &http.Server{
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 30,
		IdleTimeout:  time.Second * 30,
		Handler:      srv,
	}

	ln, err := net.Listen("tcp", config.HealthSrvAddr)
	if err != nil {
		return err
	}

	go func() {
		err := srv.httpSrv.Serve(ln)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			srv.log.Error(err)
			os.Exit(1)
		}
	}()

	srv.log.Info("health server started")
	return nil
}

func (srv *HealthServer) Shutdown(timeout time.Duration) {
	srv.log.Info("health server shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if srv.httpSrv != nil {
		if err := srv.httpSrv.Shutdown(ctx); err != nil {
			srv.log.Error(err)
		}
	}
}

func (srv *HealthServer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), time.Second*10)
	defer cancel()

	status := http.StatusOK
	for _, p := range srv.providers {
		if !p.IsHealthy(ctx) {
			srv.log.Errorf("%s is unhealthy", p.Name())
			status = http.StatusServiceUnavailable
		}
	}
	resp.WriteHeader(status)
}
