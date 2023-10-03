/*
Copyright © 2023 XigXog

This Source Code Form is subject to the terms of the Mozilla Public License,
v2.0. If a copy of the MPL was not distributed with this file, You can obtain
one at https://mozilla.org/MPL/2.0/.
*/

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var platformlog = logf.Log.WithName("platform-resource")

func (r *Platform) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-kubefox-xigxog-io-v1alpha1-platform,mutating=false,failurePolicy=fail,sideEffects=None,groups=kubefox.xigxog.io,resources=platforms,verbs=create;update,versions=v1alpha1,name=vplatform.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Platform{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Platform) ValidateCreate() (admission.Warnings, error) {
	platformlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Platform) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	platformlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Platform) ValidateDelete() (admission.Warnings, error) {
	platformlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
