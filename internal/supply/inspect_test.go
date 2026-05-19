package supply

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInspectInventoriesSupplyMetadata(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "fastlane", "metadata", "android")
	writeFile(t, filepath.Join(directory, "en-US", "title.txt"), "Example")
	writeFile(t, filepath.Join(directory, "en-US", "short_description.txt"), "Short")
	writeFile(t, filepath.Join(directory, "en-US", "changelogs", "42.txt"), "Bug fixes")
	writeFile(t, filepath.Join(directory, "en-US", "images", "phoneScreenshots", "1.png"), "png")
	writeFile(t, filepath.Join(directory, "en-US", "notes", "todo.txt"), "todo")
	writeFile(t, filepath.Join(directory, "es-ES", "full_description.txt"), "Completo")

	inventory, err := Inspect(context.Background(), InspectOptions{Directory: directory})
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if inventory.Directory != filepath.ToSlash(directory) {
		t.Fatalf("Directory = %q, want %q", inventory.Directory, filepath.ToSlash(directory))
	}
	if inventory.Summary.LocaleCount != 2 {
		t.Fatalf("LocaleCount = %d, want 2", inventory.Summary.LocaleCount)
	}
	if inventory.Summary.ListingFileCount != 3 {
		t.Fatalf("ListingFileCount = %d, want 3", inventory.Summary.ListingFileCount)
	}
	if inventory.Summary.ChangelogCount != 1 {
		t.Fatalf("ChangelogCount = %d, want 1", inventory.Summary.ChangelogCount)
	}
	if inventory.Summary.ImageFileCount != 1 {
		t.Fatalf("ImageFileCount = %d, want 1", inventory.Summary.ImageFileCount)
	}
	if inventory.Summary.UnknownFileCount != 1 {
		t.Fatalf("UnknownFileCount = %d, want 1", inventory.Summary.UnknownFileCount)
	}
	if inventory.Locales[0].Language != "en-US" || inventory.Locales[1].Language != "es-ES" {
		t.Fatalf("locale order = %#v", inventory.Locales)
	}
	if got := inventory.Locales[0].ImageSets[0].Type; got != "phoneScreenshots" {
		t.Fatalf("image set type = %q", got)
	}
}

func TestInspectRejectsMissingDirectory(t *testing.T) {
	_, err := Inspect(context.Background(), InspectOptions{Directory: filepath.Join(t.TempDir(), "missing")})
	if err == nil {
		t.Fatalf("Inspect() expected error")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("error = %v, want missing directory", err)
	}
}

func TestInspectRejectsSymlinkRoot(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("Symlink() error = %v", err)
	}
	_, err := Inspect(context.Background(), InspectOptions{Directory: link})
	if err == nil {
		t.Fatalf("Inspect() expected error")
	}
}

func TestInspectRejectsSymlinkFile(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "metadata")
	writeFile(t, filepath.Join(directory, "en-US", "title.txt"), "Example")
	target := filepath.Join(root, "outside.txt")
	writeFile(t, target, "outside")
	if err := os.Symlink(target, filepath.Join(directory, "en-US", "short_description.txt")); err != nil {
		t.Fatalf("Symlink() error = %v", err)
	}
	_, err := Inspect(context.Background(), InspectOptions{Directory: directory})
	if err == nil {
		t.Fatalf("Inspect() expected error")
	}
	if !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("error = %v, want symlink", err)
	}
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s) error = %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
}
