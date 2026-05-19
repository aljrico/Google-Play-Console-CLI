package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestDiffJSONOutputsChangesWithoutAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")
	from := writeRootTestContent(t, "from.json", `{"title":"Old","screenshots":["one.png"]}`)
	to := writeRootTestContent(t, "to.json", `{"title":"New","screenshots":["one.png","two.png"]}`)

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"diff",
		"json",
		from,
		to,
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"equal":false`, `"path":"/screenshots/1"`, `"kind":"added"`, `"path":"/title"`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestDiffJSONFailOnChangeWritesResultAndReturnsError(t *testing.T) {
	from := writeRootTestContent(t, "from.json", `{"title":"Old"}`)
	to := writeRootTestContent(t, "to.json", `{"title":"New"}`)

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"diff",
		"json",
		from,
		to,
		"--fail-on-change",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want fail-on-change error")
	}
	if !strings.Contains(err.Error(), "JSON files differ: 1 change(s)") {
		t.Fatalf("error = %v, want change count", err)
	}
	if !strings.Contains(buf.String(), `"equal":false`) {
		t.Fatalf("output = %s, want diff result before error", buf.String())
	}
}

func TestDiffJSONFailOnChangeAllowsEqualFiles(t *testing.T) {
	from := writeRootTestContent(t, "from.json", `{"title":"Same"}`)
	to := writeRootTestContent(t, "to.json", `{"title":"Same"}`)

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"diff",
		"json",
		from,
		to,
		"--fail-on-change",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"equal":true`, `"changes":[]`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
}
