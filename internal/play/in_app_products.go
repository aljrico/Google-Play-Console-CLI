package play

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

type InAppProductSKU string

func NewInAppProductSKU(value string) (InAppProductSKU, error) {
	if value == "" {
		return "", fmt.Errorf("SKU is required")
	}
	return InAppProductSKU(value), nil
}

func (s InAppProductSKU) String() string {
	return string(s)
}

type ProductPurchaseType string

const (
	ProductPurchaseTypeManagedUser  ProductPurchaseType = "managedUser"
	ProductPurchaseTypeSubscription ProductPurchaseType = "subscription"
	ProductPurchaseTypeUnspecified  ProductPurchaseType = "purchaseTypeUnspecified"
)

func (t ProductPurchaseType) String() string {
	return string(t)
}

type ProductStatus string

const (
	ProductStatusActive      ProductStatus = "active"
	ProductStatusInactive    ProductStatus = "inactive"
	ProductStatusUnspecified ProductStatus = "statusUnspecified"
)

func NewProductStatus(value string) (ProductStatus, error) {
	status := ProductStatus(strings.TrimSpace(value))
	switch status {
	case ProductStatusActive, ProductStatusInactive:
		return status, nil
	default:
		return "", fmt.Errorf("unsupported in-app product status %q; supported values: active, inactive", value)
	}
}

func (s ProductStatus) String() string {
	return string(s)
}

func (s ProductStatus) Validate() error {
	if strings.TrimSpace(s.String()) != s.String() {
		return fmt.Errorf("unsupported in-app product status %q; supported values: active, inactive", s.String())
	}
	_, err := NewProductStatus(s.String())
	return err
}

type ProductPrice struct {
	Currency    string `json:"currency,omitempty"`
	PriceMicros string `json:"priceMicros,omitempty"`
}

type RegionalProductPrice struct {
	RegionCode string       `json:"regionCode"`
	Price      ProductPrice `json:"price"`
}

func NewProductPrice(value string) (ProductPrice, error) {
	currency, priceMicros, ok := strings.Cut(strings.TrimSpace(value), ":")
	if !ok {
		return ProductPrice{}, fmt.Errorf("price must be formatted as CURRENCY:MICROS")
	}
	price := ProductPrice{Currency: strings.ToUpper(strings.TrimSpace(currency)), PriceMicros: strings.TrimSpace(priceMicros)}
	if err := price.Validate(); err != nil {
		return ProductPrice{}, err
	}
	return price, nil
}

func (p ProductPrice) Validate() error {
	if len(p.Currency) != 3 {
		return fmt.Errorf("price currency must be a three-letter ISO 4217 code")
	}
	for _, character := range p.Currency {
		if character < 'A' || character > 'Z' {
			return fmt.Errorf("price currency must be uppercase ISO 4217")
		}
	}
	if p.Currency != strings.ToUpper(p.Currency) || strings.TrimSpace(p.Currency) != p.Currency {
		return fmt.Errorf("price currency must be uppercase ISO 4217")
	}
	if p.PriceMicros == "" {
		return fmt.Errorf("price micros is required")
	}
	micros, err := strconv.ParseInt(p.PriceMicros, 10, 64)
	if err != nil || micros <= 0 {
		return fmt.Errorf("price micros must be a positive integer")
	}
	return nil
}

func NewRegionalProductPrice(value string) (RegionalProductPrice, error) {
	region, rawPrice, ok := strings.Cut(strings.TrimSpace(value), ":")
	if !ok {
		return RegionalProductPrice{}, fmt.Errorf("regional price must be formatted as REGION:CURRENCY:MICROS")
	}
	regionCode := strings.ToUpper(strings.TrimSpace(region))
	if !isValidRegionCode(regionCode) {
		return RegionalProductPrice{}, fmt.Errorf("regional price region must be a two-letter ISO 3166 code")
	}
	price, err := NewProductPrice(rawPrice)
	if err != nil {
		return RegionalProductPrice{}, err
	}
	return RegionalProductPrice{RegionCode: regionCode, Price: price}, nil
}

