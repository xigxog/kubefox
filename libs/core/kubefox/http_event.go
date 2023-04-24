package kubefox

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/lestrrat-go/jwx/jwt"
	"github.com/xigxog/kubefox/libs/core/api/uri"
	"github.com/xigxog/kubefox/libs/core/component"
	"github.com/xigxog/kubefox/libs/core/grpc"
	"github.com/xigxog/kubefox/libs/core/utils"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

type HTTPEvent interface {
	Event

	GetURL() *url.URL
	GetURLString() string
	GetPath() string

	GetStatus() string
	SetStatus(v string)
	GetStatusCode() int
	SetStatusCode(v int)
	GetMethod() string
	SetMethod(method string)

	GetHeaderKeys() []string
	GetHeader(string) string
	GetHeaderValues(string) []string
	AddHeader(string, string)
	SetHeader(string, string)
	GetHTTPHeader() http.Header
	SetHTTPHeader(http.Header)
}

type HTTPDataEvent interface {
	DataEvent
	HTTPEvent

	ParseRequest(*http.Request) error
	ParseResponse(*http.Response) error

	GetHTTPRequest(context.Context) (*http.Request, error)
	GetHTTPResponse() *http.Response

	SetURL(*url.URL)
	SetURLString(string) error
}

type httpEvent struct {
	*event

	url    *url.URL
	header http.Header
}

func (evt *httpEvent) ParseRequest(httpReq *http.Request) error {
	if evt.GetContext() == nil {
		evt.SetContext(&grpc.EventContext{})
	}

	if evt.GetContext().Environment == "" {
		env := utils.GetParamOrHeader(httpReq, EnvHeader, EnvHeaderShort, RelEnvHeader)
		env = strings.ReplaceAll(env, uri.HTTPSeparator, uri.PathSeparator)
		if _, err := uri.New(uri.Authority, uri.Environment, env); err != nil {
			return fmt.Errorf("invalid environment: %w", err)
		}
		evt.GetContext().Environment = env
	}

	if evt.GetContext().System == "" {
		sys := utils.GetParamOrHeader(httpReq, SysHeader, SysHeaderShort, RelSysHeader)
		sys = strings.ReplaceAll(sys, uri.HTTPSeparator, uri.PathSeparator)
		if _, err := uri.New(uri.Authority, uri.System, sys); err != nil {
			return fmt.Errorf("invalid system: %w", err)
		}
		evt.GetContext().System = sys
	}

	if evt.GetTarget() == nil {
		targetURI := "kubefox:component:" + utils.GetParamOrHeader(httpReq, TargetHeader, RelTargetHeader)
		target, err := component.ParseURI(targetURI)
		if err != nil {
			return fmt.Errorf("invalid target: %w", err)
		}

		evt.SetTarget(target)
		if evt.GetContext().App == "" {
			evt.GetContext().App = target.GetApp()
		}
	}

	if evt.GetType() == "" || evt.GetType() == UnknownEventType {
		if evtType := utils.GetParamOrHeader(httpReq, EventTypeHeader, RelEventTypeHeader); evtType != "" {
			evt.SetType(evtType)
		} else {
			evt.SetType(HTTPRequestType)
		}
	}

	if httpReq.Header.Get("Authorization") != "" {
		// TODO verify token signature
		// validate ensures expected public claims are present and valid but
		// does not verify the token's signature
		token, err := jwt.ParseRequest(httpReq, jwt.WithValidate(true))
		if err != nil {
			return err
		}
		evt.SetToken(token)
	}

	url := httpReq.URL
	if host := httpReq.Header.Get("X-Forwarded-Host"); host != "" {
		if port := httpReq.Header.Get("X-Forwarded-Port"); port != "" {
			url.Host = fmt.Sprintf("%s:%s", host, port)
		} else {
			url.Host = host
		}
		if scheme := httpReq.Header.Get("X-Forwarded-Proto"); scheme != "" {
			url.Scheme = scheme
		}
	}
	if url.Scheme == "" {
		url.Scheme = "http"
		if httpReq.TLS != nil {
			url.Scheme = "https"
		}
	}
	if url.Host == "" {
		url.Host = httpReq.Host
	}
	evt.SetURL(url)

	contentType := utils.GetParamOrHeader(httpReq, "Content-Type")
	evt.SetContentType(contentType)

	content, err := io.ReadAll(httpReq.Body)
	if err != nil {
		return err
	}
	evt.SetContent(content)

	evt.SetMethod(httpReq.Method)
	evt.SetHTTPHeader(httpReq.Header)

	return nil
}

