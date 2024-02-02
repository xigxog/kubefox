// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"math/rand"

	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/utils"
)

var idChars = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

func NewComponent(typ api.ComponentType, app, name, commit string) *Component {
	return &Component{
		Type:   string(typ),
		App:    utils.CleanName(app),
		Name:   utils.CleanName(name),
		Commit: commit,
	}
}

func NewTargetComponent(typ api.ComponentType, name string) *Component {
	return &Component{
		Type: string(typ),
		Name: utils.CleanName(name),
	}
}

func NewPlatformComponent(typ api.ComponentType, name, commit string) *Component {
	return &Component{
		Type:   string(typ),
		Name:   utils.CleanName(name),
		Commit: commit,
	}
}

func GenerateId() string {
	b := make([]rune, 10)
	for i := range b {
		b[i] = idChars[rand.Intn(len(idChars))]
	}

	return string(b)
}

func (c *Component) IsComplete() bool {
	if c == nil {
		return false
	}

	return c.Type != "" && c.Name != "" && c.Commit != "" && c.Id != "" && c.BrokerId != ""
}

func (c *Component) IsNameOnly() bool {
	if c == nil {
		return false
	}

	return c.Type != "" && c.Name != "" && c.Commit == "" && c.Id == "" && c.BrokerId == ""
}

func (lhs *Component) Equal(rhs *Component) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	return lhs.Type == rhs.Type &&
		lhs.App == rhs.App &&
		lhs.Name == rhs.Name &&
		lhs.Commit == rhs.Commit &&
		lhs.Id == rhs.Id &&
		lhs.BrokerId == rhs.BrokerId
}

func (c *Component) Key() string {
	return c.GroupKey() + "-" + c.Id
}

func (c *Component) GroupKey() string {
	return utils.Join("-", c.App, c.Name, c.ShortCommit())
}

func (c *Component) Subject() string {
	if c.BrokerId != "" {
		return c.BrokerSubject()
	}

	return utils.Join(".", c.GroupSubject(), c.Id)
}

func (c *Component) GroupSubject() string {
	return utils.Join(".", "evt.cmp", c.App, c.Name, c.ShortCommit())
}

func (c *Component) BrokerSubject() string {
	return utils.Join(".", "evt.brk", c.BrokerId)
}

func (c *Component) ShortCommit() string {
	if len(c.Commit) <= 7 {
		return c.Commit
	}

	return c.Commit[:7]
}
