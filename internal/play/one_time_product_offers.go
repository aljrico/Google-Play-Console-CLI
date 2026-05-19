package play

import (
	"context"
	"fmt"
	"math"
	"strings"
)

const OneTimeProductOfferWildcardID = "-"

type OneTimeProductPurchaseOptionID string

func NewOneTimeProductPurchaseOptionID(value string) (OneTimeProductPurchaseOptionID, error) {
	if value == "" {
		return "", fmt.Errorf("one-time product purchase option ID is required")
	}
	if !isValidOneTimeProductDashID(value) {
		return "", fmt.Errorf("invalid one-time product purchase option ID %q", value)
	}
	return OneTimeProductPurchaseOptionID(value), nil
}

func NewOneTimeProductOfferListProductID(value string) (OneTimeProductID, error) {
	if value == OneTimeProductOfferWildcardID {
		return OneTimeProductID(value), nil
	}
	return NewOneTimeProductID(value)
}

func NewOneTimeProductOfferListPurchaseOptionID(value string) (OneTimeProductPurchaseOptionID, error) {
	if value == OneTimeProductOfferWildcardID {
		return OneTimeProductPurchaseOptionID(value), nil
	}
	return NewOneTimeProductPurchaseOptionID(value)
}

func (p OneTimeProductPurchaseOptionID) String() string {
	return string(p)
}

type OneTimeProductOfferID string

func NewOneTimeProductOfferID(value string) (OneTimeProductOfferID, error) {
	if value == "" {
		return "", fmt.Errorf("one-time product offer ID is required")
	}
	if !isValidOneTimeProductDashID(value) {
		return "", fmt.Errorf("invalid one-time product offer ID %q", value)
	}
	return OneTimeProductOfferID(value), nil
}

func (o OneTimeProductOfferID) String() string {
	return string(o)
}

func isValidOneTimeProductDashID(value string) bool {
	if len(value) > 63 {
		return false
	}
	first := value[0]
	if !isASCIILower(first) && !isASCIIDigit(first) {
		return false
	}
	for i := 1; i < len(value); i++ {
		character := value[i]
		if !isASCIILower(character) && !isASCIIDigit(character) && character != '-' {
			return false
		}
	}
	return true
}

type OneTimeProductOfferType string

const (
	OneTimeProductOfferTypeDiscounted OneTimeProductOfferType = "discounted"
	OneTimeProductOfferTypePreOrder   OneTimeProductOfferType = "preOrder"
)

type OneTimeProductOfferStateAction string

const (
	OneTimeProductOfferStateActionActivate   OneTimeProductOfferStateAction = "activate"
	OneTimeProductOfferStateActionDeactivate OneTimeProductOfferStateAction = "deactivate"
	OneTimeProductOfferStateActionCancel     OneTimeProductOfferStateAction = "cancel"
)

func (a OneTimeProductOfferStateAction) String() string {
	return string(a)
}

func (a OneTimeProductOfferStateAction) Validate() error {
	switch a {
	case OneTimeProductOfferStateActionActivate, OneTimeProductOfferStateActionDeactivate, OneTimeProductOfferStateActionCancel:
		return nil
	default:
		return fmt.Errorf("unsupported one-time product offer state action %q", a)
	}
}

type OneTimeProductOfferAvailability string

const (
	OneTimeProductOfferAvailabilityAvailable         OneTimeProductOfferAvailability = "available"
	OneTimeProductOfferAvailabilityNoLongerAvailable OneTimeProductOfferAvailability = "noLongerAvailable"
)

func (a OneTimeProductOfferAvailability) String() string {
	return string(a)
}

func (a OneTimeProductOfferAvailability) Validate() error {
	switch a {
	case OneTimeProductOfferAvailabilityAvailable, OneTimeProductOfferAvailabilityNoLongerAvailable:
		return nil
	default:
		return fmt.Errorf("unsupported one-time product offer availability %q", a)
	}
}

type OneTimeProductOffer struct {
	PackageName      PackageName                    `json:"packageName"`
	ProductID        OneTimeProductID               `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID `json:"purchaseOptionId"`
	OfferID          OneTimeProductOfferID          `json:"offerId"`
	State            string                         `json:"state,omitempty"`
	Type             OneTimeProductOfferType        `json:"type,omitempty"`
	OfferTags        []string                       `json:"offerTags,omitempty"`
	RegionsVersion   *RegionsVersion                `json:"regionsVersion,omitempty"`
	DiscountedOffer  *OneTimeProductDiscountedOffer `json:"discountedOffer,omitempty"`
	PreOrderOffer    *OneTimeProductPreOrderOffer   `json:"preOrderOffer,omitempty"`
	RegionalConfigs  []OneTimeProductOfferRegion    `json:"regionalConfigs,omitempty"`
}

type OneTimeProductDiscountedOffer struct {
	StartTime       string `json:"startTime,omitempty"`
	EndTime         string `json:"endTime,omitempty"`
	RedemptionLimit int64  `json:"redemptionLimit,omitempty"`
}

type OneTimeProductPreOrderOffer struct {
	StartTime           string `json:"startTime,omitempty"`
	EndTime             string `json:"endTime,omitempty"`
	ReleaseTime         string `json:"releaseTime,omitempty"`
	PriceChangeBehavior string `json:"priceChangeBehavior,omitempty"`
}

type OneTimeProductOfferRegion struct {
	RegionCode       string  `json:"regionCode"`
	Availability     string  `json:"availability,omitempty"`
	AbsoluteDiscount *Money  `json:"absoluteDiscount,omitempty"`
	RelativeDiscount float64 `json:"relativeDiscount,omitempty"`
	NoOverride       bool    `json:"noOverride,omitempty"`
}

type OneTimeProductOfferListOptions struct {
	PackageName      PackageName                    `json:"packageName"`
	ProductID        OneTimeProductID               `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID `json:"purchaseOptionId"`
	PageSize         int64                          `json:"pageSize,omitempty"`
	PageToken        string                         `json:"pageToken,omitempty"`
}

func (o OneTimeProductOfferListOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewOneTimeProductOfferListProductID(o.ProductID.String()); err != nil {
		return err
	}
	if _, err := NewOneTimeProductOfferListPurchaseOptionID(o.PurchaseOptionID.String()); err != nil {
		return err
	}
	if o.ProductID.String() == OneTimeProductOfferWildcardID && o.PurchaseOptionID.String() != OneTimeProductOfferWildcardID {
		return fmt.Errorf("one-time product purchase option ID must be %q when product ID is %q", OneTimeProductOfferWildcardID, OneTimeProductOfferWildcardID)
	}
	if o.PageSize < 0 {
		return fmt.Errorf("page size cannot be negative")
	}
	if o.PageSize > 1000 {
		return fmt.Errorf("page size cannot exceed 1000")
	}
	return nil
}

type OneTimeProductOfferListResult struct {
	PackageName      PackageName                    `json:"packageName"`
	ProductID        OneTimeProductID               `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID `json:"purchaseOptionId"`
	Offers           []OneTimeProductOffer          `json:"offers"`
	NextPageToken    string                         `json:"nextPageToken,omitempty"`
	Options          OneTimeProductOfferListOptions `json:"options"`
}

type OneTimeProductOfferLister interface {
	ListOneTimeProductOffers(ctx context.Context, options OneTimeProductOfferListOptions) (OneTimeProductOfferListResult, error)
}

func ListOneTimeProductOffers(ctx context.Context, lister OneTimeProductOfferLister, options OneTimeProductOfferListOptions) (OneTimeProductOfferListResult, error) {
	if err := options.Validate(); err != nil {
		return OneTimeProductOfferListResult{}, err
	}
	if lister == nil {
		return OneTimeProductOfferListResult{}, fmt.Errorf("one-time product offer lister is required")
	}
	return lister.ListOneTimeProductOffers(ctx, options)
}

type OneTimeProductOfferGetOptions struct {
	PackageName      PackageName                    `json:"packageName"`
	ProductID        OneTimeProductID               `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID `json:"purchaseOptionId"`
	OfferID          OneTimeProductOfferID          `json:"offerId"`
}

