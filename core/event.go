package core

import (
	"bytes"
	context "context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/xigxog/kubefox/api"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type EventReader interface {
	EventType() api.EventType

	Param(key string) string
	ParamV(key string) *api.Val
	ParamDef(key string, def string) string

	URL() (*url.URL, error)

	Query(key string) string
	QueryV(key string) *api.Val
	QueryDef(key string, def string) string
	QueryAll(key string) []string

	Header(key string) string
	HeaderV(key string) *api.Val
	HeaderDef(key string, def string) string
	HeaderAll(key string) []string

	Status() int
	StatusV() *api.Val

	Bind(v any) error
	Str() string
	Bytes() []byte
}

type EventWriter interface {
	EventReader

	SetParam(key, value string) EventWriter
	SetParamV(key string, value *api.Val) EventWriter

	SetURL(u *url.URL) EventWriter
	TrimPathPrefix(prefix string) EventWriter

	SetQuery(key, value string) EventWriter
	SetQueryV(key string, value *api.Val) EventWriter
	DelQuery(key string) EventWriter

	SetHeader(key, value string) EventWriter
	SetHeaderV(key string, value *api.Val) EventWriter
	AddHeader(key, value string) EventWriter
	DelHeader(key string) EventWriter

	SetStatus(code int) EventWriter
	SetStatusV(val *api.Val) EventWriter
}

type EventOpts struct {
	Type   api.EventType
	Parent *Event
	Source *Component
	Target *Component
}

func NewResp(opts EventOpts) *Event {
	return applyOpts(NewEvent(), Category_RESPONSE, opts)
}

func NewReq(opts EventOpts) *Event {
	return applyOpts(NewEvent(), Category_REQUEST, opts)
}

func NewMsg(opts EventOpts) *Event {
	return applyOpts(NewEvent(), Category_MESSAGE, opts)
}

func NewErr(err error, opts EventOpts) *Event {
	opts.Type = api.EventTypeError

	evt := NewEvent()
	kfErr := &Err{}
	if ok := errors.As(err, &kfErr); !ok {
		kfErr = ErrUnexpected(err)
	}
	evt.SetJSON(kfErr)

	return applyOpts(evt, Category_RESPONSE, opts)
}

func CloneToResp(evt *Event, opts EventOpts) *Event {
	return clone(evt, Category_RESPONSE, opts)
}

func CloneToReq(evt *Event, opts EventOpts) *Event {
	return clone(evt, Category_REQUEST, opts)
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

func clone(evt *Event, cat Category, opts EventOpts) *Event {
	clone := proto.Clone(evt).(*Event)
	clone.Id = uuid.NewString()
	clone.CreateTime = time.Now().UnixMicro()
	if clone.Params == nil {
		clone.Params = make(map[string]*structpb.Value)
	}
	if clone.Values == nil {
		clone.Values = make(map[string]*structpb.Value)
	}
	if clone.Context == nil {
		clone.Context = &EventContext{}
	}
	return applyOpts(clone, cat, opts)
}

func applyOpts(evt *Event, cat Category, opts EventOpts) *Event {
	evt.SetParent(opts.Parent)
	evt.Category = cat
	evt.Source = opts.Source
	evt.Target = opts.Target
	if opts.Type != "" && opts.Type != api.EventTypeUnknown {
		evt.Type = string(opts.Type)
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
	if evt.Type == "" || evt.Type == string(api.EventTypeUnknown) {
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

func (evt *Event) EventType() api.EventType {
	return api.EventType(evt.Type)
}

func (evt *Event) Err() error {
	if evt.EventType() != api.EventTypeError {
		return nil
	}

	err := &Err{}
	if e := evt.Bind(err); e != nil {
		err = ErrUnexpected(e)
	}

	return err
}

func (evt *Event) Param(key string) string {
	return evt.ParamV(key).String()
}

func (evt *Event) ParamV(key string) *api.Val {
	v, _ := api.ValProto(evt.ParamProto(key))
	if !v.IsNil() {
		return v
	}
	if s := evt.Query(key); s != "" {
		return api.ValString(s)
	}
	if s := evt.Header(key); s != "" {
		return api.ValString(s)
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

func (evt *Event) SetParamV(key string, val *api.Val) EventWriter {
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

func (evt *Event) ValueV(key string) *api.Val {
	v, _ := api.ValProto(evt.ValueProto(key))
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

func (evt *Event) SetValueV(key string, val *api.Val) {
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

func (evt *Event) SetTTL(t time.Duration) {
	evt.Ttl = t.Microseconds()
}

func (evt *Event) ReduceTTL(start time.Time) time.Duration {
	evt.Ttl = evt.Ttl - time.Since(start).Microseconds()
	return evt.TTL()
}

func (evt *Event) Status() int {
	return evt.ValueV(api.ValKeyStatusCode).Int()
}

func (evt *Event) StatusV() *api.Val {
	return evt.ValueV(api.ValKeyStatusCode)
}

func (evt *Event) SetStatus(code int) EventWriter {
	evt.SetValueV(api.ValKeyStatusCode, api.ValInt(code))
	return evt
}

func (evt *Event) SetStatusV(val *api.Val) EventWriter {
	evt.SetValueV(api.ValKeyStatusCode, val)
	return evt
}

func (evt *Event) TraceId() string {
	return evt.Value(api.ValKeyTraceId)
}

func (evt *Event) SetTraceId(val string) {
	evt.SetValue(api.ValKeyTraceId, val)
}

func (evt *Event) SpanId() string {
	return evt.Value(api.ValKeySpanId)
}

func (evt *Event) SetSpanId(val string) {
	evt.SetValue(api.ValKeySpanId, val)
}

func (evt *Event) TraceFlags() byte {
	return byte(evt.ValueV(api.ValKeyTraceFlags).Float())
}

func (evt *Event) SetTraceFlags(val byte) {
	evt.SetValueV(api.ValKeyTraceFlags, api.ValInt(int(val)))
}

func (evt *Event) URL() (*url.URL, error) {
	return url.Parse(evt.Value(api.ValKeyURL))
}

func (evt *Event) SetURL(u *url.URL) EventWriter {
	if u == nil {
		u = &url.URL{}
	}

	evt.SetValue(api.ValKeyURL, u.String())
	evt.SetValue(api.ValKeyHost, strings.ToLower(u.Host))
	evt.SetValue(api.ValKeyPath, u.Path)
	evt.SetValueMap(api.ValKeyQuery, u.Query())

	return evt
}

func (evt *Event) TrimPathPrefix(prefix string) EventWriter {
	if u, err := evt.URL(); err == nil { // success
		u.Path = strings.TrimPrefix(u.Path, prefix)
		evt.SetURL(u)
	}

	return evt
}

func (evt *Event) Query(key string) string {
	return evt.ValueMapKey(api.ValKeyQuery, key)
}

func (evt *Event) QueryV(key string) *api.Val {
	return api.ValString(evt.Query(key))
}

func (evt *Event) QueryDef(key string, def string) string {
	if v := evt.Query(key); v == "" {
		return def
	} else {
		return v
	}
}

func (evt *Event) QueryAll(key string) []string {
	return evt.ValueMapKeyAll(api.ValKeyQuery, key)
}

func (evt *Event) SetQuery(key, value string) EventWriter {
	evt.SetValueMapKey(api.ValKeyQuery, key, value, true)
	return evt
}

func (evt *Event) SetQueryV(key string, value *api.Val) EventWriter {
	return evt.SetQuery(key, value.String())
}

func (evt *Event) DelQuery(key string) EventWriter {
	evt.DelValueMapKey(api.ValKeyQuery, key)
	return evt
}

func (evt *Event) Header(key string) string {
	key = textproto.CanonicalMIMEHeaderKey(key)
	return evt.ValueMapKey(api.ValKeyHeader, key)
}

func (evt *Event) HeaderV(key string) *api.Val {
	return api.ValString(evt.Header(key))
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
	return evt.ValueMapKeyAll(api.ValKeyHeader, key)
}

func (evt *Event) SetHeader(key, value string) EventWriter {
	key = textproto.CanonicalMIMEHeaderKey(key)
	evt.SetValueMapKey(api.ValKeyHeader, key, value, true)
	return evt
}

func (evt *Event) AddHeader(key, value string) EventWriter {
	key = textproto.CanonicalMIMEHeaderKey(key)
	evt.SetValueMapKey(api.ValKeyHeader, key, value, false)
	return evt
}

func (evt *Event) SetHeaderV(key string, value *api.Val) EventWriter {
	return evt.SetHeader(key, value.String())
}

func (evt *Event) DelHeader(key string) EventWriter {
	key = textproto.CanonicalMIMEHeaderKey(key)
	evt.DelValueMapKey(api.ValKeyHeader, key)
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

	evt.ContentType = fmt.Sprintf("%s; %s", api.ContentTypeJSON, api.CharSetUTF8)
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
	case strings.Contains(contType, api.ContentTypeJSON):
		dec := json.NewDecoder(bytes.NewReader(evt.Content))
		if strict {
			dec.DisallowUnknownFields()
		}
		return dec.Decode(v)
	default:
		return fmt.Errorf("%w: %s", ErrUnknownContentType(), evt.ContentType)
	}
}

func (evt *Event) HTTPRequest(ctx context.Context) (*http.Request, error) {
	body := bytes.NewReader(evt.Content)
	req, err := http.NewRequestWithContext(ctx, evt.Value(api.ValKeyMethod), evt.Value(api.ValKeyURL), body)
	if err != nil {
		return nil, err
	}
	req.Header = evt.ValueMap(api.ValKeyHeader)

	return req, nil
}

func (evt *Event) HTTPResponse() *http.Response {
	code := evt.ValueV(api.ValKeyStatusCode).Int()
	if code == 0 {
		code = http.StatusOK
	}
	return &http.Response{
		Status:     evt.Value("status"),
		StatusCode: code,
		Header:     evt.ValueMap(api.ValKeyHeader),
		Body:       io.NopCloser(bytes.NewReader(evt.Content)),
	}
}

func (evt *Event) SetHTTPRequest(httpReq *http.Request) error {
	if content, err := ReadBody(httpReq.Body, httpReq.Header); err != nil {
		return err
	} else {
		evt.Content = content
	}

	evt.ContentType = httpReq.Header.Get("Content-Type")
	evt.Category = Category_REQUEST

	if evt.Context.Environment == "" {
		evt.Context.Environment = GetParamOrHeader(httpReq, api.HeaderEnv, api.HeaderAbbrvEnv, api.HeaderShortEnv)
	}
	if evt.Context.Deployment == "" {
		evt.Context.Deployment = GetParamOrHeader(httpReq, api.HeaderDep, api.HeaderAbbrvDep, api.HeaderShortDep)
	}

	if evt.Type == "" || evt.Type == string(api.EventTypeUnknown) {
		evtType := GetParamOrHeader(httpReq, api.HeaderEventType, api.HeaderAbbrvEventType, api.HeaderShortEventType)
		if evtType != "" {
			evt.Type = evtType
		} else {
			evt.Type = string(api.EventTypeHTTP)
		}
	}

	DelParamOrHeader(httpReq,
		api.HeaderEnv, api.HeaderAbbrvEnv, api.HeaderShortEnv,
		api.HeaderDep, api.HeaderAbbrvDep, api.HeaderShortDep,
		api.HeaderEventType, api.HeaderAbbrvEventType, api.HeaderShortEventType,
	)

	u := *httpReq.URL
	if host := httpReq.Header.Get("X-Forwarded-Host"); host != "" {
		if port := httpReq.Header.Get("X-Forwarded-Port"); port != "" {
			u.Host = fmt.Sprintf("%s:%s", host, port)
		} else {
			u.Host = host
		}
		if scheme := httpReq.Header.Get("X-Forwarded-Proto"); scheme != "" {
			u.Scheme = scheme
		}
	}
	if u.Scheme == "" {
		u.Scheme = "http"
		if httpReq.TLS != nil {
			u.Scheme = "https"
		}
	}
	if u.Host == "" {
		u.Host = httpReq.Host
	}

	evt.SetURL(&u)
	evt.SetValue(api.ValKeyMethod, httpReq.Method)
	evt.SetValueMap(api.ValKeyHeader, httpReq.Header)

	return nil
}

func (evt *Event) SetHTTPResponse(httpResp *http.Response) error {
	if content, err := ReadBody(httpResp.Body, httpResp.Header); err != nil {
		return err
	} else {
		evt.Content = content
	}

	evt.ContentType = httpResp.Header.Get("Content-Type")
	evt.Category = Category_RESPONSE

	if evt.Type == "" || evt.Type == string(api.EventTypeUnknown) {
		evt.Type = string(api.EventTypeHTTP)
	}

	evt.SetValue(api.ValKeyStatus, httpResp.Status)
	evt.SetValueV(api.ValKeyStatusCode, api.ValInt(httpResp.StatusCode))
	evt.SetValueMap(api.ValKeyHeader, httpResp.Header)

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

// DelParamOrHeader deletes all query parameters and headers that match provided
// keys.
func DelParamOrHeader(httpReq *http.Request, keys ...string) {
	query := httpReq.URL.Query()
	for _, key := range keys {
		query.Del(strings.ToLower(key))
		httpReq.Header.Del(key)
	}
	httpReq.URL.RawQuery = query.Encode()
}

// ReadBody reads the body of a HTTP request or response, ensuring
// MaxContentSizeBytes is not exceeded, then closes the reader. If body is 'nil'
// then 'nil, nil' is returned.
func ReadBody(body io.ReadCloser, header http.Header) ([]byte, error) {
	if body != nil {
		defer body.Close()

		if i, err := strconv.Atoi(header.Get(api.HeaderContentLength)); err == nil { // success
			if i > api.DefaultMaxEventSizeBytes {
				return nil, ErrContentTooLarge()
			}
		}

		return io.ReadAll(&LimitedReader{body, api.DefaultMaxEventSizeBytes})
	}

	return nil, nil
}
