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

const (
	natsSvcName          = "nats-client"
	eventSubjectWildcard = "evt.>"
	compBucket           = "COMPONENTS"
)

var (
	EventStreamTTL = time.Hour * 24 * 3 // 3 days
	ComponentsTTL  = time.Hour * 12     // 12 hours
)

type RecvMsg func(*nats.Msg)

type NATSClient struct {
	nc     *nats.Conn
	compKV jetstream.KeyValue

	consumerMap map[string]bool

	brk Broker

	mutex sync.Mutex
	log   *logkf.Logger
}

func NewNATSClient(brk Broker) *NATSClient {
	return &NATSClient{
		consumerMap: make(map[string]bool),
		brk:         brk,
		log:         logkf.Global,
	}
}

func (c *NATSClient) Connect(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.nc != nil && c.nc.IsConnected() {
		c.log.Debug("nats client already connected")
		return nil
	}

	c.log.Debug("nats client connecting")

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

	if err := c.setupCompsKV(ctx); err != nil {
		return err
	}

	c.log.Info("nats client connected")
	return nil
}

func (c *NATSClient) setupCompsKV(ctx context.Context) (err error) {
	js, err := jetstream.New(c.nc)
	if err != nil {
		return c.log.ErrorN("connecting to JetStream failed: %v", err)
	}

	c.compKV, err = js.CreateKeyValue(ctx, jetstream.KeyValueConfig{
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

func (c *NATSClient) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.log.Info("nats client closing")
	if c.nc != nil {
		c.nc.Drain()
	}
}

func (c *NATSClient) ComponentsKV() jetstream.KeyValue {
	return c.compKV
}

func (c *NATSClient) Request(subject string, evt *kubefox.Event) error {
	msg, err := c.Msg(subject, evt)
	if err != nil {
		return err
	}

	resp, err := c.nc.RequestMsg(msg, evt.TTL())
	if err != nil {
		return err
	}

	c.handleMsg(resp)

	return nil
}

func (c *NATSClient) Publish(subject string, evt *kubefox.Event) error {
	msg, err := c.Msg(subject, evt)
	if err != nil {
		return err
	}

	c.log.Debugf("publishing nats msg to subj: %s", subject)

	return c.nc.PublishMsg(msg)
}

func (c *NATSClient) Msg(subject string, evt *kubefox.Event) (*nats.Msg, error) {
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

func (c *NATSClient) ConsumeEvents(ctx context.Context, name, subj string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, found := c.consumerMap[name]; found {
		return nil
	}

	c.log.Debugf("subscribing to nats; queue: %s, subj: %s", name, subj)
	sub, err := c.nc.QueueSubscribe(subj, name, c.handleMsg)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		c.log.Debugf("broker subscription context done, draining nats subscription '%s'", name)

		c.mutex.Lock()
		defer c.mutex.Unlock()

		if err := sub.Drain(); err != nil {
			c.log.Errorf("error draining nats subscription '%s': %v", name, err)
		}
		delete(c.consumerMap, name)
	}()

	return nil
}

func (c *NATSClient) handleMsg(msg *nats.Msg) {
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
		Receiver:   ReceiverNATS,
		ReceivedAt: time.Now(),
	}
	if err := c.brk.RecvEvent(lEvt); err != nil {
		log := c.log.WithEvent(evt)
		log.Debug(err)
		if evt.Target.Id == "" && errors.Is(err, kubefox.ErrSubCanceled) {
			// Any component replica can process, republish event.
			log.Debug("republishing event from component group subject")
			if err := c.nc.PublishMsg(msg); err != nil {
				log.Error("unable to republish event: %v", err)
			}
		}
	}
}

func (c *NATSClient) IsHealthy(ctx context.Context) bool {
	return c.nc != nil && c.nc.IsConnected()
}

func (c *NATSClient) Name() string {
	return natsSvcName
}
