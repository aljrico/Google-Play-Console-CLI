package play

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"

	"github.com/aljrico/Google-Play-Console-CLI/internal/googleclient"
)

func NewPublisherFromActiveProfile(ctx context.Context) (*GooglePublisher, error) {
	httpClient, err := googleclient.ActiveProfileHTTPClient(ctx, androidpublisher.AndroidpublisherScope)
	if err != nil {
		return nil, err
	}
	service, err := androidpublisher.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create Google Play API service: %w", err)
	}
	return &GooglePublisher{service: service, httpClient: httpClient, basePath: service.BasePath}, nil
}

type GooglePublisher struct {
	service    *androidpublisher.Service
	httpClient *http.Client
	basePath   string
}

func (p *GooglePublisher) doJSON(req *http.Request, target any) error {
	httpClient := p.httpClient
	if httpClient == nil {
		return fmt.Errorf("Google Play HTTP client is required")
	}
	response, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if err := googleapi.CheckResponse(response); err != nil {
		return err
	}
	decoder := json.NewDecoder(response.Body)
	if err := decoder.Decode(target); err != nil {
		return err
	}
	return nil
}

func (p *GooglePublisher) doNoContent(req *http.Request) error {
	httpClient := p.httpClient
	if httpClient == nil {
		return fmt.Errorf("Google Play HTTP client is required")
	}
	response, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if err := googleapi.CheckResponse(response); err != nil {
		return err
	}
	return nil
}

func (p *GooglePublisher) InsertEdit(ctx context.Context, packageName PackageName) (Edit, error) {
	edit, err := p.service.Edits.Insert(packageName.String(), &androidpublisher.AppEdit{}).Context(ctx).Do()
	if err != nil {
		return Edit{}, fmt.Errorf("insert edit for %s: %w", packageName, err)
	}
	return Edit{ID: edit.Id, ExpiryTimeSeconds: edit.ExpiryTimeSeconds}, nil
}

func (p *GooglePublisher) UploadBundle(ctx context.Context, packageName PackageName, editID string, bundlePath string) (BundleArtifact, error) {
	file, err := os.Open(bundlePath)
	if err != nil {
		return BundleArtifact{}, fmt.Errorf("open bundle %s: %w", bundlePath, err)
	}
	defer file.Close()

	bundle, err := p.service.Edits.Bundles.Upload(packageName.String(), editID).
		Media(file, googleapi.ContentType("application/octet-stream")).
		Context(ctx).
		Do()
	if err != nil {
		return BundleArtifact{}, fmt.Errorf("upload bundle %s for %s: %w", bundlePath, packageName, err)
	}
	return BundleArtifact{VersionCode: bundle.VersionCode, SHA1: bundle.Sha1, SHA256: bundle.Sha256}, nil
}

func (p *GooglePublisher) UploadAPK(ctx context.Context, packageName PackageName, editID string, apkPath string) (APKArtifact, error) {
	file, err := os.Open(apkPath)
	if err != nil {
		return APKArtifact{}, fmt.Errorf("open APK %s: %w", apkPath, err)
	}
	defer file.Close()

	apk, err := p.service.Edits.Apks.Upload(packageName.String(), editID).
		Media(file, googleapi.ContentType("application/octet-stream")).
		Context(ctx).
		Do()
	if err != nil {
		return APKArtifact{}, fmt.Errorf("upload APK %s for %s: %w", apkPath, packageName, err)
	}
	return apkArtifactFromAPI(apk), nil
}

func (p *GooglePublisher) ValidateEdit(ctx context.Context, packageName PackageName, editID string) error {
	if _, err := p.service.Edits.Validate(packageName.String(), editID).Context(ctx).Do(); err != nil {
		return fmt.Errorf("validate edit %s for %s: %w", editID, packageName, err)
	}
	return nil
}

type CommitEditOptions struct {
	ChangesNotSentForReview bool
}

func (p *GooglePublisher) CommitEdit(ctx context.Context, packageName PackageName, editID string) (Edit, error) {
	return p.CommitEditWithOptions(ctx, packageName, editID, CommitEditOptions{})
}

func (p *GooglePublisher) CommitEditWithOptions(ctx context.Context, packageName PackageName, editID string, opts CommitEditOptions) (Edit, error) {
	call := p.service.Edits.Commit(packageName.String(), editID).Context(ctx)
	if opts.ChangesNotSentForReview {
		call = call.ChangesNotSentForReview(true)
	}
	edit, err := call.Do()
	if err != nil {
		return Edit{}, fmt.Errorf("commit edit %s for %s: %w", editID, packageName, err)
	}
	return Edit{ID: edit.Id, ExpiryTimeSeconds: edit.ExpiryTimeSeconds}, nil
}

