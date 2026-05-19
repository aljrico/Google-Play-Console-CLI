package cmd

import (
	"bytes"
	"strings"
	"testing"
)

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