type InAppProductListing struct {
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Benefits    []string `json:"benefits,omitempty"`
}

const (
	inAppProductListingTitleLimit       = 55
	inAppProductListingDescriptionLimit = 200
)

type InAppProduct struct {
	PackageName                            PackageName                    `json:"packageName"`
	SKU                                    InAppProductSKU                `json:"sku"`
	Status                                 ProductStatus                  `json:"status,omitempty"`
	PurchaseType                           ProductPurchaseType            `json:"purchaseType,omitempty"`
	DefaultLanguage                        string                         `json:"defaultLanguage,omitempty"`
	DefaultPrice                           *ProductPrice                  `json:"defaultPrice,omitempty"`
	Prices                                 map[string]ProductPrice        `json:"prices,omitempty"`
	Listings                               map[string]InAppProductListing `json:"listings,omitempty"`
	SubscriptionPeriod                     string                         `json:"subscriptionPeriod,omitempty"`
	TrialPeriod                            string                         `json:"trialPeriod,omitempty"`
	GracePeriod                            string                         `json:"gracePeriod,omitempty"`
	ManagedProductTaxAndComplianceSettings *ProductTaxComplianceSettings  `json:"managedProductTaxAndComplianceSettings,omitempty"`
	SubscriptionTaxAndComplianceSettings   *ProductTaxComplianceSettings  `json:"subscriptionTaxAndComplianceSettings,omitempty"`
}

type ProductTaxComplianceSettings struct {
	EEAWithdrawalRightType  string                         `json:"eeaWithdrawalRightType,omitempty"`
	IsTokenizedDigitalAsset bool                           `json:"isTokenizedDigitalAsset,omitempty"`
	TaxRateInfoByRegionCode map[string]RegionalTaxRateInfo `json:"taxRateInfoByRegionCode,omitempty"`
}

type RegionalTaxRateInfo struct {
	EligibleForStreamingServiceTaxRate bool   `json:"eligibleForStreamingServiceTaxRate,omitempty"`
	StreamingTaxType                   string `json:"streamingTaxType,omitempty"`
	TaxTier                            string `json:"taxTier,omitempty"`
}

type InAppProductPagination struct {
	NextPageToken     string `json:"nextPageToken,omitempty"`
	PreviousPageToken string `json:"previousPageToken,omitempty"`
}

type InAppProductListOptions struct {
	PackageName PackageName `json:"packageName"`
	Token       string      `json:"token,omitempty"`
}

func (o InAppProductListOptions) Validate() error {
	return o.PackageName.Validate()
}

type InAppProductListResult struct {
	PackageName PackageName             `json:"packageName"`
	Products    []InAppProduct          `json:"products"`
	Pagination  *InAppProductPagination `json:"pagination,omitempty"`
	Options     InAppProductListOptions `json:"options"`
}

type InAppProductBatchGetOptions struct {
	PackageName PackageName       `json:"packageName"`
	SKUs        []InAppProductSKU `json:"skus"`
}

func (o InAppProductBatchGetOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if len(o.SKUs) == 0 {
		return fmt.Errorf("at least one in-app product SKU is required")
	}
	seen := map[InAppProductSKU]struct{}{}
	for _, sku := range o.SKUs {
		if _, err := NewInAppProductSKU(sku.String()); err != nil {
			return err
		}
		if _, ok := seen[sku]; ok {
			return fmt.Errorf("in-app product SKU %q is duplicated", sku)
		}
		seen[sku] = struct{}{}
	}
	return nil
}

type InAppProductBatchGetResult struct {
	PackageName PackageName                 `json:"packageName"`
	Products    []InAppProduct              `json:"products"`
	Options     InAppProductBatchGetOptions `json:"options"`
}

type InAppProductDeleteOptions struct {
	PackageName      PackageName                   `json:"packageName"`
	SKU              InAppProductSKU               `json:"sku"`
	LatencyTolerance ProductUpdateLatencyTolerance `json:"latencyTolerance"`
	Confirm          bool                          `json:"confirm"`
	DryRun           bool                          `json:"dryRun"`
}

