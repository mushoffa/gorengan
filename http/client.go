package http

import (
	"github.com/mushoffa/gorengan/http/client"
)

// WebClient options.
var (
	WithBaseURL        = client.WithBaseURL
	WithUserAgent      = client.WithUserAgent
	WithDefaultTimeout = client.WithDefaultTimeout
	WithDefaultHeader  = client.WithDefaultHeader
)

// Per-request options.
var (
	WithHeader  = client.WithRequestHeader
	WithQuery   = client.WithRequestQuery
	WithTimeout = client.WithRequestTimeout
)

// Client creates a new HTTP client.
func Client(opts ...client.Option) client.Factory {
	return client.New(opts...)
}