type OneTimeProductOfferBatchGetRequest struct {
	ProductID        OneTimeProductID               `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID `json:"purchaseOptionId"`
	OfferID          OneTimeProductOfferID          `json:"offerId"`
}

type OneTimeProductOfferBatchMutationRequest = OneTimeProductOfferBatchGetRequest
type OneTimeProductOfferBatchDeleteRequest = OneTimeProductOfferBatchMutationRequest

func NewOneTimeProductOfferBatchGetRequest(value string) (OneTimeProductOfferBatchGetRequest, error) {
	parts := strings.Split(strings.TrimSpace(value), "/")
	if len(parts) != 3 {
		return OneTimeProductOfferBatchGetRequest{}, fmt.Errorf("one-time product offer must use productId/purchaseOptionId/offerId")
	}
	productID, err := NewOneTimeProductID(parts[0])
	if err != nil {
		return OneTimeProductOfferBatchGetRequest{}, err
	}
	purchaseOptionID, err := NewOneTimeProductPurchaseOptionID(parts[1])
	if err != nil {
		return OneTimeProductOfferBatchGetRequest{}, err
	}
	offerID, err := NewOneTimeProductOfferID(parts[2])
	if err != nil {
		return OneTimeProductOfferBatchGetRequest{}, err
	}
	return OneTimeProductOfferBatchGetRequest{ProductID: productID, PurchaseOptionID: purchaseOptionID, OfferID: offerID}, nil
}

func NewOneTimeProductOfferBatchDeleteRequest(value string) (OneTimeProductOfferBatchDeleteRequest, error) {
	return NewOneTimeProductOfferBatchGetRequest(value)
}

func NewOneTimeProductOfferBatchMutationRequest(value string) (OneTimeProductOfferBatchMutationRequest, error) {
	return NewOneTimeProductOfferBatchGetRequest(value)
}

type OneTimeProductOfferBatchGetOptions struct {
	PackageName      PackageName                          `json:"packageName"`
	ProductID        OneTimeProductID                     `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID       `json:"purchaseOptionId"`
	Requests         []OneTimeProductOfferBatchGetRequest `json:"requests"`
}

type OneTimeProductOfferBatchDeleteOptions struct {
	PackageName      PackageName                               `json:"packageName"`
	ProductID        OneTimeProductID                          `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID            `json:"purchaseOptionId"`
	Requests         []OneTimeProductOfferBatchMutationRequest `json:"requests"`
	LatencyTolerance ProductUpdateLatencyTolerance             `json:"latencyTolerance"`
	Confirm          bool                                      `json:"confirm"`
	DryRun           bool                                      `json:"dryRun"`
}

type OneTimeProductOfferBatchStateUpdateOptions struct {
	PackageName      PackageName                               `json:"packageName"`
	ProductID        OneTimeProductID                          `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID            `json:"purchaseOptionId"`
	Requests         []OneTimeProductOfferBatchMutationRequest `json:"requests"`
	Action           OneTimeProductOfferStateAction            `json:"action"`
	LatencyTolerance ProductUpdateLatencyTolerance             `json:"latencyTolerance"`
	Confirm          bool                                      `json:"confirm"`
	DryRun           bool                                      `json:"dryRun"`
}

type OneTimeProductOfferAvailabilityPatchRequest struct {
	ProductID        OneTimeProductID                `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID  `json:"purchaseOptionId"`
	OfferID          OneTimeProductOfferID           `json:"offerId"`
	RegionCode       string                          `json:"regionCode"`
	Availability     OneTimeProductOfferAvailability `json:"availability"`
}

type OneTimeProductOfferRelativeDiscountPatchRequest struct {
	ProductID        OneTimeProductID               `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID `json:"purchaseOptionId"`
	OfferID          OneTimeProductOfferID          `json:"offerId"`
	RegionCode       string                         `json:"regionCode"`
	RelativeDiscount float64                        `json:"relativeDiscount"`
}

type OneTimeProductOfferAbsoluteDiscountPatchRequest struct {
	ProductID        OneTimeProductID               `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID `json:"purchaseOptionId"`
	OfferID          OneTimeProductOfferID          `json:"offerId"`
	RegionCode       string                         `json:"regionCode"`
	AbsoluteDiscount Money                          `json:"absoluteDiscount"`
}

type OneTimeProductOfferBatchPatchAvailabilityOptions struct {
	PackageName      PackageName                                   `json:"packageName"`
	ProductID        OneTimeProductID                              `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID                `json:"purchaseOptionId"`
	Requests         []OneTimeProductOfferAvailabilityPatchRequest `json:"requests"`
	RegionsVersion   string                                        `json:"regionsVersion"`
	LatencyTolerance ProductUpdateLatencyTolerance                 `json:"latencyTolerance"`
	Confirm          bool                                          `json:"confirm"`
	DryRun           bool                                          `json:"dryRun"`
}

type OneTimeProductOfferBatchPatchRelativeDiscountsOptions struct {
	PackageName      PackageName                                       `json:"packageName"`
	ProductID        OneTimeProductID                                  `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID                    `json:"purchaseOptionId"`
	Requests         []OneTimeProductOfferRelativeDiscountPatchRequest `json:"requests"`
	RegionsVersion   string                                            `json:"regionsVersion"`
	LatencyTolerance ProductUpdateLatencyTolerance                     `json:"latencyTolerance"`
	Confirm          bool                                              `json:"confirm"`
	DryRun           bool                                              `json:"dryRun"`
}

type OneTimeProductOfferBatchPatchAbsoluteDiscountsOptions struct {
	PackageName      PackageName                                       `json:"packageName"`
	ProductID        OneTimeProductID                                  `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID                    `json:"purchaseOptionId"`
	Requests         []OneTimeProductOfferAbsoluteDiscountPatchRequest `json:"requests"`
	RegionsVersion   string                                            `json:"regionsVersion"`
	LatencyTolerance ProductUpdateLatencyTolerance                     `json:"latencyTolerance"`
	Confirm          bool                                              `json:"confirm"`
	DryRun           bool                                              `json:"dryRun"`
}

