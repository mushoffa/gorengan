package client

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestNewNetHttp_UsesItsOwnTransport(t *testing.T) {
	// Given
	first := NewNetHttp().(*netHttp)

	// When
	second := NewNetHttp().(*netHttp)

	// Then
	if first.client.Transport == nil {
		t.Fatal("Transport = nil, want a dedicated transport")
	}
	if first.client.Transport == http.DefaultTransport {
		t.Error("Transport is http.DefaultTransport, want a clone")
	}
	if first.client.Transport == second.client.Transport {
		t.Error("two clients share one transport, want separate connection pools")
	}
}

func TestNewTransport_ReturnsAClone(t *testing.T) {
	// Given
	defaultTransport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		t.Skip("http.DefaultTransport is not an *http.Transport")
	}

	// When
	got := newTransport(newConfig(nil))

	// Then
	if got == nil {
		t.Fatal("newTransport() = nil")
	}
	if got == defaultTransport {
		t.Error("newTransport() returned http.DefaultTransport itself, want a clone")
	}
}

func TestNewNetHttp_AppliesTheDefaultTimeoutToTheHTTPClient(t *testing.T) {
	tests := []struct {
		name string
		opts []Option
		want time.Duration
	}{
		{"30 seconds when unset", nil, 30 * time.Second},
		{"a custom timeout", []Option{WithDefaultTimeout(5 * time.Second)}, 5 * time.Second},
		{"zero disables it", []Option{WithDefaultTimeout(0)}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given the options from the table

			// When
			got := NewNetHttp(tt.opts...).(*netHttp)

			// Then
			if got.client.Timeout != tt.want {
				t.Errorf("http.Client.Timeout = %v, want %v", got.client.Timeout, tt.want)
			}
		})
	}
}

func TestNewNetHttp_KeepsItsConfig(t *testing.T) {
	// Given
	opts := []Option{WithBaseURL("http://example.com"), WithUserAgent("nethttp-test")}

	// When
	got := NewNetHttp(opts...).(*netHttp)

	// Then
	if got.cfg.baseURL != "http://example.com" {
		t.Errorf("cfg.baseURL = %q, want %q", got.cfg.baseURL, "http://example.com")
	}
	if got.cfg.userAgent != "nethttp-test" {
		t.Errorf("cfg.userAgent = %q, want %q", got.cfg.userAgent, "nethttp-test")
	}
}

func TestNetHttp_RequestHeaderOverridesTheJSONContentType(t *testing.T) {
	// Given
	ts := newTestServer(t)
	cli := NewNetHttp(WithBaseURL(ts.URL))

	// When
	res, err := cli.Post(context.Background(), "/json", map[string]string{"k": "v"},
		WithRequestHeader("Content-Type", "application/vnd.api+json"))

	// Then
	if err != nil {
		t.Fatalf("Post() error = %v", err)
	}
	if got := decodeEcho(t, res).Header["Content-Type"]; got != "application/vnd.api+json" {
		t.Errorf("Content-Type = %q, want %q", got, "application/vnd.api+json")
	}
}

func TestNetHttp_Do_RejectsAnInvalidMethod(t *testing.T) {
	// Given
	cli := NewNetHttp().(*netHttp)

	// When
	res, err := cli.do(context.Background(), "BAD METHOD", "http://127.0.0.1:1/x", nil, nil)

	// Then
	if err == nil {
		t.Fatal("do() error = nil, want an error for a method with a space")
	}
	if res != nil {
		t.Errorf("do() response = %v, want nil", res)
	}
}
