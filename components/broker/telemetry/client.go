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
	"sync"
	"time"

	"github.com/xigxog/kubefox/components/broker/config"
	"github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/logkf"
	"github.com/xigxog/kubefox/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	resv1 "go.opentelemetry.io/proto/otlp/resource/v1"
	tracev1 "go.opentelemetry.io/proto/otlp/trace/v1"
)

var (
	Tracer = otel.Tracer("")
)

type Client struct {
	otelClient    otlptrace.Client
	meterProvider *metric.MeterProvider
	traceProvider *trace.TracerProvider

	spans []*tracev1.ResourceSpans

	tick  *time.Ticker
	mutex sync.Mutex
	log   *logkf.Logger
}

type OTELErrorHandler struct {
	Log *logkf.Logger
}

func (h OTELErrorHandler) Handle(err error) {
	h.Log.Warn(err)
}

func NewClient() *Client {
	//REMOVE
	config.TelemetryAddr = "kubefox-jaeger-collector.kubefox-system:4318"

	otel.SetErrorHandler(OTELErrorHandler{Log: logkf.Global})
	otel.SetTextMapPropagator(propagation.TraceContext{})

	c := &Client{
		otelClient: otlptracehttp.NewClient(
			// otlptracehttp.WithTLSClientConfig(tlsCfg),
			otlptracehttp.WithInsecure(),
			otlptracehttp.WithEndpoint(config.TelemetryAddr),
		),
		tick: time.NewTicker(time.Minute / 2),
		log:  logkf.Global,
	}
	go c.uploadTraces()

	return c
}

// func (c *Client) Start(ctx context.Context, comp *core.Component) error {
// 	c.log.Debug("telemetry client starting")

// 	otel.SetErrorHandler(OTELErrorHandler{Log: c.log})
// 	otel.SetTextMapPropagator(propagation.TraceContext{})

// tlsCfg, err := cl.tls()
// if err != nil {
// 	cl.log.Error(err)
// 	os.Exit(core.TelemetryErrorCode)
// }

// res := resource.NewWithAttributes(
// 	semconv.SchemaURL,
// 	semconv.ServiceName(comp.Key()),
// 	attribute.String("kubefox."+logkf.KeyInstance, config.Instance),
// 	attribute.String("kubefox."+logkf.KeyPlatform, config.Platform),
// 	attribute.String("kubefox."+logkf.KeyPlatformComponent, api.PlatformComponentBroker),
// )

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

// c.otelClient = otlptracehttp.NewClient(
// 	// otlptracehttp.WithTLSClientConfig(tlsCfg),
// 	otlptracehttp.WithInsecure(),
// 	otlptracehttp.WithEndpoint(config.TelemetryAddr),
// )

// exporter, err := otlptrace.New(ctx, c.otelClient)
// trExp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
// if err != nil {
// 	return c.log.ErrorN("%v", err)
// }

// bsp := trace.NewBatchSpanProcessor(trExp)
// bsp.OnEnd(nil) // this adds span to queue

// c.traceProvider = trace.NewTracerProvider(
// 	// TODO sample setup? just rely on outside request to determine if to sample?
// 	// trace.WithSampler(&Sampler{}),
// 	trace.WithSampler(trace.AlwaysSample()),
// 	trace.WithResource(res),
// 	trace.WithBatcher(exporter),
// )
// otel.SetTracerProvider(c.traceProvider)

// 	c.log.Info("telemetry client started")
// 	return nil
// }

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

func (cl *Client) AddResourceSpans(spans []*tracev1.ResourceSpans) {
	// TODO enhance attributes like instance, platform, etc.
	cl.mutex.Lock()
	cl.spans = append(cl.spans, spans...)
	cl.mutex.Unlock()
}

func (cl *Client) AddSpans(comp *core.Component, spans ...*telemetry.Span) {
	resSpans := &tracev1.ResourceSpans{
		Resource: &resv1.Resource{
			Attributes: []*commonv1.KeyValue{
				telemetry.Attr(telemetry.AttrKeySvcName, comp.Key()).KeyValue,
				telemetry.Attr(telemetry.AttrKeyComponentType, comp.Type).KeyValue,
				telemetry.Attr(telemetry.AttrKeyComponentApp, comp.App).KeyValue,
				telemetry.Attr(telemetry.AttrKeyComponentName, comp.Name).KeyValue,
				telemetry.Attr(telemetry.AttrKeyComponentHash, comp.Hash).KeyValue,
				telemetry.Attr(telemetry.AttrKeyComponentId, comp.Id).KeyValue,
				telemetry.Attr(telemetry.AttrKeyInstance, config.Instance).KeyValue,
				telemetry.Attr(telemetry.AttrKeyPlatform, config.Platform).KeyValue,
			},
		},
		ScopeSpans: []*tracev1.ScopeSpans{
			{
				Spans:     make([]*tracev1.Span, len(spans)),
				SchemaUrl: semconv.SchemaURL,
			},
		},
		SchemaUrl: semconv.SchemaURL,
	}
	for i := range spans {
		resSpans.ScopeSpans[0].Spans[i] = spans[i].Span
	}

	cl.mutex.Lock()
	cl.spans = append(cl.spans, resSpans)
	cl.mutex.Unlock()
}

func (cl *Client) uploadTraces() {
	for range cl.tick.C {
		if len(cl.spans) == 0 {
			continue
		}

		cl.mutex.Lock()
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute/4)

		cl.log.Debugf("uploading %d resource spans", len(cl.spans))

		if err := cl.otelClient.UploadTraces(ctx, cl.spans); err != nil {
			cl.log.Errorf("error uploading traces: %v", err)
		}

		cl.spans = nil

		cancel()
		cl.mutex.Unlock()
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
