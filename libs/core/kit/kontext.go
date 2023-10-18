package kit

import (
	"context"
	"fmt"

	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/logkf"
	"google.golang.org/protobuf/types/known/structpb"
)

type Kontext interface {
	kubefox.EventReader

	Env(name string) string
	EnvV(name string) *kubefox.Val
	EnvDef(name string, def string) string

	Resp() Resp

	Component(name string) Req

	Context() context.Context
	Log() *logkf.Logger
}

type Req interface {
	kubefox.EventWriter

	SendStr(s string) (kubefox.EventReader, error)
	SendHTML(h string) (kubefox.EventReader, error)
	SendJSON(v any) (kubefox.EventReader, error)
	SendBytes(contentType string, b []byte) (kubefox.EventReader, error)
	Send() (kubefox.EventReader, error)
}

type Resp interface {
	kubefox.EventWriter

	SendStr(s string) error
	SendHTML(h string) error
	SendJSON(v any) error
	SendBytes(contentType string, b []byte) error
	Send() error
}

type kontext struct {
	*kubefox.Event

	kit  *kit
	resp *kubefox.Event
	env  map[string]*structpb.Value

	ctx context.Context
	log *logkf.Logger
}

type reqKontext struct {
	*kubefox.Event

	kit *kit
	ctx context.Context
}

type respKontext struct {
	*kubefox.Event

	kit *kit
}

func (k *kontext) Context() context.Context {
	return k.ctx
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
	return &respKontext{
		Event: k.resp,
		kit:   k.kit,
	}
}

func (resp *respKontext) SendStr(val string) error {
	c := fmt.Sprintf("%s; %s", kubefox.ContentTypePlain, kubefox.CharSetUTF8)
	return resp.SendBytes(c, []byte(val))
}

func (resp *respKontext) SendHTML(val string) error {
	c := fmt.Sprintf("%s; %s", kubefox.ContentTypeHTML, kubefox.CharSetUTF8)
	return resp.SendBytes(c, []byte(val))
}

func (resp *respKontext) SendJSON(val any) error {
	if err := resp.SetJSON(val); err != nil {
		return err
	}

	return resp.Send()
}

func (resp *respKontext) SendBytes(contentType string, b []byte) error {
	resp.Event.ContentType = contentType
	resp.Event.Content = b

	return resp.Send()
}

func (resp *respKontext) Send() error {
	return resp.kit.sendEvent(resp.Event)
}

func (k *kontext) Component(component string) Req {
	return &reqKontext{
		Event: kubefox.NewReq(kubefox.EventOpts{
			Type:   kubefox.EventTypeComponent,
			Parent: k.Event,
			Source: k.kit.comp,
			Target: &kubefox.Component{Name: component},
		}),
		kit: k.kit,
		ctx: k.ctx,
	}
}

func (req *reqKontext) SendStr(val string) (kubefox.EventReader, error) {
	c := fmt.Sprintf("%s; %s", kubefox.ContentTypePlain, kubefox.CharSetUTF8)
	return req.SendBytes(c, []byte(val))
}

func (req *reqKontext) SendHTML(val string) (kubefox.EventReader, error) {
	c := fmt.Sprintf("%s; %s", kubefox.ContentTypeHTML, kubefox.CharSetUTF8)
	return req.SendBytes(c, []byte(val))
}

func (req *reqKontext) SendJSON(val any) (kubefox.EventReader, error) {
	if err := req.SetJSON(val); err != nil {
		return nil, err
	}

	return req.Send()
}

func (req *reqKontext) SendBytes(contentType string, b []byte) (kubefox.EventReader, error) {
	req.Event.ContentType = contentType
	req.Event.Content = b

	return req.Send()
}

func (req *reqKontext) Send() (kubefox.EventReader, error) {
	resp, err := req.kit.sendReq(req.ctx, req.Event)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
