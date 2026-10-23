package client

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
)

// WebClient is implemented by every web client factory.
//
// A non-2xx status is not an error: check Response.OK or StatusCode.
// body may be nil, a string or []byte (sent as-is), or any other value
// (sent as JSON with Content-Type application/json).
type WebClient interface {
	Get(ctx context.Context, path string, opts ...RequestOption) (*Response, error)
	Post(ctx context.Context, path string, body any, opts ...RequestOption) (*Response, error)
	Put(ctx context.Context, path string, body any, opts ...RequestOption) (*Response, error)
	Patch(ctx context.Context, path string, body any, opts ...RequestOption) (*Response, error)
	Delete(ctx context.Context, path string, opts ...RequestOption) (*Response, error)
	Close()
}

// Response is a fully-read copy of the server's reply,
// safe to keep after the call returns.
type Response struct {
	Body        []byte
	ContentType string
	Header      map[string][]string
	StatusCode  int
}

// OK reports whether the status is 2xx.
func (r *Response) OK() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// String returns the body as a string.
func (r *Response) String() string {
	return string(r.Body)
}

// JSON decodes the body
func (r *Response) JSON(v any) error {
	return json.Unmarshal(r.Body, v)
}

// GetHeader returns the first value of a header, matching the name
// case-insensitively. It returns "" when the header is absent.
func (r *Response) GetHeader(key string) string {
	for k, v := range r.Header {
		if strings.EqualFold(k, key) && len(v) > 0 {
			return v[0]
		}
	}
	return ""
}

// buildURL joins baseURL and path (unless path is already absolute) and
// applies query parameters.
func buildURL(baseURL, path string, query map[string]string) (string, error) {
	requestURL := path
	if !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") {
		requestURL = baseURL + path
	}
	if len(query) == 0 {
		return requestURL, nil
	}
	u, err := url.Parse(requestURL)
	if err != nil {
		return "", err
	}
	q := u.Query()
	for k, v := range query {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// buildResponse builds a Response from a backend's reply.
// Header is never nil, and ContentType is taken from the Content-Type header
func buildResponse(statusCode int, header map[string][]string, body []byte) *Response {
	res := &Response{
		StatusCode: statusCode,
		Header:     header,
		Body:       body,
	}
	if res.Header == nil {
		res.Header = map[string][]string{}
	}
	res.ContentType = res.GetHeader("Content-Type")
	return res
}

// encodeBody returns the request body and the Content-Type to set ("" when
// the caller controls it). A nil result means "no body".
func encodeBody(body any) ([]byte, string, error) {
	switch b := body.(type) {
	case nil:
		return nil, "", nil
	case []byte:
		return b, "", nil
	case string:
		return []byte(b), "", nil
	default:
		data, err := json.Marshal(b)
		if err != nil {
			return nil, "", err
		}
		return data, "application/json", nil
	}
}
