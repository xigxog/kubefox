/*
Copyright © 2023 XigXog

This Source Code Form is subject to the terms of the Mozilla Public License,
v2.0. If a copy of the MPL was not distributed with this file, You can obtain
one at https://mozilla.org/MPL/2.0/.
*/

package controller

// Set by main at init.
var (
	GitRef    string
	GitCommit string
)

const (
	LabelPlatform  string = "kubefox.xigxog.io/platform"
	LabelComponent string = "kubefox.xigxog.io/component"
	TenYears       string = "87600h"
	HundredYears   string = "876000h"
)
