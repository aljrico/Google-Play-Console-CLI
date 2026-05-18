package output

import (
	"bytes"
	"strings"
	"testing"
)

type sampleRow struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

func TestWriteJSONPretty(t *testing.T) {
	var buf bytes.Buffer

	err := Write(&buf, JSON, true, sampleRow{Name: "internal", Status: "completed"})
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if !strings.Contains(buf.String(), "\n  \"name\"") {
		t.Fatalf("pretty JSON output = %q", buf.String())
	}
}

func TestWritePlainTableForStructSlice(t *testing.T) {
	var buf bytes.Buffer

	err := Write(&buf, Table, false, []sampleRow{
		{Name: "internal", Status: "completed"},
	})
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "name") || !strings.Contains(output, "internal") {
		t.Fatalf("table output = %q", output)
	}
}

func TestDefaultOutputUsesWriter(t *testing.T) {
	var buf bytes.Buffer

	err := Write(&buf, "", false, sampleRow{Name: "internal", Status: "completed"})
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, `"name":"internal"`) {
		t.Fatalf("default output = %q, want compact JSON for non-file writer", output)
	}
}

func TestWriteMarkdownTableForStruct(t *testing.T) {
	var buf bytes.Buffer

	err := Write(&buf, Markdown, false, sampleRow{Name: "internal", Status: "completed"})
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "| name | status |") {
		t.Fatalf("markdown output = %q", output)
	}
}

func TestUnsupportedFormat(t *testing.T) {
	var format Format

	if err := format.Set("yaml"); err == nil {
		t.Fatal("expected unsupported format error")
	}
}
