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
	"github.com/xigxog/kubefox/k8s"
	admv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type SnapshotWebhook struct {
	*Client
	*admission.Decoder
}

func (r *SnapshotWebhook) Handle(ctx context.Context, req admission.Request) admission.Response {
	if req.Operation != admv1.Create {
		return admission.Allowed("🦊")
	}

	reqSnap := &v1alpha1.VirtualEnvSnapshot{}
	if err := r.DecodeRaw(req.Object, reqSnap); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	genSnap, err := r.SnapshotVirtualEnv(ctx, req.Namespace, reqSnap.Spec.Source.Name)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	if reqSnap.Data == nil {
		reqSnap.Data = genSnap.Data
		reqSnap.Details = genSnap.Details
		reqSnap.Spec.Source = genSnap.Spec.Source

	} else if !equality.Semantic.DeepEqual(&reqSnap.Spec, &genSnap.Spec) ||
		!equality.Semantic.DeepEqual(reqSnap.Data, genSnap.Data) ||
		!equality.Semantic.DeepEqual(&reqSnap.Details, &genSnap.Details) {

		return admission.Errored(http.StatusBadRequest, fmt.Errorf("VirtualEnvSnapshot does not match source VirtualEnv"))
	}

	k8s.UpdateLabel(reqSnap, api.LabelK8sVirtualEnv, reqSnap.Spec.Source.Name)
	k8s.UpdateLabel(reqSnap, api.LabelK8sSourceResourceVersion, reqSnap.Spec.Source.ResourceVersion)

	current, err := json.Marshal(reqSnap)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, current)
}
