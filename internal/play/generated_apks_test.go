package play

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestListGeneratedAPKsPassesOptionsToLister(t *testing.T) {
	lister := &fakeGeneratedAPKLister{result: GeneratedAPKListResult{
		PackageName: "com.example.app",
		VersionCode: 42,
		SigningKeys: []GeneratedAPKSigningKey{
			{CertificateSHA256Hash: "abc", SplitAPKs: []GeneratedSplitAPK{}},
		},
	}}
	options := GeneratedAPKListOptions{PackageName: "com.example.app", VersionCode: 42}

	result, err := ListGeneratedAPKs(context.Background(), lister, options)
	if err != nil {
		t.Fatalf("ListGeneratedAPKs() error = %v", err)
	}
	if len(result.SigningKeys) != 1 {
		t.Fatalf("len(SigningKeys) = %d, want 1", len(result.SigningKeys))
	}
	if !reflect.DeepEqual(lister.options, options) {
		t.Fatalf("options = %#v, want %#v", lister.options, options)
	}
}

func TestListGeneratedAPKsRejectsInvalidOptions(t *testing.T) {
	tests := []GeneratedAPKListOptions{
		{},
		{PackageName: "bad", VersionCode: 42},
		{PackageName: "com.example.app"},
		{PackageName: "com.example.app", VersionCode: -1},
	}
	for _, options := range tests {
		_, err := ListGeneratedAPKs(context.Background(), nil, options)
		if err == nil {
			t.Fatalf("ListGeneratedAPKs(%#v) expected validation error", options)
		}
	}
}

func TestGeneratedAPKJSONPreservesZeroIDs(t *testing.T) {
	payload, err := json.Marshal(GeneratedAPKSigningKey{
		SplitAPKs: []GeneratedSplitAPK{
			{DownloadID: "split-download", VariantID: 0},
		},
		StandaloneAPKs: []GeneratedStandaloneAPK{
			{DownloadID: "standalone-download", VariantID: 0},
		},
		AssetPackSlices: []GeneratedAssetPackSlice{
			{DownloadID: "asset-download", Version: 0},
		},
		RecoveryModules: []GeneratedRecoveryAPK{},
	})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	output := string(payload)
	for _, want := range []string{`"variantId":0`, `"version":0`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
}

func TestDownloadGeneratedAPKDryRunDoesNotCallDownloader(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "split.apk")

	result, err := DownloadGeneratedAPK(context.Background(), nil, GeneratedAPKDownloadOptions{
		PackageName: "com.example.app",
		VersionCode: 42,
		DownloadID:  "split-download",
		OutputPath:  outputPath,
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("DownloadGeneratedAPK() error = %v", err)
	}
	if result.Downloaded {
		t.Fatalf("Downloaded = true, want false")
	}
	if result.Plan.OutputPath != outputPath || len(result.Plan.Steps) == 0 {
		t.Fatalf("plan = %#v, want output path and steps", result.Plan)
	}
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		t.Fatalf("output file stat error = %v, want not exist", err)
	}
}

func TestDownloadGeneratedAPKRejectsExistingOutputWithoutForce(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "split.apk")
	if err := os.WriteFile(outputPath, []byte("existing"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := DownloadGeneratedAPK(context.Background(), nil, GeneratedAPKDownloadOptions{
		PackageName: "com.example.app",
		VersionCode: 42,
		DownloadID:  "split-download",
		OutputPath:  outputPath,
		DryRun:      true,
	})
	if err == nil {
		t.Fatal("expected existing output error")
	}
	if !strings.Contains(err.Error(), "--force") {
		t.Fatalf("error = %v, want force hint", err)
	}
}

func TestDownloadGeneratedAPKRejectsSymlinkOutput(t *testing.T) {
	dir := t.TempDir()
	targetPath := filepath.Join(dir, "target.apk")
	outputPath := filepath.Join(dir, "split.apk")
	if err := os.WriteFile(targetPath, []byte("target"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.Symlink(targetPath, outputPath); err != nil {
		t.Fatalf("Symlink() error = %v", err)
	}

	_, err := DownloadGeneratedAPK(context.Background(), nil, GeneratedAPKDownloadOptions{
		PackageName: "com.example.app",
		VersionCode: 42,
		DownloadID:  "split-download",
		OutputPath:  outputPath,
		Force:       true,
		DryRun:      true,
	})
	if err == nil {
		t.Fatal("expected symlink output error")
	}
	if !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("error = %v, want symlink validation", err)
	}
}

func TestDownloadGeneratedAPKPassesOptionsToDownloader(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "split.apk")
	downloader := &fakeGeneratedAPKDownloader{result: GeneratedAPKDownloadResult{Downloaded: true}}
	options := GeneratedAPKDownloadOptions{
		PackageName: "com.example.app",
		VersionCode: 42,
		DownloadID:  "split-download",
		OutputPath:  outputPath,
	}

	result, err := DownloadGeneratedAPK(context.Background(), downloader, options)
	if err != nil {
		t.Fatalf("DownloadGeneratedAPK() error = %v", err)
	}
	if !result.Downloaded {
		t.Fatalf("Downloaded = false, want true")
	}
	if !reflect.DeepEqual(downloader.options, options) {
		t.Fatalf("options = %#v, want %#v", downloader.options, options)
	}
}

func TestGeneratedAPKDownloadIDAllowsOpaqueSlashes(t *testing.T) {
	downloadID, err := NewGeneratedAPKDownloadID("bundle/split")
	if err != nil {
		t.Fatalf("NewGeneratedAPKDownloadID() error = %v", err)
	}
	if downloadID.String() != "bundle/split" {
		t.Fatalf("downloadID = %q, want opaque value", downloadID)
	}
}

func TestDownloadGeneratedAPKRejectsInvalidOptions(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "split.apk")
	tests := []GeneratedAPKDownloadOptions{
		{},
		{PackageName: "bad", VersionCode: 42, DownloadID: "split-download", OutputPath: outputPath},
		{PackageName: "com.example.app", DownloadID: "split-download", OutputPath: outputPath},
		{PackageName: "com.example.app", VersionCode: 42, OutputPath: outputPath},
		{PackageName: "com.example.app", VersionCode: 42, DownloadID: "split-download", OutputPath: filepath.Join(t.TempDir(), "split.zip")},
	}
	for _, options := range tests {
		_, err := DownloadGeneratedAPK(context.Background(), nil, options)
		if err == nil {
			t.Fatalf("DownloadGeneratedAPK(%#v) expected validation error", options)
		}
	}
}

type fakeGeneratedAPKLister struct {
	options GeneratedAPKListOptions
	result  GeneratedAPKListResult
}

func (l *fakeGeneratedAPKLister) ListGeneratedAPKs(ctx context.Context, options GeneratedAPKListOptions) (GeneratedAPKListResult, error) {
	l.options = options
	return l.result, nil
}

type fakeGeneratedAPKDownloader struct {
	options GeneratedAPKDownloadOptions
	result  GeneratedAPKDownloadResult
}

func (d *fakeGeneratedAPKDownloader) DownloadGeneratedAPK(ctx context.Context, options GeneratedAPKDownloadOptions) (GeneratedAPKDownloadResult, error) {
	d.options = options
	return d.result, nil
}
