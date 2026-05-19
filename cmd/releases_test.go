package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestReleasesUploadRejectsMalformedReleaseNoteBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"upload",
		"--package",
		"com.example.app",
		"--aab",
		"app-release.aab",
		"--release-note",
		"en-US",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected release note validation error")
	}
	if !strings.Contains(err.Error(), "language=text") {
		t.Fatalf("error = %v, want release note format validation", err)
	}
}

func TestReleasesUploadDryRunUsesRequestedTrack(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"upload",
		"--package",
		"com.example.app",
		"--track",
		"beta",
		"--aab",
		"app-release.aab",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	var result struct {
		Track string `json:"track"`
		Plan  struct {
			Track      string `json:"track"`
			Artifact   string `json:"artifact"`
			APKPath    string `json:"apkPath,omitempty"`
			BundlePath string `json:"bundlePath,omitempty"`
			Status     string `json:"status"`
		} `json:"plan"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Unmarshal() error = %v; output = %s", err, buf.String())
	}
	if result.Track != "beta" || result.Plan.Track != "beta" {
		t.Fatalf("release upload dry-run result = %#v, want beta track", result)
	}
	if result.Plan.Artifact != "bundle" || result.Plan.BundlePath != "app-release.aab" || result.Plan.APKPath != "" {
		t.Fatalf("release upload dry-run plan = %#v, want bundle-only artifact", result.Plan)
	}
	if result.Plan.Status != "completed" {
		t.Fatalf("release upload dry-run plan = %#v, want completed status", result.Plan)
	}
}

func TestReleasesUploadDryRunSupportsAPK(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"upload",
		"--package",
		"com.example.app",
		"--track",
		"internal",
		"--apk",
		"app-release.apk",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	var result struct {
		Track string `json:"track"`
		Plan  struct {
			Track      string `json:"track"`
			Artifact   string `json:"artifact"`
			APKPath    string `json:"apkPath,omitempty"`
			BundlePath string `json:"bundlePath,omitempty"`
			Status     string `json:"status"`
		} `json:"plan"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Unmarshal() error = %v; output = %s", err, buf.String())
	}
	if result.Track != "internal" || result.Plan.Track != "internal" {
		t.Fatalf("release upload dry-run result = %#v, want internal track", result)
	}
	if result.Plan.Artifact != "apk" || result.Plan.APKPath != "app-release.apk" || result.Plan.BundlePath != "" {
		t.Fatalf("release upload dry-run plan = %#v, want APK-only artifact", result.Plan)
	}
	if strings.Contains(buf.String(), "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", buf.String())
	}
}

func TestReleasesUploadRejectsMultipleArtifactsBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"upload",
		"--package",
		"com.example.app",
		"--apk",
		"app-release.apk",
		"--aab",
		"app-release.aab",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected multiple artifact validation error")
	}
	if !strings.Contains(err.Error(), "exactly one of APK path or AAB path") {
		t.Fatalf("error = %v, want artifact validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestReleasesUploadLiveRejectsMissingAPKBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"upload",
		"--package",
		"com.example.app",
		"--apk",
		t.TempDir() + "/missing.apk",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected missing APK error")
	}
	if !strings.Contains(err.Error(), "open APK") {
		t.Fatalf("error = %v, want APK preflight", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestReleasesUploadLiveRejectsInvalidUserFractionBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")
	bundlePath := writeRootTestFile(t, "app-release.aab")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"upload",
		"--package",
		"com.example.app",
		"--aab",
		bundlePath,
		"--user-fraction",
		"0.25",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected user fraction validation error")
	}
	if !strings.Contains(err.Error(), "user fraction can only be set") {
		t.Fatalf("error = %v, want user fraction validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestReleasesPromoteDryRun(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"promote",
		"--package",
		"com.example.app",
		"--from",
		"internal",
		"--to",
		"production",
		"--version-code",
		"42",
		"--release-note",
		"en-US=Production rollout.",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"toTrack":"production"`) {
		t.Fatalf("release promote dry-run output = %s", buf.String())
	}
	if !strings.Contains(buf.String(), `"text":"Production rollout."`) {
		t.Fatalf("release promote dry-run output = %s, want release note", buf.String())
	}
}

func TestReleasesHaltDryRun(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"halt",
		"--package",
		"com.example.app",
		"--track",
		"production",
		"--version-code",
		"42",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"action":"halt"`) {
		t.Fatalf("release halt dry-run output = %s", buf.String())
	}
}

func TestReleasesResumeDryRun(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"resume",
		"--package",
		"com.example.app",
		"--track",
		"production",
		"--version-code",
		"42",
		"--status",
		"inProgress",
		"--user-fraction",
		"0.25",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"action":"resume"`) {
		t.Fatalf("release resume dry-run output = %s", buf.String())
	}
}

func TestReleasesResumeCompletedDryRun(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"resume",
		"--package",
		"com.example.app",
		"--track",
		"production",
		"--version-code",
		"42",
		"--status",
		"completed",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"status":"completed"`) {
		t.Fatalf("release resume completed dry-run output = %s", buf.String())
	}
}
