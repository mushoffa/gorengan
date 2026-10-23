package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// clients lists every WebClient implementation. The behavioural tests below
// run against each of them, so all clients must act the same way.
var clients = map[string]func(...Option) WebClient{
	"netHttp": NewNetHttp,
	"fiberV2": NewFiberV2,
	"fiberV3": NewFiberV3,
}

// recordedRequest is what the test server saw, echoed back as JSON.
type recordedRequest struct {
	Method string              `json:"method"`
	Path   string              `json:"path"`
	Query  map[string][]string `json:"query"`
	Header map[string]string   `json:"header"`
	Body   string              `json:"body"`
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)

	switch r.URL.Path {
	case "/status/404":
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
		return
	case "/slow":
		select {
		case <-time.After(500 * time.Millisecond):
		case <-r.Context().Done():
			return
		}
	case "/multi":
		w.Header().Add("X-Multi", "a")
		w.Header().Add("X-Multi", "b")
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Reply", "yes")
	_ = json.NewEncoder(w).Encode(recordedRequest{
		Method: r.Method,
		Path:   r.URL.Path,
		Query:  r.URL.Query(),
		Header: map[string]string{
			"X-A":          r.Header.Get("X-A"),
			"User-Agent":   r.Header.Get("User-Agent"),
			"Content-Type": r.Header.Get("Content-Type"),
		},
		Body: string(body),
	})
}

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(requestHandler))
	t.Cleanup(ts.Close)
	return ts
}

func decodeEcho(t *testing.T, res *Response) recordedRequest {
	t.Helper()
	var echo recordedRequest
	if err := res.JSON(&echo); err != nil {
		t.Fatalf("decode echo %q: %v", res.String(), err)
	}
	return echo
}

func headerValues(res *Response, key string) []string {
	for k, v := range res.Header {
		if strings.EqualFold(k, key) {
			return v
		}
	}
	return nil
}

func forEachClient(t *testing.T, fn func(t *testing.T, newClient func(...Option) WebClient)) {
	t.Helper()
	for name, newClient := range clients {
		t.Run(name, func(t *testing.T) {
			fn(t, newClient)
		})
	}
}

func TestClients_Get_ReturnsTheResponse(t *testing.T) {
	forEachClient(t, func(t *testing.T, newClient func(...Option) WebClient) {
		// Given
		ts := newTestServer(t)
		cli := newClient(WithBaseURL(ts.URL))
		defer cli.Close()

		// When
		res, err := cli.Get(context.Background(), "/users")

		// Then
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if !res.OK() {
			t.Errorf("StatusCode = %d, want 2xx", res.StatusCode)
		}
		if res.ContentType != "application/json" {
			t.Errorf("ContentType = %q, want %q", res.ContentType, "application/json")
		}
		if got := res.GetHeader("x-reply"); got != "yes" {
			t.Errorf("X-Reply = %q, want %q", got, "yes")
		}
		echo := decodeEcho(t, res)
		if echo.Method != http.MethodGet || echo.Path != "/users" {
			t.Errorf("server saw %s %s, want GET /users", echo.Method, echo.Path)
		}
	})
}

func TestClients_BaseURL(t *testing.T) {
	t.Run("a relative path is joined to the base URL", func(t *testing.T) {
		forEachClient(t, func(t *testing.T, newClient func(...Option) WebClient) {
			// Given
			ts := newTestServer(t)
			cli := newClient(WithBaseURL(ts.URL))

			// When
			res, err := cli.Get(context.Background(), "/relative")

			// Then
			if err != nil {
				t.Fatalf("Get() error = %v", err)
			}
			if got := decodeEcho(t, res).Path; got != "/relative" {
				t.Errorf("server saw path %q, want %q", got, "/relative")
			}
		})
	})

	t.Run("an absolute URL ignores the base URL", func(t *testing.T) {
		forEachClient(t, func(t *testing.T, newClient func(...Option) WebClient) {
			// Given a base URL that nothing listens on
			ts := newTestServer(t)
			cli := newClient(WithBaseURL("http://127.0.0.1:1"))

			// When
			res, err := cli.Get(context.Background(), ts.URL+"/absolute")

			// Then
			if err != nil {
				t.Fatalf("Get() error = %v", err)
			}
			if got := decodeEcho(t, res).Path; got != "/absolute" {
				t.Errorf("server saw path %q, want %q", got, "/absolute")
			}
		})
	})
}

