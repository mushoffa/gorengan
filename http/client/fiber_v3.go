package client

import (
	"context"
	"fmt"
	"net/http"

	fiberv3 "github.com/gofiber/fiber/v3/client"
)

var _ WebClient = (*fiberV3)(nil)

type fiberV3 struct {
	client *fiberv3.Client
}

// NewFiberV3 returns a Fiber v3 client.
func NewFiberV3(opts ...Option) WebClient {
	cfg := newConfig(opts)

	fiberClient := fiberv3.New()
	if cfg.baseURL != "" {
		fiberClient.SetBaseURL(cfg.baseURL)
	}
	if cfg.userAgent != "" {
		fiberClient.SetUserAgent(cfg.userAgent)
	}
	if cfg.timeout > 0 {
		fiberClient.SetTimeout(cfg.timeout)
	}
	for k, v := range cfg.headers {
		fiberClient.SetHeader(k, v)
	}
	return &fiberV3{
		client: fiberClient,
	}
}

func (c *fiberV3) Get(ctx context.Context, path string, opts ...RequestOption) (*Response, error) {
	return c.do(ctx, http.MethodGet, path, nil, opts)
}

func (c *fiberV3) Post(ctx context.Context, path string, body any, opts ...RequestOption) (*Response, error) {
	return c.do(ctx, http.MethodPost, path, body, opts)
}

func (c *fiberV3) Put(ctx context.Context, path string, body any, opts ...RequestOption) (*Response, error) {
	return c.do(ctx, http.MethodPut, path, body, opts)
}

func (c *fiberV3) Patch(ctx context.Context, path string, body any, opts ...RequestOption) (*Response, error) {
	return c.do(ctx, http.MethodPatch, path, body, opts)
}

func (c *fiberV3) Delete(ctx context.Context, path string, opts ...RequestOption) (*Response, error) {
	return c.do(ctx, http.MethodDelete, path, nil, opts)
}

func (c *fiberV3) Close() {
	c.client.CloseIdleConnections()
}

func (c *fiberV3) do(ctx context.Context, method, path string, body any, opts []RequestOption) (*Response, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	requestCfg := newRequestConfig(opts)

	cfg := fiberv3.Config{
		Ctx:     ctx,
		Header:  requestCfg.headers,
		Param:   requestCfg.query,
		Body:    body,
		Timeout: requestCfg.timeout,
	}

	res, err := c.send(method, path, cfg)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	header := make(map[string][]string)
	for k, v := range res.Headers() {
		header[k] = append([]string(nil), v...)
	}
	return buildResponse(res.StatusCode(), header, append([]byte(nil), res.Body()...)), nil
}

// send dispatches to the Fiber client method that matches the HTTP method.
func (c *fiberV3) send(method, path string, cfg fiberv3.Config) (*fiberv3.Response, error) {
	switch method {
	case http.MethodGet:
		return c.client.Get(path, cfg)
	case http.MethodPost:
		return c.client.Post(path, cfg)
	case http.MethodPut:
		return c.client.Put(path, cfg)
	case http.MethodPatch:
		return c.client.Patch(path, cfg)
	case http.MethodDelete:
		return c.client.Delete(path, cfg)
	default:
		return nil, fmt.Errorf("client: unsupported method %q", method)
	}
}
