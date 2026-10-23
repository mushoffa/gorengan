package client

import (
	"context"
	"testing"
)

func TestNew_UsesTheNetHttpBackend(t *testing.T) {
	// Given no options

	// When
	got := New()

	// Then
	f, ok := got.(*factory)
	if !ok {
		t.Fatalf("New() = %T, want *factory", got)
	}
	if _, ok := f.WebClient.(*netHttp); !ok {
		t.Errorf("embedded client = %T, want *netHttp", f.WebClient)
	}
}

func TestFactory_ForwardsOptionsToEveryBackend(t *testing.T) {
	// Given
	ts := newTestServer(t)
	opts := []Option{
		WithBaseURL(ts.URL),
		WithUserAgent("forward-test/1.0"),
		WithDefaultHeader("X-A", "forwarded"),
	}
	factories := map[string]func() WebClient{
		"netHttp": func() WebClient { return New(opts...) },
		"fiberV2": func() WebClient { return New(opts...).Fiber().V2() },
		"fiberV3": func() WebClient { return New(opts...).Fiber().V3() },
	}

	for name, build := range factories {
		t.Run(name, func(t *testing.T) {
			// Given a client built from the factory
			cli := build()

			// When
			res, err := cli.Get(context.Background(), "/forwarded")

			// Then the base URL, user agent and default header all reached the server
			if err != nil {
				t.Fatalf("Get() error = %v", err)
			}
			echo := decodeEcho(t, res)
			if echo.Path != "/forwarded" {
				t.Errorf("path = %q, want %q", echo.Path, "/forwarded")
			}
			if echo.Header["User-Agent"] != "forward-test/1.0" {
				t.Errorf("User-Agent = %q, want %q", echo.Header["User-Agent"], "forward-test/1.0")
			}
			if echo.Header["X-A"] != "forwarded" {
				t.Errorf("X-A = %q, want %q", echo.Header["X-A"], "forwarded")
			}
		})
	}
}
