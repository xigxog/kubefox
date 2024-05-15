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
	"github.com/xigxog/kubefox/components/operator/vault"
	"github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/k8s"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type SecretsWebhook struct {
	admission.Decoder

	Instance string
}

func (r *SecretsWebhook) Handle(ctx context.Context, req admission.Request) admission.Response {
	var obj client.Object
	switch req.Kind.String() {
	case "kubefox.xigxog.io/v1alpha1, Kind=Environment":
		obj = &v1alpha1.Environment{}
	case "kubefox.xigxog.io/v1alpha1, Kind=ReleaseManifest":
		obj = &v1alpha1.ReleaseManifest{}
	case "kubefox.xigxog.io/v1alpha1, Kind=VirtualEnvironment":
		obj = &v1alpha1.VirtualEnvironment{}
	default:
		return admission.Allowed("🦊")
	}

	if err := r.DecodeRaw(req.Object, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	var (
		data    *api.Data
		dataKey api.DataKey
	)
	if dataProvider, ok := obj.(api.DataProvider); ok {
		data = dataProvider.GetData()
		dataKey = dataProvider.GetDataKey()
	} else {
		return admission.Errored(http.StatusBadRequest, core.ErrInvalid())
	}

	vaultCli, err := vault.GetClient(ctx)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	vaultData := &api.Data{
		Secrets: map[string]*api.Val{},
	}
	if err := vaultCli.GetData(ctx, dataKey, vaultData); k8s.IgnoreNotFound(err) != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	// Remove any keys currently in Vault but not in Data.
	for k := range vaultData.Secrets {
		if _, found := data.Secrets[k]; !found {
			delete(vaultData.Secrets, k)
		}
	}
	// Copy plain text secrets from Data into Vault and mask value in Data.
	for k, v := range data.Secrets {
		if v.String() != api.SecretMask {
			vaultData.Secrets[k] = v
			data.Secrets[k] = api.ValString(api.SecretMask)
		}
	}

	// Do not write to Vault if obj was deleted.
	if obj.GetDeletionTimestamp().IsZero() {
		if err := vaultCli.PutData(ctx, dataKey, vaultData); err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}
	}

	// kubectl places secrets into the annotation
	// 'kubectl.kubernetes.io/last-applied-configuration'. If present override
	// it with updated obj and re-marshal.
	if last, found := obj.GetAnnotations()[api.AnnotationLastApplied]; found {
		lastObj := map[string]any{}
		json.Unmarshal([]byte(last), &lastObj)
		lastObj["data"] = data
		cleaned, _ := json.Marshal(lastObj)

		obj.GetAnnotations()[api.AnnotationLastApplied] = string(cleaned)
	}

	current, err := json.Marshal(obj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, current)
}
