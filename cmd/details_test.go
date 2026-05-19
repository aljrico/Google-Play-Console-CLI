package cmd

import (
	"bytes"
	"strings"
	"testing"
)

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