type InAppProductBatchDeleteOptions struct {
	PackageName      PackageName                   `json:"packageName"`
	SKUs             []InAppProductSKU             `json:"skus"`
	LatencyTolerance ProductUpdateLatencyTolerance `json:"latencyTolerance"`
	Confirm          bool                          `json:"confirm"`
	DryRun           bool                          `json:"dryRun"`
}

func (o InAppProductDeleteOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewInAppProductSKU(o.SKU.String()); err != nil {
		return err
	}
	if _, err := NewProductUpdateLatencyTolerance(o.LatencyTolerance.String()); err != nil {
		return err
	}
	if o.Confirm && o.DryRun {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.Confirm && !o.DryRun {
		return fmt.Errorf("in-app product deletion requires --confirm or --dry-run")
	}
	return nil
}

func (o InAppProductBatchDeleteOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if len(o.SKUs) == 0 {
		return fmt.Errorf("at least one in-app product SKU is required")
	}
	if len(o.SKUs) > 100 {
		return fmt.Errorf("in-app product batch-delete cannot exceed 100 SKUs")
	}
	seen := map[InAppProductSKU]struct{}{}
	for _, sku := range o.SKUs {
		if _, err := NewInAppProductSKU(sku.String()); err != nil {
			return err
		}
		if _, ok := seen[sku]; ok {
			return fmt.Errorf("in-app product SKU %q is duplicated", sku)
		}
		seen[sku] = struct{}{}
	}
	if _, err := NewProductUpdateLatencyTolerance(o.LatencyTolerance.String()); err != nil {
		return err
	}
	if o.Confirm && o.DryRun {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.Confirm && !o.DryRun {
		return fmt.Errorf("in-app product batch deletion requires --confirm or --dry-run")
	}
	return nil
}

func (o InAppProductDeleteOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live in-app product deletion cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live in-app product deletion requires --confirm")
	}
	return nil
}

func (o InAppProductBatchDeleteOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live in-app product batch deletion cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live in-app product batch deletion requires --confirm")
	}
	return nil
}

type InAppProductDeletePlan struct {
	PackageName      PackageName                   `json:"packageName"`
	SKU              InAppProductSKU               `json:"sku"`
	LatencyTolerance ProductUpdateLatencyTolerance `json:"latencyTolerance"`
	Confirm          bool                          `json:"confirm"`
	Steps            []string                      `json:"steps"`
}

type InAppProductBatchDeletePlan struct {
	PackageName      PackageName                   `json:"packageName"`
	SKUs             []InAppProductSKU             `json:"skus"`
	LatencyTolerance ProductUpdateLatencyTolerance `json:"latencyTolerance"`
	Confirm          bool                          `json:"confirm"`
	Steps            []string                      `json:"steps"`
}

type InAppProductDeleteResult struct {
	PackageName PackageName            `json:"packageName"`
	SKU         InAppProductSKU        `json:"sku"`
	DryRun      bool                   `json:"dryRun"`
	Deleted     bool                   `json:"deleted"`
	Plan        InAppProductDeletePlan `json:"plan"`
}

type InAppProductBatchDeleteResult struct {
	PackageName PackageName                 `json:"packageName"`
	SKUs        []InAppProductSKU           `json:"skus"`
	DryRun      bool                        `json:"dryRun"`
	Deleted     bool                        `json:"deleted"`
	Plan        InAppProductBatchDeletePlan `json:"plan"`
}

type InAppProductCreateOptions struct {
	PackageName     PackageName         `json:"packageName"`
	SKU             InAppProductSKU     `json:"sku"`
	Status          ProductStatus       `json:"status"`
	DefaultLanguage ListingLanguage     `json:"defaultLanguage"`
	DefaultPrice    ProductPrice        `json:"defaultPrice"`
	Listing         InAppProductListing `json:"listing"`
	Confirm         bool                `json:"confirm"`
	DryRun          bool                `json:"dryRun"`
}

