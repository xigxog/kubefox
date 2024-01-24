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
	"fmt"
	"net/http"

	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/k8s"
	admv1 "k8s.io/api/admission/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type ImmutableWebhook struct {
	*admission.Decoder
}

const (
	notAllowedMsg = "update operation not allowed: %s is immutable"
)

func (r *ImmutableWebhook) Handle(ctx context.Context, req admission.Request) admission.Response {
	if req.Operation != admv1.Update {
		return admission.Allowed("🦊")
	}

	switch req.Kind.String() {
	case "kubefox.xigxog.io/v1alpha1, Kind=AppDeployment":
		obj := &v1alpha1.AppDeployment{}
		if err := r.DecodeRaw(req.Object, obj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		oldObj := &v1alpha1.AppDeployment{}
		if err := r.DecodeRaw(req.OldObject, oldObj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		if oldObj.Spec.Version != "" {
			if !k8s.DeepEqual(&obj.Spec, &oldObj.Spec) {
				return admission.Denied(fmt.Sprintf(notAllowedMsg, "AppDeployment with version"))
			}
		}

	case "kubefox.xigxog.io/v1alpha1, Kind=ReleaseManifest":
		obj := &v1alpha1.ReleaseManifest{}
		if err := r.DecodeRaw(req.Object, obj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		oldObj := &v1alpha1.ReleaseManifest{}
		if err := r.DecodeRaw(req.OldObject, oldObj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		if !k8s.DeepEqual(&obj.Spec, &oldObj.Spec) || !k8s.DeepEqual(&obj.Data, &oldObj.Data) {
			return admission.Denied(fmt.Sprintf(notAllowedMsg, "ReleaseManifest"))
		}

	case "kubefox.xigxog.io/v1alpha1, Kind=VirtualEnvironment":
		obj := &v1alpha1.VirtualEnvironment{}
		if err := r.DecodeRaw(req.Object, obj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		oldObj := &v1alpha1.VirtualEnvironment{}
		if err := r.DecodeRaw(req.OldObject, oldObj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		if obj.Spec.Environment != oldObj.Spec.Environment {
			return admission.Denied(fmt.Sprintf(notAllowedMsg, ".spec.environment"))
		}
	}

	return admission.Allowed("🦊")
}
