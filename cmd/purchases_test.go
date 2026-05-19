package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestPurchasesProductRejectsMissingTokenBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"product",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected token validation error")
	}
	if !strings.Contains(err.Error(), "purchase token") {
		t.Fatalf("error = %v, want token validation", err)
	}
}

func TestPurchasesProductAllowsTokenOnlyBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"product",
		"--package",
		"com.example.app",
		"--token",
		"token-123",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected auth error")
	}
	if strings.Contains(err.Error(), "product ID") || strings.Contains(err.Error(), "in-app product") {
		t.Fatalf("error = %v, want auth error after token-only validation", err)
	}
}

func TestPurchasesProductAcknowledgeDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"product",
		"acknowledge",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--token",
		"token-123",
		"--developer-payload",
		"order-7",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"action":"acknowledge"`, `"dryRun":true`, `"applied":false`, `"developerPayload":"order-7"`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestPurchasesProductConsumeRejectsMissingConfirmBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"product",
		"consume",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--token",
		"token-123",
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

func TestPurchasesVoidedListRejectsNegativeMaxResultsBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"voided",
		"list",
		"--package",
		"com.example.app",
		"--max-results",
		"-1",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected max results validation error")
	}
	if !strings.Contains(err.Error(), "max results") {
		t.Fatalf("error = %v, want max results validation", err)
	}
}

func TestPurchasesVoidedListRejectsInvalidTypeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"voided",
		"list",
		"--package",
		"com.example.app",
		"--type",
		"2",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected type validation error")
	}
	if !strings.Contains(err.Error(), "voided purchase type") {
		t.Fatalf("error = %v, want type validation", err)
	}
}

func TestPurchasesVoidedListRejectsTokenWithTimeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"voided",
		"list",
		"--package",
		"com.example.app",
		"--token",
		"page",
		"--start-time",
		"1700000000000",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected token/time validation error")
	}
	if !strings.Contains(err.Error(), "pagination token") {
		t.Fatalf("error = %v, want token/time validation", err)
	}
}

func TestPurchasesVoidedListRejectsFutureEndTimeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"voided",
		"list",
		"--package",
		"com.example.app",
		"--end-time",
		"4102444800000",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected future end time validation error")
	}
	if !strings.Contains(err.Error(), "future") {
		t.Fatalf("error = %v, want future end time validation", err)
	}
}

func TestPurchasesSubscriptionRejectsMissingTokenBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"subscription",
		"--package",
		"com.example.app",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected token validation error")
	}
	if !strings.Contains(err.Error(), "purchase token") {
		t.Fatalf("error = %v, want token validation", err)
	}
}

func TestPurchasesSubscriptionRevokeDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"subscription",
		"revoke",
		"--package",
		"com.example.app",
		"--token",
		"token-123",
		"--refund",
		"full",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"refundType":"full"`, `"dryRun":true`, `"applied":false`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestPurchasesSubscriptionRevokeItemRefundDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"subscription",
		"revoke",
		"--package",
		"com.example.app",
		"--token",
		"token-123",
		"--refund",
		"item",
		"--refund-product-id",
		"premium_addon",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"refundType":"item"`, `"refundProductId":"premium_addon"`, `"dryRun":true`, `"applied":false`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestPurchasesSubscriptionRevokeRejectsMissingRefundBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"subscription",
		"revoke",
		"--package",
		"com.example.app",
		"--token",
		"token-123",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected refund validation error")
	}
	if !strings.Contains(err.Error(), "refund type") {
		t.Fatalf("error = %v, want refund type validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestPurchasesSubscriptionAcknowledgeDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"subscription",
		"acknowledge",
		"--package",
		"com.example.app",
		"--subscription-id",
		"premium_monthly",
		"--token",
		"token-123",
		"--developer-payload",
		"handled-by-gpc",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"action":"acknowledge"`, `"subscriptionId":"premium_monthly"`, `"developerPayload":"handled-by-gpc"`, `"dryRun":true`, `"applied":false`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestPurchasesSubscriptionCancelRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"subscription",
		"cancel",
		"--package",
		"com.example.app",
		"--token",
		"token-123",
		"--cancellation-type",
		"userRequestedStopRenewals",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirmation gate")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestPurchasesSubscriptionCancelDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"subscription",
		"cancel",
		"--package",
		"com.example.app",
		"--token",
		"token-123",
		"--cancellation-type",
		"userRequestedStopRenewals",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"action":"cancel"`, `"cancellationType":"userRequestedStopRenewals"`, `"dryRun":true`, `"applied":false`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "subscriptionId") {
		t.Fatalf("output = %s, did not expect legacy subscription ID for v2 cancel", output)
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}
