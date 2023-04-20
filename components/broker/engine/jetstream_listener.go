package engine

import (
	"context"
	"os"
	"sync"

	"github.com/xigxog/kubefox/components/broker/jetstream"
	"github.com/xigxog/kubefox/libs/core/kubefox"
)

type JetStreamListener struct {
	Broker

	sub jetstream.Subscription

	started bool
	mutex   sync.Mutex
}

func NewJetStreamListener(brk Broker) *JetStreamListener {
	return &JetStreamListener{
		Broker: brk,
	}
}

func (c *JetStreamListener) Start() {
	defer c.mutex.Unlock()
	c.mutex.Lock()

	if c.started {
		// already started
		return
	}
	c.started = true

	sub, err := c.JetStreamClient().Subscribe(&jetstream.SubscriptionConfig{
		Subject:  c.Component().GetRequestSubject(),
		Consumer: c.Component().GetRequestConsumer(),
		Worker:   "listener",
		Handler:  c.handleEvent,
	})
	if err != nil {
		c.Log().Error(err)
		os.Exit(kubefox.JetStreamErrorCode)
	}
	c.sub = sub

	go func() {
		err := c.sub.Start()
		if err != nil {
			c.Log().Error(err)
			os.Exit(kubefox.JetStreamErrorCode)
		}
	}()
}

func (c *JetStreamListener) Shutdown() {
	defer c.mutex.Unlock()
	c.mutex.Lock()

	if !c.started {
		c.Log().Debug("not started, nothing to do")
		return
	}

	c.Log().Info("JetStream listener shutting down")

	if err := c.sub.Close(); err != nil {
		c.Log().Error(err)
	}

	c.started = false
}

func (c *JetStreamListener) handleEvent(queue chan kubefox.DataEvent) {
	for req := range queue {
		resp := c.InvokeLocalComponent(context.Background(), req)
		c.JetStreamClient().Publish(req.GetSource().GetResponseSubject(), resp)
	}
}
