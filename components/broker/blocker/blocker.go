package blocker

import (
	"context"
	"fmt"
	"sync"

	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/logger"
)

type Blocker struct {
	reqMap map[string]chan kubefox.DataEvent
	mutex  sync.RWMutex
	log    *logger.Log
}

type RespListener struct {
	id      string
	ch      chan kubefox.DataEvent
	blocker *Blocker
	ctx     context.Context
	log     *logger.Log
}

func NewBlocker(log *logger.Log) *Blocker {
	return &Blocker{
		reqMap: make(map[string]chan kubefox.DataEvent),
		log:    log,
	}
}

func (bl *Blocker) NewRespListener(ctx context.Context, id string) (*RespListener, error) {
	bl.mutex.Lock()
	defer bl.mutex.Unlock()

	bl.log.Debugf("creating channel; id: %s", id)

	if _, exists := bl.reqMap[id]; exists {
		return nil, &ErrDuplicateChannel{id: id}
	}

	ch := make(chan kubefox.DataEvent)
	bl.reqMap[id] = ch

	return &RespListener{
		ctx:     ctx,
		id:      id,
		ch:      ch,
		blocker: bl,
		log:     bl.log,
	}, nil
}

func (bl *Blocker) SendResponse(id string, e kubefox.DataEvent) error {
	ch := bl.Channel(id)
	if ch == nil {
		return &ErrChannelNotFound{id: id}
	}

	ch <- e

	return nil
}

func (bl *Blocker) Channel(id string) chan kubefox.DataEvent {
	bl.mutex.RLock()
	defer bl.mutex.RUnlock()

	return bl.reqMap[id]
}

func (bl *Blocker) CloseChannel(id string) {
	bl.mutex.Lock()
	defer bl.mutex.Unlock()

	bl.log.Debugf("closing channel; id: %s", id)

	ch := bl.reqMap[id]
	delete(bl.reqMap, id)
	if ch != nil {
		close(ch)
	}
}

func (ing *RespListener) Wait() (kubefox.DataEvent, error) {
	defer ing.blocker.CloseChannel(ing.id)

	select {
	case res := <-ing.ch:
		return res, nil
	case <-ing.ctx.Done():
		err := fmt.Errorf("wait failed: %w", ing.ctx.Err())
		ing.log.Error(err)
		return nil, err
	}
}
