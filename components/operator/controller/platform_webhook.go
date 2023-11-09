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
	"fmt"
	"net/http"

	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	kubefox "github.com/xigxog/kubefox/core"
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
			fmt.Sprintf(`The Platform "%s" is not allowed: only one Platform is allowed per namespace`, platform.Name))
	}

	if !r.Mutating {
		if !*req.DryRun {
			ns := &v1.Namespace{}
			if err := r.Get(ctx, NN("", req.Namespace), ns); err != nil {
				return admission.Errored(http.StatusInternalServerError, err)
			}

			ns.Labels[kubefox.LabelK8sPlatform] = platform.Name
			if err := r.Update(ctx, ns); err != nil {
				return admission.Errored(http.StatusInternalServerError, err)
			}
		}

		return admission.Allowed("🦊")
	}

	// TODO set default resources/probes for platform components

	return admission.Allowed("")
}
