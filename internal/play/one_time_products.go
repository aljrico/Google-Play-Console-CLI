package play

import (
	"context"
	"fmt"
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

func GetOneTimeProduct(ctx context.Context, getter OneTimeProductGetter, options OneTimeProductGetOptions) (OneTimeProduct, error) {
	if err := options.Validate(); err != nil {
		return OneTimeProduct{}, err
	}
	if getter == nil {
		return OneTimeProduct{}, fmt.Errorf("one-time product getter is required")
	}
	return getter.GetOneTimeProduct(ctx, options.PackageName, options.ProductID)
}
