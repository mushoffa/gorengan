package client

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

var _ WebClient = (*netHttp)(nil)

type netHttp struct {
	client *http.Client
	cfg    config
}

// NewNetHttp returns a web client backed by net/http.
func NewNetHttp(opts ...Option) WebClient {
	cfg := newConfig(opts)

	return &netHttp{
		client: &http.Client{
			Transport: newTransport(cfg),
			Timeout:   cfg.timeout,
		},
		cfg: cfg,
	}
}

func (c *netHttp) Get(ctx context.Context, path string, opts ...RequestOption) (*Response, error) {
	return c.do(ctx, http.MethodGet, path, nil, opts)
}

func (c *netHttp) Delete(ctx context.Context, path string, opts ...RequestOption) (*Response, error) {
	return c.do(ctx, http.MethodDelete, path, nil, opts)
}

func (c *netHttp) Post(ctx context.Context, path string, body any, opts ...RequestOption) (*Response, error) {
	return c.do(ctx, http.MethodPost, path, body, opts)
}

func (c *netHttp) Put(ctx context.Context, path string, body any, opts ...RequestOption) (*Response, error) {
	return c.do(ctx, http.MethodPut, path, body, opts)
}

func (c *netHttp) Patch(ctx context.Context, path string, body any, opts ...RequestOption) (*Response, error) {
	return c.do(ctx, http.MethodPatch, path, body, opts)
}

func (c *netHttp) Close() {
	c.client.CloseIdleConnections()
}

func (c *netHttp) do(ctx context.Context, method, path string, body any, opts []RequestOption) (*Response, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	requestCfg := newRequestConfig(opts)

	if requestCfg.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, requestCfg.timeout)
		defer cancel()
	}

	requestURL, err := buildURL(c.cfg.baseURL, path, requestCfg.query)
	if err != nil {
		return nil, err
	}

	requestBody, contentType, err := encodeBody(body)
	if err != nil {
		return nil, err
	}

	var reader io.Reader
	if requestBody != nil {
		reader = bytes.NewReader(requestBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL, reader)
	if err != nil {
		return nil, err
	}
	for k, v := range c.cfg.headers {
		req.Header.Set(k, v)
	}
	if c.cfg.userAgent != "" {
		req.Header.Set("User-Agent", c.cfg.userAgent)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for k, v := range requestCfg.headers { // per-request headers win
		req.Header.Set(k, v)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	header := make(map[string][]string, len(res.Header))
	for k, v := range res.Header {
		header[k] = append([]string(nil), v...)
	}
	return buildResponse(res.StatusCode, header, responseBody), nil
}

// newTransport clones http.DefaultTransport so each client has its own
// connection pool: closing one client's idle connections must not affect
// every other user of the default transport.
func newTransport(cfg config) *http.Transport {
	if defaultTransport, ok := http.DefaultTransport.(*http.Transport); ok {
		return defaultTransport.Clone()
	}
	return &http.Transport{Proxy: http.ProxyFromEnvironment}
}
