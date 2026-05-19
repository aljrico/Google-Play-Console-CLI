package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aljrico/Google-Play-Console-CLI/internal/config"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/aljrico/Google-Play-Console-CLI/internal/websurface"
)

func TestVersionJSON(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{"version", "--output", "json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"version"`) {
		t.Fatalf("version output = %s", buf.String())
	}
}

func TestUnknownOutputFormat(t *testing.T) {
	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{"version", "--output", "yaml"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error")
	}
}

func TestAccountStatusReportsMissingProfileWithoutAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{"account", "status", "--output", "json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"configured":false`,
		`"serviceAccountMetadataOk":false`,
		`no active auth profile`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
}

func TestAccountStatusReportsServiceAccountMetadata(t *testing.T) {
	root := t.TempDir()
	configPath := root + "/config.json"
	serviceAccountPath := root + "/service-account.json"
	t.Setenv("GPC_CONFIG", configPath)
	if err := os.WriteFile(serviceAccountPath, []byte(`{
  "type": "service_account",
  "project_id": "play-project",
  "private_key": "-----BEGIN PRIVATE KEY-----\nabc\n-----END PRIVATE KEY-----\n",
  "client_email": "gpc@example.iam.gserviceaccount.com"
}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := config.Save(config.Store{
		ActiveProfile: "default",
		Profiles: map[string]config.Profile{
			"default": {
				Name:               "default",
				ServiceAccountFile: serviceAccountPath,
			},
		},
	}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{"account", "status", "--output", "json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"configured":true`,
		`"activeProfile":"default"`,
		`"serviceAccountFileExists":true`,
		`"serviceAccountReadable":true`,
		`"serviceAccountJsonParsed":true`,
		`"serviceAccountEmail":"gpc@example.iam.gserviceaccount.com"`,
		`"projectId":"play-project"`,
		`"serviceAccountMetadataOk":true`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "BEGIN PRIVATE KEY") {
		t.Fatalf("output leaked private key: %s", output)
	}
}

func TestAccountStatusReportsMalformedServiceAccountFileAsExisting(t *testing.T) {
	root := t.TempDir()
	configPath := root + "/config.json"
	serviceAccountPath := root + "/service-account.json"
	t.Setenv("GPC_CONFIG", configPath)
	if err := os.WriteFile(serviceAccountPath, []byte("{"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := config.Save(config.Store{
		ActiveProfile: "default",
		Profiles: map[string]config.Profile{
			"default": {
				Name:               "default",
				ServiceAccountFile: serviceAccountPath,
			},
		},
	}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{"account", "status", "--output", "json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"serviceAccountFileExists":true`,
		`"serviceAccountReadable":true`,
		`"serviceAccountJsonParsed":false`,
		`parse service account file`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
}

func TestAccountStatusReportsStaleActiveProfileAsUnconfigured(t *testing.T) {
	configPath := t.TempDir() + "/config.json"
	t.Setenv("GPC_CONFIG", configPath)
	if err := config.Save(config.Store{
		ActiveProfile: "missing",
		Profiles: map[string]config.Profile{
			"other": {
				Name:               "other",
				ServiceAccountFile: "/tmp/service-account.json",
			},
		},
	}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{"account", "status", "--output", "json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"configured":false`,
		`active profile \"missing\" is missing`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
}

func TestVersionRejectsUnexpectedArgs(t *testing.T) {
	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{"version", "stray"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error")
	}
}

func TestCapabilitiesOutputsParityMatrixWithoutAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"capabilities",
		"--status",
		"tested",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, `"status":"tested"`) {
		t.Fatalf("output = %s, want tested status", output)
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestCapabilitiesRejectsUnsupportedStatus(t *testing.T) {
	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"capabilities",
		"--status",
		"done-ish",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected unsupported status error")
	}
	if !strings.Contains(err.Error(), "unsupported capability status") {
		t.Fatalf("error = %v, want status validation", err)
	}
}

func TestDocsParityOutputsMarkdownWithoutAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"docs",
		"parity",
		"--output",
		"markdown",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "# Parity Matrix") {
		t.Fatalf("output = %s, want parity matrix markdown", output)
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestDocsParityOutputsJSONDocument(t *testing.T) {
	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"docs",
		"parity",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"name":"parity"`, `"format":"markdown"`, `# Parity Matrix`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
}

func TestDocsCommandsOutputsJSONReferenceWithoutAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"docs",
		"commands",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"name":"gpc"`,
		`"path":"gpc users"`,
		`"path":"gpc users create"`,
		`"path":"gpc purchases product acknowledge"`,
		`"path":"gpc releases"`,
		`"path":"gpc vitals metric-set query"`,
		`"path":"gpc vitals errors issues search"`,
		`"path":"gpc vitals errors reports search"`,
		`"path":"gpc vitals anomalies list"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
	if strings.Contains(output, `"name":"help"`) {
		t.Fatalf("output = %s, did not expect generated help flags", output)
	}
}

func TestDocsCommandsOutputsMarkdownReference(t *testing.T) {
	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"docs",
		"commands",
		"--output",
		"markdown",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		"# Command Reference",
		"`gpc users`",
		"`gpc users create`",
		"`gpc purchases product acknowledge`",
		"`gpc releases`",
		"`gpc vitals metric-set query`",
		"`gpc vitals errors issues search`",
		"`gpc vitals errors reports search`",
		"`gpc vitals anomalies list`",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "`--help`") {
		t.Fatalf("output = %s, did not expect generated help flags", output)
	}
}

func TestInstallSkillsDryRunOutputsJSONWithoutWriting(t *testing.T) {
	directory := t.TempDir() + "/skills"

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"install-skills",
		"--directory",
		directory,
		"--skill",
		"gpc-cli-usage",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"directory":"` + directory + `"`,
		`"dryRun":true`,
		`"name":"gpc-cli-usage"`,
		`"written":false`,
		`"wouldWrite":true`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if _, err := os.Stat(directory + "/gpc-cli-usage/SKILL.md"); !os.IsNotExist(err) {
		t.Fatalf("skill file exists after dry-run or stat error = %v", err)
	}
}

func TestInstallSkillsListOutputsBundledSkillNames(t *testing.T) {
	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"install-skills",
		"list",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"name":"gpc-cli-usage"`,
		`"name":"gpc-metadata-workflow"`,
		`"name":"gpc-release-flow"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
}

func TestMigrateSupplyInspectOutputsInventoryWithoutAuth(t *testing.T) {
	root := t.TempDir()
	directory := root + "/fastlane/metadata/android"
	writeRootTestPathContent(t, directory+"/en-US/title.txt", "Example")
	writeRootTestPathContent(t, directory+"/en-US/changelogs/42.txt", "Bug fixes")
	writeRootTestPathContent(t, directory+"/en-US/images/phoneScreenshots/1.png", "png")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"migrate",
		"supply",
		"inspect",
		"--directory",
		directory,
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"localeCount":1`,
		`"language":"en-US"`,
		`"name":"title.txt"`,
		`"type":"phoneScreenshots"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestMigrateSupplyConvertOutputsMetadataWithoutAuth(t *testing.T) {
	root := t.TempDir()
	directory := root + "/fastlane/metadata/android"
	writeRootTestPathContent(t, directory+"/en-US/title.txt", "Example\n")
	writeRootTestPathContent(t, directory+"/en-US/short_description.txt", "Short\n")
	writeRootTestPathContent(t, directory+"/es-ES/title.txt", "Ejemplo")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"migrate",
		"supply",
		"convert",
		"--directory",
		directory,
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"listings"`,
		`"language":"en-US"`,
		`"title":"Example"`,
		`"shortDescription":"Short"`,
		`"language":"es-ES"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestNotifySendDryRunOutputsPayloadWithoutAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	t.Setenv("GPC_NOTIFY_WEBHOOK_URL", "https://hooks.example.com/services/T000/B000/SECRET?token=secret#fragment")

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
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	t.Setenv("GPC_NOTIFY_WEBHOOK_URL", "https://hooks.slack.com/services/T000/B000/SECRET?token=secret#fragment")

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

func TestNotifyTeamsDryRunOutputsTeamsPayloadWithoutAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	t.Setenv("GPC_NOTIFY_WEBHOOK_URL", "https://example.webhook.office.com/webhookb2/SECRET?token=secret#fragment")

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
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	t.Setenv("GPC_NOTIFY_WEBHOOK_URL", "https://chat.googleapis.com/v1/spaces/SPACE/messages?key=key-secret&token=token-secret#fragment")

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
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	t.Setenv("GPC_NOTIFY_WEBHOOK_URL", "https://discord.com/api/webhooks/123/SECRET?token=secret#fragment")

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
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	t.Setenv("GPC_NOTIFY_WEBHOOK_URL", "https://api.github.com/repos/example/project/dispatches?token=secret#fragment")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"notify",
		"github",
		"--event-type",
		"gpc.release",
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
		`"event_type":"gpc.release"`,
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
		"gpc.release",
		"--message",
		"Release shipped",
		"--confirm",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if gotPayload["event_type"] != "gpc.release" {
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
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

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

func TestSearchFindsCommandsWithoutAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"search",
		"release",
		"upload",
		"--limit",
		"3",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"query":"release upload"`,
		`"limit":3`,
		`"path":"gpc releases upload"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSearchFindsFlagsAsTyped(t *testing.T) {
	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"search",
		"--limit",
		"5",
		"--",
		"--webhook-url-env",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"query":"--webhook-url-env"`,
		`"path":"gpc notify send"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
}

func TestSnitchReportOutputsIssueURLWithoutAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"snitch",
		"report",
		"--title",
		"Confusing release output",
		"--body",
		"The track summary was hard to read.",
		"--command",
		"gpc releases list --package com.example.app",
		"--label",
		"ux",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"repository":"aljrico/Google-Play-Console-CLI"`,
		`"title":"Confusing release output"`,
		`"command":"gpc releases list --package com.example.app"`,
		`"labels":["snitch","ux"]`,
		`"issueUrl":"https://github.com/aljrico/Google-Play-Console-CLI/issues/new?`,
		`&labels=snitch%2Cux&`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestNotificationsRTDNDecodeOutputsKindWithoutAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	notification := `{"version":"1.0","packageName":"com.example.app","eventTimeMillis":1700000000000,"testNotification":{"version":"1.0"}}`
	payload := fmt.Sprintf(`{"message":{"data":%q,"messageId":"136969346945"},"subscription":"projects/example/subscriptions/play-rtdn"}`, base64.StdEncoding.EncodeToString([]byte(notification)))
	path := writeRootTestPathContent(t, filepath.Join(t.TempDir(), "pubsub.json"), payload)

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"notifications",
		"rtdn",
		"decode",
		"--file",
		path,
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"kind":"test"`,
		`"messageId":"136969346945"`,
		`"packageName":"com.example.app"`,
		`"testNotification":{"version":"1.0"}`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestNotificationsPubSubSetupDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"notifications",
		"pubsub",
		"setup",
		"--project",
		"play-project",
		"--topic",
		"play-rtdn",
		"--subscription",
		"play-rtdn-sub",
		"--push-endpoint",
		"https://example.com/rtdn",
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
		`"topicName":"projects/play-project/topics/play-rtdn"`,
		`"subscriptionName":"projects/play-project/subscriptions/play-rtdn-sub"`,
		`"publisherMember":"serviceAccount:google-play-developer-notifications@system.gserviceaccount.com"`,
		`"pushEndpoint":"https://example.com/rtdn"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestNotificationsPubSubPullRejectsAckWithoutConfirmBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"notifications",
		"pubsub",
		"pull",
		"--project",
		"play-project",
		"--subscription",
		"play-rtdn-sub",
		"--ack",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected acknowledgement confirmation validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm") {
		t.Fatalf("error = %v, want confirmation validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestInsightsAnomaliesSummarizeOutputsCountsWithoutAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	file := writeRootTestPathContent(t, filepath.Join(t.TempDir(), "anomalies.json"), `{
		"packageName": "com.example.app",
		"anomalies": [
			{"name":"a1","metricSet":"apps/com.example.app/crashRateMetricSet","metric":{"metric":"crashRate"}},
			{"name":"a2","metricSet":"apps/com.example.app/crashRateMetricSet","metric":{"metric":"crashRate"}}
		]
	}`)

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"insights",
		"anomalies",
		"summarize",
		"--file",
		file,
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"total":2`,
		`"packageName":"com.example.app"`,
		`"name":"apps/com.example.app/crashRateMetricSet","count":2`,
		`"top metric: crashRate (2)"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestInsightsReportsSummarizeOutputsFinanceAndStatsWithoutAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	dir := t.TempDir()
	financeFile := writeRootTestPathContent(t, filepath.Join(dir, "earnings.csv"), "Transaction Type,Merchant Currency,Amount (Merchant Currency)\nCharge,USD,9.99\nCharge,USD,1.01\nGoogle fee,USD,-1.50\n")
	statsFile := writeRootTestPathContent(t, filepath.Join(dir, "stats.csv"), "Date,Package name,Country,Store listing visitors,Store listing acquisitions,Conversion rate\n2026-05-01,com.example.app,US,10,2,20\n2026-05-02,com.example.app,US,20,3,22\n")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"insights",
		"reports",
		"summarize",
		"--finance-file",
		financeFile,
		"--stats-file",
		statsFile,
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"reportType":"earnings"`,
		`"transactionType":"Charge"`,
		`"packageName":"com.example.app"`,
		`"name":"Store listing acquisitions"`,
		`"kind":"finance"`,
		`"kind":"stats"`,
		`top transaction by count is Charge`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestFinanceReportsSummarizeOutputsTotalsWithoutAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	file := writeRootTestPathContent(t, filepath.Join(t.TempDir(), "earnings.csv"), "Transaction Type,Merchant Currency,Amount (Merchant Currency)\nCharge,USD,9.99\nCharge,USD,1.01\nGoogle fee,USD,-1.50\n")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"finance",
		"reports",
		"summarize",
		"--file",
		file,
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"reportType":"earnings"`,
		`"rows":3`,
		`"amountColumn":"Amount (Merchant Currency)"`,
		`"transactionType":"Charge","count":2,"total":"11","currency":"USD"`,
		`"transactionType":"Google fee","count":1,"total":"-1.5","currency":"USD"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestFinanceReportsDownloadDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	outputPath := filepath.Join(t.TempDir(), "earnings.zip")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"finance",
		"reports",
		"download",
		"--bucket",
		"pubsite_prod_rev_0123456789",
		"--object",
		"earnings/earnings_202605.zip",
		"--file",
		outputPath,
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"dryRun":true`, `"downloaded":false`, `"bucket":"pubsite_prod_rev_0123456789"`, `"object":"earnings/earnings_202605.zip"`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestAnalyticsStatsSummarizeOutputsTotalsWithoutAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	file := writeRootTestPathContent(t, filepath.Join(t.TempDir(), "store_performance.csv"), "Date,Package name,Country/region,Store listing visitors,Store listing acquisitions\n2026-05-01,com.example.app,US,10,2\n2026-05-02,com.example.app,US,15,3\n")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"analytics",
		"stats",
		"summarize",
		"--file",
		file,
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"rows":2`,
		`"packageName":"com.example.app"`,
		`"startDate":"2026-05-01"`,
		`"endDate":"2026-05-02"`,
		`"name":"Store listing visitors","aggregation":"sum","value":"25"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestAnalyticsStatsDownloadDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	outputPath := filepath.Join(t.TempDir(), "stats.csv")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"analytics",
		"stats",
		"download",
		"--bucket",
		"pubsite_prod_rev_0123456789",
		"--object",
		"stats/store_performance/store_performance_com.example.app_202605_country.csv",
		"--file",
		outputPath,
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"dryRun":true`, `"downloaded":false`, `"object":"stats/store_performance/store_performance_com.example.app_202605_country.csv"`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func writeRootTestPathContent(t *testing.T, path string, content string) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}

func TestDiffJSONOutputsChangesWithoutAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	from := writeRootTestContent(t, "from.json", `{"title":"Old","screenshots":["one.png"]}`)
	to := writeRootTestContent(t, "to.json", `{"title":"New","screenshots":["one.png","two.png"]}`)

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"diff",
		"json",
		from,
		to,
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"equal":false`, `"path":"/screenshots/1"`, `"kind":"added"`, `"path":"/title"`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestDiffJSONFailOnChangeWritesResultAndReturnsError(t *testing.T) {
	from := writeRootTestContent(t, "from.json", `{"title":"Old"}`)
	to := writeRootTestContent(t, "to.json", `{"title":"New"}`)

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"diff",
		"json",
		from,
		to,
		"--fail-on-change",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want fail-on-change error")
	}
	if !strings.Contains(err.Error(), "JSON files differ: 1 change(s)") {
		t.Fatalf("error = %v, want change count", err)
	}
	if !strings.Contains(buf.String(), `"equal":false`) {
		t.Fatalf("output = %s, want diff result before error", buf.String())
	}
}

func TestDiffJSONFailOnChangeAllowsEqualFiles(t *testing.T) {
	from := writeRootTestContent(t, "from.json", `{"title":"Same"}`)
	to := writeRootTestContent(t, "to.json", `{"title":"Same"}`)

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"diff",
		"json",
		from,
		to,
		"--fail-on-change",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"equal":true`, `"changes":[]`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
}

func TestSchemaOutputsDiscoverySummaryWithoutAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
		  "name": "androidpublisher",
		  "version": "v3",
		  "resources": {
		    "edits": {
		      "resources": {
		        "tracks": {
		          "methods": {
		            "list": {
		              "id": "androidpublisher.edits.tracks.list",
		              "path": "androidpublisher/v3/applications/{packageName}/edits/{editId}/tracks",
		              "httpMethod": "GET"
		            }
		          }
		        }
		      }
		    }
		  }
		}`))
	}))
	defer server.Close()
	previousDiscoveryURL := schemaDiscoveryURL
	schemaDiscoveryURL = server.URL
	defer func() {
		schemaDiscoveryURL = previousDiscoveryURL
	}()

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"schema",
		"--resource",
		"edits.tracks",
		"--method",
		"list",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"name":"androidpublisher"`,
		`"path":"edits.tracks"`,
		`"id":"androidpublisher.edits.tracks.list"`,
		`"httpMethod":"GET"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSchemaOutputsFlatMarkdownSummary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
		  "name": "androidpublisher",
		  "version": "v3",
		  "resources": {
		    "edits": {
		      "resources": {
		        "tracks": {
		          "methods": {
		            "list": {
		              "id": "androidpublisher.edits.tracks.list",
		              "path": "androidpublisher/v3/applications/{packageName}/edits/{editId}/tracks",
		              "httpMethod": "GET"
		            }
		          }
		        }
		      }
		    }
		  }
		}`))
	}))
	defer server.Close()
	previousDiscoveryURL := schemaDiscoveryURL
	schemaDiscoveryURL = server.URL
	defer func() {
		schemaDiscoveryURL = previousDiscoveryURL
	}()

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"schema",
		"--resource",
		"edits",
		"--method",
		"list",
		"--output",
		"markdown",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		"| resource | method | id | httpMethod | path | description |",
		"edits.tracks",
		"androidpublisher.edits.tracks.list",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
}

