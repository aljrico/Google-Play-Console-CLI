package play

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type InternalSharingArtifactKind string

const (
	InternalSharingArtifactKindAPK    InternalSharingArtifactKind = "apk"
	InternalSharingArtifactKindBundle InternalSharingArtifactKind = "bundle"
)

type InternalSharingUploadOptions struct {
	PackageName PackageName `json:"packageName"`
	APKPath     string      `json:"apkPath,omitempty"`
	BundlePath  string      `json:"bundlePath,omitempty"`
	DryRun      bool        `json:"dryRun"`
}

func (o InternalSharingUploadOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if (o.APKPath == "") == (o.BundlePath == "") {
		return fmt.Errorf("exactly one of APK path or AAB path is required")
	}
	if o.APKPath != "" && filepath.Ext(o.APKPath) != ".apk" {
		return fmt.Errorf("APK path must end with .apk")
	}
	if o.BundlePath != "" && filepath.Ext(o.BundlePath) != ".aab" {
		return fmt.Errorf("AAB path must end with .aab")
	}
	return nil
}

func (o InternalSharingUploadOptions) artifactPath() string {
	if o.APKPath != "" {
		return o.APKPath
	}
	return o.BundlePath
}

func (o InternalSharingUploadOptions) artifactKind() InternalSharingArtifactKind {
	if o.APKPath != "" {
		return InternalSharingArtifactKindAPK
	}
	return InternalSharingArtifactKindBundle
}

type InternalSharingArtifact struct {
	CertificateFingerprint string `json:"certificateFingerprint,omitempty"`
	DownloadURL            string `json:"downloadUrl,omitempty"`
	SHA256                 string `json:"sha256,omitempty"`
}

type InternalSharingUploadResult struct {
	PackageName PackageName                 `json:"packageName"`
	Kind        InternalSharingArtifactKind `json:"kind"`
	Path        string                      `json:"path"`
	DryRun      bool                        `json:"dryRun"`
	Artifact    *InternalSharingArtifact    `json:"artifact,omitempty"`
}

type InternalSharingUploader interface {
	UploadInternalSharingAPK(ctx context.Context, packageName PackageName, path string) (InternalSharingArtifact, error)
	UploadInternalSharingBundle(ctx context.Context, packageName PackageName, path string) (InternalSharingArtifact, error)
}

func UploadInternalSharingArtifact(ctx context.Context, uploader InternalSharingUploader, options InternalSharingUploadOptions) (InternalSharingUploadResult, error) {
	if err := options.Validate(); err != nil {
		return InternalSharingUploadResult{}, err
	}
	result := InternalSharingUploadResult{
		PackageName: options.PackageName,
		Kind:        options.artifactKind(),
		Path:        options.artifactPath(),
		DryRun:      options.DryRun,
	}
	if options.DryRun {
		return result, nil
	}
	if err := ValidateReadableFile(options.artifactPath()); err != nil {
		return InternalSharingUploadResult{}, err
	}
	if uploader == nil {
		return InternalSharingUploadResult{}, fmt.Errorf("internal sharing uploader is required")
	}
	var (
		artifact InternalSharingArtifact
		err      error
	)
	switch options.artifactKind() {
	case InternalSharingArtifactKindAPK:
		artifact, err = uploader.UploadInternalSharingAPK(ctx, options.PackageName, options.APKPath)
	case InternalSharingArtifactKindBundle:
		artifact, err = uploader.UploadInternalSharingBundle(ctx, options.PackageName, options.BundlePath)
	default:
		err = fmt.Errorf("unsupported internal sharing artifact kind %q", options.artifactKind())
	}
	if err != nil {
		return InternalSharingUploadResult{}, err
	}
	result.Artifact = &artifact
	return result, nil
}

func ValidateReadableFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open file %s: %w", path, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close file %s: %w", path, err)
	}
	return nil
}
