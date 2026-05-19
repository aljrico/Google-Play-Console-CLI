package notify

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendDryRunDoesNotCallSender(t *testing.T) {
	sender := failingSender{}
	result, err := Send(context.Background(), sender, SendOptions{
		WebhookURL: "https://hooks.slack.com/services/T000/B000/SECRET?token=secret#fragment-secret",
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
	for _, leaked := range []string{"T000", "B000", "SECRET", "secret", "fragment-secret"} {
		if strings.Contains(result.Webhook, leaked) {
			t.Fatalf("Webhook = %q, leaked %q", result.Webhook, leaked)
		}
	}
	if result.Webhook != "https://hooks.slack.com/redacted?redacted=true#redacted" {
		t.Fatalf("Webhook = %q, want opaque endpoint", result.Webhook)
	}
	if result.Payload.Fields[0].Name != "track" || result.Payload.Fields[0].Value != "internal" {
		t.Fatalf("fields = %#v", result.Payload.Fields)
	}
}

func TestSendReadsWebhookURLFromEnvironment(t *testing.T) {
	t.Setenv("GPC_NOTIFY_WEBHOOK_URL", "https://example.com/hook")
	result, err := Send(context.Background(), failingSender{}, SendOptions{
		Message: "Release staged",
		DryRun:  true,
	})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if result.Webhook != "https://example.com/redacted" {
		t.Fatalf("Webhook = %q, want environment URL", result.Webhook)
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

func TestSendRejectsInsecureNonLoopbackWebhook(t *testing.T) {
	_, err := Send(context.Background(), nil, SendOptions{
		WebhookURL: "http://example.com/hook",
		Message:    "Release shipped",
		DryRun:     true,
	})
	if err == nil {
		t.Fatal("Send() error = nil, want insecure URL validation")
	}
	if !strings.Contains(err.Error(), "https") {
		t.Fatalf("error = %v, want https validation", err)
	}
}

func TestSendRejectsWebhookURLUserInfo(t *testing.T) {
	_, err := Send(context.Background(), nil, SendOptions{
		WebhookURL: "https://user:pass@example.com/hook",
		Message:    "Release shipped",
		DryRun:     true,
	})
	if err == nil {
		t.Fatal("Send() error = nil, want userinfo validation")
	}
	if strings.Contains(err.Error(), "user:pass") || strings.Contains(err.Error(), "pass") {
		t.Fatalf("error = %v, leaked userinfo", err)
	}
}

func TestSendDoesNotFollowRedirects(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("redirect target should not be called")
	}))
	defer target.Close()
	redirector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL+"/leak", http.StatusTemporaryRedirect)
	}))
	defer redirector.Close()

	result, err := Send(context.Background(), WebhookSender{Client: redirector.Client()}, SendOptions{
		WebhookURL: redirector.URL + "/hook",
		Message:    "Release shipped",
		Confirm:    true,
	})
	if err == nil {
		t.Fatal("Send() error = nil, want redirect failure")
	}
	if result.Delivered {
		t.Fatalf("Delivered = true, want false")
	}
	if result.StatusCode != http.StatusTemporaryRedirect {
		t.Fatalf("StatusCode = %d, want 307", result.StatusCode)
	}
}

func TestRedactedErrorRemovesRawWebhookURL(t *testing.T) {
	rawURL := "https://user:pass@hooks.example.com/services/T000/B000/SECRET?token=secret#fragment"
	message := RedactedError(rawURL, errors.New(`Post "`+rawURL+`": dial failed`))
	for _, leaked := range []string{"user", "pass", "T000", "B000", "SECRET", "token=secret", "fragment"} {
		if strings.Contains(message, leaked) {
			t.Fatalf("message = %q, leaked %q", message, leaked)
		}
	}
}

func TestWebhookSenderTransportErrorWithUserInfoDoesNotLeak(t *testing.T) {
	rawURL := "https://user:pass@127.0.0.1:1/services/T000/B000/SECRET?token=secret#fragment"
	_, err := WebhookSender{}.Send(context.Background(), rawURL, Payload{Message: "Release shipped"})
	if err == nil {
		t.Fatal("Send() error = nil, want transport error")
	}
	for _, leaked := range []string{"user", "pass", "T000", "B000", "SECRET", "token=secret", "fragment"} {
		if strings.Contains(err.Error(), leaked) {
			t.Fatalf("error = %v, leaked %s", err, leaked)
		}
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