func TestAppsListReportsBlockedSurfaceWithoutAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"apps",
		"list",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected unsupported apps list error")
	}
	if !strings.Contains(err.Error(), "limited app discovery APIs") {
		t.Fatalf("error = %v, want app discovery limitation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestWebStatusDocumentsBlockedSurfaceWithoutAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"web",
		"status",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	var status websurface.Status
	if err := json.Unmarshal(buf.Bytes(), &status); err != nil {
		t.Fatalf("Unmarshal() error = %v; output = %s", err, buf.String())
	}
	if status.Status != "blocked" || status.Surface != "Play Console browser workflows" || len(status.Alternatives) == 0 {
		t.Fatalf("status = %#v, want blocked web boundary", status)
	}
	output := buf.String()
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestTestersUpdateDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"testers",
		"update",
		"--package",
		"com.example.app",
		"--track",
		"internal",
		"--google-group",
		"qa@example.com",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"track":"internal"`, `"dryRun":true`, `"googleGroups":["qa@example.com"]`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestTestersUpdateRejectsConfirmAndDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"testers",
		"update",
		"--package",
		"com.example.app",
		"--track",
		"internal",
		"--google-group",
		"qa@example.com",
		"--confirm",
		"--dry-run",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want confirm and dry-run conflict")
	}
	if !strings.Contains(err.Error(), "--confirm and --dry-run cannot be used together") {
		t.Fatalf("error = %v, want confirm and dry-run conflict", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestInitDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"init",
		"--directory",
		t.TempDir() + "/.gpc",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, `"dryRun":true`) {
		t.Fatalf("output = %s, want dryRun true", output)
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestWorkflowListDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	workflowFile := writeRootTestContent(t, "workflow.json", `{
	  "version": 1,
	  "workflows": {
	    "release": [{"run": "echo release"}]
	  }
	}`)

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"workflow",
		"--file",
		workflowFile,
		"list",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"name":"release"`, `"steps":1`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestWorkflowRunDryRunDoesNotExecute(t *testing.T) {
	workflowFile := writeRootTestContent(t, "workflow.json", `{
	  "version": 1,
	  "workflows": {
	    "release": [{"run": "exit 1"}]
	  }
	}`)

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"workflow",
		"--file",
		workflowFile,
		"run",
		"release",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"dryRun":true`, `"skipped":true`, `"success":true`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
}

func TestWorkflowRunFailureWritesResultAndReturnsError(t *testing.T) {
	workflowFile := writeRootTestContent(t, "workflow.json", `{
	  "version": 1,
	  "workflows": {
	    "release": [{"run": "printf nope; exit 7"}]
	  }
	}`)

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"workflow",
		"--file",
		workflowFile,
		"run",
		"release",
		"--confirm",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want workflow failure")
	}
	output := buf.String()
	for _, want := range []string{`"success":false`, `"exitCode":7`, `"stdout":"nope"`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
}

func TestWorkflowRunMissingFileReturnsErrorWithoutZeroResult(t *testing.T) {
	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"workflow",
		"--file",
		t.TempDir() + "/missing.json",
		"run",
		"release",
		"--confirm",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want missing workflow file")
	}
	if buf.String() != "" {
		t.Fatalf("output = %s, want empty output", buf.String())
	}
}

func TestWorkflowRunRejectsConfirmAndDryRun(t *testing.T) {
	workflowFile := writeRootTestContent(t, "workflow.json", `{
	  "version": 1,
	  "workflows": {
	    "release": [{"run": "echo release"}]
	  }
	}`)

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"workflow",
		"--file",
		workflowFile,
		"run",
		"release",
		"--confirm",
		"--dry-run",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want confirm and dry-run conflict")
	}
	if !strings.Contains(err.Error(), "--confirm and --dry-run cannot be used together") {
		t.Fatalf("error = %v, want confirm and dry-run conflict", err)
	}
	if buf.String() != "" {
		t.Fatalf("output = %s, want empty output", buf.String())
	}
}

func TestPublishInternalDryRunRejectsInvalidPackage(t *testing.T) {
	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"publish",
		"internal",
		"--package",
		"bad",
		"--aab",
		"app-release.aab",
		"--dry-run",
	})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error")
	}
}

func TestStatusRejectsInvalidPackageBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"status",
		"--package",
		"bad",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected package validation error")
	}
	if !strings.Contains(err.Error(), "invalid Android package name") {
		t.Fatalf("error = %v, want package validation", err)
	}
}

func TestValidateRejectsInvalidPackageBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"validate",
		"--package",
		"bad",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected package validation error")
	}
	if !strings.Contains(err.Error(), "invalid Android package name") {
		t.Fatalf("error = %v, want package validation", err)
	}
}

func TestUsersListRejectsMissingDeveloperBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"users",
		"list",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected developer validation error")
	}
	if !strings.Contains(err.Error(), "developer account") {
		t.Fatalf("error = %v, want developer validation", err)
	}
}

func TestUsersListRejectsInvalidPageSizeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"users",
		"list",
		"--developer",
		"1234567890",
		"--page-size",
		"-2",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected page size validation error")
	}
	if !strings.Contains(err.Error(), "page size") {
		t.Fatalf("error = %v, want page size validation", err)
	}
}

func TestGrantsCreateDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"grants",
		"create",
		"--developer",
		"1234567890",
		"--user-email",
		"user@example.com",
		"--package",
		"com.example.app",
		"--permission",
		"CAN_VIEW_NON_FINANCIAL_DATA",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"dryRun":true`) {
		t.Fatalf("output = %s, want dryRun true", buf.String())
	}
	if !strings.Contains(buf.String(), `"target":"developers/1234567890/users/user@example.com/grants/com.example.app"`) {
		t.Fatalf("output = %s, want full grant target", buf.String())
	}
	if !strings.Contains(buf.String(), `"appLevelPermissions":["CAN_VIEW_NON_FINANCIAL_DATA"]`) {
		t.Fatalf("output = %s, want grant permission preview", buf.String())
	}
	if strings.Contains(buf.String(), "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", buf.String())
	}
}

func TestGrantsPatchRequiresConfirmBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"grants",
		"patch",
		"--name",
		"developers/123/users/user@example.com/grants/com.example.app",
		"--permission",
		"CAN_REPLY_TO_REVIEWS",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirmation validation error")
	}
	if !strings.Contains(err.Error(), "--confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestInternalSharingUploadDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"internal-sharing",
		"upload",
		"--package",
		"com.example.app",
		"--aab",
		"app-release.aab",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, `"kind":"bundle"`) || !strings.Contains(output, `"dryRun":true`) {
		t.Fatalf("output = %s, want bundle dry run", output)
	}
}

