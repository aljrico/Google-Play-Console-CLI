package cmd

import (
	"bytes"
	"strings"
	"testing"
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

func TestVersionRejectsUnexpectedArgs(t *testing.T) {
	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{"version", "stray"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error")
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
	if !strings.Contains(buf.String(), `"track":"beta"`) {
		t.Fatalf("release upload dry-run output = %s", buf.String())
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
