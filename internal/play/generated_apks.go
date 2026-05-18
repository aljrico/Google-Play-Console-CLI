package play

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type GeneratedAPKListOptions struct {
	PackageName PackageName `json:"packageName"`
	VersionCode int64       `json:"versionCode"`
}

type GeneratedAPKDownloadID string

func NewGeneratedAPKDownloadID(value string) (GeneratedAPKDownloadID, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("download ID is required")
	}
	if strings.Contains(value, "/") {
		return "", fmt.Errorf("download ID cannot contain /")
	}
	return GeneratedAPKDownloadID(value), nil
}

func (d GeneratedAPKDownloadID) String() string {
	return string(d)
}

func (d GeneratedAPKDownloadID) Validate() error {
	_, err := NewGeneratedAPKDownloadID(d.String())
	return err
}

func (o GeneratedAPKListOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if o.VersionCode <= 0 {
		return fmt.Errorf("version code must be greater than 0")
	}
	return nil
}

type GeneratedAPKSigningKey struct {
	CertificateSHA256Hash string                     `json:"certificateSha256Hash,omitempty"`
	SplitAPKs             []GeneratedSplitAPK        `json:"generatedSplitApks"`
	StandaloneAPKs        []GeneratedStandaloneAPK   `json:"generatedStandaloneApks"`
	AssetPackSlices       []GeneratedAssetPackSlice  `json:"generatedAssetPackSlices"`
	RecoveryModules       []GeneratedRecoveryAPK     `json:"generatedRecoveryModules"`
	UniversalAPK          *GeneratedUniversalAPK     `json:"generatedUniversalApk,omitempty"`
	TargetingPackageName  string                     `json:"targetingPackageName,omitempty"`
	TargetingInfo         *GeneratedAPKTargetingInfo `json:"targetingInfo,omitempty"`
}

type GeneratedAPKTargetingInfo struct {
	PackageName    string                         `json:"packageName,omitempty"`
	Variants       []GeneratedAPKTargetingVariant `json:"variant"`
	AssetSliceSets []GeneratedAPKAssetSliceSet    `json:"assetSliceSet"`
}

type GeneratedAPKTargetingVariant struct {
	VariantNumber int64                     `json:"variantNumber"`
	ModuleNames   []string                  `json:"moduleNames,omitempty"`
	Targeting     json.RawMessage           `json:"targeting,omitempty"`
	APKs          []GeneratedAPKTargetedAPK `json:"apks"`
}

type GeneratedAPKTargetedAPK struct {
	Path      string          `json:"path,omitempty"`
	Targeting json.RawMessage `json:"targeting,omitempty"`
}

type GeneratedAPKAssetSliceSet struct {
	ModuleName   string   `json:"moduleName,omitempty"`
	DeliveryType string   `json:"deliveryType,omitempty"`
	APKPaths     []string `json:"apkPaths,omitempty"`
}

type GeneratedSplitAPK struct {
	DownloadID string `json:"downloadId,omitempty"`
	ModuleName string `json:"moduleName,omitempty"`
	SplitID    string `json:"splitId,omitempty"`
	VariantID  int64  `json:"variantId"`
}

type GeneratedStandaloneAPK struct {
	DownloadID string `json:"downloadId,omitempty"`
	VariantID  int64  `json:"variantId"`
}

type GeneratedUniversalAPK struct {
	DownloadID string `json:"downloadId,omitempty"`
}

type GeneratedAssetPackSlice struct {
	DownloadID string `json:"downloadId,omitempty"`
	ModuleName string `json:"moduleName,omitempty"`
	SliceID    string `json:"sliceId,omitempty"`
	Version    int64  `json:"version"`
}

type GeneratedRecoveryAPK struct {
	DownloadID     string `json:"downloadId,omitempty"`
	ModuleName     string `json:"moduleName,omitempty"`
	RecoveryID     string `json:"recoveryId,omitempty"`
	RecoveryStatus string `json:"recoveryStatus,omitempty"`
}

type GeneratedAPKListResult struct {
	PackageName PackageName              `json:"packageName"`
	VersionCode int64                    `json:"versionCode"`
	SigningKeys []GeneratedAPKSigningKey `json:"generatedApks"`
}

type GeneratedAPKLister interface {
	ListGeneratedAPKs(ctx context.Context, options GeneratedAPKListOptions) (GeneratedAPKListResult, error)
}

type GeneratedAPKDownloader interface {
	DownloadGeneratedAPK(ctx context.Context, options GeneratedAPKDownloadOptions) (GeneratedAPKDownloadResult, error)
}

