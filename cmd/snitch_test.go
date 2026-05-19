package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestSnitchReportOutputsIssueURLWithoutAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"snitch",
		"report",
		"--title",
		"Confusing release output",
		"--body",
		"The track summary was hard to read.",
		"--command",
		"gpc releases list --package com.example.app",
		"--label",
		"ux",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"repository":"aljrico/Google-Play-Console-CLI"`,
		`"title":"Confusing release output"`,
		`"command":"gpc releases list --package com.example.app"`,
		`"labels":["snitch","ux"]`,
		`"issueUrl":"https://github.com/aljrico/Google-Play-Console-CLI/issues/new?`,
		`&labels=snitch%2Cux&`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}
