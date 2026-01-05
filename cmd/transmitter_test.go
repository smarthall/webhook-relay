package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/smarthall/webhook-relay/internal/messaging"
)

func TestProcessDelivery_HeadersAndHost(t *testing.T) {
	tests := []struct {
		name         string
		extraHeaders bool
		preserveHost bool
	}{
		{name: "no-extra-no-preserve", extraHeaders: false, preserveHost: false},
		{name: "extra-only", extraHeaders: true, preserveHost: false},
		{name: "preserve-only", extraHeaders: false, preserveHost: true},
		{name: "extra-and-preserve", extraHeaders: true, preserveHost: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Prepare a RequestMessage and marshal it into an AMQP delivery body
			reqmsg := messaging.RequestMessage{
				Method:  "POST",
				Host:    "original.example.com",
				Path:    "/the/path",
				Headers: map[string][]string{"X-Test": {"1"}},
				Body:    "hello",
			}
			b, err := json.Marshal(reqmsg)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			// Server to capture the forwarded request
			var gotHost string
			var gotRelayPath string
			var gotRelayOriginalHost string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotHost = r.Host
				gotRelayPath = r.Header.Get("Relay-Original-Path")
				gotRelayOriginalHost = r.Header.Get("Relay-Original-Host")
				w.WriteHeader(200)
			}))
			defer srv.Close()

			// Create a dummy delivery and call processDelivery
			del := amqp.Delivery{Body: b}

			client := srv.Client()
			if err := processDelivery(del, client, srv.URL, tc.extraHeaders, tc.preserveHost); err != nil {
				t.Fatalf("processDelivery returned error: %v", err)
			}

			// Verify extra headers
			if tc.extraHeaders {
				if gotRelayPath != reqmsg.Path {
					t.Fatalf("expected Relay-Original-Path=%q, got=%q", reqmsg.Path, gotRelayPath)
				}
				if gotRelayOriginalHost != reqmsg.Host {
					t.Fatalf("expected Relay-Original-Host=%q, got=%q", reqmsg.Host, gotRelayOriginalHost)
				}
			} else {
				if gotRelayPath != "" {
					t.Fatalf("expected no Relay-Original-Path, got=%q", gotRelayPath)
				}
				if gotRelayOriginalHost != "" {
					t.Fatalf("expected no Relay-Original-Host, got=%q", gotRelayOriginalHost)
				}
			}

			// Verify host preservation
			if tc.preserveHost {
				if gotHost != reqmsg.Host {
					t.Fatalf("expected Host=%q, got=%q", reqmsg.Host, gotHost)
				}
			} else {
				// When not preserving host, the Host will be the server host (host:port).
				if gotHost == reqmsg.Host {
					t.Fatalf("did not expect Host to be preserved; got=%q", gotHost)
				}
			}
		})
	}
}
