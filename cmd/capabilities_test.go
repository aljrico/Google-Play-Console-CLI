package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestCapabilitiesOutputsParityMatrixWithoutAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"capabilities",
		"--status",
		"tested",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, `"status":"tested"`) {
		t.Fatalf("output = %s, want tested status", output)
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestCapabilitiesRejectsUnsupportedStatus(t *testing.T) {
	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"capabilities",
		"--status",
		"done-ish",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected unsupported status error")
	}
	if !strings.Contains(err.Error(), "unsupported capability status") {
		t.Fatalf("error = %v, want status validation", err)
	}
}