func (o InAppProductCreateOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if err := validateManagedProductSKU(o.SKU); err != nil {
		return err
	}
	if err := o.Status.Validate(); err != nil {
		return err
	}
	if o.Status != ProductStatusActive && o.Status != ProductStatusInactive {
		return fmt.Errorf("in-app product create status must be active or inactive")
	}
	if _, err := NewListingLanguage(o.DefaultLanguage.String()); err != nil {
		return err
	}
	if err := o.DefaultPrice.Validate(); err != nil {
		return err
	}
	if err := validateRequiredInAppProductListing(o.Listing); err != nil {
		return err
	}
	if o.DryRun && o.Confirm {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.DryRun && !o.Confirm {
		return fmt.Errorf("in-app product create requires --confirm or --dry-run")
	}
	return nil
}

func (o InAppProductCreateOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live in-app product create cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live in-app product create requires --confirm")
	}
	return nil
}

type InAppProductCreatePlan struct {
	Action                   string              `json:"action"`
	PackageName              PackageName         `json:"packageName"`
	SKU                      InAppProductSKU     `json:"sku"`
	Status                   ProductStatus       `json:"status"`
	PurchaseType             ProductPurchaseType `json:"purchaseType"`
	DefaultLanguage          ListingLanguage     `json:"defaultLanguage"`
	DefaultPrice             ProductPrice        `json:"defaultPrice"`
	AutoConvertMissingPrices bool                `json:"autoConvertMissingPrices"`
	Confirm                  bool                `json:"confirm"`
	Steps                    []string            `json:"steps"`
}

type InAppProductCreateResult struct {
	Action  string                 `json:"action"`
	DryRun  bool                   `json:"dryRun"`
	Created bool                   `json:"created"`
	Product *InAppProduct          `json:"product,omitempty"`
	Desired InAppProduct           `json:"desiredProduct"`
	Plan    InAppProductCreatePlan `json:"plan"`
}

type InAppProductCreator interface {
	CreateInAppProduct(ctx context.Context, options InAppProductCreateOptions) (InAppProduct, error)
}

func CreateInAppProduct(ctx context.Context, creator InAppProductCreator, options InAppProductCreateOptions) (InAppProductCreateResult, error) {
	if err := options.Validate(); err != nil {
		return InAppProductCreateResult{}, err
	}
	desired := inAppProductCreateDesiredProduct(options)
	result := InAppProductCreateResult{
		Action:  "create",
		DryRun:  options.DryRun,
		Desired: desired,
		Plan: InAppProductCreatePlan{
			Action:                   "create",
			PackageName:              options.PackageName,
			SKU:                      options.SKU,
			Status:                   options.Status,
			PurchaseType:             ProductPurchaseTypeManagedUser,
			DefaultLanguage:          options.DefaultLanguage,
			DefaultPrice:             options.DefaultPrice,
			AutoConvertMissingPrices: true,
			Confirm:                  options.Confirm,
			Steps:                    inAppProductCreateSteps(options.DryRun),
		},
	}
	if options.DryRun {
		return result, nil
	}
	if creator == nil {
		return InAppProductCreateResult{}, fmt.Errorf("in-app product creator is required")
	}
	product, err := creator.CreateInAppProduct(ctx, options)
	if err != nil {
		return InAppProductCreateResult{}, err
	}
	result.Created = true
	result.Product = &product
	return result, nil
}

func inAppProductCreateDesiredProduct(options InAppProductCreateOptions) InAppProduct {
	return InAppProduct{
		PackageName:     options.PackageName,
		SKU:             options.SKU,
		Status:          options.Status,
		PurchaseType:    ProductPurchaseTypeManagedUser,
		DefaultLanguage: options.DefaultLanguage.String(),
		DefaultPrice:    &options.DefaultPrice,
		Listings: map[string]InAppProductListing{
			options.DefaultLanguage.String(): options.Listing,
		},
	}
}

func inAppProductCreateSteps(dryRun bool) []string {
	if dryRun {
		return []string{"plan managed in-app product creation"}
	}
	return []string{"create managed in-app product"}
}

