package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/aljrico/Google-Play-Console-CLI/internal/config"
)

func TestAccountStatusReportsMissingProfileWithoutAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

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
	t.Setenv("PLAYPUB_CONFIG", configPath)
	if err := os.WriteFile(serviceAccountPath, []byte(`{
  "type": "service_account",
  "project_id": "play-project",
  "private_key": "-----BEGIN PRIVATE KEY-----\nabc\n-----END PRIVATE KEY-----\n",
  "client_email": "playpub@example.iam.gserviceaccount.com"
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
		`"serviceAccountEmail":"playpub@example.iam.gserviceaccount.com"`,
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
	t.Setenv("PLAYPUB_CONFIG", configPath)
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
	t.Setenv("PLAYPUB_CONFIG", configPath)
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
