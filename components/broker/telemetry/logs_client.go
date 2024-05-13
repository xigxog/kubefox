// Copyright 2024 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package telemetry

import (
	"context"

	"github.com/xigxog/kubefox/logkf"
	colv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"

	logsv1 "go.opentelemetry.io/proto/otlp/logs/v1"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type logsClient struct {
	colClient colv1.LogsServiceClient
	log       *logkf.Logger
}

func NewLogsClient(log *logkf.Logger) *logsClient {
	return &logsClient{log: log}
}

// Start establishes a gRPC connection to the collector.
func (c *logsClient) Start(ctx context.Context, conn *gogrpc.ClientConn) error {
	c.colClient = colv1.NewLogsServiceClient(conn)

	return nil
}

func (c *logsClient) UploadLogs(ctx context.Context, logs []*logsv1.ResourceLogs) error {
	resp, err := c.colClient.Export(ctx, &colv1.ExportLogsServiceRequest{
		ResourceLogs: logs,
	})
	if resp != nil && resp.PartialSuccess != nil {
		msg := resp.PartialSuccess.GetErrorMessage()
		n := resp.PartialSuccess.GetRejectedLogRecords()
		if n != 0 || msg != "" {
			c.log.Warnf("%d log records rejected: %s", n, msg)
		}
	}
	// nil is converted to OK.
	if status.Code(err) == codes.OK {
		// Success.
		return nil
	}

	return err
}
