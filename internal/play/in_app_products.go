package play

import (
	"context"
	"fmt"
	"strings"
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

type InAppProductListing struct {
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Benefits    []string `json:"benefits,omitempty"`
}

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

func GetInAppProduct(ctx context.Context, getter InAppProductGetter, options InAppProductGetOptions) (InAppProduct, error) {
	if err := options.Validate(); err != nil {
		return InAppProduct{}, err
	}
	if getter == nil {
		return InAppProduct{}, fmt.Errorf("in-app product getter is required")
	}
	return getter.GetInAppProduct(ctx, options.PackageName, options.SKU)
}

type InAppProductPatchOptions struct {
	PackageName PackageName     `json:"packageName"`
	SKU         InAppProductSKU `json:"sku"`
	Status      ProductStatus   `json:"status"`
	Confirm     bool            `json:"confirm"`
	DryRun      bool            `json:"dryRun"`
}

func (o InAppProductPatchOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewInAppProductSKU(o.SKU.String()); err != nil {
		return err
	}
	if err := o.Status.Validate(); err != nil {
		return err
	}
	if o.DryRun && o.Confirm {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.DryRun && !o.Confirm {
		return fmt.Errorf("in-app product patch requires --confirm or --dry-run")
	}
	return nil
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
	Action      string          `json:"action"`
	PackageName PackageName     `json:"packageName"`
	SKU         InAppProductSKU `json:"sku"`
	Status      ProductStatus   `json:"status"`
	Confirm     bool            `json:"confirm"`
	Steps       []string        `json:"steps"`
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
	desired := InAppProduct{
		PackageName: options.PackageName,
		SKU:         options.SKU,
		Status:      options.Status,
	}
	result := InAppProductPatchResult{
		Action:  "patch",
		DryRun:  options.DryRun,
		Desired: desired,
		Plan: InAppProductPatchPlan{
			Action:      "patch",
			PackageName: options.PackageName,
			SKU:         options.SKU,
			Status:      options.Status,
			Confirm:     options.Confirm,
			Steps:       inAppProductPatchSteps(options.DryRun),
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
