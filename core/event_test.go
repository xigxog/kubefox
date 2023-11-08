package core

import (
	"testing"
)

func Test_clone(t *testing.T) {
	parent := NewEvent()
	parent.Context = &EventContext{
		Release: "blah",
	}

	evt := NewErr(ErrNotFound(), EventOpts{})

	clone := CloneToResp(evt, EventOpts{
		Parent: parent,
	})

	if clone.ParentId != parent.Id {
		t.Fail()
	}
	if clone.Context.Release != "blah" {
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
