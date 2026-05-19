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
	writeFile(t, filepath.Join(directory, "en-US", "changelogs", "drafts", "43.txt"), "Draft")
	writeFile(t, filepath.Join(directory, "en-US", "images", "README.txt"), "notes")
	writeFile(t, filepath.Join(directory, "en-US", "images", "phoneScreenshots", "1.png"), "png")
	writeFile(t, filepath.Join(directory, "en-US", "notes", "todo.txt"), "todo")
	writeFile(t, filepath.Join(directory, "en-US", "legal", "todo.txt"), "legal")
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
	if inventory.Summary.UnknownFileCount != 4 {
		t.Fatalf("UnknownFileCount = %d, want 4", inventory.Summary.UnknownFileCount)
	}
	if inventory.Locales[0].Language != "en-US" || inventory.Locales[1].Language != "es-ES" {
		t.Fatalf("locale order = %#v", inventory.Locales)
	}
	if got := inventory.Locales[0].ImageSets[0].Type; got != "phoneScreenshots" {
		t.Fatalf("image set type = %q", got)
	}
	gotUnknownNames := make([]string, 0, len(inventory.Locales[0].UnknownFiles))
	for _, file := range inventory.Locales[0].UnknownFiles {
		gotUnknownNames = append(gotUnknownNames, file.Name)
	}
	wantUnknownNames := strings.Join([]string{"changelogs/drafts/43.txt", "images/README.txt", "legal/todo.txt", "notes/todo.txt"}, ",")
	if strings.Join(gotUnknownNames, ",") != wantUnknownNames {
		t.Fatalf("unknown files = %q, want %q", strings.Join(gotUnknownNames, ","), wantUnknownNames)
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
	createSymlinkOrSkip(t, target, link)
	_, err := Inspect(context.Background(), InspectOptions{Directory: link})
	if err == nil {
		t.Fatalf("Inspect() expected error")
	}
}

func TestInspectRejectsSymlinkAncestor(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "outside")
	directory := filepath.Join(target, "metadata", "android")
	writeFile(t, filepath.Join(directory, "en-US", "title.txt"), "Example")
	link := filepath.Join(root, "fastlane")
	createSymlinkOrSkip(t, target, link)
	previousWorkingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previousWorkingDirectory); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})
	_, err = Inspect(context.Background(), InspectOptions{Directory: filepath.Join("fastlane", "metadata", "android")})
	if err == nil {
		t.Fatalf("Inspect() expected error")
	}
	if !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("error = %v, want symlink", err)
	}
}

func TestInspectRejectsSymlinkLocale(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "metadata")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	target := filepath.Join(root, "outside-locale")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}
	createSymlinkOrSkip(t, target, filepath.Join(directory, "en-US"))
	_, err := Inspect(context.Background(), InspectOptions{Directory: directory})
	if err == nil {
		t.Fatalf("Inspect() expected error")
	}
	if !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("error = %v, want symlink", err)
	}
}

func TestInspectRejectsSymlinkFile(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "metadata")
	writeFile(t, filepath.Join(directory, "en-US", "title.txt"), "Example")
	target := filepath.Join(root, "outside.txt")
	writeFile(t, target, "outside")
	createSymlinkOrSkip(t, target, filepath.Join(directory, "en-US", "short_description.txt"))
	_, err := Inspect(context.Background(), InspectOptions{Directory: directory})
	if err == nil {
		t.Fatalf("Inspect() expected error")
	}
	if !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("error = %v, want symlink", err)
	}
}

func TestConvertCreatesMetadataFileFromSupplyListings(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "fastlane", "metadata", "android")
	writeFile(t, filepath.Join(directory, "en-US", "title.txt"), "Example\n")
	writeFile(t, filepath.Join(directory, "en-US", "short_description.txt"), "Short\r\n")
	writeFile(t, filepath.Join(directory, "en-US", "full_description.txt"), "Long\n\n")
	writeFile(t, filepath.Join(directory, "en-US", "video.txt"), "https://youtu.be/example\n")
	writeFile(t, filepath.Join(directory, "en-US", "changelogs", "42.txt"), "Bug fixes")
	writeFile(t, filepath.Join(directory, "es-ES", "title.txt"), "Ejemplo")

	metadata, err := Convert(context.Background(), ConvertOptions{Directory: directory})
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}
	if metadata.Details != nil {
		t.Fatalf("Details = %#v, want nil", metadata.Details)
	}
	if len(metadata.Listings) != 2 {
		t.Fatalf("len(Listings) = %d, want 2", len(metadata.Listings))
	}
	firstListing := metadata.Listings[0]
	if firstListing.Language != "en-US" {
		t.Fatalf("first language = %s, want en-US", firstListing.Language)
	}
	if firstListing.Title == nil || *firstListing.Title != "Example" {
		t.Fatalf("Title = %v, want Example", firstListing.Title)
	}
	if firstListing.ShortDescription == nil || *firstListing.ShortDescription != "Short" {
		t.Fatalf("ShortDescription = %v, want Short", firstListing.ShortDescription)
	}
	if firstListing.FullDescription == nil || *firstListing.FullDescription != "Long" {
		t.Fatalf("FullDescription = %v, want Long", firstListing.FullDescription)
	}
	if firstListing.Video == nil || *firstListing.Video != "https://youtu.be/example" {
		t.Fatalf("Video = %v, want URL", firstListing.Video)
	}
}

