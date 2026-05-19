package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestUsersListRejectsMissingDeveloperBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"users",
		"list",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected developer validation error")
	}
	if !strings.Contains(err.Error(), "developer account") {
		t.Fatalf("error = %v, want developer validation", err)
	}
}

func TestUsersListRejectsInvalidPageSizeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"users",
		"list",
		"--developer",
		"1234567890",
		"--page-size",
		"-2",
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
}

func TestUsersDeleteDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"users",
		"delete",
		"--developer",
		"1234567890",
		"--user-email",
		"user@example.com",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"name":"developers/1234567890/users/user@example.com"`, `"dryRun":true`, `"deleted":false`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestUsersCreateDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"users",
		"create",
		"--developer",
		"1234567890",
		"--user-email",
		"user@example.com",
		"--permission",
		"CAN_VIEW_NON_FINANCIAL_DATA_GLOBAL",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"action":"create"`, `"dryRun":true`, `"desiredUser"`, `"CAN_VIEW_NON_FINANCIAL_DATA_GLOBAL"`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestUsersCreateRejectsMissingConfirmBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"users",
		"create",
		"--developer",
		"1234567890",
		"--user-email",
		"user@example.com",
		"--permission",
		"CAN_VIEW_NON_FINANCIAL_DATA_GLOBAL",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm validation error")
	}
	if !strings.Contains(err.Error(), "--confirm") {
		t.Fatalf("error = %v, want confirm validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestUsersPatchDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"users",
		"patch",
		"--name",
		"developers/1234567890/users/user@example.com",
		"--permission",
		"CAN_REPLY_TO_REVIEWS_GLOBAL",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"action":"patch"`, `"dryRun":true`, `"desiredUser"`, `"CAN_REPLY_TO_REVIEWS_GLOBAL"`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestUsersPatchRejectsEmptyPatchBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"users",
		"patch",
		"--developer",
		"1234567890",
		"--user-email",
		"user@example.com",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected empty patch validation error")
	}
	if !strings.Contains(err.Error(), "--permission") {
		t.Fatalf("error = %v, want patch field validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestUsersDeleteRejectsMissingConfirmBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"users",
		"delete",
		"--name",
		"developers/1234567890/users/user@example.com",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm validation error")
	}
	if !strings.Contains(err.Error(), "--confirm") {
		t.Fatalf("error = %v, want confirm validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}
