package notify

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func TestSendSlackDryRunBuildsSlackTextPayload(t *testing.T) {
	sender := failingSlackSender{}
	result, err := SendSlack(context.Background(), sender, SendOptions{
		WebhookURL: "https://hooks.slack.com/services/T000/B000/SECRET?token=secret#fragment-secret",
		Title:      "Release",
		Message:    "Internal release staged",
		Severity:   "info",
		Fields:     []string{"track=internal", "version=42"},
		DryRun:     true,
	})
	if err != nil {
		t.Fatalf("SendSlack() error = %v", err)
	}
	if result.Delivered {
		t.Fatalf("Delivered = true, want false")
	}
	for _, want := range []string{"*Release*", "Internal release staged", "Severity: info", "track: internal", "version: 42"} {
		if !strings.Contains(result.Payload.Text, want) {
			t.Fatalf("Slack text = %q, want %q", result.Payload.Text, want)
		}
	}
	for _, leaked := range []string{"T000", "B000", "SECRET", "secret", "fragment-secret"} {
		if strings.Contains(result.Webhook, leaked) {
			t.Fatalf("Webhook = %q, leaked %q", result.Webhook, leaked)
		}
	}
}

func TestSendSlackPostsSlackWebhook(t *testing.T) {
	var gotPayload SlackPayload
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

	result, err := SendSlack(context.Background(), WebhookSender{Client: server.Client()}, SendOptions{
		WebhookURL: server.URL + "/hook?token=secret",
		Message:    "Release shipped",
		Confirm:    true,
	})
	if err != nil {
		t.Fatalf("SendSlack() error = %v", err)
	}
	if !result.Delivered || result.StatusCode != http.StatusAccepted {
		t.Fatalf("result = %#v, want delivered 202", result)
	}
	if gotPayload.Text != "Release shipped" {
		t.Fatalf("payload = %#v", gotPayload)
	}
	if strings.Contains(result.Webhook, "secret") {
		t.Fatalf("Webhook = %q, leaked query secret", result.Webhook)
	}
}

func TestSendTeamsDryRunBuildsTeamsTextPayload(t *testing.T) {
	sender := failingTeamsSender{}
	result, err := SendTeams(context.Background(), sender, SendOptions{
		CommandPath: "notify teams",
		WebhookURL:  "https://example.webhook.office.com/webhookb2/SECRET?token=secret#fragment-secret",
		Title:       "Release",
		Message:     "Internal release staged",
		Severity:    "info",
		Fields:      []string{"track=internal", "version=42"},
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("SendTeams() error = %v", err)
	}
	if result.Delivered {
		t.Fatalf("Delivered = true, want false")
	}
	for _, want := range []string{"Release", "Internal release staged", "Severity: info", "track: internal", "version: 42"} {
		if !strings.Contains(result.Payload.Text, want) {
			t.Fatalf("Teams text = %q, want %q", result.Payload.Text, want)
		}
	}
	for _, leaked := range []string{"SECRET", "secret", "fragment-secret"} {
		if strings.Contains(result.Webhook, leaked) {
			t.Fatalf("Webhook = %q, leaked %q", result.Webhook, leaked)
		}
	}
	if strings.Contains(result.Webhook, "example.webhook.office.com") {
		t.Fatalf("Webhook = %q, leaked Teams tenant host label", result.Webhook)
	}
}

func TestSendTeamsPostsTeamsWebhook(t *testing.T) {
	var gotPayload TeamsPayload
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

	result, err := SendTeams(context.Background(), WebhookSender{Client: server.Client()}, SendOptions{
		CommandPath: "notify teams",
		WebhookURL:  server.URL + "/hook?token=secret",
		Message:     "Release shipped",
		Confirm:     true,
	})
	if err != nil {
		t.Fatalf("SendTeams() error = %v", err)
	}
	if !result.Delivered || result.StatusCode != http.StatusAccepted {
		t.Fatalf("result = %#v, want delivered 202", result)
	}
	if gotPayload.Text != "Release shipped" {
		t.Fatalf("payload = %#v", gotPayload)
	}
	if strings.Contains(result.Webhook, "secret") {
		t.Fatalf("Webhook = %q, leaked query secret", result.Webhook)
	}
}

func TestSendTeamsRejectsPayloadAboveTeamsLimit(t *testing.T) {
	_, err := SendTeams(context.Background(), failingTeamsSender{}, SendOptions{
		CommandPath: "notify teams",
		WebhookURL:  "https://example.webhook.office.com/webhookb2/SECRET",
		Message:     strings.Repeat("a", maxTeamsWebhookPayloadBytes),
		DryRun:      true,
	})
	if err == nil {
		t.Fatal("SendTeams() error = nil, want payload length validation")
	}
	if !strings.Contains(err.Error(), "28 KB") {
		t.Fatalf("error = %v, want Teams payload length validation", err)
	}
}

func TestSendTeamsDetectsHTTPErrorInSuccessBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Microsoft Teams endpoint returned HTTP error 429"))
	}))
	defer server.Close()

	result, err := SendTeams(context.Background(), WebhookSender{Client: server.Client()}, SendOptions{
		CommandPath: "notify teams",
		WebhookURL:  server.URL + "/hook?token=secret",
		Message:     "Release shipped",
		Confirm:     true,
	})
	if err == nil {
		t.Fatal("SendTeams() error = nil, want Teams body error")
	}
	if result.Delivered {
		t.Fatalf("Delivered = true, want false")
	}
	if result.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("StatusCode = %d, want 429", result.StatusCode)
	}
	if !strings.Contains(err.Error(), "429") {
		t.Fatalf("error = %v, want body status", err)
	}
}