func TestConvertChangelogsGroupsReleaseNotesByVersionCode(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "fastlane", "metadata", "android")
	writeFile(t, filepath.Join(directory, "en-US", "changelogs", "42.txt"), "Bug fixes.\n")
	writeFile(t, filepath.Join(directory, "es-ES", "changelogs", "42.txt"), "Correcciones.\r\n")
	writeFile(t, filepath.Join(directory, "en-US", "changelogs", "43.txt"), "New flow.")

	migration, err := ConvertChangelogs(context.Background(), ConvertChangelogsOptions{Directory: directory})
	if err != nil {
		t.Fatalf("ConvertChangelogs() error = %v", err)
	}
	if migration.Directory != filepath.ToSlash(directory) {
		t.Fatalf("Directory = %q, want %q", migration.Directory, filepath.ToSlash(directory))
	}
	if len(migration.Changelogs) != 2 {
		t.Fatalf("len(Changelogs) = %d, want 2", len(migration.Changelogs))
	}
	first := migration.Changelogs[0]
	if first.VersionCode != 42 {
		t.Fatalf("VersionCode = %d, want 42", first.VersionCode)
	}
	if len(first.ReleaseNotes) != 2 {
		t.Fatalf("len(ReleaseNotes) = %d, want 2", len(first.ReleaseNotes))
	}
	if first.ReleaseNotes[0].Language != "en-US" || first.ReleaseNotes[0].Text != "Bug fixes." {
		t.Fatalf("first note = %#v, want en-US bug fixes", first.ReleaseNotes[0])
	}
	if first.ReleaseNotes[1].Language != "es-ES" || first.ReleaseNotes[1].Text != "Correcciones." {
		t.Fatalf("second note = %#v, want es-ES corrections", first.ReleaseNotes[1])
	}
	wantArg := "en-US=Bug fixes."
	if first.ReleaseNoteArgs[0] != wantArg {
		t.Fatalf("ReleaseNoteArgs[0] = %q, want %q", first.ReleaseNoteArgs[0], wantArg)
	}
}

func TestConvertChangelogsFiltersVersionCode(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "fastlane", "metadata", "android")
	writeFile(t, filepath.Join(directory, "en-US", "changelogs", "42.txt"), "Bug fixes.")
	writeFile(t, filepath.Join(directory, "en-US", "changelogs", "43.txt"), "New flow.")

	migration, err := ConvertChangelogs(context.Background(), ConvertChangelogsOptions{Directory: directory, VersionCode: 43})
	if err != nil {
		t.Fatalf("ConvertChangelogs() error = %v", err)
	}
	if len(migration.Changelogs) != 1 {
		t.Fatalf("len(Changelogs) = %d, want 1", len(migration.Changelogs))
	}
	if migration.Changelogs[0].VersionCode != 43 {
		t.Fatalf("VersionCode = %d, want 43", migration.Changelogs[0].VersionCode)
	}
}

func TestConvertChangelogsRejectsInvalidChangelogName(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "fastlane", "metadata", "android")
	writeFile(t, filepath.Join(directory, "en-US", "changelogs", "latest.txt"), "Bug fixes.")

	_, err := ConvertChangelogs(context.Background(), ConvertChangelogsOptions{Directory: directory})
	if err == nil {
		t.Fatalf("ConvertChangelogs() expected error")
	}
	if !strings.Contains(err.Error(), "VERSION_CODE.txt") {
		t.Fatalf("error = %v, want version-code filename error", err)
	}
}

func TestInspectRejectsSymlinkImageSet(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "metadata")
	writeFile(t, filepath.Join(directory, "en-US", "title.txt"), "Example")
	target := filepath.Join(root, "outside-images")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}
	createSymlinkOrSkip(t, target, filepath.Join(directory, "en-US", "images", "phoneScreenshots"))
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

func createSymlinkOrSkip(t *testing.T, oldname string, newname string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(newname), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s) error = %v", filepath.Dir(newname), err)
	}
	if err := os.Symlink(oldname, newname); err != nil {
		if os.IsPermission(err) {
			t.Skipf("symlinks unavailable: %v", err)
		}
		t.Fatalf("Symlink() error = %v", err)
	}
}
