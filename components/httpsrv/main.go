// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/build"
	"github.com/xigxog/kubefox/components/httpsrv/adapter"
	"github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/grpc"
	"github.com/xigxog/kubefox/logkf"
	"github.com/xigxog/kubefox/telemetry"
	"github.com/xigxog/kubefox/utils"
)

func main() {
	var name, hash, pod string
	var logFormat, logLevel, tokenPath string
	flag.StringVar(&adapter.Platform, "platform", "", "KubeFox Platform name. (required)")
	flag.StringVar(&name, "name", "", `Component name. (required)`)
	flag.StringVar(&hash, "hash", "", `Hash the Component was built from. (required)`)
	flag.StringVar(&pod, "pod", "", `Component pod. (required)`)
	flag.StringVar(&adapter.HTTPAddr, "http-addr", "127.0.0.1:8080", `Address and port the HTTP server should bind to, set to "false" to disable.`)
	flag.StringVar(&adapter.HTTPSAddr, "https-addr", "127.0.0.1:8443", `Address and port the HTTPS server should bind to, set to "false" to disable.`)
	flag.StringVar(&adapter.BrokerAddr, "broker-addr", "127.0.0.1:6060", "Address and port of the Broker gRPC server.")
	flag.StringVar(&adapter.HealthSrvAddr, "health-addr", "127.0.0.1:1111", `Address and port the HTTP health server should bind to, set to "false" to disable.`)
	flag.Int64Var(&adapter.MaxEventSize, "max-event-size", api.DefaultMaxEventSizeBytes, "Maximum size of event in bytes.")
	flag.IntVar(&adapter.WorkerCount, "http-worker-count", runtime.NumCPU()*2, "The number of workers to listen for events in the HTTP server.")
	flag.DurationVar(&adapter.EventTimeout, "timeout", time.Minute, "Default timeout for an event.")
	flag.StringVar(&logFormat, "log-format", "console", "Log format. [options 'json', 'console']")
	flag.StringVar(&logLevel, "log-level", "debug", "Log level. [options 'debug', 'info', 'warn', 'error']")
	flag.StringVar(&tokenPath, "token-path", api.PathSvcAccToken, "Path to Service Account Token")
	flag.Parse()

	utils.CheckRequiredFlag("platform", adapter.Platform)
	utils.CheckRequiredFlag("name", name)
	utils.CheckRequiredFlag("hash", hash)
	utils.CheckRequiredFlag("pod", pod)

	if hash != build.Info.Hash &&
		!(hash == "debug" && build.Info.Hash == "") {

		fmt.Fprintf(os.Stderr, "hash '%s' does not match build info hash '%s'", hash, build.Info.Hash)
		os.Exit(1)
	}

	comp := core.NewPlatformComponent(
		api.ComponentTypeHTTPAdapter,
		name,
		hash,
	)
	comp.Id = core.GenerateId()

	logkf.Global = logkf.
		BuildLoggerOrDie(logFormat, logLevel).
		WithComponent(comp)
	defer logkf.Global.Sync()

	telemetry.SetComponent(comp)

	broker := grpc.NewClient(grpc.ClientOpts{
		Platform:      adapter.Platform,
		Component:     comp,
		Pod:           pod,
		BrokerAddr:    adapter.BrokerAddr,
		HealthSrvAddr: adapter.HealthSrvAddr,
		TokenPath:     tokenPath,
	})

	httpClient := adapter.NewHTTPClient(broker)

	srv := adapter.New(broker, httpClient)
	defer srv.Shutdown()

	listener := adapter.NewListener(broker, httpClient)
	listener.StartWorkers(adapter.WorkerCount)

	if err := srv.Run(); err != nil {
		logkf.Global.Fatal(err)
	}
}
