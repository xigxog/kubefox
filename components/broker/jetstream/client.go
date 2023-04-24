package jetstream

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/xigxog/kubefox/components/broker/config"
	"github.com/xigxog/kubefox/libs/core/kubefox"
	"github.com/xigxog/kubefox/libs/core/logger"
	"github.com/xigxog/kubefox/libs/core/platform"
	"github.com/xigxog/kubefox/libs/core/utils"
	"google.golang.org/protobuf/proto"
)

// Content types
const (
	DataSchemaV1        = "kubefox.proto.v1.KubeFoxData"
	ProtobufContentType = "application/protobuf"
	TLSCertFile         = "/kubefox/nats/tls/tls.crt"
	TLSKeyFile          = "/kubefox/nats/tls/tls.key"
	CACertFile          = "/kubefox/nats/tls/ca.crt"
)

type Client struct {
	cfg *config.Config

	nc    *nats.Conn
	jsCtx nats.JetStreamContext
	subs  []Subscription

	mutex sync.Mutex

	log *logger.Log
}

func NewClient(cfg *config.Config, log *logger.Log) *Client {
	return &Client{
		cfg:  cfg,
		subs: make([]Subscription, 0),
		log:  log,
	}
}

func (c *Client) Connect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.nc != nil && c.nc.IsConnected() {
		return nil
	}

	nc, err := nats.Connect(
		fmt.Sprintf("nats://%s", c.cfg.NatsAddr),
		nats.Name(c.cfg.Comp.GetURI()),
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(3),
		c.natsTLS(c.cfg.Namespace),
	)
	if err != nil {
		c.log.Error(err)
		return err
	}
	c.nc = nc

	jsCtx, err := nc.JetStream()
	if err != nil {
		c.log.Error(err)
		nc.Close()
		return err
	}
	c.jsCtx = jsCtx

	_, err = jsCtx.AddStream(&nats.StreamConfig{
		Name:        c.cfg.Comp.GetStream(),
		Description: fmt.Sprintf("Durable queues for requests and responses of the %s component.", c.cfg.Comp.GetName()),
		Subjects:    []string{c.cfg.Comp.GetSubjectWildcard()},
		Storage:     nats.MemoryStorage,
		MaxMsgSize:  1024 * 1024 * 5, // 5 MiB, TODO configurable
		Retention:   nats.LimitsPolicy,
		Discard:     nats.DiscardOld,
		MaxAge:      time.Hour * 24 * 3, // 3 days, TODO configurable
	})
	if err != nil {
		c.log.Error(err)
		nc.Close()
		return err
	}

	c.log.Infof("connected to JetStream at %s", c.cfg.NatsAddr)

	return nil
}

func (c *Client) natsTLS(namespace string) nats.Option {
	return func(o *nats.Options) error {
		var err error
		var pool *x509.CertPool
		var cert tls.Certificate

		if pem, err := os.ReadFile(CACertFile); err == nil {
			c.log.Debugf("reading tls certs from file")
			pool = x509.NewCertPool()
			ok := pool.AppendCertsFromPEM(pem)
			if !ok {
				return fmt.Errorf("failed to parse root certificate from %s", CACertFile)
			}

			cert, err = tls.LoadX509KeyPair(TLSCertFile, TLSKeyFile)
			if err != nil {
				return fmt.Errorf("nats: error loading client certificate: %v", err)
			}

		} else {
			c.log.Debugf("reading tls certs from kubernetes secret")
			cert, pool, err = utils.GetCertFromSecret(namespace, platform.NATSCertSecret)
			if err != nil {
				return err
			}
		}

		cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			return err
		}

		if o.TLSConfig == nil {
			o.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		}
		o.TLSConfig.Certificates = []tls.Certificate{cert}
		o.TLSConfig.RootCAs = pool
		o.Secure = true

		return nil
	}
}

func (c *Client) Close() {
	for _, sub := range c.subs {
		sub.Close()
	}

	if c.nc != nil {
		c.nc.Close()
	}
}

func (c *Client) Healthy(ctx context.Context) bool {
	return c.nc != nil && c.nc.IsConnected()
}

func (c *Client) Name() string {
	return "jetstream-client"
}

func (c *Client) Publish(subject string, evt kubefox.DataEvent) (*nats.PubAck, error) {
	c.log.With("traceId", evt.GetTraceId()).
		Debugf("publishing data; id: %s, subject: %s", evt.GetId(), subject)

	dataBytes, err := proto.Marshal(evt.GetData())
	if err != nil {
		c.log.Error(err)
		return nil, err
	}

	h := make(nats.Header)
	h.Set("Nats-Msg-Id", evt.GetId())
	h.Set("ce_specversion", "1.0")
	h.Set("ce_type", evt.GetType())
	h.Set("ce_id", evt.GetId())
	h.Set("ce_time", time.Now().Format(time.RFC3339))
	h.Set("ce_source", evt.GetSource().GetURI())
	h.Set("ce_dataschema", DataSchemaV1)
	h.Set("ce_datacontenttype", ProtobufContentType)

	return c.jsCtx.PublishMsg(&nats.Msg{
		Subject: subject,
		Header:  h,
		Data:    dataBytes,
	})
}

func (c *Client) Subscribe(cfg *SubscriptionConfig) (Subscription, error) {
	nSub, err := c.jsCtx.PullSubscribe(cfg.Subject, cfg.Consumer)
	if err != nil {
		return nil, err
	}

	sub := NewSubscription(cfg, nSub, c.log)
	c.subs = append(c.subs, sub)
	c.log.Infof("subscription created; worker: %s.%s, subject: %s",
		cfg.Worker, cfg.Consumer, cfg.Subject)

	return sub, nil
}
