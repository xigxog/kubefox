package engine

import (
	"github.com/xigxog/kubefox/libs/core/kubefox"
)

type SendEvent func(*kubefox.MatchedEvent) error
type RecvEvent func(*ReceivedEvent) error

type ReceivedEvent struct {
	*kubefox.ActiveEvent
	Subscription ReplicaSubscription
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
