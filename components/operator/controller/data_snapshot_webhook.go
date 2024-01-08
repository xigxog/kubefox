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

type DataSnapshotWebhook struct {
	*Client
	*admission.Decoder
}

func (r *DataSnapshotWebhook) Handle(ctx context.Context, req admission.Request) admission.Response {
	if req.Operation != admv1.Create {
		return admission.Allowed("🦊")
	}

	reqSnap := &v1alpha1.DataSnapshot{}
	if err := r.DecodeRaw(req.Object, reqSnap); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	ve := &v1alpha1.VirtualEnvironment{}
	if err := r.Get(ctx, k8s.Key(req.Namespace, reqSnap.Spec.Source.Name), ve); err != nil {
		if k8s.IsNotFound(err) {
			return admission.Errored(http.StatusBadRequest, err)
		}
		return admission.Errored(http.StatusInternalServerError, err)
	}
	env := &v1alpha1.Environment{}
	if err := r.Get(ctx, k8s.Key(req.Namespace, ve.Spec.Environment), env); err != nil {
		if k8s.IsNotFound(err) {
			return admission.Errored(http.StatusBadRequest, err)
		}
		return admission.Errored(http.StatusInternalServerError, err)
	}

	dataSource := v1alpha1.DataSource{
		Kind:            api.DataSourceKindVirtualEnvironment,
		Name:            ve.Name,
		ResourceVersion: ve.ResourceVersion,
		DataChecksum:    ve.GetDataChecksum(),
	}

	if reqSnap.Data == nil {
		reqSnap.Data = &ve.Data
		reqSnap.Details = ve.Details
		reqSnap.Spec.Source = dataSource

	} else if !equality.Semantic.DeepEqual(&reqSnap.Spec.Source, &dataSource) ||
		!equality.Semantic.DeepEqual(reqSnap.Data, &ve.Data) ||
		!equality.Semantic.DeepEqual(&reqSnap.Details, &ve.Details) {

		return admission.Errored(http.StatusBadRequest, fmt.Errorf("data does not match source"))
	}

	k8s.UpdateLabel(reqSnap, api.LabelK8sSourceKind, string(dataSource.Kind))
	k8s.UpdateLabel(reqSnap, api.LabelK8sSourceName, dataSource.Name)
	k8s.UpdateLabel(reqSnap, api.LabelK8sSourceVersion, dataSource.ResourceVersion)

	current, err := json.Marshal(reqSnap)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, current)
}
