package engine

import (
	"context"
	"time"

	"github.com/xigxog/kubefox/libs/core/kubefox"
)

type EventReceiver int

const (
	EventReceiverGRPC EventReceiver = iota
	EventReceiverJetStream
	EventReceiverHTTPClient
	EventReceiverHTTPSrv
)

type SendEvent func(*kubefox.MatchedEvent) error
type RecvEvent func(*ReceivedEvent) error

type ReceivedEvent struct {
	Event        *kubefox.Event
	Receiver     EventReceiver
	RecvTime     time.Time
	Subscription ReplicaSubscription
	Context      context.Context
	ErrCh        chan error
}

func (rEvt *ReceivedEvent) Err(err error) {
	if rEvt.ErrCh != nil {
		rEvt.ErrCh <- err
	}
}

type IdSeq struct {
	EventId  string
	Sequence uint64
}