func (p *GooglePublisher) DeleteEdit(ctx context.Context, packageName PackageName, editID string) error {
	if err := p.service.Edits.Delete(packageName.String(), editID).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete edit %s for %s: %w", editID, packageName, err)
	}
	return nil
}

func isGoogleNotFound(err error) bool {
	apiError, ok := err.(*googleapi.Error)
	return ok && apiError.Code == http.StatusNotFound
}

func regionalPriceMigrationsToAPI(regions []BasePlanPriceMigrationConfig) []*androidpublisher.RegionalPriceMigrationConfig {
	apiRegions := make([]*androidpublisher.RegionalPriceMigrationConfig, 0, len(regions))
	for _, region := range regions {
		apiRegions = append(apiRegions, &androidpublisher.RegionalPriceMigrationConfig{
			RegionCode:                    region.RegionCode,
			OldestAllowedPriceVersionTime: region.OldestAllowedPriceVersionTime,
			PriceIncreaseType:             basePlanPriceIncreaseTypeToAPI(region.PriceIncreaseType),
		})
	}
	return apiRegions
}

func deviceMetadataFromAPI(apiMetadata *androidpublisher.DeviceMetadata) *DeviceMetadata {
	if apiMetadata == nil {
		return nil
	}
	return &DeviceMetadata{
		CPUMake:            apiMetadata.CpuMake,
		CPUModel:           apiMetadata.CpuModel,
		DeviceClass:        apiMetadata.DeviceClass,
		GLESVersion:        apiMetadata.GlEsVersion,
		Manufacturer:       apiMetadata.Manufacturer,
		NativePlatform:     apiMetadata.NativePlatform,
		ProductName:        apiMetadata.ProductName,
		RAMMegabytes:       apiMetadata.RamMb,
		ScreenDensityDPI:   apiMetadata.ScreenDensityDpi,
		ScreenHeightPixels: apiMetadata.ScreenHeightPx,
		ScreenWidthPixels:  apiMetadata.ScreenWidthPx,
	}
}

func timestampFromAPI(apiTimestamp *androidpublisher.Timestamp) *Timestamp {
	if apiTimestamp == nil {
		return nil
	}
	return &Timestamp{Seconds: apiTimestamp.Seconds, Nanos: apiTimestamp.Nanos}
}

func productPriceFromAPI(apiPrice *androidpublisher.Price) *ProductPrice {
	if apiPrice == nil {
		return nil
	}
	return &ProductPrice{Currency: apiPrice.Currency, PriceMicros: apiPrice.PriceMicros}
}

func productPriceToAPI(price ProductPrice) *androidpublisher.Price {
	return &androidpublisher.Price{Currency: price.Currency, PriceMicros: price.PriceMicros}
}

func productPricesToAPI(prices map[string]ProductPrice) map[string]androidpublisher.Price {
	if len(prices) == 0 {
		return nil
	}
	apiPrices := make(map[string]androidpublisher.Price, len(prices))
	for region, price := range prices {
		apiPrices[region] = androidpublisher.Price{Currency: price.Currency, PriceMicros: price.PriceMicros}
	}
	return apiPrices
}

func managedProductTaxComplianceSettingsToAPI(settings *ProductTaxComplianceSettings) *androidpublisher.ManagedProductTaxAndComplianceSettings {
	if settings == nil {
		return nil
	}
	apiSettings := &androidpublisher.ManagedProductTaxAndComplianceSettings{
		EeaWithdrawalRightType: settings.EEAWithdrawalRightType,
	}
	if settings.IsTokenizedDigitalAsset != nil {
		apiSettings.IsTokenizedDigitalAsset = *settings.IsTokenizedDigitalAsset
		apiSettings.ForceSendFields = append(apiSettings.ForceSendFields, "IsTokenizedDigitalAsset")
	}
	if len(settings.TaxRateInfoByRegionCode) > 0 {
		apiSettings.TaxRateInfoByRegionCode = regionalTaxRateInfoToAPI(settings.TaxRateInfoByRegionCode)
	}
	return apiSettings
}

