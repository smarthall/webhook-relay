package cmd

import (
	"bytes"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/smarthall/webhook-relay/internal/messaging"
)

// mockPub implements the small publisher interface expected by requestHandler.
type mockPub struct {
	called      bool
	receivedMsg messaging.RequestMessage
	errToReturn error
}

func (m *mockPub) Publish(msg messaging.RequestMessage) error {
	m.called = true
	m.receivedMsg = msg
	return m.errToReturn
}

// errReader returns an error on Read to simulate a bad request body.
type errReader struct{}

func (errReader) Read(p []byte) (int, error) { return 0, errors.New("read error") }

// TestRequestHandlerSuccess verifies that a valid HTTP request is translated
// into a RequestMessage and published, and that the handler responds 204.
func TestRequestHandlerSuccess(t *testing.T) {
	body := "hello world"
	req := httptest.NewRequest("POST", "http://example.com/path", bytes.NewBufferString(body))
	req.Header.Set("X-Test", "v")
	req.Host = "example.com"

	pub := &mockPub{}
	rr := httptest.NewRecorder()

	handler := requestHandler(pub)
	handler.ServeHTTP(rr, req)

	if rr.Code != 204 {
		t.Fatalf("expected status 204, got %d", rr.Code)
	}
	if !pub.called {
		t.Fatalf("expected publisher to be called")
	}
	if pub.receivedMsg.Method != "POST" {
		t.Fatalf("expected method POST, got %s", pub.receivedMsg.Method)
	}
	if pub.receivedMsg.Path != "/path" {
		t.Fatalf("expected path /path, got %s", pub.receivedMsg.Path)
	}
	if pub.receivedMsg.Body != body {
		t.Fatalf("expected body %q, got %q", body, pub.receivedMsg.Body)
	}
	if vals := pub.receivedMsg.Headers["X-Test"]; len(vals) == 0 || vals[0] != "v" {
		t.Fatalf("expected header X-Test=v, got %v", pub.receivedMsg.Headers["X-Test"])
	}
}

// TestRequestHandlerFromHTTPRequestError verifies that when reading the
// request body fails the handler returns HTTP 500 and does not call Publish.
func TestRequestHandlerFromHTTPRequestError(t *testing.T) {
	req := httptest.NewRequest("POST", "http://example.com/bad", errReader{})
	req.Host = "example.com"

	pub := &mockPub{}
	rr := httptest.NewRecorder()

	handler := requestHandler(pub)
	handler.ServeHTTP(rr, req)

	if rr.Code != 500 {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}
	if pub.called {
		t.Fatalf("expected publisher NOT to be called on body read error")
	}
}

// TestRequestHandlerPublishError verifies that when Publish returns an error
// the handler responds with HTTP 500.
func TestRequestHandlerPublishError(t *testing.T) {
	body := "payload"
	req := httptest.NewRequest("POST", "http://example.com/publishfail", bytes.NewBufferString(body))
	req.Host = "example.com"

	pub := &mockPub{errToReturn: errors.New("publish failed")}
	rr := httptest.NewRecorder()

	handler := requestHandler(pub)
	handler.ServeHTTP(rr, req)

	if rr.Code != 500 {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}
	if !pub.called {
		t.Fatalf("expected publisher to be called")
	}
}
