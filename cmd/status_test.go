package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestStatusRejectsInvalidPackageBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"status",
		"--package",
		"bad",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected package validation error")
	}
	if !strings.Contains(err.Error(), "invalid Android package name") {
		t.Fatalf("error = %v, want package validation", err)
	}
}
