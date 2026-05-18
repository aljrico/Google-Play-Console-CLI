package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionJSON(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{"version", "--output", "json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"version"`) {
		t.Fatalf("version output = %s", buf.String())
	}
}

func TestUnknownOutputFormat(t *testing.T) {
	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{"version", "--output", "yaml"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error")
	}
}

func TestVersionRejectsUnexpectedArgs(t *testing.T) {
	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{"version", "stray"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error")
	}
}

func TestPublishInternalDryRunRejectsInvalidPackage(t *testing.T) {
	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"publish",
		"internal",
		"--package",
		"bad",
		"--aab",
		"app-release.aab",
		"--dry-run",
	})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error")
	}
}

func TestPublishInternalDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"publish",
		"internal",
		"--package",
		"com.example.app",
		"--aab",
		"app-release.aab",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"dryRun":true`) {
		t.Fatalf("publish dry-run output = %s", buf.String())
	}
}

func TestReleasesUploadDryRunUsesRequestedTrack(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"upload",
		"--package",
		"com.example.app",
		"--track",
		"beta",
		"--aab",
		"app-release.aab",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"track":"beta"`) {
		t.Fatalf("release upload dry-run output = %s", buf.String())
	}
	if !strings.Contains(buf.String(), `"status":"completed"`) {
		t.Fatalf("release upload dry-run output = %s", buf.String())
	}
}

func TestReleasesPromoteDryRun(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"promote",
		"--package",
		"com.example.app",
		"--from",
		"internal",
		"--to",
		"production",
		"--version-code",
		"42",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"toTrack":"production"`) {
		t.Fatalf("release promote dry-run output = %s", buf.String())
	}
}

func TestReleasesHaltDryRun(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"halt",
		"--package",
		"com.example.app",
		"--track",
		"production",
		"--version-code",
		"42",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"action":"halt"`) {
		t.Fatalf("release halt dry-run output = %s", buf.String())
	}
}

func TestReleasesResumeDryRun(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"resume",
		"--package",
		"com.example.app",
		"--track",
		"production",
		"--version-code",
		"42",
		"--status",
		"inProgress",
		"--user-fraction",
		"0.25",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"action":"resume"`) {
		t.Fatalf("release resume dry-run output = %s", buf.String())
	}
}

func TestReleasesResumeCompletedDryRun(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"resume",
		"--package",
		"com.example.app",
		"--track",
		"production",
		"--version-code",
		"42",
		"--status",
		"completed",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"status":"completed"`) {
		t.Fatalf("release resume completed dry-run output = %s", buf.String())
	}
}

func TestListingsUpdateDryRun(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"listings",
		"update",
		"--package",
		"com.example.app",
		"--language",
		"en-US",
		"--title",
		"Example",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"language":"en-US"`) {
		t.Fatalf("listing update dry-run output = %s", buf.String())
	}
}

func TestListingsDeleteDryRun(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"listings",
		"delete",
		"--package",
		"com.example.app",
		"--language",
		"en-US",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"language":"en-US"`) {
		t.Fatalf("listing delete dry-run output = %s", buf.String())
	}
}

func TestListingsDeleteAllDryRun(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"listings",
		"delete-all",
		"--package",
		"com.example.app",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"all":true`) {
		t.Fatalf("listing delete-all dry-run output = %s", buf.String())
	}
}

func TestDetailsUpdateDryRun(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"details",
		"update",
		"--package",
		"com.example.app",
		"--contact-email",
		"support@example.com",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"contactEmail":"support@example.com"`) {
		t.Fatalf("details update dry-run output = %s", buf.String())
	}
}

func TestReviewsReplyDryRun(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"reviews",
		"reply",
		"--package",
		"com.example.app",
		"--review-id",
		"review-123",
		"--text",
		"Thanks for trying the app.",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"reviewId":"review-123"`) {
		t.Fatalf("reviews reply dry-run output = %s", buf.String())
	}
	if !strings.Contains(buf.String(), `"dryRun":true`) {
		t.Fatalf("reviews reply dry-run output = %s", buf.String())
	}
}

func TestReviewsListRejectsInvalidMaxResultsBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"reviews",
		"list",
		"--package",
		"com.example.app",
		"--max-results",
		"101",
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

func TestReviewsReplyRequiresDryRunOrConfirmBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"reviews",
		"reply",
		"--package",
		"com.example.app",
		"--review-id",
		"review-123",
		"--text",
		"Thanks.",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected confirmation validation error")
	}
}

func TestInAppProductsGetRejectsMissingSKUBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"get",
		"--package",
		"com.example.app",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected SKU validation error")
	}
	if !strings.Contains(err.Error(), "SKU") {
		t.Fatalf("error = %v, want SKU validation", err)
	}
}

func TestSubscriptionsListRejectsInvalidPageSizeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"list",
		"--package",
		"com.example.app",
		"--page-size",
		"1001",
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

func TestSubscriptionsGetRejectsInvalidProductIDBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"get",
		"--package",
		"com.example.app",
		"--product-id",
		"Premium",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected product ID validation error")
	}
	if !strings.Contains(err.Error(), "subscription product ID") {
		t.Fatalf("error = %v, want product ID validation", err)
	}
}

func TestSubscriptionOffersListRejectsInvalidPageSizeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"list",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--page-size",
		"1001",
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

func TestSubscriptionOffersGetRejectsMissingOfferIDBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"get",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected offer ID validation error")
	}
	if !strings.Contains(err.Error(), "subscription offer ID") {
		t.Fatalf("error = %v, want offer ID validation", err)
	}
}
