package grpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/lestrrat-go/jwx/jwt"
	"github.com/xigxog/kubefox/libs/core/api/common"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/structpb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

const (
	JSONContentType = "application/json"
)

var (
	ErrUnknownContentType = errors.New("unknown content type")
)

func (d *Data) SetType(dataType string) {
	d.Type = dataType
}

func (d *Data) SetParentId(id string) {
	d.ParentId = id
}

func (d *Data) SetContext(ctx *EventContext) {
	d.Context = ctx
}

func (d *Data) SetFabric(fabric *Fabric) {
	d.Fabric = fabric
}

func (d *Data) SetSpan(span *Span) {
	d.Span = span
}

func (d *Data) SetSource(source *Component) {
	d.Source = source
}

func (d *Data) SetTarget(target *Component) {
	d.Target = target
}

func (d *Data) SetContentType(contentType string) {
	d.ContentType = contentType
}

func (d *Data) SetContent(content []byte) {
	d.Content = content
}

func (d *Data) UpdateSpan(span trace.Span) {
	d.Span = &Span{
		TraceId:    span.SpanContext().TraceID().String(),
		SpanId:     span.SpanContext().SpanID().String(),
		TraceFlags: []byte{byte(span.SpanContext().TraceFlags())},
	}
}

func (d *Data) SetToken(token jwt.Token) {
	if token == nil {
		d.Token = nil
		return
	}

	if d.Token == nil {
		d.Token = &Token{}
	}

	claims, _ := structpb.NewStruct(token.PrivateClaims())

	d.Token.Id = token.JwtID()
	d.Token.Issuer = token.Issuer()
	d.Token.Subject = token.Subject()
	d.Token.Audience = token.Audience()
	d.Token.IssuedAt = timestamppb.New(token.IssuedAt())
	d.Token.NotBefore = timestamppb.New(token.NotBefore())
	d.Token.Expiration = timestamppb.New(token.Expiration())
	d.Token.PrivateClaims = claims.Fields
}

func (d *Data) GetArg(name string) string {
	return d.GetArgVar(name).String()
}

func (d *Data) GetArgVar(name string) *common.Var {
	v, _ := common.VarFromValue(d.GetArgProto(name))
	return v
}

func (d *Data) GetArgProto(name string) *structpb.Value {
	d.checkArgs()
	return d.Args[name]
}

func (d *Data) SetArg(k string, v string) {
	d.SetArgProto(k, structpb.NewStringValue(v))
}

func (d *Data) SetArgNumber(k string, v float64) {
	d.SetArgProto(k, structpb.NewNumberValue(v))
}

func (d *Data) SetArgProto(name string, val *structpb.Value) {
	d.checkArgs()
	if val == nil {
		delete(d.Args, name)
		return
	}
	d.Args[name] = val
}

func (d *Data) checkArgs() {
	if d.Args == nil {
		d.Args = make(map[string]*structpb.Value)
	}
}

func (d *Data) GetValue(name string) string {
	return d.GetValueVar(name).String()
}

func (d *Data) GetValueVar(name string) *common.Var {
	v, _ := common.VarFromValue(d.GetValueProto(name))
	return v
}

func (d *Data) GetValueProto(name string) *structpb.Value {
	d.checkValues()
	return d.Values[name]
}

func (d *Data) SetValue(k string, v string) {
	d.SetValueProto(k, structpb.NewStringValue(v))
}

func (d *Data) SetValueNumber(k string, v float64) {
	d.SetValueProto(k, structpb.NewNumberValue(v))
}

func (d *Data) SetValueProto(name string, val *structpb.Value) {
	d.checkValues()
	if val == nil {
		delete(d.Values, name)
		return
	}
	d.Values[name] = val
}

func (d *Data) checkValues() {
	if d.Values == nil {
		d.Values = make(map[string]*structpb.Value)
	}
}

func (d *Data) GetErrorMsg() string {
	return d.GetValue("error_msg")
}

func (d *Data) SetErrorMsg(v string) {
	d.SetValue("error_msg", v)
}

func (d *Data) Marshal(v any) error {
	if v == nil {
		v = make(map[string]any)

	}
	if d.ContentType == "" {
		d.ContentType = JSONContentType + "; charset=utf-8"
	}

	var content []byte
	var err error
	contType := strings.ToLower(d.ContentType)
	switch {
	case strings.Contains(contType, JSONContentType):
		content, err = json.Marshal(v)
	default:
		return ErrUnknownContentType
	}
	if err != nil {
		return err
	}

	d.Content = content

	return nil
}

func (d *Data) Unmarshal(v any) error {
	return d.unmarshal(v, false)
}

func (d *Data) UnmarshalStrict(v any) error {
	return d.unmarshal(v, true)
}

func (d *Data) unmarshal(v any, strict bool) error {
	contType := strings.ToLower(d.ContentType)
	switch {
	case strings.Contains(contType, JSONContentType):
		dec := json.NewDecoder(bytes.NewReader(d.Content))
		if strict {
			dec.DisallowUnknownFields()
		}
		return dec.Decode(v)
	default:
		return fmt.Errorf("%w: %s", ErrUnknownContentType, d.ContentType)
	}
}
