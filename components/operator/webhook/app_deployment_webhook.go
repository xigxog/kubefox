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
