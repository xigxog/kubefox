package kubefox

import (
	"encoding/json"
	"net/url"
	"strings"
)

type KubeEvent interface {
	Event

	GetHook() Hook
	GetTargetKind() string
}

type KubeDataEvent interface {
	DataEvent
	KubeEvent
}

type kubeEvent struct {
	*event
	hook       Hook
	targetKind string
}

func (e *kubeEvent) GetHook() Hook {
	if e.hook != Unknown {
		return e.hook
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()

	u, err := url.Parse(e.GetValue("http_url"))
	if err != nil {
		return Unknown
	}

	switch {
	case strings.HasPrefix(u.Path, "/customize"):
		e.hook = Customize
	case strings.HasPrefix(u.Path, "/sync"):
		e.hook = Sync
	}

	return e.hook
}

func (e *kubeEvent) GetTargetKind() string {
	if e.targetKind != "" {
		return e.targetKind
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()

	req := &RequestStub{}
	err := json.Unmarshal(e.GetContent(), &req)
	if err != nil {
		return ""
	}

	switch {
	case req.Object != nil:
		e.targetKind = req.Object.Kind
	case req.Parent != nil:
		e.targetKind = req.Parent.Kind
	}

	return e.targetKind
}
