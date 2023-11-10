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

	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type ReleaseWebhook struct {
	*Client
	*admission.Decoder
}

// TODO environment checks
// - check for required vars
// - check for unique vars
// - check for var type
// - check for required dependencies (comps and adapters)

func (r *ReleaseWebhook) Handle(ctx context.Context, req admission.Request) admission.Response {
	rel := &v1alpha1.Release{}
	if err := r.Decode(req, rel); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if rel.Name != rel.Spec.Environment.Name {
		return admission.Denied(
			fmt.Sprintf(`The Release "%s" is invalid: metadata.name: Value must match spec.environment.name`, rel.Name))
	}
	if rel.Spec.Environment.Vars != nil {
		return admission.Denied(
			fmt.Sprintf(`The Release "%s" is invalid: spec.environment.vars: Value is readonly`, rel.Name))
	}
	if rel.Spec.Environment.Adapters != nil {
		return admission.Denied(
			fmt.Sprintf(`The Release "%s" is invalid: spec.environment.adapters: Value is readonly`, rel.Name))
	}
	if rel.Spec.AppDeployment.App != nil {
		return admission.Denied(
			fmt.Sprintf(`The Release "%s" is invalid: spec.deployment.app: Value is readonly`, rel.Name))
	}
	if rel.Spec.AppDeployment.Components != nil {
		return admission.Denied(
			fmt.Sprintf(`The Release "%s" is invalid: spec.deployment.components: Value is readonly`, rel.Name))
	}

	env := &v1alpha1.Environment{}
	if code, err := r.checkRes(ctx, env,
		NN("", rel.Spec.Environment.Name),
		"spec.environment",
		rel.Spec.Environment.UID,
		rel.Spec.Environment.ResourceVersion,
	); err != nil {
		if code != 0 {
			return admission.Errored(code, err)
		}
		return admission.Denied(fmt.Sprintf(`The Release "%s" is invalid: %s`, rel.Name, err))
	}

	rel.Spec.Environment.UID = env.UID
	rel.Spec.Environment.ResourceVersion = env.ResourceVersion
	rel.Spec.Environment.EnvSpec = env.Spec.EnvSpec

	appDep := &v1alpha1.AppDeployment{}
	if code, err := r.checkRes(ctx, appDep,
		NN(rel.Namespace, rel.Spec.AppDeployment.Name),
		"spec.deployment",
		rel.Spec.AppDeployment.UID,
		rel.Spec.AppDeployment.ResourceVersion,
	); err != nil {
		if code != 0 {
			return admission.Errored(code, err)
		}
		return admission.Denied(fmt.Sprintf(`The Release "%s" is invalid: %s`, rel.Name, err))
	}

	if rel.Spec.Version == "" {
		rel.Spec.Version = appDep.Details.App.Tag
	}
	if rel.Spec.Version == "" {
		return admission.Denied(
			fmt.Sprintf(`The Release "%s" is invalid: spec.version: Value is required`, rel.Name))
	}
	if appDep.Details.App.Tag != "" && rel.Spec.Version != appDep.Details.App.Tag {
		return admission.Denied(
			fmt.Sprintf(`The Release "%s" is invalid: spec.version: Value does not match AppDeployment tag name`, rel.Name))
	}

	rel.Spec.AppDeployment.UID = appDep.UID
	rel.Spec.AppDeployment.ResourceVersion = appDep.ResourceVersion
	rel.SetAppDeploymentSpec(&appDep.Spec)

	current, err := json.Marshal(rel)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, current)
}

func (r *ReleaseWebhook) checkRes(ctx context.Context, obj client.Object, name types.NamespacedName, path string, uid types.UID, ver string) (int32, error) {
	if name.Name == "" {
		return 0, fmt.Errorf(`%s.name: Required value`, path)
	}
	if err := r.Get(ctx, name, obj); client.IgnoreNotFound(err) != nil {
		return http.StatusInternalServerError, err
	} else if apierrors.IsNotFound(err) {
		return http.StatusNotFound, err
	}
	if uid != "" && uid != obj.GetUID() {
		return 0, fmt.Errorf(`%s.uid: Value does not match "%s"`, path, uid)
	}
	if ver != "" && ver != obj.GetResourceVersion() {
		return 0, fmt.Errorf(`%s.uid: Value does not match "%s"`, path, ver)
	}

	return 0, nil
}