func TestClients_RequestQuery(t *testing.T) {
	forEachClient(t, func(t *testing.T, newClient func(...Option) WebClient) {
		// Given
		ts := newTestServer(t)
		cli := newClient(WithBaseURL(ts.URL))

		// When
		res, err := cli.Get(context.Background(), "/search",
			WithRequestQuery("q", "go"),
			WithRequestQuery("page", "2"),
		)

		// Then
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		echo := decodeEcho(t, res)
		if got := echo.Query["q"]; len(got) != 1 || got[0] != "go" {
			t.Errorf("query q = %v, want [go]", got)
		}
		if got := echo.Query["page"]; len(got) != 1 || got[0] != "2" {
			t.Errorf("query page = %v, want [2]", got)
		}
	})
}

func TestClients_Headers(t *testing.T) {
	t.Run("a default header is sent", func(t *testing.T) {
		forEachClient(t, func(t *testing.T, newClient func(...Option) WebClient) {
			// Given
			ts := newTestServer(t)
			cli := newClient(WithBaseURL(ts.URL), WithDefaultHeader("X-A", "default"))

			// When
			res, err := cli.Get(context.Background(), "/headers")

			// Then
			if err != nil {
				t.Fatalf("Get() error = %v", err)
			}
			if got := decodeEcho(t, res).Header["X-A"]; got != "default" {
				t.Errorf("X-A = %q, want %q", got, "default")
			}
		})
	})

	t.Run("a request header overrides the default header", func(t *testing.T) {
		forEachClient(t, func(t *testing.T, newClient func(...Option) WebClient) {
			// Given
			ts := newTestServer(t)
			cli := newClient(WithBaseURL(ts.URL), WithDefaultHeader("X-A", "default"))

			// When
			res, err := cli.Get(context.Background(), "/headers", WithRequestHeader("X-A", "request"))

			// Then
			if err != nil {
				t.Fatalf("Get() error = %v", err)
			}
			if got := decodeEcho(t, res).Header["X-A"]; got != "request" {
				t.Errorf("X-A = %q, want %q", got, "request")
			}
		})
	})

	t.Run("the user agent is sent", func(t *testing.T) {
		forEachClient(t, func(t *testing.T, newClient func(...Option) WebClient) {
			// Given
			ts := newTestServer(t)
			cli := newClient(WithBaseURL(ts.URL), WithUserAgent("gorengan-test/1.0"))

			// When
			res, err := cli.Get(context.Background(), "/headers")

			// Then
			if err != nil {
				t.Fatalf("Get() error = %v", err)
			}
			if got := decodeEcho(t, res).Header["User-Agent"]; got != "gorengan-test/1.0" {
				t.Errorf("User-Agent = %q, want %q", got, "gorengan-test/1.0")
			}
		})
	})
}

func TestClients_Post_Body(t *testing.T) {
	t.Run("a map is sent as JSON", func(t *testing.T) {
		forEachClient(t, func(t *testing.T, newClient func(...Option) WebClient) {
			// Given
			ts := newTestServer(t)
			cli := newClient(WithBaseURL(ts.URL))

			// When
			res, err := cli.Post(context.Background(), "/json", map[string]string{"k": "v"})

			// Then
			if err != nil {
				t.Fatalf("Post() error = %v", err)
			}
			echo := decodeEcho(t, res)
			if echo.Method != http.MethodPost {
				t.Errorf("method = %q, want POST", echo.Method)
			}
			if echo.Header["Content-Type"] != "application/json" {
				t.Errorf("Content-Type = %q, want %q", echo.Header["Content-Type"], "application/json")
			}
			var sent map[string]string
			if err := json.Unmarshal([]byte(echo.Body), &sent); err != nil || sent["k"] != "v" {
				t.Errorf("body = %q, want JSON with k=v (err %v)", echo.Body, err)
			}
		})
	})

	t.Run("a struct is sent as JSON", func(t *testing.T) {
		forEachClient(t, func(t *testing.T, newClient func(...Option) WebClient) {
			// Given
			type payload struct {
				Name string
			}
			ts := newTestServer(t)
			cli := newClient(WithBaseURL(ts.URL))

			// When
			res, err := cli.Post(context.Background(), "/json", payload{Name: "gorengan"})

			// Then
			if err != nil {
				t.Fatalf("Post() error = %v", err)
			}
			var sent payload
			if err := json.Unmarshal([]byte(decodeEcho(t, res).Body), &sent); err != nil || sent.Name != "gorengan" {
				t.Errorf("sent = %+v, want Name=gorengan (err %v)", sent, err)
			}
		})
	})

	t.Run("a string is sent as it is", func(t *testing.T) {
		forEachClient(t, func(t *testing.T, newClient func(...Option) WebClient) {
			// Given
			ts := newTestServer(t)
			cli := newClient(WithBaseURL(ts.URL))

			// When
			res, err := cli.Post(context.Background(), "/raw", "plain text")

			// Then
			if err != nil {
				t.Fatalf("Post() error = %v", err)
			}
			if got := decodeEcho(t, res).Body; got != "plain text" {
				t.Errorf("body = %q, want %q", got, "plain text")
			}
		})
	})

	t.Run("bytes are sent as they are", func(t *testing.T) {
		forEachClient(t, func(t *testing.T, newClient func(...Option) WebClient) {
			// Given
			ts := newTestServer(t)
			cli := newClient(WithBaseURL(ts.URL))

			// When
			res, err := cli.Post(context.Background(), "/raw", []byte("raw bytes"))

			// Then
			if err != nil {
				t.Fatalf("Post() error = %v", err)
			}
			if got := decodeEcho(t, res).Body; got != "raw bytes" {
				t.Errorf("body = %q, want %q", got, "raw bytes")
			}
		})
	})

	t.Run("a nil body sends no body", func(t *testing.T) {
		forEachClient(t, func(t *testing.T, newClient func(...Option) WebClient) {
			// Given
			ts := newTestServer(t)
			cli := newClient(WithBaseURL(ts.URL))

			// When
			res, err := cli.Post(context.Background(), "/empty", nil)

			// Then
			if err != nil {
				t.Fatalf("Post() error = %v", err)
			}
			if got := decodeEcho(t, res).Body; got != "" {
				t.Errorf("body = %q, want empty", got)
			}
		})
	})
}

