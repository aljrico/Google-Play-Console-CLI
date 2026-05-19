package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestGeneratedAPKsListRejectsMissingVersionCodeBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

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
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")
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
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

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
