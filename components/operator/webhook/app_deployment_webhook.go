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
	"net/http"

	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/k8s"
	admv1 "k8s.io/api/admission/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type AppDeploymentWebhook struct {
	*k8s.Client
	*admission.Decoder
}

func (r *AppDeploymentWebhook) Handle(ctx context.Context, req admission.Request) admission.Response {
	if req.Operation != admv1.Create {
		return admission.Allowed("🦊")
	}

	appDep := &v1alpha1.AppDeployment{}
	if err := r.DecodeRaw(req.Object, appDep); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if appDep.Spec.Version == "" {
		return admission.Allowed("🦊")
	}
	if appDep.Spec.Adapters != nil {
		return admission.Denied("create operation not allowed: .spec.adapters is a read-only field")
	}

	appDep.Spec.Adapters = &v1alpha1.Adapters{
		HTTP: map[string]v1alpha1.HTTPAdapterSpec{},
	}

	for _, comp := range appDep.Spec.Components {
		for depName, dep := range comp.Dependencies {
			switch dep.Type {
			case api.ComponentTypeHTTPAdapter:
				a := &v1alpha1.HTTPAdapter{}
				if err := r.Get(ctx, k8s.Key(req.Namespace, depName), a); err != nil {
					return admission.Errored(http.StatusBadRequest, err)
				}
				appDep.Spec.Adapters.HTTP[depName] = a.Spec
			}
		}
	}

	current, err := json.Marshal(appDep)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, current)
}
