package kubefox

import (
	"bytes"
	context "context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	ErrUnknownContentType = errors.New("unknown content type")
)

type EventReader interface {
	EventType() EventType

	Param(key string) string
	ParamV(key string) *Val
	ParamDef(key string, def string) string

	Query(key string) string
	QueryV(key string) *Val
	QueryDef(key string, def string) string
	QueryAll(key string) []string

	Header(key string) string
	HeaderV(key string) *Val
	HeaderDef(key string, def string) string
	HeaderAll(key string) []string

	Status() int
	StatusV() *Val

	Bind(v any) error
	Str() string
	Bytes() []byte
}

type EventWriter interface {
	EventReader

	SetParam(key, value string) EventWriter
	SetParamV(key string, value *Val) EventWriter

	SetQuery(key, value string) EventWriter
	SetQueryV(key string, value *Val) EventWriter
	DelQuery(key string) EventWriter

	SetHeader(key, value string) EventWriter
	SetHeaderV(key string, value *Val) EventWriter
	AddHeader(key, value string) EventWriter
	DelHeader(key string) EventWriter

	SetStatus(code int) EventWriter
	SetStatusV(val *Val) EventWriter
}

type EventOpts struct {
	Type   EventType
	Parent *Event
	Source *Component
	Target *Component
}

func NewResp(opts EventOpts) *Event {
	return newEvent(Category_RESPONSE, opts)
}

func NewReq(opts EventOpts) *Event {
	return newEvent(Category_REQUEST, opts)
}

func NewMsg(opts EventOpts) *Event {
	return newEvent(Category_MESSAGE, opts)
}

func NewEvent() *Event {
	return &Event{
		Id:         uuid.NewString(),
		CreateTime: time.Now().UnixMicro(),
		Params:     make(map[string]*structpb.Value),
		Values:     make(map[string]*structpb.Value),
		Context:    &EventContext{},
	}
}

func newEvent(cat Category, opts EventOpts) *Event {
	evt := NewEvent()
	evt.SetParent(opts.Parent)
	evt.Category = cat
	evt.Source = opts.Source
	evt.Target = opts.Target
	if opts.Type != EventTypeUnknown {
		evt.Type = string(opts.Type)
	} else if opts.Parent != nil {
		evt.Type = opts.Parent.Type
	}

	return evt
}

func (evt *Event) SetParent(parent *Event) {
	if parent == nil {
		return
	}
	evt.ParentId = parent.Id
	evt.Ttl = parent.Ttl
	evt.SetContext(parent.Context)
	evt.SetTraceId(parent.TraceId())
	evt.SetSpanId(parent.SpanId())
	evt.SetTraceFlags(parent.TraceFlags())
	if evt.Type == "" {
		evt.Type = parent.Type
	}
}

func (evt *Event) SetContext(evtCtx *EventContext) {
	if evtCtx == nil {
		return
	}
	evt.Context.Deployment = evtCtx.Deployment
	evt.Context.Environment = evtCtx.Environment
	evt.Context.Release = evtCtx.Release
}

func (evt *Event) EventType() EventType {
	return EventType(evt.Type)
}

func (evt *Event) Param(key string) string {
	return evt.ParamV(key).String()
}

func (evt *Event) ParamV(key string) *Val {
	v, _ := ValProto(evt.ParamProto(key))
	if !v.IsNil() {
		return v
	}
	if s := evt.Query(key); s != "" {
		return ValString(s)
	}
	if s := evt.Header(key); s != "" {
		return ValString(s)
	}

	return v
}

func (evt *Event) ParamDef(key string, def string) string {
	if v := evt.Param(key); v == "" {
		return def
	} else {
		return v
	}
}

func (evt *Event) ParamProto(key string) *structpb.Value {
	return evt.Params[key]
}

func (evt *Event) SetParam(key string, val string) EventWriter {
	evt.SetParamProto(key, structpb.NewStringValue(val))
	return evt
}

func (evt *Event) SetParamV(key string, val *Val) EventWriter {
	evt.SetParamProto(key, val.Proto())
	return evt
}

