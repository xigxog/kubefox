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
	"time"

	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/components/broker/config"
	"github.com/xigxog/kubefox/logkf"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type Client struct {
	meterProvider *metric.MeterProvider
	traceProvider *trace.TracerProvider

	log *logkf.Logger
}

type OTELErrorHandler struct {
	Log *logkf.Logger
}

func (h OTELErrorHandler) Handle(err error) {
	h.Log.Warn(err)
}

func NewClient() *Client {
	return &Client{
		log: logkf.Global,
	}
}

func (c *Client) Start(ctx context.Context) error {
	c.log.Debug("telemetry client starting")

	otel.SetErrorHandler(OTELErrorHandler{Log: c.log})

	// tlsCfg, err := cl.tls()
	// if err != nil {
	// 	cl.log.Error(err)
	// 	os.Exit(core.TelemetryErrorCode)
	// }

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		attribute.String("kubefox."+logkf.KeyInstance, config.Instance),
		attribute.String("kubefox."+logkf.KeyPlatform, config.Platform),
		attribute.String("kubefox."+logkf.KeyPlatformComponent, api.PlatformComponentBroker),
	)

	// metricExp, err := otlpmetrichttp.New(ctx,
	// 	// otlpmetrichttp.WithTLSClientConfig(tlsCfg),
	// 	otlpmetrichttp.WithInsecure(),
	// 	otlpmetrichttp.WithEndpoint(config.TelemetryAddr))
	// if err != nil {
	// 	return c.log.ErrorN("%v", err)
	// }

	// interval := time.Duration(config.TelemetryInterval) * time.Second
	// c.meterProvider = metric.NewMeterProvider(
	// 	metric.WithResource(res),
	// 	metric.WithReader(metric.NewPeriodicReader(metricExp, metric.WithInterval(interval))))
	// otel.SetMeterProvider(c.meterProvider)

	// if err := host.Start(); err != nil {
	// 	return c.log.ErrorN("%v", err)
	// }
	// if err := runtime.Start(runtime.WithMinimumReadMemStatsInterval(interval)); err != nil {
	// 	return c.log.ErrorN("%v", err)
	// }

	// trClient := otlptracehttp.NewClient(
	// 	// otlptracehttp.WithTLSClientConfig(tlsCfg),
	// 	otlptracehttp.WithInsecure(),
	// 	otlptracehttp.WithEndpoint(config.TelemetryAddr),
	// )
	// trExp, err := otlptrace.New(ctx, trClient)

	trExp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return c.log.ErrorN("%v", err)
	}

	c.traceProvider = trace.NewTracerProvider(
		// TODO sample setup? just rely on outside request to determine if to sample?
		trace.WithSampler(&Sampler{}),
		trace.WithResource(res),
		trace.WithBatcher(trExp),
	)
	otel.SetTracerProvider(c.traceProvider)

	// TODO ???
	//  otel.SetTextMapPropagator()

	c.log.Info("telemetry client started")
	return nil
}

func (cl *Client) Shutdown(timeout time.Duration) {
	// log from context
	cl.log.Info("telemetry client shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

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

// func (cl *Client) tls() (*tls.Config, error) {
// 	var pool *x509.CertPool

// 	if pem, err := os.ReadFile(CACertFile); err == nil {
// 		// cl.log.Debugf("reading tls certs from file")
// 		pool = x509.NewCertPool()
// 		if ok := pool.AppendCertsFromPEM(pem); !ok {
// 			return nil, fmt.Errorf("failed to parse root certificate from %s", CACertFile)
// 		}

// 	} else {
// 		// cl.log.Debugf("reading tls certs from kubernetes secret")
// 		pool, err = utils.GetCAFromSecret(ktyps.NamespacedName{
// 			// Namespace: cl.cfg.Namespace,
// 			// Name:      fmt.Sprintf("%s-%s", cl.cfg.Platform, kfp.RootCASecret),
// 		})
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to read cert from kubernetes secret: %v", err)
// 		}
// 	}

// 	return &tls.Config{RootCAs: pool}, nil
// }