type OneTimeProductOfferCreateOptions struct {
	PackageName      PackageName                    `json:"packageName"`
	ProductID        OneTimeProductID               `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID `json:"purchaseOptionId"`
	OfferID          OneTimeProductOfferID          `json:"offerId"`
	Offer            OneTimeProductOffer            `json:"offer"`
	RegionsVersion   string                         `json:"regionsVersion"`
	LatencyTolerance ProductUpdateLatencyTolerance  `json:"latencyTolerance"`
	Confirm          bool                           `json:"confirm"`
	DryRun           bool                           `json:"dryRun"`
}

func (o OneTimeProductOfferGetOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewOneTimeProductID(o.ProductID.String()); err != nil {
		return err
	}
	if _, err := NewOneTimeProductPurchaseOptionID(o.PurchaseOptionID.String()); err != nil {
		return err
	}
	if _, err := NewOneTimeProductOfferID(o.OfferID.String()); err != nil {
		return err
	}
	return nil
}

func (o OneTimeProductOfferBatchGetOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewOneTimeProductOfferListProductID(o.ProductID.String()); err != nil {
		return err
	}
	if _, err := NewOneTimeProductOfferListPurchaseOptionID(o.PurchaseOptionID.String()); err != nil {
		return err
	}
	if o.ProductID.String() == OneTimeProductOfferWildcardID && o.PurchaseOptionID.String() != OneTimeProductOfferWildcardID {
		return fmt.Errorf("one-time product purchase option ID must be %q when product ID is %q", OneTimeProductOfferWildcardID, OneTimeProductOfferWildcardID)
	}
	if len(o.Requests) == 0 {
		return fmt.Errorf("at least one one-time product offer is required")
	}
	if len(o.Requests) > 100 {
		return fmt.Errorf("one-time product offer batch-get cannot exceed 100 offers")
	}
	seen := map[string]struct{}{}
	for _, request := range o.Requests {
		if _, err := NewOneTimeProductID(request.ProductID.String()); err != nil {
			return err
		}
		if _, err := NewOneTimeProductPurchaseOptionID(request.PurchaseOptionID.String()); err != nil {
			return err
		}
		if _, err := NewOneTimeProductOfferID(request.OfferID.String()); err != nil {
			return err
		}
		if o.ProductID.String() != OneTimeProductOfferWildcardID && request.ProductID != o.ProductID {
			return fmt.Errorf("one-time product offer %s/%s/%s does not match parent product ID %s", request.ProductID, request.PurchaseOptionID, request.OfferID, o.ProductID)
		}
		if o.PurchaseOptionID.String() != OneTimeProductOfferWildcardID && request.PurchaseOptionID != o.PurchaseOptionID {
			return fmt.Errorf("one-time product offer %s/%s/%s does not match parent purchase option ID %s", request.ProductID, request.PurchaseOptionID, request.OfferID, o.PurchaseOptionID)
		}
		key := oneTimeProductOfferKey(request.ProductID, request.PurchaseOptionID, request.OfferID)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("one-time product offer %s is duplicated", key)
		}
		seen[key] = struct{}{}
	}
	return nil
}

func (o OneTimeProductOfferCreateOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewOneTimeProductID(o.ProductID.String()); err != nil {
		return err
	}
	if _, err := NewOneTimeProductPurchaseOptionID(o.PurchaseOptionID.String()); err != nil {
		return err
	}
	if _, err := NewOneTimeProductOfferID(o.OfferID.String()); err != nil {
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
		return fmt.Errorf("one-time product offer creation requires --confirm or --dry-run")
	}
	return validateOneTimeProductOfferForCreate(oneTimeProductOfferCreateDesiredOffer(o))
}

func (o OneTimeProductOfferBatchDeleteOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if err := validateOneTimeProductOfferBatchMutationParents(o.ProductID, o.PurchaseOptionID, o.Requests, "batch-delete"); err != nil {
		return err
	}
	if _, err := NewProductUpdateLatencyTolerance(o.LatencyTolerance.String()); err != nil {
		return err
	}
	if o.Confirm && o.DryRun {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.Confirm && !o.DryRun {
		return fmt.Errorf("one-time product offer batch deletion requires --confirm or --dry-run")
	}
	return nil
}

func (o OneTimeProductOfferBatchStateUpdateOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if err := validateOneTimeProductOfferBatchMutationParents(o.ProductID, o.PurchaseOptionID, o.Requests, "batch state update"); err != nil {
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
		return fmt.Errorf("one-time product offer batch state update requires --confirm or --dry-run")
	}
	return nil
}

func (o OneTimeProductOfferBatchPatchAvailabilityOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if err := validateOneTimeProductOfferAvailabilityPatchParents(o.ProductID, o.PurchaseOptionID, o.Requests); err != nil {
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
		return fmt.Errorf("one-time product offer availability batch patch requires --confirm or --dry-run")
	}
	return nil
}

func (o OneTimeProductOfferBatchPatchRelativeDiscountsOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if err := validateOneTimeProductOfferRelativeDiscountPatchParents(o.ProductID, o.PurchaseOptionID, o.Requests); err != nil {
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
		return fmt.Errorf("one-time product offer relative discount batch patch requires --confirm or --dry-run")
	}
	return nil
}

func (o OneTimeProductOfferBatchPatchAbsoluteDiscountsOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if err := validateOneTimeProductOfferAbsoluteDiscountPatchParents(o.ProductID, o.PurchaseOptionID, o.Requests); err != nil {
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
		return fmt.Errorf("one-time product offer absolute discount batch patch requires --confirm or --dry-run")
	}
	return nil
}

func (o OneTimeProductOfferCreateOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live one-time product offer creation cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live one-time product offer creation requires --confirm")
	}
	return nil
}

func (o OneTimeProductOfferBatchDeleteOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live one-time product offer batch deletion cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live one-time product offer batch deletion requires --confirm")
	}
	return nil
}

func (o OneTimeProductOfferBatchStateUpdateOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live one-time product offer batch state update cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live one-time product offer batch state update requires --confirm")
	}
	return nil
}

func (o OneTimeProductOfferBatchPatchAvailabilityOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live one-time product offer availability batch patch cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live one-time product offer availability batch patch requires --confirm")
	}
	return nil
}

func (o OneTimeProductOfferBatchPatchRelativeDiscountsOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live one-time product offer relative discount batch patch cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live one-time product offer relative discount batch patch requires --confirm")
	}
	return nil
}

func (o OneTimeProductOfferBatchPatchAbsoluteDiscountsOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live one-time product offer absolute discount batch patch cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live one-time product offer absolute discount batch patch requires --confirm")
	}
	return nil
}

func validateOneTimeProductOfferAvailabilityPatchParents(productID OneTimeProductID, purchaseOptionID OneTimeProductPurchaseOptionID, requests []OneTimeProductOfferAvailabilityPatchRequest) error {
	if len(requests) == 0 {
		return fmt.Errorf("at least one one-time product offer availability patch is required")
	}
	mutationRequests := make([]OneTimeProductOfferBatchMutationRequest, 0, len(requests))
	seenRegions := map[string]struct{}{}
	for _, request := range requests {
		if !isValidRegionCode(request.RegionCode) {
			return fmt.Errorf("one-time product offer availability region must be a two-letter ISO 3166 code")
		}
		if err := request.Availability.Validate(); err != nil {
			return err
		}
		key := oneTimeProductOfferKey(request.ProductID, request.PurchaseOptionID, request.OfferID) + "/" + request.RegionCode
		if _, ok := seenRegions[key]; ok {
			return fmt.Errorf("one-time product offer availability %s is duplicated", key)
		}
		seenRegions[key] = struct{}{}
		mutationRequests = append(mutationRequests, OneTimeProductOfferBatchMutationRequest{
			ProductID:        request.ProductID,
			PurchaseOptionID: request.PurchaseOptionID,
			OfferID:          request.OfferID,
		})
	}
	return validateOneTimeProductOfferBatchMutationParents(productID, purchaseOptionID, deduplicateOneTimeProductOfferMutationRequests(mutationRequests), "availability batch patch")
}

func validateOneTimeProductOfferRelativeDiscountPatchParents(productID OneTimeProductID, purchaseOptionID OneTimeProductPurchaseOptionID, requests []OneTimeProductOfferRelativeDiscountPatchRequest) error {
	if len(requests) == 0 {
		return fmt.Errorf("at least one one-time product offer relative discount patch is required")
	}
	mutationRequests := make([]OneTimeProductOfferBatchMutationRequest, 0, len(requests))
	seenRegions := map[string]struct{}{}
	for _, request := range requests {
		if !isValidRegionCode(request.RegionCode) {
			return fmt.Errorf("one-time product offer relative discount region must be a two-letter ISO 3166 code")
		}
		if math.IsNaN(request.RelativeDiscount) || math.IsInf(request.RelativeDiscount, 0) || request.RelativeDiscount <= 0 || request.RelativeDiscount >= 1 {
			return fmt.Errorf("one-time product offer relative discount must be greater than 0 and less than 1")
		}
		key := oneTimeProductOfferKey(request.ProductID, request.PurchaseOptionID, request.OfferID) + "/" + request.RegionCode
		if _, ok := seenRegions[key]; ok {
			return fmt.Errorf("one-time product offer relative discount %s is duplicated", key)
		}
		seenRegions[key] = struct{}{}
		mutationRequests = append(mutationRequests, OneTimeProductOfferBatchMutationRequest{
			ProductID:        request.ProductID,
			PurchaseOptionID: request.PurchaseOptionID,
			OfferID:          request.OfferID,
		})
	}
	return validateOneTimeProductOfferBatchMutationParents(productID, purchaseOptionID, deduplicateOneTimeProductOfferMutationRequests(mutationRequests), "relative discount batch patch")
}

func validateOneTimeProductOfferAbsoluteDiscountPatchParents(productID OneTimeProductID, purchaseOptionID OneTimeProductPurchaseOptionID, requests []OneTimeProductOfferAbsoluteDiscountPatchRequest) error {
	if len(requests) == 0 {
		return fmt.Errorf("at least one one-time product offer absolute discount patch is required")
	}
	mutationRequests := make([]OneTimeProductOfferBatchMutationRequest, 0, len(requests))
	seenRegions := map[string]struct{}{}
	for _, request := range requests {
		if !isValidRegionCode(request.RegionCode) {
			return fmt.Errorf("one-time product offer absolute discount region must be a two-letter ISO 3166 code")
		}
		if err := validateMoney(request.AbsoluteDiscount); err != nil {
			return fmt.Errorf("one-time product offer absolute discount for %s/%s/%s/%s: %w", request.ProductID, request.PurchaseOptionID, request.OfferID, request.RegionCode, err)
		}
		key := oneTimeProductOfferKey(request.ProductID, request.PurchaseOptionID, request.OfferID) + "/" + request.RegionCode
		if _, ok := seenRegions[key]; ok {
			return fmt.Errorf("one-time product offer absolute discount %s is duplicated", key)
		}
		seenRegions[key] = struct{}{}
		mutationRequests = append(mutationRequests, OneTimeProductOfferBatchMutationRequest{
			ProductID:        request.ProductID,
			PurchaseOptionID: request.PurchaseOptionID,
			OfferID:          request.OfferID,
		})
	}
	return validateOneTimeProductOfferBatchMutationParents(productID, purchaseOptionID, deduplicateOneTimeProductOfferMutationRequests(mutationRequests), "absolute discount batch patch")
}

func deduplicateOneTimeProductOfferMutationRequests(requests []OneTimeProductOfferBatchMutationRequest) []OneTimeProductOfferBatchMutationRequest {
	seen := map[string]struct{}{}
	deduplicated := make([]OneTimeProductOfferBatchMutationRequest, 0, len(requests))
	for _, request := range requests {
		key := oneTimeProductOfferKey(request.ProductID, request.PurchaseOptionID, request.OfferID)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		deduplicated = append(deduplicated, request)
	}
	return deduplicated
}

func validateOneTimeProductOfferBatchMutationParents(productID OneTimeProductID, purchaseOptionID OneTimeProductPurchaseOptionID, requests []OneTimeProductOfferBatchMutationRequest, operation string) error {
	if _, err := NewOneTimeProductOfferListProductID(productID.String()); err != nil {
		return err
	}
	if _, err := NewOneTimeProductOfferListPurchaseOptionID(purchaseOptionID.String()); err != nil {
		return err
	}
	if productID.String() == OneTimeProductOfferWildcardID && purchaseOptionID.String() != OneTimeProductOfferWildcardID {
		return fmt.Errorf("one-time product purchase option ID must be %q when product ID is %q", OneTimeProductOfferWildcardID, OneTimeProductOfferWildcardID)
	}
	if len(requests) == 0 {
		return fmt.Errorf("at least one one-time product offer is required")
	}
	if len(requests) > 100 {
		return fmt.Errorf("one-time product offer %s cannot exceed 100 offers", operation)
	}
	seen := map[string]struct{}{}
	seenProducts := map[OneTimeProductID]struct{}{}
	seenPurchaseOptions := map[OneTimeProductPurchaseOptionID]struct{}{}
	for _, request := range requests {
		if _, err := NewOneTimeProductID(request.ProductID.String()); err != nil {
			return err
		}
		if _, err := NewOneTimeProductPurchaseOptionID(request.PurchaseOptionID.String()); err != nil {
			return err
		}
		if _, err := NewOneTimeProductOfferID(request.OfferID.String()); err != nil {
			return err
		}
		if productID.String() != OneTimeProductOfferWildcardID && request.ProductID != productID {
			return fmt.Errorf("one-time product offer %s/%s/%s does not match parent product ID %s", request.ProductID, request.PurchaseOptionID, request.OfferID, productID)
		}
		if purchaseOptionID.String() != OneTimeProductOfferWildcardID && request.PurchaseOptionID != purchaseOptionID {
			return fmt.Errorf("one-time product offer %s/%s/%s does not match parent purchase option ID %s", request.ProductID, request.PurchaseOptionID, request.OfferID, purchaseOptionID)
		}
		key := oneTimeProductOfferKey(request.ProductID, request.PurchaseOptionID, request.OfferID)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("one-time product offer %s is duplicated", key)
		}
		seen[key] = struct{}{}
		seenProducts[request.ProductID] = struct{}{}
		seenPurchaseOptions[request.PurchaseOptionID] = struct{}{}
	}
	if len(seenProducts) == 1 && productID.String() == OneTimeProductOfferWildcardID {
		return fmt.Errorf("single-product offer %s requires parent product ID, not %q", operation, OneTimeProductOfferWildcardID)
	}
	if len(seenProducts) > 1 && productID.String() != OneTimeProductOfferWildcardID {
		return fmt.Errorf("multi-product offer %s requires parent product ID %q", operation, OneTimeProductOfferWildcardID)
	}
	if len(seenProducts) == 1 && len(seenPurchaseOptions) == 1 && purchaseOptionID.String() == OneTimeProductOfferWildcardID {
		return fmt.Errorf("single-purchase-option offer %s requires parent purchase option ID, not %q", operation, OneTimeProductOfferWildcardID)
	}
	if len(seenProducts) == 1 && len(seenPurchaseOptions) > 1 && purchaseOptionID.String() != OneTimeProductOfferWildcardID {
		return fmt.Errorf("multi-purchase-option offer %s requires parent purchase option ID %q", operation, OneTimeProductOfferWildcardID)
	}
	return nil
}

func validateOneTimeProductOfferForCreate(offer OneTimeProductOffer) error {
	if _, err := NewOneTimeProductID(offer.ProductID.String()); err != nil {
		return err
	}
	if _, err := NewOneTimeProductPurchaseOptionID(offer.PurchaseOptionID.String()); err != nil {
		return err
	}
	if _, err := NewOneTimeProductOfferID(offer.OfferID.String()); err != nil {
		return err
	}
	if len(offer.OfferTags) > 20 {
		return fmt.Errorf("one-time product offer create supports at most 20 offer tags")
	}
	seenTags := map[string]struct{}{}
	for _, tag := range offer.OfferTags {
		if err := validateSubscriptionOfferTag(tag); err != nil {
			return err
		}
		if _, ok := seenTags[tag]; ok {
			return fmt.Errorf("one-time product offer create offer tag %q is duplicated", tag)
		}
		seenTags[tag] = struct{}{}
	}
	switch offer.Type {
	case OneTimeProductOfferTypeDiscounted:
		if offer.DiscountedOffer == nil {
			return fmt.Errorf("discounted one-time product offer requires discountedOffer")
		}
		if offer.PreOrderOffer != nil {
			return fmt.Errorf("discounted one-time product offer cannot set preOrderOffer")
		}
		if err := validateOneTimeProductDiscountedOfferForCreate(*offer.DiscountedOffer); err != nil {
			return err
		}
	case OneTimeProductOfferTypePreOrder:
		if offer.PreOrderOffer == nil {
			return fmt.Errorf("pre-order one-time product offer requires preOrderOffer")
		}
		if offer.DiscountedOffer != nil {
			return fmt.Errorf("pre-order one-time product offer cannot set discountedOffer")
		}
		if err := validateOneTimeProductPreOrderOfferForCreate(*offer.PreOrderOffer); err != nil {
			return err
		}
	default:
		return fmt.Errorf("one-time product offer type must be discounted or preOrder")
	}
	if len(offer.RegionalConfigs) == 0 {
		return fmt.Errorf("one-time product offer create requires at least one regional config")
	}
	seenRegions := map[string]struct{}{}
	for _, config := range offer.RegionalConfigs {
		if !isValidRegionCode(config.RegionCode) {
			return fmt.Errorf("one-time product offer create region must be a two-letter ISO 3166 code")
		}
		if _, ok := seenRegions[config.RegionCode]; ok {
			return fmt.Errorf("one-time product offer create region %s is duplicated", config.RegionCode)
		}
		seenRegions[config.RegionCode] = struct{}{}
		if strings.TrimSpace(config.Availability) == "" {
			return fmt.Errorf("one-time product offer create region %s requires availability", config.RegionCode)
		}
		if err := OneTimeProductOfferAvailability(apiOneTimeProductOfferAvailabilityToCLI(config.Availability)).Validate(); err != nil {
			return err
		}
		priceModes := 0
		if config.AbsoluteDiscount != nil {
			priceModes++
			if err := validateMoney(*config.AbsoluteDiscount); err != nil {
				return fmt.Errorf("one-time product offer create region %s absolute discount: %w", config.RegionCode, err)
			}
		}
		if config.RelativeDiscount != 0 {
			priceModes++
			if math.IsNaN(config.RelativeDiscount) || math.IsInf(config.RelativeDiscount, 0) || config.RelativeDiscount <= 0 || config.RelativeDiscount >= 1 {
				return fmt.Errorf("one-time product offer create region %s relative discount must be greater than 0 and less than 1", config.RegionCode)
			}
		}
		if config.NoOverride {
			priceModes++
		}
		if priceModes != 1 {
			return fmt.Errorf("one-time product offer create region %s requires exactly one of absoluteDiscount, relativeDiscount, or noOverride", config.RegionCode)
		}
	}
	return nil
}

func validateOneTimeProductDiscountedOfferForCreate(offer OneTimeProductDiscountedOffer) error {
	if offer.RedemptionLimit < 0 || offer.RedemptionLimit > 50 {
		return fmt.Errorf("one-time product discounted offer redemption limit must be between 0 and 50")
	}
	return nil
}

func validateOneTimeProductPreOrderOfferForCreate(offer OneTimeProductPreOrderOffer) error {
	if strings.TrimSpace(offer.StartTime) == "" {
		return fmt.Errorf("one-time product pre-order offer requires start time")
	}
	if strings.TrimSpace(offer.EndTime) == "" {
		return fmt.Errorf("one-time product pre-order offer requires end time")
	}
	if strings.TrimSpace(offer.ReleaseTime) == "" {
		return fmt.Errorf("one-time product pre-order offer requires release time")
	}
	switch offer.PriceChangeBehavior {
	case "PRE_ORDER_PRICE_CHANGE_BEHAVIOR_TWO_POINT_LOWEST", "PRE_ORDER_PRICE_CHANGE_BEHAVIOR_NEW_ORDERS_ONLY":
		return nil
	default:
		return fmt.Errorf("one-time product pre-order offer price change behavior must be PRE_ORDER_PRICE_CHANGE_BEHAVIOR_TWO_POINT_LOWEST or PRE_ORDER_PRICE_CHANGE_BEHAVIOR_NEW_ORDERS_ONLY")
	}
}

func apiOneTimeProductOfferAvailabilityToCLI(value string) string {
	switch value {
	case "AVAILABLE":
		return OneTimeProductOfferAvailabilityAvailable.String()
	case "NO_LONGER_AVAILABLE":
		return OneTimeProductOfferAvailabilityNoLongerAvailable.String()
	default:
		return value
	}
}

type OneTimeProductOfferGetter interface {
	GetOneTimeProductOffer(ctx context.Context, options OneTimeProductOfferGetOptions) (OneTimeProductOffer, error)
}

type OneTimeProductOfferCreator interface {
	CreateOneTimeProductOffer(ctx context.Context, options OneTimeProductOfferCreateOptions) (OneTimeProductOffer, error)
}

type OneTimeProductOfferBatchGetter interface {
	BatchGetOneTimeProductOffers(ctx context.Context, options OneTimeProductOfferBatchGetOptions) (OneTimeProductOfferBatchGetResult, error)
}

type OneTimeProductOfferBatchDeleter interface {
	BatchDeleteOneTimeProductOffers(ctx context.Context, options OneTimeProductOfferBatchDeleteOptions) error
}

type OneTimeProductOfferBatchStateUpdater interface {
	BatchUpdateOneTimeProductOfferStates(ctx context.Context, options OneTimeProductOfferBatchStateUpdateOptions) (OneTimeProductOfferBatchStateUpdateResult, error)
}

type OneTimeProductOfferBatchAvailabilityPatcher interface {
	BatchPatchOneTimeProductOfferAvailability(ctx context.Context, options OneTimeProductOfferBatchPatchAvailabilityOptions) (OneTimeProductOfferBatchPatchAvailabilityResult, error)
}

type OneTimeProductOfferBatchRelativeDiscountPatcher interface {
	BatchPatchOneTimeProductOfferRelativeDiscounts(ctx context.Context, options OneTimeProductOfferBatchPatchRelativeDiscountsOptions) (OneTimeProductOfferBatchPatchRelativeDiscountsResult, error)
}

type OneTimeProductOfferBatchAbsoluteDiscountPatcher interface {
	BatchPatchOneTimeProductOfferAbsoluteDiscounts(ctx context.Context, options OneTimeProductOfferBatchPatchAbsoluteDiscountsOptions) (OneTimeProductOfferBatchPatchAbsoluteDiscountsResult, error)
}

type OneTimeProductOfferBatchGetResult struct {
	PackageName      PackageName                        `json:"packageName"`
	ProductID        OneTimeProductID                   `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID     `json:"purchaseOptionId"`
	Offers           []OneTimeProductOffer              `json:"offers"`
	Options          OneTimeProductOfferBatchGetOptions `json:"options"`
}

