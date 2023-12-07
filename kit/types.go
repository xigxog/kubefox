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

type EventHandler func(ktx Kontext) error

type Kit interface {
	// Start connects to the Broker passing the Component's Service Account
	// Token for authorization. Once connected Kit will accept incoming Events.
	// If an error occurs the program will exit with a status code of 1. Start
	// is a blocking call.
	Start()

	// Route registers a EventHandler for the specified rule. If an Event
	// matches the rule the Broker will route it to the Component and Kit will
	// call the EventHandler.
	//
	// Rules are written in a simple predicate based language that matches parts
	// of an Event. Some predicates accept inputs which should be surrounded
	// with back ticks. The boolean operators '&&' (and), '||' (or), and '!'
	// (not) can be used to combined predicates. The following predicates are
	// supported:
	//
	//   All()
	//     Matches all Events.
	//
	//   Header(`key`, `value`)
	//     Matches if a header `key` exists and is equal to `value`.
	//
	//   Host(`example.com`)
	//     Matches if the domain (host header value) is equal to input.
	//
	//   Method(`GET`, ...)
	//     Matches if the request method is one of the given methods (GET, POST,
	//     PUT, DELETE, PATCH, HEAD)
	//
	//   Path(`/path`)
	//     Matches if the request path is equal to given input.
	//
	//   PathPrefix(`/prefix`)
	//     Matches if the request path begins with given input.
	//
	//   Query(`key`, `value`)
	//     Matches if a query parameter `key` exists and is equal to `value`.
	//
	//   Type(`value`)
	//     Matches if Event type is equal to given input.
	//
	// Predicate inputs can utilize regular expressions to match and optionally
	// extract parts of an Event to a named parameter. Regular expression use
	// the format '{[REGEX]}' or '{[NAME]:[REGEX]}' to extract the matching part
	// to a parameter.
	//
	// Additionally, environment variables can be utilized in predicate inputs.
	// They are resolved at request time with the value specified in the
	// VirtualEnv. Environment variables can be used with the format
	// '{{.Env.[NAME]}}'.
	//
	// For example, the following will match Events of type 'http' that are
	// 'GET' requests and have a path with three parts. The first part of the
	// path must equal the value of th environment variable 'SUB_PATH', the
	// second part must equal 'orders', and the third part can be one or more
	// lower case letter or number. The third part of the path is extracted to
	// the parameter 'orderId' which can be used by the EventHandler:
	//
	//   kit.Route("Type(`http`) && Method(`GET`) && Path(`/{{.Env.SUB_PATH}}/orders/{orderId:[a-z0-9]+}`)",
	//     func(ktx kit.Kontext) error {
	//       return ktx.Resp().SendStr("The orderId is ", ktx.Param("orderId"))
	//     })
	Route(rule string, handler EventHandler)

	// Default registers a default EventHandler. If Kit receives an Event from
	// the Broker that does not match any registered rules the default
	// EventHandler is called.
	Default(handler EventHandler)

	// EnvVar registers an environment variable dependency with given options.
	// The returned EnvVarDep can be used by EventHandlers to retrieve the value
	// of the environment variable at request time.
	//
	// For example:
	//
	//   v := kit.EnvVar("SOME_VAR")
	//   kit.Route("Any()", func(ktx kit.Kontext) error {
	//       return ktx.Resp().SendStr("the value of SOME_VAR is ", ktx.Env(v))
	//   })
	EnvVar(name string, opts ...env.VarOption) EnvVarDep

	// Component registers a Component dependency. The returned ComponentDep can
	// be used by EventHandlers to invoke the Component at request time.
	//
	// For example:
	//
	//   b := kit.Component("backend")
	//   kit.Route("Any()", func(ktx kit.Kontext) error {
	//       r, _ := ktx.Req(backend).Send()
	//       return ktx.Resp().SendStr("the resp from backend is ", r.Str())
	//   })
	Component(name string) ComponentDep

	// HTTPAdapter registers a dependency on the HTTP Adapter. The returned
	// ComponentDep can be used by EventHandlers to invoke the Adapter at
	// request time.
	//
	// For example:
	//
	//   h := kit.HTTPAdapter("httpbin")
	//   kit.Route("Any()", func(ktx kit.Kontext) error {
	//       r, _ := ktx.HTTP(h).Get("/anything")
	//       return ktx.Resp().SendReader(r.Header.Get("content-type"), r.Body)
	//   })
	HTTPAdapter(name string) ComponentDep

	// Title sets the Component's title.
	Title(title string)

	// Description sets the Component's description.
	Description(description string)

	// Log returns a pre-configured structured logger for the Component.
	Log() *logkf.Logger
}

type Kontext interface {
	core.EventReader

	// Env returns the value of the given environment variable as a string. If
	// the environment variable does not exist or cannot be converted to a
	// string, empty string is returned. To check if an environment exists use
	// EnvV() and check if the returned Val's ValType is 'Nil'.
	Env(v EnvVarDep) string

	// EnvV returns the value of the given environment variable as a Val. It is
	// guaranteed the returned Val will not be nil. If the environment variable
	// does not exist the ValType of the returned Val will be 'Nil'.
	EnvV(v EnvVarDep) *api.Val

	// EnvDef returns the value of the given environment variable as a string.
	// If the environment variable does not exist, is empty, or cannot be
	// converted to a string, then the given 'def' string is returned.
	EnvDef(v EnvVarDep, def string) string

	// EnvDefV returns the value of the given environment variable as a Val. If
	// the environment variable does not exist, is an empty string, or an empty
	// array, then the given 'def' Val is returned. If the environment variable
	// exists and is a boolean or number, then it's value will be returned.
	EnvDefV(v EnvVarDep, def *api.Val) *api.Val

	// Resp returns a Resp object that can be used to send a response to the
	// source of the current request.
	Resp() Resp

	// Req returns an empty Req object that can be used to send a request to the
	// given Component.
	Req(c ComponentDep) Req

	// Forward returns a Req object that can be used to send a request to the
	// given Component. The Req object is a clone of the current request.
	Forward(c ComponentDep) Req

	// HTTP returns a native Go http.Client. Any requests made with the client
	// are sent to the given Component. The target Component should be capable
	// of processing HTTP requests.
	HTTP(c ComponentDep) *http.Client

	// HTTP returns a native go http.RoundTripper. This is useful to integrate
	// with HTTP based libraries.
	Transport(c ComponentDep) http.RoundTripper

	// Context returns a context.Context with it's duration set to the TTL of
	// the current request.
	Context() context.Context

	// Log returns a pre-configured structured logger for the current request.
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

	Forward(resp core.EventReader) error
	SendStr(s ...string) error
	SendHTML(h string) error
	SendJSON(v any) error
	SendAccepts(json any, html, str string) error
	SendBytes(contentType string, b []byte) error
	SendReader(contentType string, reader io.Reader) error
	Send() error
}

type EnvVarDep interface {
	Name() string
	Type() api.EnvVarType
}

type ComponentDep interface {
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
