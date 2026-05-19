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

type pointerRow struct {
	UserFraction *float64 `json:"userFraction,omitempty"`
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

func TestWriteJSONDoesNotEscapeURLSeparators(t *testing.T) {
	var buf bytes.Buffer

	err := Write(&buf, JSON, false, map[string]string{"issueUrl": "https://example.com/issues/new?title=One&labels=snitch"})
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if strings.Contains(buf.String(), `\u0026`) || !strings.Contains(buf.String(), `&labels=snitch`) {
		t.Fatalf("JSON output = %q, want literal URL separators", buf.String())
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

func TestWritePlainTableFlattensMultilineCells(t *testing.T) {
	var buf bytes.Buffer

	err := Write(&buf, Table, false, sampleRow{Name: "internal\ntrack", Status: "needs | review"})
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	output := buf.String()
	if strings.Contains(output, "internal\ntrack") || !strings.Contains(output, "internal track") {
		t.Fatalf("table output = %q, want flattened cell", output)
	}
}

func TestWritePlainTableDereferencesPointers(t *testing.T) {
	var buf bytes.Buffer
	userFraction := 0.25

	err := Write(&buf, Table, false, pointerRow{UserFraction: &userFraction})
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	output := buf.String()
	if strings.Contains(output, "0x") || !strings.Contains(output, "0.25") {
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

func TestWriteMarkdownTableEscapesCellPipes(t *testing.T) {
	var buf bytes.Buffer

	err := Write(&buf, Markdown, false, sampleRow{Name: "internal\ntrack", Status: "needs | review"})
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, `needs \| review`) || strings.Contains(output, "internal\ntrack") {
		t.Fatalf("markdown output = %q, want escaped flattened cell", output)
	}
}

func TestUnsupportedFormat(t *testing.T) {
	var format Format

	if err := format.Set("yaml"); err == nil {
		t.Fatal("expected unsupported format error")
	}
}
