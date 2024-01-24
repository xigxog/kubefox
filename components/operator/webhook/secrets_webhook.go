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
	"github.com/xigxog/kubefox/k8s"
	"github.com/xigxog/kubefox/logkf"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type SecretsWebhook struct {
	*admission.Decoder

	Instance string
}

func (r *SecretsWebhook) Handle(ctx context.Context, req admission.Request) admission.Response {
	vaultCli, err := vault.GetClient(ctx)
	if err != nil {
		logkf.Global.Error(err)
		return admission.Errored(http.StatusInternalServerError, err)
	}

	var dataProvider api.DataProvider
	switch req.Kind.String() {
	case "kubefox.xigxog.io/v1alpha1, Kind=Environment":
		dataProvider = &v1alpha1.Environment{}
	case "kubefox.xigxog.io/v1alpha1, Kind=ReleaseManifest":
		dataProvider = &v1alpha1.ReleaseManifest{}
	case "kubefox.xigxog.io/v1alpha1, Kind=VirtualEnvironment":
		dataProvider = &v1alpha1.VirtualEnvironment{}
	default:
		return admission.Allowed("🦊")
	}
	if err := r.DecodeRaw(req.Object, dataProvider.(client.Object)); err != nil {
		logkf.Global.Error(err)
		return admission.Errored(http.StatusBadRequest, err)
	}

	vaultData := &api.Data{
		Vars:    map[string]*api.Val{},
		Secrets: map[string]*api.Val{},
	}
	if err := vaultCli.GetData(ctx, dataProvider.GetDataKey(), vaultData); k8s.IgnoreNotFound(err) != nil {
		logkf.Global.Error(err)
		return admission.Errored(http.StatusInternalServerError, err)
	}

	data := dataProvider.GetData()
	vaultData.Vars = data.Vars

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
	if err := vaultCli.PutData(ctx, dataProvider.GetDataKey(), vaultData); err != nil {
		logkf.Global.Error(err)
		return admission.Errored(http.StatusInternalServerError, err)
	}

	current, err := json.Marshal(dataProvider)
	if err != nil {
		logkf.Global.Error(err)
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, current)
}
