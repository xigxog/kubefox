// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"testing"
)

func Test_clone(t *testing.T) {
	parent := NewEvent()
	parent.Context = &EventContext{
		AppDeployment: "blah",
	}

	evt := NewErr(ErrNotFound(), EventOpts{})

	clone := CloneToResp(evt, EventOpts{
		Parent: parent,
	})

	if clone.ParentId != parent.Id {
		t.Fail()
	}
	if clone.Context.AppDeployment != "blah" {
		t.Fail()
	}
	if clone.ContentType != evt.ContentType {
		t.Fail()
	}
	if clone.Category != Category_RESPONSE {
		t.Fail()
	}
	if clone.Type != evt.Type {
		t.Fail()
	}
}

func TestEvent_SetHeader(t *testing.T) {
	evt := NewEvent()
	evt.SetHeader("test", "test")
	evt.SetHeader("test2", "test2")
	if evt.Header("test") != "test" || evt.Header("test2") != "test2" {
		t.Fail()
	}
}
