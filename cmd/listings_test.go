package cmd

import (
	"bytes"
	"strings"
	"testing"
)

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
