package kubefox

import (
	"net/http"
	"testing"
)

func TestEvent_HTTPReq(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://xigxog.io/test?q1=a&q2=b", http.NoBody)
	req.Header.Add("h1", "a1")
	req.Header.Add("h1", "a2")
	req.Header.Set("h2", "b")

	evt := NewEvent()
	evt.SetHTTPRequest(req)
	t.Logf("q1: %s, q2: %s", evt.Query("q1"), evt.Query("q2"))
	t.Logf("h1: %s, h2: %s", evt.Header("h1"), evt.Header("h2"))
}
