package engine

import (
	"context"

	"github.com/xigxog/kubefox/libs/core/kubefox"
)

type EventReceiver int

const (
	GRPCSvc EventReceiver = iota
	JetStreamSvc
	HTTPClientSvc
	HTTPSrvSvc
)

type SendEvent func(*kubefox.MatchedEvent) error
type RecvEvent func(*ReceivedEvent) error

type ReceivedEvent struct {
	Event        *kubefox.Event
	Receiver     EventReceiver
	RecvTime     int64
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
