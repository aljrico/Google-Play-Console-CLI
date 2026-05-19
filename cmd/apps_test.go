package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestAppsListReportsBlockedSurfaceWithoutAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"apps",
		"list",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected unsupported apps list error")
	}
	if !strings.Contains(err.Error(), "limited app discovery APIs") {
		t.Fatalf("error = %v, want app discovery limitation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}
