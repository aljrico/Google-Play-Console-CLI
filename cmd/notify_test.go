package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNotifySendDryRunOutputsPayloadWithoutAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")
	t.Setenv("PLAYPUB_NOTIFY_WEBHOOK_URL", "https://hooks.example.com/services/T000/B000/SECRET?token=secret#fragment")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"notify",
		"send",
		"--title",
		"Release",
		"--message",
		"Internal release staged",
		"--field",
		"track=internal",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"dryRun":true`,
		`"delivered":false`,
		`"message":"Internal release staged"`,
		`"name":"track"`,
		`"value":"internal"`,
		`redacted=true`,
		`#redacted`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	for _, leaked := range []string{"T000", "B000", "SECRET", "token=secret", "fragment"} {
		if strings.Contains(output, leaked) {
			t.Fatalf("output = %s, leaked %s", output, leaked)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestNotifySendPostsWebhook(t *testing.T) {
	var gotPayload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"notify",
		"send",
		"--webhook-url",
		server.URL + "/hook?token=secret",
		"--message",
		"Release shipped",
		"--confirm",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if gotPayload["message"] != "Release shipped" {
		t.Fatalf("payload = %#v", gotPayload)
	}
	output := buf.String()
	for _, want := range []string{`"delivered":true`, `"statusCode":202`, `redacted=true`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "secret") || strings.Contains(output, "/hook") {
		t.Fatalf("output = %s, leaked webhook secret", output)
	}
}

func TestNotifySlackDryRunOutputsSlackPayloadWithoutAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")
	t.Setenv("PLAYPUB_NOTIFY_WEBHOOK_URL", "https://hooks.slack.com/services/T000/B000/SECRET?token=secret#fragment")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"notify",
		"slack",
		"--title",
		"Release",
		"--message",
		"Internal release staged",
		"--field",
		"track=internal",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"dryRun":true`,
		`"delivered":false`,
		`"text":"*Release*\nInternal release staged\ntrack: internal"`,
		`redacted=true`,
		`#redacted`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	for _, leaked := range []string{"T000", "B000", "SECRET", "token=secret", "fragment"} {
		if strings.Contains(output, leaked) {
			t.Fatalf("output = %s, leaked %s", output, leaked)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestNotifySlackPostsWebhook(t *testing.T) {
	var gotPayload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"notify",
		"slack",
		"--webhook-url",
		server.URL + "/hook?token=secret",
		"--message",
		"Release shipped",
		"--confirm",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if gotPayload["text"] != "Release shipped" {
		t.Fatalf("payload = %#v", gotPayload)
	}
	output := buf.String()
	for _, want := range []string{`"delivered":true`, `"statusCode":202`, `redacted=true`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "secret") || strings.Contains(output, "/hook") {
		t.Fatalf("output = %s, leaked webhook secret", output)
	}
}

func TestNotifyMattermostDryRunOutputsMattermostPayloadWithoutAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")
	t.Setenv("PLAYPUB_NOTIFY_WEBHOOK_URL", "https://mattermost.example.com/hooks/SECRET?token=secret#fragment")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"notify",
		"mattermost",
		"--title",
		"Release",
		"--message",
		"Internal release staged",
		"--field",
		"track=internal",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"dryRun":true`,
		`"delivered":false`,
		`"text":"**Release**\nInternal release staged\ntrack: internal"`,
		`redacted=true`,
		`#redacted`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	for _, leaked := range []string{"SECRET", "token=secret", "fragment"} {
		if strings.Contains(output, leaked) {
			t.Fatalf("output = %s, leaked %s", output, leaked)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestNotifyMattermostPostsWebhook(t *testing.T) {
	var gotPayload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"notify",
		"mattermost",
		"--webhook-url",
		server.URL + "/hooks/SECRET?token=secret",
		"--message",
		"Release shipped",
		"--confirm",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if gotPayload["text"] != "Release shipped" {
		t.Fatalf("payload = %#v", gotPayload)
	}
	output := buf.String()
	for _, want := range []string{`"delivered":true`, `"statusCode":200`, `redacted=true`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "secret") || strings.Contains(output, "/hooks/SECRET") {
		t.Fatalf("output = %s, leaked webhook secret", output)
	}
}

func TestNotifyTeamsDryRunOutputsTeamsPayloadWithoutAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")
	t.Setenv("PLAYPUB_NOTIFY_WEBHOOK_URL", "https://example.webhook.office.com/webhookb2/SECRET?token=secret#fragment")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"notify",
		"teams",
		"--title",
		"Release",
		"--message",
		"Internal release staged",
		"--field",
		"track=internal",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"dryRun":true`,
		`"delivered":false`,
		`"text":"Release\nInternal release staged\ntrack: internal"`,
		`redacted=true`,
		`#redacted`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	for _, leaked := range []string{"SECRET", "token=secret", "fragment"} {
		if strings.Contains(output, leaked) {
			t.Fatalf("output = %s, leaked %s", output, leaked)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestNotifyTeamsPostsWebhook(t *testing.T) {
	var gotPayload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"notify",
		"teams",
		"--webhook-url",
		server.URL + "/hook?token=secret",
		"--message",
		"Release shipped",
		"--confirm",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if gotPayload["text"] != "Release shipped" {
		t.Fatalf("payload = %#v", gotPayload)
	}
	output := buf.String()
	for _, want := range []string{`"delivered":true`, `"statusCode":202`, `redacted=true`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "secret") || strings.Contains(output, "/hook") {
		t.Fatalf("output = %s, leaked webhook secret", output)
	}
}

func TestNotifyGoogleChatDryRunOutputsGoogleChatPayloadWithoutAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")
	t.Setenv("PLAYPUB_NOTIFY_WEBHOOK_URL", "https://chat.googleapis.com/v1/spaces/SPACE/messages?key=key-secret&token=token-secret#fragment")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"notify",
		"google-chat",
		"--title",
		"Release",
		"--message",
		"Internal release staged",
		"--field",
		"track=internal",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"dryRun":true`,
		`"delivered":false`,
		`"text":"Release\nInternal release staged\ntrack: internal"`,
		`redacted=true`,
		`#redacted`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	for _, leaked := range []string{"SPACE", "key-secret", "token-secret", "fragment"} {
		if strings.Contains(output, leaked) {
			t.Fatalf("output = %s, leaked %s", output, leaked)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestNotifyGoogleChatPostsWebhook(t *testing.T) {
	var gotPayload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"notify",
		"google-chat",
		"--webhook-url",
		server.URL + "/v1/spaces/SPACE/messages?key=key-secret&token=token-secret",
		"--message",
		"Release shipped",
		"--confirm",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if gotPayload["text"] != "Release shipped" {
		t.Fatalf("payload = %#v", gotPayload)
	}
	output := buf.String()
	for _, want := range []string{`"delivered":true`, `"statusCode":200`, `redacted=true`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "key-secret") || strings.Contains(output, "token-secret") || strings.Contains(output, "SPACE") {
		t.Fatalf("output = %s, leaked webhook secret", output)
	}
}

func TestNotifyDiscordDryRunOutputsDiscordPayloadWithoutAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")
	t.Setenv("PLAYPUB_NOTIFY_WEBHOOK_URL", "https://discord.com/api/webhooks/123/SECRET?token=secret#fragment")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"notify",
		"discord",
		"--title",
		"Release",
		"--message",
		"Internal release staged",
		"--field",
		"track=internal",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"dryRun":true`,
		`"delivered":false`,
		`"content":"**Release**\nInternal release staged\ntrack: internal"`,
		`"allowed_mentions":{"parse":[]}`,
		`redacted=true`,
		`#redacted`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	for _, leaked := range []string{"123", "SECRET", "token=secret", "fragment"} {
		if strings.Contains(output, leaked) {
			t.Fatalf("output = %s, leaked %s", output, leaked)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestNotifyDiscordPostsWebhook(t *testing.T) {
	var gotPayload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
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

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"notify",
		"discord",
		"--webhook-url",
		server.URL + "/hook?token=secret",
		"--message",
		"Release shipped",
		"--confirm",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if gotPayload["content"] != "Release shipped" {
		t.Fatalf("payload = %#v", gotPayload)
	}
	output := buf.String()
	for _, want := range []string{`"delivered":true`, `"statusCode":200`, `redacted=true`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "secret") || strings.Contains(output, "/hook") {
		t.Fatalf("output = %s, leaked webhook secret", output)
	}
}

func TestNotifyGitHubDryRunOutputsRepositoryDispatchPayloadWithoutAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")
	t.Setenv("PLAYPUB_NOTIFY_WEBHOOK_URL", "https://api.github.com/repos/example/project/dispatches?token=secret#fragment")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"notify",
		"github",
		"--event-type",
		"playpub.release",
		"--title",
		"Release",
		"--message",
		"Internal release staged",
		"--field",
		"track=internal",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"dryRun":true`,
		`"delivered":false`,
		`"event_type":"playpub.release"`,
		`"client_payload":{"title":"Release","message":"Internal release staged"`,
		`"name":"track"`,
		`"value":"internal"`,
		`redacted=true`,
		`#redacted`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	for _, leaked := range []string{"token=secret", "fragment"} {
		if strings.Contains(output, leaked) {
			t.Fatalf("output = %s, leaked %s", output, leaked)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestNotifyGitHubPostsWebhook(t *testing.T) {
	var gotPayload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"notify",
		"github",
		"--webhook-url",
		server.URL + "/repos/example/project/dispatches?token=secret",
		"--event-type",
		"playpub.release",
		"--message",
		"Release shipped",
		"--confirm",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if gotPayload["event_type"] != "playpub.release" {
		t.Fatalf("payload = %#v, want event type", gotPayload)
	}
	clientPayload, ok := gotPayload["client_payload"].(map[string]any)
	if !ok || clientPayload["message"] != "Release shipped" {
		t.Fatalf("payload = %#v, want client payload message", gotPayload)
	}
	output := buf.String()
	for _, want := range []string{`"delivered":true`, `"statusCode":204`, `redacted=true`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "secret") || strings.Contains(output, "/repos/example/project") {
		t.Fatalf("output = %s, leaked webhook secret", output)
	}
}

func TestNotifySendTransportErrorDoesNotLeakWebhookSecret(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"notify",
		"send",
		"--webhook-url",
		"http://127.0.0.1:1/services/T000/B000/SECRET?token=secret#fragment",
		"--message",
		"Release shipped",
		"--confirm",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want transport error")
	}
	combined := buf.String() + err.Error()
	for _, leaked := range []string{"T000", "B000", "SECRET", "token=secret", "fragment"} {
		if strings.Contains(combined, leaked) {
			t.Fatalf("combined output = %s, leaked %s", combined, leaked)
		}
	}
}
