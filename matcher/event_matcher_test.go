package matcher

import (
	"testing"

	kubefox "github.com/xigxog/kubefox/core"
)

func TestPath(t *testing.T) {
	p := New()

	v := map[string]*kubefox.Val{
		"a": kubefox.ValString("127"),
		"b": kubefox.ValArrayInt([]int{0, 1}),
		"c": kubefox.ValArrayString([]string{"a", "b"}),
	}

	r1 := &kubefox.Route{
		RouteSpec: kubefox.RouteSpec{
			Id:   1,
			Rule: "PathPrefix(`/customize/{{.Env.b}}`)",
		},
		Component:    &kubefox.Component{},
		EventContext: &kubefox.EventContext{},
	}
	if err := r1.Resolve(v, nil); err != nil {
		t.Log(err)
		t.FailNow()
	}

	r2 := &kubefox.Route{
		RouteSpec: kubefox.RouteSpec{
			Id:   2,
			Rule: "Method(`PUT`,`GET`,`POST`) && (Query(`q1`, `{q[1-2]}`) && Header(`header-one`,`{[a-z0-9]+}`)) && Host(`{{.Env.a}}.0.0.{i}`) && Path(`/customize/{{.Env.b}}/{j:[a-z]+}`)",
		},
		Component:    &kubefox.Component{},
		EventContext: &kubefox.EventContext{},
	}
	if err := r2.Resolve(v, nil); err != nil {
		t.Log(err)
		t.FailNow()
	}

	if err := p.AddRoutes([]*kubefox.Route{r1, r2}); err != nil {
		t.Log(err)
		t.FailNow()
	}

	e := evt(kubefox.EventTypeHTTP)
	r, match := p.Match(e)

	if !match {
		t.Log("should have got a match :(")
		t.FailNow()
	}
	if r.Id != 2 {
		t.Log("incorrect route matched :(")
		t.FailNow()
	}
	t.Logf("matched route %d; i=%s, j=%s", r.Id, e.Param("i"), e.Param("j"))

}

func evt(evtType kubefox.EventType) *kubefox.Event {
	evt := kubefox.NewEvent()
	evt.Type = string(evtType)
	evt.SetValue(kubefox.ValKeyURL, "http://127.0.0.1/customize/1/a?q1=q1&q2=q2")
	evt.SetValue(kubefox.ValKeyHost, "127.0.0.1")
	evt.SetValue(kubefox.ValKeyPath, "/customize/1/a")
	evt.SetValue(kubefox.ValKeyMethod, "GET")
	evt.SetValueMap(kubefox.ValKeyHeader, map[string][]string{
		"Header-One": {"h1"},
	})
	evt.SetValueMap(kubefox.ValKeyQuery, map[string][]string{
		"q1": {"q1"},
		"q2": {"q2"},
	})

	return evt
}
