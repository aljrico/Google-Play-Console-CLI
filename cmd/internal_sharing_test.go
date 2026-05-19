package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestInternalSharingUploadDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

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
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

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
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")
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