func (evt *httpEvent) ParseResponse(httpResp *http.Response) error {
	defer httpResp.Body.Close()
	content, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}
	evt.SetContent(content)
	evt.SetContentType(httpResp.Header.Get("Content-Type"))

	if evt.GetType() == "" || evt.GetType() == UnknownEventType {
		evt.SetType(HTTPResponseType)
	}

	evt.SetStatus(httpResp.Status)
	evt.SetStatusCode(httpResp.StatusCode)
	evt.SetHTTPHeader(httpResp.Header)

	return nil
}

func (evt *httpEvent) GetStatus() string {
	return evt.GetValue("http_status")
}

func (evt *httpEvent) SetStatus(v string) {
	evt.SetValue("http_status", v)
}

func (evt *httpEvent) GetStatusCode() int {
	return evt.GetValueVar("http_status_code").IntOrDefault(http.StatusOK)
}

func (evt *httpEvent) SetStatusCode(v int) {
	evt.SetValueNumber("http_status_code", float64(v))
}

func (evt *httpEvent) GetMethod() string {
	return evt.GetValue("http_method")
}

func (evt *httpEvent) SetMethod(method string) {
	evt.SetValue("http_method", method)
}

func (evt *httpEvent) GetURLString() string {
	url := evt.getURL()
	if url == nil {
		return ""
	}

	return url.String()
}

func (evt *httpEvent) SetURLString(urlStr string) error {
	if u, err := url.Parse(urlStr); err != nil {
		return err
	} else {
		evt.SetURL(u)
	}

	return nil
}

// GetURL returns a copy of the event's URL. If a URL does not exist nil is
// returned. Note that a copy is returned. Changes made to the returned URL
// will not affect the event's URL.
func (evt *httpEvent) GetURL() *url.URL {
	return copyURL(evt.getURL())
}

func (evt *httpEvent) getURL() *url.URL {
	if evt.url != nil {
		return evt.url
	}

	evt.mutex.Lock()
	defer evt.mutex.Unlock()

	u, err := url.Parse(evt.GetValue("http_url"))
	if err != nil {
		return nil
	}
	evt.url = u

	return evt.url
}

// SetURL sets the event's URL to a copy of the passed URL. Note that a copy is
// set. Changes made to the passed URL will not affect the event's URL.
func (evt *httpEvent) SetURL(url *url.URL) {
	evt.mutex.Lock()
	defer evt.mutex.Unlock()

	if url != nil {
		evt.SetValue("http_url", url.String())
	} else {
		evt.SetValueProto("http_url", nil)
	}
	evt.url = copyURL(url)
}

func (evt *httpEvent) GetPath() string {
	url := evt.getURL()
	if url == nil {
		return ""
	}

	return url.Path
}

func (evt *httpEvent) GetHeaderKeys() []string {
	m := evt.headerMap(false)
	k := make([]string, len(m))
	for key := range m {
		k = append(k, key)
	}

	return k
}

// Header gets the first value associated with the given key. If there are no
// values associated with the key, Header returns "". It is case insensitive;
// keys are normalized to lowercasevt.
func (evt *httpEvent) GetHeader(key string) string {
	if v := evt.headerList(key, false).GetValues(); len(v) > 0 {
		return v[0].GetStringValue()
	}

	return ""
}

