package play

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"
)

type OneTimeProductID string

func NewOneTimeProductID(value string) (OneTimeProductID, error) {
	if value == "" {
		return "", fmt.Errorf("one-time product ID is required")
	}
	if !isValidOneTimeProductID(value) {
		return "", fmt.Errorf("invalid one-time product ID %q", value)
	}
	return OneTimeProductID(value), nil
}

func (p OneTimeProductID) String() string {
	return string(p)
}

func isValidOneTimeProductID(value string) bool {
	first := value[0]
	if !isASCIILower(first) && !isASCIIDigit(first) {
		return false
	}
	for i := 1; i < len(value); i++ {
		character := value[i]
		if !isASCIILower(character) && !isASCIIDigit(character) && character != '_' && character != '.' {
			return false
		}
	}
	return true
}

type OneTimeProductPurchaseOptionType string

const (
	OneTimeProductPurchaseOptionTypeBuy  OneTimeProductPurchaseOptionType = "buy"
	OneTimeProductPurchaseOptionTypeRent OneTimeProductPurchaseOptionType = "rent"
)

type OneTimeProductListing struct {
	LanguageCode string `json:"languageCode"`
	Title        string `json:"title,omitempty"`
	Description  string `json:"description,omitempty"`
}

type OneTimeProduct struct {
	PackageName              PackageName                         `json:"packageName"`
	ProductID                OneTimeProductID                    `json:"productId"`
	Listings                 []OneTimeProductListing             `json:"listings"`
	OfferTags                []string                            `json:"offerTags,omitempty"`
	PurchaseOptions          []OneTimeProductPurchaseOption      `json:"purchaseOptions"`
	RegionsVersion           *RegionsVersion                     `json:"regionsVersion,omitempty"`
	RestrictedCountries      []string                            `json:"restrictedCountries,omitempty"`
	TaxAndComplianceSettings *OneTimeProductTaxComplianceSetting `json:"taxAndComplianceSettings,omitempty"`
}

type OneTimeProductPurchaseOption struct {
	PurchaseOptionID         string                                             `json:"purchaseOptionId"`
	State                    string                                             `json:"state,omitempty"`
	Type                     OneTimeProductPurchaseOptionType                   `json:"type,omitempty"`
	LegacyCompatible         bool                                               `json:"legacyCompatible,omitempty"`
	MultiQuantityEnabled     bool                                               `json:"multiQuantityEnabled,omitempty"`
	RentalPeriod             string                                             `json:"rentalPeriod,omitempty"`
	ExpirationPeriod         string                                             `json:"expirationPeriod,omitempty"`
	OfferTags                []string                                           `json:"offerTags,omitempty"`
	RegionalConfigs          []OneTimeProductRegionalConfig                     `json:"regionalConfigs,omitempty"`
	NewRegionsConfig         *OneTimeProductNewRegionsPricingAndAvailability    `json:"newRegionsConfig,omitempty"`
	TaxAndComplianceSettings *OneTimeProductPurchaseOptionTaxComplianceSettings `json:"taxAndComplianceSettings,omitempty"`
}

type ProductUpdateLatencyTolerance string

const (
	ProductUpdateLatencyToleranceSensitive ProductUpdateLatencyTolerance = "latencySensitive"
	ProductUpdateLatencyToleranceTolerant  ProductUpdateLatencyTolerance = "latencyTolerant"
)

func NewProductUpdateLatencyTolerance(value string) (ProductUpdateLatencyTolerance, error) {
	switch ProductUpdateLatencyTolerance(value) {
	case "", ProductUpdateLatencyToleranceSensitive:
		return ProductUpdateLatencyToleranceSensitive, nil
	case ProductUpdateLatencyToleranceTolerant:
		return ProductUpdateLatencyToleranceTolerant, nil
	default:
		return "", fmt.Errorf("unsupported latency tolerance %q; supported values: latencySensitive, latencyTolerant", value)
	}
}

func (t ProductUpdateLatencyTolerance) String() string {
	return string(t)
}

type PurchaseOptionStateAction string

const (
	PurchaseOptionStateActionActivate   PurchaseOptionStateAction = "activate"
	PurchaseOptionStateActionDeactivate PurchaseOptionStateAction = "deactivate"
)

func (a PurchaseOptionStateAction) String() string {
	return string(a)
}

func (a PurchaseOptionStateAction) Validate() error {
	switch a {
	case PurchaseOptionStateActionActivate, PurchaseOptionStateActionDeactivate:
		return nil
	default:
		return fmt.Errorf("unsupported purchase option state action %q", a)
	}
}

type PurchaseOptionAvailability string

const (
	PurchaseOptionAvailabilityAvailable              PurchaseOptionAvailability = "available"
	PurchaseOptionAvailabilityNoLongerAvailable      PurchaseOptionAvailability = "noLongerAvailable"
	PurchaseOptionAvailabilityAvailableIfReleased    PurchaseOptionAvailability = "availableIfReleased"
	PurchaseOptionAvailabilityAvailableForOffersOnly PurchaseOptionAvailability = "availableForOffersOnly"
)

func (a PurchaseOptionAvailability) String() string {
	return string(a)
}

func (a PurchaseOptionAvailability) Validate() error {
	switch a {
	case PurchaseOptionAvailabilityAvailable,
		PurchaseOptionAvailabilityNoLongerAvailable,
		PurchaseOptionAvailabilityAvailableIfReleased,
		PurchaseOptionAvailabilityAvailableForOffersOnly:
		return nil
	default:
		return fmt.Errorf("unsupported purchase option availability %q", a)
	}
}

type OneTimeProductRegionalConfig struct {
	RegionCode   string `json:"regionCode"`
	Availability string `json:"availability,omitempty"`
	Price        *Money `json:"price,omitempty"`
}

type OneTimeProductNewRegionsPricingAndAvailability struct {
	Availability string `json:"availability,omitempty"`
	USDPrice     *Money `json:"usdPrice,omitempty"`
	EURPrice     *Money `json:"eurPrice,omitempty"`
}

type OneTimeProductTaxComplianceSetting struct {
	IsTokenizedDigitalAsset bool                `json:"isTokenizedDigitalAsset,omitempty"`
	ProductTaxCategoryCode  string              `json:"productTaxCategoryCode,omitempty"`
	RegionalAgeRatings      []RegionalAgeRating `json:"regionalAgeRatings,omitempty"`
	RegionalTaxConfigs      []RegionalTaxConfig `json:"regionalTaxConfigs,omitempty"`
}

type OneTimeProductPurchaseOptionTaxComplianceSettings struct {
	WithdrawalRightType string `json:"withdrawalRightType,omitempty"`
}

type RegionalTaxConfig struct {
	RegionCode                         string `json:"regionCode"`
	EligibleForStreamingServiceTaxRate bool   `json:"eligibleForStreamingServiceTaxRate,omitempty"`
	StreamingTaxType                   string `json:"streamingTaxType,omitempty"`
	TaxTier                            string `json:"taxTier,omitempty"`
}

type RegionalAgeRating struct {
	RegionCode           string `json:"regionCode"`
	ProductAgeRatingTier string `json:"productAgeRatingTier,omitempty"`
}

type RegionsVersion struct {
	Version string `json:"version,omitempty"`
}

type OneTimeProductListOptions struct {
	PackageName PackageName `json:"packageName"`
	PageSize    int64       `json:"pageSize,omitempty"`
	PageToken   string      `json:"pageToken,omitempty"`
}

func (o OneTimeProductListOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if o.PageSize < 0 {
		return fmt.Errorf("page size cannot be negative")
	}
	if o.PageSize > 1000 {
		return fmt.Errorf("page size cannot exceed 1000")
	}
	return nil
}

type OneTimeProductListResult struct {
	PackageName   PackageName               `json:"packageName"`
	Products      []OneTimeProduct          `json:"products"`
	NextPageToken string                    `json:"nextPageToken,omitempty"`
	Options       OneTimeProductListOptions `json:"options"`
}

type OneTimeProductLister interface {
	ListOneTimeProducts(ctx context.Context, options OneTimeProductListOptions) (OneTimeProductListResult, error)
}

func ListOneTimeProducts(ctx context.Context, lister OneTimeProductLister, options OneTimeProductListOptions) (OneTimeProductListResult, error) {
	if err := options.Validate(); err != nil {
		return OneTimeProductListResult{}, err
	}
	if lister == nil {
		return OneTimeProductListResult{}, fmt.Errorf("one-time product lister is required")
	}
	return lister.ListOneTimeProducts(ctx, options)
}

type OneTimeProductGetOptions struct {
	PackageName PackageName      `json:"packageName"`
	ProductID   OneTimeProductID `json:"productId"`
}

func (o OneTimeProductGetOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewOneTimeProductID(o.ProductID.String()); err != nil {
		return err
	}
	return nil
}

type OneTimeProductGetter interface {
	GetOneTimeProduct(ctx context.Context, packageName PackageName, productID OneTimeProductID) (OneTimeProduct, error)
}

type OneTimeProductCreator interface {
	CreateOneTimeProduct(ctx context.Context, options OneTimeProductCreateOptions) (OneTimeProduct, error)
}

