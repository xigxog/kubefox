package telemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/xigxog/kubefox/libs/core/kubefox"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var (
	otelTracer = otel.Tracer("")
)

type Span interface {
	End(kubefox.DataEvent)
}

type span struct {
	cancel   context.CancelFunc
	otelSpan trace.Span
	req      kubefox.DataEvent
}

func NewSpan(ctx context.Context, timeout time.Duration, req kubefox.DataEvent) (context.Context, Span) {
	ctx, cancel := context.WithTimeout(ctx, timeout)

	typ := kubefox.UnknownEventType
	if req.GetType() != "" {
		typ = req.GetType()
	}

	if req.GetSpan() != nil {
		trcId, _ := trace.TraceIDFromHex(req.GetSpan().TraceId)
		spnId, _ := trace.SpanIDFromHex(req.GetSpan().SpanId)
		trcFlags := trace.TraceFlags(0)
		if len(req.GetSpan().GetTraceFlags()) > 0 {
			trcFlags = trace.TraceFlags(req.GetSpan().GetTraceFlags()[0])
		}

		ctx = trace.ContextWithRemoteSpanContext(ctx, trace.NewSpanContext(
			trace.SpanContextConfig{
				TraceID:    trcId,
				SpanID:     spnId,
				TraceFlags: trcFlags,
			}))
	}

	ctx, otelSpan := otelTracer.Start(ctx,
		fmt.Sprintf("%s event", typ),
		trace.WithAttributes(traceAttrs(req)...),
	)
	req.UpdateSpan(otelSpan)

	return ctx, &span{
		cancel:   cancel,
		otelSpan: otelSpan,
		req:      req,
	}
}

func (sp *span) End(resp kubefox.DataEvent) {
	if resp != nil {
		sp.otelSpan.SetAttributes(traceAttrs(resp)...)
		resp.UpdateSpan(sp.otelSpan)

		if err := resp.GetError(); err != nil {
			sp.otelSpan.RecordError(err)
			sp.otelSpan.SetStatus(codes.Error, err.Error())
		}
	}

	sp.otelSpan.End()
	sp.cancel()
}

func traceAttrs(req kubefox.DataEvent) []attribute.KeyValue {
	attrs := []attribute.KeyValue{}

	if req != nil && req.GetType() != "" {
		attrs = append(attrs, attribute.Key("kubefox.event.type").String(req.GetType()))
	}
	if req != nil && req.GetId() != "" {
		attribute.Key("kubefox.event.id").String(req.GetId())
	}
	if req != nil && req.GetParentId() != "" {
		attribute.Key("kubefox.event.parent-id").String(req.GetParentId())
	}
	if req != nil && req.GetSource() != nil {
		attribute.Key("kubefox.event.source").String(req.GetSource().GetURI())
	}
	if req != nil && req.GetTarget() != nil {
		attribute.Key("kubefox.event.target").String(req.GetTarget().GetURI())
	}

	return attrs
}
