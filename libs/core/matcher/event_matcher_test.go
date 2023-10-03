package matcher

import (
	"testing"

	"github.com/xigxog/kubefox/libs/core/kubefox"
)

func TestPath(t *testing.T) {
	p, err := New(&kubefox.Component{
		Name:   "test",
		Commit: "0000000",
	})
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	v := map[string]*kubefox.Var{
		"a": kubefox.NewVarString("127"),
		"b": kubefox.NewVarArrayInt([]int{0, 1}),
		"c": kubefox.NewVarArrayString([]string{"a", "b"}),
	}

	r1 := &kubefox.Route{
		Id:   1,
		Rule: "PathPrefix(`/customize/{{.Env.b}}`)",
	}
	err = r1.Resolve(v)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	r2 := &kubefox.Route{
		Id:   2,
		Rule: "Method(`PUT`,`GET`,`POST`) && (Query(`q1`, `{q[1-2]}`) && Header(`header-one`,`{[a-z0-9]+}`)) && Host(`{{.Env.a}}.0.0.{i}`) && Path(`/customize/{{.Env.b}}/{j:[a-z]+}`)",
	}
	err = r2.Resolve(v)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	if err := p.AddRoutes([]*kubefox.Route{r1, r2}); err != nil {
		t.Log(err)
		t.FailNow()
	}

	e := evt(kubefox.HTTPRequestType)
	_, r, match := p.Match(e)

	if !match {
		t.Log("should have got a match :(")
		t.FailNow()
	}
	if r.Id != 2 {
		t.Log("incorrect route matched :(")
		t.FailNow()
	}
	t.Logf("matched route %d; i=%s, j=%s", r.Id, e.GetParam("i"), e.GetParam("j"))

}

func evt(evtType kubefox.EventType) *kubefox.Event {
	evt := kubefox.NewEvent()
	evt.Type = string(evtType)
	evt.SetValue("url", "http://127.0.0.1/customize/1/a?q1=q1&q2=q2")
	evt.SetValue("host", "127.0.0.1")
	evt.SetValue("path", "/customize/1/a")
	evt.SetValue("method", "GET")
	evt.SetValueMap("header", map[string][]string{
		"Header-One": {"h1"},
	})
	evt.SetValueMap("query", map[string][]string{
		"q1": {"q1"},
		"q2": {"q2"},
	})

	return evt
}
