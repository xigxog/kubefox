// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package matcher

import (
	"testing"

	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/core"
)

func TestPath(t *testing.T) {
	data := &api.Data{
		Vars: map[string]*api.Val{
			"a": api.ValString("127"),
			"b": api.ValArrayInt([]int{0, 1}),
			"c": api.ValArrayString([]string{"a", "b"}),
		},
	}

	route1, _ := core.NewRoute(1, "PathPrefix(`/customize/{{.Vars.b}}`)")
	route1.Resolve(data)

	route2, _ := core.NewRoute(2, "Type(`http`) && Method(`PUT`,`GET`,`POST`) && (Query(`q1`, `{q[1-2]}`) && Header(`header-one`,`{[a-z0-9]+}`)) && Host(`{{.Env.a}}.0.0.{i}`) && Path(`/customize/{{.Vars.b}}/{j:[a-z]+}`)")
	route2.Resolve(data)

	m := New()
	m.AddRoutes(route1, route2)

	e := evt(api.EventTypeHTTP)
	r, match := m.Match(e)

	if !match {
		t.Fatalf("should have got a match :(")
	}
	if r.Id != 2 {
		t.Fatalf("incorrect route matched :(")
	}

	t.Logf("matched route %d; i=%s, j=%s", r.Id, e.Param("i"), e.Param("j"))
}

func TestPathPrefix(t *testing.T) {
	route, _ := core.NewRoute(1, "PathPrefix(`/customize/{stuff...}`)")
	route.Resolve(nil)

	m := New()
	m.AddRoutes(route)

	e := evt(api.EventTypeHTTP)
	_, match := m.Match(e)
	if !match {
		t.Fatalf("should have got a match :(")
	}
	if s := e.Param("stuff"); s != "1/a" {
		t.Fatalf("expected 'stuff' param to be '1/a', got %s", s)
	}
}

func TestPriority(t *testing.T) {
	data := &api.Data{
		Vars: map[string]*api.Val{
			"subPath": api.ValString("qa"),
		},
	}

	route1, _ := core.NewRoute(1, "PathPrefix(`/{{.Vars.subPath}}/hasura/static`)")
	route1.Resolve(data)

	route2, _ := core.NewRoute(1, "Path(`/{{.Vars.subPath}}/hasura/{file}`)")
	route2.Resolve(data)

	m := New()
	m.AddRoutes(route1, route2)

	e := evt(api.EventTypeHTTP)
	e.SetValue(api.ValKeyURL, "http://127.0.0.1/qa/hasura/static/favicon.ico")
	e.SetValue(api.ValKeyPath, "/qa/hasura/static/favicon.ico")

	r, match := m.Match(e)
	if !match {
		t.Fatalf("should have got a match :(")
	}
	t.Logf("route: %v, suffix: %s", r, e.PathSuffix())
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
