// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package server

import (
	"fmt"
	"sync"

	"github.com/xigxog/kubefox/grpc"
	"github.com/xigxog/kubefox/logkf"
)

type Listener struct {
	sync.Mutex

	brk        *grpc.Client
	wg         sync.WaitGroup
	httpClient *HTTPClient

	log *logkf.Logger
}

func NewListener(broker *grpc.Client, httpClient *HTTPClient) *Listener {
	return &Listener{
		brk:        broker,
		httpClient: httpClient,
		log:        logkf.Global,
	}
}

func (listener *Listener) StartWorkers(workerCount int) {
	for i := 0; i < workerCount; i++ {
		listener.wg.Add(1)
		go listener.StartWorker(i)
	}
}

func (listener *Listener) StartWorker(id int) {
	log := listener.log.With(logkf.KeyWorker, fmt.Sprintf("worker-%d", id))
	log.Info("http listener started")
	defer func() {
		log.Info("worker stopped")
		listener.wg.Done()
	}()

	for req := range listener.brk.Req() {
		listener.ReceiveEvent(req)
	}
}

func (listener *Listener) ReceiveEvent(req *grpc.ComponentEvent) {
	listener.log.WithEvent(req.Event)
	listener.httpClient.SendEvent(req)
}
