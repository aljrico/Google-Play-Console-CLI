package play

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"google.golang.org/api/androidpublisher/v3"
)

func (p *GooglePublisher) ListGeneratedAPKs(ctx context.Context, options GeneratedAPKListOptions) (GeneratedAPKListResult, error) {
	response, err := p.service.Generatedapks.List(options.PackageName.String(), options.VersionCode).
		Context(ctx).
		Do()
	if err != nil {
		return GeneratedAPKListResult{}, fmt.Errorf("list generated APKs for %s version code %d: %w", options.PackageName, options.VersionCode, err)
	}
	return generatedAPKListResultFromAPI(options, response), nil
}

func (p *GooglePublisher) DownloadGeneratedAPK(ctx context.Context, options GeneratedAPKDownloadOptions) (GeneratedAPKDownloadResult, error) {
	if err := options.ValidateLive(); err != nil {
		return GeneratedAPKDownloadResult{}, err
	}
	if err := ValidateGeneratedAPKOutputPath(options.OutputPath, options.Force); err != nil {
		return GeneratedAPKDownloadResult{}, err
	}
	response, err := p.service.Generatedapks.Download(options.PackageName.String(), options.VersionCode, options.DownloadID.String()).
		Context(ctx).
		Download()
	if err != nil {
		return GeneratedAPKDownloadResult{}, fmt.Errorf("download generated APK %s for %s version code %d: %w", options.DownloadID, options.PackageName, options.VersionCode, err)
	}
	defer response.Body.Close()

	tempFile, err := createGeneratedAPKTempFile(options.OutputPath)
	if err != nil {
		return GeneratedAPKDownloadResult{}, err
	}
	tempPath := tempFile.Name()
	committed := false
	defer func() {
		if !committed {
			_ = os.Remove(tempPath)
		}
	}()

	bytesWritten, err := io.Copy(tempFile, response.Body)
	if err != nil {
		_ = tempFile.Close()
		return GeneratedAPKDownloadResult{}, fmt.Errorf("write temporary APK %s: %w", tempPath, err)
	}
	if err := tempFile.Close(); err != nil {
		return GeneratedAPKDownloadResult{}, fmt.Errorf("close temporary APK %s: %w", tempPath, err)
	}
	if err := publishGeneratedAPKTempFile(tempPath, options.OutputPath, options.Force); err != nil {
		return GeneratedAPKDownloadResult{}, err
	}
	committed = true
	plan, err := NewGeneratedAPKDownloadPlan(options)
	if err != nil {
		return GeneratedAPKDownloadResult{}, err
	}
	return GeneratedAPKDownloadResult{
		PackageName:  options.PackageName,
		VersionCode:  options.VersionCode,
		DownloadID:   options.DownloadID,
		OutputPath:   options.OutputPath,
		DryRun:       false,
		Downloaded:   true,
		BytesWritten: bytesWritten,
		Plan:         plan,
	}, nil
}

func createGeneratedAPKTempFile(outputPath string) (*os.File, error) {
	outputDirectory := filepath.Dir(outputPath)
	outputBase := filepath.Base(outputPath)
	file, err := os.CreateTemp(outputDirectory, "."+outputBase+".tmp-*")
	if err != nil {
		return nil, fmt.Errorf("create temporary APK in %s: %w", outputDirectory, err)
	}
	return file, nil
}

func publishGeneratedAPKTempFile(tempPath string, outputPath string, force bool) error {
	if force {
		if err := os.Rename(tempPath, outputPath); err != nil {
			return fmt.Errorf("replace output APK %s: %w", outputPath, err)
		}
		return nil
	}
	outputFile, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("create output APK %s: %w", outputPath, err)
	}
	outputCommitted := false
	defer func() {
		_ = outputFile.Close()
		if !outputCommitted {
			_ = os.Remove(outputPath)
		}
	}()
	tempFile, err := os.Open(tempPath)
	if err != nil {
		return fmt.Errorf("open temporary APK %s: %w", tempPath, err)
	}
	defer tempFile.Close()
	if _, err := io.Copy(outputFile, tempFile); err != nil {
		return fmt.Errorf("copy temporary APK %s to %s: %w", tempPath, outputPath, err)
	}
	if err := outputFile.Close(); err != nil {
		return fmt.Errorf("close output APK %s: %w", outputPath, err)
	}
	outputCommitted = true
	if err := os.Remove(tempPath); err != nil {
		return fmt.Errorf("remove temporary APK %s: %w", tempPath, err)
	}
	return nil
}

