package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/aljrico/Google-Play-Console-CLI/internal/agentskills"
)

func TestInstallSkillsDryRunOutputsJSONWithoutWriting(t *testing.T) {
	directory := t.TempDir() + "/skills"

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"install-skills",
		"--directory",
		directory,
		"--skill",
		"gpc-cli-usage",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"directory":"` + directory + `"`,
		`"dryRun":true`,
		`"name":"gpc-cli-usage"`,
		`"written":false`,
		`"wouldWrite":true`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if _, err := os.Stat(directory + "/gpc-cli-usage/SKILL.md"); !os.IsNotExist(err) {
		t.Fatalf("skill file exists after dry-run or stat error = %v", err)
	}
}

func TestInstallSkillsListOutputsBundledSkillNames(t *testing.T) {
	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"install-skills",
		"list",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, name := range agentskills.BundledNames() {
		want := fmt.Sprintf(`"name":%q`, name)
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
}
