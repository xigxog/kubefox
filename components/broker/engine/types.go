package engine

import (
	"sync"
	"time"

	kubefox "github.com/xigxog/kubefox/core"
)

type Receiver int

const (
	ReceiverJetStream Receiver = iota
	ReceiverGRPCServer
	ReceiverHTTPServer
	ReceiverHTTPClient
)

type SendEvent func(*LiveEvent) error
type RecvEvent func(*LiveEvent) error

type LiveEvent struct {
	*kubefox.Event

	MatchedEvent *kubefox.MatchedEvent

	Receiver     Receiver
	ReceivedAt   time.Time
	Subscription ReplicaSubscription
	ErrCh        chan error

	tick  time.Time
	mutex sync.Mutex
}

func (evt *LiveEvent) TTL() time.Duration {
	evt.mutex.Lock()
	defer evt.mutex.Unlock()

	if evt.tick.Equal(time.Time{}) {
		evt.tick = evt.ReceivedAt
	}

	evt.ReduceTTL(evt.tick)
	evt.tick = time.Now()

	return evt.Event.TTL()
}

func (evt *LiveEvent) Err(err error) {
	if evt.ErrCh != nil {
		evt.ErrCh <- err
	}
}

func (r Receiver) String() string {
	switch r {
	case ReceiverJetStream:
		return "jetstream"
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
