package kit

import (
	"context"
	"io"
	"net/http"

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
	EnvV(v EnvVar) *kubefox.Val
	EnvDef(v EnvVar, def string) string

	Resp() Resp

	Req(c Dependency) Req
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
	GetName() string
	GetType() kubefox.EnvVarType
}

type Dependency interface {
	GetName() string
	GetType() kubefox.ComponentType
	GetEventType() kubefox.EventType
}

type route struct {
	kubefox.RouteSpec

	handler EventHandler
}

type dependency struct {
	kubefox.ComponentTypeVar

	Name string
}

func (c *dependency) GetName() string {
	return c.Name
}

func (c *dependency) GetType() kubefox.ComponentType {
	return c.Type
}

func (c *dependency) GetEventType() kubefox.EventType {
	switch c.Type {
	case kubefox.ComponentTypeHTTP:
		return kubefox.EventTypeHTTP
	default:
		return kubefox.EventTypeKubeFox
	}
}
