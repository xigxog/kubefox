package admin

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/xigxog/kubefox/libs/core/api/uri"
	"github.com/xigxog/kubefox/libs/core/logger"
)

type Action uint8

const (
	APIPath         = "api/kubefox/v0"
	JSONContentType = "application/json; charset=utf-8"
	TextContentType = "text/plain; charset=utf-8"
)

type Client interface {
	Ping() (*Response, error)
	Get(uri.URI, any) (*Response, error)
	List(uri.URI) (*Response, error)
	Create(uri.URI, any) (*Response, error)
	Apply(uri.URI, any) (*Response, error)
	Patch(uri.URI, any) (*Response, error)
	Delete(uri.URI) (*Response, error)
}

type ClientConfig struct {
	URL      string
	Timeout  time.Duration
	Insecure bool

	HTTPClient *http.Client

	Log *logger.Log
}

type client struct {
	baseURL    string
	httpClient *http.Client
	log        *logger.Log
}

func NewClient(cfg ClientConfig) Client {
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.Insecure},
			},
			Timeout: cfg.Timeout,
		}
	}

	if cfg.Log == nil {
		cfg.Log = logger.ProdLogger()
	}

	return &client{
		baseURL:    cfg.URL,
		httpClient: cfg.HTTPClient,
		log:        cfg.Log,
	}
}

func (c *client) Ping() (*Response, error) {
	return c.send("GET", "ping", nil)
}

func (c *client) Get(u uri.URI, obj any) (*Response, error) {
	return c.send("GET", u.Path(), obj)
}

func (c *client) List(u uri.URI) (*Response, error) {
	return c.send("GET", u.Path(), []string{})
}

func (c *client) Create(u uri.URI, obj any) (*Response, error) {
	return c.send("POST", u.Path(), obj)
}

func (c *client) Apply(u uri.URI, obj any) (*Response, error) {
	verb := "POST"
	if u.Kind() == uri.Platform || u.SubKind() == uri.Metadata || u.SubKind() == uri.Branch {
		verb = "PUT"
	}

	return c.send(verb, u.Path(), obj)
}

func (c *client) Patch(u uri.URI, obj any) (*Response, error) {
	return c.send("PATCH", u.Path(), obj)
}

func (c *client) Delete(u uri.URI) (*Response, error) {
	return c.send("DELETE", u.Path(), nil)
}

func (c *client) send(method string, path string, obj any) (*Response, error) {
	url := fmt.Sprintf("%s/%s/%s", c.baseURL, APIPath, path)

	var body io.Reader
	var contType string
	if method == "POST" || method == "PUT" || method == "PATCH" {
		var cont []byte
		if s, ok := obj.(string); ok {
			contType = TextContentType
			cont = []byte(s)

		} else {
			contType = JSONContentType
			if j, err := json.Marshal(obj); err != nil {
				c.log.Debugf("%s %s: %v", method, url, err)
				return nil, err
			} else {
				cont = j
			}
		}
		body = bytes.NewReader(cont)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		c.log.Debugf("%s %s: %v", method, url, err)
		return nil, err
	}
	req.Header.Add("Content-Type", contType)
	req.Header.Add("Accept", JSONContentType)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.log.Debugf("%s %s: %v", method, url, err)
		return nil, err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	adminResp := &Response{}
	if obj != nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		adminResp.Data = obj
	}
	if err = json.NewDecoder(resp.Body).Decode(adminResp); err != nil {
		c.log.Debugf("%s %s %s: %v", method, url, resp.Status, err)
		return nil, err
	}

	c.log.Debugf("%s %s %s", method, url, resp.Status)

	return adminResp, nil
}
