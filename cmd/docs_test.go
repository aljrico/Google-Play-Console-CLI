package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestDocsParityOutputsMarkdownWithoutAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"docs",
		"parity",
		"--output",
		"markdown",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "# Parity Matrix") {
		t.Fatalf("output = %s, want parity matrix markdown", output)
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestDocsParityOutputsJSONDocument(t *testing.T) {
	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"docs",
		"parity",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"name":"parity"`, `"format":"markdown"`, `# Parity Matrix`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
}

func TestDocsCommandsOutputsJSONReferenceWithoutAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"docs",
		"commands",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"name":"gpc"`,
		`"path":"gpc users"`,
		`"path":"gpc users create"`,
		`"path":"gpc purchases product acknowledge"`,
		`"path":"gpc releases"`,
		`"path":"gpc vitals metric-set query"`,
		`"path":"gpc vitals errors issues search"`,
		`"path":"gpc vitals errors reports search"`,
		`"path":"gpc vitals anomalies list"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
	if strings.Contains(output, `"name":"help"`) {
		t.Fatalf("output = %s, did not expect generated help flags", output)
	}
}

func TestDocsCommandsOutputsMarkdownReference(t *testing.T) {
	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"docs",
		"commands",
		"--output",
		"markdown",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		"# Command Reference",
		"`gpc users`",
		"`gpc users create`",
		"`gpc purchases product acknowledge`",
		"`gpc releases`",
		"`gpc vitals metric-set query`",
		"`gpc vitals errors issues search`",
		"`gpc vitals errors reports search`",
		"`gpc vitals anomalies list`",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "`--help`") {
		t.Fatalf("output = %s, did not expect generated help flags", output)
	}
}