func TestClients_Methods(t *testing.T) {
	forEachClient(t, func(t *testing.T, newClient func(...Option) WebClient) {
		// Given
		ts := newTestServer(t)
		cli := newClient(WithBaseURL(ts.URL))
		ctx := context.Background()

		calls := map[string]func() (*Response, error){
			http.MethodPost:   func() (*Response, error) { return cli.Post(ctx, "/m", "b") },
			http.MethodPut:    func() (*Response, error) { return cli.Put(ctx, "/m", "b") },
			http.MethodPatch:  func() (*Response, error) { return cli.Patch(ctx, "/m", "b") },
			http.MethodDelete: func() (*Response, error) { return cli.Delete(ctx, "/m") },
		}

		for want, call := range calls {
			t.Run(want, func(t *testing.T) {
				// When
				res, err := call()

				// Then
				if err != nil {
					t.Fatalf("%s error = %v", want, err)
				}
				if got := decodeEcho(t, res).Method; got != want {
					t.Errorf("server saw method %q, want %q", got, want)
				}
			})
		}
	})
}

func TestClients_NonSuccessStatus_IsNotAnError(t *testing.T) {
	forEachClient(t, func(t *testing.T, newClient func(...Option) WebClient) {
		// Given
		ts := newTestServer(t)
		cli := newClient(WithBaseURL(ts.URL))

		// When
		res, err := cli.Get(context.Background(), "/status/404")

		// Then
		if err != nil {
			t.Fatalf("Get() error = %v, want nil for a 404", err)
		}
		if res.StatusCode != http.StatusNotFound {
			t.Errorf("StatusCode = %d, want 404", res.StatusCode)
		}
		if res.OK() {
			t.Error("OK() = true, want false")
		}
		if res.String() != "not found" {
			t.Errorf("body = %q, want %q", res.String(), "not found")
		}
	})
}

