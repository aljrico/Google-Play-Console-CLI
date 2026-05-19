package cmd

import (
	"bytes"
	"strings"
	"testing"
)

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
