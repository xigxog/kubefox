// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

// +kubebuilder:object:generate=true
package kubernetes

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// StripObjectMeta clears all fields except Name, Namespace, UID,
// ResourceVersion, and Generation.
func StripObjectMeta(meta *metav1.ObjectMeta) {
	meta.Annotations = nil
	meta.DeletionGracePeriodSeconds = nil
	meta.DeletionTimestamp = nil
	meta.Finalizers = nil
	meta.GenerateName = ""
	meta.Labels = nil
	meta.ManagedFields = nil
	meta.OwnerReferences = nil
}