func TestClients_Timeouts(t *testing.T) {
	const slowBound = 400 * time.Millisecond

	t.Run("a request timeout aborts a slow request", func(t *testing.T) {
		forEachClient(t, func(t *testing.T, newClient func(...Option) WebClient) {
			// Given
			ts := newTestServer(t)
			cli := newClient(WithBaseURL(ts.URL))
			start := time.Now()

			// When
			_, err := cli.Get(context.Background(), "/slow", WithRequestTimeout(50*time.Millisecond))

			// Then
			if err == nil {
				t.Fatal("Get() error = nil, want a timeout error")
			}
			if elapsed := time.Since(start); elapsed > slowBound {
				t.Errorf("returned after %v, want under %v", elapsed, slowBound)
			}
		})
	})

	t.Run("the default timeout aborts a slow request", func(t *testing.T) {
		forEachClient(t, func(t *testing.T, newClient func(...Option) WebClient) {
			// Given
			ts := newTestServer(t)
			cli := newClient(WithBaseURL(ts.URL), WithDefaultTimeout(50*time.Millisecond))
			start := time.Now()

			// When
			_, err := cli.Get(context.Background(), "/slow")

			// Then
			if err == nil {
				t.Fatal("Get() error = nil, want a timeout error")
			}
			if elapsed := time.Since(start); elapsed > slowBound {
				t.Errorf("returned after %v, want under %v", elapsed, slowBound)
			}
		})
	})

	t.Run("a context deadline aborts a slow request", func(t *testing.T) {
		forEachClient(t, func(t *testing.T, newClient func(...Option) WebClient) {
			// Given
			ts := newTestServer(t)
			cli := newClient(WithBaseURL(ts.URL))
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()
			start := time.Now()

			// When
			_, err := cli.Get(ctx, "/slow")

			// Then
			if err == nil {
				t.Fatal("Get() error = nil, want a deadline error")
			}
			if elapsed := time.Since(start); elapsed > slowBound {
				t.Errorf("returned after %v, want under %v", elapsed, slowBound)
			}
		})
	})
}

func TestClients_CanceledContext_ReturnsBeforeSending(t *testing.T) {
	forEachClient(t, func(t *testing.T, newClient func(...Option) WebClient) {
		// Given
		ts := newTestServer(t)
		cli := newClient(WithBaseURL(ts.URL))
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// When
		res, err := cli.Get(ctx, "/users")

		// Then
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Get() error = %v, want context.Canceled", err)
		}
		if res != nil {
			t.Errorf("Get() response = %v, want nil", res)
		}
	})
}

func TestClients_NilContext_IsTreatedAsBackground(t *testing.T) {
	forEachClient(t, func(t *testing.T, newClient func(...Option) WebClient) {
		// Given
		ts := newTestServer(t)
		cli := newClient(WithBaseURL(ts.URL))
		var ctx context.Context

		// When
		res, err := cli.Get(ctx, "/users")

		// Then
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if !res.OK() {
			t.Errorf("StatusCode = %d, want 2xx", res.StatusCode)
		}
	})
}

func TestClients_UnreachableServer_ReturnsAnError(t *testing.T) {
	forEachClient(t, func(t *testing.T, newClient func(...Option) WebClient) {
		// Given a server that has been shut down
		ts := httptest.NewServer(http.HandlerFunc(requestHandler))
		url := ts.URL
		ts.Close()
		cli := newClient(WithBaseURL(url), WithDefaultTimeout(time.Second))

		// When
		res, err := cli.Get(context.Background(), "/users")

		// Then
		if err == nil {
			t.Fatal("Get() error = nil, want a connection error")
		}
		if res != nil {
			t.Errorf("Get() response = %v, want nil", res)
		}
	})
}

func TestClients_Close_CanBeCalledRepeatedly(t *testing.T) {
	forEachClient(t, func(t *testing.T, newClient func(...Option) WebClient) {
		// Given
		cli := newClient()

		// When
		cli.Close()
		cli.Close()

		// Then no panic occurs, and the client is still usable
		ts := newTestServer(t)
		res, err := newClient(WithBaseURL(ts.URL)).Get(context.Background(), "/users")
		if err != nil || !res.OK() {
			t.Errorf("Get() = %v, %v, want a successful response", res, err)
		}
	})
}

func TestClients_UserAgentOption_TakesPrecedenceOverADefaultHeader(t *testing.T) {
	forEachClient(t, func(t *testing.T, newClient func(...Option) WebClient) {
		// Given
		ts := newTestServer(t)
		cli := newClient(
			WithBaseURL(ts.URL),
			WithDefaultHeader("User-Agent", "from-header"),
			WithUserAgent("from-option"),
		)

		// When
		res, err := cli.Get(context.Background(), "/ua")

		// Then
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if got := decodeEcho(t, res).Header["User-Agent"]; got != "from-option" {
			t.Errorf("User-Agent = %q, want %q", got, "from-option")
		}
	})
}

