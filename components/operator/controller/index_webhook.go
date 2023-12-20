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
	"net/http"

	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	"github.com/xigxog/kubefox/k8s"
	"github.com/xigxog/kubefox/utils"
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
		k8s.UpdateLabel(appDep, api.LabelK8sAppCommit, appDep.Spec.App.Commit)
		k8s.UpdateLabel(appDep, api.LabelK8sAppCommitShort, utils.ShortCommit(appDep.Spec.App.Commit))
		k8s.UpdateLabel(appDep, api.LabelK8sAppTag, appDep.Spec.App.Tag)
		k8s.UpdateLabel(appDep, api.LabelK8sAppBranch, appDep.Spec.App.Branch)

	case "kubefox.xigxog.io/v1alpha1, Kind=VirtualEnv":
		env := &v1alpha1.VirtualEnv{}
		if err := r.DecodeRaw(req.Object, env); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		obj = env

		if env.Spec.Release != nil {
			k8s.UpdateLabel(env, api.LabelK8sAppDeployment, env.Spec.Release.AppDeployment.Name)
			k8s.UpdateLabel(env, api.LabelK8sAppVersion, env.Spec.Release.AppDeployment.Version)
			k8s.UpdateLabel(env, api.LabelK8sVirtualEnvSnapshot, env.Spec.Release.VirtualEnvSnapshot)
		} else {
			k8s.UpdateLabel(env, api.LabelK8sAppDeployment, "")
			k8s.UpdateLabel(env, api.LabelK8sAppVersion, "")
			k8s.UpdateLabel(env, api.LabelK8sVirtualEnvSnapshot, "")
		}
		k8s.UpdateLabel(env, api.LabelK8sVirtualEnvParent, env.Spec.Parent)

	case "kubefox.xigxog.io/v1alpha1, Kind=VirtualEnvSnapshot":
		env := &v1alpha1.VirtualEnvSnapshot{}
		if err := r.DecodeRaw(req.Object, env); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		obj = env

		k8s.UpdateLabel(env, api.LabelK8sVirtualEnv, env.Spec.Source.Name)
		k8s.UpdateLabel(env, api.LabelK8sSourceResourceVersion, env.Spec.Source.ResourceVersion)

	default:
		return admission.Allowed("🦊")
	}

	current, err := json.Marshal(obj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, current)
}