func validateManagedProductSKU(sku InAppProductSKU) error {
	value := sku.String()
	if _, err := NewInAppProductSKU(value); err != nil {
		return err
	}
	if strings.HasPrefix(value, "android.test") || !isValidOneTimeProductID(value) {
		return fmt.Errorf("invalid managed product SKU %q", value)
	}
	return nil
}

type InAppProductLister interface {
	ListInAppProducts(ctx context.Context, options InAppProductListOptions) (InAppProductListResult, error)
}

func ListInAppProducts(ctx context.Context, lister InAppProductLister, options InAppProductListOptions) (InAppProductListResult, error) {
	if err := options.Validate(); err != nil {
		return InAppProductListResult{}, err
	}
	if lister == nil {
		return InAppProductListResult{}, fmt.Errorf("in-app product lister is required")
	}
	return lister.ListInAppProducts(ctx, options)
}

type InAppProductGetOptions struct {
	PackageName PackageName     `json:"packageName"`
	SKU         InAppProductSKU `json:"sku"`
}

func (o InAppProductGetOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewInAppProductSKU(o.SKU.String()); err != nil {
		return err
	}
	return nil
}

type InAppProductGetter interface {
	GetInAppProduct(ctx context.Context, packageName PackageName, sku InAppProductSKU) (InAppProduct, error)
}

type InAppProductBatchGetter interface {
	BatchGetInAppProducts(ctx context.Context, options InAppProductBatchGetOptions) (InAppProductBatchGetResult, error)
}

type InAppProductDeleter interface {
	DeleteInAppProduct(ctx context.Context, options InAppProductDeleteOptions) error
}

type InAppProductBatchDeleter interface {
	BatchDeleteInAppProducts(ctx context.Context, options InAppProductBatchDeleteOptions) error
}

func GetInAppProduct(ctx context.Context, getter InAppProductGetter, options InAppProductGetOptions) (InAppProduct, error) {
	if err := options.Validate(); err != nil {
		return InAppProduct{}, err
	}
	if getter == nil {
		return InAppProduct{}, fmt.Errorf("in-app product getter is required")
	}
	return getter.GetInAppProduct(ctx, options.PackageName, options.SKU)
}

func BatchGetInAppProducts(ctx context.Context, getter InAppProductBatchGetter, options InAppProductBatchGetOptions) (InAppProductBatchGetResult, error) {
	if err := options.Validate(); err != nil {
		return InAppProductBatchGetResult{}, err
	}
	if getter == nil {
		return InAppProductBatchGetResult{}, fmt.Errorf("in-app product batch getter is required")
	}
	return getter.BatchGetInAppProducts(ctx, options)
}

func DeleteInAppProduct(ctx context.Context, deleter InAppProductDeleter, options InAppProductDeleteOptions) (InAppProductDeleteResult, error) {
	if err := options.Validate(); err != nil {
		return InAppProductDeleteResult{}, err
	}
	result := InAppProductDeleteResult{
		PackageName: options.PackageName,
		SKU:         options.SKU,
		DryRun:      options.DryRun,
		Plan: InAppProductDeletePlan{
			PackageName:      options.PackageName,
			SKU:              options.SKU,
			LatencyTolerance: options.LatencyTolerance,
			Confirm:          options.Confirm,
			Steps:            inAppProductDeleteSteps(options),
		},
	}
	if options.DryRun {
		return result, nil
	}
	if deleter == nil {
		return InAppProductDeleteResult{}, fmt.Errorf("in-app product deleter is required")
	}
	getter, ok := deleter.(InAppProductGetter)
	if !ok {
		return InAppProductDeleteResult{}, fmt.Errorf("in-app product getter is required for live deletion preflight")
	}
	current, err := getter.GetInAppProduct(ctx, options.PackageName, options.SKU)
	if err != nil {
		return InAppProductDeleteResult{}, err
	}
	if current.PurchaseType != ProductPurchaseTypeManagedUser {
		return InAppProductDeleteResult{}, fmt.Errorf("only managed in-app products can be deleted with in-app-products; got purchase type %q", current.PurchaseType)
	}
	if err := deleter.DeleteInAppProduct(ctx, options); err != nil {
		return InAppProductDeleteResult{}, err
	}
	result.Deleted = true
	return result, nil
}

