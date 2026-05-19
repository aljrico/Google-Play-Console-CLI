package play

import (
	"context"
	"fmt"

	"google.golang.org/api/androidpublisher/v3"
)

func (p *GooglePublisher) ListSystemAPKVariants(ctx context.Context, options SystemAPKVariantListOptions) (SystemAPKVariantListResult, error) {
	response, err := p.service.Systemapks.Variants.List(options.PackageName.String(), options.VersionCode).
		Context(ctx).
		Do()
	if err != nil {
		return SystemAPKVariantListResult{}, fmt.Errorf("list system APK variants for %s version code %d: %w", options.PackageName, options.VersionCode, err)
	}
	return systemAPKVariantListResultFromAPI(options, response), nil
}

func systemAPKVariantListResultFromAPI(options SystemAPKVariantListOptions, response *androidpublisher.SystemApksListResponse) SystemAPKVariantListResult {
	result := SystemAPKVariantListResult{
		PackageName: options.PackageName,
		VersionCode: options.VersionCode,
		Variants:    []SystemAPKVariant{},
	}
	if response == nil {
		return result
	}
	for _, apiVariant := range response.Variants {
		if apiVariant == nil {
			continue
		}
		result.Variants = append(result.Variants, SystemAPKVariant{
			VariantID:  apiVariant.VariantId,
			DeviceSpec: systemAPKDeviceSpecFromAPI(apiVariant.DeviceSpec),
			Options:    systemAPKOptionsFromAPI(apiVariant.Options),
		})
	}
	return result
}

func systemAPKDeviceSpecFromAPI(apiSpec *androidpublisher.DeviceSpec) *SystemAPKDeviceSpec {
	if apiSpec == nil {
		return nil
	}
	return &SystemAPKDeviceSpec{
		ScreenDensity:    apiSpec.ScreenDensity,
		SupportedABIs:    append([]string(nil), apiSpec.SupportedAbis...),
		SupportedLocales: append([]string(nil), apiSpec.SupportedLocales...),
	}
}

func systemAPKOptionsFromAPI(apiOptions *androidpublisher.SystemApkOptions) *SystemAPKOptions {
	if apiOptions == nil {
		return nil
	}
	return &SystemAPKOptions{
		Rotated:                     apiOptions.Rotated,
		UncompressedDexFiles:        apiOptions.UncompressedDexFiles,
		UncompressedNativeLibraries: apiOptions.UncompressedNativeLibraries,
	}
}
