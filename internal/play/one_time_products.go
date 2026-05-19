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

type OneTimeProductDeleter interface {
	DeleteOneTimeProduct(ctx context.Context, options OneTimeProductDeleteOptions) error
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

type OneTimeProductDeletePlan struct {
	PackageName      PackageName                   `json:"packageName"`
	ProductID        OneTimeProductID              `json:"productId"`
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

type PurchaseOptionStateUpdateOptions struct {
	PackageName      PackageName                    `json:"packageName"`
	ProductID        OneTimeProductID               `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID `json:"purchaseOptionId"`
	Action           PurchaseOptionStateAction      `json:"action"`
	LatencyTolerance ProductUpdateLatencyTolerance  `json:"latencyTolerance"`
	Confirm          bool                           `json:"confirm"`
	DryRun           bool                           `json:"dryRun"`
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

func purchaseOptionStateUpdateSteps(options PurchaseOptionStateUpdateOptions) []string {
	if options.DryRun {
		return []string{fmt.Sprintf("plan %s purchase option", options.Action)}
	}
	return []string{fmt.Sprintf("%s purchase option", options.Action)}
}