func BatchDeleteInAppProducts(ctx context.Context, deleter InAppProductBatchDeleter, options InAppProductBatchDeleteOptions) (InAppProductBatchDeleteResult, error) {
	if err := options.Validate(); err != nil {
		return InAppProductBatchDeleteResult{}, err
	}
	skus := append([]InAppProductSKU(nil), options.SKUs...)
	result := InAppProductBatchDeleteResult{
		PackageName: options.PackageName,
		SKUs:        skus,
		DryRun:      options.DryRun,
		Plan: InAppProductBatchDeletePlan{
			PackageName:      options.PackageName,
			SKUs:             skus,
			LatencyTolerance: options.LatencyTolerance,
			Confirm:          options.Confirm,
			Steps:            inAppProductBatchDeleteSteps(options),
		},
	}
	if options.DryRun {
		return result, nil
	}
	if deleter == nil {
		return InAppProductBatchDeleteResult{}, fmt.Errorf("in-app product batch deleter is required")
	}
	getter, ok := deleter.(InAppProductBatchGetter)
	if !ok {
		return InAppProductBatchDeleteResult{}, fmt.Errorf("in-app product batch getter is required for live batch deletion preflight")
	}
	products, err := getter.BatchGetInAppProducts(ctx, InAppProductBatchGetOptions{
		PackageName: options.PackageName,
		SKUs:        skus,
	})
	if err != nil {
		return InAppProductBatchDeleteResult{}, err
	}
	if err := validateInAppProductsBatchDeletePreflight(skus, products.Products); err != nil {
		return InAppProductBatchDeleteResult{}, err
	}
	if err := deleter.BatchDeleteInAppProducts(ctx, options); err != nil {
		return InAppProductBatchDeleteResult{}, err
	}
	result.Deleted = true
	return result, nil
}

func inAppProductDeleteSteps(options InAppProductDeleteOptions) []string {
	if options.DryRun {
		return []string{"plan in-app product deletion"}
	}
	return []string{"fetch current in-app product", "delete in-app product"}
}

func inAppProductBatchDeleteSteps(options InAppProductBatchDeleteOptions) []string {
	if options.DryRun {
		return []string{"plan in-app product batch deletion"}
	}
	return []string{"batch fetch current in-app products", "batch delete in-app products"}
}

func validateInAppProductsBatchDeletePreflight(requestedSKUs []InAppProductSKU, products []InAppProduct) error {
	bySKU := make(map[InAppProductSKU]InAppProduct, len(products))
	for _, product := range products {
		if product.SKU != "" {
			bySKU[product.SKU] = product
		}
	}
	for _, sku := range requestedSKUs {
		product, ok := bySKU[sku]
		if !ok {
			return fmt.Errorf("in-app product %q was not returned by live deletion preflight", sku)
		}
		if product.PurchaseType != ProductPurchaseTypeManagedUser {
			return fmt.Errorf("only managed in-app products can be deleted with in-app-products; SKU %q has purchase type %q", sku, product.PurchaseType)
		}
	}
	return nil
}

type InAppProductPatchOptions struct {
	PackageName              PackageName            `json:"packageName"`
	SKU                      InAppProductSKU        `json:"sku"`
	Status                   ProductStatus          `json:"status,omitempty"`
	DefaultLanguage          ListingLanguage        `json:"defaultLanguage,omitempty"`
	DefaultPrice             *ProductPrice          `json:"defaultPrice,omitempty"`
	RegionalPrices           []RegionalProductPrice `json:"regionalPrices,omitempty"`
	ListingLanguage          ListingLanguage        `json:"listingLanguage,omitempty"`
	Listing                  *InAppProductListing   `json:"listing,omitempty"`
	AutoConvertMissingPrices bool                   `json:"autoConvertMissingPrices"`
	Confirm                  bool                   `json:"confirm"`
	DryRun                   bool                   `json:"dryRun"`
}

