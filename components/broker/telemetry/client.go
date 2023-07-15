package telemetry

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/xigxog/kubefox/components/broker/config"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/logger"
	"github.com/xigxog/kubefox/libs/core/platform"
	"github.com/xigxog/kubefox/libs/core/utils"
	"go.opentelemetry.io/contrib/instrumentation/host"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	ktyps "k8s.io/apimachinery/pkg/types"
)

const (
	TLSCertFile = "/kubefox/telemetry/tls/tls.crt"
	TLSKeyFile  = "/kubefox/telemetry/tls/tls.key"
	CACertFile  = "/kubefox/telemetry/tls/ca.crt"
)

type Client struct {
	meterProvider *metric.MeterProvider
	traceProvider *trace.TracerProvider

	cfg *config.Config
	log *logger.Log
}

func NewClient(cfg *config.Config, log *logger.Log) *Client {
	otel.SetErrorHandler(OTELErrorHandler{Log: log})

	return &Client{
		cfg: cfg,
		log: log,
	}
}

func (cl *Client) Start(ctx context.Context) {
	// tlsCfg, err := cl.tls()
	// if err != nil {
	// 	cl.log.Error(err)
	// 	os.Exit(kubefox.TelemetryErrorCode)
	// }

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(cl.cfg.Comp.GetName()),
		attribute.String("kubefox.component.id", cl.cfg.Comp.GetId()),
		attribute.String("kubefox.component.git-hash", cl.cfg.Comp.GetGitHash()),
		attribute.String("kubefox.component.name", cl.cfg.Comp.GetName()),
	)

	metricExp, err := otlpmetrichttp.New(ctx,
		// otlpmetrichttp.WithTLSClientConfig(tlsCfg),
		otlpmetrichttp.WithInsecure(),
		otlpmetrichttp.WithEndpoint(cl.cfg.TelemetryAgentAddr))
	if err != nil {
		cl.log.Error(err)
		os.Exit(kubefox.TelemetryErrorCode)
	}

	intSecs := time.Duration(cl.cfg.MetricsInterval) * time.Second
	cl.meterProvider = metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(metricExp, metric.WithInterval(intSecs))))
	global.SetMeterProvider(cl.meterProvider)

	if err := host.Start(); err != nil {
		cl.log.Error(err)
		os.Exit(kubefox.TelemetryErrorCode)
	}
	if err := runtime.Start(runtime.WithMinimumReadMemStatsInterval(intSecs)); err != nil {
		cl.log.Error(err)
		os.Exit(kubefox.TelemetryErrorCode)
	}

	// TODO do not exit if cannot connect, just show warnings it should keep
	// trying to connect, hopefully just enabling retry on clients will do looks
	// like the exporters do the right thing and throw away things if queue
	// fills up

	trClient := otlptracehttp.NewClient(
		// otlptracehttp.WithTLSClientConfig(tlsCfg),
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithEndpoint(cl.cfg.TelemetryAgentAddr),
	)
	trExp, err := otlptrace.New(ctx, trClient)
	if err != nil {
		cl.log.Error(err)
		os.Exit(kubefox.TelemetryErrorCode)
	}

	cl.traceProvider = trace.NewTracerProvider(
		// TODO sample setup? just rely on outside request to determine if to sample?
		// trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(res),
		trace.WithBatcher(trExp),
	)
	otel.SetTracerProvider(cl.traceProvider)

	cl.log.Infof("telemetry client started; addr: %s", cl.cfg.TelemetryAgentAddr)
}

func (cl *Client) Shutdown(ctx context.Context) {
	cl.log.Info("telemetry client shutting down")

	if cl.meterProvider != nil {
		if err := cl.meterProvider.Shutdown(ctx); err != nil {
			cl.log.Warn(err)
		}
	}

	if cl.traceProvider != nil {
		if err := cl.traceProvider.Shutdown(ctx); err != nil {
			cl.log.Warn(err)
		}
	}
}

func (cl *Client) tls() (*tls.Config, error) {
	var pool *x509.CertPool

	if pem, err := os.ReadFile(CACertFile); err == nil {
		cl.log.Debugf("reading tls certs from file")
		pool = x509.NewCertPool()
		if ok := pool.AppendCertsFromPEM(pem); !ok {
			return nil, fmt.Errorf("failed to parse root certificate from %s", CACertFile)
		}

	} else {
		cl.log.Debugf("reading tls certs from kubernetes secret")
		pool, err = utils.GetCAFromSecret(ktyps.NamespacedName{
			Namespace: cl.cfg.Namespace,
			Name:      fmt.Sprintf("%s-%s", cl.cfg.Platform, platform.RootCASecret),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to read cert from kubernetes secret: %v", err)
		}
	}

	return &tls.Config{RootCAs: pool}, nil
}