func generatedAPKListResultFromAPI(options GeneratedAPKListOptions, response *androidpublisher.GeneratedApksListResponse) GeneratedAPKListResult {
	result := GeneratedAPKListResult{
		PackageName: options.PackageName,
		VersionCode: options.VersionCode,
		SigningKeys: []GeneratedAPKSigningKey{},
	}
	if response == nil {
		return result
	}
	for _, apiSigningKey := range response.GeneratedApks {
		if apiSigningKey == nil {
			continue
		}
		result.SigningKeys = append(result.SigningKeys, generatedAPKSigningKeyFromAPI(apiSigningKey))
	}
	return result
}

func generatedAPKSigningKeyFromAPI(apiSigningKey *androidpublisher.GeneratedApksPerSigningKey) GeneratedAPKSigningKey {
	signingKey := GeneratedAPKSigningKey{
		CertificateSHA256Hash: apiSigningKey.CertificateSha256Hash,
		SplitAPKs:             []GeneratedSplitAPK{},
		StandaloneAPKs:        []GeneratedStandaloneAPK{},
		AssetPackSlices:       []GeneratedAssetPackSlice{},
		RecoveryModules:       []GeneratedRecoveryAPK{},
		UniversalAPK:          generatedUniversalAPKFromAPI(apiSigningKey.GeneratedUniversalApk),
	}
	if apiSigningKey.TargetingInfo != nil {
		signingKey.TargetingPackageName = apiSigningKey.TargetingInfo.PackageName
		signingKey.TargetingInfo = generatedAPKTargetingInfoFromAPI(apiSigningKey.TargetingInfo)
	}
	for _, apiAPK := range apiSigningKey.GeneratedSplitApks {
		if apiAPK == nil {
			continue
		}
		signingKey.SplitAPKs = append(signingKey.SplitAPKs, GeneratedSplitAPK{
			DownloadID: apiAPK.DownloadId,
			ModuleName: apiAPK.ModuleName,
			SplitID:    apiAPK.SplitId,
			VariantID:  apiAPK.VariantId,
		})
	}
	for _, apiAPK := range apiSigningKey.GeneratedStandaloneApks {
		if apiAPK == nil {
			continue
		}
		signingKey.StandaloneAPKs = append(signingKey.StandaloneAPKs, GeneratedStandaloneAPK{
			DownloadID: apiAPK.DownloadId,
			VariantID:  apiAPK.VariantId,
		})
	}
	for _, apiSlice := range apiSigningKey.GeneratedAssetPackSlices {
		if apiSlice == nil {
			continue
		}
		signingKey.AssetPackSlices = append(signingKey.AssetPackSlices, GeneratedAssetPackSlice{
			DownloadID: apiSlice.DownloadId,
			ModuleName: apiSlice.ModuleName,
			SliceID:    apiSlice.SliceId,
			Version:    apiSlice.Version,
		})
	}
	for _, apiRecoveryAPK := range apiSigningKey.GeneratedRecoveryModules {
		if apiRecoveryAPK == nil {
			continue
		}
		signingKey.RecoveryModules = append(signingKey.RecoveryModules, GeneratedRecoveryAPK{
			DownloadID:     apiRecoveryAPK.DownloadId,
			ModuleName:     apiRecoveryAPK.ModuleName,
			RecoveryID:     strconv.FormatInt(apiRecoveryAPK.RecoveryId, 10),
			RecoveryStatus: apiRecoveryAPK.RecoveryStatus,
		})
	}
	return signingKey
}

func generatedUniversalAPKFromAPI(apiAPK *androidpublisher.GeneratedUniversalApk) *GeneratedUniversalAPK {
	if apiAPK == nil {
		return nil
	}
	return &GeneratedUniversalAPK{DownloadID: apiAPK.DownloadId}
}