func (evt *Event) SetParamProto(key string, val *structpb.Value) EventWriter {
	if val == nil {
		delete(evt.Params, key)
	} else {
		evt.Params[key] = val
	}
	return evt
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

func (evt *Event) ValueMapKey(valKey, key string) string {
	s := evt.ValueMapKeyAll(valKey, key)
	if len(s) == 0 {
		return ""
	}
	return s[0]
}

func (evt *Event) ValueMapKeyAll(valKey, key string) []string {
	s := []string{}
	m := evt.ValueProto(valKey).GetStructValue()
	if m == nil {
		return s
	}

	f := m.Fields[key]
	if f == nil || f.GetListValue() == nil {
		return s
	}

	for _, v := range f.GetListValue().Values {
		s = append(s, v.GetStringValue())
	}

	return s
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

func (evt *Event) SetValueMapKey(valKey, key, value string, overwrite bool) {
	m := evt.ValueProto(valKey).GetStructValue()
	if m == nil {
		v, _ := structpb.NewValue(map[string]interface{}{
			key: value,
		})
		evt.SetValueProto(valKey, v)
		return
	}

	f := m.Fields[key]
	if overwrite || f == nil || f.GetListValue() == nil {
		v, _ := structpb.NewValue([]any{value})
		m.Fields[key] = v
		return
	}

	v, _ := structpb.NewValue(value)
	f.GetListValue().Values = append(f.GetListValue().Values, v)
}

func (evt *Event) DelValueMapKey(valKey, key string) {
	m := evt.ValueProto(valKey).GetStructValue()
	if m == nil {
		return
	}
	delete(m.Fields, key)
}

func (evt *Event) SetValueProto(key string, val *structpb.Value) {
	if val == nil {
		delete(evt.Values, key)
		return
	}
	evt.Values[key] = val
}

func (evt *Event) TTL() time.Duration {
	return time.Duration(evt.Ttl) * time.Microsecond
}

func (evt *Event) ReduceTTL(start time.Time) time.Duration {
	evt.Ttl = evt.Ttl - time.Since(start).Microseconds()
	return evt.TTL()
}

func (evt *Event) Status() int {
	return evt.ValueV(ValKeyStatusCode).Int()
}

func (evt *Event) StatusV() *Val {
	return evt.ValueV(ValKeyStatusCode)
}

func (evt *Event) SetStatus(code int) EventWriter {
	evt.SetValueV(ValKeyStatusCode, ValInt(code))
	return evt
}

func (evt *Event) SetStatusV(val *Val) EventWriter {
	evt.SetValueV(ValKeyStatusCode, val)
	return evt
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

func (evt *Event) Query(key string) string {
	return evt.ValueMapKey(ValKeyQuery, key)
}

func (evt *Event) QueryV(key string) *Val {
	return ValString(evt.Query(key))
}

func (evt *Event) QueryDef(key string, def string) string {
	if v := evt.Query(key); v == "" {
		return def
	} else {
		return v
	}
}

func (evt *Event) QueryAll(key string) []string {
	return evt.ValueMapKeyAll(ValKeyQuery, key)
}

func (evt *Event) SetQuery(key, value string) EventWriter {
	evt.SetValueMapKey(ValKeyQuery, key, value, true)
	return evt
}

func (evt *Event) SetQueryV(key string, value *Val) EventWriter {
	return evt.SetQuery(key, value.String())
}

func (evt *Event) DelQuery(key string) EventWriter {
	evt.DelValueMapKey(ValKeyQuery, key)
	return evt
}

func (evt *Event) Header(key string) string {
	key = textproto.CanonicalMIMEHeaderKey(key)
	return evt.ValueMapKey(ValKeyHeader, key)
}

func (evt *Event) HeaderV(key string) *Val {
	return ValString(evt.Header(key))
}

func (evt *Event) HeaderDef(key string, def string) string {
	if v := evt.Header(key); v == "" {
		return def
	} else {
		return v
	}
}

func (evt *Event) HeaderAll(key string) []string {
	key = textproto.CanonicalMIMEHeaderKey(key)
	return evt.ValueMapKeyAll(ValKeyHeader, key)
}

func (evt *Event) SetHeader(key, value string) EventWriter {
	key = textproto.CanonicalMIMEHeaderKey(key)
	evt.SetValueMapKey(ValKeyHeader, key, value, true)
	return evt
}

func (evt *Event) AddHeader(key, value string) EventWriter {
	key = textproto.CanonicalMIMEHeaderKey(key)
	evt.SetValueMapKey(ValKeyHeader, key, value, false)
	return evt
}

func (evt *Event) SetHeaderV(key string, value *Val) EventWriter {
	return evt.SetHeader(key, value.String())
}

func (evt *Event) DelHeader(key string) EventWriter {
	key = textproto.CanonicalMIMEHeaderKey(key)
	evt.DelValueMapKey(ValKeyHeader, key)
	return evt
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

func (evt *Event) Str() string {
	return string(evt.Bytes())
}

func (evt *Event) Bytes() []byte {
	return evt.Content
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
	evt.Category = Category_REQUEST

	if evt.Context.Environment == "" {
		evt.Context.Environment = GetParamOrHeader(httpReq, HeaderEnv, HeaderAbbrvEnv, HeaderShortEnv)
	}
	if evt.Context.Deployment == "" {
		evt.Context.Deployment = GetParamOrHeader(httpReq, HeaderDep, HeaderAbbrvDep, HeaderShortDep)
	}

	if evt.Type == "" || evt.Type == string(EventTypeUnknown) {
		evtType := GetParamOrHeader(httpReq, HeaderEventType, HeaderAbbrvEventType, HeaderShortEventType)
		if evtType != "" {
			evt.Type = evtType
		} else {
			evt.Type = string(EventTypeHTTP)
		}
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
	evt.Category = Category_RESPONSE

	if evt.Type == "" || evt.Type == string(EventTypeUnknown) {
		evt.Type = string(EventTypeHTTP)
	}

	evt.SetValue("status", httpResp.Status)
	evt.SetValueV(ValKeyStatusCode, ValInt(httpResp.StatusCode))
	evt.SetValueMap(ValKeyHeader, httpResp.Header)

	return nil
}

func (ctx *EventContext) IsRelease() bool {
	return ctx.Release != "" || (ctx.Deployment == "" && ctx.Environment == "")
}

func (ctx *EventContext) IsDeployment() bool {
	return ctx.Deployment != "" && ctx.Environment != ""
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
