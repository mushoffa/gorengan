package client

import (
	"encoding/json"
	"net/http"
	"testing"

	fiberv3 "github.com/gofiber/fiber/v3/client"
)

func TestNewFiberV3_ReturnsAFiberV3Client(t *testing.T) {
	// Given no options

	// When
	got := NewFiberV3()

	// Then
	c, ok := got.(*fiberV3)
	if !ok {
		t.Fatalf("NewFiberV3() = %T, want *fiberV3", got)
	}
	if c.client == nil {
		t.Error("client = nil, want a Fiber v3 client")
	}
}

func TestFiberV3_Send_DispatchesEverySupportedMethod(t *testing.T) {
	// Given
	ts := newTestServer(t)
	c := NewFiberV3().(*fiberV3)
	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			// When
			res, err := c.send(method, ts.URL+"/dispatch", fiberv3.Config{})

			// Then
			if err != nil {
				t.Fatalf("send() error = %v", err)
			}
			defer res.Close()

			var echo recordedRequest
			if err := json.Unmarshal(res.Body(), &echo); err != nil {
				t.Fatalf("decode %q: %v", res.Body(), err)
			}
			if echo.Method != method {
				t.Errorf("server saw method %q, want %q", echo.Method, method)
			}
			if echo.Path != "/dispatch" {
				t.Errorf("server saw path %q, want %q", echo.Path, "/dispatch")
			}
		})
	}
}

func TestFiberV3_Send_RejectsAnUnsupportedMethod(t *testing.T) {
	// Given
	c := NewFiberV3().(*fiberV3)

	// When
	res, err := c.send("TRACE", "http://127.0.0.1:1/x", fiberv3.Config{})

	// Then
	if err == nil {
		t.Fatal("send() error = nil, want an error for TRACE")
	}
	if want := `client: unsupported method "TRACE"`; err.Error() != want {
		t.Errorf("send() error = %q, want %q", err.Error(), want)
	}
	if res != nil {
		t.Errorf("send() response = %v, want nil", res)
	}
}
