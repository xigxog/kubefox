package main

import (
	"flag"
	"path"
	"runtime"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/go-logr/logr"
	"github.com/xigxog/kubefox/components/broker/config"
	"github.com/xigxog/kubefox/components/broker/engine"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/logkf"
	"github.com/xigxog/kubefox/libs/core/utils"

	ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
	flag.StringVar(&config.Instance, "instance", "", "KubeFox instance Broker is part of. (required)")
	flag.StringVar(&config.Platform, "platform", "", "Platform instance Broker if part of. (required)")
	flag.StringVar(&config.Namespace, "namespace", "", "Namespace of Platform instance. (required)")
	flag.StringVar(&config.GRPCSrvAddr, "grpc-addr", "127.0.0.1:6060", "Address and port the gRPC server should bind to.")
	flag.StringVar(&config.HTTPSrvAddr, "http-addr", "127.0.0.1:8080", `Address and port the HTTP server should bind to, set to "false" to disable.`)
	flag.StringVar(&config.HTTPSSrvAddr, "https-addr", "127.0.0.1:8443", `Address and port the HTTPS server should bind to, set to "false" to disable.`)
	flag.StringVar(&config.HealthSrvAddr, "health-addr", "127.0.0.1:1111", `Address and port the HTTP health server should bind to, set to "false" to disable.`)
	flag.StringVar(&config.CertDir, "cert-dir", path.Join(kubefox.KubeFoxHome, "broker"), "Path of dir containing TLS certificate, private key, and root CA certificate.")
	flag.StringVar(&config.VaultAddr, "vault-addr", "127.0.0.1:8200", "Address and port of Vault server.")
	flag.StringVar(&config.NATSAddr, "nats-addr", "127.0.0.1:4222", "Address and port of NATS JetStream server.")
	flag.StringVar(&config.TelemetryAddr, "telemetry-addr", "127.0.0.1:4318", `Address and port of telemetry collector, set to "false" to disable.`)
	flag.DurationVar(&config.TelemetryInterval, "telemetry-interval", time.Minute, "Interval at which to report telemetry.")
	flag.IntVar(&config.NumWorkers, "num-workers", runtime.NumCPU(), "Number of worker threads to start, default is number of logical CPUs.")
	flag.DurationVar(&config.EventTTL, "ttl", time.Minute, "Default time-to-live for an event.")
	flag.StringVar(&config.LogFormat, "log-format", "console", `Log format; one of ["json", "console"].`)
	flag.StringVar(&config.LogLevel, "log-level", "debug", `Log level; one of ["debug", "info", "warn", "error"].`)
	flag.Parse()

	utils.CheckRequiredFlag("instance", config.Instance)
	utils.CheckRequiredFlag("platform", config.Platform)
	utils.CheckRequiredFlag("namespace", config.Namespace)

	logkf.Global = logkf.
		BuildLoggerOrDie(config.LogFormat, config.LogLevel).
		WithInstance(config.Instance).
		WithPlatform(config.Platform).
		WithService("broker")
	defer logkf.Global.Sync()

	ctrl.SetLogger(logr.Logger{})

	engine.New().Start()

}
