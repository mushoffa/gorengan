package client

import (
	"context"
	"testing"
	"time"
)

func TestResolveTimeout(t *testing.T) {
	tests := []struct {
		name           string
		defaultTimeout time.Duration
		perRequest     time.Duration
		hasDeadline    bool
		deadlineIn     time.Duration
		wantExact      time.Duration
		wantAtMost     time.Duration // when set, the context deadline decides: 0 < got <= wantAtMost
	}{
		{
			name:           "uses the client default",
			defaultTimeout: 30 * time.Second,
			wantExact:      30 * time.Second,
		},
		{
			name:           "per-request timeout overrides the default",
			defaultTimeout: 30 * time.Second,
			perRequest:     5 * time.Second,
			wantExact:      5 * time.Second,
		},
		{
			name:           "per-request timeout wins even when it is longer",
			defaultTimeout: 30 * time.Second,
			perRequest:     60 * time.Second,
			wantExact:      60 * time.Second,
		},
		{
			name:           "a sooner context deadline shortens the timeout",
			defaultTimeout: 30 * time.Second,
			hasDeadline:    true,
			deadlineIn:     2 * time.Second,
			wantAtMost:     2 * time.Second,
		},
		{
			name:           "a sooner context deadline also shortens a per-request timeout",
			defaultTimeout: 30 * time.Second,
			perRequest:     20 * time.Second,
			hasDeadline:    true,
			deadlineIn:     2 * time.Second,
			wantAtMost:     2 * time.Second,
		},
		{
			name:           "a later context deadline does not lengthen the timeout",
			defaultTimeout: 30 * time.Second,
			hasDeadline:    true,
			deadlineIn:     time.Hour,
			wantExact:      30 * time.Second,
		},
		{
			name:        "no timeouts and no deadline means no timeout",
			wantExact:   0,
			hasDeadline: false,
		},
		{
			name:        "the deadline applies when no timeout is set",
			hasDeadline: true,
			deadlineIn:  2 * time.Second,
			wantAtMost:  2 * time.Second,
		},
		{
			name:           "an expired deadline is ignored",
			defaultTimeout: 30 * time.Second,
			hasDeadline:    true,
			deadlineIn:     -time.Second,
			wantExact:      30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			ctx := context.Background()
			if tt.hasDeadline {
				var cancel context.CancelFunc
				ctx, cancel = context.WithDeadline(ctx, time.Now().Add(tt.deadlineIn))
				defer cancel()
			}

			// When
			got := resolveTimeout(ctx, tt.defaultTimeout, tt.perRequest)

			// Then
			if tt.wantAtMost > 0 {
				if got <= 0 || got > tt.wantAtMost {
					t.Errorf("resolveTimeout() = %v, want 0 < timeout <= %v", got, tt.wantAtMost)
				}
				return
			}
			if got != tt.wantExact {
				t.Errorf("resolveTimeout() = %v, want %v", got, tt.wantExact)
			}
		})
	}
}

func TestNewFiberV2_ReturnsAFiberV2ClientWithItsConfig(t *testing.T) {
	// Given
	opts := []Option{WithBaseURL("http://example.com"), WithUserAgent("v2-test")}

	// When
	got := NewFiberV2(opts...)

	// Then
	c, ok := got.(*fiberV2)
	if !ok {
		t.Fatalf("NewFiberV2() = %T, want *fiberV2", got)
	}
	if c.client == nil {
		t.Error("client = nil, want a Fiber v2 client")
	}
	if c.cfg.baseURL != "http://example.com" {
		t.Errorf("cfg.baseURL = %q, want %q", c.cfg.baseURL, "http://example.com")
	}
	if c.cfg.userAgent != "v2-test" {
		t.Errorf("cfg.userAgent = %q, want %q", c.cfg.userAgent, "v2-test")
	}
}

func TestFiberV2_RequestHeaderOverridesTheJSONContentType(t *testing.T) {
	// Given
	ts := newTestServer(t)
	cli := NewFiberV2(WithBaseURL(ts.URL))

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
