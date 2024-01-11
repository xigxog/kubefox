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
	"github.com/xigxog/kubefox/utils"
	admv1 "k8s.io/api/admission/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type IndexWebhook struct {
	*admission.Decoder
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

		k8s.UpdateLabel(appDep, api.LabelK8sAppVersion, appDep.Spec.Version)
		k8s.UpdateLabel(appDep, api.LabelK8sAppCommit, appDep.Spec.Commit)
		k8s.UpdateLabel(appDep, api.LabelK8sAppCommitShort, utils.ShortCommit(appDep.Spec.Commit))
		k8s.UpdateLabel(appDep, api.LabelK8sAppTag, appDep.Spec.Tag)
		k8s.UpdateLabel(appDep, api.LabelK8sAppBranch, appDep.Spec.Branch)

	case "kubefox.xigxog.io/v1alpha1, Kind=Environment":
		if req.Operation == admv1.Create {
			env := &v1alpha1.Environment{}
			if err := r.DecodeRaw(req.Object, env); err != nil {
				return admission.Errored(http.StatusBadRequest, err)
			}
			obj = env

			k8s.AddFinalizer(env, api.FinalizerEnvironmentProtection)
		}

	case "kubefox.xigxog.io/v1alpha1, Kind=ReleaseManifest":
		env := &v1alpha1.ReleaseManifest{}
		if err := r.DecodeRaw(req.Object, env); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		obj = env

		k8s.UpdateLabel(env, api.LabelK8sEnvironment, string(env.Spec.VirtualEnvironment.Name))
		k8s.UpdateLabel(env, api.LabelK8sVirtualEnvironment, env.Spec.VirtualEnvironment.Environment)

	case "kubefox.xigxog.io/v1alpha1, Kind=VirtualEnvironment":
		env := &v1alpha1.VirtualEnvironment{}
		if err := r.DecodeRaw(req.Object, env); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		obj = env

		k8s.UpdateLabel(env, api.LabelK8sEnvironment, env.Spec.Environment)

	default:
		return admission.Allowed("🦊")
	}

	current, err := json.Marshal(obj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, current)
}