func (o InAppProductPatchOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewInAppProductSKU(o.SKU.String()); err != nil {
		return err
	}
	if o.Status != "" {
		if err := o.Status.Validate(); err != nil {
			return err
		}
	}
	if o.DefaultLanguage != "" {
		if _, err := NewListingLanguage(o.DefaultLanguage.String()); err != nil {
			return err
		}
	}
	if o.DefaultPrice != nil {
		if err := o.DefaultPrice.Validate(); err != nil {
			return err
		}
	}
	if err := validateRegionalProductPrices(o.RegionalPrices); err != nil {
		return err
	}
	if o.ListingLanguage != "" {
		if _, err := NewListingLanguage(o.ListingLanguage.String()); err != nil {
			return err
		}
	}
	if o.Listing != nil {
		if o.ListingLanguage == "" {
			return fmt.Errorf("listing patch requires --listing-language")
		}
		if err := validateRequiredInAppProductListing(*o.Listing); err != nil {
			return err
		}
	} else if o.ListingLanguage != "" {
		return fmt.Errorf("--listing-language requires --title and --description")
	}
	if !o.HasMutation() {
		return fmt.Errorf("in-app product patch requires at least one of --status, --default-price, --regional-price, --default-language, --listing-language, --title, or --description")
	}
	if o.DryRun && o.Confirm {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.DryRun && !o.Confirm {
		return fmt.Errorf("in-app product patch requires --confirm or --dry-run")
	}
	return nil
}

func (o InAppProductPatchOptions) HasMutation() bool {
	return o.Status != "" || o.DefaultPrice != nil || len(o.RegionalPrices) > 0 || o.DefaultLanguage != "" || o.ListingLanguage != "" || o.Listing != nil
}

func (o InAppProductPatchOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live in-app product patch cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live in-app product patch requires --confirm")
	}
	return nil
}

type InAppProductPatchPlan struct {
	Action                   string                 `json:"action"`
	PackageName              PackageName            `json:"packageName"`
	SKU                      InAppProductSKU        `json:"sku"`
	Status                   ProductStatus          `json:"status,omitempty"`
	DefaultLanguage          ListingLanguage        `json:"defaultLanguage,omitempty"`
	DefaultPrice             *ProductPrice          `json:"defaultPrice,omitempty"`
	RegionalPrices           []RegionalProductPrice `json:"regionalPrices,omitempty"`
	ListingLanguage          ListingLanguage        `json:"listingLanguage,omitempty"`
	Listing                  *InAppProductListing   `json:"listing,omitempty"`
	AutoConvertMissingPrices bool                   `json:"autoConvertMissingPrices"`
	Confirm                  bool                   `json:"confirm"`
	Steps                    []string               `json:"steps"`
}

type InAppProductPatchResult struct {
	Action  string                `json:"action"`
	DryRun  bool                  `json:"dryRun"`
	Applied bool                  `json:"applied"`
	Product *InAppProduct         `json:"product,omitempty"`
	Desired InAppProduct          `json:"desiredProduct"`
	Plan    InAppProductPatchPlan `json:"plan"`
}

type InAppProductPatcher interface {
	PatchInAppProduct(ctx context.Context, options InAppProductPatchOptions) (InAppProduct, error)
}

