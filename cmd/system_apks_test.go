package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestSystemAPKVariantsListRejectsMissingVersionCodeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"system-apks",
		"variants",
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
