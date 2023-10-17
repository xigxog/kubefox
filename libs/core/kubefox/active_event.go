package kubefox

import (
	"net/http"
	"net/url"
	"time"
)

type EventReader interface {
	EventType() EventType

	Param(name string) string
	ParamV(name string) *Val
	ParamDef(name string, def string) string

	Query(name string) string
	QueryV(name string) *Val
	QueryDef(name string, def string) string

	Header(name string) string
	HeaderV(name string) *Val
	HeaderDef(name string, def string) string

	Status() int
	StatusV() *Val

	Bind(v any) error
	String() string
	Bytes() []byte
}

type EventWriter interface {
	EventReader

	SetParam(name, value string) EventWriter
	SetParamV(name string, value *Val) EventWriter

	SetQuery(name, value string) EventWriter
	SetQueryV(name string, value *Val) EventWriter

	SetHeader(name, value string) EventWriter
	SetHeaderV(name string, value *Val) EventWriter

	SetStatus(code int) EventWriter
	SetStatusV(val *Val) EventWriter
}

type ActiveEvent struct {
	*Event

	tick time.Time

	queryMap  url.Values
	headerMap http.Header
}

type EventOpts struct {
	Type   EventType
	Parent *Event
	Source *Component
	Target *Component
}

func StartResp(opts EventOpts) *ActiveEvent {
	return StartNewEvent(Category_RESPONSE, opts)
}

func StartReq(opts EventOpts) *ActiveEvent {
	return StartNewEvent(Category_REQUEST, opts)
}

func StartMsg(opts EventOpts) *ActiveEvent {
	return StartNewEvent(Category_MESSAGE, opts)
}

func StartNewEvent(cat Category, opts EventOpts) *ActiveEvent {
	evt := NewEvent()
	evt.SetParent(opts.Parent)
	evt.Category = cat
	evt.Source = opts.Source
	evt.Target = opts.Target
	if opts.Type != EventTypeUnknown {
		evt.Type = string(opts.Type)
	}

	return StartEvent(evt)
}

func StartEvent(evt *Event) *ActiveEvent {
	return &ActiveEvent{
		Event:     evt,
		tick:      time.Now(),
		queryMap:  evt.ValueMap(ValKeyQuery),
		headerMap: evt.ValueMap(ValKeyHeader),
	}
}

func (evt *ActiveEvent) Flush() *Event {
	evt.SetValueMap(ValKeyQuery, evt.queryMap)
	evt.SetValueMap(ValKeyHeader, evt.headerMap)
	evt.ReduceTTL(evt.tick)
	evt.tick = time.Now()

	return evt.Event
}

func (evt *ActiveEvent) TTL() time.Duration {
	return time.Duration(evt.Ttl)*time.Microsecond - time.Since(evt.tick)
}

func (evt *ActiveEvent) Param(name string) string {
	return evt.ParamV(name).String()
}

func (evt *ActiveEvent) ParamV(name string) *Val {
	v := evt.Event.ParamV(name)
	if !v.IsNil() {
		return v
	}
	if s := evt.Query(name); s != "" {
		return ValString(s)
	}
	if s := evt.Header(name); s != "" {
		return ValString(s)
	}

	return ValNil()
}

func (evt *ActiveEvent) ParamDef(name string, def string) string {
	if v := evt.Param(name); v == "" {
		return def
	} else {
		return v
	}
}

func (evt *ActiveEvent) SetParam(name, value string) EventWriter {
	evt.Event.SetParam(name, value)
	return evt
}

func (evt *ActiveEvent) SetParamV(name string, value *Val) EventWriter {
	evt.Event.SetParamV(name, value)
	return evt
}

func (evt *ActiveEvent) QueryMap() url.Values {
	return evt.queryMap
}

func (evt *ActiveEvent) Query(name string) string {
	return evt.queryMap.Get(name)
}

func (evt *ActiveEvent) QueryV(name string) *Val {
	return ValString(evt.Query(name))
}

func (evt *ActiveEvent) QueryDef(name string, def string) string {
	if v := evt.Query(name); v == "" {
		return def
	} else {
		return v
	}
}

func (evt *ActiveEvent) SetQuery(name, value string) EventWriter {
	evt.queryMap.Set(name, value)
	return evt
}

func (evt *ActiveEvent) SetQueryV(name string, value *Val) EventWriter {
	return evt.SetQuery(name, value.String())
}

func (evt *ActiveEvent) DelQuery(name string) EventWriter {
	evt.queryMap.Del(name)
	return evt
}

func (evt *ActiveEvent) HeaderMap() http.Header {
	return evt.headerMap
}

func (evt *ActiveEvent) Header(name string) string {
	return evt.headerMap.Get(name)
}

func (evt *ActiveEvent) HeaderV(name string) *Val {
	return ValString(evt.Header(name))
}

func (evt *ActiveEvent) HeaderDef(name string, def string) string {
	if v := evt.Header(name); v == "" {
		return def
	} else {
		return v
	}
}

func (evt *ActiveEvent) SetHeader(name, value string) EventWriter {
	evt.headerMap.Set(name, value)
	return evt
}

func (evt *ActiveEvent) AddHeader(name, value string) EventWriter {
	evt.headerMap.Add(name, value)
	return evt
}

func (evt *ActiveEvent) SetHeaderV(name string, value *Val) EventWriter {
	return evt.SetHeader(name, value.String())
}

func (evt *ActiveEvent) DelHeader(name string) EventWriter {
	evt.headerMap.Del(name)
	return evt
}

func (evt *ActiveEvent) SetStatus(code int) EventWriter {
	evt.Event.SetStatus(code)
	return evt
}

func (evt *ActiveEvent) SetStatusV(val *Val) EventWriter {
	evt.Event.SetStatusV(val)
	return evt
}

func (evt *ActiveEvent) String() string {
	return string(evt.Bytes())
}

func (evt *ActiveEvent) Bytes() []byte {
	return evt.Event.Content
}