func TestSendDiscordDryRunBuildsDiscordContentPayload(t *testing.T) {
	sender := failingDiscordSender{}
	result, err := SendDiscord(context.Background(), sender, SendOptions{
		CommandPath: "notify discord",
		WebhookURL:  "https://discord.com/api/webhooks/123/SECRET?token=secret#fragment-secret",
		Title:       "Release",
		Message:     "Internal release staged",
		Severity:    "info",
		Fields:      []string{"track=internal", "version=42"},
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("SendDiscord() error = %v", err)
	}
	if result.Delivered {
		t.Fatalf("Delivered = true, want false")
	}
	for _, want := range []string{"**Release**", "Internal release staged", "Severity: info", "track: internal", "version: 42"} {
		if !strings.Contains(result.Payload.Content, want) {
			t.Fatalf("Discord content = %q, want %q", result.Payload.Content, want)
		}
	}
	if result.Payload.AllowedMentions.Parse == nil || len(result.Payload.AllowedMentions.Parse) != 0 {
		t.Fatalf("AllowedMentions.Parse = %#v, want empty mention parse list", result.Payload.AllowedMentions.Parse)
	}
	for _, leaked := range []string{"123", "SECRET", "secret", "fragment-secret"} {
		if strings.Contains(result.Webhook, leaked) {
			t.Fatalf("Webhook = %q, leaked %q", result.Webhook, leaked)
		}
	}
}

func TestSendDiscordPostsDiscordWebhook(t *testing.T) {
	var gotPayload DiscordPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json", got)
		}
		if got := r.URL.Query().Get("wait"); got != "true" {
			t.Fatalf("wait = %q, want true", got)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	result, err := SendDiscord(context.Background(), WebhookSender{Client: server.Client()}, SendOptions{
		CommandPath: "notify discord",
		WebhookURL:  server.URL + "/hook?token=secret",
		Message:     "Release shipped",
		Confirm:     true,
	})
	if err != nil {
		t.Fatalf("SendDiscord() error = %v", err)
	}
	if !result.Delivered || result.StatusCode != http.StatusOK {
		t.Fatalf("result = %#v, want delivered 200", result)
	}
	if gotPayload.Content != "Release shipped" {
		t.Fatalf("payload = %#v", gotPayload)
	}
	if gotPayload.AllowedMentions.Parse == nil || len(gotPayload.AllowedMentions.Parse) != 0 {
		t.Fatalf("AllowedMentions.Parse = %#v, want empty mention parse list", gotPayload.AllowedMentions.Parse)
	}
	if strings.Contains(result.Webhook, "secret") {
		t.Fatalf("Webhook = %q, leaked query secret", result.Webhook)
	}
}

func TestSendDiscordRejectsContentAboveDiscordLimit(t *testing.T) {
	_, err := SendDiscord(context.Background(), failingDiscordSender{}, SendOptions{
		CommandPath: "notify discord",
		WebhookURL:  "https://discord.com/api/webhooks/123/SECRET",
		Message:     strings.Repeat("a", 2001),
		DryRun:      true,
	})
	if err == nil {
		t.Fatal("SendDiscord() error = nil, want content length validation")
	}
	if !strings.Contains(err.Error(), "2000 characters") {
		t.Fatalf("error = %v, want content length validation", err)
	}
}

func TestSendGitHubDryRunBuildsRepositoryDispatchPayload(t *testing.T) {
	sender := failingGitHubSender{}
	result, err := SendGitHub(context.Background(), sender, SendOptions{
		CommandPath: "notify github",
		WebhookURL:  "https://api.github.com/repos/example/project/dispatches?token=secret#fragment-secret",
		EventType:   "gpc.release",
		Title:       "Release",
		Message:     "Internal release staged",
		Severity:    "info",
		Fields:      []string{"track=internal", "version=42"},
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("SendGitHub() error = %v", err)
	}
	if result.Delivered {
		t.Fatalf("Delivered = true, want false")
	}
	if result.Payload.EventType != "gpc.release" {
		t.Fatalf("EventType = %q, want gpc.release", result.Payload.EventType)
	}
	if result.Payload.ClientPayload.Message != "Internal release staged" || result.Payload.ClientPayload.Fields[0].Name != "track" {
		t.Fatalf("Payload = %#v, want client payload", result.Payload)
	}
	for _, leaked := range []string{"token=secret", "fragment-secret"} {
		if strings.Contains(result.Webhook, leaked) {
			t.Fatalf("Webhook = %q, leaked %q", result.Webhook, leaked)
		}
	}
}

func TestSendGitHubPostsRepositoryDispatchWebhook(t *testing.T) {
	var gotPayload GitHubPayload
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
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	result, err := SendGitHub(context.Background(), WebhookSender{Client: server.Client()}, SendOptions{
		CommandPath: "notify github",
		WebhookURL:  server.URL + "/repos/example/project/dispatches?token=secret",
		EventType:   "gpc.release",
		Message:     "Release shipped",
		Confirm:     true,
	})
	if err != nil {
		t.Fatalf("SendGitHub() error = %v", err)
	}
	if !result.Delivered || result.StatusCode != http.StatusNoContent {
		t.Fatalf("result = %#v, want delivered 204", result)
	}
	if gotPayload.EventType != "gpc.release" || gotPayload.ClientPayload.Message != "Release shipped" {
		t.Fatalf("payload = %#v, want repository dispatch payload", gotPayload)
	}
	if strings.Contains(result.Webhook, "secret") {
		t.Fatalf("Webhook = %q, leaked query secret", result.Webhook)
	}
}

func TestSendGitHubRejectsInvalidEventType(t *testing.T) {
	_, err := SendGitHub(context.Background(), failingGitHubSender{}, SendOptions{
		CommandPath: "notify github",
		WebhookURL:  "https://api.github.com/repos/example/project/dispatches",
		EventType:   " release",
		Message:     "Release shipped",
		DryRun:      true,
	})
	if err == nil {
		t.Fatal("SendGitHub() error = nil, want event type validation")
	}
	if !strings.Contains(err.Error(), "event type") {
		t.Fatalf("error = %v, want event type validation", err)
	}
}

func TestDiscordWebhookURLWithWaitPreservesAndOverridesQuery(t *testing.T) {
	waitURL, err := DiscordWebhookURLWithWait("https://discord.com/api/webhooks/123/SECRET?thread_id=abc&wait=false")
	if err != nil {
		t.Fatalf("DiscordWebhookURLWithWait() error = %v", err)
	}
	parsedURL, err := url.Parse(waitURL)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got := parsedURL.Query().Get("thread_id"); got != "abc" {
		t.Fatalf("thread_id = %q, want abc", got)
	}
	if got := parsedURL.Query().Get("wait"); got != "true" {
		t.Fatalf("wait = %q, want true", got)
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

func TestSendSlackRequiresConfirmOrDryRunWithSlackCommandName(t *testing.T) {
	_, err := SendSlack(context.Background(), nil, SendOptions{
		CommandPath: "notify slack",
		WebhookURL:  "https://example.com/hook",
		Message:     "Release shipped",
	})
	if err == nil {
		t.Fatal("SendSlack() error = nil, want confirmation validation")
	}
	if !strings.Contains(err.Error(), "notify slack requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want slack command validation", err)
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

type failingSlackSender struct{}

func (failingSlackSender) SendSlack(context.Context, string, SlackPayload) (int, error) {
	panic("sender should not be called")
}

type failingTeamsSender struct{}

func (failingTeamsSender) SendTeams(context.Context, string, TeamsPayload) (int, error) {
	panic("sender should not be called")
}

type failingDiscordSender struct{}

func (failingDiscordSender) SendDiscord(context.Context, string, DiscordPayload) (int, error) {
	panic("sender should not be called")
}

type failingGitHubSender struct{}

func (failingGitHubSender) SendGitHub(context.Context, string, GitHubPayload) (int, error) {
	panic("sender should not be called")
}
