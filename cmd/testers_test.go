package cmd

import (
	"bytes"
	"strings"
	"testing"
)

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
