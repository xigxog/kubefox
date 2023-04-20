package kubefox

import (
	"testing"
)

func TestPath(t *testing.T) {
	p, err := NewEventMatcher("Path(`customize/{id}`) && !Path(`erg`)")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	e := evt(HTTPRequestType)
	match := p.Match(e)
	if !match {
		t.Log("should have got a match :(")
		t.Fail()
	}
}

func TestHook(t *testing.T) {
	p, err := NewEventMatcher("Hook(`customize`)")
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	match := p.Match(evt(KubernetesRequestType))
	if !match {
		t.Log("should have got a match :(")
		t.Fail()
	}
}

func evt(evtType string) Event {
	evt := NewDataEvent(evtType)
	evt.HTTPData().SetURLString("http://127.0.0.1/customize/1")
	return evt
}
