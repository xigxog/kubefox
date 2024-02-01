// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/utils"
)

func GenerateId() string {
	_, id := GenerateNameAndId()
	return id
}

func GenerateNameAndId() (string, string) {
	id := uuid.NewString()
	name := id
	if p, _ := os.LookupEnv(api.EnvPodName); p != "" {
		name = p
		s := strings.Split(p, "-")
		if len(s) > 1 {
			id = s[len(s)-1]
		}
	} else if h, _ := os.Hostname(); h != "" {
		name = h
	}

	return utils.CleanName(name), id
}

func (c *Component) IsComplete() bool {
	return c.IsGroupComplete() && c.Id != "" && c.BrokerId != ""
}

func (c *Component) IsGroupComplete() bool {
	return c.Type != "" && c.Name != "" && c.Commit != ""
}

func (c *Component) IsNameOnly() bool {
	return c.Type != "" && c.Name != "" && c.Commit == "" && c.Id == "" && c.BrokerId == ""
}

func (lhs *Component) Equal(rhs *Component) bool {
	if lhs == nil || rhs == nil {
		return false
	}
	return lhs.Type == rhs.Type &&
		lhs.Name == rhs.Name &&
		lhs.Commit == rhs.Commit &&
		lhs.Id == rhs.Id &&
		lhs.BrokerId == rhs.BrokerId
}

func (c *Component) Key() string {
	return fmt.Sprintf("%s-%s", c.GroupKey(), c.Id)
}

// TODO add back app as part of comp id

func (c *Component) GroupKey() string {
	return utils.CleanName(fmt.Sprintf("%s-%s", c.Name, c.Commit))
}

func (c *Component) Subject() string {
	if c.BrokerId != "" {
		return c.BrokerSubject()
	}
	if c.Id == "" {
		return c.GroupSubject()
	}
	return fmt.Sprintf("evt.js.%s.%s.%s", c.Name, c.Commit, c.Id)
}

func (c *Component) GroupSubject() string {
	return utils.CleanName(fmt.Sprintf("evt.js.%s.%s", c.Name, c.Commit))
}

func (c *Component) BrokerSubject() string {
	return fmt.Sprintf("evt.brk.%s", c.BrokerId)
}