type OneTimeProductOfferCreatePlan struct {
	PackageName      PackageName                    `json:"packageName"`
	ProductID        OneTimeProductID               `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID `json:"purchaseOptionId"`
	OfferID          OneTimeProductOfferID          `json:"offerId"`
	RegionsVersion   string                         `json:"regionsVersion"`
	LatencyTolerance ProductUpdateLatencyTolerance  `json:"latencyTolerance"`
	Confirm          bool                           `json:"confirm"`
	Steps            []string                       `json:"steps"`
}

type OneTimeProductOfferCreateResult struct {
	PackageName      PackageName                    `json:"packageName"`
	ProductID        OneTimeProductID               `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID `json:"purchaseOptionId"`
	OfferID          OneTimeProductOfferID          `json:"offerId"`
	DryRun           bool                           `json:"dryRun"`
	Created          bool                           `json:"created"`
	Desired          OneTimeProductOffer            `json:"desiredOffer"`
	Offer            *OneTimeProductOffer           `json:"offer,omitempty"`
	Plan             OneTimeProductOfferCreatePlan  `json:"plan"`
}

type OneTimeProductOfferBatchDeletePlan struct {
	PackageName      PackageName                               `json:"packageName"`
	ProductID        OneTimeProductID                          `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID            `json:"purchaseOptionId"`
	Requests         []OneTimeProductOfferBatchMutationRequest `json:"requests"`
	LatencyTolerance ProductUpdateLatencyTolerance             `json:"latencyTolerance"`
	Confirm          bool                                      `json:"confirm"`
	Steps            []string                                  `json:"steps"`
}

type OneTimeProductOfferBatchDeleteResult struct {
	PackageName      PackageName                               `json:"packageName"`
	ProductID        OneTimeProductID                          `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID            `json:"purchaseOptionId"`
	Requests         []OneTimeProductOfferBatchMutationRequest `json:"requests"`
	DryRun           bool                                      `json:"dryRun"`
	Deleted          bool                                      `json:"deleted"`
	Plan             OneTimeProductOfferBatchDeletePlan        `json:"plan"`
}

