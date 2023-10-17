package engine

import (
	"github.com/xigxog/kubefox/libs/core/kubefox"
)

type Receiver int

const (
	ReceiverJetStream Receiver = iota
	ReceiverGRPCServer
	ReceiverHTTPServer
	ReceiverHTTPClient
)

type SendEvent func(*kubefox.MatchedEvent) error
type RecvEvent func(*ReceivedEvent) error

type ReceivedEvent struct {
	*kubefox.ActiveEvent

	Receiver     Receiver
	Subscription ReplicaSubscription
	ErrCh        chan error
}

type IdSeq struct {
	EventId  string
	Sequence uint64
}

func (rEvt *ReceivedEvent) Err(err error) {
	if rEvt.ErrCh != nil {
		rEvt.ErrCh <- err
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
