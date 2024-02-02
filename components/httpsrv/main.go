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
	"time"

	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/build"
	"github.com/xigxog/kubefox/components/httpsrv/server"
	"github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/logkf"
	"github.com/xigxog/kubefox/utils"
)

func main() {
	var name, commit, pod string
	var logFormat, logLevel string
	flag.StringVar(&server.Platform, "platform", "", "KubeFox Platform name. (required)")
	flag.StringVar(&name, "name", "", `Component name. (required)`)
	flag.StringVar(&commit, "commit", "", `Commit the Component was built from. (required)`)
	flag.StringVar(&pod, "pod", "", `Component pod. (required)`)
	flag.StringVar(&server.HTTPAddr, "http-addr", "127.0.0.1:8080", `Address and port the HTTP server should bind to, set to "false" to disable.`)
	flag.StringVar(&server.HTTPSAddr, "https-addr", "127.0.0.1:8443", `Address and port the HTTPS server should bind to, set to "false" to disable.`)
	flag.StringVar(&server.BrokerAddr, "broker-addr", "127.0.0.1:6060", "Address and port of the Broker gRPC server.")
	flag.StringVar(&server.HealthSrvAddr, "health-addr", "127.0.0.1:1111", `Address and port the HTTP health server should bind to, set to "false" to disable.`)
	flag.Int64Var(&server.MaxEventSize, "max-event-size", api.DefaultMaxEventSizeBytes, "Maximum size of event in bytes.")
	flag.DurationVar(&server.EventTimeout, "timeout", time.Minute, "Default timeout for an event.")
	flag.StringVar(&logFormat, "log-format", "console", "Log format. [options 'json', 'console']")
	flag.StringVar(&logLevel, "log-level", "debug", "Log level. [options 'debug', 'info', 'warn', 'error']")
	flag.Parse()

	utils.CheckRequiredFlag("platform", server.Platform)
	utils.CheckRequiredFlag("name", name)
	utils.CheckRequiredFlag("commit", commit)
	utils.CheckRequiredFlag("pod", pod)

	if commit != build.Info.Commit {
		fmt.Fprintf(os.Stderr, "commit '%s' does not match build info commit '%s'", commit, build.Info.Commit)
		os.Exit(1)
	}

	comp := core.NewPlatformComponent(
		api.ComponentTypeHTTPAdapter,
		name,
		commit,
	)
	comp.Id = core.GenerateId()

	logkf.Global = logkf.
		BuildLoggerOrDie(logFormat, logLevel).
		WithComponent(comp)
	defer logkf.Global.Sync()

	srv := server.New(comp, pod)
	defer srv.Shutdown()

	if err := srv.Run(); err != nil {
		logkf.Global.Fatal(err)
	}
}