func regionalTaxRateInfoToAPI(taxRateInfo map[string]RegionalTaxRateInfo) map[string]androidpublisher.RegionalTaxRateInfo {
	if len(taxRateInfo) == 0 {
		return nil
	}
	apiTaxRateInfo := make(map[string]androidpublisher.RegionalTaxRateInfo, len(taxRateInfo))
	for region, info := range taxRateInfo {
		apiTaxRateInfo[region] = androidpublisher.RegionalTaxRateInfo{
			EligibleForStreamingServiceTaxRate: info.EligibleForStreamingServiceTaxRate,
			StreamingTaxType:                   info.StreamingTaxType,
			TaxTier:                            info.TaxTier,
		}
	}
	return apiTaxRateInfo
}

func productPricesFromAPI(apiPrices map[string]androidpublisher.Price) map[string]ProductPrice {
	if len(apiPrices) == 0 {
		return nil
	}
	prices := make(map[string]ProductPrice, len(apiPrices))
	for region, apiPrice := range apiPrices {
		prices[region] = ProductPrice{Currency: apiPrice.Currency, PriceMicros: apiPrice.PriceMicros}
	}
	return prices
}

type rawOfferTag struct {
	Tag string `json:"tag,omitempty"`
}

type rawRegionalProductAgeRating struct {
	RegionCode           string `json:"regionCode,omitempty"`
	ProductAgeRatingTier string `json:"productAgeRatingTier,omitempty"`
}

type rawRegionalTaxConfig struct {
	RegionCode                         string `json:"regionCode,omitempty"`
	EligibleForStreamingServiceTaxRate bool   `json:"eligibleForStreamingServiceTaxRate,omitempty"`
	StreamingTaxType                   string `json:"streamingTaxType,omitempty"`
	TaxTier                            string `json:"taxTier,omitempty"`
}

type rawRestrictedPaymentCountries struct {
	RegionCodes []string `json:"regionCodes,omitempty"`
}

type rawRegionsVersion struct {
	Version string `json:"version,omitempty"`
}

type rawMoney struct {
	CurrencyCode string        `json:"currencyCode,omitempty"`
	Units        flexibleInt64 `json:"units,omitempty"`
	Nanos        int64         `json:"nanos,omitempty"`
}

type flexibleInt64 int64

func (i *flexibleInt64) UnmarshalJSON(data []byte) error {
	var stringValue string
	if err := json.Unmarshal(data, &stringValue); err == nil {
		parsed, err := strconv.ParseInt(stringValue, 10, 64)
		if err != nil {
			return err
		}
		*i = flexibleInt64(parsed)
		return nil
	}
	var numberValue int64
	if err := json.Unmarshal(data, &numberValue); err != nil {
		return err
	}
	*i = flexibleInt64(numberValue)
	return nil
}

func productUpdateLatencyToleranceToAPI(latencyTolerance ProductUpdateLatencyTolerance) string {
	switch latencyTolerance {
	case ProductUpdateLatencyToleranceTolerant:
		return "PRODUCT_UPDATE_LATENCY_TOLERANCE_LATENCY_TOLERANT"
	default:
		return "PRODUCT_UPDATE_LATENCY_TOLERANCE_LATENCY_SENSITIVE"
	}
}

func offerTagsToAPI(tags []string) []*androidpublisher.OfferTag {
	apiTags := make([]*androidpublisher.OfferTag, 0, len(tags))
	for _, tag := range tags {
		apiTags = append(apiTags, &androidpublisher.OfferTag{Tag: tag})
	}
	return apiTags
}

func regionalAgeRatingsFromAPI(apiRatings []rawRegionalProductAgeRating) []RegionalAgeRating {
	if len(apiRatings) == 0 {
		return nil
	}
	ratings := make([]RegionalAgeRating, 0, len(apiRatings))
	for _, apiRating := range apiRatings {
		ratings = append(ratings, RegionalAgeRating{
			RegionCode:           apiRating.RegionCode,
			ProductAgeRatingTier: apiRating.ProductAgeRatingTier,
		})
	}
	return ratings
}

func regionalTaxConfigsFromAPI(apiConfigs []rawRegionalTaxConfig) []RegionalTaxConfig {
	if len(apiConfigs) == 0 {
		return nil
	}
	configs := make([]RegionalTaxConfig, 0, len(apiConfigs))
	for _, apiConfig := range apiConfigs {
		configs = append(configs, RegionalTaxConfig{
			RegionCode:                         apiConfig.RegionCode,
			EligibleForStreamingServiceTaxRate: apiConfig.EligibleForStreamingServiceTaxRate,
			StreamingTaxType:                   apiConfig.StreamingTaxType,
			TaxTier:                            apiConfig.TaxTier,
		})
	}
	return configs
}

func regionsVersionFromAPI(apiVersion *rawRegionsVersion) *RegionsVersion {
	if apiVersion == nil {
		return nil
	}
	return &RegionsVersion{Version: apiVersion.Version}
}

