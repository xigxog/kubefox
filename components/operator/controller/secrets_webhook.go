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

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type SecretsWebhook struct {
	*Client
	*admission.Decoder
}

func (r *SecretsWebhook) Handle(ctx context.Context, req admission.Request) admission.Response {
	return admission.Allowed("🦊")

	// reqSnap := &v1alpha1.DataSnapshot{}
	// if err := r.DecodeRaw(req.Object, reqSnap); err != nil {
	// 	return admission.Errored(http.StatusBadRequest, err)
	// }

	// genSnap, err := r.SnapshotVirtualEnv(ctx, req.Namespace, reqSnap.Spec.Source.Name)
	// if err != nil {
	// 	return admission.Errored(http.StatusInternalServerError, err)
	// }

	// if reqSnap.Data == nil {
	// 	reqSnap.Data = genSnap.Data
	// 	reqSnap.Details = genSnap.Details
	// 	reqSnap.Spec.Source = genSnap.Spec.Source

	// } else if !equality.Semantic.DeepEqual(&reqSnap.Spec, &genSnap.Spec) ||
	// 	!equality.Semantic.DeepEqual(reqSnap.Data, genSnap.Data) ||
	// 	!equality.Semantic.DeepEqual(&reqSnap.Details, &genSnap.Details) {

	// 	return admission.Errored(http.StatusBadRequest, fmt.Errorf("data does not match source"))
	// }

	// k8s.UpdateLabel(reqSnap, api.LabelK8sSourceKind, string(reqSnap.Spec.Source.Kind))
	// k8s.UpdateLabel(reqSnap, api.LabelK8sSourceName, reqSnap.Spec.Source.Name)
	// k8s.UpdateLabel(reqSnap, api.LabelK8sSourceVersion, reqSnap.Spec.Source.ResourceVersion)

	// current, err := json.Marshal(reqSnap)
	// if err != nil {
	// 	return admission.Errored(http.StatusInternalServerError, err)
	// }

	// return admission.PatchResponseFromRaw(req.Object.Raw, current)
}
