// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package kit

import (
	"net/http"
)

type EventRoundTripper struct {
	req *reqKontext
}

func (rt *EventRoundTripper) RoundTrip(httpReq *http.Request) (*http.Response, error) {
	if err := rt.req.SetHTTPRequest(httpReq, rt.req.ktx.kit.maxEventSize); err != nil {
		return nil, err
	}

	resp, err := rt.req.ktx.sendReq(rt.req.Event)
	if err != nil {
		return nil, err
	}

	return resp.HTTPResponse(), nil
}

func (rt *EventRoundTripper) Do(httpReq *http.Request) (*http.Response, error) {
	return rt.RoundTrip(httpReq)
}
