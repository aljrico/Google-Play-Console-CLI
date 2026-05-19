package agentskills

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestListReturnsBundledSkills(t *testing.T) {
	skills := List()
	if len(skills) != 3 {
		t.Fatalf("len(List()) = %d, want 3", len(skills))
	}
	names := make(map[string]bool, len(skills))
	for _, skill := range skills {
		names[skill.Name] = true
		if skill.Description == "" {
			t.Fatalf("skill %s has empty description", skill.Name)
		}
	}
	for _, want := range []string{"gpc-cli-usage", "gpc-metadata-workflow", "gpc-release-flow"} {
		if !names[want] {
			t.Fatalf("List() missing %s: %#v", want, skills)
		}
	}
}

func TestInstallDryRunDoesNotWriteFiles(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "skills")
	result, err := Install(context.Background(), InstallOptions{
		Directory: directory,
		Skills:    []string{"gpc-cli-usage"},
		DryRun:    true,
	})
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if len(result.Skills) != 1 {
		t.Fatalf("len(result.Skills) = %d, want 1", len(result.Skills))
	}
	if !result.Skills[0].Written {
		t.Fatalf("Written = false, want true for dry-run plan")
	}
	if _, err := os.Stat(result.Skills[0].Path); !os.IsNotExist(err) {
		t.Fatalf("skill file exists after dry-run or stat error = %v", err)
	}
}

func TestInstallWritesSelectedSkill(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "skills")
	result, err := Install(context.Background(), InstallOptions{
		Directory: directory,
		Skills:    []string{"gpc-release-flow"},
	})
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if len(result.Skills) != 1 {
		t.Fatalf("len(result.Skills) = %d, want 1", len(result.Skills))
	}
	if !result.Skills[0].Written || result.Skills[0].Overwrote {
		t.Fatalf("skill result = %#v, want written without overwrite", result.Skills[0])
	}
	content, err := os.ReadFile(filepath.Join(directory, "gpc-release-flow", "SKILL.md"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if len(content) == 0 {
		t.Fatalf("installed skill is empty")
	}
}

func TestInstallSkipsExistingWithoutForce(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "skills")
	path := filepath.Join(directory, "gpc-cli-usage", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte("custom"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	result, err := Install(context.Background(), InstallOptions{
		Directory: directory,
		Skills:    []string{"gpc-cli-usage"},
	})
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if !result.Skills[0].Exists || result.Skills[0].Written {
		t.Fatalf("skill result = %#v, want existing skipped", result.Skills[0])
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(content) != "custom" {
		t.Fatalf("content = %q, want custom", content)
	}
}

func TestInstallForceOverwritesExisting(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "skills")
	path := filepath.Join(directory, "gpc-cli-usage", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte("custom"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	result, err := Install(context.Background(), InstallOptions{
		Directory: directory,
		Skills:    []string{"gpc-cli-usage"},
		Force:     true,
	})
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if !result.Skills[0].Exists || !result.Skills[0].Written || !result.Skills[0].Overwrote {
		t.Fatalf("skill result = %#v, want forced overwrite", result.Skills[0])
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(content) == "custom" {
		t.Fatalf("content was not overwritten")
	}
}

func TestInstallRejectsUnknownSkill(t *testing.T) {
	_, err := Install(context.Background(), InstallOptions{
		Directory: t.TempDir(),
		Skills:    []string{"nope"},
	})
	if err == nil {
		t.Fatalf("Install() expected error")
	}
}

func TestInstallRejectsSymlinkDirectory(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("Symlink() error = %v", err)
	}
	_, err := Install(context.Background(), InstallOptions{Directory: link})
	if err == nil {
		t.Fatalf("Install() expected error")
	}
}
