package kit

import (
	"context"
	"fmt"
	"io"
	"net/http"

	kubefox "github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/grpc"
	"github.com/xigxog/kubefox/logkf"

	"google.golang.org/protobuf/types/known/structpb"
)

type kontext struct {
	*kubefox.Event

	kit  *kit
	resp *kubefox.Event
	env  map[string]*structpb.Value

	ctx context.Context
	log *logkf.Logger
}

type respKontext struct {
	*kubefox.Event

	brk *grpc.Client
}

type reqKontext struct {
	*kubefox.Event

	brk *grpc.Client
	ctx context.Context
}

func (k *kontext) Context() context.Context {
	return k.ctx
}

func (k *kontext) Log() *logkf.Logger {
	return k.log
}

func (k *kontext) Env(v EnvVar) string {
	return k.EnvV(v).String()
}

func (k *kontext) EnvV(v EnvVar) *kubefox.Val {
	val, _ := kubefox.ValProto(k.env[v.GetName()])
	return val
}

func (k *kontext) EnvDef(v EnvVar, def string) string {
	if val := k.Env(v); val == "" {
		return def
	} else {
		return val
	}
}

func (k *kontext) Resp() Resp {
	return &respKontext{
		Event: k.resp,
		brk:   k.kit.brk,
	}
}

func (k *kontext) Req(c Dependency) Req {
	return &reqKontext{
		Event: kubefox.NewReq(kubefox.EventOpts{
			Type:   c.GetEventType(),
			Parent: k.Event,
			Source: k.kit.brk.Component,
			Target: &kubefox.Component{Name: c.GetName()},
		}),
		brk: k.kit.brk,
		ctx: k.ctx,
	}
}

func (k *kontext) HTTP(c Dependency) *http.Client {
	return &http.Client{
		Transport: k.Transport(c),
	}
}

func (k *kontext) Transport(c Dependency) http.RoundTripper {
	return &EventRoundTripper{
		req: &reqKontext{
			Event: kubefox.NewReq(kubefox.EventOpts{
				Type:   c.GetEventType(),
				Parent: k.Event,
				Source: k.kit.brk.Component,
				Target: &kubefox.Component{Name: c.GetName()},
			}),
			brk: k.kit.brk,
			ctx: k.ctx,
		},
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

func (resp *respKontext) SendReader(contentType string, reader io.Reader) error {
	bytes, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	return resp.SendBytes(contentType, bytes)
}

func (resp *respKontext) SendBytes(contentType string, b []byte) error {
	resp.Event.ContentType = contentType
	resp.Event.Content = b

	return resp.Send()
}

func (resp *respKontext) Send() error {
	return resp.brk.SendResp(resp.Event)
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

func (req *reqKontext) SendReader(contentType string, reader io.Reader) (kubefox.EventReader, error) {
	bytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return req.SendBytes(contentType, bytes)
}

func (req *reqKontext) SendBytes(contentType string, b []byte) (kubefox.EventReader, error) {
	req.Event.ContentType = contentType
	req.Event.Content = b

	return req.Send()
}

func (req *reqKontext) Send() (kubefox.EventReader, error) {
	resp, err := req.brk.SendReq(req.ctx, req.Event)
	if err != nil {
		return nil, err
	}

	return resp, nil
}