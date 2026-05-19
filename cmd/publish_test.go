package cmd

import (
	"bytes"
	"strings"
	"testing"
)

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
