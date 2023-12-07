package kit

import (
	"context"
	"io"
	"net/http"

	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/kit/env"
	"github.com/xigxog/kubefox/logkf"
)

type EventHandler func(kit Kontext) error

type Kit interface {
	Start()

	Route(string, EventHandler)
	Default(EventHandler)

	EnvVar(name string, opts ...env.VarOption) EnvVar

	Component(name string) Dependency
	HTTPAdapter(name string) Dependency

	Title(title string)
	Description(description string)

	Log() *logkf.Logger
}

type Kontext interface {
	core.EventReader

	Env(v EnvVar) string
	EnvV(v EnvVar) *api.Val
	EnvDef(v EnvVar, def string) string
	EnvDefV(v EnvVar, def *api.Val) *api.Val

	Resp() Resp
	ForwardResp(resp core.EventReader) Resp

	Req(c Dependency) Req
	Forward(c Dependency) Req
	HTTP(c Dependency) *http.Client
	Transport(c Dependency) http.RoundTripper

	Context() context.Context
	Log() *logkf.Logger
}

type Req interface {
	core.EventWriter

	SendStr(s string) (core.EventReader, error)
	SendHTML(h string) (core.EventReader, error)
	SendJSON(v any) (core.EventReader, error)
	SendBytes(contentType string, b []byte) (core.EventReader, error)
	SendReader(contentType string, reader io.Reader) (core.EventReader, error)
	Send() (core.EventReader, error)
}

type Resp interface {
	core.EventWriter

	SendStr(s string) error
	SendHTML(h string) error
	SendJSON(v any) error
	SendAccepts(json any, html, str string) error
	SendBytes(contentType string, b []byte) error
	SendReader(contentType string, reader io.Reader) error
	Send() error
}

type EnvVar interface {
	Name() string
	Type() api.EnvVarType
}

type Dependency interface {
	Name() string
	Type() api.ComponentType
	EventType() api.EventType
}

type route struct {
	api.RouteSpec

	handler EventHandler
}

type dependency struct {
	typ  api.ComponentType
	name string
}

func (c *dependency) Name() string {
	return c.name
}

func (c *dependency) Type() api.ComponentType {
	return c.typ
}

func (c *dependency) EventType() api.EventType {
	switch c.typ {
	case api.ComponentTypeHTTP:
		return api.EventTypeHTTP
	default:
		return api.EventTypeKubeFox
	}
}
