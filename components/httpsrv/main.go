package main

import (
	"flag"
	"time"

	"github.com/xigxog/kubefox/build"
	kubefox "github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/logkf"
	"github.com/xigxog/kubefox/utils"
)

var (
	name                   string
	httpAddr, httpsAddr    string
	brokerAddr, healthAddr string
	logFormat, logLevel    string
	eventTTL               time.Duration
	comp                   *kubefox.Component
	spec                   *kubefox.ComponentSpec
)

func main() {
	flag.StringVar(&name, "name", "", `Name of httpsrv. (required)`)
	flag.StringVar(&httpAddr, "http-addr", "127.0.0.1:8080", `Address and port the HTTP server should bind to, set to "false" to disable.`)
	flag.StringVar(&httpsAddr, "https-addr", "127.0.0.1:8443", `Address and port the HTTPS server should bind to, set to "false" to disable.`)
	flag.StringVar(&brokerAddr, "broker-addr", "127.0.0.1:6060", "Address and port of the Broker gRPC server.")
	flag.StringVar(&healthAddr, "health-addr", "127.0.0.1:1111", `Address and port the HTTP health server should bind to, set to "false" to disable.`)
	flag.DurationVar(&eventTTL, "ttl", time.Minute, "Default time-to-live for an event.")
	flag.StringVar(&logFormat, "log-format", "console", "Log format. [options 'json', 'console']")
	flag.StringVar(&logLevel, "log-level", "debug", "Log level. [options 'debug', 'info', 'warn', 'error']")
	flag.Parse()

	utils.CheckRequiredFlag("name", name)

	_, id := kubefox.GenerateNameAndId()
	comp = &kubefox.Component{
		Name:   name,
		Commit: build.Info.Commit,
		Id:     id,
	}
	spec = &kubefox.ComponentSpec{
		ComponentTypeVar: kubefox.ComponentTypeVar{
			Type: kubefox.ComponentTypeGenesis,
		},
	}

	logkf.Global = logkf.
		BuildLoggerOrDie(logFormat, logLevel).
		WithComponent(comp)
	defer logkf.Global.Sync()

	srv := NewHTTPServer()
	defer srv.Shutdown()

	if err := srv.Run(); err != nil {
		logkf.Global.Fatal(err)
	}
}
