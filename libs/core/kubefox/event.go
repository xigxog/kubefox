package kubefox

import (
	"bytes"
	context "context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwt"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	JSONContentType = "application/json"
)

var (
	ErrUnknownContentType = errors.New("unknown content type")
)

func NewEvent() *Event {
	return &Event{
		Id:         uuid.NewString(),
		CreateTime: time.Now().UnixMicro(),
		Params:     make(map[string]*structpb.Value),
		Values:     make(map[string]*structpb.Value),
	}
}

func (evt *Event) SetParent(parent *Event) {
	evt.ParentId = parent.Id
	evt.Ttl = parent.Ttl
	evt.Deployment = parent.Deployment
	evt.Environment = parent.Environment
	evt.Release = parent.Release
	evt.SetTraceId(parent.GetTraceId())
	evt.SetSpanId(parent.GetSpanId())
	evt.SetTraceFlags(parent.GetTraceFlags())
}

func (evt *Event) GetParam(key string) string {
	return evt.GetParamVar(key).String()
}

func (evt *Event) GetParamVar(key string) *Var {
	v, _ := VarFromValue(evt.GetParamProto(key))
	return v
}

func (evt *Event) GetParamProto(key string) *structpb.Value {
	return evt.Params[key]
}

func (evt *Event) SetParam(key string, val string) {
	evt.SetParamProto(key, structpb.NewStringValue(val))
}

func (evt *Event) SetParamVar(key string, val *Var) {
	evt.SetParamProto(key, val.Value())
}

func (evt *Event) SetParamNumber(key string, val float64) {
	evt.SetParamProto(key, structpb.NewNumberValue(val))
}

func (evt *Event) SetParamProto(key string, val *structpb.Value) {
	if val == nil {
		delete(evt.Params, key)
		return
	}
	evt.Params[key] = val
}

func (evt *Event) GetValue(key string) string {
	return evt.GetValueVar(key).String()
}

func (evt *Event) GetValueVar(key string) *Var {
	v, _ := VarFromValue(evt.GetValueProto(key))
	return v
}

func (evt *Event) GetValueMap(key string) map[string][]string {
	m := make(map[string][]string)
	p := evt.GetValueProto(key).GetStructValue()
	if p == nil {
		return m
	}

	for k, v := range p.Fields {
		for _, h := range v.GetListValue().Values {
			m[k] = append(m[k], h.GetStringValue())
		}
	}

	return m
}

func (evt *Event) GetValueProto(key string) *structpb.Value {
	return evt.Values[key]
}

func (evt *Event) SetValue(key string, val string) {
	evt.SetValueProto(key, structpb.NewStringValue(val))
}

func (evt *Event) SetValueVar(key string, val *Var) {
	evt.SetValueProto(key, val.Value())
}

func (evt *Event) SetValueNumber(key string, val float64) {
	evt.SetValueProto(key, structpb.NewNumberValue(val))
}

func (evt *Event) SetValueMap(key string, m map[string][]string) {
	mapAny := make(map[string]interface{}, len(m))
	for k, v := range m {
		l := make([]interface{}, len(v))
		for i, h := range v {
			l[i] = h
		}
		mapAny[k] = l
	}

	v, _ := structpb.NewValue(mapAny)
	evt.SetValueProto(key, v)
}

func (evt *Event) SetValueProto(key string, val *structpb.Value) {
	if val == nil {
		delete(evt.Values, key)
		return
	}
	evt.Values[key] = val
}

func (evt *Event) GetTraceId() string {
	return evt.GetValue("traceId")
}

func (evt *Event) SetTraceId(val string) {
	evt.SetValue("traceId", val)
}

func (evt *Event) GetSpanId() string {
	return evt.GetValue("spanId")
}

func (evt *Event) SetSpanId(val string) {
	evt.SetValue("spanId", val)
}

func (evt *Event) GetTraceFlags() byte {
	return byte(evt.GetValueVar("traceFlags").Float())
}

func (evt *Event) SetTraceFlags(val byte) {
	evt.SetValueNumber("traceFlags", float64(val))
}

func (evt *Event) Marshal(v any) error {
	if v == nil {
		v = make(map[string]any)

	}
	if evt.ContentType == "" {
		evt.ContentType = JSONContentType + "; charset=utf-8"
	}

	var content []byte
	var err error
	contType := strings.ToLower(evt.ContentType)
	switch {
	case strings.Contains(contType, JSONContentType):
		content, err = json.Marshal(v)
	default:
		return ErrUnknownContentType
	}
	if err != nil {
		return err
	}

	evt.Content = content

	return nil
}

func (evt *Event) Unmarshal(v any) error {
	return evt.unmarshal(v, false)
}

