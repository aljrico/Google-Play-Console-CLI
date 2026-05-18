package play

import (
	"context"
	"fmt"
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