type OneTimeProductDeleter interface {
	DeleteOneTimeProduct(ctx context.Context, options OneTimeProductDeleteOptions) error
}

type OneTimeProductPatcher interface {
	PatchOneTimeProduct(ctx context.Context, options OneTimeProductPatchOptions) (OneTimeProduct, error)
}

type OneTimeProductBatchListingsPatcher interface {
	BatchPatchOneTimeProductListings(ctx context.Context, options OneTimeProductBatchPatchListingsOptions) (OneTimeProductBatchPatchListingsResult, error)
}

type PurchaseOptionBatchAvailabilityPatcher interface {
	BatchPatchPurchaseOptionAvailability(ctx context.Context, options PurchaseOptionBatchPatchAvailabilityOptions) (PurchaseOptionBatchPatchAvailabilityResult, error)
}

type PurchaseOptionBatchPricePatcher interface {
	BatchPatchPurchaseOptionPrices(ctx context.Context, options PurchaseOptionBatchPatchPriceOptions) (PurchaseOptionBatchPatchPriceResult, error)
}

type OneTimeProductBatchDeleter interface {
	BatchDeleteOneTimeProducts(ctx context.Context, options OneTimeProductBatchDeleteOptions) error
}

func GetOneTimeProduct(ctx context.Context, getter OneTimeProductGetter, options OneTimeProductGetOptions) (OneTimeProduct, error) {
	if err := options.Validate(); err != nil {
		return OneTimeProduct{}, err
	}
	if getter == nil {
		return OneTimeProduct{}, fmt.Errorf("one-time product getter is required")
	}
	return getter.GetOneTimeProduct(ctx, options.PackageName, options.ProductID)
}

type OneTimeProductBatchGetOptions struct {
	PackageName PackageName        `json:"packageName"`
	ProductIDs  []OneTimeProductID `json:"productIds"`
}

func (o OneTimeProductBatchGetOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if len(o.ProductIDs) == 0 {
		return fmt.Errorf("at least one one-time product ID is required")
	}
	if len(o.ProductIDs) > 100 {
		return fmt.Errorf("one-time product batch-get cannot exceed 100 product IDs")
	}
	seen := map[OneTimeProductID]struct{}{}
	for _, productID := range o.ProductIDs {
		if _, err := NewOneTimeProductID(productID.String()); err != nil {
			return err
		}
		if _, ok := seen[productID]; ok {
			return fmt.Errorf("one-time product ID %q is duplicated", productID)
		}
		seen[productID] = struct{}{}
	}
	return nil
}

type OneTimeProductBatchGetResult struct {
	PackageName PackageName                   `json:"packageName"`
	Products    []OneTimeProduct              `json:"products"`
	Options     OneTimeProductBatchGetOptions `json:"options"`
}

type OneTimeProductBatchGetter interface {
	BatchGetOneTimeProducts(ctx context.Context, options OneTimeProductBatchGetOptions) (OneTimeProductBatchGetResult, error)
}

func BatchGetOneTimeProducts(ctx context.Context, getter OneTimeProductBatchGetter, options OneTimeProductBatchGetOptions) (OneTimeProductBatchGetResult, error) {
	if err := options.Validate(); err != nil {
		return OneTimeProductBatchGetResult{}, err
	}
	if getter == nil {
		return OneTimeProductBatchGetResult{}, fmt.Errorf("one-time product batch getter is required")
	}
	return getter.BatchGetOneTimeProducts(ctx, options)
}

type OneTimeProductDeleteOptions struct {
	PackageName      PackageName                   `json:"packageName"`
	ProductID        OneTimeProductID              `json:"productId"`
	LatencyTolerance ProductUpdateLatencyTolerance `json:"latencyTolerance"`
	Confirm          bool                          `json:"confirm"`
	DryRun           bool                          `json:"dryRun"`
}

type OneTimeProductCreateOptions struct {
	PackageName      PackageName                   `json:"packageName"`
	ProductID        OneTimeProductID              `json:"productId"`
	Product          OneTimeProduct                `json:"product"`
	RegionsVersion   string                        `json:"regionsVersion"`
	LatencyTolerance ProductUpdateLatencyTolerance `json:"latencyTolerance"`
	Confirm          bool                          `json:"confirm"`
	DryRun           bool                          `json:"dryRun"`
}

type OneTimeProductPatchOptions struct {
	PackageName      PackageName                   `json:"packageName"`
	ProductID        OneTimeProductID              `json:"productId"`
	Listing          OneTimeProductListing         `json:"listing"`
	TitleSet         bool                          `json:"titleSet,omitempty"`
	DescriptionSet   bool                          `json:"descriptionSet,omitempty"`
	RegionsVersion   string                        `json:"regionsVersion"`
	LatencyTolerance ProductUpdateLatencyTolerance `json:"latencyTolerance"`
	Confirm          bool                          `json:"confirm"`
	DryRun           bool                          `json:"dryRun"`
}

type OneTimeProductBatchPatchListingRequest struct {
	ProductID OneTimeProductID      `json:"productId"`
	Listing   OneTimeProductListing `json:"listing"`
}

type OneTimeProductBatchPatchListingsOptions struct {
	PackageName      PackageName                              `json:"packageName"`
	Requests         []OneTimeProductBatchPatchListingRequest `json:"requests"`
	RegionsVersion   string                                   `json:"regionsVersion"`
	LatencyTolerance ProductUpdateLatencyTolerance            `json:"latencyTolerance"`
	Confirm          bool                                     `json:"confirm"`
	DryRun           bool                                     `json:"dryRun"`
}

func (o OneTimeProductDeleteOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewOneTimeProductID(o.ProductID.String()); err != nil {
		return err
	}
	if _, err := NewProductUpdateLatencyTolerance(o.LatencyTolerance.String()); err != nil {
		return err
	}
	if o.Confirm && o.DryRun {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.Confirm && !o.DryRun {
		return fmt.Errorf("one-time product deletion requires --confirm or --dry-run")
	}
	return nil
}

func (o OneTimeProductCreateOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewOneTimeProductID(o.ProductID.String()); err != nil {
		return err
	}
	if strings.TrimSpace(o.RegionsVersion) == "" {
		return fmt.Errorf("regions version is required")
	}
	if strings.TrimSpace(o.RegionsVersion) != o.RegionsVersion {
		return fmt.Errorf("regions version cannot have leading or trailing whitespace")
	}
	if _, err := NewProductUpdateLatencyTolerance(o.LatencyTolerance.String()); err != nil {
		return err
	}
	if o.Confirm && o.DryRun {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.Confirm && !o.DryRun {
		return fmt.Errorf("one-time product creation requires --confirm or --dry-run")
	}
	return validateOneTimeProductForCreate(oneTimeProductCreateDesiredProduct(o))
}

func (o OneTimeProductPatchOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewOneTimeProductID(o.ProductID.String()); err != nil {
		return err
	}
	if err := validateOneTimeProductListing(o.Listing); err != nil {
		return err
	}
	if strings.TrimSpace(o.RegionsVersion) == "" {
		return fmt.Errorf("regions version is required")
	}
	if strings.TrimSpace(o.RegionsVersion) != o.RegionsVersion {
		return fmt.Errorf("regions version cannot have leading or trailing whitespace")
	}
	if _, err := NewProductUpdateLatencyTolerance(o.LatencyTolerance.String()); err != nil {
		return err
	}
	if o.Confirm && o.DryRun {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.Confirm && !o.DryRun {
		return fmt.Errorf("one-time product patch requires --confirm or --dry-run")
	}
	return nil
}

func (o OneTimeProductBatchPatchListingsOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if len(o.Requests) == 0 {
		return fmt.Errorf("at least one one-time product listing patch is required")
	}
	seenListings := map[string]struct{}{}
	seenProducts := map[OneTimeProductID]struct{}{}
	for _, request := range o.Requests {
		if _, err := NewOneTimeProductID(request.ProductID.String()); err != nil {
			return err
		}
		if err := validateRequiredOneTimeProductListing(request.Listing); err != nil {
			return err
		}
		key := oneTimeProductBatchPatchListingKey(request.ProductID, request.Listing.LanguageCode)
		if _, ok := seenListings[key]; ok {
			return fmt.Errorf("one-time product listing %s is duplicated", key)
		}
		seenListings[key] = struct{}{}
		seenProducts[request.ProductID] = struct{}{}
	}
	if len(seenProducts) > 100 {
		return fmt.Errorf("one-time product listing batch patch cannot exceed 100 products")
	}
	if strings.TrimSpace(o.RegionsVersion) == "" {
		return fmt.Errorf("regions version is required")
	}
	if strings.TrimSpace(o.RegionsVersion) != o.RegionsVersion {
		return fmt.Errorf("regions version cannot have leading or trailing whitespace")
	}
	if _, err := NewProductUpdateLatencyTolerance(o.LatencyTolerance.String()); err != nil {
		return err
	}
	if o.Confirm && o.DryRun {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.Confirm && !o.DryRun {
		return fmt.Errorf("one-time product listing batch patch requires --confirm or --dry-run")
	}
	return nil
}

func (o OneTimeProductCreateOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live one-time product creation cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live one-time product creation requires --confirm")
	}
	return nil
}

func (o OneTimeProductDeleteOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live one-time product deletion cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live one-time product deletion requires --confirm")
	}
	return nil
}

func (o OneTimeProductPatchOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live one-time product patch cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live one-time product patch requires --confirm")
	}
	return nil
}

func (o OneTimeProductBatchPatchListingsOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live one-time product listing batch patch cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live one-time product listing batch patch requires --confirm")
	}
	return nil
}