func ListGeneratedAPKs(ctx context.Context, lister GeneratedAPKLister, options GeneratedAPKListOptions) (GeneratedAPKListResult, error) {
	if err := options.Validate(); err != nil {
		return GeneratedAPKListResult{}, err
	}
	if lister == nil {
		return GeneratedAPKListResult{}, fmt.Errorf("generated APK lister is required")
	}
	return lister.ListGeneratedAPKs(ctx, options)
}

type GeneratedAPKDownloadOptions struct {
	PackageName PackageName            `json:"packageName"`
	VersionCode int64                  `json:"versionCode"`
	DownloadID  GeneratedAPKDownloadID `json:"downloadId"`
	OutputPath  string                 `json:"outputPath"`
	Force       bool                   `json:"force"`
	DryRun      bool                   `json:"dryRun"`
}

func (o GeneratedAPKDownloadOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if o.VersionCode <= 0 {
		return fmt.Errorf("version code must be greater than 0")
	}
	if err := o.DownloadID.Validate(); err != nil {
		return err
	}
	if o.OutputPath == "" {
		return fmt.Errorf("output path is required")
	}
	if filepath.Ext(o.OutputPath) != ".apk" {
		return fmt.Errorf("output path must end with .apk")
	}
	return nil
}

type GeneratedAPKDownloadPlan struct {
	PackageName PackageName            `json:"packageName"`
	VersionCode int64                  `json:"versionCode"`
	DownloadID  GeneratedAPKDownloadID `json:"downloadId"`
	OutputPath  string                 `json:"outputPath"`
	Force       bool                   `json:"force"`
	Steps       []string               `json:"steps"`
}

type GeneratedAPKDownloadResult struct {
	PackageName  PackageName              `json:"packageName"`
	VersionCode  int64                    `json:"versionCode"`
	DownloadID   GeneratedAPKDownloadID   `json:"downloadId"`
	OutputPath   string                   `json:"outputPath"`
	DryRun       bool                     `json:"dryRun"`
	Downloaded   bool                     `json:"downloaded"`
	BytesWritten int64                    `json:"bytesWritten,omitempty"`
	Plan         GeneratedAPKDownloadPlan `json:"plan"`
}

func NewGeneratedAPKDownloadPlan(options GeneratedAPKDownloadOptions) (GeneratedAPKDownloadPlan, error) {
	if err := options.Validate(); err != nil {
		return GeneratedAPKDownloadPlan{}, err
	}
	steps := []string{"download generated APK"}
	if options.Force {
		steps = append(steps, "overwrite output file")
	} else {
		steps = append(steps, "create output file")
	}
	return GeneratedAPKDownloadPlan{
		PackageName: options.PackageName,
		VersionCode: options.VersionCode,
		DownloadID:  options.DownloadID,
		OutputPath:  options.OutputPath,
		Force:       options.Force,
		Steps:       steps,
	}, nil
}

func DownloadGeneratedAPK(ctx context.Context, downloader GeneratedAPKDownloader, options GeneratedAPKDownloadOptions) (GeneratedAPKDownloadResult, error) {
	plan, err := NewGeneratedAPKDownloadPlan(options)
	if err != nil {
		return GeneratedAPKDownloadResult{}, err
	}
	result := GeneratedAPKDownloadResult{
		PackageName: options.PackageName,
		VersionCode: options.VersionCode,
		DownloadID:  options.DownloadID,
		OutputPath:  options.OutputPath,
		DryRun:      options.DryRun,
		Downloaded:  false,
		Plan:        plan,
	}
	if err := ValidateGeneratedAPKOutputPath(options.OutputPath, options.Force); err != nil {
		return GeneratedAPKDownloadResult{}, err
	}
	if options.DryRun {
		return result, nil
	}
	if downloader == nil {
		return GeneratedAPKDownloadResult{}, fmt.Errorf("generated APK downloader is required")
	}
	return downloader.DownloadGeneratedAPK(ctx, options)
}

func ValidateGeneratedAPKOutputPath(path string, force bool) error {
	if path == "" {
		return fmt.Errorf("output path is required")
	}
	if filepath.Ext(path) != ".apk" {
		return fmt.Errorf("output path must end with .apk")
	}
	info, err := os.Lstat(path)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("output path %s is a symlink", path)
		}
		if info.IsDir() {
			return fmt.Errorf("output path %s is a directory", path)
		}
		if !force {
			return fmt.Errorf("output path %s already exists; pass --force to overwrite", path)
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("inspect output path %s: %w", path, err)
	}
	parent := filepath.Dir(path)
	parentInfo, err := os.Stat(parent)
	if err != nil {
		return fmt.Errorf("inspect output directory %s: %w", parent, err)
	}
	if !parentInfo.IsDir() {
		return fmt.Errorf("output directory %s is not a directory", parent)
	}
	return nil
}
