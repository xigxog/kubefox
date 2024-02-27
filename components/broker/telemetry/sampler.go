// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package telemetry

import (
	"github.com/xigxog/kubefox/logkf"
	"go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type Sampler struct {
	count int
}

func (s *Sampler) ShouldSample(p trace.SamplingParameters) trace.SamplingResult {
	s.count = s.count + 1
	spCtx := oteltrace.SpanContextFromContext(p.ParentContext)
	logkf.Global.Debugf("sampler called; count: %d, traceId: %s, spanId: %s, attrs: %s, traceState: %s",
		s.count, p.TraceID, spCtx.SpanID(), p.Attributes, spCtx.TraceState())

	return trace.SamplingResult{
		Decision:   trace.RecordAndSample,
		Tracestate: spCtx.TraceState(),
	}
}

func (s *Sampler) Description() string {
	return ""
}
