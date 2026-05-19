package cmd

import (
	"bytes"
	"strings"
	"testing"
)

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
