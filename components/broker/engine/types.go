package engine

import (
	"sync"
	"time"

	kubefox "github.com/xigxog/kubefox/core"
)

type Receiver int

const (
	ReceiverNATS Receiver = iota
	ReceiverGRPCServer
	ReceiverHTTPServer
	ReceiverHTTPClient
)

type SendEvent func(*LiveEvent) error
type RecvEvent func(*LiveEvent) error

type LiveEvent struct {
	*kubefox.Event

	MatchedEvent *kubefox.MatchedEvent

	Receiver   Receiver
	ReceivedAt time.Time
	SentCh     chan struct{}
	ErrCh      chan error

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

func (evt *LiveEvent) Err() chan error {
	return evt.ErrCh
}

func (evt *LiveEvent) Sent() chan struct{} {
	return evt.SentCh
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
