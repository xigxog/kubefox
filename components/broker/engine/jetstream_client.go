package engine

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/xigxog/kubefox/components/broker/config"
	kubefox "github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/logkf"
	"google.golang.org/protobuf/proto"
)

// Content types
const (
	jsSvcName            = "jetstream-client"
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
	nc     *nats.Conn
	js     jetstream.JetStream
	compKV jetstream.KeyValue

	consumerMap map[string]bool

	brk Broker

	mutex sync.Mutex
	log   *logkf.Logger
}

func NewJetStreamClient(brk Broker) *JetStreamClient {
	return &JetStreamClient{
		consumerMap: make(map[string]bool),
		brk:         brk,
		log:         logkf.Global,
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

	c.nc, err = nats.Connect(
		fmt.Sprintf("nats://%s", config.NATSAddr),
		nats.Name("broker-"+c.brk.Component().Id),
		nats.RootCAs(kubefox.PathCACert),
		nats.ClientCert(kubefox.PathTLSCert, kubefox.PathTLSKey),
	)
	if err != nil {
		return c.log.ErrorN("connecting to NATS failed: %v", err)
	}

	if c.js, err = jetstream.New(c.nc); err != nil {
		return c.log.ErrorN("connecting to JetStream failed: %v", err)
	}

	// if err := c.setupStream(ctx); err != nil {
	// 	return err
	// }
	if err := c.setupCompsKV(ctx); err != nil {
		return err
	}

	c.log.Info("jetstream client connected")
	return nil
}

func (c *JetStreamClient) setupStream(ctx context.Context) error {
	if _, err := c.js.Stream(ctx, eventStream); err != nil {
		_, err = c.js.CreateStream(ctx, jetstream.StreamConfig{
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
	_, err := c.js.UpdateStream(ctx, jetstream.StreamConfig{
		Name:        eventStream,
		Description: "Durable disk backed event stream. Events are retained for 3 days.",
		Subjects:    []string{eventSubjectWildcard},
		MaxMsgSize:  maxMsgSize,
		Discard:     jetstream.DiscardOld,
		MaxAge:      EventStreamTTL,
		Duplicates:  time.Millisecond * 100, // minimum value
	})
	if err != nil {
		return c.log.ErrorN("unable to create events stream: %v", err)
	}

	return nil
}

func (c *JetStreamClient) setupCompsKV(ctx context.Context) (err error) {
	c.compKV, err = c.js.CreateKeyValue(ctx, jetstream.KeyValueConfig{
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

func (c *JetStreamClient) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.log.Info("jetstream client closing")
	if c.nc != nil {
		c.nc.Drain()
	}
}

func (c *JetStreamClient) ComponentsKV() jetstream.KeyValue {
	return c.compKV
}

func (c *JetStreamClient) Request(subject string, evt *kubefox.Event) error {
	msg, err := c.Msg(subject, evt)
	if err != nil {
		return err
	}

	// Note use of NATS directly instead of JetStream. This is done for
	// performance and memory efficiency. The risk is a msg not getting
	// delivered as there is no ACK from the server.
	resp, err := c.nc.RequestMsg(msg, evt.TTL())
	if err != nil {
		return err
	}

	c.handleMsg(resp)

	return nil
}

func (c *JetStreamClient) Publish(subject string, evt *kubefox.Event) error {
	msg, err := c.Msg(subject, evt)
	if err != nil {
		return err
	}

	c.log.Debugf("publishing nats msg to subj: %s", subject)

	// Note use of NATS directly instead of JetStream. This is done for
	// performance and memory efficiency. The risk is a msg not getting
	// delivered as there is no ACK from the server.
	return c.nc.PublishMsg(msg)
}

func (c *JetStreamClient) Msg(subject string, evt *kubefox.Event) (*nats.Msg, error) {
	dataBytes, err := proto.Marshal(evt)
	if err != nil {
		return nil, err
	}

	h := make(nats.Header)
	// Note, use of `Nats-Msg-Id` enables de-dupe and increases NATS mem usage.
	h.Set(kubefox.CloudEventId, evt.Id)

	// Headers create sizeable overhead for storage. Disabling most for now.
	//
	// h.Set("ce_specversion", "1.0")
	// h.Set("ce_type", evt.Type)
	// h.Set("ce_time", time.Now().Format(time.RFC3339))
	// h.Set("ce_source", fmt.Sprintf("kubefox:component:%s", evt.Source.Key()))
	// h.Set("ce_dataschema", kubefox.DataSchemaKubefox)
	// h.Set("ce_datacontenttype", kubefox.ContentTypeProtobuf)
	//

	return &nats.Msg{
		Subject: subject,
		Header:  h,
		Data:    dataBytes,
	}, nil
}

func (c *JetStreamClient) ConsumeEvents(ctx context.Context, name, subj string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, found := c.consumerMap[name]; found {
		return nil
	}

	c.log.Debugf("subscribing to nats; queue: %s, subj: %s", name, subj)
	sub, err := c.nc.QueueSubscribe(subj, name, c.handleMsg)
	// sub, err := c.nc.Subscribe(subj, c.handleMsg)
	if err != nil {
		return err
	}

	// go func() {
	// 	sub, err := c.nc.SubscribeSync(subj)
	// 	if err != nil {
	// 		c.log.Error(err)
	// 	}

	// 	m, err := sub.NextMsg(time.Minute * 5)
	// 	if err != nil {
	// 		c.log.Error(err)
	// 	}
	// 	c.log.Debug("got a msg")
	// 	c.handleMsg(m)
	// }()

	go func() {
		<-ctx.Done()
		c.log.Debug("context done, draining subscription '%s'", name)

		c.mutex.Lock()
		defer c.mutex.Unlock()

		if err := sub.Drain(); err != nil {
			c.log.Error("error draining subscription '%s': %v", name, err)
		}
		delete(c.consumerMap, name)
	}()

	return nil
}

func (c *JetStreamClient) handleMsg(msg *nats.Msg) {
	c.log.Debugf("handling msg from nats")

	evt := kubefox.NewEvent()
	if err := proto.Unmarshal(msg.Data, evt); err != nil {
		evtId := msg.Header.Get(kubefox.CloudEventId)
		c.log.With(logkf.KeyEventId, evtId).Warn("message contains invalid event data: %v", err)
		return
	}
	if md, err := msg.Metadata(); err == nil { // success
		evt.ReduceTTL(md.Timestamp)
	}

	lEvt := &LiveEvent{
		Event:      evt,
		Receiver:   ReceiverJetStream,
		ReceivedAt: time.Now(),
	}
	if err := c.brk.RecvEvent(lEvt); err != nil {
		log := c.log.WithEvent(evt)
		log.Debug(err)
		if evt.Target.Id == "" && errors.Is(err, ErrSubCanceled) {
			// Any component replica can process, redeliver event.
			log.Debug("nacking event from component group subject")
			// msg.Nak()
			return
		}
	}
	// msg.Ack()
}

func (c *JetStreamClient) IsHealthy(ctx context.Context) bool {
	return c.nc != nil && c.nc.IsConnected()
}

func (c *JetStreamClient) Name() string {
	return jsSvcName
}
