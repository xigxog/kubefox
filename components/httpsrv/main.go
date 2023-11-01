package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/xigxog/kubefox/build"
	kubefox "github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/logkf"
	"github.com/xigxog/kubefox/utils"
)

var (
	httpAddr, httpsAddr    string
	brokerAddr, healthAddr string
	logFormat, logLevel    string
	eventTTL               time.Duration
)

var (
	comp = new(kubefox.Component)
	spec = new(kubefox.ComponentSpec)
)

func main() {
	flag.StringVar(&comp.Name, "name", "", `Component name. (required)`)
	flag.StringVar(&comp.Commit, "commit", "", `Commit the Component was built from. (required)`)
	flag.StringVar(&httpAddr, "http-addr", "127.0.0.1:8080", `Address and port the HTTP server should bind to, set to "false" to disable.`)
	flag.StringVar(&httpsAddr, "https-addr", "127.0.0.1:8443", `Address and port the HTTPS server should bind to, set to "false" to disable.`)
	flag.StringVar(&brokerAddr, "broker-addr", "127.0.0.1:6060", "Address and port of the Broker gRPC server.")
	flag.StringVar(&healthAddr, "health-addr", "127.0.0.1:1111", `Address and port the HTTP health server should bind to, set to "false" to disable.`)
	flag.DurationVar(&eventTTL, "ttl", time.Minute, "Default time-to-live for an event.")
	flag.StringVar(&logFormat, "log-format", "console", "Log format. [options 'json', 'console']")
	flag.StringVar(&logLevel, "log-level", "debug", "Log level. [options 'debug', 'info', 'warn', 'error']")
	flag.Parse()

	utils.CheckRequiredFlag("name", comp.Name)
	utils.CheckRequiredFlag("commit", comp.Commit)

	if comp.Commit != build.Info.Commit {
		fmt.Fprintf(os.Stderr, "commit '%s' does not match build info commit '%s'", comp.Commit, build.Info.Commit)
		os.Exit(1)
	}

	comp.Id = kubefox.GenerateId()

	spec = &kubefox.ComponentSpec{
		ComponentTypeVar: kubefox.ComponentTypeVar{
			Type: kubefox.ComponentTypeGenesis,
		},
	}

	logkf.Global = logkf.
		BuildLoggerOrDie(logFormat, logLevel).
		WithComponent(comp)
	defer logkf.Global.Sync()

	srv := NewHTTPSrv()
	defer srv.Shutdown()

	if err := srv.Run(); err != nil {
		logkf.Global.Fatal(err)
	}
}
