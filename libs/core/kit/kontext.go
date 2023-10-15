package kit

import (
	"context"
	"time"

	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/logkf"
	"google.golang.org/protobuf/types/known/structpb"
)

type Kontext interface {
	context.Context
	kubefox.EventReader

	Env(name string) string
	EnvV(name string) *kubefox.Val
	EnvDef(name string, def string) string

	Resp() Resp

	Component(name string) Req

	Log() *logkf.Logger
}

type Req interface {
	kubefox.EventWriter

	SendString(s string) (kubefox.EventReader, error)
	SendHTML(s string) (kubefox.EventReader, error)
	SendJSON(v any) (kubefox.EventReader, error)
	SendBytes(contentType string, b []byte) (kubefox.EventReader, error)
	Send() (kubefox.EventReader, error)
}

type Resp interface {
	kubefox.EventWriter

	SendString(s string) error
	SendHTML(s string) error
	SendJSON(v any) error
	SendBytes(contentType string, b []byte) error
	Send() error
}

type kontext struct {
	context.Context
	*kubefox.EventRW

	kit  *kit
	resp *kubefox.EventRW
	env  map[string]*structpb.Value

	start time.Time

	log *logkf.Logger
}

type reqKtx struct {
	*kubefox.EventRW
	*kontext
}

type respKtx struct {
	*kubefox.EventRW
	*kontext
}

func (k *kontext) Log() *logkf.Logger {
	return k.log
}

func (k *kontext) Env(key string) string {
	return k.EnvV(key).String()
}

func (k *kontext) EnvV(key string) *kubefox.Val {
	v, _ := kubefox.ValProto(k.env[key])
	return v
}

func (k *kontext) EnvDef(key string, def string) string {
	if v := k.Env(key); v == "" {
		return def
	} else {
		return v
	}
}

func (k *kontext) Resp() Resp {
	return &respKtx{
		kontext: k,
		EventRW: k.resp,
	}
}

func (s *respKtx) SendString(val string) error {
	return s.SendBytes("text/plain; charset=UTF-8", []byte(val))
}

func (s *respKtx) SendHTML(val string) error {
	return s.SendBytes("text/html; charset=UTF-8", []byte(val))
}

func (s *respKtx) SendJSON(val any) error {
	if err := s.SetJSON(val); err != nil {
		return err
	}

	return s.Send()
}

func (s *respKtx) SendBytes(contentType string, b []byte) error {
	s.EventRW.ContentType = contentType
	s.EventRW.Content = b

	return s.Send()
}

func (s *respKtx) Send() error {
	s.Event.ReduceTTL(s.start)
	return s.kit.sendEvent(s.Event)
}

func (k *kontext) Component(component string) Req {
	return &reqKtx{
		kontext: k,
		EventRW: kubefox.NewReq(
			kubefox.ComponentRequestType,
			k.Event,
			k.kit.comp,
			&kubefox.Component{
				Name: component,
			},
		),
	}
}

func (s *reqKtx) SendString(val string) (kubefox.EventReader, error) {
	return s.SendBytes("text/plain; charset=UTF-8", []byte(val))
}

func (s *reqKtx) SendHTML(val string) (kubefox.EventReader, error) {
	return s.SendBytes("text/html; charset=UTF-8", []byte(val))
}

func (s *reqKtx) SendJSON(val any) (kubefox.EventReader, error) {
	if err := s.SetJSON(val); err != nil {
		return nil, err
	}

	return s.Send()
}

func (s *reqKtx) SendBytes(contentType string, b []byte) (kubefox.EventReader, error) {
	s.EventRW.ContentType = contentType
	s.EventRW.Content = b

	return s.Send()
}

func (s *reqKtx) Send() (kubefox.EventReader, error) {
	s.Event.ReduceTTL(s.start)
	resp, err := s.kit.sendReq(s.kontext, s.Event)
	if err != nil {
		return nil, err
	}

	return kubefox.NewEventRW(resp), nil
}
