package kit

import (
	"context"
	"io"
	"net/http"

	"github.com/xigxog/kubefox/api"
	kubefox "github.com/xigxog/kubefox/core"
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
	kubefox.EventReader

	Env(v EnvVar) string
	EnvV(v EnvVar) *api.Val
	EnvDef(v EnvVar, def string) string
	EnvDefV(v EnvVar, def *api.Val) *api.Val

	Resp() Resp
	ForwardResp(resp kubefox.EventReader) Resp

	Req(c Dependency) Req
	Forward(c Dependency) Req
	HTTP(c Dependency) *http.Client
	Transport(c Dependency) http.RoundTripper

	Context() context.Context
	Log() *logkf.Logger
}

type Req interface {
	kubefox.EventWriter

	SendStr(s string) (kubefox.EventReader, error)
	SendHTML(h string) (kubefox.EventReader, error)
	SendJSON(v any) (kubefox.EventReader, error)
	SendBytes(contentType string, b []byte) (kubefox.EventReader, error)
	SendReader(contentType string, reader io.Reader) (kubefox.EventReader, error)
	Send() (kubefox.EventReader, error)
}

type Resp interface {
	kubefox.EventWriter

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
