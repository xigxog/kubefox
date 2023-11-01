package grpc

import (
	context "context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	kubefox "github.com/xigxog/kubefox/core"

	"github.com/xigxog/kubefox/logkf"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type ClientOpts struct {
	Component     *kubefox.Component
	BrokerAddr    string
	HealthSrvAddr string
}

type Client struct {
	ClientOpts

	brk    Broker_SubscribeClient
	reqMap map[string]chan *kubefox.Event

	recvCh chan *kubefox.MatchedEvent
	errCh  chan error

	reqMapMutex sync.RWMutex
	sendMutex   sync.Mutex

	healthSrv *http.Server
	healthy   atomic.Bool

	log *logkf.Logger
}

func NewClient(opts ClientOpts) *Client {
	return &Client{
		ClientOpts: opts,
		reqMap:     make(map[string]chan *kubefox.Event),
		recvCh:     make(chan *kubefox.MatchedEvent),
		errCh:      make(chan error),
		log:        logkf.Global,
	}
}

func (c *Client) Start(spec *kubefox.ComponentSpec) error {
	creds, err := credentials.NewClientTLSFromFile(kubefox.PathCACert, "")
	if err != nil {
		return fmt.Errorf("unable to load root CA certificate: %v", err)
	}
	grpcCfg := `{
		"methodConfig": [{
		  "name": [{"service": "", "method": ""}],
		  "waitForReady": false,
		  "retryPolicy": {
			  "MaxAttempts": 3,
			  "InitialBackoff": "3s",
			  "MaxBackoff": "6s",
			  "BackoffMultiplier": 2.0,
			  "RetryableStatusCodes": [ "UNAVAILABLE" ]
		  }
		}]}`

	conn, err := gogrpc.Dial(c.BrokerAddr,
		gogrpc.WithPerRPCCredentials(c),
		gogrpc.WithTransportCredentials(creds),
		gogrpc.WithDefaultServiceConfig(grpcCfg),
	)
	if err != nil {
		return fmt.Errorf("unable to connect to broker: %v", err)
	}

	if c.brk, err = NewBrokerClient(conn).Subscribe(context.Background()); err != nil {
		return fmt.Errorf("subscribing to broker failed: %v", err)
	}

	evt := kubefox.NewMsg(kubefox.EventOpts{
		Type:   kubefox.EventTypeRegister,
		Source: c.Component,
	})
	if err := evt.SetJSON(spec); err != nil {
		return fmt.Errorf("unable to marshal component spec: %v", err)
	}
	if err := c.send(evt); err != nil {
		return fmt.Errorf("unable to register with broker: %v", err)
	}

	c.healthy.Store(true)
	c.log.Info("subscribed to broker")

	go func() {
		defer func() {
			if err := conn.Close(); err != nil {
				c.log.Error(err)
			}
		}()

		for {
			evt, err := c.brk.Recv()
			if err != nil {
				c.healthy.Store(false)
				c.errCh <- err
				return
			}

			switch evt.Event.Category {
			case kubefox.Category_REQUEST:
				go c.recvReq(evt)

			case kubefox.Category_RESPONSE:
				go c.recvResp(evt.Event)

			default:
				c.log.Debug("default")
			}
		}
	}()

	return nil
}

func (c *Client) Err() chan error {
	return c.errCh
}

func (c *Client) Req() chan *kubefox.MatchedEvent {
	return c.recvCh
}

func (c *Client) SendReq(ctx context.Context, req *kubefox.Event) (*kubefox.Event, error) {
	respCh, err := c.SendReqChan(req)
	if err != nil {
		return nil, err
	}

	select {
	case resp := <-respCh:
		return resp, nil

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *Client) SendReqChan(req *kubefox.Event) (chan *kubefox.Event, error) {
	c.log.WithEvent(req).Debug("send request")

	c.reqMapMutex.Lock()
	respCh := make(chan *kubefox.Event)
	c.reqMap[req.Id] = respCh
	c.reqMapMutex.Unlock()

	go func() {
		time.Sleep(req.TTL())

		c.reqMapMutex.Lock()
		delete(c.reqMap, req.Id)
		c.reqMapMutex.Unlock()
	}()

	if err := c.send(req); err != nil {
		return nil, err
	}

	return respCh, nil
}

func (c *Client) SendResp(resp *kubefox.Event) error {
	c.log.WithEvent(resp).Debug("send response")
	return c.send(resp)
}

func (c *Client) recvReq(req *kubefox.MatchedEvent) {
	c.log.WithEvent(req.Event).Debug("receive request")
	c.recvCh <- req
}

func (c *Client) recvResp(resp *kubefox.Event) {
	log := c.log.WithEvent(resp)
	log.Debug("receive response")

	c.reqMapMutex.RLock()
	respCh, found := c.reqMap[resp.ParentId]
	c.reqMapMutex.RUnlock()

	if !found {
		log.Error("request for response not found")
		return
	}

	respCh <- resp
}

func (c *Client) send(evt *kubefox.Event) error {
	// Need to protect the stream from being called by multiple threads.
	c.sendMutex.Lock()
	defer c.sendMutex.Unlock()

	return c.brk.Send(evt)
}

func (c *Client) StartHealthSrv() error {
	if c.HealthSrvAddr == "" || c.HealthSrvAddr == "false" {
		return nil
	}

	c.healthSrv = &http.Server{
		WriteTimeout: time.Second * 3,
		ReadTimeout:  time.Second * 3,
		IdleTimeout:  time.Second * 30,
		Handler:      c,
	}

	ln, err := net.Listen("tcp", c.HealthSrvAddr)
	if err != nil {
		return fmt.Errorf("unable to open tcp socket for health server: %v", err)
	}

	go func() {
		err := c.healthSrv.Serve(ln)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			c.log.Fatal(err)
		}
	}()

	c.log.Info("health server started")
	return nil
}

func (c *Client) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	status := http.StatusOK
	if !c.healthy.Load() {
		status = http.StatusServiceUnavailable
	}
	resp.WriteHeader(status)
}

func (c *Client) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	b, err := os.ReadFile(kubefox.PathSvcAccToken)
	if err != nil {
		return nil, err
	}
	token := string(b)

	return map[string]string{
		"componentId":     c.Component.Id,
		"componentName":   c.Component.Name,
		"componentCommit": c.Component.Commit,
		"authToken":       token,
	}, nil
}

func (svc *Client) RequireTransportSecurity() bool {
	return true
}
