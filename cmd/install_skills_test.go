package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
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
		"playpub-cli-usage",
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
		`"name":"playpub-cli-usage"`,
		`"written":false`,
		`"wouldWrite":true`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if _, err := os.Stat(directory + "/playpub-cli-usage/SKILL.md"); !os.IsNotExist(err) {
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
	for _, want := range []string{
		`"name":"playpub-cli-usage"`,
		`"name":"playpub-metadata-workflow"`,
		`"name":"playpub-release-flow"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
}