type OneTimeProductOfferBatchStateUpdatePlan struct {
	PackageName      PackageName                               `json:"packageName"`
	ProductID        OneTimeProductID                          `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID            `json:"purchaseOptionId"`
	Requests         []OneTimeProductOfferBatchMutationRequest `json:"requests"`
	Action           OneTimeProductOfferStateAction            `json:"action"`
	LatencyTolerance ProductUpdateLatencyTolerance             `json:"latencyTolerance"`
	Confirm          bool                                      `json:"confirm"`
	Steps            []string                                  `json:"steps"`
}

type OneTimeProductOfferBatchStateUpdateResult struct {
	PackageName      PackageName                               `json:"packageName"`
	ProductID        OneTimeProductID                          `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID            `json:"purchaseOptionId"`
	Requests         []OneTimeProductOfferBatchMutationRequest `json:"requests"`
	Action           OneTimeProductOfferStateAction            `json:"action"`
	DryRun           bool                                      `json:"dryRun"`
	Applied          bool                                      `json:"applied"`
	Offers           []OneTimeProductOffer                     `json:"offers,omitempty"`
	Plan             OneTimeProductOfferBatchStateUpdatePlan   `json:"plan"`
}

type OneTimeProductOfferBatchPatchAvailabilityPlan struct {
	PackageName      PackageName                                   `json:"packageName"`
	ProductID        OneTimeProductID                              `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID                `json:"purchaseOptionId"`
	Requests         []OneTimeProductOfferAvailabilityPatchRequest `json:"requests"`
	UpdateMask       string                                        `json:"updateMask"`
	RegionsVersion   string                                        `json:"regionsVersion"`
	LatencyTolerance ProductUpdateLatencyTolerance                 `json:"latencyTolerance"`
	Confirm          bool                                          `json:"confirm"`
	Steps            []string                                      `json:"steps"`
}

type OneTimeProductOfferBatchPatchAvailabilityDesiredOffer struct {
	PackageName      PackageName                                              `json:"packageName"`
	ProductID        OneTimeProductID                                         `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID                           `json:"purchaseOptionId"`
	OfferID          OneTimeProductOfferID                                    `json:"offerId"`
	RegionalConfigs  []OneTimeProductOfferBatchPatchAvailabilityDesiredRegion `json:"regionalConfigs"`
}

type OneTimeProductOfferBatchPatchAvailabilityDesiredRegion struct {
	RegionCode   string                          `json:"regionCode"`
	Availability OneTimeProductOfferAvailability `json:"availability"`
}

type OneTimeProductOfferBatchPatchAvailabilityResult struct {
	PackageName      PackageName                                             `json:"packageName"`
	ProductID        OneTimeProductID                                        `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID                          `json:"purchaseOptionId"`
	Requests         []OneTimeProductOfferAvailabilityPatchRequest           `json:"requests"`
	DryRun           bool                                                    `json:"dryRun"`
	Applied          bool                                                    `json:"applied"`
	Offers           []OneTimeProductOffer                                   `json:"offers,omitempty"`
	Desired          []OneTimeProductOfferBatchPatchAvailabilityDesiredOffer `json:"desiredOffers"`
	Plan             OneTimeProductOfferBatchPatchAvailabilityPlan           `json:"plan"`
}

type OneTimeProductOfferBatchPatchRelativeDiscountsPlan struct {
	PackageName      PackageName                                       `json:"packageName"`
	ProductID        OneTimeProductID                                  `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID                    `json:"purchaseOptionId"`
	Requests         []OneTimeProductOfferRelativeDiscountPatchRequest `json:"requests"`
	UpdateMask       string                                            `json:"updateMask"`
	RegionsVersion   string                                            `json:"regionsVersion"`
	LatencyTolerance ProductUpdateLatencyTolerance                     `json:"latencyTolerance"`
	Confirm          bool                                              `json:"confirm"`
	Steps            []string                                          `json:"steps"`
}

type OneTimeProductOfferBatchPatchRelativeDiscountsDesiredOffer struct {
	PackageName      PackageName                                                   `json:"packageName"`
	ProductID        OneTimeProductID                                              `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID                                `json:"purchaseOptionId"`
	OfferID          OneTimeProductOfferID                                         `json:"offerId"`
	RegionalConfigs  []OneTimeProductOfferBatchPatchRelativeDiscountsDesiredRegion `json:"regionalConfigs"`
}

type OneTimeProductOfferBatchPatchRelativeDiscountsDesiredRegion struct {
	RegionCode       string  `json:"regionCode"`
	RelativeDiscount float64 `json:"relativeDiscount"`
}

type OneTimeProductOfferBatchPatchRelativeDiscountsResult struct {
	PackageName      PackageName                                                  `json:"packageName"`
	ProductID        OneTimeProductID                                             `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID                               `json:"purchaseOptionId"`
	Requests         []OneTimeProductOfferRelativeDiscountPatchRequest            `json:"requests"`
	DryRun           bool                                                         `json:"dryRun"`
	Applied          bool                                                         `json:"applied"`
	Offers           []OneTimeProductOffer                                        `json:"offers,omitempty"`
	Desired          []OneTimeProductOfferBatchPatchRelativeDiscountsDesiredOffer `json:"desiredOffers"`
	Plan             OneTimeProductOfferBatchPatchRelativeDiscountsPlan           `json:"plan"`
}