func TestClients_RequestHeaders_DoNotLeakIntoLaterRequests(t *testing.T) {
	forEachClient(t, func(t *testing.T, newClient func(...Option) WebClient) {
		// Given a client with a default header
		ts := newTestServer(t)
		cli := newClient(WithBaseURL(ts.URL), WithDefaultHeader("X-A", "default"))
		ctx := context.Background()

		// When one request overrides the header and the next one does not
		first, err := cli.Get(ctx, "/first", WithRequestHeader("X-A", "request"))
		if err != nil {
			t.Fatalf("first Get() error = %v", err)
		}
		second, err := cli.Get(ctx, "/second")
		if err != nil {
			t.Fatalf("second Get() error = %v", err)
		}

		// Then the override applied only to the first request
		if got := decodeEcho(t, first).Header["X-A"]; got != "request" {
			t.Errorf("first X-A = %q, want %q", got, "request")
		}
		if got := decodeEcho(t, second).Header["X-A"]; got != "default" {
			t.Errorf("second X-A = %q, want %q", got, "default")
		}
	})
}

func TestClients_RepeatedResponseHeaders_AreAllKept(t *testing.T) {
	forEachClient(t, func(t *testing.T, newClient func(...Option) WebClient) {
		// Given a server that sends X-Multi twice
		ts := newTestServer(t)
		cli := newClient(WithBaseURL(ts.URL))

		// When
		res, err := cli.Get(context.Background(), "/multi")

		// Then
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		got := headerValues(res, "X-Multi")
		if len(got) != 2 || got[0] != "a" || got[1] != "b" {
			t.Errorf("X-Multi = %v, want [a b]", got)
		}
	})
}

func TestClients_Response_StaysValidAfterLaterRequests(t *testing.T) {
	forEachClient(t, func(t *testing.T, newClient func(...Option) WebClient) {
		// Given a response that was already returned
		ts := newTestServer(t)
		cli := newClient(WithBaseURL(ts.URL))
		ctx := context.Background()
		first, err := cli.Get(ctx, "/first")
		if err != nil {
			t.Fatalf("first Get() error = %v", err)
		}

		// When more requests reuse the backend's pooled objects
		for i := 0; i < 5; i++ {
			if _, err := cli.Get(ctx, fmt.Sprintf("/later/%d", i)); err != nil {
				t.Fatalf("later Get() error = %v", err)
			}
		}

		// Then the first response still holds its own data
		if got := decodeEcho(t, first).Path; got != "/first" {
			t.Errorf("first response path = %q, want %q", got, "/first")
		}
		if first.ContentType != "application/json" {
			t.Errorf("first response ContentType = %q, want %q", first.ContentType, "application/json")
		}
	})
}

func TestClients_ConcurrentRequests_DoNotInterfere(t *testing.T) {
	forEachClient(t, func(t *testing.T, newClient func(...Option) WebClient) {
		// Given
		ts := newTestServer(t)
		cli := newClient(WithBaseURL(ts.URL))
		const workers = 20
		failures := make(chan string, workers)
		var wg sync.WaitGroup

		// When many goroutines use the same client at once
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				path := fmt.Sprintf("/worker/%d", i)
				res, err := cli.Get(context.Background(), path)
				if err != nil {
					failures <- fmt.Sprintf("%s: %v", path, err)
					return
				}
				var echo recordedRequest
				if err := res.JSON(&echo); err != nil || echo.Path != path {
					failures <- fmt.Sprintf("%s: got path %q (err %v)", path, echo.Path, err)
				}
			}(i)
		}
		wg.Wait()
		close(failures)

		// Then every response matches its own request
		for failure := range failures {
			t.Error(failure)
		}
	})
}

func TestClients_ZeroDefaultTimeout_DisablesTheClientTimeout(t *testing.T) {
	forEachClient(t, func(t *testing.T, newClient func(...Option) WebClient) {
		// Given a client without a timeout and a server that answers after 500ms
		ts := newTestServer(t)
		cli := newClient(WithBaseURL(ts.URL), WithDefaultTimeout(0))

		// When
		res, err := cli.Get(context.Background(), "/slow")

		// Then the request waits for the answer instead of timing out
		if err != nil {
			t.Fatalf("Get() error = %v, want the slow response", err)
		}
		if !res.OK() {
			t.Errorf("StatusCode = %d, want 2xx", res.StatusCode)
		}
	})
}

func TestClients_InvalidURL_ReturnsAnError(t *testing.T) {
	forEachClient(t, func(t *testing.T, newClient func(...Option) WebClient) {
		// Given
		cli := newClient(WithDefaultTimeout(time.Second))

		// When
		res, err := cli.Get(context.Background(), "http://[::1")

		// Then
		if err == nil {
			t.Fatal("Get() error = nil, want an error for an invalid URL")
		}
		if res != nil {
			t.Errorf("Get() response = %v, want nil", res)
		}
	})
}
