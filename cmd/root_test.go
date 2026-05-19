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
