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
	"google.golang.org/protobuf/types/known/structpb"
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
	evt.SetTraceId(parent.TraceId())
	evt.SetSpanId(parent.SpanId())
	evt.SetTraceFlags(parent.TraceFlags())
}

func (evt *Event) EventType() EventType {
	return EventType(evt.Type)
}

func (evt *Event) Param(key string) string {
	return evt.ParamV(key).String()
}

func (evt *Event) ParamV(key string) *Val {
	v, _ := ValProto(evt.ParamProto(key))
	return v
}

func (evt *Event) ParamProto(key string) *structpb.Value {
	return evt.Params[key]
}

func (evt *Event) SetParam(key string, val string) {
	evt.SetParamProto(key, structpb.NewStringValue(val))
}

func (evt *Event) SetParamV(key string, val *Val) {
	evt.SetParamProto(key, val.Proto())
}

func (evt *Event) SetParamProto(key string, val *structpb.Value) {
	if val == nil {
		delete(evt.Params, key)
		return
	}
	evt.Params[key] = val
}

func (evt *Event) Value(key string) string {
	return evt.ValueV(key).String()
}

func (evt *Event) ValueV(key string) *Val {
	v, _ := ValProto(evt.ValueProto(key))
	return v
}

func (evt *Event) ValueMap(key string) map[string][]string {
	m := make(map[string][]string)
	p := evt.ValueProto(key).GetStructValue()
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

func (evt *Event) ValueProto(key string) *structpb.Value {
	return evt.Values[key]
}

func (evt *Event) SetValue(key string, val string) {
	evt.SetValueProto(key, structpb.NewStringValue(val))
}

func (evt *Event) SetValueV(key string, val *Val) {
	evt.SetValueProto(key, val.Proto())
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

func (evt *Event) TTL() time.Duration {
	return time.Microsecond * time.Duration(evt.Ttl)
}

func (evt *Event) ReduceTTL(start time.Time) {
	evt.Ttl = evt.Ttl - time.Since(start).Microseconds()
}

func (evt *Event) Status() int {
	return evt.ValueV(ValKeyStatusCode).Int()
}

func (evt *Event) StatusV() *Val {
	return evt.ValueV(ValKeyStatusCode)
}

func (evt *Event) SetStatus(code int) {
	evt.SetValueV(ValKeyStatusCode, ValInt(code))
}

func (evt *Event) SetStatusV(val *Val) {
	evt.SetValueV(ValKeyStatusCode, val)
}

func (evt *Event) TraceId() string {
	return evt.Value(ValKeyTraceId)
}

func (evt *Event) SetTraceId(val string) {
	evt.SetValue(ValKeyTraceId, val)
}

func (evt *Event) SpanId() string {
	return evt.Value(ValKeySpanId)
}

func (evt *Event) SetSpanId(val string) {
	evt.SetValue(ValKeySpanId, val)
}

func (evt *Event) TraceFlags() byte {
	return byte(evt.ValueV(ValKeyTraceFlags).Float())
}

func (evt *Event) SetTraceFlags(val byte) {
	evt.SetValueV(ValKeyTraceFlags, ValInt(int(val)))
}

func (evt *Event) SetJSON(v any) error {
	if v == nil {
		v = make(map[string]any)

	}

	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	evt.ContentType = fmt.Sprintf("%s; %s", ContentTypeJSON, CharSetUTF8)
	evt.Content = b

	return nil
}

func (evt *Event) Bind(v any) error {
	return evt.bind(v, false)
}

func (evt *Event) BindStrict(v any) error {
	return evt.bind(v, true)
}

func (evt *Event) bind(v any, strict bool) error {
	contType := strings.ToLower(evt.ContentType)
	switch {
	case strings.Contains(contType, ContentTypeJSON):
		dec := json.NewDecoder(bytes.NewReader(evt.Content))
		if strict {
			dec.DisallowUnknownFields()
		}
		return dec.Decode(v)
	default:
		return fmt.Errorf("%w: %s", ErrUnknownContentType, evt.ContentType)
	}
}

func (evt *Event) HTTPRequest(ctx context.Context) (*http.Request, error) {
	body := bytes.NewReader(evt.Content)
	req, err := http.NewRequestWithContext(ctx, evt.Value(ValKeyMethod), evt.Value(ValKeyURL), body)
	if err != nil {
		return nil, err
	}
	req.Header = evt.ValueMap(ValKeyHeader)

	return req, nil
}

func (evt *Event) HTTPResponse() *http.Response {
	code := evt.ValueV(ValKeyStatusCode).Int()
	if code == 0 {
		code = http.StatusOK
	}
	return &http.Response{
		Status:     evt.Value("status"),
		StatusCode: code,
		Header:     evt.ValueMap(ValKeyHeader),
		Body:       io.NopCloser(bytes.NewReader(evt.Content)),
	}
}

func (evt *Event) SetHTTPRequest(httpReq *http.Request) error {
	content, err := io.ReadAll(httpReq.Body)
	if err != nil {
		return err
	}
	evt.Content = content
	evt.ContentType = httpReq.Header.Get("Content-Type")
	evt.Category = Category_CATEGORY_REQUEST

	if evt.Environment == "" {
		evt.Environment = GetParamOrHeader(httpReq, HeaderEnv, HeaderShortEnv, HeaderAbbrvEnv)
	}

	if evt.Deployment == "" {
		evt.Deployment = GetParamOrHeader(httpReq, HeaderDep, HeaderShortDep, HeaderAbbrvDep)
	}

	if evt.Type == "" || evt.Type == string(EventTypeUnknown) {
		evtType := GetParamOrHeader(httpReq, HeaderEventType, HeaderAbbrvEventType)
		if evtType != "" {
			evt.Type = evtType
		} else {
			evt.Type = string(EventTypeHTTP)
		}
	}

	// TODO parse and validate jwt
	// NOTE the jwt lib has lots of deps :(
	// token, err := jwt.ParseRequest(httpReq, jwt.WithValidate(true))
	// if err == nil {
	// 	b, _ := jwt.NewSerializer().Serialize(token)
	// 	evt.SetValue("authToken", string(b))
	// }

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

	evt.SetValue(ValKeyURL, url.String())
	evt.SetValue(ValKeyHost, strings.ToLower(url.Host))
	evt.SetValue(ValKeyPath, url.Path)
	evt.SetValue(ValKeyMethod, httpReq.Method)
	evt.SetValueMap(ValKeyHeader, httpReq.Header)
	evt.SetValueMap(ValKeyQuery, url.Query())

	return nil
}

func (evt *Event) SetHTTPResponse(httpResp *http.Response) error {
	defer httpResp.Body.Close()
	content, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}
	evt.Content = content
	evt.ContentType = httpResp.Header.Get("Content-Type")
	evt.Category = Category_CATEGORY_RESPONSE

	if evt.Type == "" || evt.Type == string(EventTypeUnknown) {
		evt.Type = string(EventTypeHTTP)
	}

	evt.SetValue("status", httpResp.Status)
	evt.SetValueV(ValKeyStatusCode, ValInt(httpResp.StatusCode))
	evt.SetValueMap(ValKeyHeader, httpResp.Header)

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
