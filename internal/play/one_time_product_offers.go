package play

import (
	"context"
	"fmt"
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

type OneTimeProductOfferBatchGetOptions struct {
	PackageName      PackageName                          `json:"packageName"`
	ProductID        OneTimeProductID                     `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID       `json:"purchaseOptionId"`
	Requests         []OneTimeProductOfferBatchGetRequest `json:"requests"`
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

type OneTimeProductOfferGetter interface {
	GetOneTimeProductOffer(ctx context.Context, options OneTimeProductOfferGetOptions) (OneTimeProductOffer, error)
}

type OneTimeProductOfferBatchGetter interface {
	BatchGetOneTimeProductOffers(ctx context.Context, options OneTimeProductOfferBatchGetOptions) (OneTimeProductOfferBatchGetResult, error)
}

type OneTimeProductOfferBatchGetResult struct {
	PackageName      PackageName                        `json:"packageName"`
	ProductID        OneTimeProductID                   `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID     `json:"purchaseOptionId"`
	Offers           []OneTimeProductOffer              `json:"offers"`
	Options          OneTimeProductOfferBatchGetOptions `json:"options"`
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