type OneTimeProductOfferBatchPatchAbsoluteDiscountsPlan struct {
	PackageName      PackageName                                       `json:"packageName"`
	ProductID        OneTimeProductID                                  `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID                    `json:"purchaseOptionId"`
	Requests         []OneTimeProductOfferAbsoluteDiscountPatchRequest `json:"requests"`
	UpdateMask       string                                            `json:"updateMask"`
	RegionsVersion   string                                            `json:"regionsVersion"`
	LatencyTolerance ProductUpdateLatencyTolerance                     `json:"latencyTolerance"`
	Confirm          bool                                              `json:"confirm"`
	Steps            []string                                          `json:"steps"`
}

type OneTimeProductOfferBatchPatchAbsoluteDiscountsDesiredOffer struct {
	PackageName      PackageName                                                   `json:"packageName"`
	ProductID        OneTimeProductID                                              `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID                                `json:"purchaseOptionId"`
	OfferID          OneTimeProductOfferID                                         `json:"offerId"`
	RegionalConfigs  []OneTimeProductOfferBatchPatchAbsoluteDiscountsDesiredRegion `json:"regionalConfigs"`
}

type OneTimeProductOfferBatchPatchAbsoluteDiscountsDesiredRegion struct {
	RegionCode       string `json:"regionCode"`
	AbsoluteDiscount Money  `json:"absoluteDiscount"`
}

