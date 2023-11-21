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
	v1 "k8s.io/api/admission/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type ImmutableWebhook struct {
	*Client
	*admission.Decoder
}

func (r *ImmutableWebhook) Handle(ctx context.Context, req admission.Request) admission.Response {
	if req.Operation != v1.Update {
		return admission.Allowed("🦊")
	}

	var (
		lhs, rhs any
	)
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
		if obj.Spec.Version != "" {
			lhs = &obj.Spec
			rhs = &oldObj.Spec

		}

	case "kubefox.xigxog.io/v1alpha1, Kind=Release":
		obj := &v1alpha1.Release{}
		if err := r.DecodeRaw(req.Object, obj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		oldObj := &v1alpha1.Release{}
		if err := r.DecodeRaw(req.OldObject, oldObj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		if obj.Spec.Version != "" {
			lhs = &obj.Spec
			rhs = &oldObj.Spec

		}

	case "kubefox.xigxog.io/v1alpha1, Kind=ResolvedEnvironment":
		obj := &v1alpha1.ResolvedEnvironment{}
		if err := r.DecodeRaw(req.Object, obj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		oldObj := &v1alpha1.ResolvedEnvironment{}
		if err := r.DecodeRaw(req.OldObject, oldObj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		lhs = &obj.Data
		rhs = &oldObj.Data

	}

	if !HashesEqual(lhs, rhs) {
		return admission.Denied(fmt.Sprintf(
			"update operation not allowed: %s spec is immutable", req.Kind.Kind))
	}

	return admission.Allowed("🦊")
}
