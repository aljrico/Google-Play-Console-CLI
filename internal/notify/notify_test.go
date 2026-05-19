package notify

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendDryRunDoesNotCallSender(t *testing.T) {
	sender := failingSender{}
	result, err := Send(context.Background(), sender, SendOptions{
		WebhookURL: "https://example.com/hook?token=secret",
		Title:      "Release",
		Message:    "Internal release staged",
		Severity:   "info",
		Fields:     []string{"track=internal"},
		DryRun:     true,
	})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if result.Delivered {
		t.Fatalf("Delivered = true, want false")
	}
	if !strings.Contains(result.Webhook, "redacted=true") {
		t.Fatalf("Webhook = %q, want redacted query", result.Webhook)
	}
	if result.Payload.Fields[0].Name != "track" || result.Payload.Fields[0].Value != "internal" {
		t.Fatalf("fields = %#v", result.Payload.Fields)
	}
}

func TestSendPostsJSONWebhook(t *testing.T) {
	var gotPayload Payload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json", got)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	result, err := Send(context.Background(), WebhookSender{Client: server.Client()}, SendOptions{
		WebhookURL: server.URL + "/hook?token=secret",
		Message:    "Release shipped",
		Confirm:    true,
	})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if !result.Delivered || result.StatusCode != http.StatusAccepted {
		t.Fatalf("result = %#v, want delivered 202", result)
	}
	if gotPayload.Message != "Release shipped" {
		t.Fatalf("payload = %#v", gotPayload)
	}
	if strings.Contains(result.Webhook, "secret") {
		t.Fatalf("Webhook = %q, leaked query secret", result.Webhook)
	}
}

func TestSendRequiresConfirmOrDryRun(t *testing.T) {
	_, err := Send(context.Background(), nil, SendOptions{
		WebhookURL: "https://example.com/hook",
		Message:    "Release shipped",
	})
	if err == nil {
		t.Fatal("Send() error = nil, want confirmation validation")
	}
	if !strings.Contains(err.Error(), "--confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation validation", err)
	}
}

func TestParseFieldRequiresNameValue(t *testing.T) {
	_, err := ParseField("track")
	if err == nil {
		t.Fatal("ParseField() error = nil, want validation")
	}
}

type failingSender struct{}

func (failingSender) Send(context.Context, string, Payload) (int, error) {
	panic("sender should not be called")
}
