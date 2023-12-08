package matcher

import (
	"testing"

	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/core"
)

func TestPath(t *testing.T) {
	p := New()

	v := map[string]*api.Val{
		"a": api.ValString("127"),
		"b": api.ValArrayInt([]int{0, 1}),
		"c": api.ValArrayString([]string{"a", "b"}),
	}

	r1 := &core.Route{
		RouteSpec: api.RouteSpec{
			Id:   1,
			Rule: "PathPrefix(`/customize/{{b}}`)",
		},
		Component:    &core.Component{},
		EventContext: &core.EventContext{},
	}
	if err := r1.Resolve(v); err != nil {
		t.Log(err)
		t.FailNow()
	}

	r2 := &core.Route{
		RouteSpec: api.RouteSpec{
			Id:   2,
			Rule: "Type(`http`) && Method(`PUT`,`GET`,`POST`) && (Query(`q1`, `{q[1-2]}`) && Header(`header-one`,`{[a-z0-9]+}`)) && Host(`{{a}}.0.0.{i}`) && Path(`/customize/{{b}}/{j:[a-z]+}`)",
		},
		Component:    &core.Component{},
		EventContext: &core.EventContext{},
	}
	if err := r2.Resolve(v); err != nil {
		t.Log(err)
		t.FailNow()
	}

	if err := p.AddRoutes([]*core.Route{r1, r2}); err != nil {
		t.Log(err)
		t.FailNow()
	}

	e := evt(api.EventTypeHTTP)
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

func evt(evtType api.EventType) *core.Event {
	evt := core.NewEvent()
	evt.Type = string(evtType)
	evt.SetValue(api.ValKeyURL, "http://127.0.0.1/customize/1/a?q1=q1&q2=q2")
	evt.SetValue(api.ValKeyHost, "127.0.0.1")
	evt.SetValue(api.ValKeyPath, "/customize/1/a")
	evt.SetValue(api.ValKeyMethod, "GET")
	evt.SetValueMap(api.ValKeyHeader, map[string][]string{
		"Header-One": {"h1"},
	})
	evt.SetValueMap(api.ValKeyQuery, map[string][]string{
		"q1": {"q1"},
		"q2": {"q2"},
	})

	return evt
}
