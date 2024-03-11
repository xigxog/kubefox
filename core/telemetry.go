// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package core

import (
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math/rand"
	sync "sync"
	"time"

	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	v1 "go.opentelemetry.io/proto/otlp/trace/v1"
)

var (
	randSrcMutex sync.Mutex
	randSrc      *rand.Rand
)

func init() {
	var rngSeed int64
	_ = binary.Read(crand.Reader, binary.LittleEndian, &rngSeed)
	randSrc = rand.New(rand.NewSource(rngSeed))
}

type SpanContext struct {
	TraceId    [16]byte
	SpanId     [8]byte
	TraceState string
	Flags      byte
}

type SpanAttribute struct {
	Key   string
	Value any
}

type Span struct {
	otelSpan *v1.Span
}

func SpanFromContext(ctx *SpanContext, attrs ...SpanAttribute) *v1.Span {
	return &v1.Span{
		TraceId:    ctx.TraceId[:],
		SpanId:     ctx.SpanId[:],
		Attributes: convertAttributes(attrs),
		TraceState: ctx.TraceState,
		Flags:      uint32(ctx.Flags),
	}
}

func StartSpan(name string, traceParent *v1.Span, attrs ...SpanAttribute) *Span {
	otelSpan := &v1.Span{
		Name:              name,
		TraceId:           traceParent.TraceId[:],
		ParentSpanId:      traceParent.SpanId[:],
		SpanId:            make([]byte, 8),
		Attributes:        convertAttributes(attrs),
		TraceState:        traceParent.TraceState,
		Flags:             uint32(traceParent.Flags),
		StartTimeUnixNano: uint64(time.Now().UnixNano()),
	}

	// Generate spanId
	randSrcMutex.Lock()
	_, _ = randSrc.Read(otelSpan.SpanId[:])
	randSrcMutex.Unlock()

	return &Span{otelSpan: otelSpan}
}

func (s *Span) End() {
	s.otelSpan.EndTimeUnixNano = uint64(time.Now().UnixNano())
}

func convertAttributes(attrs []SpanAttribute) []*commonv1.KeyValue {
	converted := make([]*commonv1.KeyValue, len(attrs))

	for i, attr := range attrs {
		anyVal := &commonv1.AnyValue{}
		converted[i] = &commonv1.KeyValue{Key: attr.Key, Value: anyVal}

		switch t := attr.Value.(type) {
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

		default:
			anyVal.Value = &commonv1.AnyValue_StringValue{StringValue: fmt.Sprint(t)}
		}
	}

	return converted
}
