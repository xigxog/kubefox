package kubefox

import (
	"sync"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/xigxog/kubefox/libs/core/api/common"
	"github.com/xigxog/kubefox/libs/core/component"
	"github.com/xigxog/kubefox/libs/core/grpc"
	"go.opentelemetry.io/otel/trace"
)

// TODO create interfaces for grpc stuff
type Event interface {
	GetId() string
	GetParentId() string

	GetType() string
	SetType(string)

	GetToken() *grpc.Token
	GetSpan() *grpc.Span
	GetTraceId() string
	GetSource() component.Component
	GetTarget() component.Component
	GetContext() *grpc.EventContext

	GetArg(string) string
	GetArgVar(string) *common.Var
	SetArg(string, string)
	SetArgNumber(string, float64)
	GetValue(string) string
	GetValueVar(string) *common.Var

	GetContentType() string
	SetContentType(string)
	GetContent() []byte
	SetContent([]byte)
	Marshal(any) error
	Unmarshal(any) error
	UnmarshalStrict(any) error

	HTTP() HTTPEvent
	Kube() KubeEvent
}

type DataEvent interface {
	Event

	GetData() *grpc.Data

	ChildEvent() DataEvent
	ChildErrorEvent(error) DataEvent

	SetParent(Event)
	SetParentId(string)

	SetSource(component.Component, string)
	SetTarget(component.Component)

	SetContext(*grpc.EventContext)

	GetFabric() *grpc.Fabric
	SetFabric(*grpc.Fabric)

	SetSpan(*grpc.Span)
	UpdateSpan(trace.Span)

	SetToken(jwt.Token)
	SetValue(string, string)
	SetValueNumber(string, float64)

	HTTPData() HTTPDataEvent
	KubeData() KubeDataEvent

	SetError(error)
	GetError() error
	GetErrorMsg() string
}

type event struct {
	*grpc.Data

	http       *httpEvent
	kubernetes *kubeEvent

	err error

	// used to lock when changing private fields
	mutex sync.RWMutex
}

func IsPlatformEvent(evt DataEvent) bool {
	t := evt.GetType()
	return t == BootstrapRequestType || t == BootstrapResponseType ||
		t == FabricRequestType || t == FabricResponseType
}

func NewEvent(eventType string) Event {
	return NewDataEvent(eventType)
}

func NewErrorEvent(err error) DataEvent {
	evt := NewDataEvent(ErrorEventType)
	evt.SetError(err)

	return evt
}

func NewDataEvent(eventType string) DataEvent {
	evt := EmptyDataEvent()
	evt.SetType(eventType)

	return evt
}

func EmptyDataEvent() DataEvent {
	return newEvent(nil)
}

func EventFromData(data *grpc.Data) DataEvent {
	return newEvent(data)
}

func newEvent(data *grpc.Data) *event {
	evt := &event{}
	evt.Data = data

	if evt.Data == nil {
		evt.Data = &grpc.Data{}
	}
	if evt.Data.Id == "" {
		evt.Data.Id = uuid.NewString()
	}
	if evt.Data.Type == "" {
		evt.Data.Type = UnknownEventType
	}

	evt.http = &httpEvent{event: evt}
	evt.kubernetes = &kubeEvent{event: evt}

	return evt
}

func (evt *event) HTTP() HTTPEvent {
	return evt.http
}

func (evt *event) HTTPData() HTTPDataEvent {
	return evt.http
}

func (evt *event) Kube() KubeEvent {
	return evt.kubernetes
}

func (evt *event) KubeData() KubeDataEvent {
	return evt.kubernetes
}

func (evt *event) GetData() *grpc.Data {
	if evt != nil {
		return evt.Data
	}
	return nil
}

func (evt *event) ChildEvent() DataEvent {
	child := EmptyDataEvent()
	if evt == nil {
		return child
	}
	child.SetParent(evt)

	return child
}

func (evt *event) ChildErrorEvent(err error) DataEvent {
	if evt != nil && evt.GetType() == ErrorEventType && evt.GetError() == err {
		return evt
	}

	errEvent := evt.ChildEvent()
	errEvent.SetError(err)

	return errEvent
}

func (evt *event) SetParent(parent Event) {
	if parent == nil {
		evt.SetParentId("")
	} else {
		evt.SetParentId(parent.GetId())
		evt.SetContext(parent.GetContext())
		evt.SetSpan(parent.GetSpan())
	}
}

func (evt *event) GetSource() component.Component {
	if evt.GetData() == nil || evt.GetData().GetSource() == nil {
		return nil
	}

	return evt.GetData().GetSource()
}

// SetSource creates copy
func (evt *event) SetSource(src component.Component, app string) {
	if src == nil {
		evt.GetData().SetSource(nil)
	} else {
		evt.GetData().SetSource(&grpc.Component{
			App:     app,
			Name:    src.GetName(),
			GitHash: src.GetGitHash(),
			Id:      src.GetId(),
		})
	}
}

func (evt *event) GetTarget() component.Component {
	// avoid returning non-nil interface
	if evt.GetData() == nil || evt.GetData().GetTarget() == nil {
		return nil
	}

	return evt.GetData().GetTarget()
}

func (evt *event) SetTarget(trgt component.Component) {
	if trgt == nil {
		evt.GetData().SetTarget(nil)
	} else {
		evt.GetData().SetTarget(&grpc.Component{
			App:     trgt.GetApp(),
			Name:    trgt.GetName(),
			GitHash: trgt.GetGitHash(),
			Id:      trgt.GetId(),
		})
	}
}

func (evt *event) GetTraceId() string {
	if evt.GetSpan() != nil {
		return evt.GetSpan().GetTraceId()
	}
	return ""
}

func (evt *event) GetError() error {
	if evt != nil {
		return evt.err
	}
	return nil
}

func (evt *event) SetError(err error) {
	evt.err = err
	if err != nil {
		evt.SetType(ErrorEventType)
		evt.SetErrorMsg(err.Error())
	} else {
		evt.SetErrorMsg("")
	}
}
