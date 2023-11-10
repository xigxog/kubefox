/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type PlatformWebhook struct {
	*Client
	*admission.Decoder

	Mutating bool
}

func (r *PlatformWebhook) Handle(ctx context.Context, req admission.Request) admission.Response {
	platform := &v1alpha1.Platform{}
	if err := r.Decode(req, platform); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	l := &v1alpha1.PlatformList{}
	if err := r.List(ctx, l, client.InNamespace(platform.Namespace)); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	allowed := false
	switch len(l.Items) {
	case 0:
		allowed = true
	case 1:
		if l.Items[0].Name == platform.Name {
			allowed = true
		}
	}
	if !allowed {
		return admission.Denied(
			fmt.Sprintf(`The Platform "%s" is not allowed: More than one Platform found in Namespace "%s"`,
				platform.Name, platform.Namespace))
	}

	if !r.Mutating {
		if !*req.DryRun {
			ns := &v1.Namespace{}
			if err := r.Get(ctx, NN("", req.Namespace), ns); err != nil {
				return admission.Errored(http.StatusInternalServerError, err)
			}

			ns.Labels[api.LabelK8sPlatform] = platform.Name
			if err := r.Update(ctx, ns); err != nil {
				return admission.Errored(http.StatusInternalServerError, err)
			}
		}

		return admission.Allowed("🦊")
	}

	// Setup defaults.
	s := &platform.Spec
	if s.Logger.Format == "" {
		s.Logger.Format = api.DefaultLogFormat
	}
	if s.Logger.Level == "" {
		s.Logger.Level = api.DefaultLogLevel
	}
	if s.Events.TimeoutSeconds == 0 {
		s.Events.TimeoutSeconds = api.DefaultTimeoutSeconds
	}
	if s.Events.MaxSize.IsZero() {
		s.Events.MaxSize.Set(api.DefaultMaxEventSizeBytes)
	}
	svc := &s.HTTPSrv.Service
	if svc.Type == "" {
		svc.Type = "ClusterIP"
	}
	if svc.Ports.HTTP == 0 {
		svc.Ports.HTTP = 80
	}
	if svc.Ports.HTTPS == 0 {
		svc.Ports.HTTPS = 443
	}

	SetDefaults(&s.NATS.ContainerSpec, &NATSDefaults)
	SetDefaults(&s.Broker.ContainerSpec, &BrokerDefaults)
	SetDefaults(&s.HTTPSrv.ContainerSpec, &HTTPSrvDefaults)

	current, err := json.Marshal(platform)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, current)
}
