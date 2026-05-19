package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/aljrico/Google-Play-Console-CLI/internal/websurface"
)

func TestWebStatusDocumentsBlockedSurfaceWithoutAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"web",
		"status",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	var status websurface.Status
	if err := json.Unmarshal(buf.Bytes(), &status); err != nil {
		t.Fatalf("Unmarshal() error = %v; output = %s", err, buf.String())
	}
	if status.Status != "blocked" || status.Surface != "Play Console browser workflows" || len(status.Alternatives) == 0 {
		t.Fatalf("status = %#v, want blocked web boundary", status)
	}
	output := buf.String()
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}