// HeaderValues returns all values associated with the given key. It is case
// insensitive; keys are normalized to lowercasevt.
func (evt *httpEvent) GetHeaderValues(key string) []string {
	li := evt.headerList(key, false).GetValues()
	v := make([]string, len(li))
	for _, h := range li {
		v = append(v, h.GetStringValue())
	}

	return v
}

// AddHeader adds the key, value pair to the header. It appends to any existing
// values associated with key. The key is case insensitive; it is canonicalized
// by CanonicalHeaderKey.
func (evt *httpEvent) AddHeader(key, v string) {
	evt.mutex.Lock()
	defer evt.mutex.Unlock()
	evt.header = nil

	li := evt.headerList(key, true)
	li.Values = append(li.Values, structpb.NewStringValue(v))
}

// SetHeader sets the header entries associated with key to the single element
// valuevt. It replaces any existing values associated with key. The key is case
// insensitive; keys are normalized to lowercasevt.
func (evt *httpEvent) SetHeader(key, v string) {
	evt.mutex.Lock()
	defer evt.mutex.Unlock()
	evt.header = nil

	li := evt.headerList(key, true)
	li.Values = []*structpb.Value{structpb.NewStringValue(v)}

}

func (evt *httpEvent) GetHTTPHeader() http.Header {
	evt.mutex.Lock()
	defer evt.mutex.Unlock()

	header := make(http.Header)
	for k, v := range evt.headerMap(false) {
		for _, h := range v.GetListValue().Values {
			evt.header.Add(k, h.GetStringValue())
		}
	}

	return header
}

// ? maybe make a close method that is called by sdk which pushes down to data
func (evt *httpEvent) SetHTTPHeader(header http.Header) {
	evt.mutex.Lock()
	defer evt.mutex.Unlock()

	headerMap := make(map[string]interface{}, len(header))
	for k, v := range header {
		l := make([]interface{}, len(v))
		for i, h := range v {
			l[i] = h
		}
		headerMap[strings.ToLower(k)] = l
	}

	v, _ := structpb.NewValue(headerMap)
	evt.SetValueProto("http_header", v)
	evt.header = header
}

func (evt *httpEvent) headerMap(create bool) map[string]*structpb.Value {
	if v := evt.GetValueProto("http_header"); v == nil && create {
		evt.SetValueProto("http_header", structpb.NewStructValue(nil))
	}

	return evt.GetValueProto("http_header").GetStructValue().GetFields()
}

func (evt *httpEvent) headerList(key string, create bool) *structpb.ListValue {
	key = strings.ToLower(key)
	fields := evt.headerMap(create)
	// it is not possible for fields to be null if create is true
	if fields == nil {
		return nil
	}

	v := fields[key].GetListValue()
	if v == nil && create {
		v, _ := structpb.NewList(nil)
		fields[key] = structpb.NewListValue(v)
	}

	return v
}

func (evt *httpEvent) GetHTTPRequest(ctx context.Context) (*http.Request, error) {
	body := bytes.NewReader(evt.GetContent())
	req, err := http.NewRequestWithContext(ctx, evt.GetMethod(), evt.GetURL().String(), body)
	if err != nil {
		return nil, err
	}
	req.Header = evt.GetHTTPHeader()

	return req, nil
}

func (evt *httpEvent) GetHTTPResponse() *http.Response {
	httpRes := &http.Response{
		Status:     evt.GetStatus(),
		StatusCode: evt.GetStatusCode(),
		Header:     evt.GetHTTPHeader(),
		Body:       io.NopCloser(bytes.NewReader(evt.GetContent())),
	}

	return httpRes
}

func copyURL(url *url.URL) *url.URL {
	if url == nil {
		return nil
	}

	urlCopy := *url
	if url.User != nil {
		userCopy := *url.User
		urlCopy.User = &userCopy
	}

	return &urlCopy
}
