package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestAppRecoveryListRejectsMissingVersionCodeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"app-recovery",
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
}

func TestAppRecoveryDeployDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"app-recovery",
		"deploy",
		"--package",
		"com.example.app",
		"--id",
		"7",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"action":"deploy"`, `"dryRun":true`, `"applied":false`, `"appRecoveryId":"7"`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestAppRecoveryCreateDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"app-recovery",
		"create",
		"--package",
		"com.example.app",
		"--version-code",
		"42",
		"--region",
		"US",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"dryRun":true`, `"created":false`, `"versionCodes":[42]`, `"regionCodes":["US"]`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestAppRecoveryCreateRejectsMissingTargetingBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"app-recovery",
		"create",
		"--package",
		"com.example.app",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected targeting validation error")
	}
	if !strings.Contains(err.Error(), "targeting") {
		t.Fatalf("error = %v, want targeting validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestAppRecoveryAddTargetingDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"app-recovery",
		"add-targeting",
		"--package",
		"com.example.app",
		"--id",
		"7",
		"--sdk-level",
		"26",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"dryRun":true`, `"applied":false`, `"appRecoveryId":"7"`, `"sdkLevels":[26]`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestAppRecoveryAddTargetingRejectsMissingTargetBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"app-recovery",
		"add-targeting",
		"--package",
		"com.example.app",
		"--id",
		"7",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected targeting validation error")
	}
	if !strings.Contains(err.Error(), "targeting") {
		t.Fatalf("error = %v, want targeting validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestAppRecoveryCancelRejectsMissingConfirmBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"app-recovery",
		"cancel",
		"--package",
		"com.example.app",
		"--id",
		"7",
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
