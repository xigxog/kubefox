package kubefox

import (
	"net/http"
	"net/url"
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

type EventRW struct {
	*Event

	queryMap  url.Values
	headerMap http.Header
}

func NewResp(parent *Event, src *Component) *EventRW {
	resp := NewEvent()
	resp.SetParent(parent)
	resp.Category = Category_CATEGORY_RESPONSE
	resp.Source = src
	resp.Target = parent.Source

	return NewEventRW(resp)
}

func NewReq(typ EventType, parent *Event, src, tgt *Component) *EventRW {
	resp := NewEvent()
	resp.SetParent(parent)
	resp.Category = Category_CATEGORY_REQUEST
	resp.Type = string(typ)
	resp.Source = src
	resp.Target = tgt

	return NewEventRW(resp)
}

func NewEventRW(evt *Event) *EventRW {
	return &EventRW{
		Event:     evt,
		queryMap:  evt.ValueMap(ValKeyQuery),
		headerMap: evt.ValueMap(ValKeyHeader),
	}
}

func (evt *EventRW) Flush() {
	evt.SetValueMap(ValKeyQuery, evt.queryMap)
	evt.SetValueMap(ValKeyHeader, evt.headerMap)
}

func (evt *EventRW) Param(name string) string {
	return evt.ParamV(name).String()
}

func (evt *EventRW) ParamV(name string) *Val {
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

func (evt *EventRW) ParamDef(name string, def string) string {
	if v := evt.Param(name); v == "" {
		return def
	} else {
		return v
	}
}

func (evt *EventRW) SetParam(name, value string) EventWriter {
	evt.Event.SetParam(name, value)
	return evt
}

func (evt *EventRW) SetParamV(name string, value *Val) EventWriter {
	evt.Event.SetParamV(name, value)
	return evt
}

func (evt *EventRW) QueryMap() url.Values {
	return evt.queryMap
}

func (evt *EventRW) Query(name string) string {
	return evt.queryMap.Get(name)
}

func (evt *EventRW) QueryV(name string) *Val {
	return ValString(evt.Query(name))
}

func (evt *EventRW) QueryDef(name string, def string) string {
	if v := evt.Query(name); v == "" {
		return def
	} else {
		return v
	}
}

func (evt *EventRW) SetQuery(name, value string) EventWriter {
	evt.queryMap.Set(name, value)
	return evt
}

func (evt *EventRW) SetQueryV(name string, value *Val) EventWriter {
	return evt.SetQuery(name, value.String())
}

func (evt *EventRW) DelQuery(name string) EventWriter {
	evt.queryMap.Del(name)
	return evt
}

func (evt *EventRW) HeaderMap() http.Header {
	return evt.headerMap
}

func (evt *EventRW) Header(name string) string {
	return evt.headerMap.Get(name)
}

func (evt *EventRW) HeaderV(name string) *Val {
	return ValString(evt.Header(name))
}

func (evt *EventRW) HeaderDef(name string, def string) string {
	if v := evt.Header(name); v == "" {
		return def
	} else {
		return v
	}
}

func (evt *EventRW) SetHeader(name, value string) EventWriter {
	evt.headerMap.Set(name, value)
	return evt
}

func (evt *EventRW) AddHeader(name, value string) EventWriter {
	evt.headerMap.Add(name, value)
	return evt
}

func (evt *EventRW) SetHeaderV(name string, value *Val) EventWriter {
	return evt.SetHeader(name, value.String())
}

func (evt *EventRW) DelHeader(name string) EventWriter {
	evt.headerMap.Del(name)
	return evt
}

func (evt *EventRW) SetStatus(code int) EventWriter {
	evt.Event.SetStatus(code)
	return evt
}

func (evt *EventRW) SetStatusV(val *Val) EventWriter {
	evt.Event.SetStatusV(val)
	return evt
}

func (evt *EventRW) String() string {
	return string(evt.Bytes())
}

func (evt *EventRW) Bytes() []byte {
	return evt.Event.Content
}
