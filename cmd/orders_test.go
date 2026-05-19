package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestOrdersGetRejectsMissingOrderIDBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"orders",
		"get",
		"--package",
		"com.example.app",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected order ID validation error")
	}
	if !strings.Contains(err.Error(), "order ID") {
		t.Fatalf("error = %v, want order ID validation", err)
	}
}

func TestOrdersBatchGetRejectsDuplicateOrderIDBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"orders",
		"batch-get",
		"--package",
		"com.example.app",
		"--order-id",
		"GPA.123",
		"--order-id",
		"GPA.123",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected duplicate order ID validation error")
	}
	if !strings.Contains(err.Error(), "duplicated") {
		t.Fatalf("error = %v, want duplicate order ID validation", err)
	}
}

func TestOrdersRefundDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"orders",
		"refund",
		"--package",
		"com.example.app",
		"--order-id",
		"GPA.123",
		"--revoke",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	var result struct {
		OrderID string `json:"orderId"`
		Revoke  bool   `json:"revoke"`
		DryRun  bool   `json:"dryRun"`
		Applied bool   `json:"applied"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Unmarshal() error = %v; output = %s", err, buf.String())
	}
	if result.OrderID != "GPA.123" || !result.Revoke || !result.DryRun || result.Applied {
		t.Fatalf("result = %#v, want revoked refund dry-run", result)
	}
	if strings.Contains(buf.String(), "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", buf.String())
	}
}

func TestOrdersRefundRequiresConfirmBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"orders",
		"refund",
		"--package",
		"com.example.app",
		"--order-id",
		"GPA.123",
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

func TestOrdersRefundRejectsConfirmDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"orders",
		"refund",
		"--package",
		"com.example.app",
		"--order-id",
		"GPA.123",
		"--confirm",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected conflicting flag validation error")
	}
	if !strings.Contains(err.Error(), "cannot be used together") {
		t.Fatalf("error = %v, want conflicting flag validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}