func oneTimeProductBatchPatchListingKey(productID OneTimeProductID, languageCode string) string {
	return productID.String() + "/" + languageCode
}

func validateOneTimeProductListing(listing OneTimeProductListing) error {
	if _, err := NewListingLanguage(listing.LanguageCode); err != nil {
		return err
	}
	if strings.TrimSpace(listing.Title) != listing.Title {
		return fmt.Errorf("one-time product listing title cannot have leading or trailing whitespace")
	}
	if utf8.RuneCountInString(listing.Title) > 55 {
		return fmt.Errorf("one-time product listing title cannot exceed 55 characters")
	}
	if strings.TrimSpace(listing.Description) != listing.Description {
		return fmt.Errorf("one-time product listing description cannot have leading or trailing whitespace")
	}
	if utf8.RuneCountInString(listing.Description) > 200 {
		return fmt.Errorf("one-time product listing description cannot exceed 200 characters")
	}
	return nil
}

func validateRequiredOneTimeProductListing(listing OneTimeProductListing) error {
	if err := validateOneTimeProductListing(listing); err != nil {
		return err
	}
	if strings.TrimSpace(listing.Title) == "" {
		return fmt.Errorf("one-time product listing title is required")
	}
	if strings.TrimSpace(listing.Description) == "" {
		return fmt.Errorf("one-time product listing description is required")
	}
	return nil
}

func validateOneTimeProductForCreate(product OneTimeProduct) error {
	if _, err := NewOneTimeProductID(product.ProductID.String()); err != nil {
		return err
	}
	if len(product.Listings) == 0 {
		return fmt.Errorf("one-time product create requires at least one listing")
	}
	seenListings := map[string]struct{}{}
	for _, listing := range product.Listings {
		if err := validateRequiredOneTimeProductListing(listing); err != nil {
			return err
		}
		language, err := NewListingLanguage(listing.LanguageCode)
		if err != nil {
			return err
		}
		if _, ok := seenListings[language.String()]; ok {
			return fmt.Errorf("one-time product create listing %s is duplicated", language)
		}
		seenListings[language.String()] = struct{}{}
	}
	if len(product.PurchaseOptions) == 0 {
		return fmt.Errorf("one-time product create requires at least one purchase option")
	}
	legacyCompatibleCount := 0
	seenOptions := map[string]struct{}{}
	for _, option := range product.PurchaseOptions {
		if err := validateOneTimeProductPurchaseOptionForCreate(option); err != nil {
			return err
		}
		if _, ok := seenOptions[option.PurchaseOptionID]; ok {
			return fmt.Errorf("one-time product purchase option %s is duplicated", option.PurchaseOptionID)
		}
		seenOptions[option.PurchaseOptionID] = struct{}{}
		if option.LegacyCompatible {
			legacyCompatibleCount++
			if legacyCompatibleCount > 1 {
				return fmt.Errorf("one-time product create can set at most one legacy-compatible purchase option")
			}
		}
	}
	for _, tag := range product.OfferTags {
		if err := validateSubscriptionOfferTag(tag); err != nil {
			return err
		}
	}
	for _, country := range product.RestrictedCountries {
		if !isValidRegionCode(country) {
			return fmt.Errorf("restricted country must be a two-letter ISO 3166 code")
		}
	}
	if product.TaxAndComplianceSettings != nil {
		if err := validateOneTimeProductTaxComplianceSettingsForCreate(*product.TaxAndComplianceSettings); err != nil {
			return err
		}
	}
	return nil
}

func validateOneTimeProductPurchaseOptionForCreate(option OneTimeProductPurchaseOption) error {
	if _, err := NewOneTimeProductPurchaseOptionID(option.PurchaseOptionID); err != nil {
		return err
	}
	for _, tag := range option.OfferTags {
		if err := validateSubscriptionOfferTag(tag); err != nil {
			return err
		}
	}
	switch option.Type {
	case OneTimeProductPurchaseOptionTypeBuy:
		if option.RentalPeriod != "" || option.ExpirationPeriod != "" {
			return fmt.Errorf("buy purchase options cannot set rent fields")
		}
	case OneTimeProductPurchaseOptionTypeRent:
		if option.LegacyCompatible || option.MultiQuantityEnabled {
			return fmt.Errorf("rent purchase options cannot set buy fields")
		}
		if err := validateISODuration("rental period", option.RentalPeriod); err != nil {
			return err
		}
		if err := validateISODuration("expiration period", option.ExpirationPeriod); err != nil {
			return err
		}
	default:
		return fmt.Errorf("purchase option type must be buy or rent")
	}
	if err := validateOneTimeProductRegionalConfigsForCreate(option.RegionalConfigs); err != nil {
		return err
	}
	if option.NewRegionsConfig != nil {
		if err := validateOneTimeProductNewRegionsConfigForCreate(*option.NewRegionsConfig); err != nil {
			return err
		}
	}
	if option.TaxAndComplianceSettings != nil && strings.TrimSpace(option.TaxAndComplianceSettings.WithdrawalRightType) == "" {
		return fmt.Errorf("purchase option tax compliance settings require withdrawal right type")
	}
	return nil
}

func validateOneTimeProductRegionalConfigsForCreate(configs []OneTimeProductRegionalConfig) error {
	seen := map[string]struct{}{}
	for _, config := range configs {
		if !isValidRegionCode(config.RegionCode) {
			return fmt.Errorf("regional config region must be a two-letter ISO 3166 code")
		}
		if _, ok := seen[config.RegionCode]; ok {
			return fmt.Errorf("regional config %s is duplicated", config.RegionCode)
		}
		seen[config.RegionCode] = struct{}{}
		if config.Availability != "" {
			if err := validateOneTimeProductAvailability(config.Availability); err != nil {
				return err
			}
		}
		if config.Price != nil {
			if err := validateMoney(*config.Price); err != nil {
				return fmt.Errorf("regional config %s: %w", config.RegionCode, err)
			}
		}
		if config.Price == nil {
			return fmt.Errorf("regional config %s requires price", config.RegionCode)
		}
		if strings.TrimSpace(config.Availability) == "" {
			return fmt.Errorf("regional config %s requires availability", config.RegionCode)
		}
	}
	return nil
}

func validateOneTimeProductNewRegionsConfigForCreate(config OneTimeProductNewRegionsPricingAndAvailability) error {
	if strings.TrimSpace(config.Availability) == "" {
		return fmt.Errorf("new regions config requires availability")
	}
	if config.Availability != "" {
		if err := validateOneTimeProductAvailability(config.Availability); err != nil {
			return fmt.Errorf("new regions %w", err)
		}
	}
	if config.USDPrice == nil {
		return fmt.Errorf("new regions config requires USD price")
	}
	if config.USDPrice != nil {
		if config.USDPrice.CurrencyCode != "USD" {
			return fmt.Errorf("new regions USD price currency must be USD")
		}
		if err := validateMoney(*config.USDPrice); err != nil {
			return fmt.Errorf("new regions USD price: %w", err)
		}
	}
	if config.EURPrice == nil {
		return fmt.Errorf("new regions config requires EUR price")
	}
	if config.EURPrice != nil {
		if config.EURPrice.CurrencyCode != "EUR" {
			return fmt.Errorf("new regions EUR price currency must be EUR")
		}
		if err := validateMoney(*config.EURPrice); err != nil {
			return fmt.Errorf("new regions EUR price: %w", err)
		}
	}
	return nil
}

func validateOneTimeProductTaxComplianceSettingsForCreate(settings OneTimeProductTaxComplianceSetting) error {
	if strings.TrimSpace(settings.ProductTaxCategoryCode) != "" {
		return fmt.Errorf("one-time product product tax category code is not supported by the pinned Google API client")
	}
	if len(settings.RegionalAgeRatings) > 0 {
		return fmt.Errorf("one-time product regional age ratings are not supported by the pinned Google API client")
	}
	for _, config := range settings.RegionalTaxConfigs {
		if !isValidRegionCode(config.RegionCode) {
			return fmt.Errorf("regional tax config region must be a two-letter ISO 3166 code")
		}
	}
	return nil
}

func validateOneTimeProductAvailability(value string) error {
	switch value {
	case "available", "noLongerAvailable", "availableIfReleased", "availableForOffersOnly",
		"AVAILABLE", "NO_LONGER_AVAILABLE", "AVAILABLE_IF_RELEASED", "AVAILABLE_FOR_OFFERS_ONLY":
		return nil
	default:
		return fmt.Errorf("unsupported purchase option availability %q", value)
	}
}

type OneTimeProductDeletePlan struct {
	PackageName      PackageName                   `json:"packageName"`
	ProductID        OneTimeProductID              `json:"productId"`
	LatencyTolerance ProductUpdateLatencyTolerance `json:"latencyTolerance"`
	Confirm          bool                          `json:"confirm"`
	Steps            []string                      `json:"steps"`
}