func PatchInAppProduct(ctx context.Context, patcher InAppProductPatcher, options InAppProductPatchOptions) (InAppProductPatchResult, error) {
	if err := options.Validate(); err != nil {
		return InAppProductPatchResult{}, err
	}
	desired := inAppProductPatchDesiredProduct(options)
	result := InAppProductPatchResult{
		Action:  "patch",
		DryRun:  options.DryRun,
		Desired: desired,
		Plan: InAppProductPatchPlan{
			Action:                   "patch",
			PackageName:              options.PackageName,
			SKU:                      options.SKU,
			Status:                   options.Status,
			DefaultLanguage:          options.DefaultLanguage,
			DefaultPrice:             options.DefaultPrice,
			RegionalPrices:           append([]RegionalProductPrice(nil), options.RegionalPrices...),
			ListingLanguage:          options.ListingLanguage,
			Listing:                  options.Listing,
			AutoConvertMissingPrices: shouldAutoConvertInAppProductPatchPrices(options),
			Confirm:                  options.Confirm,
			Steps:                    inAppProductPatchSteps(options.DryRun),
		},
	}
	if options.DryRun {
		return result, nil
	}
	if patcher == nil {
		return InAppProductPatchResult{}, fmt.Errorf("in-app product patcher is required")
	}
	getter, ok := patcher.(InAppProductGetter)
	if !ok {
		return InAppProductPatchResult{}, fmt.Errorf("in-app product getter is required for live patch preflight")
	}
	current, err := getter.GetInAppProduct(ctx, options.PackageName, options.SKU)
	if err != nil {
		return InAppProductPatchResult{}, err
	}
	if current.PurchaseType == ProductPurchaseTypeSubscription {
		return InAppProductPatchResult{}, fmt.Errorf("legacy subscription products cannot be patched with in-app-products; use subscriptions commands")
	}
	product, err := patcher.PatchInAppProduct(ctx, options)
	if err != nil {
		return InAppProductPatchResult{}, err
	}
	result.Applied = true
	result.Product = &product
	return result, nil
}

func inAppProductPatchSteps(dryRun bool) []string {
	if dryRun {
		return []string{"plan in-app product patch"}
	}
	return []string{"patch in-app product"}
}

func inAppProductPatchDesiredProduct(options InAppProductPatchOptions) InAppProduct {
	product := InAppProduct{
		PackageName: options.PackageName,
		SKU:         options.SKU,
		Status:      options.Status,
	}
	if options.DefaultLanguage != "" {
		product.DefaultLanguage = options.DefaultLanguage.String()
	}
	if options.DefaultPrice != nil {
		product.DefaultPrice = options.DefaultPrice
	}
	if len(options.RegionalPrices) > 0 {
		product.Prices = regionalProductPricesToMap(options.RegionalPrices)
	}
	if options.Listing != nil {
		product.Listings = map[string]InAppProductListing{
			options.ListingLanguage.String(): *options.Listing,
		}
	}
	return product
}

func shouldAutoConvertInAppProductPatchPrices(options InAppProductPatchOptions) bool {
	return options.DefaultPrice != nil || len(options.RegionalPrices) > 0
}

func validateRegionalProductPrices(prices []RegionalProductPrice) error {
	seen := make(map[string]struct{}, len(prices))
	for _, regionalPrice := range prices {
		regionCode := strings.ToUpper(strings.TrimSpace(regionalPrice.RegionCode))
		if !isValidRegionCode(regionCode) || regionCode != regionalPrice.RegionCode {
			return fmt.Errorf("regional price region must be a two-letter uppercase ISO 3166 code")
		}
		if _, ok := seen[regionCode]; ok {
			return fmt.Errorf("regional price for region %q is duplicated", regionCode)
		}
		seen[regionCode] = struct{}{}
		if err := regionalPrice.Price.Validate(); err != nil {
			return fmt.Errorf("regional price for region %q: %w", regionCode, err)
		}
	}
	return nil
}

func regionalProductPricesToMap(prices []RegionalProductPrice) map[string]ProductPrice {
	if len(prices) == 0 {
		return nil
	}
	mappedPrices := make(map[string]ProductPrice, len(prices))
	for _, regionalPrice := range prices {
		mappedPrices[regionalPrice.RegionCode] = regionalPrice.Price
	}
	return mappedPrices
}

func validateRequiredInAppProductListing(listing InAppProductListing) error {
	if strings.TrimSpace(listing.Title) == "" {
		return fmt.Errorf("listing title is required")
	}
	if utf8.RuneCountInString(listing.Title) > inAppProductListingTitleLimit {
		return fmt.Errorf("listing title must be %d characters or fewer", inAppProductListingTitleLimit)
	}
	if strings.TrimSpace(listing.Description) == "" {
		return fmt.Errorf("listing description is required")
	}
	if utf8.RuneCountInString(listing.Description) > inAppProductListingDescriptionLimit {
		return fmt.Errorf("listing description must be %d characters or fewer", inAppProductListingDescriptionLimit)
	}
	return nil
}
