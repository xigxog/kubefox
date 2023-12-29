package grpc

import (
	context "context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xigxog/kubefox/api"
	core "github.com/xigxog/kubefox/core"

	"github.com/xigxog/kubefox/logkf"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type ClientOpts struct {
	Component     *core.Component
	BrokerAddr    string
	HealthSrvAddr string
}

type Client struct {
	ClientOpts

	brk    Broker_SubscribeClient
	reqMap map[string]*ActiveReq

	recvCh chan *ComponentEvent
	errCh  chan error

	reqMapMutex sync.RWMutex
	sendMutex   sync.Mutex

	healthSrv *http.Server
	healthy   atomic.Bool

	log *logkf.Logger
}

type ComponentEvent struct {
	*core.MatchedEvent

	ReceivedAt time.Time
}

type ActiveReq struct {
	respCh     chan *core.Event
	expiration time.Time
}

func NewClient(opts ClientOpts) *Client {
	c := &Client{
		ClientOpts: opts,
		reqMap:     make(map[string]*ActiveReq),
		recvCh:     make(chan *ComponentEvent),
		errCh:      make(chan error),
		log:        logkf.Global,
	}
	go c.startReqMapReaper()

	return c
}

// Start connects to the broker and begins sending and receiving messages. It is
// a blocking call.
func (c *Client) Start(def *api.ComponentDefinition, maxAttempts int) {
	var (
		attempt int
		err     error
	)
	for attempt < maxAttempts {
		c.log.Infof("subscribing to broker, attempt %d/%d", attempt+1, maxAttempts)

		attempt, err = c.run(def, attempt)

		c.log.Warnf("broker subscription closed: %v", err)
		time.Sleep(time.Second * time.Duration(rand.Intn(2)+1))
	}

	c.errCh <- err
	close(c.errCh)
}

func (c *Client) run(def *api.ComponentDefinition, retry int) (int, error) {
	creds, err := credentials.NewClientTLSFromFile(api.PathCACert, "")
	if err != nil {
		return retry + 1, fmt.Errorf("unable to load root CA certificate: %v", err)
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
		return retry + 1, fmt.Errorf("unable to connect to broker: %v", err)
	}

	defer func() {
		c.healthy.Store(false)
		if err := conn.Close(); err != nil {
			c.log.Error(err)
		}
	}()

	if c.brk, err = NewBrokerClient(conn).Subscribe(context.Background()); err != nil {
		return retry + 1, fmt.Errorf("subscribing to broker failed: %v", err)
	}

	evt := core.NewReq(core.EventOpts{
		Type:   api.EventTypeRegister,
		Source: c.Component,
	})
	if err := evt.SetJSON(def); err != nil {
		return retry + 1, fmt.Errorf("unable to marshal component spec: %v", err)
	}
	if err := c.send(evt, time.Now()); err != nil {
		return retry + 1, err
	}
	if _, err := c.brk.Recv(); err != nil {
		// TODO deal with redirect when broker removed from host network
		return retry + 1, err
	}

	c.healthy.Store(true)
	c.log.Info("subscribed to broker")

	for {
		evt, err := c.brk.Recv()
		if err != nil {
			// Reset retry count after successful connection.
			return 0, err
		}

		if c.brk.Context().Err() != nil {
			// Reset retry count after successful connection.
			return 0, err
		}

		switch evt.Event.Category {
		case core.Category_REQUEST:
			go c.recvReq(evt)

		case core.Category_RESPONSE:
			go c.recvResp(evt.Event)

		default:
			c.log.WithEvent(evt.Event).Debug("received event on unexpected category, dropping")
		}
	}
}

func (c *Client) Err() chan error {
	return c.errCh
}

func (c *Client) Req() chan *ComponentEvent {
	return c.recvCh
}

func (c *Client) SendReq(ctx context.Context, req *core.Event, start time.Time) (*core.Event, error) {
	respCh, err := c.SendReqChan(req, start)
	if err != nil {
		return nil, err
	}

	select {
	case resp := <-respCh:
		return resp, resp.Err()

	case <-ctx.Done():
		return nil, core.ErrTimeout(err)

	case <-c.brk.Context().Done():
		return nil, core.ErrBrokerUnavailable(err)
	}
}

func (c *Client) SendReqChan(req *core.Event, start time.Time) (chan *core.Event, error) {
	c.log.WithEvent(req).Debug("send request")

	respCh := make(chan *core.Event)
	c.reqMapMutex.Lock()
	c.reqMap[req.Id] = &ActiveReq{
		respCh:     respCh,
		expiration: time.Now().Add(req.TTL()),
	}
	c.reqMapMutex.Unlock()

	if err := c.send(req, start); err != nil {
		return nil, err
	}

	return respCh, nil
}

func (c *Client) SendResp(resp *core.Event, start time.Time) error {
	c.log.WithEvent(resp).Debug("send response")
	return c.send(resp, start)
}

func (c *Client) recvReq(req *core.MatchedEvent) {
	c.log.WithEvent(req.Event).Debug("receive request")
	c.recvCh <- &ComponentEvent{MatchedEvent: req, ReceivedAt: time.Now()}
}

func (c *Client) recvResp(resp *core.Event) {
	log := c.log.WithEvent(resp)
	log.Debug("receive response")

	c.reqMapMutex.Lock()
	respCh := c.reqMap[resp.ParentId]
	delete(c.reqMap, resp.ParentId)
	c.reqMapMutex.Unlock()

	if respCh == nil {
		log.Warn("request for response not found")
		return
	}

	respCh.respCh <- resp
}

func (c *Client) send(evt *core.Event, start time.Time) error {
	// Need to protect the stream from being called by multiple threads.
	c.sendMutex.Lock()
	defer c.sendMutex.Unlock()

	evt.ReduceTTL(start)
	if evt.TTL() < 0 {
		return core.ErrTimeout()
	}

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
	b, err := os.ReadFile(api.PathSvcAccToken)
	if err != nil {
		return nil, err
	}
	token := string(b)

	return map[string]string{
		"componentId":     c.Component.Id,
		"componentName":   c.Component.Name,
		"componentCommit": c.Component.Commit,
		"componentType":   c.Component.Type,
		"authToken":       token,
	}, nil
}

func (c *Client) RequireTransportSecurity() bool {
	return true
}

func (c *Client) startReqMapReaper() {
	log := c.log.With(logkf.KeyWorker, "request-map-reaper")
	defer func() {
		log.Info("request-map-reaper")
	}()

	ticker := time.NewTicker(time.Second * 30)
	for range ticker.C {
		log.Debugf("reaping request map of size %d", len(c.reqMap))
		c.reqMapMutex.RLock()
		// Add a 5 second buffer to expiration.
		now := time.Now().Add(time.Second * -30)
		for k, v := range c.reqMap {
			// If request has expired delete it.
			if now.After(v.expiration) {
				c.reqMapMutex.RUnlock()
				c.reqMapMutex.Lock()
				log.Debugf("request '%s' expired, deleting", k)
				delete(c.reqMap, k)
				c.reqMapMutex.Unlock()
				c.reqMapMutex.RLock()
			}
		}
		c.reqMapMutex.RUnlock()
	}
}
