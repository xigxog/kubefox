// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"regexp"

	"github.com/xigxog/kubefox/api"
)

var (
	RuleParamRegexp = regexp.MustCompile(`([^\\])({[^}]+})`)
)

type Route struct {
	*api.EnvTemplate

	Id           int
	ResolvedRule string
	Priority     int

	Component    *Component
	EventContext *EventContext
}

func NewRoute(id int, rule string) (*Route, error) {
	tpl := api.NewEnvTemplate("route", rule)
	if tpl.ParseError() != nil {
		return nil, tpl.ParseError()
	}

	return &Route{
		EnvTemplate: tpl,
		Id:          id,
	}, nil
}

func (r *Route) Resolve(data *api.Data) (err error) {
	if r.ResolvedRule, err = r.EnvTemplate.Resolve(data, false); err != nil {
		return
	}
	// Normalize path args so they don't affect length.
	r.Priority = len(RuleParamRegexp.ReplaceAllString(r.ResolvedRule, "$1{}"))

	return
}
