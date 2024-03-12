// Copyright 2024 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package telemetry

import (
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math/rand"
	sync "sync"
	"time"

	"github.com/xigxog/kubefox/core"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	resv1 "go.opentelemetry.io/proto/otlp/resource/v1"
	tracev1 "go.opentelemetry.io/proto/otlp/trace/v1"
)

var (
	randSrcMutex sync.Mutex
	randSrc      *rand.Rand
	resource     *resv1.Resource
)

type Span struct {
	otelSpan *tracev1.Span
}

type Attribute struct {
	kv *commonv1.KeyValue
}

func SetComponent(comp *core.Component) {
	resource = &resv1.Resource{
		Attributes: []*commonv1.KeyValue{
			SpanAttribute(string(semconv.ServiceNameKey), comp.Key()).kv,
			SpanAttribute("kubefox.component.type", comp.Type).kv,
			SpanAttribute("kubefox.component.app", comp.App).kv,
			SpanAttribute("kubefox.component.name", comp.Name).kv,
			SpanAttribute("kubefox.component.hash", comp.Hash).kv,
			SpanAttribute("kubefox.component.id", comp.Id).kv,
		},
	}
}

func Resource() *resv1.Resource {
	return resource
}

func SpanContextFromOTEL(otelSpan trace.SpanContext) *core.SpanContext {
	tid, sid := otelSpan.TraceID(), otelSpan.SpanID()
	return &core.SpanContext{
		TraceId:    tid[:],
		SpanId:     sid[:],
		TraceState: otelSpan.TraceState().String(),
		Flags:      uint32(otelSpan.TraceFlags()),
	}
}

func StartSpan(name string, traceParent *core.SpanContext, attrs ...Attribute) *Span {
	if traceParent == nil {
		traceParent = &core.SpanContext{
			TraceId: make([]byte, 16),
		}
		// Generate traceId
		randSrcMutex.Lock()
		_, _ = randSrc.Read(traceParent.TraceId)
		randSrcMutex.Unlock()
	}

	otelAttrs := make([]*commonv1.KeyValue, len(attrs))
	for i := range attrs {
		otelAttrs[i] = attrs[i].kv
	}

	otelSpan := &tracev1.Span{
		Name:              name,
		TraceId:           traceParent.TraceId,
		ParentSpanId:      traceParent.SpanId,
		SpanId:            make([]byte, 8),
		Attributes:        otelAttrs,
		TraceState:        traceParent.TraceState,
		Flags:             traceParent.Flags,
		StartTimeUnixNano: uint64(time.Now().UnixNano()),
	}

	// Generate spanId
	randSrcMutex.Lock()
	_, _ = randSrc.Read(otelSpan.SpanId)
	randSrcMutex.Unlock()

	return &Span{
		otelSpan: otelSpan,
	}
}

func (s *Span) SpanContext() *core.SpanContext {
	return &core.SpanContext{
		TraceId:    s.otelSpan.TraceId,
		SpanId:     s.otelSpan.SpanId,
		TraceState: s.otelSpan.TraceState,
		Flags:      s.otelSpan.Flags,
	}
}

func (s *Span) OTELSpan() *tracev1.Span {
	return s.otelSpan
}

func (s *Span) End() {
	s.otelSpan.EndTimeUnixNano = uint64(time.Now().UnixNano())
}

func SpanAttribute(key string, val any) Attribute {
	anyVal := &commonv1.AnyValue{}
	switch t := val.(type) {
	case float32:
		anyVal.Value = &commonv1.AnyValue_DoubleValue{DoubleValue: float64(t)}
	case float64:
		anyVal.Value = &commonv1.AnyValue_DoubleValue{DoubleValue: t}

	case int:
		anyVal.Value = &commonv1.AnyValue_IntValue{IntValue: int64(t)}
	case int8:
		anyVal.Value = &commonv1.AnyValue_IntValue{IntValue: int64(t)}
	case int16:
		anyVal.Value = &commonv1.AnyValue_IntValue{IntValue: int64(t)}
	case int32:
		anyVal.Value = &commonv1.AnyValue_IntValue{IntValue: int64(t)}
	case int64:
		anyVal.Value = &commonv1.AnyValue_IntValue{IntValue: int64(t)}
	case uint:
		anyVal.Value = &commonv1.AnyValue_IntValue{IntValue: int64(t)}
	case uint8:
		anyVal.Value = &commonv1.AnyValue_IntValue{IntValue: int64(t)}
	case uint16:
		anyVal.Value = &commonv1.AnyValue_IntValue{IntValue: int64(t)}
	case uint32:
		anyVal.Value = &commonv1.AnyValue_IntValue{IntValue: int64(t)}
	case uint64:
		anyVal.Value = &commonv1.AnyValue_IntValue{IntValue: int64(t)}

	case bool:
		anyVal.Value = &commonv1.AnyValue_BoolValue{BoolValue: t}

	case []byte:
		anyVal.Value = &commonv1.AnyValue_BytesValue{BytesValue: t}

	default:
		anyVal.Value = &commonv1.AnyValue_StringValue{StringValue: fmt.Sprint(t)}
	}

	return Attribute{&commonv1.KeyValue{Key: key, Value: anyVal}}
}

func init() {
	var rngSeed int64
	_ = binary.Read(crand.Reader, binary.LittleEndian, &rngSeed)
	randSrc = rand.New(rand.NewSource(rngSeed))
}