func TestInternalSharingUploadRejectsMissingArtifactBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"internal-sharing",
		"upload",
		"--package",
		"com.example.app",
		"--apk",
		t.TempDir() + "/missing.apk",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected missing artifact error")
	}
	if !strings.Contains(err.Error(), "open file") {
		t.Fatalf("error = %v, want local file preflight", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestInternalSharingUploadRejectsDirectoryArtifactBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	artifactPath := t.TempDir() + "/directory.apk"
	if err := os.Mkdir(artifactPath, 0o755); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"internal-sharing",
		"upload",
		"--package",
		"com.example.app",
		"--apk",
		artifactPath,
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected directory artifact error")
	}
	if !strings.Contains(err.Error(), "regular file") {
		t.Fatalf("error = %v, want regular file validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestAppRecoveryListRejectsMissingVersionCodeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"app-recovery",
		"list",
		"--package",
		"com.example.app",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected version code validation error")
	}
	if !strings.Contains(err.Error(), "version code") {
		t.Fatalf("error = %v, want version code validation", err)
	}
}

func TestAppRecoveryDeployDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"app-recovery",
		"deploy",
		"--package",
		"com.example.app",
		"--id",
		"7",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"action":"deploy"`, `"dryRun":true`, `"applied":false`, `"appRecoveryId":"7"`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestAppRecoveryCreateDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"app-recovery",
		"create",
		"--package",
		"com.example.app",
		"--version-code",
		"42",
		"--region",
		"US",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"dryRun":true`, `"created":false`, `"versionCodes":[42]`, `"regionCodes":["US"]`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestAppRecoveryCreateRejectsMissingTargetingBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"app-recovery",
		"create",
		"--package",
		"com.example.app",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected targeting validation error")
	}
	if !strings.Contains(err.Error(), "targeting") {
		t.Fatalf("error = %v, want targeting validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestAppRecoveryAddTargetingDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"app-recovery",
		"add-targeting",
		"--package",
		"com.example.app",
		"--id",
		"7",
		"--sdk-level",
		"26",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"dryRun":true`, `"applied":false`, `"appRecoveryId":"7"`, `"sdkLevels":[26]`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestAppRecoveryAddTargetingRejectsMissingTargetBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"app-recovery",
		"add-targeting",
		"--package",
		"com.example.app",
		"--id",
		"7",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected targeting validation error")
	}
	if !strings.Contains(err.Error(), "targeting") {
		t.Fatalf("error = %v, want targeting validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestAppRecoveryCancelRejectsMissingConfirmBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"app-recovery",
		"cancel",
		"--package",
		"com.example.app",
		"--id",
		"7",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm validation error")
	}
	if !strings.Contains(err.Error(), "--confirm") {
		t.Fatalf("error = %v, want confirm validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestGeneratedAPKsListRejectsMissingVersionCodeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"generated-apks",
		"list",
		"--package",
		"com.example.app",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected version code validation error")
	}
	if !strings.Contains(err.Error(), "version code") {
		t.Fatalf("error = %v, want version code validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestGeneratedAPKsDownloadDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	outputPath := t.TempDir() + "/split.apk"

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"generated-apks",
		"download",
		"--package",
		"com.example.app",
		"--version-code",
		"42",
		"--download-id",
		"split-download",
		"--file",
		outputPath,
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"dryRun":true`, `"downloaded":false`, `"outputPath":"` + outputPath + `"`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestGeneratedAPKsDownloadRejectsMissingFileBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"generated-apks",
		"download",
		"--package",
		"com.example.app",
		"--version-code",
		"42",
		"--download-id",
		"split-download",
		"--dry-run",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected output path validation error")
	}
	if !strings.Contains(err.Error(), "output path") {
		t.Fatalf("error = %v, want output path validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestSystemAPKVariantsListRejectsMissingVersionCodeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"system-apks",
		"variants",
		"list",
		"--package",
		"com.example.app",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected version code validation error")
	}
	if !strings.Contains(err.Error(), "version code") {
		t.Fatalf("error = %v, want version code validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestDeviceTierConfigsListRejectsInvalidPageSizeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"device-tier-configs",
		"list",
		"--package",
		"com.example.app",
		"--page-size",
		"101",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected page size validation error")
	}
	if !strings.Contains(err.Error(), "page size") {
		t.Fatalf("error = %v, want page size validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestDeviceTierConfigsGetRejectsMissingIDBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"device-tier-configs",
		"get",
		"--package",
		"com.example.app",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected device tier config ID validation error")
	}
	if !strings.Contains(err.Error(), "device tier config ID") {
		t.Fatalf("error = %v, want device tier config ID validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestPublishInternalDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"publish",
		"internal",
		"--package",
		"com.example.app",
		"--aab",
		"app-release.aab",
		"--release-note",
		"en-US=Bug fixes.",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"dryRun":true`) {
		t.Fatalf("publish dry-run output = %s", buf.String())
	}
	if !strings.Contains(buf.String(), `"releaseNotes":[{"language":"en-US","text":"Bug fixes."}]`) {
		t.Fatalf("publish dry-run output = %s, want release note", buf.String())
	}
}

func TestPublishInternalLiveRejectsMissingBundleBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"publish",
		"internal",
		"--package",
		"com.example.app",
		"--aab",
		t.TempDir() + "/missing.aab",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected missing bundle error")
	}
	if !strings.Contains(err.Error(), "open bundle") {
		t.Fatalf("error = %v, want bundle preflight", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestMigrateSupplyChangelogsDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	directory := filepath.Join(t.TempDir(), "fastlane", "metadata", "android")
	writeNestedRootTestFile(t, filepath.Join(directory, "en-US", "changelogs", "42.txt"), "Bug fixes.\n")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"migrate",
		"supply",
		"changelogs",
		"--directory",
		directory,
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"versionCode":42`, `"releaseNotes":[{"language":"en-US","text":"Bug fixes."}]`, `"releaseNoteArgs":["en-US=Bug fixes."]`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestMigrateSupplyImagesDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	directory := filepath.Join(t.TempDir(), "fastlane", "metadata", "android")
	imagePath := filepath.Join(directory, "en-US", "images", "phoneScreenshots", "1.png")
	writeNestedRootTestFile(t, imagePath, "png")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"migrate",
		"supply",
		"images",
		"--directory",
		directory,
		"--language",
		"en-US",
		"--type",
		"phoneScreenshots",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"language":"en-US"`, `"type":"phoneScreenshots"`, `"uploadArgs":["--language","en-US","--type","phoneScreenshots","--file","` + filepath.ToSlash(imagePath) + `"]`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestPublishInternalLiveRejectsInvalidUserFractionBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	bundlePath := writeRootTestFile(t, "app-release.aab")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"publish",
		"internal",
		"--package",
		"com.example.app",
		"--aab",
		bundlePath,
		"--user-fraction",
		"0.25",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected user fraction validation error")
	}
	if !strings.Contains(err.Error(), "user fraction can only be set") {
		t.Fatalf("error = %v, want user fraction validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestReleasesUploadRejectsMalformedReleaseNoteBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"upload",
		"--package",
		"com.example.app",
		"--aab",
		"app-release.aab",
		"--release-note",
		"en-US",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected release note validation error")
	}
	if !strings.Contains(err.Error(), "language=text") {
		t.Fatalf("error = %v, want release note format validation", err)
	}
}

func TestReleasesUploadDryRunUsesRequestedTrack(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"upload",
		"--package",
		"com.example.app",
		"--track",
		"beta",
		"--aab",
		"app-release.aab",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	var result struct {
		Track string `json:"track"`
		Plan  struct {
			Track      string `json:"track"`
			Artifact   string `json:"artifact"`
			APKPath    string `json:"apkPath,omitempty"`
			BundlePath string `json:"bundlePath,omitempty"`
			Status     string `json:"status"`
		} `json:"plan"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Unmarshal() error = %v; output = %s", err, buf.String())
	}
	if result.Track != "beta" || result.Plan.Track != "beta" {
		t.Fatalf("release upload dry-run result = %#v, want beta track", result)
	}
	if result.Plan.Artifact != "bundle" || result.Plan.BundlePath != "app-release.aab" || result.Plan.APKPath != "" {
		t.Fatalf("release upload dry-run plan = %#v, want bundle-only artifact", result.Plan)
	}
	if result.Plan.Status != "completed" {
		t.Fatalf("release upload dry-run plan = %#v, want completed status", result.Plan)
	}
}

func TestReleasesUploadDryRunSupportsAPK(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"upload",
		"--package",
		"com.example.app",
		"--track",
		"internal",
		"--apk",
		"app-release.apk",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	var result struct {
		Track string `json:"track"`
		Plan  struct {
			Track      string `json:"track"`
			Artifact   string `json:"artifact"`
			APKPath    string `json:"apkPath,omitempty"`
			BundlePath string `json:"bundlePath,omitempty"`
			Status     string `json:"status"`
		} `json:"plan"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Unmarshal() error = %v; output = %s", err, buf.String())
	}
	if result.Track != "internal" || result.Plan.Track != "internal" {
		t.Fatalf("release upload dry-run result = %#v, want internal track", result)
	}
	if result.Plan.Artifact != "apk" || result.Plan.APKPath != "app-release.apk" || result.Plan.BundlePath != "" {
		t.Fatalf("release upload dry-run plan = %#v, want APK-only artifact", result.Plan)
	}
	if strings.Contains(buf.String(), "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", buf.String())
	}
}

func TestReleasesUploadRejectsMultipleArtifactsBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"upload",
		"--package",
		"com.example.app",
		"--apk",
		"app-release.apk",
		"--aab",
		"app-release.aab",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected multiple artifact validation error")
	}
	if !strings.Contains(err.Error(), "exactly one of APK path or AAB path") {
		t.Fatalf("error = %v, want artifact validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestReleasesUploadLiveRejectsMissingAPKBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"upload",
		"--package",
		"com.example.app",
		"--apk",
		t.TempDir() + "/missing.apk",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected missing APK error")
	}
	if !strings.Contains(err.Error(), "open APK") {
		t.Fatalf("error = %v, want APK preflight", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestReleasesUploadLiveRejectsInvalidUserFractionBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	bundlePath := writeRootTestFile(t, "app-release.aab")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"upload",
		"--package",
		"com.example.app",
		"--aab",
		bundlePath,
		"--user-fraction",
		"0.25",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected user fraction validation error")
	}
	if !strings.Contains(err.Error(), "user fraction can only be set") {
		t.Fatalf("error = %v, want user fraction validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestReleasesPromoteDryRun(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"promote",
		"--package",
		"com.example.app",
		"--from",
		"internal",
		"--to",
		"production",
		"--version-code",
		"42",
		"--release-note",
		"en-US=Production rollout.",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"toTrack":"production"`) {
		t.Fatalf("release promote dry-run output = %s", buf.String())
	}
	if !strings.Contains(buf.String(), `"text":"Production rollout."`) {
		t.Fatalf("release promote dry-run output = %s, want release note", buf.String())
	}
}

func TestReleasesHaltDryRun(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"halt",
		"--package",
		"com.example.app",
		"--track",
		"production",
		"--version-code",
		"42",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"action":"halt"`) {
		t.Fatalf("release halt dry-run output = %s", buf.String())
	}
}

func TestReleasesResumeDryRun(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"resume",
		"--package",
		"com.example.app",
		"--track",
		"production",
		"--version-code",
		"42",
		"--status",
		"inProgress",
		"--user-fraction",
		"0.25",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"action":"resume"`) {
		t.Fatalf("release resume dry-run output = %s", buf.String())
	}
}

func TestReleasesResumeCompletedDryRun(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"resume",
		"--package",
		"com.example.app",
		"--track",
		"production",
		"--version-code",
		"42",
		"--status",
		"completed",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"status":"completed"`) {
		t.Fatalf("release resume completed dry-run output = %s", buf.String())
	}
}

func TestListingsUpdateDryRun(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"listings",
		"update",
		"--package",
		"com.example.app",
		"--language",
		"en-US",
		"--title",
		"Example",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"language":"en-US"`) {
		t.Fatalf("listing update dry-run output = %s", buf.String())
	}
}

func TestListingsDeleteDryRun(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"listings",
		"delete",
		"--package",
		"com.example.app",
		"--language",
		"en-US",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"language":"en-US"`) {
		t.Fatalf("listing delete dry-run output = %s", buf.String())
	}
}

func TestListingsDeleteAllDryRun(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"listings",
		"delete-all",
		"--package",
		"com.example.app",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"all":true`) {
		t.Fatalf("listing delete-all dry-run output = %s", buf.String())
	}
}

func TestImagesListRejectsInvalidTypeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"images",
		"list",
		"--package",
		"com.example.app",
		"--language",
		"en-US",
		"--type",
		"poster",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected image type validation error")
	}
	if !strings.Contains(err.Error(), "unsupported image type") {
		t.Fatalf("error = %v, want image type validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestImagesListRejectsMissingTypeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"images",
		"list",
		"--package",
		"com.example.app",
		"--language",
		"en-US",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected image type validation error")
	}
	if !strings.Contains(err.Error(), "image type is required") {
		t.Fatalf("error = %v, want missing image type validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestImagesUploadDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	imagePath := writeRootTestFile(t, "feature.png")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"images",
		"upload",
		"--package",
		"com.example.app",
		"--language",
		"en-US",
		"--type",
		"featureGraphic",
		"--file",
		imagePath,
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	var result struct {
		Path   string `json:"path"`
		DryRun bool   `json:"dryRun"`
		Plan   struct {
			Path string `json:"path"`
		} `json:"plan"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Unmarshal() error = %v; output = %s", err, buf.String())
	}
	if result.Path != imagePath || result.Plan.Path != imagePath || !result.DryRun {
		t.Fatalf("result = %#v, want image upload dry-run", result)
	}
	if strings.Contains(buf.String(), "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", buf.String())
	}
}

func TestImagesUploadDryRunRejectsMissingFileBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"images",
		"upload",
		"--package",
		"com.example.app",
		"--language",
		"en-US",
		"--type",
		"featureGraphic",
		"--file",
		t.TempDir() + "/missing.png",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected missing image error")
	}
	if !strings.Contains(err.Error(), "open image") {
		t.Fatalf("error = %v, want image preflight", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestImagesUploadRejectsMissingFileBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"images",
		"upload",
		"--package",
		"com.example.app",
		"--language",
		"en-US",
		"--type",
		"featureGraphic",
		"--file",
		t.TempDir() + "/missing.png",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected missing image error")
	}
	if !strings.Contains(err.Error(), "open image") {
		t.Fatalf("error = %v, want image preflight", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestImagesDeleteDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"images",
		"delete",
		"--package",
		"com.example.app",
		"--language",
		"en-US",
		"--type",
		"phoneScreenshots",
		"--image-id",
		"image-1",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	var result struct {
		ImageID string `json:"imageId"`
		DryRun  bool   `json:"dryRun"`
		Plan    struct {
			ImageID string `json:"imageId"`
			All     bool   `json:"all"`
		} `json:"plan"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Unmarshal() error = %v; output = %s", err, buf.String())
	}
	if result.ImageID != "image-1" || result.Plan.ImageID != "image-1" || result.Plan.All || !result.DryRun {
		t.Fatalf("result = %#v, want single image delete dry-run", result)
	}
	if strings.Contains(buf.String(), "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", buf.String())
	}
}

func TestImagesDeleteAllDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"images",
		"delete-all",
		"--package",
		"com.example.app",
		"--language",
		"en-US",
		"--type",
		"featureGraphic",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	var result struct {
		All    bool `json:"all"`
		DryRun bool `json:"dryRun"`
		Plan   struct {
			All bool `json:"all"`
		} `json:"plan"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Unmarshal() error = %v; output = %s", err, buf.String())
	}
	if !result.All || !result.Plan.All || !result.DryRun {
		t.Fatalf("result = %#v, want delete-all dry-run", result)
	}
	if strings.Contains(buf.String(), "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", buf.String())
	}
}

func TestImagesDeleteRejectsMissingImageIDBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"images",
		"delete",
		"--package",
		"com.example.app",
		"--language",
		"en-US",
		"--type",
		"phoneScreenshots",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected missing image ID error")
	}
	if !strings.Contains(err.Error(), "image ID is required") {
		t.Fatalf("error = %v, want image ID validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestImagesDeleteAllRejectsInvalidTypeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"images",
		"delete-all",
		"--package",
		"com.example.app",
		"--language",
		"en-US",
		"--type",
		"bad",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected image type validation error")
	}
	if !strings.Contains(err.Error(), "unsupported image type") {
		t.Fatalf("error = %v, want image type validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestDetailsUpdateDryRun(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"details",
		"update",
		"--package",
		"com.example.app",
		"--contact-email",
		"support@example.com",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"contactEmail":"support@example.com"`) {
		t.Fatalf("details update dry-run output = %s", buf.String())
	}
}

func TestMetadataApplyDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	metadataPath := writeRootTestContent(t, "metadata.json", `{
  "details": {
    "contactEmail": "support@example.com"
  },
  "listings": [
    {
      "language": "en-US",
      "title": "Example"
    }
  ]
}`)

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"metadata",
		"apply",
		"--package",
		"com.example.app",
		"--file",
		metadataPath,
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
		`"contactEmail":"support@example.com"`,
		`"patch app details"`,
		`"patch en-US listing"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestMetadataApplyRejectsUnknownFileFieldsBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	metadataPath := writeRootTestContent(t, "metadata.json", `{"details":{"contactEmail":"support@example.com"},"surprise":true}`)

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"metadata",
		"apply",
		"--package",
		"com.example.app",
		"--file",
		metadataPath,
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected unknown field validation error")
	}
	if !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("error = %v, want unknown field validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestMetadataApplyRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	metadataPath := writeRootTestContent(t, "metadata.json", `{"details":{"contactEmail":"support@example.com"}}`)

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"metadata",
		"apply",
		"--package",
		"com.example.app",
		"--file",
		metadataPath,
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirmation gate error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestMetadataApplyRejectsConfirmAndDryRunBeforeFileRead(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"metadata",
		"apply",
		"--package",
		"com.example.app",
		"--file",
		t.TempDir() + "/missing.json",
		"--confirm",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected flag conflict error")
	}
	if !strings.Contains(err.Error(), "--confirm and --dry-run cannot be used together") {
		t.Fatalf("error = %v, want flag conflict", err)
	}
	if strings.Contains(err.Error(), "open file") {
		t.Fatalf("error = %v, did not expect file read error", err)
	}
}

func TestDataSafetyUpdateDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	csvPath := writeRootTestFile(t, "data-safety.csv")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"data-safety",
		"update",
		"--package",
		"com.example.app",
		"--csv",
		csvPath,
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"dryRun":true`) {
		t.Fatalf("output = %s, want dryRun true", buf.String())
	}
	if strings.Contains(buf.String(), "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", buf.String())
	}
}

func TestDataSafetyUpdateRequiresConfirmBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"data-safety",
		"update",
		"--package",
		"com.example.app",
		"--csv",
		t.TempDir() + "/missing.csv",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirmation validation error")
	}
	if !strings.Contains(err.Error(), "--confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestReviewsReplyDryRun(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"reviews",
		"reply",
		"--package",
		"com.example.app",
		"--review-id",
		"review-123",
		"--text",
		"Thanks for trying the app.",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"reviewId":"review-123"`) {
		t.Fatalf("reviews reply dry-run output = %s", buf.String())
	}
	if !strings.Contains(buf.String(), `"dryRun":true`) {
		t.Fatalf("reviews reply dry-run output = %s", buf.String())
	}
}

func TestReviewsListRejectsInvalidMaxResultsBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"reviews",
		"list",
		"--package",
		"com.example.app",
		"--max-results",
		"101",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected max results validation error")
	}
	if !strings.Contains(err.Error(), "max results") {
		t.Fatalf("error = %v, want max results validation", err)
	}
}

func TestReviewsReplyRequiresDryRunOrConfirmBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"reviews",
		"reply",
		"--package",
		"com.example.app",
		"--review-id",
		"review-123",
		"--text",
		"Thanks.",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected confirmation validation error")
	}
}

func TestInAppProductsGetRejectsMissingSKUBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"get",
		"--package",
		"com.example.app",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected SKU validation error")
	}
	if !strings.Contains(err.Error(), "SKU") {
		t.Fatalf("error = %v, want SKU validation", err)
	}
}

func TestInAppProductsBatchGetRejectsMissingSKUBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"batch-get",
		"--package",
		"com.example.app",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected SKU validation error")
	}
	if !strings.Contains(err.Error(), "at least one in-app product SKU") {
		t.Fatalf("error = %v, want SKU validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestInAppProductsBatchGetRejectsDuplicateSKUBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"batch-get",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--sku",
		"coins_100",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected duplicate SKU validation error")
	}
	if !strings.Contains(err.Error(), "duplicated") {
		t.Fatalf("error = %v, want duplicate validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestInAppProductsDeleteDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"delete",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--latency-tolerance",
		"latencyTolerant",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"sku":"coins_100"`,
		`"latencyTolerance":"latencyTolerant"`,
		`"dryRun":true`,
		`"deleted":false`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestInAppProductsDeleteRequiresDryRunOrConfirmBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"delete",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirmation validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestInAppProductsBatchDeleteDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"batch-delete",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--sku",
		"coins_500",
		"--latency-tolerance",
		"latencyTolerant",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"skus":["coins_100","coins_500"]`,
		`"latencyTolerance":"latencyTolerant"`,
		`"dryRun":true`,
		`"deleted":false`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestInAppProductsBatchDeleteRequiresDryRunOrConfirmBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"batch-delete",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirmation validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestInAppProductsBatchDeleteRejectsDuplicateSKUBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"batch-delete",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--sku",
		"coins_100",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected duplicate SKU validation error")
	}
	if !strings.Contains(err.Error(), "duplicated") {
		t.Fatalf("error = %v, want duplicate validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestInAppProductsCreateDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"create",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--default-language",
		"en-US",
		"--default-price",
		"USD:1990000",
		"--title",
		"100 coins",
		"--description",
		"A small coin pack.",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"action":"create"`,
		`"dryRun":true`,
		`"purchaseType":"managedUser"`,
		`"priceMicros":"1990000"`,
		`"autoConvertMissingPrices":true`,
		`"created":false`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestInAppProductsCreateRequiresDryRunOrConfirmBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"create",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--default-language",
		"en-US",
		"--default-price",
		"USD:1990000",
		"--title",
		"100 coins",
		"--description",
		"A small coin pack.",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirmation validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation validation", err)
	}
}

func TestInAppProductsCreateRejectsBadPriceBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"create",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--default-language",
		"en-US",
		"--default-price",
		"USD:free",
		"--title",
		"100 coins",
		"--description",
		"A small coin pack.",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected price validation error")
	}
	if !strings.Contains(err.Error(), "price micros") {
		t.Fatalf("error = %v, want price micros validation", err)
	}
}

func TestInAppProductsPatchDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"patch",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--status",
		"inactive",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"action":"patch"`,
		`"dryRun":true`,
		`"status":"inactive"`,
		`"applied":false`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestInAppProductsPatchPriceAndListingDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"patch",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--listing-language",
		"en-US",
		"--default-price",
		"USD:2990000",
		"--title",
		"100 coins",
		"--description",
		"A better coin pack.",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"action":"patch"`,
		`"dryRun":true`,
		`"priceMicros":"2990000"`,
		`"autoConvertMissingPrices":true`,
		`"description":"A better coin pack."`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestInAppProductsPatchRegionalPricesDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"patch",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--regional-price",
		"US:USD:2990000",
		"--regional-price",
		"BR:BRL:9990000",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"regionCode":"US"`,
		`"currency":"BRL"`,
		`"priceMicros":"9990000"`,
		`"autoConvertMissingPrices":true`,
		`"applied":false`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestInAppProductsPatchTaxComplianceDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"patch",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--eea-withdrawal-right-type",
		"WITHDRAWAL_RIGHT_DIGITAL_CONTENT",
		"--tokenized-digital-asset",
		"false",
		"--regional-tax-tier",
		"FR:TAX_TIER_NEWS_1",
		"--regional-streaming-tax",
		"US:STREAMING_TAX_TYPE_TELCO_VIDEO_SALES",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"eeaWithdrawalRightType":"WITHDRAWAL_RIGHT_DIGITAL_CONTENT"`,
		`"isTokenizedDigitalAsset":false`,
		`"taxTier":"TAX_TIER_NEWS_1"`,
		`"streamingTaxType":"STREAMING_TAX_TYPE_TELCO_VIDEO_SALES"`,
		`"dryRun":true`,
		`"applied":false`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestInAppProductsPatchRejectsBadTokenizedDigitalAssetBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"patch",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--tokenized-digital-asset",
		"maybe",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected tokenized digital asset validation error")
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestInAppProductsPatchRequiresDryRunOrConfirmBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"patch",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--status",
		"active",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirmation validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation validation", err)
	}
}

func TestInAppProductsPatchRejectsMissingMutationBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"patch",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected mutation validation error")
	}
	if !strings.Contains(err.Error(), "requires at least one") {
		t.Fatalf("error = %v, want mutation validation", err)
	}
}

func TestInAppProductsPatchRejectsListingWithoutLanguageBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"patch",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--title",
		"100 coins",
		"--description",
		"A better coin pack.",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected listing language validation error")
	}
	if !strings.Contains(err.Error(), "requires --listing-language") {
		t.Fatalf("error = %v, want default language validation", err)
	}
}

func TestInAppProductsPatchRejectsPartialListingBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"patch",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--listing-language",
		"en-US",
		"--title",
		"100 coins",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected partial listing validation error")
	}
	if !strings.Contains(err.Error(), "listing description is required") {
		t.Fatalf("error = %v, want listing description validation", err)
	}
}

func TestInAppProductsPatchRejectsConfirmAndDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"patch",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--status",
		"active",
		"--confirm",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected mutually exclusive flag validation error")
	}
	if !strings.Contains(err.Error(), "cannot be used together") {
		t.Fatalf("error = %v, want mutually exclusive validation", err)
	}
}

func TestSubscriptionsListRejectsInvalidPageSizeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"list",
		"--package",
		"com.example.app",
		"--page-size",
		"1001",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected page size validation error")
	}
	if !strings.Contains(err.Error(), "page size") {
		t.Fatalf("error = %v, want page size validation", err)
	}
}

func TestSubscriptionsBasePlanActivateDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"base-plan",
		"activate",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--latency-tolerance",
		"latencyTolerant",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"action":"activate"`,
		`"basePlanId":"monthly"`,
		`"latencyTolerance":"latencyTolerant"`,
		`"dryRun":true`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionsCreateDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	bodyPath := filepath.Join(t.TempDir(), "subscription.json")
	if err := os.WriteFile(bodyPath, []byte(`{
		"packageName":"ignored",
		"productId":"ignored",
		"listings":[{"languageCode":"en-US","title":"Premium","description":"Full access"}],
		"basePlans":[{
			"basePlanId":"monthly",
			"state":"ACTIVE",
			"autoRenewingBasePlanType":{"billingPeriodDuration":"P1M"},
			"offerTags":[{"tag":"public"}],
			"regionalConfigs":[{"regionCode":"US","newSubscriberAvailability":true,"price":{"currencyCode":"USD","units":"4","nanos":990000000}}],
			"otherRegionsConfig":{"newSubscriberAvailability":true,"usdPrice":{"currencyCode":"USD","units":"4"},"eurPrice":{"currencyCode":"EUR","units":"4"}}
		}]
	}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--from-json",
		bodyPath,
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"action":"create"`,
		`"dryRun":true`,
		`"created":false`,
		`"packageName":"com.example.app"`,
		`"productId":"premium"`,
		`"basePlanId":"monthly"`,
		`"regionsVersion":"2026/05"`,
		`"currencyCode":"USD"`,
		`"nanos":990000000`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, `"state":"ACTIVE"`) || strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect state or auth", output)
	}
}

func TestSubscriptionsCreateBasicFlagsDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--listing",
		"en-US,Premium,Full access",
		"--base-plan-id",
		"monthly",
		"--billing-period",
		"P1M",
		"--price",
		"us:USD:4:990000000",
		"--restricted-country",
		"br",
		"--eea-withdrawal-right-type",
		"WITHDRAWAL_RIGHT_SERVICE",
		"--tokenized-digital-asset",
		"false",
		"--regional-tax-tier",
		"FR:TAX_TIER_NEWS_1",
		"--regional-streaming-tax",
		"US:STREAMING_TAX_TYPE_TELCO_VIDEO_SALES",
		"--offer-tag",
		"public",
		"--regions-version",
		"2026/05",
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
		`"created":false`,
		`"productId":"premium"`,
		`"languageCode":"en-US"`,
		`"title":"Premium"`,
		`"basePlanId":"monthly"`,
		`"type":"autoRenewing"`,
		`"billingPeriodDuration":"P1M"`,
		`"legacyCompatible":true`,
		`"offerTags":["public"]`,
		`"regionCode":"US"`,
		`"newSubscriberAvailability":true`,
		`"currencyCode":"USD"`,
		`"nanos":990000000`,
		`"restrictedCountries":["BR"]`,
		`"eeaWithdrawalRightType":"WITHDRAWAL_RIGHT_SERVICE"`,
		`"isTokenizedDigitalAsset":false`,
		`"taxTier":"TAX_TIER_NEWS_1"`,
		`"streamingTaxType":"STREAMING_TAX_TYPE_TELCO_VIDEO_SALES"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionsCreateBasicFlagsCanDisableLegacyCompatibility(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--listing",
		"en-US,Premium,Full access",
		"--base-plan-id",
		"monthly",
		"--billing-period",
		"P1M",
		"--price",
		"US:USD:4",
		"--legacy-compatible=false",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	if strings.Contains(output, `"legacyCompatible":true`) {
		t.Fatalf("output = %s, did not expect legacy-compatible base plan", output)
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionsCreateBasicPrepaidFlagsDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--listing",
		"en-US,Premium,Full access",
		"--base-plan-id",
		"monthly-prepaid",
		"--prepaid",
		"--billing-period",
		"P1M",
		"--time-extension",
		"TIME_EXTENSION_ACTIVE",
		"--price",
		"us:USD:4:990000000",
		"--offer-tag",
		"public",
		"--regions-version",
		"2026/05",
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
		`"created":false`,
		`"productId":"premium"`,
		`"basePlanId":"monthly-prepaid"`,
		`"type":"prepaid"`,
		`"billingPeriodDuration":"P1M"`,
		`"timeExtension":"TIME_EXTENSION_ACTIVE"`,
		`"offerTags":["public"]`,
		`"regionCode":"US"`,
		`"newSubscriberAvailability":true`,
		`"currencyCode":"USD"`,
		`"nanos":990000000`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, `"legacyCompatible":true`) || strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect legacy compatibility or auth", output)
	}
}

func TestSubscriptionsCreateBasicInstallmentsFlagsDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--listing",
		"en-US,Premium,Full access",
		"--base-plan-id",
		"monthly-installments",
		"--installments",
		"--billing-period",
		"P1M",
		"--committed-payments",
		"12",
		"--renewal-type",
		"RENEWAL_TYPE_RENEWS_WITHOUT_COMMITMENT",
		"--price",
		"us:USD:4:990000000",
		"--offer-tag",
		"public",
		"--regions-version",
		"2026/05",
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
		`"created":false`,
		`"productId":"premium"`,
		`"basePlanId":"monthly-installments"`,
		`"type":"installments"`,
		`"billingPeriodDuration":"P1M"`,
		`"committedPaymentsCount":12`,
		`"renewalType":"RENEWAL_TYPE_RENEWS_WITHOUT_COMMITMENT"`,
		`"offerTags":["public"]`,
		`"regionCode":"US"`,
		`"newSubscriberAvailability":true`,
		`"currencyCode":"USD"`,
		`"nanos":990000000`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, `"legacyCompatible":true`) || strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect legacy compatibility or auth", output)
	}
}

