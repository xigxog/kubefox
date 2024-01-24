// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/components/operator/defaults"
	"github.com/xigxog/kubefox/k8s"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type PlatformWebhook struct {
	*k8s.Client
	*admission.Decoder
}

func (r *PlatformWebhook) Handle(ctx context.Context, req admission.Request) admission.Response {
	l := &v1alpha1.PlatformList{}
	if err := r.List(ctx, l, client.InNamespace(req.Namespace)); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	allowed := false
	switch len(l.Items) {
	case 0:
		allowed = true
	case 1:
		if l.Items[0].Name == req.Name {
			allowed = true
		}
	}
	if !allowed {
		return admission.Denied(
			fmt.Sprintf(`The Platform "%s" is not allowed: More than one Platform found in Namespace "%s"`,
				req.Name, req.Namespace))
	}

	platform := &v1alpha1.Platform{}
	if err := r.DecodeRaw(req.Object, platform); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if platform.Spec.Logger.Format == "" {
		platform.Spec.Logger.Format = api.DefaultLogFormat
	}
	if platform.Spec.Logger.Level == "" {
		platform.Spec.Logger.Level = api.DefaultLogLevel
	}
	if platform.Spec.Events.TimeoutSeconds == 0 {
		platform.Spec.Events.TimeoutSeconds = api.DefaultTimeoutSeconds
	}
	if platform.Spec.Events.MaxSize.IsZero() {
		platform.Spec.Events.MaxSize.Set(api.DefaultMaxEventSizeBytes)
	}

	svc := &platform.Spec.HTTPSrv.Service
	if svc.Type == "" {
		svc.Type = "ClusterIP"
	}
	if svc.Ports.HTTP == 0 {
		svc.Ports.HTTP = 80
	}
	if svc.Ports.HTTPS == 0 {
		svc.Ports.HTTPS = 443
	}

	defaults.Set(&platform.Spec.NATS.ContainerSpec, &defaults.NATS)
	defaults.Set(&platform.Spec.Broker.ContainerSpec, &defaults.Broker)
	defaults.Set(&platform.Spec.HTTPSrv.ContainerSpec, &defaults.HTTPSrv)

	current, err := json.Marshal(platform)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, current)
}
