package kit

import (
	"context"
	"io"
	"net/http"

	kubefox "github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/logkf"
)

type EventHandler func(kit Kontext) error

type EnvVarOption func(*kubefox.EnvVarSchema)

type EnvVarType kubefox.EnvVarType

const (
	Array   EnvVarType = EnvVarType(kubefox.EnvVarTypeArray)
	Boolean EnvVarType = EnvVarType(kubefox.EnvVarTypeBoolean)
	Number  EnvVarType = EnvVarType(kubefox.EnvVarTypeNumber)
	String  EnvVarType = EnvVarType(kubefox.EnvVarTypeString)
)

type Kit interface {
	Start()

	Route(string, EventHandler)
	Default(EventHandler)

	EnvVar(name string, opts ...EnvVarOption) EnvVar

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
}

type Dependency interface {
	GetName() string
	GetType() kubefox.ComponentType
	GetEventType() kubefox.EventType
}