func generatedAPKTargetingInfoFromAPI(apiTargetingInfo *androidpublisher.TargetingInfo) *GeneratedAPKTargetingInfo {
	if apiTargetingInfo == nil {
		return nil
	}
	targetingInfo := &GeneratedAPKTargetingInfo{
		PackageName:    apiTargetingInfo.PackageName,
		Variants:       []GeneratedAPKTargetingVariant{},
		AssetSliceSets: []GeneratedAPKAssetSliceSet{},
	}
	for _, apiVariant := range apiTargetingInfo.Variant {
		if apiVariant == nil {
			continue
		}
		targetingInfo.Variants = append(targetingInfo.Variants, generatedAPKTargetingVariantFromAPI(apiVariant))
	}
	for _, apiAssetSliceSet := range apiTargetingInfo.AssetSliceSet {
		if apiAssetSliceSet == nil {
			continue
		}
		targetingInfo.AssetSliceSets = append(targetingInfo.AssetSliceSets, generatedAPKAssetSliceSetFromAPI(apiAssetSliceSet))
	}
	return targetingInfo
}

func generatedAPKTargetingVariantFromAPI(apiVariant *androidpublisher.SplitApkVariant) GeneratedAPKTargetingVariant {
	moduleNames := make([]string, 0, len(apiVariant.ApkSet))
	targetedAPKs := []GeneratedAPKTargetedAPK{}
	for _, apiAPKSet := range apiVariant.ApkSet {
		if apiAPKSet == nil || apiAPKSet.ModuleMetadata == nil {
			if apiAPKSet != nil {
				targetedAPKs = append(targetedAPKs, generatedAPKTargetedAPKsFromAPI(apiAPKSet.ApkDescription)...)
			}
			continue
		}
		moduleNames = append(moduleNames, apiAPKSet.ModuleMetadata.Name)
		targetedAPKs = append(targetedAPKs, generatedAPKTargetedAPKsFromAPI(apiAPKSet.ApkDescription)...)
	}
	return GeneratedAPKTargetingVariant{
		VariantNumber: apiVariant.VariantNumber,
		ModuleNames:   moduleNames,
		Targeting:     generatedAPKRawJSONFromAPI(apiVariant.Targeting),
		APKs:          targetedAPKs,
	}
}

func generatedAPKTargetedAPKsFromAPI(apiAPKDescriptions []*androidpublisher.ApkDescription) []GeneratedAPKTargetedAPK {
	targetedAPKs := make([]GeneratedAPKTargetedAPK, 0, len(apiAPKDescriptions))
	for _, apiAPKDescription := range apiAPKDescriptions {
		if apiAPKDescription == nil {
			continue
		}
		targetedAPKs = append(targetedAPKs, GeneratedAPKTargetedAPK{
			Path:      apiAPKDescription.Path,
			Targeting: generatedAPKRawJSONFromAPI(apiAPKDescription.Targeting),
		})
	}
	return targetedAPKs
}

func generatedAPKAssetSliceSetFromAPI(apiAssetSliceSet *androidpublisher.AssetSliceSet) GeneratedAPKAssetSliceSet {
	assetSliceSet := GeneratedAPKAssetSliceSet{APKPaths: []string{}}
	if apiAssetSliceSet.AssetModuleMetadata != nil {
		assetSliceSet.ModuleName = apiAssetSliceSet.AssetModuleMetadata.Name
		assetSliceSet.DeliveryType = apiAssetSliceSet.AssetModuleMetadata.DeliveryType
	}
	for _, apiAPKDescription := range apiAssetSliceSet.ApkDescription {
		if apiAPKDescription == nil || apiAPKDescription.Path == "" {
			continue
		}
		assetSliceSet.APKPaths = append(assetSliceSet.APKPaths, apiAPKDescription.Path)
	}
	return assetSliceSet
}

func generatedAPKRawJSONFromAPI(value any) json.RawMessage {
	if value == nil {
		return nil
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	if string(payload) == "{}" {
		return nil
	}
	return payload
}