type OneTimeProductOfferBatchPatchAbsoluteDiscountsResult struct {
	PackageName      PackageName                                                  `json:"packageName"`
	ProductID        OneTimeProductID                                             `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID                               `json:"purchaseOptionId"`
	Requests         []OneTimeProductOfferAbsoluteDiscountPatchRequest            `json:"requests"`
	DryRun           bool                                                         `json:"dryRun"`
	Applied          bool                                                         `json:"applied"`
	Offers           []OneTimeProductOffer                                        `json:"offers,omitempty"`
	Desired          []OneTimeProductOfferBatchPatchAbsoluteDiscountsDesiredOffer `json:"desiredOffers"`
	Plan             OneTimeProductOfferBatchPatchAbsoluteDiscountsPlan           `json:"plan"`
}

func GetOneTimeProductOffer(ctx context.Context, getter OneTimeProductOfferGetter, options OneTimeProductOfferGetOptions) (OneTimeProductOffer, error) {
	if err := options.Validate(); err != nil {
		return OneTimeProductOffer{}, err
	}
	if getter == nil {
		return OneTimeProductOffer{}, fmt.Errorf("one-time product offer getter is required")
	}
	return getter.GetOneTimeProductOffer(ctx, options)
}

func BatchGetOneTimeProductOffers(ctx context.Context, getter OneTimeProductOfferBatchGetter, options OneTimeProductOfferBatchGetOptions) (OneTimeProductOfferBatchGetResult, error) {
	if err := options.Validate(); err != nil {
		return OneTimeProductOfferBatchGetResult{}, err
	}
	if getter == nil {
		return OneTimeProductOfferBatchGetResult{}, fmt.Errorf("one-time product offer batch getter is required")
	}
	return getter.BatchGetOneTimeProductOffers(ctx, options)
}

func NewOneTimeProductOfferCreatePlan(options OneTimeProductOfferCreateOptions) (OneTimeProductOfferCreatePlan, error) {
	if err := options.Validate(); err != nil {
		return OneTimeProductOfferCreatePlan{}, err
	}
	return OneTimeProductOfferCreatePlan{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		PurchaseOptionID: options.PurchaseOptionID,
		OfferID:          options.OfferID,
		RegionsVersion:   options.RegionsVersion,
		LatencyTolerance: options.LatencyTolerance,
		Confirm:          options.Confirm,
		Steps:            oneTimeProductOfferCreateSteps(options),
	}, nil
}

func CreateOneTimeProductOffer(ctx context.Context, creator OneTimeProductOfferCreator, options OneTimeProductOfferCreateOptions) (OneTimeProductOfferCreateResult, error) {
	plan, err := NewOneTimeProductOfferCreatePlan(options)
	if err != nil {
		return OneTimeProductOfferCreateResult{}, err
	}
	result := OneTimeProductOfferCreateResult{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		PurchaseOptionID: options.PurchaseOptionID,
		OfferID:          options.OfferID,
		DryRun:           options.DryRun,
		Desired:          oneTimeProductOfferCreateDesiredOffer(options),
		Plan:             plan,
	}
	if options.DryRun {
		return result, nil
	}
	if creator == nil {
		return OneTimeProductOfferCreateResult{}, fmt.Errorf("one-time product offer creator is required")
	}
	offer, err := creator.CreateOneTimeProductOffer(ctx, options)
	if err != nil {
		return OneTimeProductOfferCreateResult{}, err
	}
	result.Created = true
	result.Offer = &offer
	return result, nil
}

func BatchDeleteOneTimeProductOffers(ctx context.Context, deleter OneTimeProductOfferBatchDeleter, options OneTimeProductOfferBatchDeleteOptions) (OneTimeProductOfferBatchDeleteResult, error) {
	if err := options.Validate(); err != nil {
		return OneTimeProductOfferBatchDeleteResult{}, err
	}
	requests := append([]OneTimeProductOfferBatchMutationRequest(nil), options.Requests...)
	result := OneTimeProductOfferBatchDeleteResult{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		PurchaseOptionID: options.PurchaseOptionID,
		Requests:         requests,
		DryRun:           options.DryRun,
		Plan: OneTimeProductOfferBatchDeletePlan{
			PackageName:      options.PackageName,
			ProductID:        options.ProductID,
			PurchaseOptionID: options.PurchaseOptionID,
			Requests:         requests,
			LatencyTolerance: options.LatencyTolerance,
			Confirm:          options.Confirm,
			Steps:            oneTimeProductOfferBatchDeleteSteps(options),
		},
	}
	if options.DryRun {
		return result, nil
	}
	if deleter == nil {
		return OneTimeProductOfferBatchDeleteResult{}, fmt.Errorf("one-time product offer batch deleter is required")
	}
	if err := deleter.BatchDeleteOneTimeProductOffers(ctx, options); err != nil {
		return OneTimeProductOfferBatchDeleteResult{}, err
	}
	result.Deleted = true
	return result, nil
}

func BatchUpdateOneTimeProductOfferStates(ctx context.Context, updater OneTimeProductOfferBatchStateUpdater, options OneTimeProductOfferBatchStateUpdateOptions) (OneTimeProductOfferBatchStateUpdateResult, error) {
	if err := options.Validate(); err != nil {
		return OneTimeProductOfferBatchStateUpdateResult{}, err
	}
	requests := append([]OneTimeProductOfferBatchMutationRequest(nil), options.Requests...)
	result := OneTimeProductOfferBatchStateUpdateResult{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		PurchaseOptionID: options.PurchaseOptionID,
		Requests:         requests,
		Action:           options.Action,
		DryRun:           options.DryRun,
		Plan: OneTimeProductOfferBatchStateUpdatePlan{
			PackageName:      options.PackageName,
			ProductID:        options.ProductID,
			PurchaseOptionID: options.PurchaseOptionID,
			Requests:         requests,
			Action:           options.Action,
			LatencyTolerance: options.LatencyTolerance,
			Confirm:          options.Confirm,
			Steps:            oneTimeProductOfferBatchStateUpdateSteps(options),
		},
	}
	if options.DryRun {
		return result, nil
	}
	if updater == nil {
		return OneTimeProductOfferBatchStateUpdateResult{}, fmt.Errorf("one-time product offer batch state updater is required")
	}
	updated, err := updater.BatchUpdateOneTimeProductOfferStates(ctx, options)
	if err != nil {
		return OneTimeProductOfferBatchStateUpdateResult{}, err
	}
	updated.PackageName = options.PackageName
	updated.ProductID = options.ProductID
	updated.PurchaseOptionID = options.PurchaseOptionID
	updated.Plan = result.Plan
	updated.Requests = requests
	updated.Action = options.Action
	updated.DryRun = false
	updated.Applied = true
	return updated, nil
}

func BatchPatchOneTimeProductOfferAvailability(ctx context.Context, patcher OneTimeProductOfferBatchAvailabilityPatcher, options OneTimeProductOfferBatchPatchAvailabilityOptions) (OneTimeProductOfferBatchPatchAvailabilityResult, error) {
	plan, err := NewOneTimeProductOfferBatchPatchAvailabilityPlan(options)
	if err != nil {
		return OneTimeProductOfferBatchPatchAvailabilityResult{}, err
	}
	requests := append([]OneTimeProductOfferAvailabilityPatchRequest(nil), options.Requests...)
	result := OneTimeProductOfferBatchPatchAvailabilityResult{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		PurchaseOptionID: options.PurchaseOptionID,
		Requests:         requests,
		DryRun:           options.DryRun,
		Desired:          desiredOneTimeProductOffersForAvailabilityPatch(options),
		Plan:             plan,
	}
	if options.DryRun {
		return result, nil
	}
	if patcher == nil {
		return OneTimeProductOfferBatchPatchAvailabilityResult{}, fmt.Errorf("one-time product offer availability batch patcher is required")
	}
	updated, err := patcher.BatchPatchOneTimeProductOfferAvailability(ctx, options)
	if err != nil {
		return OneTimeProductOfferBatchPatchAvailabilityResult{}, err
	}
	updated.PackageName = options.PackageName
	updated.ProductID = options.ProductID
	updated.PurchaseOptionID = options.PurchaseOptionID
	updated.Requests = requests
	updated.DryRun = false
	updated.Applied = true
	updated.Desired = result.Desired
	updated.Plan = plan
	return updated, nil
}

func BatchPatchOneTimeProductOfferRelativeDiscounts(ctx context.Context, patcher OneTimeProductOfferBatchRelativeDiscountPatcher, options OneTimeProductOfferBatchPatchRelativeDiscountsOptions) (OneTimeProductOfferBatchPatchRelativeDiscountsResult, error) {
	plan, err := NewOneTimeProductOfferBatchPatchRelativeDiscountsPlan(options)
	if err != nil {
		return OneTimeProductOfferBatchPatchRelativeDiscountsResult{}, err
	}
	requests := append([]OneTimeProductOfferRelativeDiscountPatchRequest(nil), options.Requests...)
	result := OneTimeProductOfferBatchPatchRelativeDiscountsResult{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		PurchaseOptionID: options.PurchaseOptionID,
		Requests:         requests,
		DryRun:           options.DryRun,
		Desired:          desiredOneTimeProductOffersForRelativeDiscountPatch(options),
		Plan:             plan,
	}
	if options.DryRun {
		return result, nil
	}
	if patcher == nil {
		return OneTimeProductOfferBatchPatchRelativeDiscountsResult{}, fmt.Errorf("one-time product offer relative discount batch patcher is required")
	}
	updated, err := patcher.BatchPatchOneTimeProductOfferRelativeDiscounts(ctx, options)
	if err != nil {
		return OneTimeProductOfferBatchPatchRelativeDiscountsResult{}, err
	}
	updated.PackageName = options.PackageName
	updated.ProductID = options.ProductID
	updated.PurchaseOptionID = options.PurchaseOptionID
	updated.Requests = requests
	updated.DryRun = false
	updated.Applied = true
	updated.Desired = result.Desired
	updated.Plan = plan
	return updated, nil
}

func BatchPatchOneTimeProductOfferAbsoluteDiscounts(ctx context.Context, patcher OneTimeProductOfferBatchAbsoluteDiscountPatcher, options OneTimeProductOfferBatchPatchAbsoluteDiscountsOptions) (OneTimeProductOfferBatchPatchAbsoluteDiscountsResult, error) {
	plan, err := NewOneTimeProductOfferBatchPatchAbsoluteDiscountsPlan(options)
	if err != nil {
		return OneTimeProductOfferBatchPatchAbsoluteDiscountsResult{}, err
	}
	requests := append([]OneTimeProductOfferAbsoluteDiscountPatchRequest(nil), options.Requests...)
	result := OneTimeProductOfferBatchPatchAbsoluteDiscountsResult{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		PurchaseOptionID: options.PurchaseOptionID,
		Requests:         requests,
		DryRun:           options.DryRun,
		Desired:          desiredOneTimeProductOffersForAbsoluteDiscountPatch(options),
		Plan:             plan,
	}
	if options.DryRun {
		return result, nil
	}
	if patcher == nil {
		return OneTimeProductOfferBatchPatchAbsoluteDiscountsResult{}, fmt.Errorf("one-time product offer absolute discount batch patcher is required")
	}
	updated, err := patcher.BatchPatchOneTimeProductOfferAbsoluteDiscounts(ctx, options)
	if err != nil {
		return OneTimeProductOfferBatchPatchAbsoluteDiscountsResult{}, err
	}
	updated.PackageName = options.PackageName
	updated.ProductID = options.ProductID
	updated.PurchaseOptionID = options.PurchaseOptionID
	updated.Requests = requests
	updated.DryRun = false
	updated.Applied = true
	updated.Desired = result.Desired
	updated.Plan = plan
	return updated, nil
}

func NewOneTimeProductOfferBatchPatchAvailabilityPlan(options OneTimeProductOfferBatchPatchAvailabilityOptions) (OneTimeProductOfferBatchPatchAvailabilityPlan, error) {
	if err := options.Validate(); err != nil {
		return OneTimeProductOfferBatchPatchAvailabilityPlan{}, err
	}
	return OneTimeProductOfferBatchPatchAvailabilityPlan{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		PurchaseOptionID: options.PurchaseOptionID,
		Requests:         append([]OneTimeProductOfferAvailabilityPatchRequest(nil), options.Requests...),
		UpdateMask:       oneTimeProductOfferAvailabilityUpdateMask,
		RegionsVersion:   options.RegionsVersion,
		LatencyTolerance: options.LatencyTolerance,
		Confirm:          options.Confirm,
		Steps:            oneTimeProductOfferBatchPatchAvailabilitySteps(options),
	}, nil
}

func NewOneTimeProductOfferBatchPatchRelativeDiscountsPlan(options OneTimeProductOfferBatchPatchRelativeDiscountsOptions) (OneTimeProductOfferBatchPatchRelativeDiscountsPlan, error) {
	if err := options.Validate(); err != nil {
		return OneTimeProductOfferBatchPatchRelativeDiscountsPlan{}, err
	}
	return OneTimeProductOfferBatchPatchRelativeDiscountsPlan{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		PurchaseOptionID: options.PurchaseOptionID,
		Requests:         append([]OneTimeProductOfferRelativeDiscountPatchRequest(nil), options.Requests...),
		UpdateMask:       oneTimeProductOfferRegionalConfigsUpdateMask,
		RegionsVersion:   options.RegionsVersion,
		LatencyTolerance: options.LatencyTolerance,
		Confirm:          options.Confirm,
		Steps:            oneTimeProductOfferBatchPatchRelativeDiscountsSteps(options),
	}, nil
}

func NewOneTimeProductOfferBatchPatchAbsoluteDiscountsPlan(options OneTimeProductOfferBatchPatchAbsoluteDiscountsOptions) (OneTimeProductOfferBatchPatchAbsoluteDiscountsPlan, error) {
	if err := options.Validate(); err != nil {
		return OneTimeProductOfferBatchPatchAbsoluteDiscountsPlan{}, err
	}
	return OneTimeProductOfferBatchPatchAbsoluteDiscountsPlan{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		PurchaseOptionID: options.PurchaseOptionID,
		Requests:         append([]OneTimeProductOfferAbsoluteDiscountPatchRequest(nil), options.Requests...),
		UpdateMask:       oneTimeProductOfferRegionalConfigsUpdateMask,
		RegionsVersion:   options.RegionsVersion,
		LatencyTolerance: options.LatencyTolerance,
		Confirm:          options.Confirm,
		Steps:            oneTimeProductOfferBatchPatchAbsoluteDiscountsSteps(options),
	}, nil
}

func desiredOneTimeProductOffersForAvailabilityPatch(options OneTimeProductOfferBatchPatchAvailabilityOptions) []OneTimeProductOfferBatchPatchAvailabilityDesiredOffer {
	byOffer := map[string]int{}
	offers := make([]OneTimeProductOfferBatchPatchAvailabilityDesiredOffer, 0)
	for _, request := range options.Requests {
		key := oneTimeProductOfferKey(request.ProductID, request.PurchaseOptionID, request.OfferID)
		index, ok := byOffer[key]
		if !ok {
			byOffer[key] = len(offers)
			offers = append(offers, OneTimeProductOfferBatchPatchAvailabilityDesiredOffer{
				PackageName:      options.PackageName,
				ProductID:        request.ProductID,
				PurchaseOptionID: request.PurchaseOptionID,
				OfferID:          request.OfferID,
				RegionalConfigs:  []OneTimeProductOfferBatchPatchAvailabilityDesiredRegion{},
			})
			index = len(offers) - 1
		}
		offers[index].RegionalConfigs = append(offers[index].RegionalConfigs, OneTimeProductOfferBatchPatchAvailabilityDesiredRegion{
			RegionCode:   request.RegionCode,
			Availability: request.Availability,
		})
	}
	return offers
}

func desiredOneTimeProductOffersForRelativeDiscountPatch(options OneTimeProductOfferBatchPatchRelativeDiscountsOptions) []OneTimeProductOfferBatchPatchRelativeDiscountsDesiredOffer {
	byOffer := map[string]int{}
	offers := make([]OneTimeProductOfferBatchPatchRelativeDiscountsDesiredOffer, 0)
	for _, request := range options.Requests {
		key := oneTimeProductOfferKey(request.ProductID, request.PurchaseOptionID, request.OfferID)
		index, ok := byOffer[key]
		if !ok {
			byOffer[key] = len(offers)
			offers = append(offers, OneTimeProductOfferBatchPatchRelativeDiscountsDesiredOffer{
				PackageName:      options.PackageName,
				ProductID:        request.ProductID,
				PurchaseOptionID: request.PurchaseOptionID,
				OfferID:          request.OfferID,
				RegionalConfigs:  []OneTimeProductOfferBatchPatchRelativeDiscountsDesiredRegion{},
			})
			index = len(offers) - 1
		}
		offers[index].RegionalConfigs = append(offers[index].RegionalConfigs, OneTimeProductOfferBatchPatchRelativeDiscountsDesiredRegion{
			RegionCode:       request.RegionCode,
			RelativeDiscount: request.RelativeDiscount,
		})
	}
	return offers
}

func desiredOneTimeProductOffersForAbsoluteDiscountPatch(options OneTimeProductOfferBatchPatchAbsoluteDiscountsOptions) []OneTimeProductOfferBatchPatchAbsoluteDiscountsDesiredOffer {
	byOffer := map[string]int{}
	offers := make([]OneTimeProductOfferBatchPatchAbsoluteDiscountsDesiredOffer, 0)
	for _, request := range options.Requests {
		key := oneTimeProductOfferKey(request.ProductID, request.PurchaseOptionID, request.OfferID)
		index, ok := byOffer[key]
		if !ok {
			byOffer[key] = len(offers)
			offers = append(offers, OneTimeProductOfferBatchPatchAbsoluteDiscountsDesiredOffer{
				PackageName:      options.PackageName,
				ProductID:        request.ProductID,
				PurchaseOptionID: request.PurchaseOptionID,
				OfferID:          request.OfferID,
				RegionalConfigs:  []OneTimeProductOfferBatchPatchAbsoluteDiscountsDesiredRegion{},
			})
			index = len(offers) - 1
		}
		offers[index].RegionalConfigs = append(offers[index].RegionalConfigs, OneTimeProductOfferBatchPatchAbsoluteDiscountsDesiredRegion{
			RegionCode:       request.RegionCode,
			AbsoluteDiscount: request.AbsoluteDiscount,
		})
	}
	return offers
}

func oneTimeProductOfferCreateDesiredOffer(options OneTimeProductOfferCreateOptions) OneTimeProductOffer {
	offer := options.Offer
	offer.PackageName = options.PackageName
	offer.ProductID = options.ProductID
	offer.PurchaseOptionID = options.PurchaseOptionID
	offer.OfferID = options.OfferID
	offer.State = ""
	offer.RegionsVersion = nil
	if offer.RegionalConfigs == nil {
		offer.RegionalConfigs = []OneTimeProductOfferRegion{}
	}
	return offer
}

func oneTimeProductOfferCreateSteps(options OneTimeProductOfferCreateOptions) []string {
	if options.DryRun {
		return []string{"plan one-time product offer create"}
	}
	return []string{"verify one-time product offer does not exist", "batch update one-time product offer with allowMissing=true"}
}

func oneTimeProductOfferBatchDeleteSteps(options OneTimeProductOfferBatchDeleteOptions) []string {
	if options.DryRun {
		return []string{"plan one-time product offer batch deletion"}
	}
	return []string{"batch delete one-time product offers"}
}

func oneTimeProductOfferBatchStateUpdateSteps(options OneTimeProductOfferBatchStateUpdateOptions) []string {
	if options.Action == OneTimeProductOfferStateActionCancel {
		if options.DryRun {
			return []string{"plan batch cancel pre-order one-time product offers and pending orders"}
		}
		return []string{"batch cancel pre-order one-time product offers and pending orders"}
	}
	if options.DryRun {
		return []string{fmt.Sprintf("plan batch %s one-time product offers", options.Action)}
	}
	return []string{fmt.Sprintf("batch %s one-time product offers", options.Action)}
}

const oneTimeProductOfferRegionalConfigsUpdateMask = "regionalPricingAndAvailabilityConfigs"

const oneTimeProductOfferCreateUpdateMask = "offerTags,discountedOffer,preOrderOffer,regionalPricingAndAvailabilityConfigs"

const oneTimeProductOfferAvailabilityUpdateMask = oneTimeProductOfferRegionalConfigsUpdateMask

func oneTimeProductOfferBatchPatchAvailabilitySteps(options OneTimeProductOfferBatchPatchAvailabilityOptions) []string {
	if options.DryRun {
		return []string{"plan one-time product offer availability batch patch"}
	}
	return []string{"fetch current one-time product offers", "merge regional availability", "batch patch one-time product offer availability"}
}

func oneTimeProductOfferBatchPatchRelativeDiscountsSteps(options OneTimeProductOfferBatchPatchRelativeDiscountsOptions) []string {
	if options.DryRun {
		return []string{"plan one-time product offer relative discount batch patch"}
	}
	return []string{"fetch current one-time product offers", "merge regional relative discounts", "batch patch one-time product offer relative discounts"}
}

func oneTimeProductOfferBatchPatchAbsoluteDiscountsSteps(options OneTimeProductOfferBatchPatchAbsoluteDiscountsOptions) []string {
	if options.DryRun {
		return []string{"plan one-time product offer absolute discount batch patch"}
	}
	return []string{"fetch current one-time product offers", "merge regional absolute discounts", "batch patch one-time product offer absolute discounts"}
}

type OneTimeProductOfferStateUpdateOptions struct {
	PackageName      PackageName                    `json:"packageName"`
	ProductID        OneTimeProductID               `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID `json:"purchaseOptionId"`
	OfferID          OneTimeProductOfferID          `json:"offerId"`
	Action           OneTimeProductOfferStateAction `json:"action"`
	LatencyTolerance ProductUpdateLatencyTolerance  `json:"latencyTolerance"`
	Confirm          bool                           `json:"confirm"`
	DryRun           bool                           `json:"dryRun"`
}

func (o OneTimeProductOfferStateUpdateOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewOneTimeProductID(o.ProductID.String()); err != nil {
		return err
	}
	if _, err := NewOneTimeProductPurchaseOptionID(o.PurchaseOptionID.String()); err != nil {
		return err
	}
	if _, err := NewOneTimeProductOfferID(o.OfferID.String()); err != nil {
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
		return fmt.Errorf("one-time product offer state update requires --confirm or --dry-run")
	}
	return nil
}

func (o OneTimeProductOfferStateUpdateOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live one-time product offer state update cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live one-time product offer state update requires --confirm")
	}
	return nil
}

type OneTimeProductOfferStateUpdatePlan struct {
	PackageName      PackageName                    `json:"packageName"`
	ProductID        OneTimeProductID               `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID `json:"purchaseOptionId"`
	OfferID          OneTimeProductOfferID          `json:"offerId"`
	Action           OneTimeProductOfferStateAction `json:"action"`
	LatencyTolerance ProductUpdateLatencyTolerance  `json:"latencyTolerance"`
	Confirm          bool                           `json:"confirm"`
	Steps            []string                       `json:"steps"`
}

type OneTimeProductOfferStateUpdateResult struct {
	PackageName      PackageName                        `json:"packageName"`
	ProductID        OneTimeProductID                   `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID     `json:"purchaseOptionId"`
	OfferID          OneTimeProductOfferID              `json:"offerId"`
	Action           OneTimeProductOfferStateAction     `json:"action"`
	DryRun           bool                               `json:"dryRun"`
	Applied          bool                               `json:"applied"`
	Offer            *OneTimeProductOffer               `json:"offer,omitempty"`
	Plan             OneTimeProductOfferStateUpdatePlan `json:"plan"`
}

type OneTimeProductOfferStateUpdater interface {
	UpdateOneTimeProductOfferState(ctx context.Context, options OneTimeProductOfferStateUpdateOptions) (OneTimeProductOffer, error)
}

func NewOneTimeProductOfferStateUpdatePlan(options OneTimeProductOfferStateUpdateOptions) (OneTimeProductOfferStateUpdatePlan, error) {
	if err := options.Validate(); err != nil {
		return OneTimeProductOfferStateUpdatePlan{}, err
	}
	return OneTimeProductOfferStateUpdatePlan{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		PurchaseOptionID: options.PurchaseOptionID,
		OfferID:          options.OfferID,
		Action:           options.Action,
		LatencyTolerance: options.LatencyTolerance,
		Confirm:          options.Confirm,
		Steps:            oneTimeProductOfferStateUpdateSteps(options),
	}, nil
}

func UpdateOneTimeProductOfferState(ctx context.Context, updater OneTimeProductOfferStateUpdater, options OneTimeProductOfferStateUpdateOptions) (OneTimeProductOfferStateUpdateResult, error) {
	plan, err := NewOneTimeProductOfferStateUpdatePlan(options)
	if err != nil {
		return OneTimeProductOfferStateUpdateResult{}, err
	}
	result := OneTimeProductOfferStateUpdateResult{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		PurchaseOptionID: options.PurchaseOptionID,
		OfferID:          options.OfferID,
		Action:           options.Action,
		DryRun:           options.DryRun,
		Applied:          false,
		Plan:             plan,
	}
	if options.DryRun {
		return result, nil
	}
	if updater == nil {
		return OneTimeProductOfferStateUpdateResult{}, fmt.Errorf("one-time product offer state updater is required")
	}
	offer, err := updater.UpdateOneTimeProductOfferState(ctx, options)
	if err != nil {
		return OneTimeProductOfferStateUpdateResult{}, err
	}
	result.Applied = true
	result.Offer = &offer
	return result, nil
}

func oneTimeProductOfferStateUpdateSteps(options OneTimeProductOfferStateUpdateOptions) []string {
	if options.DryRun {
		return []string{fmt.Sprintf("plan %s one-time product offer", options.Action)}
	}
	return []string{fmt.Sprintf("%s one-time product offer", options.Action)}
}
