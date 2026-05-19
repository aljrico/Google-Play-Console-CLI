package project

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestInitWritesWorkspaceFiles(t *testing.T) {
	directory := filepath.Join(t.TempDir(), ".gpc")

	plan, err := Init(context.Background(), InitOptions{Directory: directory})
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	if len(plan.Files) != 2 {
		t.Fatalf("len(Files) = %d, want 2", len(plan.Files))
	}
	for _, file := range plan.Files {
		if !file.Written {
			t.Fatalf("%s Written = false, want true", file.Path)
		}
		if _, err := os.Stat(file.Path); err != nil {
			t.Fatalf("Stat(%s) error = %v", file.Path, err)
		}
	}
}

func TestInitSkipsExistingFilesWithoutForce(t *testing.T) {
	directory := filepath.Join(t.TempDir(), ".gpc")
	readme := filepath.Join(directory, "README.md")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(readme, []byte("keep me"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	plan, err := Init(context.Background(), InitOptions{Directory: directory})
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	for _, file := range plan.Files {
		if file.Path == readme && file.Written {
			t.Fatalf("%s Written = true, want false", file.Path)
		}
	}
	content, err := os.ReadFile(readme)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(content) != "keep me" {
		t.Fatalf("content = %q, want original", content)
	}
}

func TestInitDryRunDoesNotWriteFiles(t *testing.T) {
	directory := filepath.Join(t.TempDir(), ".gpc")

	plan, err := Init(context.Background(), InitOptions{Directory: directory, DryRun: true})
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	if len(plan.Files) != 2 {
		t.Fatalf("len(Files) = %d, want 2", len(plan.Files))
	}
	for _, file := range plan.Files {
		if !file.Written {
			t.Fatalf("%s Written = false, want planned write", file.Path)
		}
		if _, err := os.Stat(file.Path); !os.IsNotExist(err) {
			t.Fatalf("Stat(%s) error = %v, want not exist", file.Path, err)
		}
	}
}

func TestInitForceOverwritesExistingFiles(t *testing.T) {
	directory := filepath.Join(t.TempDir(), ".gpc")
	readme := filepath.Join(directory, "README.md")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(readme, []byte("replace me"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	plan, err := Init(context.Background(), InitOptions{Directory: directory, Force: true})
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	foundOverwrite := false
	for _, file := range plan.Files {
		if file.Path == readme {
			foundOverwrite = file.Overwrote && file.Written
		}
	}
	if !foundOverwrite {
		t.Fatal("expected README overwrite")
	}
}

func TestInitRejectsSymlinkedDirectory(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	link := filepath.Join(root, "link")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("Symlink() error = %v", err)
	}

	_, err := Init(context.Background(), InitOptions{Directory: link})
	if err == nil {
		t.Fatal("expected symlink directory error")
	}
	if _, statErr := os.Stat(filepath.Join(target, "README.md")); !os.IsNotExist(statErr) {
		t.Fatalf("target README stat error = %v, want not exist", statErr)
	}
}

func TestInitRejectsSymlinkedFileEvenWithForce(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, ".gpc")
	target := filepath.Join(root, "outside.md")
	link := filepath.Join(directory, "README.md")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(target, []byte("keep me"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("Symlink() error = %v", err)
	}

	_, err := Init(context.Background(), InitOptions{Directory: directory, Force: true})
	if err == nil {
		t.Fatal("expected symlink file error")
	}
	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(content) != "keep me" {
		t.Fatalf("target content = %q, want unchanged", content)
	}
}
