package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/components/broker/config"
	"github.com/xigxog/kubefox/core"
	"github.com/xigxog/kubefox/logkf"
	"google.golang.org/protobuf/proto"
)

const (
	CloudEventId = "ce_id"
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
	nc *nats.Conn

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
		nats.RootCAs(api.PathCACert),
		nats.ClientCert(api.PathTLSCert, api.PathTLSKey),
	)
	if err != nil {
		return c.log.ErrorN("connecting to NATS failed: %v", err)
	}

	c.log.Info("nats client connected")
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

func (c *NATSClient) Request(subject string, evt *core.Event) error {
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

func (c *NATSClient) Publish(subject string, evt *core.Event) error {
	msg, err := c.Msg(subject, evt)
	if err != nil {
		return err
	}

	c.log.Debugf("publishing nats msg to subj: %s", subject)

	return c.nc.PublishMsg(msg)
}

func (c *NATSClient) Msg(subject string, evt *core.Event) (*nats.Msg, error) {
	dataBytes, err := proto.Marshal(evt)
	if err != nil {
		return nil, err
	}

	h := make(nats.Header)
	h.Set(CloudEventId, evt.Id)

	// Headers create sizeable overhead for small msgs. Disabling most for now.
	//
	// h.Set("ce_specversion", "1.0")
	// h.Set("ce_type", evt.Type)
	// h.Set("ce_time", time.Now().Format(time.RFC3339))
	// h.Set("ce_source", fmt.Sprintf("kubefox:component:%s", evt.Source.Key()))
	// h.Set("ce_dataschema", core.DataSchemaKubefox)
	// h.Set("ce_datacontenttype", core.ContentTypeProtobuf)
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

	evt := core.NewEvent()
	if err := proto.Unmarshal(msg.Data, evt); err != nil {
		evtId := msg.Header.Get(CloudEventId)
		c.log.With(logkf.KeyEventId, evtId).Warn("message contains invalid event data: %v", err)
		return
	}
	if md, err := msg.Metadata(); err == nil { // success
		evt.ReduceTTL(md.Timestamp)
	}

	c.brk.RecvEvent(evt, ReceiverNATS)
}

func (c *NATSClient) IsHealthy(ctx context.Context) bool {
	return c.nc != nil && c.nc.IsConnected()
}

func (c *NATSClient) Name() string {
	return natsSvcName
}
