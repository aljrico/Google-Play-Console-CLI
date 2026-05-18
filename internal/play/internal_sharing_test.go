package play

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestUploadInternalSharingArtifactDryRunDoesNotRequireUploader(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := UploadInternalSharingArtifact(context.Background(), nil, InternalSharingUploadOptions{
		PackageName: packageName,
		BundlePath:  "app-release.aab",
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("UploadInternalSharingArtifact() error = %v", err)
	}
	if result.Kind != InternalSharingArtifactKindBundle {
		t.Fatalf("Kind = %q, want bundle", result.Kind)
	}
	if !result.DryRun {
		t.Fatal("DryRun = false, want true")
	}
}

func TestUploadInternalSharingArtifactUploadsAPK(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	apkPath := writeTestFile(t, "app.apk")
	uploader := &fakeInternalSharingUploader{artifact: InternalSharingArtifact{DownloadURL: "https://example.com/download"}}

	result, err := UploadInternalSharingArtifact(context.Background(), uploader, InternalSharingUploadOptions{
		PackageName: packageName,
		APKPath:     apkPath,
	})
	if err != nil {
		t.Fatalf("UploadInternalSharingArtifact() error = %v", err)
	}
	if uploader.apkPath != apkPath {
		t.Fatalf("apkPath = %q, want %q", uploader.apkPath, apkPath)
	}
	if result.Artifact == nil || result.Artifact.DownloadURL == "" {
		t.Fatalf("Artifact = %#v, want download URL", result.Artifact)
	}
}

func TestInternalSharingUploadOptionsRejectsInvalidInputs(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	tests := []InternalSharingUploadOptions{
		{PackageName: "bad", APKPath: "app.apk"},
		{PackageName: packageName},
		{PackageName: packageName, APKPath: "app.apk", BundlePath: "app.aab"},
		{PackageName: packageName, APKPath: "app.aab"},
		{PackageName: packageName, BundlePath: "app.apk"},
	}
	for _, options := range tests {
		if err := options.Validate(); err == nil {
			t.Fatalf("Validate(%#v) expected error", options)
		}
	}
}

func TestUploadInternalSharingArtifactChecksFileBeforeUploader(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	_, err = UploadInternalSharingArtifact(context.Background(), nil, InternalSharingUploadOptions{
		PackageName: packageName,
		APKPath:     filepath.Join(t.TempDir(), "missing.apk"),
	})
	if err == nil {
		t.Fatal("expected missing APK error")
	}
}

func TestValidateReadableFileRejectsDirectory(t *testing.T) {
	directoryPath := filepath.Join(t.TempDir(), "app.apk")
	if err := os.Mkdir(directoryPath, 0o755); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}

	if err := ValidateReadableFile(directoryPath); err == nil {
		t.Fatal("expected directory rejection")
	}
}

func TestValidateReadableFileRejectsSymlink(t *testing.T) {
	targetPath := writeTestFile(t, "app.apk")
	linkPath := filepath.Join(t.TempDir(), "linked.apk")
	if err := os.Symlink(targetPath, linkPath); err != nil {
		t.Skipf("Symlink() error = %v", err)
	}

	if err := ValidateReadableFile(linkPath); err == nil {
		t.Fatal("expected symlink rejection")
	}
}

type fakeInternalSharingUploader struct {
	apkPath    string
	bundlePath string
	artifact   InternalSharingArtifact
}

func (u *fakeInternalSharingUploader) UploadInternalSharingAPK(ctx context.Context, packageName PackageName, path string) (InternalSharingArtifact, error) {
	u.apkPath = path
	return u.artifact, nil
}

func (u *fakeInternalSharingUploader) UploadInternalSharingBundle(ctx context.Context, packageName PackageName, path string) (InternalSharingArtifact, error) {
	u.bundlePath = path
	return u.artifact, nil
}

func writeTestFile(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte("artifact"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}
