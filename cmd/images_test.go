package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestImagesListRejectsInvalidTypeBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"images",
		"list",
		"--package",
		"com.example.app",
		"--language",
		"en-US",
		"--type",
		"poster",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected image type validation error")
	}
	if !strings.Contains(err.Error(), "unsupported image type") {
		t.Fatalf("error = %v, want image type validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestImagesListRejectsMissingTypeBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"images",
		"list",
		"--package",
		"com.example.app",
		"--language",
		"en-US",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected image type validation error")
	}
	if !strings.Contains(err.Error(), "image type is required") {
		t.Fatalf("error = %v, want missing image type validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestImagesUploadDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")
	imagePath := writeRootTestFile(t, "feature.png")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"images",
		"upload",
		"--package",
		"com.example.app",
		"--language",
		"en-US",
		"--type",
		"featureGraphic",
		"--file",
		imagePath,
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	var result struct {
		Path   string `json:"path"`
		DryRun bool   `json:"dryRun"`
		Plan   struct {
			Path string `json:"path"`
		} `json:"plan"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Unmarshal() error = %v; output = %s", err, buf.String())
	}
	if result.Path != imagePath || result.Plan.Path != imagePath || !result.DryRun {
		t.Fatalf("result = %#v, want image upload dry-run", result)
	}
	if strings.Contains(buf.String(), "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", buf.String())
	}
}

func TestImagesUploadDryRunRejectsMissingFileBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"images",
		"upload",
		"--package",
		"com.example.app",
		"--language",
		"en-US",
		"--type",
		"featureGraphic",
		"--file",
		t.TempDir() + "/missing.png",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected missing image error")
	}
	if !strings.Contains(err.Error(), "open image") {
		t.Fatalf("error = %v, want image preflight", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestImagesUploadRejectsMissingFileBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"images",
		"upload",
		"--package",
		"com.example.app",
		"--language",
		"en-US",
		"--type",
		"featureGraphic",
		"--file",
		t.TempDir() + "/missing.png",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected missing image error")
	}
	if !strings.Contains(err.Error(), "open image") {
		t.Fatalf("error = %v, want image preflight", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestImagesDeleteDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"images",
		"delete",
		"--package",
		"com.example.app",
		"--language",
		"en-US",
		"--type",
		"phoneScreenshots",
		"--image-id",
		"image-1",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	var result struct {
		ImageID string `json:"imageId"`
		DryRun  bool   `json:"dryRun"`
		Plan    struct {
			ImageID string `json:"imageId"`
			All     bool   `json:"all"`
		} `json:"plan"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Unmarshal() error = %v; output = %s", err, buf.String())
	}
	if result.ImageID != "image-1" || result.Plan.ImageID != "image-1" || result.Plan.All || !result.DryRun {
		t.Fatalf("result = %#v, want single image delete dry-run", result)
	}
	if strings.Contains(buf.String(), "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", buf.String())
	}
}

func TestImagesDeleteAllDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"images",
		"delete-all",
		"--package",
		"com.example.app",
		"--language",
		"en-US",
		"--type",
		"featureGraphic",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	var result struct {
		All    bool `json:"all"`
		DryRun bool `json:"dryRun"`
		Plan   struct {
			All bool `json:"all"`
		} `json:"plan"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Unmarshal() error = %v; output = %s", err, buf.String())
	}
	if !result.All || !result.Plan.All || !result.DryRun {
		t.Fatalf("result = %#v, want delete-all dry-run", result)
	}
	if strings.Contains(buf.String(), "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", buf.String())
	}
}

func TestImagesDeleteRejectsMissingImageIDBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"images",
		"delete",
		"--package",
		"com.example.app",
		"--language",
		"en-US",
		"--type",
		"phoneScreenshots",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected missing image ID error")
	}
	if !strings.Contains(err.Error(), "image ID is required") {
		t.Fatalf("error = %v, want image ID validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestImagesDeleteAllRejectsInvalidTypeBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"images",
		"delete-all",
		"--package",
		"com.example.app",
		"--language",
		"en-US",
		"--type",
		"bad",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected image type validation error")
	}
	if !strings.Contains(err.Error(), "unsupported image type") {
		t.Fatalf("error = %v, want image type validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}
