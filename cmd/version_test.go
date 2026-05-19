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

func TestVersionRejectsUnexpectedArgs(t *testing.T) {
	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{"version", "stray"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error")
	}
}
