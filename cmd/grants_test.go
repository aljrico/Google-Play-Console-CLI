package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestGrantsCreateDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"grants",
		"create",
		"--developer",
		"1234567890",
		"--user-email",
		"user@example.com",
		"--package",
		"com.example.app",
		"--permission",
		"CAN_VIEW_NON_FINANCIAL_DATA",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"dryRun":true`) {
		t.Fatalf("output = %s, want dryRun true", buf.String())
	}
	if !strings.Contains(buf.String(), `"target":"developers/1234567890/users/user@example.com/grants/com.example.app"`) {
		t.Fatalf("output = %s, want full grant target", buf.String())
	}
	if !strings.Contains(buf.String(), `"appLevelPermissions":["CAN_VIEW_NON_FINANCIAL_DATA"]`) {
		t.Fatalf("output = %s, want grant permission preview", buf.String())
	}
	if strings.Contains(buf.String(), "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", buf.String())
	}
}

func TestGrantsPatchRequiresConfirmBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"grants",
		"patch",
		"--name",
		"developers/123/users/user@example.com/grants/com.example.app",
		"--permission",
		"CAN_REPLY_TO_REVIEWS",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirmation validation error")
	}
	if !strings.Contains(err.Error(), "--confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}