type OneTimeProductCreatePlan struct {
	PackageName      PackageName                   `json:"packageName"`
	ProductID        OneTimeProductID              `json:"productId"`
	RegionsVersion   string                        `json:"regionsVersion"`
	LatencyTolerance ProductUpdateLatencyTolerance `json:"latencyTolerance"`
	Confirm          bool                          `json:"confirm"`
	Steps            []string                      `json:"steps"`
}

type OneTimeProductDeleteResult struct {
	PackageName PackageName              `json:"packageName"`
	ProductID   OneTimeProductID         `json:"productId"`
	DryRun      bool                     `json:"dryRun"`
	Deleted     bool                     `json:"deleted"`
	Plan        OneTimeProductDeletePlan `json:"plan"`
}

type OneTimeProductCreateResult struct {
	PackageName PackageName              `json:"packageName"`
	ProductID   OneTimeProductID         `json:"productId"`
	DryRun      bool                     `json:"dryRun"`
	Created     bool                     `json:"created"`
	Desired     OneTimeProduct           `json:"desiredProduct"`
	Product     *OneTimeProduct          `json:"product,omitempty"`
	Plan        OneTimeProductCreatePlan `json:"plan"`
}

type OneTimeProductPatchPlan struct {
	PackageName      PackageName                   `json:"packageName"`
	ProductID        OneTimeProductID              `json:"productId"`
	Listing          OneTimeProductListing         `json:"listing"`
	TitleSet         bool                          `json:"titleSet,omitempty"`
	DescriptionSet   bool                          `json:"descriptionSet,omitempty"`
	UpdateMask       string                        `json:"updateMask"`
	RegionsVersion   string                        `json:"regionsVersion"`
	LatencyTolerance ProductUpdateLatencyTolerance `json:"latencyTolerance"`
	Confirm          bool                          `json:"confirm"`
	Steps            []string                      `json:"steps"`
}

type OneTimeProductBatchPatchListingsPlan struct {
	PackageName      PackageName                              `json:"packageName"`
	Requests         []OneTimeProductBatchPatchListingRequest `json:"requests"`
	UpdateMask       string                                   `json:"updateMask"`
	RegionsVersion   string                                   `json:"regionsVersion"`
	LatencyTolerance ProductUpdateLatencyTolerance            `json:"latencyTolerance"`
	Confirm          bool                                     `json:"confirm"`
	Steps            []string                                 `json:"steps"`
}

type OneTimeProductPatchResult struct {
	PackageName PackageName             `json:"packageName"`
	ProductID   OneTimeProductID        `json:"productId"`
	DryRun      bool                    `json:"dryRun"`
	Applied     bool                    `json:"applied"`
	Desired     OneTimeProduct          `json:"desiredProduct"`
	Product     *OneTimeProduct         `json:"product,omitempty"`
	Plan        OneTimeProductPatchPlan `json:"plan"`
}

type OneTimeProductBatchPatchListingsResult struct {
	PackageName PackageName                          `json:"packageName"`
	DryRun      bool                                 `json:"dryRun"`
	Applied     bool                                 `json:"applied"`
	Desired     []OneTimeProduct                     `json:"desiredProducts"`
	Products    []OneTimeProduct                     `json:"products,omitempty"`
	Plan        OneTimeProductBatchPatchListingsPlan `json:"plan"`
}

func NewOneTimeProductCreatePlan(options OneTimeProductCreateOptions) (OneTimeProductCreatePlan, error) {
	if err := options.Validate(); err != nil {
		return OneTimeProductCreatePlan{}, err
	}
	return OneTimeProductCreatePlan{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		RegionsVersion:   options.RegionsVersion,
		LatencyTolerance: options.LatencyTolerance,
		Confirm:          options.Confirm,
		Steps:            []string{"patch one-time product with allowMissing=true"},
	}, nil
}

func CreateOneTimeProduct(ctx context.Context, creator OneTimeProductCreator, options OneTimeProductCreateOptions) (OneTimeProductCreateResult, error) {
	plan, err := NewOneTimeProductCreatePlan(options)
	if err != nil {
		return OneTimeProductCreateResult{}, err
	}
	result := OneTimeProductCreateResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		DryRun:      options.DryRun,
		Desired:     oneTimeProductCreateDesiredProduct(options),
		Plan:        plan,
	}
	if options.DryRun {
		return result, nil
	}
	if creator == nil {
		return OneTimeProductCreateResult{}, fmt.Errorf("one-time product creator is required")
	}
	product, err := creator.CreateOneTimeProduct(ctx, options)
	if err != nil {
		return OneTimeProductCreateResult{}, err
	}
	result.Created = true
	result.Product = &product
	return result, nil
}

func DeleteOneTimeProduct(ctx context.Context, deleter OneTimeProductDeleter, options OneTimeProductDeleteOptions) (OneTimeProductDeleteResult, error) {
	if err := options.Validate(); err != nil {
		return OneTimeProductDeleteResult{}, err
	}
	result := OneTimeProductDeleteResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		DryRun:      options.DryRun,
		Plan: OneTimeProductDeletePlan{
			PackageName:      options.PackageName,
			ProductID:        options.ProductID,
			LatencyTolerance: options.LatencyTolerance,
			Confirm:          options.Confirm,
			Steps:            []string{"delete one-time product"},
		},
	}
	if options.DryRun {
		return result, nil
	}
	if deleter == nil {
		return OneTimeProductDeleteResult{}, fmt.Errorf("one-time product deleter is required")
	}
	if err := deleter.DeleteOneTimeProduct(ctx, options); err != nil {
		return OneTimeProductDeleteResult{}, err
	}
	result.Deleted = true
	return result, nil
}

func NewOneTimeProductPatchPlan(options OneTimeProductPatchOptions) (OneTimeProductPatchPlan, error) {
	if err := options.Validate(); err != nil {
		return OneTimeProductPatchPlan{}, err
	}
	return OneTimeProductPatchPlan{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		Listing:          options.Listing,
		TitleSet:         options.TitleSet,
		DescriptionSet:   options.DescriptionSet,
		UpdateMask:       oneTimeProductPatchUpdateMask,
		RegionsVersion:   options.RegionsVersion,
		LatencyTolerance: options.LatencyTolerance,
		Confirm:          options.Confirm,
		Steps:            oneTimeProductPatchSteps(options),
	}, nil
}

func NewOneTimeProductBatchPatchListingsPlan(options OneTimeProductBatchPatchListingsOptions) (OneTimeProductBatchPatchListingsPlan, error) {
	if err := options.Validate(); err != nil {
		return OneTimeProductBatchPatchListingsPlan{}, err
	}
	return OneTimeProductBatchPatchListingsPlan{
		PackageName:      options.PackageName,
		Requests:         options.Requests,
		UpdateMask:       oneTimeProductPatchUpdateMask,
		RegionsVersion:   options.RegionsVersion,
		LatencyTolerance: options.LatencyTolerance,
		Confirm:          options.Confirm,
		Steps:            oneTimeProductBatchPatchListingsSteps(options),
	}, nil
}

func PatchOneTimeProduct(ctx context.Context, patcher OneTimeProductPatcher, options OneTimeProductPatchOptions) (OneTimeProductPatchResult, error) {
	plan, err := NewOneTimeProductPatchPlan(options)
	if err != nil {
		return OneTimeProductPatchResult{}, err
	}
	result := OneTimeProductPatchResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		DryRun:      options.DryRun,
		Desired: OneTimeProduct{
			PackageName:     options.PackageName,
			ProductID:       options.ProductID,
			Listings:        []OneTimeProductListing{options.Listing},
			PurchaseOptions: []OneTimeProductPurchaseOption{},
		},
		Plan: plan,
	}
	if options.DryRun {
		return result, nil
	}
	if patcher == nil {
		return OneTimeProductPatchResult{}, fmt.Errorf("one-time product patcher is required")
	}
	product, err := patcher.PatchOneTimeProduct(ctx, options)
	if err != nil {
		return OneTimeProductPatchResult{}, err
	}
	result.Applied = true
	result.Product = &product
	return result, nil
}

func BatchPatchOneTimeProductListings(ctx context.Context, patcher OneTimeProductBatchListingsPatcher, options OneTimeProductBatchPatchListingsOptions) (OneTimeProductBatchPatchListingsResult, error) {
	plan, err := NewOneTimeProductBatchPatchListingsPlan(options)
	if err != nil {
		return OneTimeProductBatchPatchListingsResult{}, err
	}
	result := OneTimeProductBatchPatchListingsResult{
		PackageName: options.PackageName,
		DryRun:      options.DryRun,
		Applied:     false,
		Desired:     desiredOneTimeProductsForBatchListingPatch(options),
		Plan:        plan,
	}
	if options.DryRun {
		return result, nil
	}
	if patcher == nil {
		return OneTimeProductBatchPatchListingsResult{}, fmt.Errorf("one-time product listing batch patcher is required")
	}
	liveResult, err := patcher.BatchPatchOneTimeProductListings(ctx, options)
	if err != nil {
		return OneTimeProductBatchPatchListingsResult{}, err
	}
	liveResult.PackageName = options.PackageName
	liveResult.DryRun = false
	liveResult.Applied = true
	liveResult.Desired = result.Desired
	liveResult.Plan = plan
	return liveResult, nil
}

