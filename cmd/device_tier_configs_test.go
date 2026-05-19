package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestDeviceTierConfigsListRejectsInvalidPageSizeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"device-tier-configs",
		"list",
		"--package",
		"com.example.app",
		"--page-size",
		"101",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected page size validation error")
	}
	if !strings.Contains(err.Error(), "page size") {
		t.Fatalf("error = %v, want page size validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestDeviceTierConfigsGetRejectsMissingIDBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"device-tier-configs",
		"get",
		"--package",
		"com.example.app",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected device tier config ID validation error")
	}
	if !strings.Contains(err.Error(), "device tier config ID") {
		t.Fatalf("error = %v, want device tier config ID validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}
