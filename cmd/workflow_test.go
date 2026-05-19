package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestWorkflowListDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	workflowFile := writeRootTestContent(t, "workflow.json", `{
	  "version": 1,
	  "workflows": {
	    "release": [{"run": "echo release"}]
	  }
	}`)

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"workflow",
		"--file",
		workflowFile,
		"list",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"name":"release"`, `"steps":1`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestWorkflowRunDryRunDoesNotExecute(t *testing.T) {
	workflowFile := writeRootTestContent(t, "workflow.json", `{
	  "version": 1,
	  "workflows": {
	    "release": [{"run": "exit 1"}]
	  }
	}`)

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"workflow",
		"--file",
		workflowFile,
		"run",
		"release",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"dryRun":true`, `"skipped":true`, `"success":true`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
}

func TestWorkflowRunFailureWritesResultAndReturnsError(t *testing.T) {
	workflowFile := writeRootTestContent(t, "workflow.json", `{
	  "version": 1,
	  "workflows": {
	    "release": [{"run": "printf nope; exit 7"}]
	  }
	}`)

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"workflow",
		"--file",
		workflowFile,
		"run",
		"release",
		"--confirm",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want workflow failure")
	}
	output := buf.String()
	for _, want := range []string{`"success":false`, `"exitCode":7`, `"stdout":"nope"`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
}

func TestWorkflowRunMissingFileReturnsErrorWithoutZeroResult(t *testing.T) {
	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"workflow",
		"--file",
		t.TempDir() + "/missing.json",
		"run",
		"release",
		"--confirm",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want missing workflow file")
	}
	if buf.String() != "" {
		t.Fatalf("output = %s, want empty output", buf.String())
	}
}

func TestWorkflowRunRejectsConfirmAndDryRun(t *testing.T) {
	workflowFile := writeRootTestContent(t, "workflow.json", `{
	  "version": 1,
	  "workflows": {
	    "release": [{"run": "echo release"}]
	  }
	}`)

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"workflow",
		"--file",
		workflowFile,
		"run",
		"release",
		"--confirm",
		"--dry-run",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want confirm and dry-run conflict")
	}
	if !strings.Contains(err.Error(), "--confirm and --dry-run cannot be used together") {
		t.Fatalf("error = %v, want confirm and dry-run conflict", err)
	}
	if buf.String() != "" {
		t.Fatalf("output = %s, want empty output", buf.String())
	}
}
