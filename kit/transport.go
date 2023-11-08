package kit

import (
	"net/http"
)

type EventRoundTripper struct {
	req *reqKontext
}

func (rt *EventRoundTripper) RoundTrip(httpReq *http.Request) (*http.Response, error) {
	if err := rt.req.SetHTTPRequest(httpReq); err != nil {
		return nil, err
	}

	resp, err := rt.req.brk.SendReq(rt.req.ctx, rt.req.Event, rt.req.start)
	if err != nil {
		return nil, err
	}

	return resp.HTTPResponse(), nil
}

func (rt *EventRoundTripper) Do(httpReq *http.Request) (*http.Response, error) {
	return rt.RoundTrip(httpReq)
}
