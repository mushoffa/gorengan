package client

import (
	"time"
)

// config holds the web client settings
type config struct {
	baseURL   string
	userAgent string
	timeout   time.Duration
	headers   map[string]string
}

type Options []Option

// Option configures a web client.
type Option func(*config)

func newConfig(opts Options) config {
	// Set default timeout and headers config
	cfg := config{
		timeout: 30 * time.Second,
		headers: map[string]string{},
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

// WithBaseURL sets a base URL that is prepended to relative request URLs.
func WithBaseURL(url string) Option {
	return func(c *config) {
		c.baseURL = url
	}
}

// WithUserAgent sets the User-Agent header.
func WithUserAgent(userAgent string) Option {
	return func(c *config) {
		c.userAgent = userAgent
	}
}

// WithDefaultTimeout sets the default timeout for every request (30s if unset).
func WithDefaultTimeout(timeout time.Duration) Option {
	return func(c *config) {
		c.timeout = timeout
	}
}

// WithDefaultHeader sets a header that is sent with every request.
func WithDefaultHeader(key, value string) Option {
	return func(c *config) {
		c.headers[key] = value
	}
}

type RequestOptions []RequestOption

// RequestOption customises a single client call.
type RequestOption func(*requestConfig)

type requestConfig struct {
	headers map[string]string
	query   map[string]string
	timeout time.Duration
}

func newRequestConfig(opts RequestOptions) requestConfig {
	var config requestConfig
	for _, opt := range opts {
		opt(&config)
	}
	return config
}

// WithRequestHeader sets a header on one request. It overrides default headers.
func WithRequestHeader(key, value string) RequestOption {
	return func(c *requestConfig) {
		if c.headers == nil {
			c.headers = map[string]string{}
		}
		c.headers[key] = value
	}
}

// WithRequestQuery adds a query-string parameter to one request.
func WithRequestQuery(key, value string) RequestOption {
	return func(c *requestConfig) {
		if c.query == nil {
			c.query = map[string]string{}
		}
		c.query[key] = value
	}
}

// WithRequestTimeout overrides the client timeout for one request.
func WithRequestTimeout(timeout time.Duration) RequestOption {
	return func(c *requestConfig) {
		c.timeout = timeout
	}
}
