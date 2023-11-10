package kit

import (
	"context"
	"io"
	"net/http"

	common "github.com/xigxog/kubefox/api/kubernetes"
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
	EnvV(v EnvVar) *common.Val
	EnvDef(v EnvVar, def string) string
	EnvDefV(v EnvVar, def *common.Val) *common.Val

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
	SendBytes(contentType string, b []byte) error
	SendReader(contentType string, reader io.Reader) error
	Send() error
}

type EnvVar interface {
	Name() string
	Type() common.EnvVarType
}

type Dependency interface {
	Name() string
	Type() common.ComponentType
	EventType() kubefox.EventType
}

type route struct {
	common.RouteSpec

	handler EventHandler
}

type dependency struct {
	typ  common.ComponentType
	name string
}

func (c *dependency) Name() string {
	return c.name
}

func (c *dependency) Type() common.ComponentType {
	return c.typ
}

func (c *dependency) EventType() kubefox.EventType {
	switch c.typ {
	case common.ComponentTypeHTTP:
		return kubefox.EventTypeHTTP
	default:
		return kubefox.EventTypeKubeFox
	}
}
