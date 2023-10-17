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
	*kubefox.ActiveEvent

	kit  *kit
	resp *kubefox.ActiveEvent
	env  map[string]*structpb.Value

	ctx context.Context
	log *logkf.Logger
}

type reqKtx struct {
	*kubefox.ActiveEvent

	kit *kit
}

type respKtx struct {
	*kubefox.ActiveEvent

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
	return &respKtx{
		ActiveEvent: k.resp,
		kit:         k.kit,
	}
}

func (resp *respKtx) SendString(val string) error {
	c := fmt.Sprintf("%s; %s", kubefox.ContentTypePlain, kubefox.CharSetUTF8)
	return resp.SendBytes(c, []byte(val))
}

func (resp *respKtx) SendHTML(val string) error {
	c := fmt.Sprintf("%s; %s", kubefox.ContentTypeHTML, kubefox.CharSetUTF8)
	return resp.SendBytes(c, []byte(val))
}

func (resp *respKtx) SendJSON(val any) error {
	if err := resp.SetJSON(val); err != nil {
		return err
	}

	return resp.Send()
}

func (resp *respKtx) SendBytes(contentType string, b []byte) error {
	resp.ActiveEvent.ContentType = contentType
	resp.ActiveEvent.Content = b

	return resp.Send()
}

func (resp *respKtx) Send() error {
	return resp.kit.sendEvent(resp.ActiveEvent)
}

func (k *kontext) Component(component string) Req {
	return &reqKtx{
		ActiveEvent: kubefox.StartReq(kubefox.EventOpts{
			Type:   kubefox.EventTypeComponent,
			Parent: k.Event,
			Source: k.kit.comp,
			Target: &kubefox.Component{Name: component},
		}),
		kit: k.kit,
	}
}

func (req *reqKtx) SendString(val string) (kubefox.EventReader, error) {
	c := fmt.Sprintf("%s; %s", kubefox.ContentTypePlain, kubefox.CharSetUTF8)
	return req.SendBytes(c, []byte(val))
}

func (req *reqKtx) SendHTML(val string) (kubefox.EventReader, error) {
	c := fmt.Sprintf("%s; %s", kubefox.ContentTypeHTML, kubefox.CharSetUTF8)
	return req.SendBytes(c, []byte(val))
}

func (req *reqKtx) SendJSON(val any) (kubefox.EventReader, error) {
	if err := req.SetJSON(val); err != nil {
		return nil, err
	}

	return req.Send()
}

func (req *reqKtx) SendBytes(contentType string, b []byte) (kubefox.EventReader, error) {
	req.ActiveEvent.ContentType = contentType
	req.ActiveEvent.Content = b

	return req.Send()
}

func (req *reqKtx) Send() (kubefox.EventReader, error) {
	resp, err := req.kit.sendReq(req.ActiveEvent)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
