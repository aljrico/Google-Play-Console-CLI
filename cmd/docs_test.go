package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestDocsParityOutputsMarkdownWithoutAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

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
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

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
		`"name":"playpub"`,
		`"path":"playpub users"`,
		`"path":"playpub users create"`,
		`"path":"playpub purchases product acknowledge"`,
		`"path":"playpub releases"`,
		`"path":"playpub vitals metric-set query"`,
		`"path":"playpub vitals errors issues search"`,
		`"path":"playpub vitals errors reports search"`,
		`"path":"playpub vitals anomalies list"`,
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
		"`playpub users`",
		"`playpub users create`",
		"`playpub purchases product acknowledge`",
		"`playpub releases`",
		"`playpub vitals metric-set query`",
		"`playpub vitals errors issues search`",
		"`playpub vitals errors reports search`",
		"`playpub vitals anomalies list`",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "`--help`") {
		t.Fatalf("output = %s, did not expect generated help flags", output)
	}
}
