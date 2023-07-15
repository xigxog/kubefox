package engine

import (
	"context"
	"os"
	"sync"

	"github.com/xigxog/kubefox/components/broker/jetstream"

	"github.com/xigxog/kubefox/libs/core/kubefox"
)

type JetStreamSender struct {
	Broker

	sub jetstream.Subscription

	started bool
	mutex   sync.RWMutex
}

func NewJetStreamSender(brk Broker) *JetStreamSender {
	return &JetStreamSender{
		Broker: brk,
	}
}

func (js *JetStreamSender) Start() {
	defer js.mutex.Unlock()
	js.mutex.Lock()

	if js.started {
		return
	}

	sub, err := js.JetStreamClient().Subscribe(&jetstream.SubscriptionConfig{
		Subject:            js.Component().GetResponseSubject(),
		Consumer:           js.Component().GetResponseConsumer(),
		Worker:             "sender",
		Handler:            js.processQueue,
		UnsubscribeOnClose: true,
	})
	if err != nil {
		js.Log().Error(err)
		os.Exit(kubefox.JetStreamErrorCode)
	}
	js.sub = sub

	go func() {
		err := js.sub.Start()
		if err != nil {
			js.Log().Error(err)
			os.Exit(kubefox.JetStreamErrorCode)
		}
	}()
}

func (js *JetStreamSender) Shutdown() {
	defer js.mutex.Unlock()
	js.mutex.Lock()

	if !js.started {
		return
	}

	js.Log().Info("JetStream sender shutting down")

	if err := js.sub.Close(); err != nil {
		js.Log().Error(err)
	}

	js.started = false
}

func (js *JetStreamSender) SendEvent(ctx context.Context, req kubefox.DataEvent) (resp kubefox.DataEvent) {
	ing, err := js.Blocker().NewRespListener(ctx, req.GetId())
	if err != nil {
		resp = req.ChildErrorEvent(err)
		js.Log().Error(resp.GetError())
		return
	}

	// Send request to remote component via NATS
	if _, err = js.JetStreamClient().Publish(req.GetTarget().GetRequestSubject(), req); err != nil {
		resp = req.ChildErrorEvent(err)
		js.Log().Error(resp.GetError())
		return
	}

	// Wait for the response
	if resp, err = ing.Wait(); err != nil {
		resp = req.ChildErrorEvent(err)
		js.Log().Error(resp.GetError())
		return
	}
	resp.SetParent(req)

	return
}

func (js *JetStreamSender) processQueue(queue chan kubefox.DataEvent) {
	for resp := range queue {
		// TODO need better handling
		if err := js.Blocker().SendResponse(resp.GetParentId(), resp); err != nil {
			js.Log().Error(err)
		}
	}
}
