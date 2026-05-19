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

type OneTimeProductPatcher interface {
	PatchOneTimeProduct(ctx context.Context, options OneTimeProductPatchOptions) (OneTimeProduct, error)
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

type OneTimeProductPatchResult struct {
	PackageName PackageName             `json:"packageName"`
	ProductID   OneTimeProductID        `json:"productId"`
	DryRun      bool                    `json:"dryRun"`
	Applied     bool                    `json:"applied"`
	Desired     OneTimeProduct          `json:"desiredProduct"`
	Product     *OneTimeProduct         `json:"product,omitempty"`
	Plan        OneTimeProductPatchPlan `json:"plan"`
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

const oneTimeProductPatchUpdateMask = "listings"

func oneTimeProductPatchSteps(options OneTimeProductPatchOptions) []string {
	if options.DryRun {
		return []string{"plan one-time product listing patch"}
	}
	return []string{"fetch current one-time product", "merge localized listing", "patch one-time product listings"}
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
