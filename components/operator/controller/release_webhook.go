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
	"github.com/xigxog/kubefox/logkf"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type ReleaseWebhook struct {
	*Client
	*admission.Decoder

	Mutating bool
}

func (r *ReleaseWebhook) Handle(ctx context.Context, req admission.Request) admission.Response {
	if !r.Mutating {
		if _, err := r.Client.GetPlatform(ctx, req.Namespace); err != nil {
			switch err {
			case ErrNotFound:
				return admission.Denied(
					fmt.Sprintf(`The Release "%s" not allowed: Platform not found in Namespace "%s"`, req.Name, req.Namespace))
			case ErrTooManyPlatforms:
				return admission.Denied(
					fmt.Sprintf(`The Release "%s" not allowed: More than one Platform found in Namespace "%s"`, req.Name, req.Namespace))
			default:
				return admission.Errored(http.StatusInternalServerError, err)
			}
		}
	}

	rel := &v1alpha1.Release{}
	if err := r.Decode(req, rel); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	logkf.Global.DebugInterface("userInfo", req.UserInfo)

	if rel.Spec.Environment.Name == "" {
		return admission.Denied(
			fmt.Sprintf(`The Release "%s" is invalid: spec.environment.name: Required value`, rel.Name))
	}
	if rel.Spec.Environment.Vars != nil {
		return admission.Denied(
			fmt.Sprintf(`The Release "%s" is invalid: spec.environment.vars: Value must be null`, rel.Name))
	}
	if rel.Spec.Environment.Adapters != nil {
		return admission.Denied(
			fmt.Sprintf(`The Release "%s" is invalid: spec.environment.adapters: Value must be null`, rel.Name))
	}

	env := &v1alpha1.Environment{}
	if err := r.Get(ctx, NN("", rel.Spec.Environment.Name), env); client.IgnoreNotFound(err) != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	} else if apierrors.IsNotFound(err) {
		return admission.Errored(http.StatusNotFound, err)
	}

	if rel.Spec.Environment.UID != "" && rel.Spec.Environment.UID != env.UID {
		return admission.Denied(
			fmt.Sprintf(`The Release "%s" is invalid: spec.environment.uid: Value does not match "%s"`, rel.Name, env.UID))
	}
	if rel.Spec.Environment.ResourceVersion != "" && rel.Spec.Environment.ResourceVersion != env.ResourceVersion {
		return admission.Denied(
			fmt.Sprintf(`The Release "%s" is invalid: spec.environment.resourceVersion: Value does not match "%s"`, rel.Name, env.ResourceVersion))
	}

	rel.Spec.Environment.UID = env.UID
	rel.Spec.Environment.ResourceVersion = env.ResourceVersion
	rel.Spec.Environment.Vars = env.Spec.Vars
	rel.Spec.Environment.Adapters = env.Spec.Adapters

	current, err := json.Marshal(rel)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, current)

}
