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
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	logsv1 "go.opentelemetry.io/proto/otlp/logs/v1"
	resv1 "go.opentelemetry.io/proto/otlp/resource/v1"
	tracev1 "go.opentelemetry.io/proto/otlp/trace/v1"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	maxQueueSize = 1_000
)

var (
	Tracer = otel.Tracer("")
)

type Client struct {
	conn        *gogrpc.ClientConn
	traceClient otlptrace.Client
	logsClient  *logsClient

	spans []*tracev1.ResourceSpans
	logs  []*logsv1.ResourceLogs

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
	otel.SetErrorHandler(OTELErrorHandler{Log: logkf.Global})
	otel.SetTextMapPropagator(propagation.TraceContext{})

	c := &Client{
		tick: time.NewTicker(time.Second * 15),
		log:  logkf.Global,
	}

	return c
}

func (c *Client) Start(ctx context.Context) error {
	c.log.Debug("telemetry client starting")

	conn, err := gogrpc.NewClient(config.TelemetryAddr,
		gogrpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}
	c.conn = conn

	c.traceClient = otlptracegrpc.NewClient(otlptracegrpc.WithGRPCConn(conn))
	if err := c.traceClient.Start(ctx); err != nil {
		return err
	}

	c.logsClient = NewLogsClient(c.log)
	c.logsClient.Start(ctx, conn)

	go c.publishTelemetry()
	c.log.Info("telemetry client started")

	return nil
}

func (cl *Client) Shutdown(timeout time.Duration) {
	// log from context
	cl.log.Info("telemetry client shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if cl.traceClient != nil {
		if err := cl.traceClient.Stop(ctx); err != nil {
			cl.log.Warn(err)
		}
	}
	if cl.conn != nil {
		if err := cl.conn.Close(); err != nil {
			cl.log.Warn(err)
		}
	}
}

func (cl *Client) AddTelemetry(comp *core.Component, tel *core.Telemetry) {
	cl.AddProtoSpans(comp, tel.Spans)
	cl.AddProtoLogs(comp, tel.LogRecords)
}

func (cl *Client) AddSpans(comp *core.Component, spans ...*telemetry.Span) {
	if len(spans) == 0 {
		return
	}

	protoSpans := make([]*tracev1.Span, len(spans))
	for i := range spans {
		protoSpans[i] = spans[i].Span
		cl.AddSpans(comp, spans[i].ChildSpans...)
	}

	cl.AddProtoSpans(comp, protoSpans)
}

func (cl *Client) AddProtoSpans(comp *core.Component, spans []*tracev1.Span) {
	cl.mutex.Lock()
	defer cl.mutex.Unlock()

	switch {
	case len(spans) == 0:
		return
	case len(cl.spans) > maxQueueSize:
		cl.log.Warnf("maximum number of queued spans exceeded, discarding %d incoming", len(spans))
		return
	}

	i := 0
	for _, s := range spans {
		if s.TraceState == "kf=1" {
			spans[i] = s
			i++
		}
	}
	if diff := len(spans) - i; diff > 0 {
		cl.log.Debugf("%d spans do not have record state set, discarding", diff)
	}
	// Truncate spans that do not need to be recorded.
	spans = spans[:i]
	if len(spans) == 0 {
		return
	}

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
				Spans:     spans,
				SchemaUrl: semconv.SchemaURL,
			},
		},
		SchemaUrl: semconv.SchemaURL,
	}

	cl.spans = append(cl.spans, resSpans)
}

func (cl *Client) AddProtoLogs(comp *core.Component, logRecords []*logsv1.LogRecord) {
	cl.mutex.Lock()
	defer cl.mutex.Unlock()

	switch {
	case len(logRecords) == 0:
		return
	case len(cl.logs) > maxQueueSize:
		cl.log.Warnf("maximum number of queued log records exceeded, discarding %d incoming", len(logRecords))
		return
	}

	resSpans := &logsv1.ResourceLogs{
		Resource: buildResource(comp),
		ScopeLogs: []*logsv1.ScopeLogs{
			{
				LogRecords: logRecords,
				SchemaUrl:  semconv.SchemaURL,
			},
		},
		SchemaUrl: semconv.SchemaURL,
	}

	cl.logs = append(cl.logs, resSpans)
}

// TODO have broker/grpc server create resource and pass that instead of comp so
// it doesn't need to be recreated over and over.
func buildResource(comp *core.Component) *resv1.Resource {
	return &resv1.Resource{
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
	}
}

func (cl *Client) publishTelemetry() {
	for range cl.tick.C {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			cl.publishSpans(ctx)
			wg.Done()
		}()
		go func() {
			cl.publishLogs(ctx)
			wg.Done()
		}()
		wg.Wait()

		cancel()
	}
}

func (cl *Client) publishSpans(ctx context.Context) {
	if cl.traceClient == nil || len(cl.spans) == 0 {
		return
	}

	cl.mutex.Lock()
	defer cl.mutex.Unlock()

	cl.log.Debugf("uploading %d resource spans", len(cl.spans))
	if err := cl.traceClient.UploadTraces(ctx, cl.spans); err != nil {
		cl.log.Errorf("error uploading traces: %v", err)
	}
	cl.spans = nil
}

func (cl *Client) publishLogs(ctx context.Context) {
	if cl.logsClient == nil || len(cl.logs) == 0 {
		return
	}

	cl.mutex.Lock()
	defer cl.mutex.Unlock()

	cl.log.Debugf("uploading %d resource logs", len(cl.logs))
	if err := cl.logsClient.UploadLogs(ctx, cl.logs); err != nil {
		cl.log.Errorf("error uploading logs: %v", err)
	}
	cl.logs = nil
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
