package kit

import (
	"context"
	"encoding/json"
	"time"

	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/logkf"
	"google.golang.org/protobuf/types/known/structpb"
)

type Kontext interface {
	context.Context

	//
	// Retrieve from request event.
	//

	Env(name string) string
	EnvVar(name string) *kubefox.Var
	EnvOrDefault(key string, def string) string

	Param(name string) string
	ParamVar(name string) *kubefox.Var
	ParamOrDefault(key string, def string) string

	Query(name string) string
	QueryVar(name string) *kubefox.Var
	QueryOrDefault(name string, def string) string

	Header(name string) string
	HeaderVar(name string) *kubefox.Var
	HeaderOrDefault(name string, def string) string

	Bind(v any) error

	//
	// Modify response event.
	//

	Status(code int)

	//
	// Send response event.
	//

	String(s string) error
	JSON(v any) error
	Bytes(contentType string, b []byte) error

	//
	// Send request event.
	//

	//
	// Other
	//
	Log() *logkf.Logger
}

type kontext struct {
	context.Context

	kitSvc *kit
	req    *kubefox.Event
	resp   *kubefox.Event
	env    map[string]*structpb.Value

	start int64

	log *logkf.Logger
}

func (k *kontext) Log() *logkf.Logger {
	return k.log
}

func (k *kontext) Env(key string) string {
	return k.EnvVar(key).String()
}

func (k *kontext) EnvVar(key string) *kubefox.Var {
	v, _ := kubefox.VarFromValue(k.env[key])
	return v
}

func (k *kontext) EnvOrDefault(key string, def string) string {
	if v := k.Env(key); v == "" {
		return def
	} else {
		return v
	}
}

func (k *kontext) Param(key string) string {
	v := k.req.GetParamVar(key)
	if !v.IsNil() {
		return v.String()
	}
	if s := k.Query(key); s != "" {
		return s
	}
	if s := k.Header(key); s != "" {
		return s
	}

	return ""
}

func (k *kontext) ParamVar(key string) *kubefox.Var {
	return k.req.GetParamVar(key)
}

func (k *kontext) ParamOrDefault(key string, def string) string {
	if v := k.Param(key); v == "" {
		return def
	} else {
		return v
	}
}

func (k *kontext) Query(key string) string {
	m := k.req.GetValueMap(kubefox.QueryValKey)
	if q := m[key]; len(q) >= 1 {
		return q[0]
	}

	return ""
}

func (k *kontext) QueryVar(key string) *kubefox.Var {
	return kubefox.NewVarString(k.Query(key))
}

func (k *kontext) QueryOrDefault(key string, def string) string {
	if v := k.Query(key); v == "" {
		return def
	} else {
		return v
	}
}

func (k *kontext) Header(key string) string {
	m := k.req.GetValueMap(kubefox.HeaderValKey)
	if h := m[key]; len(h) >= 1 {
		return h[0]
	}

	return ""
}

func (k *kontext) HeaderVar(key string) *kubefox.Var {
	return kubefox.NewVarString(k.Header(key))
}

func (k *kontext) HeaderOrDefault(key string, def string) string {
	if v := k.Header(key); v == "" {
		return def
	} else {
		return v
	}
}

func (k *kontext) Bind(v any) error {
	return k.req.Unmarshal(v)
}

func (k *kontext) Status(code int) {
	k.req.SetValueNumber(kubefox.StatusCodeValKey, float64(code))
}

func (k *kontext) String(s string) error {
	return k.Bytes("text/plain; charset=UTF-8", []byte(s))
}

func (k *kontext) JSON(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return k.Bytes("application/json; charset=UTF-8", b)
}

func (k *kontext) Bytes(contentType string, b []byte) error {
	k.resp.ContentType = contentType
	k.resp.Content = b

	k.resp.Ttl = k.req.Ttl - (time.Now().UnixMicro() - k.start)

	return k.kitSvc.sendEvent(k.resp, k.start)
}
