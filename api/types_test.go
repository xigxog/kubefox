/*
Copyright © 2023 XigXog

This Source Code Form is subject to the terms of the Mozilla Public License,
v2.0. If a copy of the MPL was not distributed with this file, You can obtain
one at https://mozilla.org/MPL/2.0/.
*/

// +kubebuilder:object:generate=true
package api

import (
	"testing"

	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestUncomparableTime(t *testing.T) {
	var t1 struct {
		T UncomparableTime
	}
	var t2 struct {
		T UncomparableTime
	}
	t1.T = UncomparableTime(metav1.Now())
	t2.T = UncomparableTime{}

	if !equality.Semantic.DeepEqual(t1, t2) {
		t.Fail()
	}
}