func TestSubscriptionsCreateBasicFlagsRejectTimeExtensionWithoutPrepaidBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--listing",
		"en-US,Premium,Full access",
		"--base-plan-id",
		"monthly",
		"--billing-period",
		"P1M",
		"--time-extension=",
		"--price",
		"US:USD:4",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected time-extension validation error")
	}
	if !strings.Contains(err.Error(), "--time-extension requires --prepaid") {
		t.Fatalf("error = %v, want time-extension prepaid validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestSubscriptionsCreateBasicFlagsRejectDuplicateRestrictedCountryBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--listing",
		"en-US,Premium,Full access",
		"--base-plan-id",
		"monthly",
		"--billing-period",
		"P1M",
		"--price",
		"US:USD:4",
		"--restricted-country",
		"br",
		"--restricted-country",
		"BR",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected duplicate restricted country validation error")
	}
	if !strings.Contains(err.Error(), "restricted country BR is duplicated") {
		t.Fatalf("error = %v, want duplicate restricted country validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestSubscriptionsCreateBasicFlagsRejectInvalidTokenizedDigitalAssetBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--listing",
		"en-US,Premium,Full access",
		"--base-plan-id",
		"monthly",
		"--billing-period",
		"P1M",
		"--price",
		"US:USD:4",
		"--tokenized-digital-asset",
		"maybe",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected tokenized digital asset validation error")
	}
	if !strings.Contains(err.Error(), "invalid syntax") {
		t.Fatalf("error = %v, want bool parse validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestSubscriptionsCreateBasicFlagsRejectCommittedPaymentsWithoutInstallmentsBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--listing",
		"en-US,Premium,Full access",
		"--base-plan-id",
		"monthly",
		"--billing-period",
		"P1M",
		"--committed-payments",
		"0",
		"--price",
		"US:USD:4",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected committed-payments validation error")
	}
	if !strings.Contains(err.Error(), "--committed-payments requires --installments") {
		t.Fatalf("error = %v, want committed-payments installments validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestSubscriptionsCreateBasicFlagsRejectLegacyCompatibleWithInstallmentsBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--listing",
		"en-US,Premium,Full access",
		"--base-plan-id",
		"monthly-installments",
		"--installments",
		"--billing-period",
		"P1M",
		"--committed-payments",
		"12",
		"--renewal-type",
		"RENEWAL_TYPE_RENEWS_WITHOUT_COMMITMENT",
		"--price",
		"US:USD:4",
		"--legacy-compatible=false",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected legacy-compatible validation error")
	}
	if !strings.Contains(err.Error(), "--legacy-compatible cannot be used with --installments") {
		t.Fatalf("error = %v, want legacy-compatible installments validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestSubscriptionsCreateBasicFlagsRejectPrepaidWithInstallmentsBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--listing",
		"en-US,Premium,Full access",
		"--base-plan-id",
		"monthly-installments",
		"--prepaid",
		"--installments",
		"--billing-period",
		"P1M",
		"--price",
		"US:USD:4",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected mutually exclusive base plan type validation error")
	}
	if !strings.Contains(err.Error(), "--prepaid and --installments cannot be used together") {
		t.Fatalf("error = %v, want mutually exclusive base plan type validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestSubscriptionsCreateBasicFlagsRejectLegacyCompatibleWithPrepaidBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--listing",
		"en-US,Premium,Full access",
		"--base-plan-id",
		"monthly-prepaid",
		"--prepaid",
		"--billing-period",
		"P1M",
		"--price",
		"US:USD:4",
		"--legacy-compatible=false",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected legacy-compatible validation error")
	}
	if !strings.Contains(err.Error(), "--legacy-compatible cannot be used with --prepaid") {
		t.Fatalf("error = %v, want legacy-compatible prepaid validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestSubscriptionsCreateRejectsJSONWithBasicFlagsBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	bodyPath := filepath.Join(t.TempDir(), "subscription.json")
	if err := os.WriteFile(bodyPath, []byte(`{
		"listings":[{"languageCode":"en-US","title":"Premium","description":"Full access"}],
		"basePlans":[{"basePlanId":"monthly","autoRenewingBasePlanType":{"billingPeriodDuration":"P1M"},"regionalConfigs":[{"regionCode":"US","newSubscriberAvailability":true,"price":{"currencyCode":"USD","units":"4"}}]}]
	}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--from-json",
		bodyPath,
		"--listing",
		"en-US,Premium,Full access",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected from-json and basic flags validation error")
	}
	if !strings.Contains(err.Error(), "cannot be combined") {
		t.Fatalf("error = %v, want combination validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestSubscriptionsCreateRejectsInvalidBodyBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	bodyPath := filepath.Join(t.TempDir(), "subscription.json")
	if err := os.WriteFile(bodyPath, []byte(`{"listings":[{"languageCode":"en-US","title":"Premium"}]}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--from-json",
		bodyPath,
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected body validation error")
	}
	if !strings.Contains(err.Error(), "requires at least one base plan") {
		t.Fatalf("error = %v, want base plan validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestSubscriptionsBasePlanRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"base-plan",
		"deactivate",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestSubscriptionsBasePlanDeleteDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"base-plan",
		"delete",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"productId":"premium"`,
		`"basePlanId":"monthly"`,
		`"dryRun":true`,
		`"deleted":false`,
		`"steps":["delete base plan"]`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionsBasePlanDeleteRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"base-plan",
		"delete",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestSubscriptionsBatchPatchListingsDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"batch-patch-listings",
		"--package",
		"com.example.app",
		"--listing",
		`premium,en-US,"Premium, Plus","Full access"`,
		"--listing",
		"vip,es-ES,VIP,Acceso completo",
		"--regions-version",
		"2026/05",
		"--latency-tolerance",
		"latencyTolerant",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"productId":"premium"`,
		`"languageCode":"en-US"`,
		`"title":"Premium, Plus"`,
		`"productId":"vip"`,
		`"regionsVersion":"2026/05"`,
		`"latencyTolerance":"latencyTolerant"`,
		`"dryRun":true`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionsBatchPatchListingsRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"batch-patch-listings",
		"--package",
		"com.example.app",
		"--listing",
		"premium,en-US,Premium,Full access",
		"--regions-version",
		"2026/05",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestSubscriptionsBatchPatchListingsRejectsMultipleCSVRecordsBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"batch-patch-listings",
		"--package",
		"com.example.app",
		"--listing",
		"premium,en-US,Premium,Full access\nvip,en-US,VIP,All access",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected CSV record validation error")
	}
	if !strings.Contains(err.Error(), "exactly one CSV record") {
		t.Fatalf("error = %v, want CSV record validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestSubscriptionsBasePlanBatchDeactivateDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"base-plan",
		"batch-deactivate",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--base-plan-id",
		"annual",
		"--latency-tolerance",
		"latencyTolerant",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"action":"deactivate"`,
		`"basePlanId":"monthly"`,
		`"basePlanId":"annual"`,
		`"latencyTolerance":"latencyTolerant"`,
		`"dryRun":true`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionsBasePlanBatchDeactivateDryRunInfersWildcardProductBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"base-plan",
		"batch-deactivate",
		"--package",
		"com.example.app",
		"--base-plan",
		"premium/monthly",
		"--base-plan",
		"vip/annual",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"productId":"-"`,
		`"productId":"premium"`,
		`"productId":"vip"`,
		`"basePlanId":"monthly"`,
		`"basePlanId":"annual"`,
		`"dryRun":true`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionsBasePlanBatchActivateRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"base-plan",
		"batch-activate",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestSubscriptionsBasePlanBatchActivateRejectsMissingBasePlanBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"base-plan",
		"batch-activate",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected missing base plan validation error")
	}
	if !strings.Contains(err.Error(), "at least one subscription base plan is required") {
		t.Fatalf("error = %v, want missing base plan validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestSubscriptionsBasePlanBatchMigratePricesDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"base-plan",
		"batch-migrate-prices",
		"--package",
		"com.example.app",
		"--regions-version",
		"2026/05",
		"--migration",
		"premium/monthly/US/2026-05-01T00:00:00Z",
		"--migration",
		"premium/monthly/BR/2026-05-01T00:00:00Z",
		"--price-increase-type",
		"optOut",
		"--latency-tolerance",
		"latencyTolerant",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"productId":"premium"`,
		`"basePlanId":"monthly"`,
		`"regionCode":"US"`,
		`"regionCode":"BR"`,
		`"priceIncreaseType":"optOut"`,
		`"regionsVersion":"2026/05"`,
		`"dryRun":true`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionsBasePlanBatchMigratePricesRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"base-plan",
		"batch-migrate-prices",
		"--package",
		"com.example.app",
		"--regions-version",
		"2026/05",
		"--migration",
		"premium/monthly/US/2026-05-01T00:00:00Z",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestSubscriptionsBasePlanBatchPatchPricesDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"base-plan",
		"batch-patch-prices",
		"--package",
		"com.example.app",
		"--regions-version",
		"2026/05",
		"--price",
		"premium/monthly/US:USD:4:990000000",
		"--price",
		"premium/monthly/BR:BRL:19:990000000",
		"--latency-tolerance",
		"latencyTolerant",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"productId":"premium"`,
		`"basePlanId":"monthly"`,
		`"regionCode":"US"`,
		`"currencyCode":"USD"`,
		`"nanos":990000000`,
		`"regionsVersion":"2026/05"`,
		`"dryRun":true`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionsBasePlanBatchPatchPricesRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"base-plan",
		"batch-patch-prices",
		"--package",
		"com.example.app",
		"--regions-version",
		"2026/05",
		"--price",
		"premium/monthly/US:USD:4:990000000",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductsListRejectsInvalidPageSizeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"list",
		"--package",
		"com.example.app",
		"--page-size",
		"1001",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected page size validation error")
	}
	if !strings.Contains(err.Error(), "page size") {
		t.Fatalf("error = %v, want page size validation", err)
	}
}

func TestOneTimeProductsGetRejectsInvalidProductIDBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"get",
		"--package",
		"com.example.app",
		"--product-id",
		"Coins",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected product ID validation error")
	}
	if !strings.Contains(err.Error(), "one-time product ID") {
		t.Fatalf("error = %v, want product ID validation", err)
	}
}

func TestOneTimeProductsBatchGetRejectsDuplicatesBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"batch-get",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--product-id",
		"coins_100",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected duplicate product ID validation error")
	}
	if !strings.Contains(err.Error(), "duplicated") {
		t.Fatalf("error = %v, want duplicate validation", err)
	}
}

func TestOneTimeProductsBatchGetRejectsMissingProductIDBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"batch-get",
		"--package",
		"com.example.app",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected missing product ID validation error")
	}
	if !strings.Contains(err.Error(), "at least one") {
		t.Fatalf("error = %v, want missing product ID validation", err)
	}
}

func TestOneTimeProductsCreateDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	bodyPath := filepath.Join(t.TempDir(), "one-time-product.json")
	if err := os.WriteFile(bodyPath, []byte(`{
		"packageName":"ignored.by.flags",
		"productId":"ignored_by_flags",
		"listings":[{"languageCode":"en-US","title":"100 coins","description":"Buy coins."}],
		"purchaseOptions":[{
			"purchaseOptionId":"buy",
			"state":"ACTIVE",
			"buyOption":{"legacyCompatible":true},
			"regionalPricingAndAvailabilityConfigs":[{"regionCode":"US","availability":"AVAILABLE","price":{"currencyCode":"USD","units":"1","nanos":990000000}}],
			"newRegionsConfig":{"availability":"AVAILABLE","usdPrice":{"currencyCode":"USD","units":"1"},"eurPrice":{"currencyCode":"EUR","units":"1"}}
		}]
	}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--from-json",
		bodyPath,
		"--regions-version",
		"2026/05",
		"--latency-tolerance",
		"latencyTolerant",
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
		`"created":false`,
		`"packageName":"com.example.app"`,
		`"productId":"coins_100"`,
		`"purchaseOptionId":"buy"`,
		`"regionsVersion":"2026/05"`,
		`"latencyTolerance":"latencyTolerant"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
	if strings.Contains(output, `"state":"ACTIVE"`) {
		t.Fatalf("output = %s, did not expect output-only state from input JSON", output)
	}
}

func TestOneTimeProductsCreateBasicFlagsDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--listing",
		"en-US,100 coins,Buy coins.",
		"--price",
		"us:USD:1:990000000",
		"--purchase-option-id",
		"buy",
		"--offer-tag",
		"public",
		"--multi-quantity",
		"--regions-version",
		"2026/05",
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
		`"created":false`,
		`"productId":"coins_100"`,
		`"languageCode":"en-US"`,
		`"title":"100 coins"`,
		`"purchaseOptionId":"buy"`,
		`"type":"buy"`,
		`"legacyCompatible":true`,
		`"multiQuantityEnabled":true`,
		`"offerTags":["public"]`,
		`"regionCode":"US"`,
		`"availability":"available"`,
		`"currencyCode":"USD"`,
		`"nanos":990000000`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestOneTimeProductsCreateRejectsJSONWithBasicFlagsBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	bodyPath := filepath.Join(t.TempDir(), "one-time-product.json")
	if err := os.WriteFile(bodyPath, []byte(`{
		"listings":[{"languageCode":"en-US","title":"100 coins","description":"Buy coins."}],
		"purchaseOptions":[{"purchaseOptionId":"buy","buyOption":{"legacyCompatible":true},"regionalPricingAndAvailabilityConfigs":[{"regionCode":"US","availability":"AVAILABLE","price":{"currencyCode":"USD","units":"1"}}]}]
	}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--from-json",
		bodyPath,
		"--listing",
		"en-US,100 coins,Buy coins.",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected from-json and basic flags validation error")
	}
	if !strings.Contains(err.Error(), "cannot be combined") {
		t.Fatalf("error = %v, want combination validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductsCreateBasicFlagsRejectsTooManyOfferTagsBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	args := []string{
		"one-time-products",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--listing",
		"en-US,100 coins,Buy coins.",
		"--price",
		"US:USD:1",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	}
	for index := range 21 {
		args = append(args, "--offer-tag", fmt.Sprintf("tag%d", index))
	}
	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected offer tag limit validation error")
	}
	if !strings.Contains(err.Error(), "at most 20 offer tags") {
		t.Fatalf("error = %v, want offer tag limit validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductsCreateRejectsInvalidBodyBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	bodyPath := filepath.Join(t.TempDir(), "one-time-product.json")
	if err := os.WriteFile(bodyPath, []byte(`{"listings":[{"languageCode":"en-US","title":"100 coins","description":"Buy coins."}]}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--from-json",
		bodyPath,
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected product body validation error")
	}
	if !strings.Contains(err.Error(), "requires at least one") {
		t.Fatalf("error = %v, want body validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductsPatchDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"patch",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--listing-language",
		"en-US",
		"--title",
		"100 coins",
		"--description",
		"Buy a stack of coins.",
		"--regions-version",
		"2026/05",
		"--latency-tolerance",
		"latencyTolerant",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"productId":"coins_100"`,
		`"languageCode":"en-US"`,
		`"title":"100 coins"`,
		`"updateMask":"listings"`,
		`"latencyTolerance":"latencyTolerant"`,
		`"dryRun":true`,
		`"applied":false`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestOneTimeProductsPatchRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"patch",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--listing-language",
		"en-US",
		"--title",
		"100 coins",
		"--regions-version",
		"2026/05",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductsBatchPatchListingsDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"batch-patch-listings",
		"--package",
		"com.example.app",
		"--listing",
		"coins_100,en-US,100 coins,Buy coins.",
		"--listing",
		"coins_500,es-ES,500 monedas,Compra monedas.",
		"--regions-version",
		"2026/05",
		"--latency-tolerance",
		"latencyTolerant",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"productId":"coins_100"`,
		`"languageCode":"es-ES"`,
		`"title":"500 monedas"`,
		`"updateMask":"listings"`,
		`"latencyTolerance":"latencyTolerant"`,
		`"dryRun":true`,
		`"applied":false`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestOneTimeProductsBatchPatchListingsRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"batch-patch-listings",
		"--package",
		"com.example.app",
		"--listing",
		"coins_100,en-US,100 coins,Buy coins.",
		"--regions-version",
		"2026/05",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductsBatchPatchListingsRejectsMultipleCSVRecordsBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"batch-patch-listings",
		"--package",
		"com.example.app",
		"--listing",
		"coins_100,en-US,100 coins,Buy coins.\ncoins_500,en-US,500 coins,Buy more coins.",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected CSV record validation error")
	}
	if !strings.Contains(err.Error(), "exactly one CSV record") {
		t.Fatalf("error = %v, want CSV record validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductsDeleteDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"delete",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--latency-tolerance",
		"latencyTolerant",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"productId":"coins_100"`,
		`"latencyTolerance":"latencyTolerant"`,
		`"dryRun":true`,
		`"deleted":false`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestOneTimeProductsDeleteRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"delete",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
}

func TestOneTimeProductsBatchDeleteDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"batch-delete",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--product-id",
		"coins_500",
		"--latency-tolerance",
		"latencyTolerant",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"productIds":["coins_100","coins_500"]`,
		`"latencyTolerance":"latencyTolerant"`,
		`"dryRun":true`,
		`"deleted":false`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestOneTimeProductsBatchDeleteRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"batch-delete",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
}

func TestOneTimeProductsBatchDeleteRejectsDuplicatesBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"batch-delete",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--product-id",
		"coins_100",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected duplicate product ID validation error")
	}
	if !strings.Contains(err.Error(), "duplicated") {
		t.Fatalf("error = %v, want duplicate validation", err)
	}
}

func TestOneTimeProductsPurchaseOptionDeactivateDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"purchase-option",
		"deactivate",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--latency-tolerance",
		"latencyTolerant",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"action":"deactivate"`,
		`"purchaseOptionId":"buy"`,
		`"latencyTolerance":"latencyTolerant"`,
		`"dryRun":true`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestOneTimeProductsPurchaseOptionRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"purchase-option",
		"activate",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductsPurchaseOptionBatchDeleteDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"purchase-option",
		"batch-delete",
		"--package",
		"com.example.app",
		"--product-id",
		"-",
		"--purchase-option",
		"coins_100/buy",
		"--purchase-option",
		"coins_500/rent",
		"--latency-tolerance",
		"latencyTolerant",
		"--force",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"parentProductId":"-"`,
		`"purchaseOptionId":"buy"`,
		`"latencyTolerance":"latencyTolerant"`,
		`"force":true`,
		`"dryRun":true`,
		`"deleted":false`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestOneTimeProductsPurchaseOptionBatchDeleteRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"purchase-option",
		"batch-delete",
		"--package",
		"com.example.app",
		"--purchase-option",
		"coins_100/buy",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductsPurchaseOptionBatchDeleteInfersSingleParentBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"purchase-option",
		"batch-delete",
		"--package",
		"com.example.app",
		"--purchase-option",
		"coins_100/buy",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, `"parentProductId":"coins_100"`) {
		t.Fatalf("output = %s, want inferred parent product", output)
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestOneTimeProductsPurchaseOptionBatchDeleteRejectsRepeatedProductBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"purchase-option",
		"batch-delete",
		"--package",
		"com.example.app",
		"--purchase-option",
		"coins_100/buy",
		"--purchase-option",
		"coins_100/rent",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected repeated product validation error")
	}
	if !strings.Contains(err.Error(), "at most one request per one-time product") {
		t.Fatalf("error = %v, want repeated product validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductOffersListRejectsInvalidWildcardParentBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"list",
		"--package",
		"com.example.app",
		"--product-id",
		"-",
		"--purchase-option-id",
		"buy",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected wildcard validation error")
	}
	if !strings.Contains(err.Error(), "purchase option ID") {
		t.Fatalf("error = %v, want purchase option validation", err)
	}
}

func TestOneTimeProductOffersGetRejectsInvalidOfferIDBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"get",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"Intro",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected offer ID validation error")
	}
	if !strings.Contains(err.Error(), "one-time product offer ID") {
		t.Fatalf("error = %v, want offer ID validation", err)
	}
}

func TestOneTimeProductOffersBatchGetRejectsMissingOfferBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-get",
		"--package",
		"com.example.app",
		"--product-id",
		"-",
		"--purchase-option-id",
		"-",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected offer validation error")
	}
	if !strings.Contains(err.Error(), "at least one one-time product offer") {
		t.Fatalf("error = %v, want missing offer validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestOneTimeProductOffersBatchGetRejectsParentMismatchBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-get",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer",
		"coins_500/buy/intro",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected parent mismatch validation error")
	}
	if !strings.Contains(err.Error(), "does not match parent product ID") {
		t.Fatalf("error = %v, want parent product validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestOneTimeProductOffersBatchGetRejectsInvalidOfferIDBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-get",
		"--package",
		"com.example.app",
		"--product-id",
		"-",
		"--purchase-option-id",
		"-",
		"--offer",
		"coins_100/buy/Intro",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected offer ID validation error")
	}
	if !strings.Contains(err.Error(), "one-time product offer ID") {
		t.Fatalf("error = %v, want offer ID validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestOneTimeProductOffersCreateDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	bodyPath := filepath.Join(t.TempDir(), "offer.json")
	if err := os.WriteFile(bodyPath, []byte(`{
		"packageName":"ignored.by.flags",
		"productId":"ignored_by_flags",
		"purchaseOptionId":"ignored",
		"offerId":"ignored",
		"state":"ACTIVE",
		"discountedOffer":{"startTime":"2026-06-01T00:00:00Z","endTime":"2026-07-01T00:00:00Z","redemptionLimit":"5"},
		"regionalPricingAndAvailabilityConfigs":[{"regionCode":"US","availability":"AVAILABLE","relativeDiscount":0.5}]
	}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"intro",
		"--from-json",
		bodyPath,
		"--regions-version",
		"2026/05",
		"--latency-tolerance",
		"latencyTolerant",
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
		`"created":false`,
		`"productId":"coins_100"`,
		`"purchaseOptionId":"buy"`,
		`"offerId":"intro"`,
		`"regionsVersion":"2026/05"`,
		`"latencyTolerance":"latencyTolerant"`,
		`"relativeDiscount":0.5`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
	if strings.Contains(output, `"state":"ACTIVE"`) {
		t.Fatalf("output = %s, did not expect output-only state from input JSON", output)
	}
}

func TestOneTimeProductOffersCreateBasicRelativeDiscountDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"intro",
		"--offer-tag",
		"public",
		"--start-time",
		"2026-06-01T00:00:00Z",
		"--end-time",
		"2026-07-01T00:00:00Z",
		"--redemption-limit",
		"5",
		"--relative-discount",
		"us:0.5",
		"--regions-version",
		"2026/05",
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
		`"created":false`,
		`"productId":"coins_100"`,
		`"purchaseOptionId":"buy"`,
		`"offerId":"intro"`,
		`"type":"discounted"`,
		`"offerTags":["public"]`,
		`"startTime":"2026-06-01T00:00:00Z"`,
		`"endTime":"2026-07-01T00:00:00Z"`,
		`"redemptionLimit":5`,
		`"regionCode":"US"`,
		`"availability":"available"`,
		`"relativeDiscount":0.5`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestOneTimeProductOffersCreateBasicAbsoluteDiscountDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"intro",
		"--start-time",
		"2026-06-01T00:00:00Z",
		"--end-time",
		"2026-07-01T00:00:00Z",
		"--absolute-discount",
		"us:USD:1:500000000",
		"--regions-version",
		"2026/05",
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
		`"created":false`,
		`"type":"discounted"`,
		`"regionCode":"US"`,
		`"availability":"available"`,
		`"absoluteDiscount":{"currencyCode":"USD","units":1,"nanos":500000000}`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestOneTimeProductOffersCreateBasicNoOverrideDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"intro",
		"--start-time",
		"2026-06-01T00:00:00Z",
		"--end-time",
		"2026-07-01T00:00:00Z",
		"--no-override",
		"us",
		"--regions-version",
		"2026/05",
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
		`"created":false`,
		`"type":"discounted"`,
		`"regionCode":"US"`,
		`"availability":"available"`,
		`"noOverride":true`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestOneTimeProductOffersCreateBasicPreOrderDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"preorder",
		"--pre-order",
		"--start-time",
		"2026-06-01T00:00:00Z",
		"--end-time",
		"2026-07-01T00:00:00Z",
		"--release-time",
		"2026-08-01T00:00:00Z",
		"--price-change-behavior",
		"PRE_ORDER_PRICE_CHANGE_BEHAVIOR_NEW_ORDERS_ONLY",
		"--no-override",
		"us",
		"--regions-version",
		"2026/05",
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
		`"created":false`,
		`"type":"preOrder"`,
		`"preOrderOffer":{"startTime":"2026-06-01T00:00:00Z","endTime":"2026-07-01T00:00:00Z","releaseTime":"2026-08-01T00:00:00Z","priceChangeBehavior":"PRE_ORDER_PRICE_CHANGE_BEHAVIOR_NEW_ORDERS_ONLY"}`,
		`"regionCode":"US"`,
		`"availability":"available"`,
		`"noOverride":true`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestOneTimeProductOffersCreateBasicMixedDiscountModesDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"intro",
		"--relative-discount",
		"US:0.5",
		"--absolute-discount",
		"JP:JPY:100",
		"--no-override",
		"BR",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"regionCode":"US"`,
		`"relativeDiscount":0.5`,
		`"regionCode":"JP"`,
		`"absoluteDiscount":{"currencyCode":"JPY","units":100}`,
		`"regionCode":"BR"`,
		`"noOverride":true`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestOneTimeProductOffersCreateBasicFlagsRejectsDuplicateDiscountRegionBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"intro",
		"--relative-discount",
		"US:0.5",
		"--absolute-discount",
		"US:USD:1",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected duplicated region validation error")
	}
	if !strings.Contains(err.Error(), "region US is duplicated") {
		t.Fatalf("error = %v, want duplicate region validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductOffersCreateBasicFlagsRejectsReleaseTimeWithoutPreOrderBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"intro",
		"--relative-discount",
		"US:0.5",
		"--release-time",
		"not-a-time",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected release-time validation error")
	}
	if !strings.Contains(err.Error(), "--release-time requires --pre-order") {
		t.Fatalf("error = %v, want release-time pre-order validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestOneTimeProductOffersCreateBasicFlagsRejectsRedemptionLimitWithPreOrderBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"preorder",
		"--pre-order",
		"--start-time",
		"2026-06-01T00:00:00Z",
		"--end-time",
		"2026-07-01T00:00:00Z",
		"--release-time",
		"2026-08-01T00:00:00Z",
		"--price-change-behavior",
		"PRE_ORDER_PRICE_CHANGE_BEHAVIOR_NEW_ORDERS_ONLY",
		"--no-override",
		"US",
		"--redemption-limit",
		"0",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected redemption-limit validation error")
	}
	if !strings.Contains(err.Error(), "--redemption-limit cannot be used with --pre-order") {
		t.Fatalf("error = %v, want redemption-limit pre-order validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestOneTimeProductOffersCreateBasicFlagsRejectsPriceBehaviorWithoutPreOrderBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"intro",
		"--relative-discount",
		"US:0.5",
		"--price-change-behavior=",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected price-change-behavior validation error")
	}
	if !strings.Contains(err.Error(), "--price-change-behavior requires --pre-order") {
		t.Fatalf("error = %v, want price-change-behavior pre-order validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestOneTimeProductOffersCreateRejectsJSONWithBasicFlagsBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	bodyPath := filepath.Join(t.TempDir(), "offer.json")
	if err := os.WriteFile(bodyPath, []byte(`{
		"discountedOffer":{"startTime":"2026-06-01T00:00:00Z","endTime":"2026-07-01T00:00:00Z"},
		"regionalPricingAndAvailabilityConfigs":[{"regionCode":"US","availability":"AVAILABLE","relativeDiscount":0.5}]
	}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"intro",
		"--from-json",
		bodyPath,
		"--relative-discount",
		"US:0.5",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected from-json and basic flags validation error")
	}
	if !strings.Contains(err.Error(), "cannot be combined") {
		t.Fatalf("error = %v, want combination validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductOffersCreateBasicFlagsRejectsInvalidTimeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	for _, tc := range []struct {
		name string
		flag string
	}{
		{name: "start", flag: "--start-time"},
		{name: "end", flag: "--end-time"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := newRootCommand(&buf)
			cmd.SetArgs([]string{
				"one-time-product-offers",
				"create",
				"--package",
				"com.example.app",
				"--product-id",
				"coins_100",
				"--purchase-option-id",
				"buy",
				"--offer-id",
				"intro",
				"--relative-discount",
				"US:0.5",
				tc.flag,
				"not-a-time",
				"--regions-version",
				"2026/05",
				"--dry-run",
				"--output",
				"json",
			})

			err := cmd.Execute()
			if err == nil {
				t.Fatal("expected RFC3339 validation error")
			}
			if !strings.Contains(err.Error(), "must be RFC3339") {
				t.Fatalf("error = %v, want RFC3339 validation", err)
			}
			if strings.Contains(err.Error(), "no active auth profile") {
				t.Fatalf("error = %v, did not expect auth error", err)
			}
		})
	}
}

func TestOneTimeProductOffersCreateBasicFlagsRejectsInvalidRelativeDiscountBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"intro",
		"--relative-discount",
		"US:0",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected relative discount validation error")
	}
	if !strings.Contains(err.Error(), "relative discount must be greater than 0 and less than 1") {
		t.Fatalf("error = %v, want relative discount range validation", err)
	}
	if strings.Contains(err.Error(), "requires exactly one") {
		t.Fatalf("error = %v, did not expect downstream price mode error", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductOffersCreateBasicFlagsRejectsMalformedAbsoluteDiscountBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"intro",
		"--absolute-discount",
		"US:USD:x",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected absolute discount format validation error")
	}
	if !strings.Contains(err.Error(), "absolute discount must use REGION:CURRENCY:UNITS[:NANOS]") {
		t.Fatalf("error = %v, want absolute discount format validation", err)
	}
	if strings.Contains(err.Error(), "price units") {
		t.Fatalf("error = %v, did not expect generic price units error", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductOffersCreateRejectsInvalidBodyBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	bodyPath := filepath.Join(t.TempDir(), "offer.json")
	if err := os.WriteFile(bodyPath, []byte(`{"discountedOffer":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"intro",
		"--from-json",
		bodyPath,
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected offer body validation error")
	}
	if !strings.Contains(err.Error(), "requires at least one regional config") {
		t.Fatalf("error = %v, want regional config validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductOffersBatchDeleteDryRunInfersParentsBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-delete",
		"--package",
		"com.example.app",
		"--offer",
		"coins_100/buy/intro",
		"--offer",
		"coins_100/rent/preorder",
		"--latency-tolerance",
		"latencyTolerant",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"productId":"coins_100"`,
		`"purchaseOptionId":"-"`,
		`"offerId":"intro"`,
		`"latencyTolerance":"latencyTolerant"`,
		`"dryRun":true`,
		`"deleted":false`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestOneTimeProductOffersBatchDeleteRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-delete",
		"--package",
		"com.example.app",
		"--offer",
		"coins_100/buy/intro",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirmation gate")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductOffersBatchDeleteInfersOmittedPurchaseOptionBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-delete",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--offer",
		"coins_100/buy/intro",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, `"purchaseOptionId":"buy"`) {
		t.Fatalf("output = %s, want inferred purchase option", output)
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestOneTimeProductOffersBatchPatchAvailabilityDryRunInfersParentsBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-patch-availability",
		"--package",
		"com.example.app",
		"--availability",
		"coins_100/buy/intro/us:noLongerAvailable",
		"--availability",
		"coins_100/rent/winback/FR:available",
		"--regions-version",
		"2026/05",
		"--latency-tolerance",
		"latencyTolerant",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"productId":"coins_100"`,
		`"purchaseOptionId":"-"`,
		`"offerId":"intro"`,
		`"regionCode":"US"`,
		`"availability":"noLongerAvailable"`,
		`"updateMask":"regionalPricingAndAvailabilityConfigs"`,
		`"regionsVersion":"2026/05"`,
		`"latencyTolerance":"latencyTolerant"`,
		`"dryRun":true`,
		`"applied":false`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestOneTimeProductOffersBatchPatchAvailabilityRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-patch-availability",
		"--package",
		"com.example.app",
		"--availability",
		"coins_100/buy/intro/US:available",
		"--regions-version",
		"2026/05",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirmation gate")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductOffersBatchPatchAvailabilityRejectsMalformedPatchBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-patch-availability",
		"--package",
		"com.example.app",
		"--availability",
		"coins_100/buy/intro:available",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected availability format validation error")
	}
	if !strings.Contains(err.Error(), "productId/purchaseOptionId/offerId/REGION:available|noLongerAvailable") {
		t.Fatalf("error = %v, want availability format validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductOffersBatchPatchRelativeDiscountsDryRunInfersParentsBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-patch-relative-discounts",
		"--package",
		"com.example.app",
		"--relative-discount",
		"coins_100/buy/intro/us:0.5",
		"--relative-discount",
		"coins_100/rent/winback/FR:0.25",
		"--regions-version",
		"2026/05",
		"--latency-tolerance",
		"latencyTolerant",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"productId":"coins_100"`,
		`"purchaseOptionId":"-"`,
		`"offerId":"intro"`,
		`"regionCode":"US"`,
		`"relativeDiscount":0.5`,
		`"updateMask":"regionalPricingAndAvailabilityConfigs"`,
		`"regionsVersion":"2026/05"`,
		`"latencyTolerance":"latencyTolerant"`,
		`"dryRun":true`,
		`"applied":false`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestOneTimeProductOffersBatchPatchRelativeDiscountsRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-patch-relative-discounts",
		"--package",
		"com.example.app",
		"--relative-discount",
		"coins_100/buy/intro/US:0.5",
		"--regions-version",
		"2026/05",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirmation gate")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductOffersBatchPatchRelativeDiscountsRejectsMalformedPatchBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-patch-relative-discounts",
		"--package",
		"com.example.app",
		"--relative-discount",
		"coins_100/buy/intro:0.5",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected relative discount format validation error")
	}
	if !strings.Contains(err.Error(), "productId/purchaseOptionId/offerId/REGION:0.5") {
		t.Fatalf("error = %v, want relative discount format validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductOffersBatchPatchAbsoluteDiscountsDryRunInfersParentsBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-patch-absolute-discounts",
		"--package",
		"com.example.app",
		"--absolute-discount",
		"coins_100/buy/intro/us:USD:1:500000000",
		"--absolute-discount",
		"coins_100/rent/winback/FR:EUR:2",
		"--regions-version",
		"2026/05",
		"--latency-tolerance",
		"latencyTolerant",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"productId":"coins_100"`,
		`"purchaseOptionId":"-"`,
		`"offerId":"intro"`,
		`"regionCode":"US"`,
		`"currencyCode":"USD"`,
		`"units":1`,
		`"nanos":500000000`,
		`"updateMask":"regionalPricingAndAvailabilityConfigs"`,
		`"regionsVersion":"2026/05"`,
		`"latencyTolerance":"latencyTolerant"`,
		`"dryRun":true`,
		`"applied":false`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestOneTimeProductOffersBatchPatchAbsoluteDiscountsRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-patch-absolute-discounts",
		"--package",
		"com.example.app",
		"--absolute-discount",
		"coins_100/buy/intro/US:USD:1",
		"--regions-version",
		"2026/05",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirmation gate")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductOffersBatchPatchAbsoluteDiscountsRejectsMalformedPatchBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-patch-absolute-discounts",
		"--package",
		"com.example.app",
		"--absolute-discount",
		"coins_100/buy/intro:USD:1",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected absolute discount format validation error")
	}
	if !strings.Contains(err.Error(), "productId/purchaseOptionId/offerId/REGION:CURRENCY:UNITS[:NANOS]") {
		t.Fatalf("error = %v, want absolute discount format validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductOffersBatchPatchAbsoluteDiscountsRejectsMalformedMoneyBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-patch-absolute-discounts",
		"--package",
		"com.example.app",
		"--absolute-discount",
		"coins_100/buy/intro/US:USD",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected absolute discount money format validation error")
	}
	if !strings.Contains(err.Error(), "one-time product offer absolute discount must use productId/purchaseOptionId/offerId/REGION:CURRENCY:UNITS[:NANOS]") {
		t.Fatalf("error = %v, want absolute discount format validation", err)
	}
	if strings.Contains(err.Error(), "purchase option price") {
		t.Fatalf("error = %v, did not expect purchase option price message", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductOffersBatchPatchNoOverridesDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-patch-no-overrides",
		"--package",
		"com.example.app",
		"--no-override",
		"coins_100/buy/intro/us",
		"--no-override",
		"coins_100/rent/winback/FR",
		"--regions-version",
		"2026/05",
		"--latency-tolerance",
		"latencyTolerant",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"productId":"coins_100"`,
		`"purchaseOptionId":"-"`,
		`"offerId":"intro"`,
		`"regionCode":"US"`,
		`"noOverride":true`,
		`"updateMask":"regionalPricingAndAvailabilityConfigs"`,
		`"regionsVersion":"2026/05"`,
		`"latencyTolerance":"latencyTolerant"`,
		`"dryRun":true`,
		`"applied":false`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestOneTimeProductOffersBatchPatchNoOverridesRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-patch-no-overrides",
		"--package",
		"com.example.app",
		"--no-override",
		"coins_100/buy/intro/US",
		"--regions-version",
		"2026/05",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirmation gate")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestPurchaseOptionBatchPatchAvailabilityDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"purchase-option",
		"batch-patch-availability",
		"--package",
		"com.example.app",
		"--availability",
		"coins_100/buy/us:noLongerAvailable",
		"--availability",
		"coins_100/buy/FR:availableForOffersOnly",
		"--regions-version",
		"2026/05",
		"--latency-tolerance",
		"latencyTolerant",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"productId":"coins_100"`,
		`"purchaseOptionId":"buy"`,
		`"regionCode":"US"`,
		`"availability":"noLongerAvailable"`,
		`"availability":"availableForOffersOnly"`,
		`"updateMask":"purchaseOptions"`,
		`"regionsVersion":"2026/05"`,
		`"latencyTolerance":"latencyTolerant"`,
		`"dryRun":true`,
		`"applied":false`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestPurchaseOptionBatchPatchAvailabilityRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"purchase-option",
		"batch-patch-availability",
		"--package",
		"com.example.app",
		"--availability",
		"coins_100/buy/US:available",
		"--regions-version",
		"2026/05",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirmation gate")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestPurchaseOptionBatchPatchAvailabilityRejectsMalformedPatchBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"purchase-option",
		"batch-patch-availability",
		"--package",
		"com.example.app",
		"--availability",
		"coins_100/buy:available",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected availability format validation error")
	}
	if !strings.Contains(err.Error(), "productId/purchaseOptionId/REGION") {
		t.Fatalf("error = %v, want availability format validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestPurchaseOptionBatchPatchPricesDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"purchase-option",
		"batch-patch-prices",
		"--package",
		"com.example.app",
		"--price",
		"coins_100/buy/us:USD:3:490000000",
		"--price",
		"coins_100/buy/FR:EUR:2",
		"--regions-version",
		"2026/05",
		"--latency-tolerance",
		"latencyTolerant",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"productId":"coins_100"`,
		`"purchaseOptionId":"buy"`,
		`"regionCode":"US"`,
		`"currencyCode":"USD"`,
		`"units":3`,
		`"nanos":490000000`,
		`"updateMask":"purchaseOptions"`,
		`"regionsVersion":"2026/05"`,
		`"latencyTolerance":"latencyTolerant"`,
		`"dryRun":true`,
		`"applied":false`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestPurchaseOptionBatchPatchPricesRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"purchase-option",
		"batch-patch-prices",
		"--package",
		"com.example.app",
		"--price",
		"coins_100/buy/US:USD:3:490000000",
		"--regions-version",
		"2026/05",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirmation gate")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestPurchaseOptionBatchPatchPricesRejectsMalformedPatchBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"purchase-option",
		"batch-patch-prices",
		"--package",
		"com.example.app",
		"--price",
		"coins_100/buy:USD:3",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected price format validation error")
	}
	if !strings.Contains(err.Error(), "productId/purchaseOptionId/REGION") {
		t.Fatalf("error = %v, want price format validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductOffersBatchDeactivateDryRunInfersParentsBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-deactivate",
		"--package",
		"com.example.app",
		"--offer",
		"coins_100/buy/intro",
		"--offer",
		"coins_100/rent/winback",
		"--latency-tolerance",
		"latencyTolerant",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"action":"deactivate"`,
		`"productId":"coins_100"`,
		`"purchaseOptionId":"-"`,
		`"offerId":"intro"`,
		`"latencyTolerance":"latencyTolerant"`,
		`"dryRun":true`,
		`"applied":false`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestOneTimeProductOffersBatchActivateRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-activate",
		"--package",
		"com.example.app",
		"--offer",
		"coins_100/buy/intro",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirmation gate")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductOffersBatchCancelDryRunCallsOutPendingOrders(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-cancel",
		"--package",
		"com.example.app",
		"--offer",
		"coins_100/buy/preorder",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "pending orders") {
		t.Fatalf("output = %s, want pending orders warning in plan", output)
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth error", output)
	}
}

func TestOneTimeProductOffersDeactivateDryRunPrintsPlanBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"deactivate",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"intro",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	var result play.OneTimeProductOfferStateUpdateResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Unmarshal() error = %v\n%s", err, buf.String())
	}
	if result.Action != play.OneTimeProductOfferStateActionDeactivate {
		t.Fatalf("Action = %q, want deactivate", result.Action)
	}
	if result.Applied {
		t.Fatal("Applied = true, want dry-run plan")
	}
}

func TestOneTimeProductOffersCancelRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"cancel",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"preorder",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirmation gate")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestVitalsMetricSetGetRejectsUnsupportedMetricSetBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"vitals",
		"metric-set",
		"get",
		"--package",
		"com.example.app",
		"--metric-set",
		"crashes",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected metric set validation error")
	}
	if !strings.Contains(err.Error(), "unsupported vitals metric set") {
		t.Fatalf("error = %v, want metric set validation", err)
	}
}

func TestVitalsMetricSetGetRejectsMissingMetricSetBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"vitals",
		"metric-set",
		"get",
		"--package",
		"com.example.app",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected metric set validation error")
	}
	if !strings.Contains(err.Error(), "vitals metric set is required") {
		t.Fatalf("error = %v, want required metric set validation", err)
	}
}

func TestVitalsMetricSetQueryRejectsMissingMetricBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"vitals",
		"metric-set",
		"query",
		"--package",
		"com.example.app",
		"--metric-set",
		"crash-rate",
		"--aggregation",
		"DAILY",
		"--start-date",
		"2026-05-01",
		"--end-date",
		"2026-05-19",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected metric validation error")
	}
	if !strings.Contains(err.Error(), "at least one metric is required") {
		t.Fatalf("error = %v, want metric validation", err)
	}
}

func TestVitalsMetricSetQueryRejectsInvalidStartDateBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"vitals",
		"metric-set",
		"query",
		"--package",
		"com.example.app",
		"--metric-set",
		"crash-rate",
		"--metric",
		"crashRate",
		"--aggregation",
		"DAILY",
		"--start-date",
		"2026/05/01",
		"--end-date",
		"2026-05-19",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected start date validation error")
	}
	if !strings.Contains(err.Error(), "must use YYYY-MM-DD") {
		t.Fatalf("error = %v, want start date validation", err)
	}
}

func TestVitalsMetricSetQueryRejectsInvalidRangeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"vitals",
		"metric-set",
		"query",
		"--package",
		"com.example.app",
		"--metric-set",
		"crash-rate",
		"--metric",
		"crashRate",
		"--aggregation",
		"DAILY",
		"--start-date",
		"2026-05-19",
		"--end-date",
		"2026-05-01",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected date range validation error")
	}
	if !strings.Contains(err.Error(), "start date must be before end date") {
		t.Fatalf("error = %v, want date range validation", err)
	}
}

func TestVitalsMetricSetQueryRejectsUnsupportedAggregationBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"vitals",
		"metric-set",
		"query",
		"--package",
		"com.example.app",
		"--metric-set",
		"error-count",
		"--metric",
		"errorReportCount",
		"--aggregation",
		"HOURLY",
		"--start-date",
		"2026-05-01",
		"--end-date",
		"2026-05-19",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected aggregation validation error")
	}
	if !strings.Contains(err.Error(), "aggregation period HOURLY is not supported") {
		t.Fatalf("error = %v, want aggregation validation", err)
	}
}

func TestVitalsErrorsIssuesSearchRejectsInvalidRangeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"vitals",
		"errors",
		"issues",
		"search",
		"--package",
		"com.example.app",
		"--start-date",
		"2026-05-19",
		"--end-date",
		"2026-05-01",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected date range validation error")
	}
	if !strings.Contains(err.Error(), "start date must be before end date") {
		t.Fatalf("error = %v, want date range validation", err)
	}
}

func TestVitalsErrorsReportsSearchRejectsUnsupportedTimeZoneBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"vitals",
		"errors",
		"reports",
		"search",
		"--package",
		"com.example.app",
		"--start-date",
		"2026-05-01",
		"--end-date",
		"2026-05-19",
		"--time-zone",
		"America/Los_Angeles",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected timezone validation error")
	}
	if !strings.Contains(err.Error(), "only support UTC") {
		t.Fatalf("error = %v, want timezone validation", err)
	}
}

func TestVitalsAnomaliesListRejectsNegativePageSizeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"vitals",
		"anomalies",
		"list",
		"--package",
		"com.example.app",
		"--page-size",
		"-1",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected page size validation error")
	}
	if !strings.Contains(err.Error(), "page size cannot be negative") {
		t.Fatalf("error = %v, want page size validation", err)
	}
}

