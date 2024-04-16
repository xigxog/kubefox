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
	"reflect"
	sync "sync"
	"time"

	"github.com/xigxog/kubefox/core"
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
	*tracev1.Span

	Recorded []*tracev1.Span
}

type Attribute struct {
	*commonv1.KeyValue
}

func SetComponent(comp *core.Component, attrs ...Attribute) {
	resource = &resv1.Resource{
		Attributes: []*commonv1.KeyValue{
			Attr(AttrKeySvcName, comp.Key()).KeyValue,
			Attr(AttrKeyComponentType, comp.Type).KeyValue,
			Attr(AttrKeyComponentApp, comp.App).KeyValue,
			Attr(AttrKeyComponentName, comp.Name).KeyValue,
			Attr(AttrKeyComponentHash, comp.Hash).KeyValue,
			Attr(AttrKeyComponentId, comp.Id).KeyValue,
		},
	}
	for _, a := range attrs {
		resource.Attributes = append(resource.Attributes, a.KeyValue)
	}
}

func Resource() *resv1.Resource {
	return resource
}

func Attr(key string, val any) Attribute {
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

	return Attribute{&commonv1.KeyValue{Key: string(key), Value: anyVal}}
}

func StartSpan(name string, parent *core.SpanContext, attrs ...Attribute) *Span {
	if parent == nil {
		parent = &core.SpanContext{
			TraceId: make([]byte, 16),
		}
		// Generate traceId
		randSrcMutex.Lock()
		_, _ = randSrc.Read(parent.TraceId)
		randSrcMutex.Unlock()
	}

	var otelAttrs []*commonv1.KeyValue
	if len(attrs) > 0 {
		otelAttrs = make([]*commonv1.KeyValue, len(attrs))
		for i := range attrs {
			otelAttrs[i] = attrs[i].KeyValue
		}
	}

	protoSpan := &tracev1.Span{
		Name:              name,
		TraceId:           parent.TraceId,
		ParentSpanId:      parent.SpanId,
		SpanId:            make([]byte, 8),
		Attributes:        otelAttrs,
		TraceState:        parent.TraceState,
		Flags:             parent.Flags,
		StartTimeUnixNano: now(),
	}

	// Generate spanId
	randSrcMutex.Lock()
	_, _ = randSrc.Read(protoSpan.SpanId)
	randSrcMutex.Unlock()

	return &Span{
		Span:     protoSpan,
		Recorded: []*tracev1.Span{protoSpan},
	}
}

func (s *Span) StartChildSpan(name string, attrs ...Attribute) *Span {
	return StartSpan(name, s.SpanContext(), attrs...)
}

func (s *Span) SpanContext() *core.SpanContext {
	return &core.SpanContext{
		TraceId:    s.TraceId,
		SpanId:     s.SpanId,
		TraceState: s.TraceState,
		Flags:      s.Flags,
	}
}

func (s *Span) SetAttributes(attrs ...Attribute) {
	for _, a := range attrs {
		found := false
		for i, c := range s.Attributes {
			if c.Key == a.Key {
				s.Attributes[i] = a.KeyValue
				found = true
				break
			}
		}
		if !found {
			s.Attributes = append(s.Attributes, a.KeyValue)
		}
	}
}

func (s *Span) RecordErr(err error) {
	if err == nil {
		return
	}

	s.Events = append(s.Events, &tracev1.Span_Event{
		TimeUnixNano: now(),
		Name:         EventNameException,
		Attributes: []*commonv1.KeyValue{
			Attr(AttrKeyExceptionType, typeStr(err)).KeyValue,
			Attr(AttrKeyExceptionMsg, err.Error()).KeyValue,
		},
	})
}

func (s *Span) End(errs ...error) {
	if len(errs) > 0 {
		for _, e := range errs {
			s.RecordErr(e)
		}

		s.Status = &tracev1.Status{
			Message: errs[0].Error(),
			Code:    tracev1.Status_STATUS_CODE_ERROR,
		}
	}

	s.EndTimeUnixNano = now()
}

func now() uint64 {
	return uint64(time.Now().UnixNano())
}

func typeStr(i interface{}) string {
	t := reflect.TypeOf(i)
	if t.PkgPath() == "" && t.Name() == "" {
		// Likely a builtin type.
		return t.String()
	}
	return fmt.Sprintf("%s.%s", t.PkgPath(), t.Name())
}

func init() {
	var rngSeed int64
	_ = binary.Read(crand.Reader, binary.LittleEndian, &rngSeed)
	randSrc = rand.New(rand.NewSource(rngSeed))
}