func desiredOneTimeProductsForBatchListingPatch(options OneTimeProductBatchPatchListingsOptions) []OneTimeProduct {
	byProduct := map[OneTimeProductID]int{}
	products := make([]OneTimeProduct, 0)
	for _, request := range options.Requests {
		index, ok := byProduct[request.ProductID]
		if !ok {
			byProduct[request.ProductID] = len(products)
			products = append(products, OneTimeProduct{
				PackageName:     options.PackageName,
				ProductID:       request.ProductID,
				Listings:        []OneTimeProductListing{},
				PurchaseOptions: []OneTimeProductPurchaseOption{},
			})
			index = len(products) - 1
		}
		products[index].Listings = append(products[index].Listings, request.Listing)
	}
	return products
}

func oneTimeProductCreateDesiredProduct(options OneTimeProductCreateOptions) OneTimeProduct {
	product := options.Product
	product.PackageName = options.PackageName
	product.ProductID = options.ProductID
	if product.Listings == nil {
		product.Listings = []OneTimeProductListing{}
	}
	if product.PurchaseOptions == nil {
		product.PurchaseOptions = []OneTimeProductPurchaseOption{}
	}
	for index := range product.Listings {
		if language, err := NewListingLanguage(product.Listings[index].LanguageCode); err == nil {
			product.Listings[index].LanguageCode = language.String()
		}
	}
	for index := range product.PurchaseOptions {
		product.PurchaseOptions[index].State = ""
	}
	return product
}

type OneTimeProductBatchDeleteOptions struct {
	PackageName      PackageName                   `json:"packageName"`
	ProductIDs       []OneTimeProductID            `json:"productIds"`
	LatencyTolerance ProductUpdateLatencyTolerance `json:"latencyTolerance"`
	Confirm          bool                          `json:"confirm"`
	DryRun           bool                          `json:"dryRun"`
}

func (o OneTimeProductBatchDeleteOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if len(o.ProductIDs) == 0 {
		return fmt.Errorf("at least one one-time product ID is required")
	}
	if len(o.ProductIDs) > 100 {
		return fmt.Errorf("one-time product batch-delete cannot exceed 100 product IDs")
	}
	seen := map[OneTimeProductID]struct{}{}
	for _, productID := range o.ProductIDs {
		if _, err := NewOneTimeProductID(productID.String()); err != nil {
			return err
		}
		if _, ok := seen[productID]; ok {
			return fmt.Errorf("one-time product ID %q is duplicated", productID)
		}
		seen[productID] = struct{}{}
	}
	if _, err := NewProductUpdateLatencyTolerance(o.LatencyTolerance.String()); err != nil {
		return err
	}
	if o.Confirm && o.DryRun {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.Confirm && !o.DryRun {
		return fmt.Errorf("one-time product batch deletion requires --confirm or --dry-run")
	}
	return nil
}

func (o OneTimeProductBatchDeleteOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live one-time product batch deletion cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live one-time product batch deletion requires --confirm")
	}
	return nil
}

type OneTimeProductBatchDeletePlan struct {
	PackageName      PackageName                   `json:"packageName"`
	ProductIDs       []OneTimeProductID            `json:"productIds"`
	LatencyTolerance ProductUpdateLatencyTolerance `json:"latencyTolerance"`
	Confirm          bool                          `json:"confirm"`
	Steps            []string                      `json:"steps"`
}

type OneTimeProductBatchDeleteResult struct {
	PackageName PackageName                   `json:"packageName"`
	ProductIDs  []OneTimeProductID            `json:"productIds"`
	DryRun      bool                          `json:"dryRun"`
	Deleted     bool                          `json:"deleted"`
	Plan        OneTimeProductBatchDeletePlan `json:"plan"`
}

func BatchDeleteOneTimeProducts(ctx context.Context, deleter OneTimeProductBatchDeleter, options OneTimeProductBatchDeleteOptions) (OneTimeProductBatchDeleteResult, error) {
	if err := options.Validate(); err != nil {
		return OneTimeProductBatchDeleteResult{}, err
	}
	productIDs := append([]OneTimeProductID(nil), options.ProductIDs...)
	result := OneTimeProductBatchDeleteResult{
		PackageName: options.PackageName,
		ProductIDs:  productIDs,
		DryRun:      options.DryRun,
		Plan: OneTimeProductBatchDeletePlan{
			PackageName:      options.PackageName,
			ProductIDs:       productIDs,
			LatencyTolerance: options.LatencyTolerance,
			Confirm:          options.Confirm,
			Steps:            []string{"delete one-time products"},
		},
	}
	if options.DryRun {
		return result, nil
	}
	if deleter == nil {
		return OneTimeProductBatchDeleteResult{}, fmt.Errorf("one-time product batch deleter is required")
	}
	if err := deleter.BatchDeleteOneTimeProducts(ctx, options); err != nil {
		return OneTimeProductBatchDeleteResult{}, err
	}
	result.Deleted = true
	return result, nil
}

const (
	oneTimeProductCreateUpdateMask = "listings,offerTags,purchaseOptions,restrictedPaymentCountries,taxAndComplianceSettings"
	oneTimeProductPatchUpdateMask  = "listings"
)

func oneTimeProductPatchSteps(options OneTimeProductPatchOptions) []string {
	if options.DryRun {
		return []string{"plan one-time product listing patch"}
	}
	return []string{"fetch current one-time product", "merge localized listing", "patch one-time product listings"}
}

func oneTimeProductBatchPatchListingsSteps(options OneTimeProductBatchPatchListingsOptions) []string {
	if options.DryRun {
		return []string{"plan one-time product listing batch patch"}
	}
	return []string{"fetch current one-time products", "merge localized listings", "batch patch one-time product listings"}
}

const OneTimeProductWildcardID = "-"

type OneTimeProductBatchParentProductID string

func NewOneTimeProductBatchParentProductID(value string) (OneTimeProductBatchParentProductID, error) {
	if value == OneTimeProductWildcardID {
		return OneTimeProductBatchParentProductID(value), nil
	}
	if _, err := NewOneTimeProductID(value); err != nil {
		return "", err
	}
	return OneTimeProductBatchParentProductID(value), nil
}

func (p OneTimeProductBatchParentProductID) String() string {
	return string(p)
}

func (p OneTimeProductBatchParentProductID) IsWildcard() bool {
	return p == OneTimeProductBatchParentProductID(OneTimeProductWildcardID)
}

