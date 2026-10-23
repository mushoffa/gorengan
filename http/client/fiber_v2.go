package client

import (
	"context"
	"errors"
	"net/http"
	"time"

	fiberv2 "github.com/gofiber/fiber/v2"
)

var _ WebClient = (*fiberV2)(nil)

type fiberV2 struct {
	client *fiberv2.Client
	cfg    config
}

// NewFiberV2 returns a web client backed by the Fiber v2 client (fasthttp Agent).
func NewFiberV2(opts ...Option) WebClient {
	return &fiberV2{
		client: &fiberv2.Client{},
		cfg:    newConfig(opts),
	}
}

func (c *fiberV2) Get(ctx context.Context, path string, opts ...RequestOption) (*Response, error) {
	return c.do(ctx, http.MethodGet, path, nil, opts)
}

func (c *fiberV2) Delete(ctx context.Context, path string, opts ...RequestOption) (*Response, error) {
	return c.do(ctx, http.MethodDelete, path, nil, opts)
}

func (c *fiberV2) Post(ctx context.Context, path string, body any, opts ...RequestOption) (*Response, error) {
	return c.do(ctx, http.MethodPost, path, body, opts)
}

func (c *fiberV2) Put(ctx context.Context, path string, body any, opts ...RequestOption) (*Response, error) {
	return c.do(ctx, http.MethodPut, path, body, opts)
}

func (c *fiberV2) Patch(ctx context.Context, path string, body any, opts ...RequestOption) (*Response, error) {
	return c.do(ctx, http.MethodPatch, path, body, opts)
}

func (c *fiberV2) Close() {}

func (c *fiberV2) do(ctx context.Context, method, path string, body any, opts []RequestOption) (*Response, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	requestCfg := newRequestConfig(opts)
	requestURL, err := buildURL(c.cfg.baseURL, path, requestCfg.query)
	if err != nil {
		return nil, err
	}

	requestBody, contentType, err := encodeBody(body)
	if err != nil {
		return nil, err
	}

	agent := c.client.Get(requestURL) // creates the agent; the method is set below
	agent.Request().Header.SetMethod(method)
	for k, v := range c.cfg.headers {
		agent.Set(k, v)
	}

	if c.cfg.userAgent != "" {
		agent.UserAgent(c.cfg.userAgent)
	}

	if contentType != "" {
		agent.Set("Content-Type", contentType)
	}

	for k, v := range requestCfg.headers { // per-request headers win
		agent.Set(k, v)
	}

	if requestBody != nil {
		agent.Body(requestBody)
	}

	if t := resolveTimeout(ctx, c.cfg.timeout, requestCfg.timeout); t > 0 {
		agent.Timeout(t)
	}

	res := fiberv2.AcquireResponse()
	defer fiberv2.ReleaseResponse(res)
	agent.SetResponse(res)

	responseCode, responseBody, errs := agent.Bytes()
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	header := map[string][]string{}
	res.Header.VisitAll(func(k, v []byte) {
		key := string(k)
		header[key] = append(header[key], string(v))
	})
	return buildResponse(responseCode, header, append([]byte(nil), responseBody...)), nil
}

// resolveTimeout picks the per-request timeout, else the client default,
// then tightens it to the context deadline if that comes sooner.
func resolveTimeout(ctx context.Context, defaultTimeout, perRequest time.Duration) time.Duration {
	timeout := defaultTimeout
	if perRequest > 0 {
		timeout = perRequest
	}
	if deadline, ok := ctx.Deadline(); ok {
		if timeLeft := time.Until(deadline); timeLeft > 0 && (timeout == 0 || timeLeft < timeout) {
			timeout = timeLeft
		}
	}
	return timeout
}
