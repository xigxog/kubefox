package engine

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/xigxog/kubefox/components/broker/config"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/logkf"
	"google.golang.org/protobuf/proto"
)

// Content types
const (
	name                 = "jetstream-client"
	eventStream          = "EVENTS"
	eventSubjectWildcard = "evt.>"
	compBucket           = "COMPONENTS"
)

var (
	maxMsgSize     = int32(1024 * 1024 * 5) // 5 MiB
	EventStreamTTL = time.Hour * 24 * 3     // 3 days
	ComponentsTTL  = time.Hour * 12         // 12 hours
)

type RecvMsg func(*nats.Msg)

type JetStreamClient struct {
	jetstream.JetStream

	nc       *nats.Conn
	eventsKV jetstream.KeyValue
	compKV   jetstream.KeyValue

	brk Broker

	mutex sync.Mutex
	log   *logkf.Logger
}

func NewJetStreamClient(brk Broker) *JetStreamClient {
	return &JetStreamClient{
		log: logkf.Global,
		brk: brk,
	}
}

func (c *JetStreamClient) Connect(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.nc != nil && c.nc.IsConnected() {
		c.log.Debug("jetstream client already connected")
		return nil
	}

	c.log.Debug("jetstream client connecting")

	var err error

	name, _ := os.LookupEnv("POD_NAME")
	if name == "" {
		name, _ = os.Hostname()
	}
	if name == "" {
		name = "unknown"
	}

	c.nc, err = nats.Connect(
		fmt.Sprintf("nats://%s", config.NATSAddr),
		nats.Name("broker-"+name),
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(3),
		nats.PingInterval(time.Second*30),
		// c.natsTLS(c.cfg.Namespace),
		nats.RootCAs(kubefox.PathCACert),
		nats.ClientCert(kubefox.PathTLSCert, kubefox.PathTLSKey),
	)
	if err != nil {
		return c.log.ErrorN("connecting to NATS failed: %v", err)
	}
	// if c.JetStreamContext, err = c.nc.JetStream(); err != nil {
	// 	return c.log.ErrorN("connecting to JetStream failed: %v", err)
	// }

	if c.JetStream, err = jetstream.New(c.nc); err != nil {
		return c.log.ErrorN("connecting to JetStream failed: %v", err)
	}

	if err := c.setupStream(ctx); err != nil {
		return err
	}
	if err := c.setupCompsKV(ctx); err != nil {
		return err
	}
	if err := c.setupEventsKV(ctx); err != nil {
		return err
	}
	if err := c.startKVUpdater(); err != nil {
		return err
	}

	c.log.Info("jetstream client connected")
	return nil
}

func (c *JetStreamClient) setupStream(ctx context.Context) error {
	if _, err := c.Stream(ctx, eventStream); err != nil {
		_, err = c.CreateStream(ctx, jetstream.StreamConfig{
			Name: eventStream,
			// cannot be updated
			Storage:   jetstream.FileStorage,
			Retention: jetstream.LimitsPolicy,
			//
		})
		if err != nil {
			return c.log.ErrorN("unable to create events stream: %v", err)
		}
	}
	_, err := c.UpdateStream(ctx, jetstream.StreamConfig{
		Name:        eventStream,
		Description: "Durable disk backed event stream. Events are retained for 3 days.",
		Subjects:    []string{eventSubjectWildcard},
		MaxMsgSize:  maxMsgSize,
		Discard:     jetstream.DiscardOld,
		MaxAge:      EventStreamTTL,
		NoAck:       true,
	})
	if err != nil {
		return c.log.ErrorN("unable to create events stream: %v", err)
	}

	return nil
}

func (c *JetStreamClient) setupCompsKV(ctx context.Context) (err error) {
	c.compKV, err = c.CreateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket:      compBucket,
		Description: "Durable key/value store used by Brokers to register Components. Values are retained for 12 hours.",
		Storage:     jetstream.FileStorage,
		TTL:         ComponentsTTL,
	})
	if err != nil {
		return c.log.ErrorN("unable to create components key/value store: %v", err)
	}

	return nil
}

func (c *JetStreamClient) setupEventsKV(ctx context.Context) (err error) {
	c.eventsKV, err = c.CreateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket:      eventStream,
		Description: "Durable disk backed key/value store for Event ids. Values are retained for 3 days.",
		Storage:     jetstream.FileStorage,
		TTL:         EventStreamTTL,
	})
	if err != nil {
		return c.log.ErrorN("unable to create archive key/value store: %w", err)
	}

	return nil
}

func (c *JetStreamClient) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.log.Info("jetstream client closing")

	if c.nc != nil {
		c.nc.Close()
	}
}

func (c *JetStreamClient) EventsKV() jetstream.KeyValue {
	return c.eventsKV
}

func (c *JetStreamClient) ComponentsKV() jetstream.KeyValue {
	return c.compKV
}