type PurchaseOptionBatchDeleteRequest struct {
	ProductID        OneTimeProductID               `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID `json:"purchaseOptionId"`
}

func NewPurchaseOptionBatchDeleteRequest(value string) (PurchaseOptionBatchDeleteRequest, error) {
	productIDValue, purchaseOptionIDValue, ok := strings.Cut(value, "/")
	if !ok {
		return PurchaseOptionBatchDeleteRequest{}, fmt.Errorf("purchase option must be formatted as productId/purchaseOptionId")
	}
	productID, err := NewOneTimeProductID(productIDValue)
	if err != nil {
		return PurchaseOptionBatchDeleteRequest{}, err
	}
	purchaseOptionID, err := NewOneTimeProductPurchaseOptionID(purchaseOptionIDValue)
	if err != nil {
		return PurchaseOptionBatchDeleteRequest{}, err
	}
	return PurchaseOptionBatchDeleteRequest{ProductID: productID, PurchaseOptionID: purchaseOptionID}, nil
}

type PurchaseOptionBatchDeleteOptions struct {
	PackageName      PackageName                        `json:"packageName"`
	ParentProductID  OneTimeProductBatchParentProductID `json:"parentProductId"`
	Requests         []PurchaseOptionBatchDeleteRequest `json:"requests"`
	LatencyTolerance ProductUpdateLatencyTolerance      `json:"latencyTolerance"`
	Force            bool                               `json:"force"`
	Confirm          bool                               `json:"confirm"`
	DryRun           bool                               `json:"dryRun"`
}

func (o PurchaseOptionBatchDeleteOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewOneTimeProductBatchParentProductID(o.ParentProductID.String()); err != nil {
		return err
	}
	if len(o.Requests) == 0 {
		return fmt.Errorf("at least one purchase option is required")
	}
	if len(o.Requests) > 100 {
		return fmt.Errorf("purchase option batch-delete cannot exceed 100 purchase options")
	}
	seenRequests := map[PurchaseOptionBatchDeleteRequest]struct{}{}
	seenProducts := map[OneTimeProductID]struct{}{}
	for _, request := range o.Requests {
		if _, err := NewOneTimeProductID(request.ProductID.String()); err != nil {
			return err
		}
		if _, err := NewOneTimeProductPurchaseOptionID(request.PurchaseOptionID.String()); err != nil {
			return err
		}
		if !o.ParentProductID.IsWildcard() && request.ProductID.String() != o.ParentProductID.String() {
			return fmt.Errorf("purchase option %s/%s does not match parent product ID %q", request.ProductID, request.PurchaseOptionID, o.ParentProductID)
		}
		if _, ok := seenRequests[request]; ok {
			return fmt.Errorf("purchase option %s/%s is duplicated", request.ProductID, request.PurchaseOptionID)
		}
		seenRequests[request] = struct{}{}
		if _, ok := seenProducts[request.ProductID]; ok {
			return fmt.Errorf("purchase option batch-delete accepts at most one request per one-time product; product %q is repeated", request.ProductID)
		}
		seenProducts[request.ProductID] = struct{}{}
	}
	if len(seenProducts) == 1 && o.ParentProductID.IsWildcard() {
		return fmt.Errorf("single-product purchase option batch-delete requires parent product ID, not %q", OneTimeProductWildcardID)
	}
	if len(seenProducts) > 1 && !o.ParentProductID.IsWildcard() {
		return fmt.Errorf("multi-product purchase option batch-delete requires parent product ID %q", OneTimeProductWildcardID)
	}
	if _, err := NewProductUpdateLatencyTolerance(o.LatencyTolerance.String()); err != nil {
		return err
	}
	if o.Confirm && o.DryRun {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.Confirm && !o.DryRun {
		return fmt.Errorf("purchase option batch deletion requires --confirm or --dry-run")
	}
	return nil
}

func (o PurchaseOptionBatchDeleteOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live purchase option batch deletion cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live purchase option batch deletion requires --confirm")
	}
	return nil
}

type PurchaseOptionBatchDeletePlan struct {
	PackageName      PackageName                        `json:"packageName"`
	ParentProductID  OneTimeProductBatchParentProductID `json:"parentProductId"`
	Requests         []PurchaseOptionBatchDeleteRequest `json:"requests"`
	LatencyTolerance ProductUpdateLatencyTolerance      `json:"latencyTolerance"`
	Force            bool                               `json:"force"`
	Confirm          bool                               `json:"confirm"`
	Steps            []string                           `json:"steps"`
}

type PurchaseOptionBatchDeleteResult struct {
	PackageName     PackageName                        `json:"packageName"`
	ParentProductID OneTimeProductBatchParentProductID `json:"parentProductId"`
	Requests        []PurchaseOptionBatchDeleteRequest `json:"requests"`
	Force           bool                               `json:"force"`
	DryRun          bool                               `json:"dryRun"`
	Deleted         bool                               `json:"deleted"`
	Plan            PurchaseOptionBatchDeletePlan      `json:"plan"`
}

type PurchaseOptionBatchDeleter interface {
	BatchDeletePurchaseOptions(ctx context.Context, options PurchaseOptionBatchDeleteOptions) error
}

func BatchDeletePurchaseOptions(ctx context.Context, deleter PurchaseOptionBatchDeleter, options PurchaseOptionBatchDeleteOptions) (PurchaseOptionBatchDeleteResult, error) {
	if err := options.Validate(); err != nil {
		return PurchaseOptionBatchDeleteResult{}, err
	}
	requests := append([]PurchaseOptionBatchDeleteRequest(nil), options.Requests...)
	result := PurchaseOptionBatchDeleteResult{
		PackageName:     options.PackageName,
		ParentProductID: options.ParentProductID,
		Requests:        requests,
		Force:           options.Force,
		DryRun:          options.DryRun,
		Plan: PurchaseOptionBatchDeletePlan{
			PackageName:      options.PackageName,
			ParentProductID:  options.ParentProductID,
			Requests:         requests,
			LatencyTolerance: options.LatencyTolerance,
			Force:            options.Force,
			Confirm:          options.Confirm,
			Steps:            purchaseOptionBatchDeleteSteps(options),
		},
	}
	if options.DryRun {
		return result, nil
	}
	if deleter == nil {
		return PurchaseOptionBatchDeleteResult{}, fmt.Errorf("purchase option batch deleter is required")
	}
	if err := deleter.BatchDeletePurchaseOptions(ctx, options); err != nil {
		return PurchaseOptionBatchDeleteResult{}, err
	}
	result.Deleted = true
	return result, nil
}

func purchaseOptionBatchDeleteSteps(options PurchaseOptionBatchDeleteOptions) []string {
	if options.DryRun {
		return []string{"plan purchase option batch deletion"}
	}
	return []string{"batch delete purchase options"}
}

type PurchaseOptionStateUpdateOptions struct {
	PackageName      PackageName                    `json:"packageName"`
	ProductID        OneTimeProductID               `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID `json:"purchaseOptionId"`
	Action           PurchaseOptionStateAction      `json:"action"`
	LatencyTolerance ProductUpdateLatencyTolerance  `json:"latencyTolerance"`
	Confirm          bool                           `json:"confirm"`
	DryRun           bool                           `json:"dryRun"`
}

type PurchaseOptionAvailabilityPatchRequest struct {
	ProductID        OneTimeProductID               `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID `json:"purchaseOptionId"`
	RegionCode       string                         `json:"regionCode"`
	Availability     PurchaseOptionAvailability     `json:"availability"`
}

type PurchaseOptionBatchPatchAvailabilityOptions struct {
	PackageName      PackageName                              `json:"packageName"`
	Requests         []PurchaseOptionAvailabilityPatchRequest `json:"requests"`
	RegionsVersion   string                                   `json:"regionsVersion"`
	LatencyTolerance ProductUpdateLatencyTolerance            `json:"latencyTolerance"`
	Confirm          bool                                     `json:"confirm"`
	DryRun           bool                                     `json:"dryRun"`
}

type PurchaseOptionPricePatchRequest struct {
	ProductID        OneTimeProductID               `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID `json:"purchaseOptionId"`
	RegionCode       string                         `json:"regionCode"`
	Price            Money                          `json:"price"`
}

type PurchaseOptionBatchPatchPriceOptions struct {
	PackageName      PackageName                       `json:"packageName"`
	Requests         []PurchaseOptionPricePatchRequest `json:"requests"`
	RegionsVersion   string                            `json:"regionsVersion"`
	LatencyTolerance ProductUpdateLatencyTolerance     `json:"latencyTolerance"`
	Confirm          bool                              `json:"confirm"`
	DryRun           bool                              `json:"dryRun"`
}

func (o PurchaseOptionStateUpdateOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewOneTimeProductID(o.ProductID.String()); err != nil {
		return err
	}
	if _, err := NewOneTimeProductPurchaseOptionID(o.PurchaseOptionID.String()); err != nil {
		return err
	}
	if err := o.Action.Validate(); err != nil {
		return err
	}
	if _, err := NewProductUpdateLatencyTolerance(o.LatencyTolerance.String()); err != nil {
		return err
	}
	if o.Confirm && o.DryRun {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.Confirm && !o.DryRun {
		return fmt.Errorf("purchase option state update requires --confirm or --dry-run")
	}
	return nil
}

func (o PurchaseOptionBatchPatchAvailabilityOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if len(o.Requests) == 0 {
		return fmt.Errorf("at least one purchase option availability patch is required")
	}
	seen := map[string]struct{}{}
	seenProducts := map[OneTimeProductID]struct{}{}
	for _, request := range o.Requests {
		if _, err := NewOneTimeProductID(request.ProductID.String()); err != nil {
			return err
		}
		if _, err := NewOneTimeProductPurchaseOptionID(request.PurchaseOptionID.String()); err != nil {
			return err
		}
		if !isValidRegionCode(request.RegionCode) {
			return fmt.Errorf("purchase option availability region must be a two-letter ISO 3166 code")
		}
		if err := request.Availability.Validate(); err != nil {
			return err
		}
		key := purchaseOptionAvailabilityPatchKey(request.ProductID, request.PurchaseOptionID, request.RegionCode)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("purchase option availability %s is duplicated", key)
		}
		seen[key] = struct{}{}
		seenProducts[request.ProductID] = struct{}{}
	}
	if len(seenProducts) > 100 {
		return fmt.Errorf("purchase option availability batch patch cannot exceed 100 products")
	}
	if strings.TrimSpace(o.RegionsVersion) == "" {
		return fmt.Errorf("regions version is required")
	}
	if strings.TrimSpace(o.RegionsVersion) != o.RegionsVersion {
		return fmt.Errorf("regions version cannot have leading or trailing whitespace")
	}
	if _, err := NewProductUpdateLatencyTolerance(o.LatencyTolerance.String()); err != nil {
		return err
	}
	if o.Confirm && o.DryRun {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.Confirm && !o.DryRun {
		return fmt.Errorf("purchase option availability batch patch requires --confirm or --dry-run")
	}
	return nil
}

func (o PurchaseOptionBatchPatchAvailabilityOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live purchase option availability batch patch cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live purchase option availability batch patch requires --confirm")
	}
	return nil
}

func purchaseOptionAvailabilityPatchKey(productID OneTimeProductID, purchaseOptionID OneTimeProductPurchaseOptionID, regionCode string) string {
	return productID.String() + "/" + purchaseOptionID.String() + "/" + regionCode
}

func (o PurchaseOptionBatchPatchPriceOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if len(o.Requests) == 0 {
		return fmt.Errorf("at least one purchase option price patch is required")
	}
	seen := map[string]struct{}{}
	seenProducts := map[OneTimeProductID]struct{}{}
	for _, request := range o.Requests {
		if _, err := NewOneTimeProductID(request.ProductID.String()); err != nil {
			return err
		}
		if _, err := NewOneTimeProductPurchaseOptionID(request.PurchaseOptionID.String()); err != nil {
			return err
		}
		if !isValidRegionCode(request.RegionCode) {
			return fmt.Errorf("purchase option price region must be a two-letter ISO 3166 code")
		}
		if err := validateMoney(request.Price); err != nil {
			return fmt.Errorf("purchase option price for %s/%s/%s: %w", request.ProductID, request.PurchaseOptionID, request.RegionCode, err)
		}
		key := purchaseOptionPricePatchKey(request.ProductID, request.PurchaseOptionID, request.RegionCode)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("purchase option price %s is duplicated", key)
		}
		seen[key] = struct{}{}
		seenProducts[request.ProductID] = struct{}{}
	}
	if len(seenProducts) > 100 {
		return fmt.Errorf("purchase option price batch patch cannot exceed 100 products")
	}
	if strings.TrimSpace(o.RegionsVersion) == "" {
		return fmt.Errorf("regions version is required")
	}
	if strings.TrimSpace(o.RegionsVersion) != o.RegionsVersion {
		return fmt.Errorf("regions version cannot have leading or trailing whitespace")
	}
	if _, err := NewProductUpdateLatencyTolerance(o.LatencyTolerance.String()); err != nil {
		return err
	}
	if o.Confirm && o.DryRun {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.Confirm && !o.DryRun {
		return fmt.Errorf("purchase option price batch patch requires --confirm or --dry-run")
	}
	return nil
}

func validateMoney(money Money) error {
	currency, err := NewCurrencyCode(money.CurrencyCode)
	if err != nil {
		return err
	}
	if currency.String() != money.CurrencyCode {
		return fmt.Errorf("currency code must be normalized uppercase without surrounding whitespace")
	}
	if money.Units < 0 {
		return fmt.Errorf("price units must be 0 or greater")
	}
	if money.Nanos < 0 || money.Nanos > 999999999 {
		return fmt.Errorf("price nanos must be between 0 and 999999999")
	}
	if money.Units == 0 && money.Nanos == 0 {
		return fmt.Errorf("price must be greater than 0")
	}
	return nil
}

func (o PurchaseOptionBatchPatchPriceOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live purchase option price batch patch cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live purchase option price batch patch requires --confirm")
	}
	return nil
}

func purchaseOptionPricePatchKey(productID OneTimeProductID, purchaseOptionID OneTimeProductPurchaseOptionID, regionCode string) string {
	return productID.String() + "/" + purchaseOptionID.String() + "/" + regionCode
}

func (o PurchaseOptionStateUpdateOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live purchase option state update cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live purchase option state update requires --confirm")
	}
	return nil
}

type PurchaseOptionStateUpdatePlan struct {
	PackageName      PackageName                    `json:"packageName"`
	ProductID        OneTimeProductID               `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID `json:"purchaseOptionId"`
	Action           PurchaseOptionStateAction      `json:"action"`
	LatencyTolerance ProductUpdateLatencyTolerance  `json:"latencyTolerance"`
	Confirm          bool                           `json:"confirm"`
	Steps            []string                       `json:"steps"`
}

type PurchaseOptionStateUpdateResult struct {
	PackageName      PackageName                    `json:"packageName"`
	ProductID        OneTimeProductID               `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID `json:"purchaseOptionId"`
	Action           PurchaseOptionStateAction      `json:"action"`
	DryRun           bool                           `json:"dryRun"`
	Applied          bool                           `json:"applied"`
	Product          *OneTimeProduct                `json:"product,omitempty"`
	Plan             PurchaseOptionStateUpdatePlan  `json:"plan"`
}

type PurchaseOptionBatchPatchAvailabilityPlan struct {
	PackageName      PackageName                              `json:"packageName"`
	Requests         []PurchaseOptionAvailabilityPatchRequest `json:"requests"`
	UpdateMask       string                                   `json:"updateMask"`
	RegionsVersion   string                                   `json:"regionsVersion"`
	LatencyTolerance ProductUpdateLatencyTolerance            `json:"latencyTolerance"`
	Confirm          bool                                     `json:"confirm"`
	Steps            []string                                 `json:"steps"`
}

type PurchaseOptionBatchPatchAvailabilityDesiredProduct struct {
	PackageName     PackageName                                         `json:"packageName"`
	ProductID       OneTimeProductID                                    `json:"productId"`
	PurchaseOptions []PurchaseOptionBatchPatchAvailabilityDesiredOption `json:"purchaseOptions"`
}

type PurchaseOptionBatchPatchAvailabilityDesiredOption struct {
	PurchaseOptionID OneTimeProductPurchaseOptionID                      `json:"purchaseOptionId"`
	RegionalConfigs  []PurchaseOptionBatchPatchAvailabilityDesiredRegion `json:"regionalConfigs"`
}

type PurchaseOptionBatchPatchAvailabilityDesiredRegion struct {
	RegionCode   string                     `json:"regionCode"`
	Availability PurchaseOptionAvailability `json:"availability"`
}

type PurchaseOptionBatchPatchAvailabilityResult struct {
	PackageName PackageName                                          `json:"packageName"`
	Requests    []PurchaseOptionAvailabilityPatchRequest             `json:"requests"`
	DryRun      bool                                                 `json:"dryRun"`
	Applied     bool                                                 `json:"applied"`
	Products    []OneTimeProduct                                     `json:"products,omitempty"`
	Desired     []PurchaseOptionBatchPatchAvailabilityDesiredProduct `json:"desiredProducts"`
	Plan        PurchaseOptionBatchPatchAvailabilityPlan             `json:"plan"`
}

type PurchaseOptionBatchPatchPricePlan struct {
	PackageName      PackageName                       `json:"packageName"`
	Requests         []PurchaseOptionPricePatchRequest `json:"requests"`
	UpdateMask       string                            `json:"updateMask"`
	RegionsVersion   string                            `json:"regionsVersion"`
	LatencyTolerance ProductUpdateLatencyTolerance     `json:"latencyTolerance"`
	Confirm          bool                              `json:"confirm"`
	Steps            []string                          `json:"steps"`
}

type PurchaseOptionBatchPatchPriceDesiredProduct struct {
	PackageName     PackageName                                  `json:"packageName"`
	ProductID       OneTimeProductID                             `json:"productId"`
	PurchaseOptions []PurchaseOptionBatchPatchPriceDesiredOption `json:"purchaseOptions"`
}

type PurchaseOptionBatchPatchPriceDesiredOption struct {
	PurchaseOptionID OneTimeProductPurchaseOptionID               `json:"purchaseOptionId"`
	RegionalConfigs  []PurchaseOptionBatchPatchPriceDesiredRegion `json:"regionalConfigs"`
}

type PurchaseOptionBatchPatchPriceDesiredRegion struct {
	RegionCode string `json:"regionCode"`
	Price      Money  `json:"price"`
}

type PurchaseOptionBatchPatchPriceResult struct {
	PackageName PackageName                                   `json:"packageName"`
	Requests    []PurchaseOptionPricePatchRequest             `json:"requests"`
	DryRun      bool                                          `json:"dryRun"`
	Applied     bool                                          `json:"applied"`
	Products    []OneTimeProduct                              `json:"products,omitempty"`
	Desired     []PurchaseOptionBatchPatchPriceDesiredProduct `json:"desiredProducts"`
	Plan        PurchaseOptionBatchPatchPricePlan             `json:"plan"`
}

type PurchaseOptionStateUpdater interface {
	UpdatePurchaseOptionState(ctx context.Context, options PurchaseOptionStateUpdateOptions) (OneTimeProduct, error)
}

func NewPurchaseOptionStateUpdatePlan(options PurchaseOptionStateUpdateOptions) (PurchaseOptionStateUpdatePlan, error) {
	if err := options.Validate(); err != nil {
		return PurchaseOptionStateUpdatePlan{}, err
	}
	return PurchaseOptionStateUpdatePlan{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		PurchaseOptionID: options.PurchaseOptionID,
		Action:           options.Action,
		LatencyTolerance: options.LatencyTolerance,
		Confirm:          options.Confirm,
		Steps:            purchaseOptionStateUpdateSteps(options),
	}, nil
}

func UpdatePurchaseOptionState(ctx context.Context, updater PurchaseOptionStateUpdater, options PurchaseOptionStateUpdateOptions) (PurchaseOptionStateUpdateResult, error) {
	plan, err := NewPurchaseOptionStateUpdatePlan(options)
	if err != nil {
		return PurchaseOptionStateUpdateResult{}, err
	}
	result := PurchaseOptionStateUpdateResult{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		PurchaseOptionID: options.PurchaseOptionID,
		Action:           options.Action,
		DryRun:           options.DryRun,
		Applied:          false,
		Plan:             plan,
	}
	if options.DryRun {
		return result, nil
	}
	if updater == nil {
		return PurchaseOptionStateUpdateResult{}, fmt.Errorf("purchase option state updater is required")
	}
	product, err := updater.UpdatePurchaseOptionState(ctx, options)
	if err != nil {
		return PurchaseOptionStateUpdateResult{}, err
	}
	result.Applied = true
	result.Product = &product
	return result, nil
}