func TestSubscriptionsGetRejectsInvalidProductIDBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"get",
		"--package",
		"com.example.app",
		"--product-id",
		"Premium",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected product ID validation error")
	}
	if !strings.Contains(err.Error(), "subscription product ID") {
		t.Fatalf("error = %v, want product ID validation", err)
	}
}

func TestSubscriptionsBatchGetRejectsMissingProductIDBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"batch-get",
		"--package",
		"com.example.app",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected product ID validation error")
	}
	if !strings.Contains(err.Error(), "at least one subscription product ID") {
		t.Fatalf("error = %v, want missing product ID validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestSubscriptionsPatchDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"patch",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--listing-language",
		"en-US",
		"--title",
		"Premium",
		"--description",
		"Full access",
		"--benefit",
		"Unlimited projects",
		"--regions-version",
		"2022/02",
		"--latency-tolerance",
		"latencyTolerant",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"productId":"premium"`,
		`"dryRun":true`,
		`"languageCode":"en-US"`,
		`"title":"Premium"`,
		`"updateMask":"listings"`,
		`"regionsVersion":"2022/02"`,
		`"latencyTolerance":"latencyTolerant"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionsPatchRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"patch",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--listing-language",
		"en-US",
		"--title",
		"Premium",
		"--regions-version",
		"2022/02",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestSubscriptionsDeleteDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"delete",
		"--package",
		"com.example.app",
		"--product-id",
		"premium_monthly",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"productId":"premium_monthly"`,
		`"dryRun":true`,
		`"deleted":false`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionsDeleteRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"delete",
		"--package",
		"com.example.app",
		"--product-id",
		"premium_monthly",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestSubscriptionOffersListRejectsInvalidPageSizeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"list",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--page-size",
		"1001",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected page size validation error")
	}
	if !strings.Contains(err.Error(), "page size") {
		t.Fatalf("error = %v, want page size validation", err)
	}
}

func TestSubscriptionOffersGetRejectsMissingOfferIDBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"get",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected offer ID validation error")
	}
	if !strings.Contains(err.Error(), "subscription offer ID") {
		t.Fatalf("error = %v, want offer ID validation", err)
	}
}

func TestSubscriptionOffersDeactivateDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"deactivate",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro",
		"--latency-tolerance",
		"latencyTolerant",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"action":"deactivate"`,
		`"offerId":"intro"`,
		`"latencyTolerance":"latencyTolerant"`,
		`"dryRun":true`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionOffersBatchDeactivateDryRunInfersParentsBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-deactivate",
		"--package",
		"com.example.app",
		"--offer",
		"premium/monthly/intro",
		"--offer",
		"premium/annual/winback",
		"--latency-tolerance",
		"latencyTolerant",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"action":"deactivate"`,
		`"productId":"premium"`,
		`"basePlanId":"-"`,
		`"offerId":"intro"`,
		`"latencyTolerance":"latencyTolerant"`,
		`"dryRun":true`,
		`"applied":false`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionOffersBatchPatchAvailabilityDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-patch-availability",
		"--package",
		"com.example.app",
		"--availability",
		"premium/monthly/intro/us: false",
		"--availability",
		"premium/annual/winback/FR:true",
		"--regions-version",
		"2026/05",
		"--latency-tolerance",
		"latencyTolerant",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"productId":"premium"`,
		`"basePlanId":"-"`,
		`"offerId":"intro"`,
		`"regionCode":"US"`,
		`"availability":false`,
		`"newSubscriberAvailability":false`,
		`"updateMask":"regionalConfigs"`,
		`"regionsVersion":"2026/05"`,
		`"latencyTolerance":"latencyTolerant"`,
		`"dryRun":true`,
		`"applied":false`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionOffersBatchPatchAvailabilityRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-patch-availability",
		"--package",
		"com.example.app",
		"--availability",
		"premium/monthly/intro/US:true",
		"--regions-version",
		"2026/05",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestSubscriptionOffersBatchPatchAvailabilityRejectsMalformedPatchBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-patch-availability",
		"--package",
		"com.example.app",
		"--availability",
		"premium/monthly/intro:true",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected availability format validation error")
	}
	if !strings.Contains(err.Error(), "productId/basePlanId/offerId/REGION:true|false") {
		t.Fatalf("error = %v, want availability format validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestSubscriptionOffersBatchPatchAvailabilityRejectsInvalidBooleanBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-patch-availability",
		"--package",
		"com.example.app",
		"--availability",
		"premium/monthly/intro/US:notabool",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected availability boolean validation error")
	}
	if !strings.Contains(err.Error(), "productId/basePlanId/offerId/REGION:true|false") {
		t.Fatalf("error = %v, want availability format validation", err)
	}
	if strings.Contains(err.Error(), "strconv.ParseBool") {
		t.Fatalf("error = %v, did not expect raw strconv error", err)
	}
}

func TestSubscriptionOffersBatchPatchPhaseRelativeDiscountsDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-patch-phase-relative-discounts",
		"--package",
		"com.example.app",
		"--relative-discount",
		"premium/monthly/intro/0/us:0.75",
		"--relative-discount",
		"premium/annual/winback/1/FR:0.5",
		"--regions-version",
		"2026/05",
		"--latency-tolerance",
		"latencyTolerant",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"productId":"premium"`,
		`"basePlanId":"-"`,
		`"offerId":"intro"`,
		`"phaseIndex":0`,
		`"regionCode":"US"`,
		`"relativeDiscount":0.75`,
		`"updateMask":"phases"`,
		`"regionsVersion":"2026/05"`,
		`"latencyTolerance":"latencyTolerant"`,
		`"dryRun":true`,
		`"applied":false`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionOffersBatchPatchPhaseRelativeDiscountsRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-patch-phase-relative-discounts",
		"--package",
		"com.example.app",
		"--relative-discount",
		"premium/monthly/intro/0/US:0.75",
		"--regions-version",
		"2026/05",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestSubscriptionOffersBatchPatchPhaseRelativeDiscountsRejectsMalformedPatchBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-patch-phase-relative-discounts",
		"--package",
		"com.example.app",
		"--relative-discount",
		"premium/monthly/intro/US:0.75",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected phase relative discount format validation error")
	}
	if !strings.Contains(err.Error(), "productId/basePlanId/offerId/phaseIndex/REGION:0.75") {
		t.Fatalf("error = %v, want phase relative discount format validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestSubscriptionOffersBatchPatchPhaseAbsoluteDiscountsDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-patch-phase-absolute-discounts",
		"--package",
		"com.example.app",
		"--absolute-discount",
		"premium/monthly/intro/0/us:USD:1:500000000",
		"--absolute-discount",
		"premium/annual/winback/1/FR:EUR:2",
		"--regions-version",
		"2026/05",
		"--latency-tolerance",
		"latencyTolerant",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"productId":"premium"`,
		`"basePlanId":"-"`,
		`"offerId":"intro"`,
		`"phaseIndex":0`,
		`"regionCode":"US"`,
		`"currencyCode":"USD"`,
		`"units":1`,
		`"nanos":500000000`,
		`"updateMask":"phases"`,
		`"regionsVersion":"2026/05"`,
		`"latencyTolerance":"latencyTolerant"`,
		`"dryRun":true`,
		`"applied":false`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionOffersCreateDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	bodyPath := filepath.Join(t.TempDir(), "offer.json")
	if err := os.WriteFile(bodyPath, []byte(`{
		"state":"ACTIVE",
		"offerTags":[{"tag":"intro"}],
		"regionalConfigs":[
			{"regionCode":"US","newSubscriberAvailability":true},
			{"regionCode":"FR","newSubscriberAvailability":true}
		],
		"otherRegionsConfig":{"otherRegionsNewSubscriberAvailability":true},
		"phases":[{
			"duration":"P1M",
			"recurrenceCount":1,
			"regionalConfigs":[
				{"regionCode":"US","price":{"currencyCode":"USD","units":"1"}},
				{"regionCode":"FR","price":{"currencyCode":"EUR","nanos":990000000}}
			],
			"otherRegionsConfig":{"free":{}}
		}],
		"targeting":{"acquisitionRule":{"scope":{"thisSubscription":{}}}}
	}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro",
		"--from-json",
		bodyPath,
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"action":"create"`,
		`"dryRun":true`,
		`"created":false`,
		`"productId":"premium"`,
		`"basePlanId":"monthly"`,
		`"offerId":"intro"`,
		`"regionsVersion":"2026/05"`,
		`"currencyCode":"USD"`,
		`"nanos":990000000`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
	if strings.Contains(output, `"state":"ACTIVE"`) {
		t.Fatalf("output = %s, did not expect output-only state from input JSON", output)
	}
}

func TestSubscriptionOffersCreateBasicFreePhaseDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro",
		"--offer-tag",
		"trial",
		"--free-region",
		"us",
		"--free-region",
		"FR",
		"--phase-duration",
		"P7D",
		"--phase-recurrence",
		"1",
		"--regions-version",
		"2026/05",
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
		`"created":false`,
		`"productId":"premium"`,
		`"basePlanId":"monthly"`,
		`"offerId":"intro"`,
		`"offerTags":["trial"]`,
		`"regionCode":"US"`,
		`"regionCode":"FR"`,
		`"newSubscriberAvailability":true`,
		`"duration":"P7D"`,
		`"recurrenceCount":1`,
		`"free":true`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionOffersCreateBasicPricePhaseDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro",
		"--offer-tag",
		"paid-intro",
		"--price",
		"us:USD:1:990000000",
		"--price",
		"FR:EUR:0:990000000",
		"--phase-duration",
		"P1M",
		"--phase-recurrence",
		"1",
		"--regions-version",
		"2026/05",
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
		`"created":false`,
		`"productId":"premium"`,
		`"basePlanId":"monthly"`,
		`"offerId":"intro"`,
		`"offerTags":["paid-intro"]`,
		`"regionCode":"US"`,
		`"regionCode":"FR"`,
		`"newSubscriberAvailability":true`,
		`"duration":"P1M"`,
		`"recurrenceCount":1`,
		`"price":{"currencyCode":"USD","units":1,"nanos":990000000}`,
		`"price":{"currencyCode":"EUR","nanos":990000000}`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionOffersCreateBasicRelativeDiscountPhaseDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro",
		"--offer-tag",
		"half-off",
		"--relative-discount",
		"us:0.5",
		"--relative-discount",
		"FR:0.25",
		"--phase-duration",
		"P1M",
		"--phase-recurrence",
		"1",
		"--regions-version",
		"2026/05",
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
		`"created":false`,
		`"productId":"premium"`,
		`"basePlanId":"monthly"`,
		`"offerId":"intro"`,
		`"offerTags":["half-off"]`,
		`"regionCode":"US"`,
		`"regionCode":"FR"`,
		`"newSubscriberAvailability":true`,
		`"duration":"P1M"`,
		`"recurrenceCount":1`,
		`"relativeDiscount":0.5`,
		`"relativeDiscount":0.25`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionOffersCreateBasicFlagsRejectDuplicatePhaseRegionBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro",
		"--free-region",
		"US",
		"--price",
		"us:USD:1",
		"--phase-duration",
		"P1M",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected duplicate region validation error")
	}
	if !strings.Contains(err.Error(), "subscription offer create region US is duplicated") {
		t.Fatalf("error = %v, want duplicate region validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestSubscriptionOffersCreateBasicFlagsRejectInvalidRelativeDiscountBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro",
		"--relative-discount",
		"US:not-a-number",
		"--phase-duration",
		"P1M",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected relative discount validation error")
	}
	if !strings.Contains(err.Error(), "subscription offer create relative discount must use REGION:0.5") {
		t.Fatalf("error = %v, want relative discount format validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestSubscriptionOffersCreateRejectsJSONWithBasicFlagsBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	bodyPath := filepath.Join(t.TempDir(), "offer.json")
	if err := os.WriteFile(bodyPath, []byte(`{
		"regionalConfigs":[{"regionCode":"US","newSubscriberAvailability":true}],
		"phases":[{"duration":"P7D","recurrenceCount":1,"regionalConfigs":[{"regionCode":"US","free":{}}]}]
	}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro",
		"--from-json",
		bodyPath,
		"--free-region",
		"US",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected from-json and basic flags validation error")
	}
	if !strings.Contains(err.Error(), "cannot be combined") {
		t.Fatalf("error = %v, want combination validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestSubscriptionOffersCreateRejectsInvalidBodyBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	bodyPath := filepath.Join(t.TempDir(), "offer.json")
	if err := os.WriteFile(bodyPath, []byte(`{"regionalConfigs":[{"regionCode":"US","newSubscriberAvailability":true}]}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro",
		"--from-json",
		bodyPath,
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected offer body validation error")
	}
	if !strings.Contains(err.Error(), "requires one or two phases") {
		t.Fatalf("error = %v, want phase validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestSubscriptionOffersBatchPatchPhaseAbsoluteDiscountsRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-patch-phase-absolute-discounts",
		"--package",
		"com.example.app",
		"--absolute-discount",
		"premium/monthly/intro/0/US:USD:1",
		"--regions-version",
		"2026/05",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestSubscriptionOffersBatchPatchPhaseAbsoluteDiscountsRejectsMalformedPatchBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-patch-phase-absolute-discounts",
		"--package",
		"com.example.app",
		"--absolute-discount",
		"premium/monthly/intro/US:USD:1",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected phase absolute discount format validation error")
	}
	if !strings.Contains(err.Error(), "productId/basePlanId/offerId/phaseIndex/REGION:CURRENCY:UNITS[:NANOS]") {
		t.Fatalf("error = %v, want phase absolute discount format validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestSubscriptionOffersBatchPatchPhasePricesDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-patch-phase-prices",
		"--package",
		"com.example.app",
		"--price",
		"premium/monthly/intro/0/us:USD:1:990000000",
		"--price",
		"premium/annual/winback/1/FR:EUR:2",
		"--regions-version",
		"2026/05",
		"--latency-tolerance",
		"latencyTolerant",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"productId":"premium"`,
		`"basePlanId":"-"`,
		`"offerId":"intro"`,
		`"phaseIndex":0`,
		`"regionCode":"US"`,
		`"currencyCode":"USD"`,
		`"units":1`,
		`"nanos":990000000`,
		`"updateMask":"phases"`,
		`"regionsVersion":"2026/05"`,
		`"latencyTolerance":"latencyTolerant"`,
		`"dryRun":true`,
		`"applied":false`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionOffersBatchPatchPhasePricesRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-patch-phase-prices",
		"--package",
		"com.example.app",
		"--price",
		"premium/monthly/intro/0/US:USD:1",
		"--regions-version",
		"2026/05",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestSubscriptionOffersBatchPatchPhasePricesRejectsMalformedPatchBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-patch-phase-prices",
		"--package",
		"com.example.app",
		"--price",
		"premium/monthly/intro/US:USD:1",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected phase price format validation error")
	}
	if !strings.Contains(err.Error(), "productId/basePlanId/offerId/phaseIndex/REGION:CURRENCY:UNITS[:NANOS]") {
		t.Fatalf("error = %v, want phase price format validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestSubscriptionOffersBatchPatchPhaseFreeDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-patch-phase-free",
		"--package",
		"com.example.app",
		"--free",
		"premium/monthly/intro/0/us",
		"--free",
		"premium/annual/winback/1/FR",
		"--regions-version",
		"2026/05",
		"--latency-tolerance",
		"latencyTolerant",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"productId":"premium"`,
		`"basePlanId":"-"`,
		`"offerId":"intro"`,
		`"phaseIndex":0`,
		`"regionCode":"US"`,
		`"free":true`,
		`"updateMask":"phases"`,
		`"regionsVersion":"2026/05"`,
		`"latencyTolerance":"latencyTolerant"`,
		`"dryRun":true`,
		`"applied":false`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionOffersBatchPatchPhaseFreeRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-patch-phase-free",
		"--package",
		"com.example.app",
		"--free",
		"premium/monthly/intro/0/US",
		"--regions-version",
		"2026/05",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestSubscriptionOffersBatchPatchPhaseFreeRejectsMalformedPatchBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-patch-phase-free",
		"--package",
		"com.example.app",
		"--free",
		"premium/monthly/intro/US",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected phase free format validation error")
	}
	if !strings.Contains(err.Error(), "productId/basePlanId/offerId/phaseIndex/REGION") {
		t.Fatalf("error = %v, want phase free format validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestSubscriptionOffersBatchActivateRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-activate",
		"--package",
		"com.example.app",
		"--offer",
		"premium/monthly/intro",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestSubscriptionOffersBatchActivateRejectsMissingOfferBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-activate",
		"--package",
		"com.example.app",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected missing offer validation error")
	}
	if !strings.Contains(err.Error(), "at least one subscription offer is required") {
		t.Fatalf("error = %v, want missing offer validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestSubscriptionOffersDeleteDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"delete",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"productId":"premium"`,
		`"basePlanId":"monthly"`,
		`"offerId":"intro"`,
		`"dryRun":true`,
		`"deleted":false`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionOffersDeleteRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"delete",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestSubscriptionOffersRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"activate",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestSubscriptionOffersListAcceptsWildcardsBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"list",
		"--package",
		"com.example.app",
		"--product-id",
		"-",
		"--base-plan-id",
		"-",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected auth error after wildcard validation succeeds")
	}
	if strings.Contains(err.Error(), "invalid subscription product ID") || strings.Contains(err.Error(), "base plan") {
		t.Fatalf("error = %v, want auth error after wildcard validation", err)
	}
}

func TestSubscriptionOffersGetRejectsWildcardBasePlanBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"get",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"-",
		"--offer-id",
		"intro",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected base plan validation error")
	}
	if !strings.Contains(err.Error(), "subscription base plan ID") {
		t.Fatalf("error = %v, want base plan validation", err)
	}
}

func TestSubscriptionOffersBatchGetRejectsMissingOfferBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-get",
		"--package",
		"com.example.app",
		"--product-id",
		"-",
		"--base-plan-id",
		"-",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected offer validation error")
	}
	if !strings.Contains(err.Error(), "at least one subscription offer") {
		t.Fatalf("error = %v, want missing offer validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestSubscriptionOffersBatchGetRejectsParentMismatchBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-get",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer",
		"other/monthly/intro",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected parent mismatch validation error")
	}
	if !strings.Contains(err.Error(), "does not match parent product ID") {
		t.Fatalf("error = %v, want parent product validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestSubscriptionOffersBatchGetRejectsInvalidOfferIDBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-get",
		"--package",
		"com.example.app",
		"--product-id",
		"-",
		"--base-plan-id",
		"-",
		"--offer",
		"premium/monthly/Intro",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected offer ID validation error")
	}
	if !strings.Contains(err.Error(), "subscription offer ID") {
		t.Fatalf("error = %v, want offer ID validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestSubscriptionOffersBatchGetRejectsOverlongOfferIDBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-get",
		"--package",
		"com.example.app",
		"--product-id",
		"-",
		"--base-plan-id",
		"-",
		"--offer",
		"premium/monthly/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected offer ID validation error")
	}
	if !strings.Contains(err.Error(), "cannot exceed 63 characters") {
		t.Fatalf("error = %v, want offer ID length validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestPurchasesProductRejectsMissingTokenBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"product",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected token validation error")
	}
	if !strings.Contains(err.Error(), "purchase token") {
		t.Fatalf("error = %v, want token validation", err)
	}
}

func TestPurchasesProductAllowsTokenOnlyBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"product",
		"--package",
		"com.example.app",
		"--token",
		"token-123",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected auth error")
	}
	if strings.Contains(err.Error(), "product ID") || strings.Contains(err.Error(), "in-app product") {
		t.Fatalf("error = %v, want auth error after token-only validation", err)
	}
}

func TestPurchasesProductAcknowledgeDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"product",
		"acknowledge",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--token",
		"token-123",
		"--developer-payload",
		"order-7",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"action":"acknowledge"`, `"dryRun":true`, `"applied":false`, `"developerPayload":"order-7"`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestPurchasesProductConsumeRejectsMissingConfirmBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"product",
		"consume",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--token",
		"token-123",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm validation error")
	}
	if !strings.Contains(err.Error(), "--confirm") {
		t.Fatalf("error = %v, want confirm validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestPurchasesVoidedListRejectsNegativeMaxResultsBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"voided",
		"list",
		"--package",
		"com.example.app",
		"--max-results",
		"-1",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected max results validation error")
	}
	if !strings.Contains(err.Error(), "max results") {
		t.Fatalf("error = %v, want max results validation", err)
	}
}

func TestPurchasesVoidedListRejectsInvalidTypeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"voided",
		"list",
		"--package",
		"com.example.app",
		"--type",
		"2",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected type validation error")
	}
	if !strings.Contains(err.Error(), "voided purchase type") {
		t.Fatalf("error = %v, want type validation", err)
	}
}

func TestPurchasesVoidedListRejectsTokenWithTimeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"voided",
		"list",
		"--package",
		"com.example.app",
		"--token",
		"page",
		"--start-time",
		"1700000000000",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected token/time validation error")
	}
	if !strings.Contains(err.Error(), "pagination token") {
		t.Fatalf("error = %v, want token/time validation", err)
	}
}

func TestPurchasesVoidedListRejectsFutureEndTimeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"voided",
		"list",
		"--package",
		"com.example.app",
		"--end-time",
		"4102444800000",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected future end time validation error")
	}
	if !strings.Contains(err.Error(), "future") {
		t.Fatalf("error = %v, want future end time validation", err)
	}
}

func TestPurchasesSubscriptionRejectsMissingTokenBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"subscription",
		"--package",
		"com.example.app",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected token validation error")
	}
	if !strings.Contains(err.Error(), "purchase token") {
		t.Fatalf("error = %v, want token validation", err)
	}
}

func TestPurchasesSubscriptionRevokeDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"subscription",
		"revoke",
		"--package",
		"com.example.app",
		"--token",
		"token-123",
		"--refund",
		"full",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"refundType":"full"`, `"dryRun":true`, `"applied":false`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestPurchasesSubscriptionRevokeItemRefundDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"subscription",
		"revoke",
		"--package",
		"com.example.app",
		"--token",
		"token-123",
		"--refund",
		"item",
		"--refund-product-id",
		"premium_addon",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"refundType":"item"`, `"refundProductId":"premium_addon"`, `"dryRun":true`, `"applied":false`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestPurchasesSubscriptionRevokeRejectsMissingRefundBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"subscription",
		"revoke",
		"--package",
		"com.example.app",
		"--token",
		"token-123",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected refund validation error")
	}
	if !strings.Contains(err.Error(), "refund type") {
		t.Fatalf("error = %v, want refund type validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestPurchasesSubscriptionAcknowledgeDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"subscription",
		"acknowledge",
		"--package",
		"com.example.app",
		"--subscription-id",
		"premium_monthly",
		"--token",
		"token-123",
		"--developer-payload",
		"handled-by-gpc",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"action":"acknowledge"`, `"subscriptionId":"premium_monthly"`, `"developerPayload":"handled-by-gpc"`, `"dryRun":true`, `"applied":false`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestPurchasesSubscriptionCancelRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"subscription",
		"cancel",
		"--package",
		"com.example.app",
		"--token",
		"token-123",
		"--cancellation-type",
		"userRequestedStopRenewals",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirmation gate")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestPurchasesSubscriptionCancelDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"subscription",
		"cancel",
		"--package",
		"com.example.app",
		"--token",
		"token-123",
		"--cancellation-type",
		"userRequestedStopRenewals",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"action":"cancel"`, `"cancellationType":"userRequestedStopRenewals"`, `"dryRun":true`, `"applied":false`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "subscriptionId") {
		t.Fatalf("output = %s, did not expect legacy subscription ID for v2 cancel", output)
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestUsersDeleteDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"users",
		"delete",
		"--developer",
		"1234567890",
		"--user-email",
		"user@example.com",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"name":"developers/1234567890/users/user@example.com"`, `"dryRun":true`, `"deleted":false`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestUsersCreateDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"users",
		"create",
		"--developer",
		"1234567890",
		"--user-email",
		"user@example.com",
		"--permission",
		"CAN_VIEW_NON_FINANCIAL_DATA_GLOBAL",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"action":"create"`, `"dryRun":true`, `"desiredUser"`, `"CAN_VIEW_NON_FINANCIAL_DATA_GLOBAL"`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestUsersCreateRejectsMissingConfirmBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"users",
		"create",
		"--developer",
		"1234567890",
		"--user-email",
		"user@example.com",
		"--permission",
		"CAN_VIEW_NON_FINANCIAL_DATA_GLOBAL",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm validation error")
	}
	if !strings.Contains(err.Error(), "--confirm") {
		t.Fatalf("error = %v, want confirm validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestUsersPatchDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"users",
		"patch",
		"--name",
		"developers/1234567890/users/user@example.com",
		"--permission",
		"CAN_REPLY_TO_REVIEWS_GLOBAL",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"action":"patch"`, `"dryRun":true`, `"desiredUser"`, `"CAN_REPLY_TO_REVIEWS_GLOBAL"`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestUsersPatchRejectsEmptyPatchBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"users",
		"patch",
		"--developer",
		"1234567890",
		"--user-email",
		"user@example.com",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected empty patch validation error")
	}
	if !strings.Contains(err.Error(), "--permission") {
		t.Fatalf("error = %v, want patch field validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestUsersDeleteRejectsMissingConfirmBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"users",
		"delete",
		"--name",
		"developers/1234567890/users/user@example.com",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm validation error")
	}
	if !strings.Contains(err.Error(), "--confirm") {
		t.Fatalf("error = %v, want confirm validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOrdersGetRejectsMissingOrderIDBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"orders",
		"get",
		"--package",
		"com.example.app",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected order ID validation error")
	}
	if !strings.Contains(err.Error(), "order ID") {
		t.Fatalf("error = %v, want order ID validation", err)
	}
}

func TestOrdersBatchGetRejectsDuplicateOrderIDBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"orders",
		"batch-get",
		"--package",
		"com.example.app",
		"--order-id",
		"GPA.123",
		"--order-id",
		"GPA.123",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected duplicate order ID validation error")
	}
	if !strings.Contains(err.Error(), "duplicated") {
		t.Fatalf("error = %v, want duplicate order ID validation", err)
	}
}

func TestOrdersRefundDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"orders",
		"refund",
		"--package",
		"com.example.app",
		"--order-id",
		"GPA.123",
		"--revoke",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	var result struct {
		OrderID string `json:"orderId"`
		Revoke  bool   `json:"revoke"`
		DryRun  bool   `json:"dryRun"`
		Applied bool   `json:"applied"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Unmarshal() error = %v; output = %s", err, buf.String())
	}
	if result.OrderID != "GPA.123" || !result.Revoke || !result.DryRun || result.Applied {
		t.Fatalf("result = %#v, want revoked refund dry-run", result)
	}
	if strings.Contains(buf.String(), "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", buf.String())
	}
}

func TestOrdersRefundRequiresConfirmBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"orders",
		"refund",
		"--package",
		"com.example.app",
		"--order-id",
		"GPA.123",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirmation validation error")
	}
	if !strings.Contains(err.Error(), "--confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOrdersRefundRejectsConfirmDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"orders",
		"refund",
		"--package",
		"com.example.app",
		"--order-id",
		"GPA.123",
		"--confirm",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected conflicting flag validation error")
	}
	if !strings.Contains(err.Error(), "cannot be used together") {
		t.Fatalf("error = %v, want conflicting flag validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestPricingConvertRejectsInvalidPriceBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"pricing",
		"convert-region-prices",
		"--package",
		"com.example.app",
		"--currency",
		"USD",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected price validation error")
	}
	if !strings.Contains(err.Error(), "price must be greater than 0") {
		t.Fatalf("error = %v, want price validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func writeRootTestFile(t *testing.T, name string) string {
	t.Helper()
	path := t.TempDir() + "/" + name
	if err := os.WriteFile(path, []byte("artifact"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}

func writeRootTestContent(t *testing.T, name string, content string) string {
	t.Helper()
	path := t.TempDir() + "/" + name
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}

func writeNestedRootTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}
