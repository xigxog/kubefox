package engine

import (
	"sync"
	"time"

	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/api/kubernetes/v1alpha1"
	kubefox "github.com/xigxog/kubefox/core"
	"google.golang.org/protobuf/types/known/structpb"
)

type Receiver int

const (
	ReceiverNATS Receiver = iota
	ReceiverGRPCServer
	ReceiverHTTPServer
	ReceiverHTTPClient
)

type SendEvent func(*BrokerEvent) error

type BrokerEvent struct {
	*kubefox.Event

	EnvVars map[string]*api.Val
	RouteId int64

	TargetAdapter *v1alpha1.EnvAdapter
	Adapters      map[string]*v1alpha1.EnvAdapter

	Receiver   Receiver
	ReceivedAt time.Time
	DoneCh     chan *kubefox.Err

	tick  time.Time
	mutex sync.Mutex
}

func (evt *BrokerEvent) TTL() time.Duration {
	evt.mutex.Lock()
	defer evt.mutex.Unlock()

	if evt.tick.Equal(time.Time{}) {
		evt.tick = evt.ReceivedAt
	}

	evt.ReduceTTL(evt.tick)
	evt.tick = time.Now()

	return evt.Event.TTL()
}

func (evt *BrokerEvent) Done() chan *kubefox.Err {
	return evt.DoneCh
}

func (evt *BrokerEvent) MatchedEvent() *kubefox.MatchedEvent {
	var env map[string]*structpb.Value
	if evt.EnvVars != nil {
		env = make(map[string]*structpb.Value, len(evt.EnvVars))
		for k, v := range evt.EnvVars {
			env[k] = v.Proto()
		}
	}

	return &kubefox.MatchedEvent{
		Event:   evt.Event,
		RouteId: evt.RouteId,
		Env:     env,
	}
}

func (r Receiver) String() string {
	switch r {
	case ReceiverNATS:
		return "nats-client"
	case ReceiverGRPCServer:
		return "grpc-server"
	case ReceiverHTTPServer:
		return "http-server"
	case ReceiverHTTPClient:
		return "http-client"
	default:
		return "unknown"
	}
}