func NewPurchaseOptionBatchPatchAvailabilityPlan(options PurchaseOptionBatchPatchAvailabilityOptions) (PurchaseOptionBatchPatchAvailabilityPlan, error) {
	if err := options.Validate(); err != nil {
		return PurchaseOptionBatchPatchAvailabilityPlan{}, err
	}
	return PurchaseOptionBatchPatchAvailabilityPlan{
		PackageName:      options.PackageName,
		Requests:         append([]PurchaseOptionAvailabilityPatchRequest(nil), options.Requests...),
		UpdateMask:       purchaseOptionAvailabilityUpdateMask,
		RegionsVersion:   options.RegionsVersion,
		LatencyTolerance: options.LatencyTolerance,
		Confirm:          options.Confirm,
		Steps:            purchaseOptionBatchPatchAvailabilitySteps(options),
	}, nil
}

func BatchPatchPurchaseOptionAvailability(ctx context.Context, patcher PurchaseOptionBatchAvailabilityPatcher, options PurchaseOptionBatchPatchAvailabilityOptions) (PurchaseOptionBatchPatchAvailabilityResult, error) {
	plan, err := NewPurchaseOptionBatchPatchAvailabilityPlan(options)
	if err != nil {
		return PurchaseOptionBatchPatchAvailabilityResult{}, err
	}
	requests := append([]PurchaseOptionAvailabilityPatchRequest(nil), options.Requests...)
	result := PurchaseOptionBatchPatchAvailabilityResult{
		PackageName: options.PackageName,
		Requests:    requests,
		DryRun:      options.DryRun,
		Desired:     desiredOneTimeProductsForPurchaseOptionAvailabilityPatch(options),
		Plan:        plan,
	}
	if options.DryRun {
		return result, nil
	}
	if patcher == nil {
		return PurchaseOptionBatchPatchAvailabilityResult{}, fmt.Errorf("purchase option availability batch patcher is required")
	}
	updated, err := patcher.BatchPatchPurchaseOptionAvailability(ctx, options)
	if err != nil {
		return PurchaseOptionBatchPatchAvailabilityResult{}, err
	}
	updated.PackageName = options.PackageName
	updated.Requests = requests
	updated.DryRun = false
	updated.Applied = true
	updated.Desired = result.Desired
	updated.Plan = plan
	return updated, nil
}

func NewPurchaseOptionBatchPatchPricePlan(options PurchaseOptionBatchPatchPriceOptions) (PurchaseOptionBatchPatchPricePlan, error) {
	if err := options.Validate(); err != nil {
		return PurchaseOptionBatchPatchPricePlan{}, err
	}
	return PurchaseOptionBatchPatchPricePlan{
		PackageName:      options.PackageName,
		Requests:         append([]PurchaseOptionPricePatchRequest(nil), options.Requests...),
		UpdateMask:       purchaseOptionAvailabilityUpdateMask,
		RegionsVersion:   options.RegionsVersion,
		LatencyTolerance: options.LatencyTolerance,
		Confirm:          options.Confirm,
		Steps:            purchaseOptionBatchPatchPriceSteps(options),
	}, nil
}

func BatchPatchPurchaseOptionPrices(ctx context.Context, patcher PurchaseOptionBatchPricePatcher, options PurchaseOptionBatchPatchPriceOptions) (PurchaseOptionBatchPatchPriceResult, error) {
	plan, err := NewPurchaseOptionBatchPatchPricePlan(options)
	if err != nil {
		return PurchaseOptionBatchPatchPriceResult{}, err
	}
	requests := append([]PurchaseOptionPricePatchRequest(nil), options.Requests...)
	result := PurchaseOptionBatchPatchPriceResult{
		PackageName: options.PackageName,
		Requests:    requests,
		DryRun:      options.DryRun,
		Desired:     desiredOneTimeProductsForPurchaseOptionPricePatch(options),
		Plan:        plan,
	}
	if options.DryRun {
		return result, nil
	}
	if patcher == nil {
		return PurchaseOptionBatchPatchPriceResult{}, fmt.Errorf("purchase option price batch patcher is required")
	}
	updated, err := patcher.BatchPatchPurchaseOptionPrices(ctx, options)
	if err != nil {
		return PurchaseOptionBatchPatchPriceResult{}, err
	}
	updated.PackageName = options.PackageName
	updated.Requests = requests
	updated.DryRun = false
	updated.Applied = true
	updated.Desired = result.Desired
	updated.Plan = plan
	return updated, nil
}

func desiredOneTimeProductsForPurchaseOptionAvailabilityPatch(options PurchaseOptionBatchPatchAvailabilityOptions) []PurchaseOptionBatchPatchAvailabilityDesiredProduct {
	byProduct := map[OneTimeProductID]int{}
	byOption := map[string]int{}
	products := make([]PurchaseOptionBatchPatchAvailabilityDesiredProduct, 0)
	for _, request := range options.Requests {
		productIndex, ok := byProduct[request.ProductID]
		if !ok {
			byProduct[request.ProductID] = len(products)
			products = append(products, PurchaseOptionBatchPatchAvailabilityDesiredProduct{
				PackageName:     options.PackageName,
				ProductID:       request.ProductID,
				PurchaseOptions: []PurchaseOptionBatchPatchAvailabilityDesiredOption{},
			})
			productIndex = len(products) - 1
		}
		optionKey := request.ProductID.String() + "/" + request.PurchaseOptionID.String()
		optionIndex, ok := byOption[optionKey]
		if !ok {
			byOption[optionKey] = len(products[productIndex].PurchaseOptions)
			products[productIndex].PurchaseOptions = append(products[productIndex].PurchaseOptions, PurchaseOptionBatchPatchAvailabilityDesiredOption{
				PurchaseOptionID: request.PurchaseOptionID,
				RegionalConfigs:  []PurchaseOptionBatchPatchAvailabilityDesiredRegion{},
			})
			optionIndex = len(products[productIndex].PurchaseOptions) - 1
		}
		products[productIndex].PurchaseOptions[optionIndex].RegionalConfigs = append(products[productIndex].PurchaseOptions[optionIndex].RegionalConfigs, PurchaseOptionBatchPatchAvailabilityDesiredRegion{
			RegionCode:   request.RegionCode,
			Availability: request.Availability,
		})
	}
	return products
}

func desiredOneTimeProductsForPurchaseOptionPricePatch(options PurchaseOptionBatchPatchPriceOptions) []PurchaseOptionBatchPatchPriceDesiredProduct {
	byProduct := map[OneTimeProductID]int{}
	byOption := map[string]int{}
	products := make([]PurchaseOptionBatchPatchPriceDesiredProduct, 0)
	for _, request := range options.Requests {
		productIndex, ok := byProduct[request.ProductID]
		if !ok {
			byProduct[request.ProductID] = len(products)
			products = append(products, PurchaseOptionBatchPatchPriceDesiredProduct{
				PackageName:     options.PackageName,
				ProductID:       request.ProductID,
				PurchaseOptions: []PurchaseOptionBatchPatchPriceDesiredOption{},
			})
			productIndex = len(products) - 1
		}
		optionKey := request.ProductID.String() + "/" + request.PurchaseOptionID.String()
		optionIndex, ok := byOption[optionKey]
		if !ok {
			byOption[optionKey] = len(products[productIndex].PurchaseOptions)
			products[productIndex].PurchaseOptions = append(products[productIndex].PurchaseOptions, PurchaseOptionBatchPatchPriceDesiredOption{
				PurchaseOptionID: request.PurchaseOptionID,
				RegionalConfigs:  []PurchaseOptionBatchPatchPriceDesiredRegion{},
			})
			optionIndex = len(products[productIndex].PurchaseOptions) - 1
		}
		products[productIndex].PurchaseOptions[optionIndex].RegionalConfigs = append(products[productIndex].PurchaseOptions[optionIndex].RegionalConfigs, PurchaseOptionBatchPatchPriceDesiredRegion{
			RegionCode: request.RegionCode,
			Price:      request.Price,
		})
	}
	return products
}

func purchaseOptionStateUpdateSteps(options PurchaseOptionStateUpdateOptions) []string {
	if options.DryRun {
		return []string{fmt.Sprintf("plan %s purchase option", options.Action)}
	}
	return []string{fmt.Sprintf("%s purchase option", options.Action)}
}

const purchaseOptionAvailabilityUpdateMask = "purchaseOptions"

func purchaseOptionBatchPatchAvailabilitySteps(options PurchaseOptionBatchPatchAvailabilityOptions) []string {
	if options.DryRun {
		return []string{"plan purchase option availability batch patch"}
	}
	return []string{"fetch current one-time products", "merge purchase option regional availability", "batch patch purchase option availability"}
}

func purchaseOptionBatchPatchPriceSteps(options PurchaseOptionBatchPatchPriceOptions) []string {
	if options.DryRun {
		return []string{"plan purchase option price batch patch"}
	}
	return []string{"fetch current one-time products", "merge purchase option regional prices", "batch patch purchase option prices"}
}
