// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package controller

import (
	"testing"

	"golang.org/x/mod/semver"
)

func TestSemVerCompare(t *testing.T) {
	compare("main", "v0.1.0-alpha", t)
	compare("v0.1.0-alpha", "v0.2.0-alpha", t)
	compare("v0.2.0-alpha", "v0.1.0-alpha", t)
	compare("v0.2.0-alpha", "v0.2.0-beta", t)
}

func compare(v, w string, t *testing.T) {
	switch semver.Compare(v, w) {
	case -1:
		t.Logf("%s < %s", v, w)
	case 0:
		t.Logf("%s == %s", v, w)
	case 1:
		t.Logf("%s > %s", v, w)
	}
}