func (c *JetStreamClient) Publish(subject string, evt *kubefox.Event) error {
	dataBytes, err := proto.Marshal(evt)
	if err != nil {
		return err
	}

	h := make(nats.Header)
	h.Set(nats.MsgIdHdr, evt.Id)
	// h.Set("ce_specversion", "1.0")
	// h.Set("ce_type", evt.Type)
	// h.Set("ce_id", evt.Id)
	// h.Set("ce_time", time.Now().Format(time.RFC3339))
	// h.Set("ce_source", fmt.Sprintf("kubefox:component:%s", evt.Source.Key()))
	// h.Set("ce_dataschema", kubefox.DataSchemaKubefox)
	// h.Set("ce_datacontenttype", kubefox.ContentTypeProtobuf)

	// Note use of NATS directly instead of JetStream. This is done for
	// performance and memory efficiency. The risk is a msg not getting
	// delivered as there is no ACK from the server.
	return c.nc.PublishMsg(&nats.Msg{
		Subject: subject,
		Header:  h,
		Data:    dataBytes,
	})
}

func (c *JetStreamClient) PullEvents(sub ReplicaSubscription) error {
	log := c.log.WithComponent(sub.Component())

	consumer, err := c.JetStream.CreateOrUpdateConsumer(sub.Context(), eventStream, jetstream.ConsumerConfig{
		Name:          sub.Component().Key(),
		FilterSubject: sub.Component().Subject(),
		DeliverPolicy: jetstream.DeliverNewPolicy,
		AckPolicy:     jetstream.AckNonePolicy,
	})
	if err != nil {
		return log.ErrorN("unable to create JetStream consumer for component: %v", err)
	}

	var grpConsumer jetstream.Consumer
	if sub.IsGroupEnabled() {
		grpConsumer, err = c.JetStream.CreateOrUpdateConsumer(sub.Context(), eventStream, jetstream.ConsumerConfig{
			Name:              sub.Component().GroupKey(),
			FilterSubject:     sub.Component().GroupSubject(),
			AckPolicy:         jetstream.AckNonePolicy,
			InactiveThreshold: EventStreamTTL,
		})
		if err != nil {
			return log.ErrorN("unable to create JetStream consumer for group: %v", err)
		}
	}

	recvMsg := func(msg jetstream.Msg) {
		evt := kubefox.NewEvent()
		if err := proto.Unmarshal(msg.Data(), evt); err != nil {
			evtId := msg.Headers().Get(jetstream.MsgIDHeader)
			log.With(logkf.KeyEventId, evtId).Warn("message contains invalid event data: %v", err)
			return
		}
		if md, err := msg.Metadata(); err == nil { // success
			evt.ReduceTTL(md.Timestamp)
		}

		rEvt := &LiveEvent{
			Event:        evt,
			Receiver:     ReceiverJetStream,
			ReceivedAt:   time.Now(),
			Subscription: sub,
		}
		if err := c.brk.RecvEvent(rEvt); err != nil {
			c.log.WithEvent(evt).Debug(err)
			if evt.Target.Id == "" && errors.Is(err, ErrSubCanceled) {
				// Any component replica can process, republish event.
				if err := c.nc.PublishMsg(&nats.Msg{
					Subject: msg.Subject(),
					Header:  msg.Headers(),
					Data:    msg.Data(),
				}); err != nil {
					c.log.WithEvent(evt).Debug(err)
				}
			}
		}
	}

	var (
		conCtx    jetstream.ConsumeContext
		grpConCtx jetstream.ConsumeContext
	)
	if conCtx, err = consumer.Consume(recvMsg); err != nil {
		return err
	}
	if grpConsumer != nil {
		if grpConCtx, err = grpConsumer.Consume(recvMsg); err != nil {
			conCtx.Stop()
			return err
		}
	}
	go func() {
		<-sub.Context().Done()
		log.Debug("subscription closed, stopping consumers")
		if grpConCtx != nil {
			grpConCtx.Stop()
		}
		conCtx.Stop()

	}()
	log.Debug("consumers started")

	return nil
}

func (c *JetStreamClient) startKVUpdater() error {
	// log := c.log.With(logkf.KeyWorker, "kv-updater")
	// cfg := &nats.ConsumerConfig{
	// 	Name:      "kv-updater",
	// 	Durable:   "kv-updater",
	// 	AckPolicy: nats.AckNonePolicy,
	// 	// FilterSubject:     grpSubj,
	// 	InactiveThreshold: EventStreamTTL,
	// 	HeadersOnly:       true,
	// }
	// if _, err := c.AddConsumer(eventStream, cfg); err != nil {
	// 	if _, err := c.UpdateConsumer(eventStream, cfg); err != nil {
	// 		return log.ErrorN("unable to update JetStream consumer: %v", err)
	// 	}

	// }
	// sub, err := c.PullSubscribe(
	// 	eventSubjectWildcard,
	// 	"kv-updater",
	// 	nats.Bind(eventStream, "kv-updater"),
	// 	// nats.Context(sub.Context()),
	// 	nats.AckNone(),
	// 	nats.HeadersOnly(),
	// )
	// if err != nil {
	// 	return log.ErrorN("unable to create JetStream pull subscription: %v", err)
	// }

	// go c.pullEvents(log, sub, func(m *nats.Msg) {
	// 	md, err := m.Metadata()
	// 	if err != nil {
	// 		log.Warnf("unable to get sequence: %v", err)
	// 	}

	// 	log.DebugInterface("", m)
	// 	log.Debugf("seq: %d", md.Sequence.Stream)
	// })

	return nil
}

func (c *JetStreamClient) IsHealthy(ctx context.Context) bool {
	return c.nc != nil && c.nc.IsConnected()
}

func (c *JetStreamClient) Name() string {
	return name
}