func (evt *Event) UnmarshalStrict(v any) error {
	return evt.unmarshal(v, true)
}

func (evt *Event) unmarshal(v any, strict bool) error {
	contType := strings.ToLower(evt.ContentType)
	switch {
	case strings.Contains(contType, JSONContentType):
		dec := json.NewDecoder(bytes.NewReader(evt.Content))
		if strict {
			dec.DisallowUnknownFields()
		}
		return dec.Decode(v)
	default:
		return fmt.Errorf("%w: %s", ErrUnknownContentType, evt.ContentType)
	}
}

func (evt *Event) ToHTTPRequest(ctx context.Context) (*http.Request, error) {
	body := bytes.NewReader(evt.Content)
	req, err := http.NewRequestWithContext(ctx, evt.GetValue(MethodValKey), evt.GetValue(URLValKey), body)
	if err != nil {
		return nil, err
	}
	req.Header = evt.GetValueMap(HeaderValKey)

	return req, nil
}

func (evt *Event) ToHTTPResponse() *http.Response {
	code := evt.GetValueVar(StatusCodeValKey).Int()
	if code == 0 {
		code = http.StatusOK
	}
	return &http.Response{
		Status:     evt.GetValue("status"),
		StatusCode: code,
		Header:     evt.GetValueMap(HeaderValKey),
		Body:       io.NopCloser(bytes.NewReader(evt.Content)),
	}
}

func (evt *Event) ParseHTTPRequest(httpReq *http.Request) error {
	content, err := io.ReadAll(httpReq.Body)
	if err != nil {
		return err
	}
	evt.Content = content
	evt.ContentType = httpReq.Header.Get("Content-Type")

	if evt.Environment == "" {
		evt.Environment = GetParamOrHeader(httpReq, EnvHeader, EnvHeaderShort, EnvHeaderAbbrv)
	}

	if evt.Deployment == "" {
		evt.Deployment = GetParamOrHeader(httpReq, DepHeader, DepHeaderShort, DepHeaderAbbrv)
	}

	if evt.Type == "" || evt.Type == string(UnknownEventType) {
		evtType := GetParamOrHeader(httpReq, EventTypeHeader, EventTypeHeaderAbbrv)
		if evtType != "" {
			evt.Type = evtType
		} else {
			evt.Type = string(HTTPRequestType)
		}
	}

	token, err := jwt.ParseRequest(httpReq, jwt.WithValidate(true))
	if err == nil {
		b, _ := jwt.NewSerializer().Serialize(token)
		evt.SetValue("authToken", string(b))
	}

	url := httpReq.URL
	if host := httpReq.Header.Get("X-Forwarded-Host"); host != "" {
		if port := httpReq.Header.Get("X-Forwarded-Port"); port != "" {
			url.Host = fmt.Sprintf("%s:%s", host, port)
		} else {
			url.Host = host
		}
		if scheme := httpReq.Header.Get("X-Forwarded-Proto"); scheme != "" {
			url.Scheme = scheme
		}
	}
	if url.Scheme == "" {
		url.Scheme = "http"
		if httpReq.TLS != nil {
			url.Scheme = "https"
		}
	}
	if url.Host == "" {
		url.Host = httpReq.Host
	}

	evt.SetValue(URLValKey, url.String())
	evt.SetValue(HostValKey, strings.ToLower(url.Host))
	evt.SetValue(PathValKey, url.Path)
	evt.SetValue(MethodValKey, httpReq.Method)
	evt.SetValueMap(HeaderValKey, httpReq.Header)
	evt.SetValueMap(QueryValKey, url.Query())

	return nil
}

func (evt *Event) ParseHTTPResponse(httpResp *http.Response) error {
	defer httpResp.Body.Close()
	content, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}
	evt.Content = content
	evt.ContentType = httpResp.Header.Get("Content-Type")

	if evt.Type == "" || evt.Type == string(UnknownEventType) {
		evt.Type = string(HTTPResponseType)
	}

	evt.SetValue("status", httpResp.Status)
	evt.SetValueNumber(StatusCodeValKey, float64(httpResp.StatusCode))
	evt.SetValueMap(HeaderValKey, httpResp.Header)

	return nil
}

// GetParamOrHeader looks for query parameters and headers for the provided
// keys. Keys are checked in order. Query parameters take precedence over
// headers.
func GetParamOrHeader(httpReq *http.Request, keys ...string) string {
	for _, key := range keys {
		val := httpReq.URL.Query().Get(strings.ToLower(key))
		if val == "" {
			val = httpReq.Header.Get(key)
		}
		if val != "" {
			return val
		}
	}

	return ""
}
