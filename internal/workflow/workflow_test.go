package workflow

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestListReturnsSortedWorkflowSummaries(t *testing.T) {
	file := writeWorkflowFile(t, `{
	  "version": 1,
	  "workflows": {
	    "release": [{"run": "echo release"}],
	    "doctor": [{"run": "echo one"}, {"run": "echo two"}]
	  }
	}`)

	summaries, err := List(ListOptions{File: file})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	want := []Summary{{Name: "doctor", Steps: 2}, {Name: "release", Steps: 1}}
	if !reflect.DeepEqual(summaries, want) {
		t.Fatalf("summaries = %#v, want %#v", summaries, want)
	}
}

func TestRunDryRunSkipsExecution(t *testing.T) {
	file := writeWorkflowFile(t, `{
	  "version": 1,
	  "workflows": {
	    "release": [{"name": "plan", "run": "exit 1"}]
	  }
	}`)

	result, err := Run(context.Background(), nil, RunOptions{
		File:   file,
		Name:   "release",
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !result.Success || !result.Steps[0].Skipped {
		t.Fatalf("result = %#v, want successful skipped step", result)
	}
}

func TestRunExecutesStepsInOrder(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell workflow execution is not supported on windows yet")
	}
	file := writeWorkflowFile(t, `{
	  "version": 1,
	  "workflows": {
	    "release": [
	      {"run": "printf first"},
	      {"run": "printf second"}
	    ]
	  }
	}`)

	result, err := Run(context.Background(), nil, RunOptions{
		File:    file,
		Name:    "release",
		Confirm: true,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Steps[0].Stdout != "first" || result.Steps[1].Stdout != "second" {
		t.Fatalf("stdout = %q, %q; want first, second", result.Steps[0].Stdout, result.Steps[1].Stdout)
	}
}

func TestRunStopsOnFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell workflow execution is not supported on windows yet")
	}
	file := writeWorkflowFile(t, `{
	  "version": 1,
	  "workflows": {
	    "release": [
	      {"run": "printf before"},
	      {"run": "printf nope; exit 7"},
	      {"run": "printf after"}
	    ]
	  }
	}`)

	result, err := Run(context.Background(), nil, RunOptions{
		File:    file,
		Name:    "release",
		Confirm: true,
	})
	if err == nil {
		t.Fatal("Run() error = nil, want step failure")
	}
	if result.Success {
		t.Fatal("Success = true, want false")
	}
	if len(result.Steps) != 2 {
		t.Fatalf("len(Steps) = %d, want stopped after failure", len(result.Steps))
	}
	if result.Steps[1].ExitCode != 7 {
		t.Fatalf("ExitCode = %d, want 7", result.Steps[1].ExitCode)
	}
}

func TestLoadRejectsUnknownFields(t *testing.T) {
	file := writeWorkflowFile(t, `{
	  "version": 1,
	  "surprise": true,
	  "workflows": {
	    "release": [{"run": "echo release"}]
	  }
	}`)

	_, err := Load(file)
	if err == nil {
		t.Fatal("Load() error = nil, want unknown field error")
	}
}

func TestLoadRejectsTrailingJSON(t *testing.T) {
	file := writeWorkflowFile(t, `{
	  "version": 1,
	  "workflows": {
	    "release": [{"run": "echo release"}]
	  }
	}
	{"extra": true}`)

	_, err := Load(file)
	if err == nil {
		t.Fatal("Load() error = nil, want trailing JSON error")
	}
}

func TestLoadRejectsWorkflowNameWhitespace(t *testing.T) {
	file := writeWorkflowFile(t, `{
	  "version": 1,
	  "workflows": {
	    " release ": [{"run": "echo release"}]
	  }
	}`)

	_, err := Load(file)
	if err == nil {
		t.Fatal("Load() error = nil, want workflow name whitespace error")
	}
}

func TestRunRequiresConfirmOrDryRun(t *testing.T) {
	file := writeWorkflowFile(t, `{
	  "version": 1,
	  "workflows": {
	    "release": [{"run": "echo release"}]
	  }
	}`)

	_, err := Run(context.Background(), nil, RunOptions{
		File: file,
		Name: "release",
	})
	if err == nil {
		t.Fatal("Run() error = nil, want confirm or dry-run error")
	}
}

func TestRunRejectsConfirmAndDryRunTogether(t *testing.T) {
	file := writeWorkflowFile(t, `{
	  "version": 1,
	  "workflows": {
	    "release": [{"run": "echo release"}]
	  }
	}`)

	_, err := Run(context.Background(), nil, RunOptions{
		File:    file,
		Name:    "release",
		DryRun:  true,
		Confirm: true,
	})
	if err == nil {
		t.Fatal("Run() error = nil, want confirm and dry-run conflict")
	}
}

func TestRunDefaultsWorkDirToWorkflowRoot(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, ".playpub", "workflow.json")
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(file, []byte(`{
	  "version": 1,
	  "workflows": {
	    "release": [{"run": "echo release"}]
	  }
	}`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	result, err := Run(context.Background(), nil, RunOptions{
		File:   file,
		Name:   "release",
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.WorkDir != root {
		t.Fatalf("WorkDir = %q, want workflow root %q", result.WorkDir, root)
	}
}

func writeWorkflowFile(t *testing.T, content string) string {
	t.Helper()
	file := filepath.Join(t.TempDir(), "workflow.json")
	if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return file
}
