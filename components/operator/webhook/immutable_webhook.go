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
				return admission.Denied(fmt.Sprintf(notAllowedMsg, req.Kind.Kind))
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
			return admission.Denied(fmt.Sprintf(notAllowedMsg, req.Kind.Kind))
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

		if obj.Spec.Release != nil && oldObj.Spec.Release != nil &&
			obj.Spec.Release.Id == oldObj.Spec.Release.Id &&
			!k8s.DeepEqual(obj.Spec.Release.Apps, oldObj.Spec.Release.Apps) {

			return admission.Denied(".spec.release.Id but be updated if changes are made to .spec.release.apps")
		}
	}

	return admission.Allowed("🦊")
}
