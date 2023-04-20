package jetstream

import (
	"errors"
	"fmt"
	"math"
	"runtime"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/xigxog/kubefox/libs/core/component"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
)

type SubscriptionConfig struct {
	Subject            string
	Consumer           string
	Worker             string
	Concurrent         int
	UnsubscribeOnClose bool

	Handler func(queue chan kubefox.DataEvent)
}

type Subscription interface {
	Start() error
	Stop()
	Close() error
	ErrCount() int
	Stopped() bool
	Closed() bool
}

type subscription struct {
	cfg *SubscriptionConfig

	errCount int
	stopped  bool
	closed   bool

	natsSub *nats.Subscription
	stopCh  chan bool
	mutex   sync.Mutex

	tracer trace.Tracer
	log    *logger.Log
}

func NewSubscription(cfg *SubscriptionConfig, natsSub *nats.Subscription, log *logger.Log) Subscription {
	if cfg.Concurrent <= 0 {
		cfg.Concurrent = runtime.NumCPU() / 2
	}

	return &subscription{
		cfg:      cfg,
		errCount: 0,
		stopped:  true,
		closed:   false,
		natsSub:  natsSub,
		stopCh:   make(chan bool, 1),
		tracer:   otel.Tracer(""),
		log:      log,
	}
}

func (sub *subscription) Start() error {
	sub.mutex.Lock()
	if !sub.stopped {
		sub.log.Debugf("subscription for subject %s is already started", sub.cfg.Subject)
		sub.mutex.Unlock()
		return nil
	}

	if sub.closed {
		sub.log.Errorf("subscription for subject %s is closed", sub.cfg.Subject)
		sub.mutex.Unlock()
		return kubefox.ErrSubscriptionClosed
	}

	sub.stopped = false
	queue := make(chan kubefox.DataEvent)
	sub.mutex.Unlock()

	// start handlers
	sub.log.Infof("starting %d handlers for subject %s", sub.cfg.Concurrent, sub.cfg.Subject)
	var wg sync.WaitGroup
	for i := 0; i < sub.cfg.Concurrent; i++ {
		wg.Add(1)
		go func() {
			sub.cfg.Handler(queue)
			wg.Done()
		}()
	}

	for !sub.stopped {
		// TODO fetch count can be increased for higher throughput vs lower
		// latency perhaps turn into a config of the component?
		msgs, err := sub.natsSub.Fetch(1)
		if err != nil {
			if errors.Is(err, nats.ErrTimeout) {
				continue
			}

			sub.errCount += 1
			sub.log.Error(err)
			if errors.Is(err, nats.ErrConnectionClosed) {
				return err
			}

			// simple backoff, max 3 seconds
			sleepTime := math.Min(3, float64(sub.errCount-1))
			time.Sleep(time.Duration(sleepTime) * time.Second)
			continue
		}
		sub.errCount = 0

		for _, msg := range msgs {
			evt := kubefox.EmptyDataEvent()
			if err := proto.Unmarshal(msg.Data, evt.GetData()); err != nil {
				evt = evt.ChildErrorEvent(fmt.Errorf("unmarshal error: %w", err))
				sub.log.Error(evt.GetError())
			}

			log := sub.log
			if evt.GetTraceId() != "" {
				log = sub.log.With("traceId", evt.GetTraceId())
			}
			log.Debugf("received data; id: %s, subject: %s", evt.GetId(), sub.cfg.Subject)

			if _, err := component.ParseURI(msg.Header.Get("ce_source")); err != nil {
				evt = evt.ChildErrorEvent(fmt.Errorf("component address parse error: %w", err))
				sub.log.Error(evt.GetError())
			}

			queue <- evt

			msg.Ack()
		}
	}

	close(queue)
	wg.Wait() // wait for handlers to finish
	sub.stopCh <- true

	return nil
}

func (s *subscription) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.closed {
		return
	}

	if s.stopped {
		return
	}

	s.stopped = true
	<-s.stopCh // wait for fetch to stop
}

// Once closed a subscription can no longer be used
func (s *subscription) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.closed {
		return nil
	}

	s.log.Infof("closing %s.%s for subject %s", s.cfg.Worker, s.cfg.Consumer, s.cfg.Subject)
	s.closed = true

	if !s.stopped {
		s.stopped = true
		<-s.stopCh // wait for fetch to stop
	}

	if s.cfg.UnsubscribeOnClose {
		err := s.natsSub.Unsubscribe()
		if err != nil {
			s.log.Error(err)
			return err
		} else {
			s.log.Infof("unsubscribed '%s.%s' from '%s'",
				s.cfg.Worker, s.cfg.Consumer, s.cfg.Subject)
		}
	}

	return nil
}

func (s *subscription) ErrCount() int {
	return s.errCount
}

func (s *subscription) Stopped() bool {
	return s.stopped
}

func (s *subscription) Closed() bool {
	return s.closed
}
