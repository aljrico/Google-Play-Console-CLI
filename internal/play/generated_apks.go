package play

import (
	"context"
	"fmt"
)

type GeneratedAPKListOptions struct {
	PackageName PackageName `json:"packageName"`
	VersionCode int64       `json:"versionCode"`
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
	VariantNumber int64    `json:"variantNumber"`
	ModuleNames   []string `json:"moduleNames,omitempty"`
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

func ListGeneratedAPKs(ctx context.Context, lister GeneratedAPKLister, options GeneratedAPKListOptions) (GeneratedAPKListResult, error) {
	if err := options.Validate(); err != nil {
		return GeneratedAPKListResult{}, err
	}
	if lister == nil {
		return GeneratedAPKListResult{}, fmt.Errorf("generated APK lister is required")
	}
	return lister.ListGeneratedAPKs(ctx, options)
}
