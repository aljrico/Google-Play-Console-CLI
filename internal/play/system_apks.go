package play

import (
	"context"
	"fmt"
)

type SystemAPKVariantListOptions struct {
	PackageName PackageName `json:"packageName"`
	VersionCode int64       `json:"versionCode"`
}

func (o SystemAPKVariantListOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if o.VersionCode <= 0 {
		return fmt.Errorf("version code must be greater than 0")
	}
	return nil
}

type SystemAPKVariant struct {
	VariantID  int64                `json:"variantId"`
	DeviceSpec *SystemAPKDeviceSpec `json:"deviceSpec,omitempty"`
	Options    *SystemAPKOptions    `json:"options,omitempty"`
}

type SystemAPKDeviceSpec struct {
	ScreenDensity    int64    `json:"screenDensity,omitempty"`
	SupportedABIs    []string `json:"supportedAbis,omitempty"`
	SupportedLocales []string `json:"supportedLocales,omitempty"`
}

type SystemAPKOptions struct {
	Rotated                     bool `json:"rotated,omitempty"`
	UncompressedDexFiles        bool `json:"uncompressedDexFiles,omitempty"`
	UncompressedNativeLibraries bool `json:"uncompressedNativeLibraries,omitempty"`
}

type SystemAPKVariantListResult struct {
	PackageName PackageName        `json:"packageName"`
	VersionCode int64              `json:"versionCode"`
	Variants    []SystemAPKVariant `json:"variants"`
}

type SystemAPKVariantLister interface {
	ListSystemAPKVariants(ctx context.Context, options SystemAPKVariantListOptions) (SystemAPKVariantListResult, error)
}

func ListSystemAPKVariants(ctx context.Context, lister SystemAPKVariantLister, options SystemAPKVariantListOptions) (SystemAPKVariantListResult, error) {
	if err := options.Validate(); err != nil {
		return SystemAPKVariantListResult{}, err
	}
	if lister == nil {
		return SystemAPKVariantListResult{}, fmt.Errorf("system APK variant lister is required")
	}
	return lister.ListSystemAPKVariants(ctx, options)
}
