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
	"github.com/xigxog/kubefox/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type IndexWebhook struct {
	admission.Decoder
}

func (r *IndexWebhook) Handle(ctx context.Context, req admission.Request) admission.Response {
	var obj client.Object
	switch req.Kind.String() {
	case "kubefox.xigxog.io/v1alpha1, Kind=AppDeployment":
		appDep := &v1alpha1.AppDeployment{}
		if err := r.DecodeRaw(req.Object, appDep); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		obj = appDep

		k8s.UpdateLabel(appDep, api.LabelK8sAppBranch, appDep.Spec.Branch)
		k8s.UpdateLabel(appDep, api.LabelK8sAppCommit, appDep.Spec.Commit)
		k8s.UpdateLabel(appDep, api.LabelK8sAppCommitShort, utils.ShortHash(appDep.Spec.Commit))
		k8s.UpdateLabel(appDep, api.LabelK8sAppName, appDep.Spec.AppName)
		k8s.UpdateLabel(appDep, api.LabelK8sAppTag, appDep.Spec.Tag)
		k8s.UpdateLabel(appDep, api.LabelK8sAppVersion, appDep.Spec.Version)

	case "kubefox.xigxog.io/v1alpha1, Kind=ReleaseManifest":
		manifest := &v1alpha1.ReleaseManifest{}
		if err := r.DecodeRaw(req.Object, manifest); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		obj = manifest

		k8s.UpdateLabel(manifest, api.LabelK8sEnvironment, manifest.Spec.Environment.Name)
		k8s.UpdateLabel(manifest, api.LabelK8sVirtualEnvironment, string(manifest.Spec.VirtualEnvironment.Name))

	case "kubefox.xigxog.io/v1alpha1, Kind=VirtualEnvironment":
		ve := &v1alpha1.VirtualEnvironment{}
		if err := r.DecodeRaw(req.Object, ve); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		obj = ve

		k8s.UpdateLabel(ve, api.LabelK8sEnvironment, ve.Spec.Environment)

	default:
		return admission.Allowed("🦊")
	}

	current, err := json.Marshal(obj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, current)
}
