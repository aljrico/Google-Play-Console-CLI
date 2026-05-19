package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrateSupplyInspectOutputsInventoryWithoutAuth(t *testing.T) {
	root := t.TempDir()
	directory := root + "/fastlane/metadata/android"
	writeRootTestPathContent(t, directory+"/en-US/title.txt", "Example")
	writeRootTestPathContent(t, directory+"/en-US/changelogs/42.txt", "Bug fixes")
	writeRootTestPathContent(t, directory+"/en-US/images/phoneScreenshots/1.png", "png")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"migrate",
		"supply",
		"inspect",
		"--directory",
		directory,
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"localeCount":1`,
		`"language":"en-US"`,
		`"name":"title.txt"`,
		`"type":"phoneScreenshots"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestMigrateSupplyConvertOutputsMetadataWithoutAuth(t *testing.T) {
	root := t.TempDir()
	directory := root + "/fastlane/metadata/android"
	writeRootTestPathContent(t, directory+"/en-US/title.txt", "Example\n")
	writeRootTestPathContent(t, directory+"/en-US/short_description.txt", "Short\n")
	writeRootTestPathContent(t, directory+"/es-ES/title.txt", "Ejemplo")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"migrate",
		"supply",
		"convert",
		"--directory",
		directory,
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"listings"`,
		`"language":"en-US"`,
		`"title":"Example"`,
		`"shortDescription":"Short"`,
		`"language":"es-ES"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestMigrateSupplyChangelogsDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	directory := filepath.Join(t.TempDir(), "fastlane", "metadata", "android")
	writeNestedRootTestFile(t, filepath.Join(directory, "en-US", "changelogs", "42.txt"), "Bug fixes.\n")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"migrate",
		"supply",
		"changelogs",
		"--directory",
		directory,
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"versionCode":42`, `"releaseNotes":[{"language":"en-US","text":"Bug fixes."}]`, `"releaseNoteArgs":["en-US=Bug fixes."]`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestMigrateSupplyImagesDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	directory := filepath.Join(t.TempDir(), "fastlane", "metadata", "android")
	imagePath := filepath.Join(directory, "en-US", "images", "phoneScreenshots", "1.png")
	writeNestedRootTestFile(t, imagePath, "png")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"migrate",
		"supply",
		"images",
		"--directory",
		directory,
		"--language",
		"en-US",
		"--type",
		"phoneScreenshots",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"language":"en-US"`, `"type":"phoneScreenshots"`, `"uploadArgs":["--language","en-US","--type","phoneScreenshots","--file","` + filepath.ToSlash(imagePath) + `"]`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}