func moneyToAPI(money *Money) *androidpublisher.Money {
	if money == nil {
		return nil
	}
	return &androidpublisher.Money{
		CurrencyCode: money.CurrencyCode,
		Units:        money.Units,
		Nanos:        money.Nanos,
	}
}

func regionsVersionFromGeneratedAPI(apiVersion *androidpublisher.RegionsVersion) *RegionsVersion {
	if apiVersion == nil {
		return nil
	}
	return &RegionsVersion{Version: apiVersion.Version}
}

func managedProductTaxComplianceSettingsFromAPI(apiSettings *androidpublisher.ManagedProductTaxAndComplianceSettings) *ProductTaxComplianceSettings {
	if apiSettings == nil {
		return nil
	}
	tokenizedDigitalAsset := apiSettings.IsTokenizedDigitalAsset
	return &ProductTaxComplianceSettings{
		EEAWithdrawalRightType:  apiSettings.EeaWithdrawalRightType,
		IsTokenizedDigitalAsset: &tokenizedDigitalAsset,
		TaxRateInfoByRegionCode: regionalTaxRateInfoFromAPI(apiSettings.TaxRateInfoByRegionCode),
	}
}

func regionalTaxRateInfoFromAPI(apiTaxRateInfo map[string]androidpublisher.RegionalTaxRateInfo) map[string]RegionalTaxRateInfo {
	if len(apiTaxRateInfo) == 0 {
		return nil
	}
	taxRateInfo := make(map[string]RegionalTaxRateInfo, len(apiTaxRateInfo))
	for region, apiInfo := range apiTaxRateInfo {
		taxRateInfo[region] = RegionalTaxRateInfo{
			EligibleForStreamingServiceTaxRate: apiInfo.EligibleForStreamingServiceTaxRate,
			StreamingTaxType:                   apiInfo.StreamingTaxType,
			TaxTier:                            apiInfo.TaxTier,
		}
	}
	return taxRateInfo
}

func moneyFromAPI(apiMoney *androidpublisher.Money) *Money {
	if apiMoney == nil {
		return nil
	}
	return &Money{
		CurrencyCode: apiMoney.CurrencyCode,
		Units:        apiMoney.Units,
		Nanos:        apiMoney.Nanos,
	}
}

func rawMoneyFromAPI(apiMoney *rawMoney) *Money {
	if apiMoney == nil {
		return nil
	}
	return &Money{
		CurrencyCode: apiMoney.CurrencyCode,
		Units:        int64(apiMoney.Units),
		Nanos:        apiMoney.Nanos,
	}
}

func offerTagsFromAPI(apiOfferTags []*androidpublisher.OfferTag) []string {
	if len(apiOfferTags) == 0 {
		return nil
	}
	tags := make([]string, 0, len(apiOfferTags))
	for _, apiOfferTag := range apiOfferTags {
		if apiOfferTag == nil {
			continue
		}
		tags = append(tags, apiOfferTag.Tag)
	}
	return tags
}

func rawOfferTagsFromAPI(apiOfferTags []rawOfferTag) []string {
	if len(apiOfferTags) == 0 {
		return nil
	}
	tags := make([]string, 0, len(apiOfferTags))
	for _, apiOfferTag := range apiOfferTags {
		tags = append(tags, apiOfferTag.Tag)
	}
	return tags
}

func restrictedCountriesFromAPI(apiCountries *androidpublisher.RestrictedPaymentCountries) []string {
	if apiCountries == nil {
		return nil
	}
	return apiCountries.RegionCodes
}

func restrictedCountriesToAPI(countries []string) *androidpublisher.RestrictedPaymentCountries {
	if len(countries) == 0 {
		return nil
	}
	return &androidpublisher.RestrictedPaymentCountries{RegionCodes: append([]string(nil), countries...)}
}

func rawRestrictedCountriesFromAPI(apiCountries *rawRestrictedPaymentCountries) []string {
	if apiCountries == nil {
		return nil
	}
	return apiCountries.RegionCodes
}

func apkArtifactFromAPI(apiAPK *androidpublisher.Apk) APKArtifact {
	if apiAPK == nil {
		return APKArtifact{}
	}
	artifact := APKArtifact{VersionCode: apiAPK.VersionCode}
	if apiAPK.Binary != nil {
		artifact.SHA1 = apiAPK.Binary.Sha1
		artifact.SHA256 = apiAPK.Binary.Sha256
	}
	return artifact
}

func stringPointer(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
