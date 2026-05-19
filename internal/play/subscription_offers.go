package play

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"
)

type SubscriptionBasePlanID string

const SubscriptionOfferWildcardID = "-"

var (
	subscriptionOfferTagPattern = regexp.MustCompile(`^[a-z]([a-z0-9-]{0,18}[a-z0-9])?$`)
	isoDurationPattern          = regexp.MustCompile(`^P(?:\d+(?:[.,]\d+)?Y)?(?:\d+(?:[.,]\d+)?M)?(?:\d+(?:[.,]\d+)?W)?(?:\d+(?:[.,]\d+)?D)?(?:T(?:\d+(?:[.,]\d+)?H)?(?:\d+(?:[.,]\d+)?M)?(?:\d+(?:[.,]\d+)?S)?)?$`)
)

func NewSubscriptionBasePlanID(value string) (SubscriptionBasePlanID, error) {
	if value == "" {
		return "", fmt.Errorf("subscription base plan ID is required")
	}
	if len(value) > 63 {
		return "", fmt.Errorf("subscription base plan ID cannot exceed 63 characters")
	}
	if !isValidSubscriptionBasePlanID(value) {
		return "", fmt.Errorf("invalid subscription base plan ID %q", value)
	}
	return SubscriptionBasePlanID(value), nil
}

func (b SubscriptionBasePlanID) String() string {
	return string(b)
}

func NewSubscriptionOfferListProductID(value string) (SubscriptionProductID, error) {
	if value == SubscriptionOfferWildcardID {
		return SubscriptionProductID(value), nil
	}
	return NewSubscriptionProductID(value)
}

func NewSubscriptionOfferListBasePlanID(value string) (SubscriptionBasePlanID, error) {
	if value == SubscriptionOfferWildcardID {
		return SubscriptionBasePlanID(value), nil
	}
	return NewSubscriptionBasePlanID(value)
}

func isValidSubscriptionBasePlanID(value string) bool {
	first := value[0]
	if !isASCIILower(first) && !isASCIIDigit(first) {
		return false
	}
	last := value[len(value)-1]
	if !isASCIILower(last) && !isASCIIDigit(last) {
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

type SubscriptionOfferID string

func NewSubscriptionOfferID(value string) (SubscriptionOfferID, error) {
	if value == "" {
		return "", fmt.Errorf("subscription offer ID is required")
	}
	if len(value) > 63 {
		return "", fmt.Errorf("subscription offer ID cannot exceed 63 characters")
	}
	if !isValidSubscriptionBasePlanID(value) {
		return "", fmt.Errorf("invalid subscription offer ID %q", value)
	}
	return SubscriptionOfferID(value), nil
}

func (o SubscriptionOfferID) String() string {
	return string(o)
}

type SubscriptionOfferState string

const (
	SubscriptionOfferStateDraft       SubscriptionOfferState = "DRAFT"
	SubscriptionOfferStateActive      SubscriptionOfferState = "ACTIVE"
	SubscriptionOfferStateInactive    SubscriptionOfferState = "INACTIVE"
	SubscriptionOfferStateUnspecified SubscriptionOfferState = "STATE_UNSPECIFIED"
)

type SubscriptionOffer struct {
	PackageName        PackageName                          `json:"packageName"`
	ProductID          SubscriptionProductID                `json:"productId"`
	BasePlanID         SubscriptionBasePlanID               `json:"basePlanId"`
	OfferID            SubscriptionOfferID                  `json:"offerId"`
	State              SubscriptionOfferState               `json:"state,omitempty"`
	OfferTags          []string                             `json:"offerTags,omitempty"`
	RegionalConfigs    []SubscriptionOfferRegionalConfig    `json:"regionalConfigs"`
	OtherRegionsConfig *SubscriptionOfferOtherRegionsConfig `json:"otherRegionsConfig,omitempty"`
	Phases             []SubscriptionOfferPhase             `json:"phases"`
	Targeting          *SubscriptionOfferTargeting          `json:"targeting,omitempty"`
}

type SubscriptionOfferStateAction string

const (
	SubscriptionOfferStateActionActivate   SubscriptionOfferStateAction = "activate"
	SubscriptionOfferStateActionDeactivate SubscriptionOfferStateAction = "deactivate"
)

func (a SubscriptionOfferStateAction) String() string {
	return string(a)
}

func (a SubscriptionOfferStateAction) Validate() error {
	switch a {
	case SubscriptionOfferStateActionActivate, SubscriptionOfferStateActionDeactivate:
		return nil
	default:
		return fmt.Errorf("unsupported subscription offer state action %q", a)
	}
}

type SubscriptionOfferRegionalConfig struct {
	RegionCode                string `json:"regionCode"`
	NewSubscriberAvailability bool   `json:"newSubscriberAvailability,omitempty"`
}

type SubscriptionOfferOtherRegionsConfig struct {
	NewSubscriberAvailability bool `json:"newSubscriberAvailability,omitempty"`
}

type SubscriptionOfferPhase struct {
	Duration           string                                    `json:"duration,omitempty"`
	RecurrenceCount    int64                                     `json:"recurrenceCount,omitempty"`
	RegionalConfigs    []SubscriptionOfferPhaseRegionalConfig    `json:"regionalConfigs"`
	OtherRegionsConfig *SubscriptionOfferPhaseOtherRegionsConfig `json:"otherRegionsConfig,omitempty"`
}

type SubscriptionOfferPhaseRegionalConfig struct {
	RegionCode       string  `json:"regionCode"`
	Price            *Money  `json:"price,omitempty"`
	AbsoluteDiscount *Money  `json:"absoluteDiscount,omitempty"`
	RelativeDiscount float64 `json:"relativeDiscount,omitempty"`
	Free             bool    `json:"free,omitempty"`
}

type SubscriptionOfferPhaseOtherRegionsConfig struct {
	OtherRegionsPrices *SubscriptionOfferOtherRegionsPrices `json:"otherRegionsPrices,omitempty"`
	AbsoluteDiscounts  *SubscriptionOfferOtherRegionsPrices `json:"absoluteDiscounts,omitempty"`
	RelativeDiscount   float64                              `json:"relativeDiscount,omitempty"`
	Free               bool                                 `json:"free,omitempty"`
}

type SubscriptionOfferOtherRegionsPrices struct {
	USDPrice *Money `json:"usdPrice,omitempty"`
	EURPrice *Money `json:"eurPrice,omitempty"`
}

type SubscriptionOfferTargeting struct {
	Acquisition *SubscriptionOfferAcquisitionTargeting `json:"acquisition,omitempty"`
	Upgrade     *SubscriptionOfferUpgradeTargeting     `json:"upgrade,omitempty"`
}

type SubscriptionOfferAcquisitionTargeting struct {
	Scope *SubscriptionOfferTargetingScope `json:"scope,omitempty"`
}

type SubscriptionOfferUpgradeTargeting struct {
	Scope                 *SubscriptionOfferTargetingScope `json:"scope,omitempty"`
	BillingPeriodDuration string                           `json:"billingPeriodDuration,omitempty"`
	OncePerUser           bool                             `json:"oncePerUser,omitempty"`
}

type SubscriptionOfferTargetingScope struct {
	AnySubscriptionInApp      bool   `json:"anySubscriptionInApp,omitempty"`
	ThisSubscription          bool   `json:"thisSubscription,omitempty"`
	SpecificSubscriptionInApp string `json:"specificSubscriptionInApp,omitempty"`
}

type SubscriptionOfferListOptions struct {
	PackageName PackageName            `json:"packageName"`
	ProductID   SubscriptionProductID  `json:"productId"`
	BasePlanID  SubscriptionBasePlanID `json:"basePlanId"`
	PageSize    int64                  `json:"pageSize,omitempty"`
	PageToken   string                 `json:"pageToken,omitempty"`
}

func (o SubscriptionOfferListOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewSubscriptionOfferListProductID(o.ProductID.String()); err != nil {
		return err
	}
	if _, err := NewSubscriptionOfferListBasePlanID(o.BasePlanID.String()); err != nil {
		return err
	}
	if o.ProductID.String() == SubscriptionOfferWildcardID && o.BasePlanID.String() != SubscriptionOfferWildcardID {
		return fmt.Errorf("subscription base plan ID must be %q when product ID is %q", SubscriptionOfferWildcardID, SubscriptionOfferWildcardID)
	}
	if o.PageSize < 0 {
		return fmt.Errorf("page size cannot be negative")
	}
	if o.PageSize > 1000 {
		return fmt.Errorf("page size cannot exceed 1000")
	}
	return nil
}

type SubscriptionOfferListResult struct {
	PackageName   PackageName                  `json:"packageName"`
	ProductID     SubscriptionProductID        `json:"productId"`
	BasePlanID    SubscriptionBasePlanID       `json:"basePlanId"`
	Offers        []SubscriptionOffer          `json:"offers"`
	NextPageToken string                       `json:"nextPageToken,omitempty"`
	Options       SubscriptionOfferListOptions `json:"options"`
}

type SubscriptionOfferLister interface {
	ListSubscriptionOffers(ctx context.Context, options SubscriptionOfferListOptions) (SubscriptionOfferListResult, error)
}

func ListSubscriptionOffers(ctx context.Context, lister SubscriptionOfferLister, options SubscriptionOfferListOptions) (SubscriptionOfferListResult, error) {
	if err := options.Validate(); err != nil {
		return SubscriptionOfferListResult{}, err
	}
	if lister == nil {
		return SubscriptionOfferListResult{}, fmt.Errorf("subscription offer lister is required")
	}
	return lister.ListSubscriptionOffers(ctx, options)
}

type SubscriptionOfferGetOptions struct {
	PackageName PackageName            `json:"packageName"`
	ProductID   SubscriptionProductID  `json:"productId"`
	BasePlanID  SubscriptionBasePlanID `json:"basePlanId"`
	OfferID     SubscriptionOfferID    `json:"offerId"`
}

func (o SubscriptionOfferGetOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewSubscriptionProductID(o.ProductID.String()); err != nil {
		return err
	}
	if _, err := NewSubscriptionBasePlanID(o.BasePlanID.String()); err != nil {
		return err
	}
	if _, err := NewSubscriptionOfferID(o.OfferID.String()); err != nil {
		return err
	}
	return nil
}

type SubscriptionOfferGetter interface {
	GetSubscriptionOffer(ctx context.Context, packageName PackageName, productID SubscriptionProductID, basePlanID SubscriptionBasePlanID, offerID SubscriptionOfferID) (SubscriptionOffer, error)
}

func GetSubscriptionOffer(ctx context.Context, getter SubscriptionOfferGetter, options SubscriptionOfferGetOptions) (SubscriptionOffer, error) {
	if err := options.Validate(); err != nil {
		return SubscriptionOffer{}, err
	}
	if getter == nil {
		return SubscriptionOffer{}, fmt.Errorf("subscription offer getter is required")
	}
	return getter.GetSubscriptionOffer(ctx, options.PackageName, options.ProductID, options.BasePlanID, options.OfferID)
}

type SubscriptionOfferCreateOptions struct {
	PackageName    PackageName            `json:"packageName"`
	ProductID      SubscriptionProductID  `json:"productId"`
	BasePlanID     SubscriptionBasePlanID `json:"basePlanId"`
	OfferID        SubscriptionOfferID    `json:"offerId"`
	Offer          SubscriptionOffer      `json:"offer"`
	RegionsVersion string                 `json:"regionsVersion"`
	Confirm        bool                   `json:"confirm"`
	DryRun         bool                   `json:"dryRun"`
}

type SubscriptionOfferBatchGetRequest struct {
	ProductID  SubscriptionProductID  `json:"productId"`
	BasePlanID SubscriptionBasePlanID `json:"basePlanId"`
	OfferID    SubscriptionOfferID    `json:"offerId"`
}

type SubscriptionOfferBatchMutationRequest = SubscriptionOfferBatchGetRequest

func NewSubscriptionOfferBatchGetRequest(value string) (SubscriptionOfferBatchGetRequest, error) {
	parts := strings.Split(strings.TrimSpace(value), "/")
	if len(parts) != 3 {
		return SubscriptionOfferBatchGetRequest{}, fmt.Errorf("subscription offer must use productId/basePlanId/offerId")
	}
	productID, err := NewSubscriptionProductID(parts[0])
	if err != nil {
		return SubscriptionOfferBatchGetRequest{}, err
	}
	basePlanID, err := NewSubscriptionBasePlanID(parts[1])
	if err != nil {
		return SubscriptionOfferBatchGetRequest{}, err
	}
	offerID, err := NewSubscriptionOfferID(parts[2])
	if err != nil {
		return SubscriptionOfferBatchGetRequest{}, err
	}
	return SubscriptionOfferBatchGetRequest{ProductID: productID, BasePlanID: basePlanID, OfferID: offerID}, nil
}

func NewSubscriptionOfferBatchMutationRequest(value string) (SubscriptionOfferBatchMutationRequest, error) {
	return NewSubscriptionOfferBatchGetRequest(value)
}

type SubscriptionOfferBatchGetOptions struct {
	PackageName PackageName                        `json:"packageName"`
	ProductID   SubscriptionProductID              `json:"productId"`
	BasePlanID  SubscriptionBasePlanID             `json:"basePlanId"`
	Requests    []SubscriptionOfferBatchGetRequest `json:"requests"`
}

type SubscriptionOfferBatchStateUpdateOptions struct {
	PackageName      PackageName                             `json:"packageName"`
	ProductID        SubscriptionProductID                   `json:"productId"`
	BasePlanID       SubscriptionBasePlanID                  `json:"basePlanId"`
	Requests         []SubscriptionOfferBatchMutationRequest `json:"requests"`
	Action           SubscriptionOfferStateAction            `json:"action"`
	LatencyTolerance ProductUpdateLatencyTolerance           `json:"latencyTolerance"`
	Confirm          bool                                    `json:"confirm"`
	DryRun           bool                                    `json:"dryRun"`
}

type SubscriptionOfferAvailabilityPatchRequest struct {
	ProductID    SubscriptionProductID  `json:"productId"`
	BasePlanID   SubscriptionBasePlanID `json:"basePlanId"`
	OfferID      SubscriptionOfferID    `json:"offerId"`
	RegionCode   string                 `json:"regionCode"`
	Availability bool                   `json:"availability"`
}

type SubscriptionOfferPhaseRelativeDiscountPatchRequest struct {
	ProductID        SubscriptionProductID  `json:"productId"`
	BasePlanID       SubscriptionBasePlanID `json:"basePlanId"`
	OfferID          SubscriptionOfferID    `json:"offerId"`
	PhaseIndex       int                    `json:"phaseIndex"`
	RegionCode       string                 `json:"regionCode"`
	RelativeDiscount float64                `json:"relativeDiscount"`
}

type SubscriptionOfferPhaseAbsoluteDiscountPatchRequest struct {
	ProductID        SubscriptionProductID  `json:"productId"`
	BasePlanID       SubscriptionBasePlanID `json:"basePlanId"`
	OfferID          SubscriptionOfferID    `json:"offerId"`
	PhaseIndex       int                    `json:"phaseIndex"`
	RegionCode       string                 `json:"regionCode"`
	AbsoluteDiscount Money                  `json:"absoluteDiscount"`
}

type SubscriptionOfferPhasePricePatchRequest struct {
	ProductID  SubscriptionProductID  `json:"productId"`
	BasePlanID SubscriptionBasePlanID `json:"basePlanId"`
	OfferID    SubscriptionOfferID    `json:"offerId"`
	PhaseIndex int                    `json:"phaseIndex"`
	RegionCode string                 `json:"regionCode"`
	Price      Money                  `json:"price"`
}

type SubscriptionOfferPhaseFreePatchRequest struct {
	ProductID  SubscriptionProductID  `json:"productId"`
	BasePlanID SubscriptionBasePlanID `json:"basePlanId"`
	OfferID    SubscriptionOfferID    `json:"offerId"`
	PhaseIndex int                    `json:"phaseIndex"`
	RegionCode string                 `json:"regionCode"`
}

type SubscriptionOfferBatchPatchAvailabilityOptions struct {
	PackageName      PackageName                                 `json:"packageName"`
	ProductID        SubscriptionProductID                       `json:"productId"`
	BasePlanID       SubscriptionBasePlanID                      `json:"basePlanId"`
	Requests         []SubscriptionOfferAvailabilityPatchRequest `json:"requests"`
	RegionsVersion   string                                      `json:"regionsVersion"`
	LatencyTolerance ProductUpdateLatencyTolerance               `json:"latencyTolerance"`
	Confirm          bool                                        `json:"confirm"`
	DryRun           bool                                        `json:"dryRun"`
}

type SubscriptionOfferBatchPatchPhaseRelativeDiscountsOptions struct {
	PackageName      PackageName                                          `json:"packageName"`
	ProductID        SubscriptionProductID                                `json:"productId"`
	BasePlanID       SubscriptionBasePlanID                               `json:"basePlanId"`
	Requests         []SubscriptionOfferPhaseRelativeDiscountPatchRequest `json:"requests"`
	RegionsVersion   string                                               `json:"regionsVersion"`
	LatencyTolerance ProductUpdateLatencyTolerance                        `json:"latencyTolerance"`
	Confirm          bool                                                 `json:"confirm"`
	DryRun           bool                                                 `json:"dryRun"`
}

type SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsOptions struct {
	PackageName      PackageName                                          `json:"packageName"`
	ProductID        SubscriptionProductID                                `json:"productId"`
	BasePlanID       SubscriptionBasePlanID                               `json:"basePlanId"`
	Requests         []SubscriptionOfferPhaseAbsoluteDiscountPatchRequest `json:"requests"`
	RegionsVersion   string                                               `json:"regionsVersion"`
	LatencyTolerance ProductUpdateLatencyTolerance                        `json:"latencyTolerance"`
	Confirm          bool                                                 `json:"confirm"`
	DryRun           bool                                                 `json:"dryRun"`
}

type SubscriptionOfferBatchPatchPhasePricesOptions struct {
	PackageName      PackageName                               `json:"packageName"`
	ProductID        SubscriptionProductID                     `json:"productId"`
	BasePlanID       SubscriptionBasePlanID                    `json:"basePlanId"`
	Requests         []SubscriptionOfferPhasePricePatchRequest `json:"requests"`
	RegionsVersion   string                                    `json:"regionsVersion"`
	LatencyTolerance ProductUpdateLatencyTolerance             `json:"latencyTolerance"`
	Confirm          bool                                      `json:"confirm"`
	DryRun           bool                                      `json:"dryRun"`
}

type SubscriptionOfferBatchPatchPhaseFreeOptions struct {
	PackageName      PackageName                              `json:"packageName"`
	ProductID        SubscriptionProductID                    `json:"productId"`
	BasePlanID       SubscriptionBasePlanID                   `json:"basePlanId"`
	Requests         []SubscriptionOfferPhaseFreePatchRequest `json:"requests"`
	RegionsVersion   string                                   `json:"regionsVersion"`
	LatencyTolerance ProductUpdateLatencyTolerance            `json:"latencyTolerance"`
	Confirm          bool                                     `json:"confirm"`
	DryRun           bool                                     `json:"dryRun"`
}

func (o SubscriptionOfferBatchGetOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewSubscriptionOfferListProductID(o.ProductID.String()); err != nil {
		return err
	}
	if _, err := NewSubscriptionOfferListBasePlanID(o.BasePlanID.String()); err != nil {
		return err
	}
	if o.ProductID.String() == SubscriptionOfferWildcardID && o.BasePlanID.String() != SubscriptionOfferWildcardID {
		return fmt.Errorf("subscription base plan ID must be %q when product ID is %q", SubscriptionOfferWildcardID, SubscriptionOfferWildcardID)
	}
	if len(o.Requests) == 0 {
		return fmt.Errorf("at least one subscription offer is required")
	}
	if len(o.Requests) > 100 {
		return fmt.Errorf("subscription offer batch-get cannot exceed 100 offers")
	}
	seen := map[string]struct{}{}
	for _, request := range o.Requests {
		if _, err := NewSubscriptionProductID(request.ProductID.String()); err != nil {
			return err
		}
		if _, err := NewSubscriptionBasePlanID(request.BasePlanID.String()); err != nil {
			return err
		}
		if _, err := NewSubscriptionOfferID(request.OfferID.String()); err != nil {
			return err
		}
		if o.ProductID.String() != SubscriptionOfferWildcardID && request.ProductID != o.ProductID {
			return fmt.Errorf("subscription offer %s/%s/%s does not match parent product ID %s", request.ProductID, request.BasePlanID, request.OfferID, o.ProductID)
		}
		if o.BasePlanID.String() != SubscriptionOfferWildcardID && request.BasePlanID != o.BasePlanID {
			return fmt.Errorf("subscription offer %s/%s/%s does not match parent base plan ID %s", request.ProductID, request.BasePlanID, request.OfferID, o.BasePlanID)
		}
		key := request.ProductID.String() + "/" + request.BasePlanID.String() + "/" + request.OfferID.String()
		if _, ok := seen[key]; ok {
			return fmt.Errorf("subscription offer %s is duplicated", key)
		}
		seen[key] = struct{}{}
	}
	return nil
}

func (o SubscriptionOfferCreateOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewSubscriptionProductID(o.ProductID.String()); err != nil {
		return err
	}
	if _, err := NewSubscriptionBasePlanID(o.BasePlanID.String()); err != nil {
		return err
	}
	if _, err := NewSubscriptionOfferID(o.OfferID.String()); err != nil {
		return err
	}
	if strings.TrimSpace(o.RegionsVersion) == "" {
		return fmt.Errorf("regions version is required")
	}
	if strings.TrimSpace(o.RegionsVersion) != o.RegionsVersion {
		return fmt.Errorf("regions version cannot have leading or trailing whitespace")
	}
	offer := subscriptionOfferCreateDesiredOffer(o)
	if err := validateSubscriptionOfferForCreate(offer); err != nil {
		return err
	}
	if o.Confirm && o.DryRun {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.Confirm && !o.DryRun {
		return fmt.Errorf("subscription offer create requires --confirm or --dry-run")
	}
	return nil
}

func (o SubscriptionOfferCreateOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live subscription offer create cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live subscription offer create requires --confirm")
	}
	return nil
}

func (o SubscriptionOfferBatchStateUpdateOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if err := validateSubscriptionOfferBatchMutationParents(o.ProductID, o.BasePlanID, o.Requests, "batch state update"); err != nil {
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
		return fmt.Errorf("subscription offer batch state update requires --confirm or --dry-run")
	}
	return nil
}

func (o SubscriptionOfferBatchPatchAvailabilityOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if err := validateSubscriptionOfferAvailabilityPatchParents(o.ProductID, o.BasePlanID, o.Requests); err != nil {
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
		return fmt.Errorf("subscription offer availability batch patch requires --confirm or --dry-run")
	}
	return nil
}

func (o SubscriptionOfferBatchPatchPhaseRelativeDiscountsOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if err := validateSubscriptionOfferPhaseRelativeDiscountPatchParents(o.ProductID, o.BasePlanID, o.Requests); err != nil {
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
		return fmt.Errorf("subscription offer phase relative discount batch patch requires --confirm or --dry-run")
	}
	return nil
}

func (o SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if err := validateSubscriptionOfferPhaseAbsoluteDiscountPatchParents(o.ProductID, o.BasePlanID, o.Requests); err != nil {
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
		return fmt.Errorf("subscription offer phase absolute discount batch patch requires --confirm or --dry-run")
	}
	return nil
}

func (o SubscriptionOfferBatchPatchPhasePricesOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if err := validateSubscriptionOfferPhasePricePatchParents(o.ProductID, o.BasePlanID, o.Requests); err != nil {
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
		return fmt.Errorf("subscription offer phase price batch patch requires --confirm or --dry-run")
	}
	return nil
}

func (o SubscriptionOfferBatchPatchPhaseFreeOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if err := validateSubscriptionOfferPhaseFreePatchParents(o.ProductID, o.BasePlanID, o.Requests); err != nil {
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
		return fmt.Errorf("subscription offer phase free batch patch requires --confirm or --dry-run")
	}
	return nil
}

func (o SubscriptionOfferBatchStateUpdateOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live subscription offer batch state update cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live subscription offer batch state update requires --confirm")
	}
	return nil
}

func (o SubscriptionOfferBatchPatchAvailabilityOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live subscription offer availability batch patch cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live subscription offer availability batch patch requires --confirm")
	}
	return nil
}

func (o SubscriptionOfferBatchPatchPhaseRelativeDiscountsOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live subscription offer phase relative discount batch patch cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live subscription offer phase relative discount batch patch requires --confirm")
	}
	return nil
}

func (o SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live subscription offer phase absolute discount batch patch cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live subscription offer phase absolute discount batch patch requires --confirm")
	}
	return nil
}

func (o SubscriptionOfferBatchPatchPhasePricesOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live subscription offer phase price batch patch cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live subscription offer phase price batch patch requires --confirm")
	}
	return nil
}

func (o SubscriptionOfferBatchPatchPhaseFreeOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live subscription offer phase free batch patch cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live subscription offer phase free batch patch requires --confirm")
	}
	return nil
}

func validateSubscriptionOfferBatchMutationParents(productID SubscriptionProductID, basePlanID SubscriptionBasePlanID, requests []SubscriptionOfferBatchMutationRequest, operation string) error {
	if len(requests) == 0 {
		return fmt.Errorf("at least one subscription offer is required")
	}
	if _, err := NewSubscriptionOfferListProductID(productID.String()); err != nil {
		return err
	}
	if _, err := NewSubscriptionOfferListBasePlanID(basePlanID.String()); err != nil {
		return err
	}
	if len(requests) > 100 {
		return fmt.Errorf("subscription offer %s cannot exceed 100 offers", operation)
	}
	seen := map[string]struct{}{}
	seenProducts := map[SubscriptionProductID]struct{}{}
	seenBasePlans := map[SubscriptionBasePlanID]struct{}{}
	for _, request := range requests {
		if _, err := NewSubscriptionProductID(request.ProductID.String()); err != nil {
			return err
		}
		if _, err := NewSubscriptionBasePlanID(request.BasePlanID.String()); err != nil {
			return err
		}
		if _, err := NewSubscriptionOfferID(request.OfferID.String()); err != nil {
			return err
		}
		if productID.String() != SubscriptionOfferWildcardID && request.ProductID != productID {
			return fmt.Errorf("subscription offer %s/%s/%s does not match parent product ID %s", request.ProductID, request.BasePlanID, request.OfferID, productID)
		}
		if basePlanID.String() != SubscriptionOfferWildcardID && request.BasePlanID != basePlanID {
			return fmt.Errorf("subscription offer %s/%s/%s does not match parent base plan ID %s", request.ProductID, request.BasePlanID, request.OfferID, basePlanID)
		}
		key := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("subscription offer %s is duplicated", key)
		}
		seen[key] = struct{}{}
		seenProducts[request.ProductID] = struct{}{}
		seenBasePlans[request.BasePlanID] = struct{}{}
	}
	if len(seenProducts) == 1 && productID.String() == SubscriptionOfferWildcardID {
		return fmt.Errorf("single-product offer %s requires parent product ID, not %q", operation, SubscriptionOfferWildcardID)
	}
	if len(seenProducts) > 1 && productID.String() != SubscriptionOfferWildcardID {
		return fmt.Errorf("multi-product offer %s requires parent product ID %q", operation, SubscriptionOfferWildcardID)
	}
	if len(seenProducts) == 1 && len(seenBasePlans) == 1 && basePlanID.String() == SubscriptionOfferWildcardID {
		return fmt.Errorf("single-base-plan offer %s requires parent base plan ID, not %q", operation, SubscriptionOfferWildcardID)
	}
	if len(seenProducts) == 1 && len(seenBasePlans) > 1 && basePlanID.String() != SubscriptionOfferWildcardID {
		return fmt.Errorf("multi-base-plan offer %s requires parent base plan ID %q", operation, SubscriptionOfferWildcardID)
	}
	return nil
}

func validateSubscriptionOfferAvailabilityPatchParents(productID SubscriptionProductID, basePlanID SubscriptionBasePlanID, requests []SubscriptionOfferAvailabilityPatchRequest) error {
	if len(requests) == 0 {
		return fmt.Errorf("at least one subscription offer availability patch is required")
	}
	mutationRequests := make([]SubscriptionOfferBatchMutationRequest, 0, len(requests))
	seenRegions := map[string]struct{}{}
	for _, request := range requests {
		if !isValidRegionCode(request.RegionCode) {
			return fmt.Errorf("subscription offer availability region must be a two-letter ISO 3166 code")
		}
		key := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID) + "/" + request.RegionCode
		if _, ok := seenRegions[key]; ok {
			return fmt.Errorf("subscription offer availability %s is duplicated", key)
		}
		seenRegions[key] = struct{}{}
		mutationRequests = append(mutationRequests, SubscriptionOfferBatchMutationRequest{
			ProductID:  request.ProductID,
			BasePlanID: request.BasePlanID,
			OfferID:    request.OfferID,
		})
	}
	return validateSubscriptionOfferBatchMutationParents(productID, basePlanID, deduplicateSubscriptionOfferMutationRequests(mutationRequests), "availability batch patch")
}

func validateSubscriptionOfferPhaseRelativeDiscountPatchParents(productID SubscriptionProductID, basePlanID SubscriptionBasePlanID, requests []SubscriptionOfferPhaseRelativeDiscountPatchRequest) error {
	if len(requests) == 0 {
		return fmt.Errorf("at least one subscription offer phase relative discount patch is required")
	}
	mutationRequests := make([]SubscriptionOfferBatchMutationRequest, 0, len(requests))
	seenRegions := map[string]struct{}{}
	for _, request := range requests {
		if request.PhaseIndex < 0 {
			return fmt.Errorf("subscription offer phase index must be 0 or greater")
		}
		if !isValidRegionCode(request.RegionCode) {
			return fmt.Errorf("subscription offer phase relative discount region must be a two-letter ISO 3166 code")
		}
		if math.IsNaN(request.RelativeDiscount) || math.IsInf(request.RelativeDiscount, 0) || request.RelativeDiscount <= 0 || request.RelativeDiscount >= 1 {
			return fmt.Errorf("subscription offer phase relative discount must be greater than 0 and less than 1")
		}
		key := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID) + fmt.Sprintf("/%d/%s", request.PhaseIndex, request.RegionCode)
		if _, ok := seenRegions[key]; ok {
			return fmt.Errorf("subscription offer phase relative discount %s is duplicated", key)
		}
		seenRegions[key] = struct{}{}
		mutationRequests = append(mutationRequests, SubscriptionOfferBatchMutationRequest{
			ProductID:  request.ProductID,
			BasePlanID: request.BasePlanID,
			OfferID:    request.OfferID,
		})
	}
	return validateSubscriptionOfferBatchMutationParents(productID, basePlanID, deduplicateSubscriptionOfferMutationRequests(mutationRequests), "phase relative discount batch patch")
}

func validateSubscriptionOfferPhaseAbsoluteDiscountPatchParents(productID SubscriptionProductID, basePlanID SubscriptionBasePlanID, requests []SubscriptionOfferPhaseAbsoluteDiscountPatchRequest) error {
	if len(requests) == 0 {
		return fmt.Errorf("at least one subscription offer phase absolute discount patch is required")
	}
	mutationRequests := make([]SubscriptionOfferBatchMutationRequest, 0, len(requests))
	seenRegions := map[string]struct{}{}
	for _, request := range requests {
		if request.PhaseIndex < 0 {
			return fmt.Errorf("subscription offer phase index must be 0 or greater")
		}
		if !isValidRegionCode(request.RegionCode) {
			return fmt.Errorf("subscription offer phase absolute discount region must be a two-letter ISO 3166 code")
		}
		if err := validateMoney(request.AbsoluteDiscount); err != nil {
			return fmt.Errorf("subscription offer phase absolute discount for %s/%s/%s/%d/%s: %w", request.ProductID, request.BasePlanID, request.OfferID, request.PhaseIndex, request.RegionCode, err)
		}
		key := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID) + fmt.Sprintf("/%d/%s", request.PhaseIndex, request.RegionCode)
		if _, ok := seenRegions[key]; ok {
			return fmt.Errorf("subscription offer phase absolute discount %s is duplicated", key)
		}
		seenRegions[key] = struct{}{}
		mutationRequests = append(mutationRequests, SubscriptionOfferBatchMutationRequest{
			ProductID:  request.ProductID,
			BasePlanID: request.BasePlanID,
			OfferID:    request.OfferID,
		})
	}
	return validateSubscriptionOfferBatchMutationParents(productID, basePlanID, deduplicateSubscriptionOfferMutationRequests(mutationRequests), "phase absolute discount batch patch")
}

func validateSubscriptionOfferPhasePricePatchParents(productID SubscriptionProductID, basePlanID SubscriptionBasePlanID, requests []SubscriptionOfferPhasePricePatchRequest) error {
	if len(requests) == 0 {
		return fmt.Errorf("at least one subscription offer phase price patch is required")
	}
	mutationRequests := make([]SubscriptionOfferBatchMutationRequest, 0, len(requests))
	seenRegions := map[string]struct{}{}
	for _, request := range requests {
		if request.PhaseIndex < 0 {
			return fmt.Errorf("subscription offer phase index must be 0 or greater")
		}
		if !isValidRegionCode(request.RegionCode) {
			return fmt.Errorf("subscription offer phase price region must be a two-letter ISO 3166 code")
		}
		if err := validateMoney(request.Price); err != nil {
			return fmt.Errorf("subscription offer phase price for %s/%s/%s/%d/%s: %w", request.ProductID, request.BasePlanID, request.OfferID, request.PhaseIndex, request.RegionCode, err)
		}
		key := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID) + fmt.Sprintf("/%d/%s", request.PhaseIndex, request.RegionCode)
		if _, ok := seenRegions[key]; ok {
			return fmt.Errorf("subscription offer phase price %s is duplicated", key)
		}
		seenRegions[key] = struct{}{}
		mutationRequests = append(mutationRequests, SubscriptionOfferBatchMutationRequest{
			ProductID:  request.ProductID,
			BasePlanID: request.BasePlanID,
			OfferID:    request.OfferID,
		})
	}
	return validateSubscriptionOfferBatchMutationParents(productID, basePlanID, deduplicateSubscriptionOfferMutationRequests(mutationRequests), "phase price batch patch")
}

func validateSubscriptionOfferPhaseFreePatchParents(productID SubscriptionProductID, basePlanID SubscriptionBasePlanID, requests []SubscriptionOfferPhaseFreePatchRequest) error {
	if len(requests) == 0 {
		return fmt.Errorf("at least one subscription offer phase free patch is required")
	}
	mutationRequests := make([]SubscriptionOfferBatchMutationRequest, 0, len(requests))
	seenRegions := map[string]struct{}{}
	for _, request := range requests {
		if request.PhaseIndex < 0 {
			return fmt.Errorf("subscription offer phase index must be 0 or greater")
		}
		if !isValidRegionCode(request.RegionCode) {
			return fmt.Errorf("subscription offer phase free region must be a two-letter ISO 3166 code")
		}
		key := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID) + fmt.Sprintf("/%d/%s", request.PhaseIndex, request.RegionCode)
		if _, ok := seenRegions[key]; ok {
			return fmt.Errorf("subscription offer phase free %s is duplicated", key)
		}
		seenRegions[key] = struct{}{}
		mutationRequests = append(mutationRequests, SubscriptionOfferBatchMutationRequest{
			ProductID:  request.ProductID,
			BasePlanID: request.BasePlanID,
			OfferID:    request.OfferID,
		})
	}
	return validateSubscriptionOfferBatchMutationParents(productID, basePlanID, deduplicateSubscriptionOfferMutationRequests(mutationRequests), "phase free batch patch")
}

func validateSubscriptionOfferForCreate(offer SubscriptionOffer) error {
	if len(offer.OfferTags) > 20 {
		return fmt.Errorf("subscription offer create supports at most 20 offer tags")
	}
	seenTags := map[string]struct{}{}
	for _, tag := range offer.OfferTags {
		if err := validateSubscriptionOfferTag(tag); err != nil {
			return err
		}
		if _, ok := seenTags[tag]; ok {
			return fmt.Errorf("subscription offer create offer tag %q is duplicated", tag)
		}
		seenTags[tag] = struct{}{}
	}
	if len(offer.RegionalConfigs) == 0 {
		return fmt.Errorf("subscription offer create requires at least one regional config")
	}
	regions := map[string]struct{}{}
	for _, config := range offer.RegionalConfigs {
		if !isValidRegionCode(config.RegionCode) {
			return fmt.Errorf("subscription offer create region must be a two-letter ISO 3166 code")
		}
		if _, ok := regions[config.RegionCode]; ok {
			return fmt.Errorf("subscription offer create region %s is duplicated", config.RegionCode)
		}
		regions[config.RegionCode] = struct{}{}
	}
	if len(offer.Phases) == 0 || len(offer.Phases) > 2 {
		return fmt.Errorf("subscription offer create requires one or two phases")
	}
	for phaseIndex, phase := range offer.Phases {
		if err := validateISODuration("subscription offer create phase duration", phase.Duration); err != nil {
			return fmt.Errorf("subscription offer create phase %d: %w", phaseIndex, err)
		}
		if phase.RecurrenceCount <= 0 {
			return fmt.Errorf("subscription offer create phase %d recurrence count must be greater than 0", phaseIndex)
		}
		if len(phase.RegionalConfigs) != len(regions) {
			return fmt.Errorf("subscription offer create phase %d must have exactly one regional price config for each offer region", phaseIndex)
		}
		phaseRegions := map[string]struct{}{}
		for _, config := range phase.RegionalConfigs {
			if _, ok := regions[config.RegionCode]; !ok {
				return fmt.Errorf("subscription offer create phase %d region %s is not configured on the offer", phaseIndex, config.RegionCode)
			}
			if _, ok := phaseRegions[config.RegionCode]; ok {
				return fmt.Errorf("subscription offer create phase %d region %s is duplicated", phaseIndex, config.RegionCode)
			}
			phaseRegions[config.RegionCode] = struct{}{}
			if err := validateSubscriptionOfferPhaseRegionalPriceMode(config); err != nil {
				return fmt.Errorf("subscription offer create phase %d region %s: %w", phaseIndex, config.RegionCode, err)
			}
		}
		if phase.OtherRegionsConfig != nil {
			if err := validateSubscriptionOfferPhaseOtherRegionsPriceMode(*phase.OtherRegionsConfig); err != nil {
				return fmt.Errorf("subscription offer create phase %d other regions: %w", phaseIndex, err)
			}
		}
	}
	if offer.Targeting != nil {
		if err := validateSubscriptionOfferTargetingForCreate(*offer.Targeting); err != nil {
			return err
		}
	}
	return nil
}

func validateSubscriptionOfferTag(tag string) error {
	if strings.TrimSpace(tag) == "" {
		return fmt.Errorf("subscription offer create offer tags cannot be empty")
	}
	if len(tag) > 20 {
		return fmt.Errorf("subscription offer create offer tag %q cannot exceed 20 characters", tag)
	}
	if !subscriptionOfferTagPattern.MatchString(tag) {
		return fmt.Errorf("subscription offer create offer tag %q must use RFC-1034 format with lower-case letters, digits, and hyphens", tag)
	}
	return nil
}

func validateSubscriptionOfferPhaseRegionalPriceMode(config SubscriptionOfferPhaseRegionalConfig) error {
	modes := 0
	if config.Price != nil {
		if err := validateMoney(*config.Price); err != nil {
			return err
		}
		modes++
	}
	if config.AbsoluteDiscount != nil {
		if err := validateMoney(*config.AbsoluteDiscount); err != nil {
			return err
		}
		modes++
	}
	if config.RelativeDiscount != 0 {
		if config.RelativeDiscount <= 0 || config.RelativeDiscount >= 1 {
			return fmt.Errorf("relative discount must be greater than 0 and less than 1")
		}
		modes++
	}
	if config.Free {
		modes++
	}
	if modes != 1 {
		return fmt.Errorf("exactly one of price, absoluteDiscount, relativeDiscount, or free is required")
	}
	return nil
}

func validateSubscriptionOfferPhaseOtherRegionsPriceMode(config SubscriptionOfferPhaseOtherRegionsConfig) error {
	modes := 0
	if config.OtherRegionsPrices != nil {
		if err := validateSubscriptionOfferOtherRegionsPrices(*config.OtherRegionsPrices); err != nil {
			return err
		}
		modes++
	}
	if config.AbsoluteDiscounts != nil {
		if err := validateSubscriptionOfferOtherRegionsPrices(*config.AbsoluteDiscounts); err != nil {
			return err
		}
		modes++
	}
	if config.RelativeDiscount != 0 {
		if config.RelativeDiscount <= 0 || config.RelativeDiscount >= 1 {
			return fmt.Errorf("relative discount must be greater than 0 and less than 1")
		}
		modes++
	}
	if config.Free {
		modes++
	}
	if modes != 1 {
		return fmt.Errorf("exactly one of otherRegionsPrices, absoluteDiscounts, relativeDiscount, or free is required")
	}
	return nil
}

func validateSubscriptionOfferOtherRegionsPrices(prices SubscriptionOfferOtherRegionsPrices) error {
	if prices.USDPrice == nil || prices.EURPrice == nil {
		return fmt.Errorf("usdPrice and eurPrice are required")
	}
	if prices.USDPrice.CurrencyCode != "USD" {
		return fmt.Errorf("usdPrice currencyCode must be USD")
	}
	if prices.EURPrice.CurrencyCode != "EUR" {
		return fmt.Errorf("eurPrice currencyCode must be EUR")
	}
	if err := validateMoney(*prices.USDPrice); err != nil {
		return err
	}
	if err := validateMoney(*prices.EURPrice); err != nil {
		return err
	}
	return nil
}

func validateSubscriptionOfferTargetingForCreate(targeting SubscriptionOfferTargeting) error {
	hasAcquisition := targeting.Acquisition != nil
	hasUpgrade := targeting.Upgrade != nil
	if hasAcquisition == hasUpgrade {
		return fmt.Errorf("subscription offer create targeting requires exactly one of acquisition or upgrade")
	}
	if hasAcquisition {
		return validateSubscriptionOfferAcquisitionTargetingForCreate(*targeting.Acquisition)
	}
	return validateSubscriptionOfferUpgradeTargetingForCreate(*targeting.Upgrade)
}

func validateSubscriptionOfferAcquisitionTargetingForCreate(targeting SubscriptionOfferAcquisitionTargeting) error {
	if targeting.Scope == nil {
		return fmt.Errorf("subscription offer create acquisition targeting requires scope")
	}
	scope, err := validateSubscriptionOfferTargetingScope(*targeting.Scope)
	if err != nil {
		return err
	}
	if scope == "specificSubscriptionInApp" {
		return fmt.Errorf("subscription offer create acquisition targeting cannot use specificSubscriptionInApp")
	}
	return nil
}

func validateSubscriptionOfferUpgradeTargetingForCreate(targeting SubscriptionOfferUpgradeTargeting) error {
	if targeting.Scope == nil {
		return fmt.Errorf("subscription offer create upgrade targeting requires scope")
	}
	scope, err := validateSubscriptionOfferTargetingScope(*targeting.Scope)
	if err != nil {
		return err
	}
	if scope == "anySubscriptionInApp" {
		return fmt.Errorf("subscription offer create upgrade targeting cannot use anySubscriptionInApp")
	}
	if targeting.BillingPeriodDuration != "" {
		if err := validateISODuration("subscription offer create upgrade billing period duration", targeting.BillingPeriodDuration); err != nil {
			return err
		}
	}
	return nil
}

func validateSubscriptionOfferTargetingScope(scope SubscriptionOfferTargetingScope) (string, error) {
	modes := 0
	selected := ""
	if scope.AnySubscriptionInApp {
		modes++
		selected = "anySubscriptionInApp"
	}
	if scope.ThisSubscription {
		modes++
		selected = "thisSubscription"
	}
	if scope.SpecificSubscriptionInApp != "" {
		if _, err := NewSubscriptionProductID(scope.SpecificSubscriptionInApp); err != nil {
			return "", err
		}
		modes++
		selected = "specificSubscriptionInApp"
	}
	if modes != 1 {
		return "", fmt.Errorf("subscription offer create targeting scope requires exactly one of anySubscriptionInApp, thisSubscription, or specificSubscriptionInApp")
	}
	return selected, nil
}

func validateISODuration(fieldName, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	if strings.TrimSpace(value) != value {
		return fmt.Errorf("%s cannot have leading or trailing whitespace", fieldName)
	}
	if !isoDurationPattern.MatchString(value) || !strings.ContainsAny(value, "0123456789") || value == "P" || value == "PT" || strings.HasSuffix(value, "T") {
		return fmt.Errorf("%s must use ISO 8601 duration format", fieldName)
	}
	return nil
}

func deduplicateSubscriptionOfferMutationRequests(requests []SubscriptionOfferBatchMutationRequest) []SubscriptionOfferBatchMutationRequest {
	seen := map[string]struct{}{}
	deduplicated := make([]SubscriptionOfferBatchMutationRequest, 0, len(requests))
	for _, request := range requests {
		key := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		deduplicated = append(deduplicated, request)
	}
	return deduplicated
}

type SubscriptionOfferBatchGetResult struct {
	PackageName PackageName                      `json:"packageName"`
	ProductID   SubscriptionProductID            `json:"productId"`
	BasePlanID  SubscriptionBasePlanID           `json:"basePlanId"`
	Offers      []SubscriptionOffer              `json:"offers"`
	Options     SubscriptionOfferBatchGetOptions `json:"options"`
}

type SubscriptionOfferCreatePlan struct {
	Action         string                 `json:"action"`
	PackageName    PackageName            `json:"packageName"`
	ProductID      SubscriptionProductID  `json:"productId"`
	BasePlanID     SubscriptionBasePlanID `json:"basePlanId"`
	OfferID        SubscriptionOfferID    `json:"offerId"`
	RegionsVersion string                 `json:"regionsVersion"`
	Confirm        bool                   `json:"confirm"`
	Steps          []string               `json:"steps"`
}

type SubscriptionOfferCreateResult struct {
	Action  string                      `json:"action"`
	DryRun  bool                        `json:"dryRun"`
	Created bool                        `json:"created"`
	Offer   *SubscriptionOffer          `json:"offer,omitempty"`
	Desired SubscriptionOffer           `json:"desiredOffer"`
	Plan    SubscriptionOfferCreatePlan `json:"plan"`
}

type SubscriptionOfferBatchStateUpdatePlan struct {
	PackageName      PackageName                             `json:"packageName"`
	ProductID        SubscriptionProductID                   `json:"productId"`
	BasePlanID       SubscriptionBasePlanID                  `json:"basePlanId"`
	Requests         []SubscriptionOfferBatchMutationRequest `json:"requests"`
	Action           SubscriptionOfferStateAction            `json:"action"`
	LatencyTolerance ProductUpdateLatencyTolerance           `json:"latencyTolerance"`
	Confirm          bool                                    `json:"confirm"`
	Steps            []string                                `json:"steps"`
}

type SubscriptionOfferBatchStateUpdateResult struct {
	PackageName PackageName                             `json:"packageName"`
	ProductID   SubscriptionProductID                   `json:"productId"`
	BasePlanID  SubscriptionBasePlanID                  `json:"basePlanId"`
	Requests    []SubscriptionOfferBatchMutationRequest `json:"requests"`
	Action      SubscriptionOfferStateAction            `json:"action"`
	DryRun      bool                                    `json:"dryRun"`
	Applied     bool                                    `json:"applied"`
	Offers      []SubscriptionOffer                     `json:"offers,omitempty"`
	Plan        SubscriptionOfferBatchStateUpdatePlan   `json:"plan"`
}

type SubscriptionOfferBatchPatchAvailabilityPlan struct {
	PackageName      PackageName                                 `json:"packageName"`
	ProductID        SubscriptionProductID                       `json:"productId"`
	BasePlanID       SubscriptionBasePlanID                      `json:"basePlanId"`
	Requests         []SubscriptionOfferAvailabilityPatchRequest `json:"requests"`
	UpdateMask       string                                      `json:"updateMask"`
	RegionsVersion   string                                      `json:"regionsVersion"`
	LatencyTolerance ProductUpdateLatencyTolerance               `json:"latencyTolerance"`
	Confirm          bool                                        `json:"confirm"`
	Steps            []string                                    `json:"steps"`
}

type SubscriptionOfferBatchPatchAvailabilityDesiredOffer struct {
	PackageName     PackageName                                                    `json:"packageName"`
	ProductID       SubscriptionProductID                                          `json:"productId"`
	BasePlanID      SubscriptionBasePlanID                                         `json:"basePlanId"`
	OfferID         SubscriptionOfferID                                            `json:"offerId"`
	RegionalConfigs []SubscriptionOfferBatchPatchAvailabilityDesiredRegionalConfig `json:"regionalConfigs"`
}

type SubscriptionOfferBatchPatchAvailabilityDesiredRegionalConfig struct {
	RegionCode                string `json:"regionCode"`
	NewSubscriberAvailability bool   `json:"newSubscriberAvailability"`
}

type SubscriptionOfferBatchPatchAvailabilityResult struct {
	PackageName PackageName                                           `json:"packageName"`
	ProductID   SubscriptionProductID                                 `json:"productId"`
	BasePlanID  SubscriptionBasePlanID                                `json:"basePlanId"`
	Requests    []SubscriptionOfferAvailabilityPatchRequest           `json:"requests"`
	DryRun      bool                                                  `json:"dryRun"`
	Applied     bool                                                  `json:"applied"`
	Offers      []SubscriptionOffer                                   `json:"offers,omitempty"`
	Desired     []SubscriptionOfferBatchPatchAvailabilityDesiredOffer `json:"desiredOffers"`
	Plan        SubscriptionOfferBatchPatchAvailabilityPlan           `json:"plan"`
}

type SubscriptionOfferBatchPatchPhaseRelativeDiscountsPlan struct {
	PackageName      PackageName                                          `json:"packageName"`
	ProductID        SubscriptionProductID                                `json:"productId"`
	BasePlanID       SubscriptionBasePlanID                               `json:"basePlanId"`
	Requests         []SubscriptionOfferPhaseRelativeDiscountPatchRequest `json:"requests"`
	UpdateMask       string                                               `json:"updateMask"`
	RegionsVersion   string                                               `json:"regionsVersion"`
	LatencyTolerance ProductUpdateLatencyTolerance                        `json:"latencyTolerance"`
	Confirm          bool                                                 `json:"confirm"`
	Steps            []string                                             `json:"steps"`
}

type SubscriptionOfferBatchPatchPhaseRelativeDiscountsDesiredOffer struct {
	PackageName PackageName                                                     `json:"packageName"`
	ProductID   SubscriptionProductID                                           `json:"productId"`
	BasePlanID  SubscriptionBasePlanID                                          `json:"basePlanId"`
	OfferID     SubscriptionOfferID                                             `json:"offerId"`
	Phases      []SubscriptionOfferBatchPatchPhaseRelativeDiscountsDesiredPhase `json:"phases"`
}

type SubscriptionOfferBatchPatchPhaseRelativeDiscountsDesiredPhase struct {
	PhaseIndex      int                                                                      `json:"phaseIndex"`
	RegionalConfigs []SubscriptionOfferBatchPatchPhaseRelativeDiscountsDesiredRegionalConfig `json:"regionalConfigs"`
}

type SubscriptionOfferBatchPatchPhaseRelativeDiscountsDesiredRegionalConfig struct {
	RegionCode       string  `json:"regionCode"`
	RelativeDiscount float64 `json:"relativeDiscount"`
}

type SubscriptionOfferBatchPatchPhaseRelativeDiscountsResult struct {
	PackageName PackageName                                                     `json:"packageName"`
	ProductID   SubscriptionProductID                                           `json:"productId"`
	BasePlanID  SubscriptionBasePlanID                                          `json:"basePlanId"`
	Requests    []SubscriptionOfferPhaseRelativeDiscountPatchRequest            `json:"requests"`
	DryRun      bool                                                            `json:"dryRun"`
	Applied     bool                                                            `json:"applied"`
	Offers      []SubscriptionOffer                                             `json:"offers,omitempty"`
	Desired     []SubscriptionOfferBatchPatchPhaseRelativeDiscountsDesiredOffer `json:"desiredOffers"`
	Plan        SubscriptionOfferBatchPatchPhaseRelativeDiscountsPlan           `json:"plan"`
}

type SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsPlan struct {
	PackageName      PackageName                                          `json:"packageName"`
	ProductID        SubscriptionProductID                                `json:"productId"`
	BasePlanID       SubscriptionBasePlanID                               `json:"basePlanId"`
	Requests         []SubscriptionOfferPhaseAbsoluteDiscountPatchRequest `json:"requests"`
	UpdateMask       string                                               `json:"updateMask"`
	RegionsVersion   string                                               `json:"regionsVersion"`
	LatencyTolerance ProductUpdateLatencyTolerance                        `json:"latencyTolerance"`
	Confirm          bool                                                 `json:"confirm"`
	Steps            []string                                             `json:"steps"`
}

type SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsDesiredOffer struct {
	PackageName PackageName                                                     `json:"packageName"`
	ProductID   SubscriptionProductID                                           `json:"productId"`
	BasePlanID  SubscriptionBasePlanID                                          `json:"basePlanId"`
	OfferID     SubscriptionOfferID                                             `json:"offerId"`
	Phases      []SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsDesiredPhase `json:"phases"`
}

type SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsDesiredPhase struct {
	PhaseIndex      int                                                                      `json:"phaseIndex"`
	RegionalConfigs []SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsDesiredRegionalConfig `json:"regionalConfigs"`
}

type SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsDesiredRegionalConfig struct {
	RegionCode       string `json:"regionCode"`
	AbsoluteDiscount Money  `json:"absoluteDiscount"`
}

type SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsResult struct {
	PackageName PackageName                                                     `json:"packageName"`
	ProductID   SubscriptionProductID                                           `json:"productId"`
	BasePlanID  SubscriptionBasePlanID                                          `json:"basePlanId"`
	Requests    []SubscriptionOfferPhaseAbsoluteDiscountPatchRequest            `json:"requests"`
	DryRun      bool                                                            `json:"dryRun"`
	Applied     bool                                                            `json:"applied"`
	Offers      []SubscriptionOffer                                             `json:"offers,omitempty"`
	Desired     []SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsDesiredOffer `json:"desiredOffers"`
	Plan        SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsPlan           `json:"plan"`
}

type SubscriptionOfferBatchPatchPhasePricesPlan struct {
	PackageName      PackageName                               `json:"packageName"`
	ProductID        SubscriptionProductID                     `json:"productId"`
	BasePlanID       SubscriptionBasePlanID                    `json:"basePlanId"`
	Requests         []SubscriptionOfferPhasePricePatchRequest `json:"requests"`
	UpdateMask       string                                    `json:"updateMask"`
	RegionsVersion   string                                    `json:"regionsVersion"`
	LatencyTolerance ProductUpdateLatencyTolerance             `json:"latencyTolerance"`
	Confirm          bool                                      `json:"confirm"`
	Steps            []string                                  `json:"steps"`
}

type SubscriptionOfferBatchPatchPhasePricesDesiredOffer struct {
	PackageName PackageName                                          `json:"packageName"`
	ProductID   SubscriptionProductID                                `json:"productId"`
	BasePlanID  SubscriptionBasePlanID                               `json:"basePlanId"`
	OfferID     SubscriptionOfferID                                  `json:"offerId"`
	Phases      []SubscriptionOfferBatchPatchPhasePricesDesiredPhase `json:"phases"`
}

type SubscriptionOfferBatchPatchPhasePricesDesiredPhase struct {
	PhaseIndex      int                                                           `json:"phaseIndex"`
	RegionalConfigs []SubscriptionOfferBatchPatchPhasePricesDesiredRegionalConfig `json:"regionalConfigs"`
}

type SubscriptionOfferBatchPatchPhasePricesDesiredRegionalConfig struct {
	RegionCode string `json:"regionCode"`
	Price      Money  `json:"price"`
}

type SubscriptionOfferBatchPatchPhasePricesResult struct {
	PackageName PackageName                                          `json:"packageName"`
	ProductID   SubscriptionProductID                                `json:"productId"`
	BasePlanID  SubscriptionBasePlanID                               `json:"basePlanId"`
	Requests    []SubscriptionOfferPhasePricePatchRequest            `json:"requests"`
	DryRun      bool                                                 `json:"dryRun"`
	Applied     bool                                                 `json:"applied"`
	Offers      []SubscriptionOffer                                  `json:"offers,omitempty"`
	Desired     []SubscriptionOfferBatchPatchPhasePricesDesiredOffer `json:"desiredOffers"`
	Plan        SubscriptionOfferBatchPatchPhasePricesPlan           `json:"plan"`
}

type SubscriptionOfferBatchPatchPhaseFreePlan struct {
	PackageName      PackageName                              `json:"packageName"`
	ProductID        SubscriptionProductID                    `json:"productId"`
	BasePlanID       SubscriptionBasePlanID                   `json:"basePlanId"`
	Requests         []SubscriptionOfferPhaseFreePatchRequest `json:"requests"`
	UpdateMask       string                                   `json:"updateMask"`
	RegionsVersion   string                                   `json:"regionsVersion"`
	LatencyTolerance ProductUpdateLatencyTolerance            `json:"latencyTolerance"`
	Confirm          bool                                     `json:"confirm"`
	Steps            []string                                 `json:"steps"`
}

type SubscriptionOfferBatchPatchPhaseFreeDesiredOffer struct {
	PackageName PackageName                                        `json:"packageName"`
	ProductID   SubscriptionProductID                              `json:"productId"`
	BasePlanID  SubscriptionBasePlanID                             `json:"basePlanId"`
	OfferID     SubscriptionOfferID                                `json:"offerId"`
	Phases      []SubscriptionOfferBatchPatchPhaseFreeDesiredPhase `json:"phases"`
}

type SubscriptionOfferBatchPatchPhaseFreeDesiredPhase struct {
	PhaseIndex      int                                                         `json:"phaseIndex"`
	RegionalConfigs []SubscriptionOfferBatchPatchPhaseFreeDesiredRegionalConfig `json:"regionalConfigs"`
}

type SubscriptionOfferBatchPatchPhaseFreeDesiredRegionalConfig struct {
	RegionCode string `json:"regionCode"`
	Free       bool   `json:"free"`
}

type SubscriptionOfferBatchPatchPhaseFreeResult struct {
	PackageName PackageName                                        `json:"packageName"`
	ProductID   SubscriptionProductID                              `json:"productId"`
	BasePlanID  SubscriptionBasePlanID                             `json:"basePlanId"`
	Requests    []SubscriptionOfferPhaseFreePatchRequest           `json:"requests"`
	DryRun      bool                                               `json:"dryRun"`
	Applied     bool                                               `json:"applied"`
	Offers      []SubscriptionOffer                                `json:"offers,omitempty"`
	Desired     []SubscriptionOfferBatchPatchPhaseFreeDesiredOffer `json:"desiredOffers"`
	Plan        SubscriptionOfferBatchPatchPhaseFreePlan           `json:"plan"`
}

type SubscriptionOfferBatchGetter interface {
	BatchGetSubscriptionOffers(ctx context.Context, options SubscriptionOfferBatchGetOptions) (SubscriptionOfferBatchGetResult, error)
}

type SubscriptionOfferCreator interface {
	CreateSubscriptionOffer(ctx context.Context, options SubscriptionOfferCreateOptions) (SubscriptionOffer, error)
}

type SubscriptionOfferBatchStateUpdater interface {
	BatchUpdateSubscriptionOfferStates(ctx context.Context, options SubscriptionOfferBatchStateUpdateOptions) (SubscriptionOfferBatchStateUpdateResult, error)
}

type SubscriptionOfferBatchAvailabilityPatcher interface {
	BatchPatchSubscriptionOfferAvailability(ctx context.Context, options SubscriptionOfferBatchPatchAvailabilityOptions) (SubscriptionOfferBatchPatchAvailabilityResult, error)
}

type SubscriptionOfferBatchPhaseRelativeDiscountPatcher interface {
	BatchPatchSubscriptionOfferPhaseRelativeDiscounts(ctx context.Context, options SubscriptionOfferBatchPatchPhaseRelativeDiscountsOptions) (SubscriptionOfferBatchPatchPhaseRelativeDiscountsResult, error)
}

type SubscriptionOfferBatchPhaseAbsoluteDiscountPatcher interface {
	BatchPatchSubscriptionOfferPhaseAbsoluteDiscounts(ctx context.Context, options SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsOptions) (SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsResult, error)
}

type SubscriptionOfferBatchPhasePricePatcher interface {
	BatchPatchSubscriptionOfferPhasePrices(ctx context.Context, options SubscriptionOfferBatchPatchPhasePricesOptions) (SubscriptionOfferBatchPatchPhasePricesResult, error)
}

type SubscriptionOfferBatchPhaseFreePatcher interface {
	BatchPatchSubscriptionOfferPhaseFree(ctx context.Context, options SubscriptionOfferBatchPatchPhaseFreeOptions) (SubscriptionOfferBatchPatchPhaseFreeResult, error)
}

type SubscriptionOfferDeleter interface {
	DeleteSubscriptionOffer(ctx context.Context, options SubscriptionOfferDeleteOptions) error
}

func BatchGetSubscriptionOffers(ctx context.Context, getter SubscriptionOfferBatchGetter, options SubscriptionOfferBatchGetOptions) (SubscriptionOfferBatchGetResult, error) {
	if err := options.Validate(); err != nil {
		return SubscriptionOfferBatchGetResult{}, err
	}
	if getter == nil {
		return SubscriptionOfferBatchGetResult{}, fmt.Errorf("subscription offer batch getter is required")
	}
	return getter.BatchGetSubscriptionOffers(ctx, options)
}

func CreateSubscriptionOffer(ctx context.Context, creator SubscriptionOfferCreator, options SubscriptionOfferCreateOptions) (SubscriptionOfferCreateResult, error) {
	if err := options.Validate(); err != nil {
		return SubscriptionOfferCreateResult{}, err
	}
	desired := subscriptionOfferCreateDesiredOffer(options)
	result := SubscriptionOfferCreateResult{
		Action:  "create",
		DryRun:  options.DryRun,
		Desired: desired,
		Plan: SubscriptionOfferCreatePlan{
			Action:         "create",
			PackageName:    options.PackageName,
			ProductID:      options.ProductID,
			BasePlanID:     options.BasePlanID,
			OfferID:        options.OfferID,
			RegionsVersion: options.RegionsVersion,
			Confirm:        options.Confirm,
			Steps:          subscriptionOfferCreateSteps(options.DryRun),
		},
	}
	if options.DryRun {
		return result, nil
	}
	if creator == nil {
		return SubscriptionOfferCreateResult{}, fmt.Errorf("subscription offer creator is required")
	}
	offer, err := creator.CreateSubscriptionOffer(ctx, options)
	if err != nil {
		return SubscriptionOfferCreateResult{}, err
	}
	result.Created = true
	result.Offer = &offer
	return result, nil
}

func subscriptionOfferCreateDesiredOffer(options SubscriptionOfferCreateOptions) SubscriptionOffer {
	offer := options.Offer
	offer.PackageName = options.PackageName
	offer.ProductID = options.ProductID
	offer.BasePlanID = options.BasePlanID
	offer.OfferID = options.OfferID
	offer.State = ""
	return offer
}

func BatchUpdateSubscriptionOfferStates(ctx context.Context, updater SubscriptionOfferBatchStateUpdater, options SubscriptionOfferBatchStateUpdateOptions) (SubscriptionOfferBatchStateUpdateResult, error) {
	if err := options.Validate(); err != nil {
		return SubscriptionOfferBatchStateUpdateResult{}, err
	}
	requests := append([]SubscriptionOfferBatchMutationRequest(nil), options.Requests...)
	result := SubscriptionOfferBatchStateUpdateResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		BasePlanID:  options.BasePlanID,
		Requests:    requests,
		Action:      options.Action,
		DryRun:      options.DryRun,
		Plan: SubscriptionOfferBatchStateUpdatePlan{
			PackageName:      options.PackageName,
			ProductID:        options.ProductID,
			BasePlanID:       options.BasePlanID,
			Requests:         requests,
			Action:           options.Action,
			LatencyTolerance: options.LatencyTolerance,
			Confirm:          options.Confirm,
			Steps:            subscriptionOfferBatchStateUpdateSteps(options),
		},
	}
	if options.DryRun {
		return result, nil
	}
	if updater == nil {
		return SubscriptionOfferBatchStateUpdateResult{}, fmt.Errorf("subscription offer batch state updater is required")
	}
	updated, err := updater.BatchUpdateSubscriptionOfferStates(ctx, options)
	if err != nil {
		return SubscriptionOfferBatchStateUpdateResult{}, err
	}
	updated.PackageName = options.PackageName
	updated.ProductID = options.ProductID
	updated.BasePlanID = options.BasePlanID
	updated.Plan = result.Plan
	updated.Requests = requests
	updated.Action = options.Action
	updated.DryRun = false
	updated.Applied = true
	return updated, nil
}

func BatchPatchSubscriptionOfferAvailability(ctx context.Context, patcher SubscriptionOfferBatchAvailabilityPatcher, options SubscriptionOfferBatchPatchAvailabilityOptions) (SubscriptionOfferBatchPatchAvailabilityResult, error) {
	plan, err := NewSubscriptionOfferBatchPatchAvailabilityPlan(options)
	if err != nil {
		return SubscriptionOfferBatchPatchAvailabilityResult{}, err
	}
	requests := append([]SubscriptionOfferAvailabilityPatchRequest(nil), options.Requests...)
	result := SubscriptionOfferBatchPatchAvailabilityResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		BasePlanID:  options.BasePlanID,
		Requests:    requests,
		DryRun:      options.DryRun,
		Desired:     desiredSubscriptionOffersForAvailabilityPatch(options),
		Plan:        plan,
	}
	if options.DryRun {
		return result, nil
	}
	if patcher == nil {
		return SubscriptionOfferBatchPatchAvailabilityResult{}, fmt.Errorf("subscription offer availability batch patcher is required")
	}
	updated, err := patcher.BatchPatchSubscriptionOfferAvailability(ctx, options)
	if err != nil {
		return SubscriptionOfferBatchPatchAvailabilityResult{}, err
	}
	updated.PackageName = options.PackageName
	updated.ProductID = options.ProductID
	updated.BasePlanID = options.BasePlanID
	updated.Requests = requests
	updated.DryRun = false
	updated.Applied = true
	updated.Desired = result.Desired
	updated.Plan = plan
	return updated, nil
}

func BatchPatchSubscriptionOfferPhaseRelativeDiscounts(ctx context.Context, patcher SubscriptionOfferBatchPhaseRelativeDiscountPatcher, options SubscriptionOfferBatchPatchPhaseRelativeDiscountsOptions) (SubscriptionOfferBatchPatchPhaseRelativeDiscountsResult, error) {
	plan, err := NewSubscriptionOfferBatchPatchPhaseRelativeDiscountsPlan(options)
	if err != nil {
		return SubscriptionOfferBatchPatchPhaseRelativeDiscountsResult{}, err
	}
	requests := append([]SubscriptionOfferPhaseRelativeDiscountPatchRequest(nil), options.Requests...)
	result := SubscriptionOfferBatchPatchPhaseRelativeDiscountsResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		BasePlanID:  options.BasePlanID,
		Requests:    requests,
		DryRun:      options.DryRun,
		Desired:     desiredSubscriptionOffersForPhaseRelativeDiscountPatch(options),
		Plan:        plan,
	}
	if options.DryRun {
		return result, nil
	}
	if patcher == nil {
		return SubscriptionOfferBatchPatchPhaseRelativeDiscountsResult{}, fmt.Errorf("subscription offer phase relative discount batch patcher is required")
	}
	updated, err := patcher.BatchPatchSubscriptionOfferPhaseRelativeDiscounts(ctx, options)
	if err != nil {
		return SubscriptionOfferBatchPatchPhaseRelativeDiscountsResult{}, err
	}
	updated.PackageName = options.PackageName
	updated.ProductID = options.ProductID
	updated.BasePlanID = options.BasePlanID
	updated.Requests = requests
	updated.DryRun = false
	updated.Applied = true
	updated.Desired = result.Desired
	updated.Plan = plan
	return updated, nil
}

func BatchPatchSubscriptionOfferPhaseAbsoluteDiscounts(ctx context.Context, patcher SubscriptionOfferBatchPhaseAbsoluteDiscountPatcher, options SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsOptions) (SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsResult, error) {
	plan, err := NewSubscriptionOfferBatchPatchPhaseAbsoluteDiscountsPlan(options)
	if err != nil {
		return SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsResult{}, err
	}
	requests := append([]SubscriptionOfferPhaseAbsoluteDiscountPatchRequest(nil), options.Requests...)
	result := SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		BasePlanID:  options.BasePlanID,
		Requests:    requests,
		DryRun:      options.DryRun,
		Desired:     desiredSubscriptionOffersForPhaseAbsoluteDiscountPatch(options),
		Plan:        plan,
	}
	if options.DryRun {
		return result, nil
	}
	if patcher == nil {
		return SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsResult{}, fmt.Errorf("subscription offer phase absolute discount batch patcher is required")
	}
	updated, err := patcher.BatchPatchSubscriptionOfferPhaseAbsoluteDiscounts(ctx, options)
	if err != nil {
		return SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsResult{}, err
	}
	updated.PackageName = options.PackageName
	updated.ProductID = options.ProductID
	updated.BasePlanID = options.BasePlanID
	updated.Requests = requests
	updated.DryRun = false
	updated.Applied = true
	updated.Desired = result.Desired
	updated.Plan = plan
	return updated, nil
}

func BatchPatchSubscriptionOfferPhasePrices(ctx context.Context, patcher SubscriptionOfferBatchPhasePricePatcher, options SubscriptionOfferBatchPatchPhasePricesOptions) (SubscriptionOfferBatchPatchPhasePricesResult, error) {
	plan, err := NewSubscriptionOfferBatchPatchPhasePricesPlan(options)
	if err != nil {
		return SubscriptionOfferBatchPatchPhasePricesResult{}, err
	}
	requests := append([]SubscriptionOfferPhasePricePatchRequest(nil), options.Requests...)
	result := SubscriptionOfferBatchPatchPhasePricesResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		BasePlanID:  options.BasePlanID,
		Requests:    requests,
		DryRun:      options.DryRun,
		Desired:     desiredSubscriptionOffersForPhasePricePatch(options),
		Plan:        plan,
	}
	if options.DryRun {
		return result, nil
	}
	if patcher == nil {
		return SubscriptionOfferBatchPatchPhasePricesResult{}, fmt.Errorf("subscription offer phase price batch patcher is required")
	}
	updated, err := patcher.BatchPatchSubscriptionOfferPhasePrices(ctx, options)
	if err != nil {
		return SubscriptionOfferBatchPatchPhasePricesResult{}, err
	}
	updated.PackageName = options.PackageName
	updated.ProductID = options.ProductID
	updated.BasePlanID = options.BasePlanID
	updated.Requests = requests
	updated.DryRun = false
	updated.Applied = true
	updated.Desired = result.Desired
	updated.Plan = plan
	return updated, nil
}

func BatchPatchSubscriptionOfferPhaseFree(ctx context.Context, patcher SubscriptionOfferBatchPhaseFreePatcher, options SubscriptionOfferBatchPatchPhaseFreeOptions) (SubscriptionOfferBatchPatchPhaseFreeResult, error) {
	plan, err := NewSubscriptionOfferBatchPatchPhaseFreePlan(options)
	if err != nil {
		return SubscriptionOfferBatchPatchPhaseFreeResult{}, err
	}
	requests := append([]SubscriptionOfferPhaseFreePatchRequest(nil), options.Requests...)
	result := SubscriptionOfferBatchPatchPhaseFreeResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		BasePlanID:  options.BasePlanID,
		Requests:    requests,
		DryRun:      options.DryRun,
		Desired:     desiredSubscriptionOffersForPhaseFreePatch(options),
		Plan:        plan,
	}
	if options.DryRun {
		return result, nil
	}
	if patcher == nil {
		return SubscriptionOfferBatchPatchPhaseFreeResult{}, fmt.Errorf("subscription offer phase free batch patcher is required")
	}
	updated, err := patcher.BatchPatchSubscriptionOfferPhaseFree(ctx, options)
	if err != nil {
		return SubscriptionOfferBatchPatchPhaseFreeResult{}, err
	}
	updated.PackageName = options.PackageName
	updated.ProductID = options.ProductID
	updated.BasePlanID = options.BasePlanID
	updated.Requests = requests
	updated.DryRun = false
	updated.Applied = true
	updated.Desired = result.Desired
	updated.Plan = plan
	return updated, nil
}

func NewSubscriptionOfferBatchPatchAvailabilityPlan(options SubscriptionOfferBatchPatchAvailabilityOptions) (SubscriptionOfferBatchPatchAvailabilityPlan, error) {
	if err := options.Validate(); err != nil {
		return SubscriptionOfferBatchPatchAvailabilityPlan{}, err
	}
	return SubscriptionOfferBatchPatchAvailabilityPlan{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		BasePlanID:       options.BasePlanID,
		Requests:         append([]SubscriptionOfferAvailabilityPatchRequest(nil), options.Requests...),
		UpdateMask:       subscriptionOfferAvailabilityUpdateMask,
		RegionsVersion:   options.RegionsVersion,
		LatencyTolerance: options.LatencyTolerance,
		Confirm:          options.Confirm,
		Steps:            subscriptionOfferBatchPatchAvailabilitySteps(options),
	}, nil
}

func NewSubscriptionOfferBatchPatchPhaseRelativeDiscountsPlan(options SubscriptionOfferBatchPatchPhaseRelativeDiscountsOptions) (SubscriptionOfferBatchPatchPhaseRelativeDiscountsPlan, error) {
	if err := options.Validate(); err != nil {
		return SubscriptionOfferBatchPatchPhaseRelativeDiscountsPlan{}, err
	}
	return SubscriptionOfferBatchPatchPhaseRelativeDiscountsPlan{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		BasePlanID:       options.BasePlanID,
		Requests:         append([]SubscriptionOfferPhaseRelativeDiscountPatchRequest(nil), options.Requests...),
		UpdateMask:       subscriptionOfferPhasesUpdateMask,
		RegionsVersion:   options.RegionsVersion,
		LatencyTolerance: options.LatencyTolerance,
		Confirm:          options.Confirm,
		Steps:            subscriptionOfferBatchPatchPhaseRelativeDiscountsSteps(options),
	}, nil
}

func NewSubscriptionOfferBatchPatchPhaseAbsoluteDiscountsPlan(options SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsOptions) (SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsPlan, error) {
	if err := options.Validate(); err != nil {
		return SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsPlan{}, err
	}
	return SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsPlan{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		BasePlanID:       options.BasePlanID,
		Requests:         append([]SubscriptionOfferPhaseAbsoluteDiscountPatchRequest(nil), options.Requests...),
		UpdateMask:       subscriptionOfferPhasesUpdateMask,
		RegionsVersion:   options.RegionsVersion,
		LatencyTolerance: options.LatencyTolerance,
		Confirm:          options.Confirm,
		Steps:            subscriptionOfferBatchPatchPhaseAbsoluteDiscountsSteps(options),
	}, nil
}

func NewSubscriptionOfferBatchPatchPhasePricesPlan(options SubscriptionOfferBatchPatchPhasePricesOptions) (SubscriptionOfferBatchPatchPhasePricesPlan, error) {
	if err := options.Validate(); err != nil {
		return SubscriptionOfferBatchPatchPhasePricesPlan{}, err
	}
	return SubscriptionOfferBatchPatchPhasePricesPlan{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		BasePlanID:       options.BasePlanID,
		Requests:         append([]SubscriptionOfferPhasePricePatchRequest(nil), options.Requests...),
		UpdateMask:       subscriptionOfferPhasesUpdateMask,
		RegionsVersion:   options.RegionsVersion,
		LatencyTolerance: options.LatencyTolerance,
		Confirm:          options.Confirm,
		Steps:            subscriptionOfferBatchPatchPhasePricesSteps(options),
	}, nil
}

func NewSubscriptionOfferBatchPatchPhaseFreePlan(options SubscriptionOfferBatchPatchPhaseFreeOptions) (SubscriptionOfferBatchPatchPhaseFreePlan, error) {
	if err := options.Validate(); err != nil {
		return SubscriptionOfferBatchPatchPhaseFreePlan{}, err
	}
	return SubscriptionOfferBatchPatchPhaseFreePlan{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		BasePlanID:       options.BasePlanID,
		Requests:         append([]SubscriptionOfferPhaseFreePatchRequest(nil), options.Requests...),
		UpdateMask:       subscriptionOfferPhasesUpdateMask,
		RegionsVersion:   options.RegionsVersion,
		LatencyTolerance: options.LatencyTolerance,
		Confirm:          options.Confirm,
		Steps:            subscriptionOfferBatchPatchPhaseFreeSteps(options),
	}, nil
}

func desiredSubscriptionOffersForAvailabilityPatch(options SubscriptionOfferBatchPatchAvailabilityOptions) []SubscriptionOfferBatchPatchAvailabilityDesiredOffer {
	byOffer := map[string]int{}
	offers := make([]SubscriptionOfferBatchPatchAvailabilityDesiredOffer, 0)
	for _, request := range options.Requests {
		key := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID)
		index, ok := byOffer[key]
		if !ok {
			byOffer[key] = len(offers)
			offers = append(offers, SubscriptionOfferBatchPatchAvailabilityDesiredOffer{
				PackageName:     options.PackageName,
				ProductID:       request.ProductID,
				BasePlanID:      request.BasePlanID,
				OfferID:         request.OfferID,
				RegionalConfigs: []SubscriptionOfferBatchPatchAvailabilityDesiredRegionalConfig{},
			})
			index = len(offers) - 1
		}
		offers[index].RegionalConfigs = append(offers[index].RegionalConfigs, SubscriptionOfferBatchPatchAvailabilityDesiredRegionalConfig{
			RegionCode:                request.RegionCode,
			NewSubscriberAvailability: request.Availability,
		})
	}
	return offers
}

func desiredSubscriptionOffersForPhaseRelativeDiscountPatch(options SubscriptionOfferBatchPatchPhaseRelativeDiscountsOptions) []SubscriptionOfferBatchPatchPhaseRelativeDiscountsDesiredOffer {
	byOffer := map[string]int{}
	phaseIndexesByOffer := map[string]map[int]int{}
	offers := make([]SubscriptionOfferBatchPatchPhaseRelativeDiscountsDesiredOffer, 0)
	for _, request := range options.Requests {
		offerKey := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID)
		offerIndex, ok := byOffer[offerKey]
		if !ok {
			byOffer[offerKey] = len(offers)
			phaseIndexesByOffer[offerKey] = map[int]int{}
			offers = append(offers, SubscriptionOfferBatchPatchPhaseRelativeDiscountsDesiredOffer{
				PackageName: options.PackageName,
				ProductID:   request.ProductID,
				BasePlanID:  request.BasePlanID,
				OfferID:     request.OfferID,
				Phases:      []SubscriptionOfferBatchPatchPhaseRelativeDiscountsDesiredPhase{},
			})
			offerIndex = len(offers) - 1
		}
		phaseIndex, ok := phaseIndexesByOffer[offerKey][request.PhaseIndex]
		if !ok {
			phaseIndexesByOffer[offerKey][request.PhaseIndex] = len(offers[offerIndex].Phases)
			offers[offerIndex].Phases = append(offers[offerIndex].Phases, SubscriptionOfferBatchPatchPhaseRelativeDiscountsDesiredPhase{
				PhaseIndex:      request.PhaseIndex,
				RegionalConfigs: []SubscriptionOfferBatchPatchPhaseRelativeDiscountsDesiredRegionalConfig{},
			})
			phaseIndex = len(offers[offerIndex].Phases) - 1
		}
		offers[offerIndex].Phases[phaseIndex].RegionalConfigs = append(offers[offerIndex].Phases[phaseIndex].RegionalConfigs, SubscriptionOfferBatchPatchPhaseRelativeDiscountsDesiredRegionalConfig{
			RegionCode:       request.RegionCode,
			RelativeDiscount: request.RelativeDiscount,
		})
	}
	return offers
}

func desiredSubscriptionOffersForPhaseAbsoluteDiscountPatch(options SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsOptions) []SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsDesiredOffer {
	byOffer := map[string]int{}
	phaseIndexesByOffer := map[string]map[int]int{}
	offers := make([]SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsDesiredOffer, 0)
	for _, request := range options.Requests {
		offerKey := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID)
		offerIndex, ok := byOffer[offerKey]
		if !ok {
			byOffer[offerKey] = len(offers)
			phaseIndexesByOffer[offerKey] = map[int]int{}
			offers = append(offers, SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsDesiredOffer{
				PackageName: options.PackageName,
				ProductID:   request.ProductID,
				BasePlanID:  request.BasePlanID,
				OfferID:     request.OfferID,
				Phases:      []SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsDesiredPhase{},
			})
			offerIndex = len(offers) - 1
		}
		phaseIndex, ok := phaseIndexesByOffer[offerKey][request.PhaseIndex]
		if !ok {
			phaseIndexesByOffer[offerKey][request.PhaseIndex] = len(offers[offerIndex].Phases)
			offers[offerIndex].Phases = append(offers[offerIndex].Phases, SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsDesiredPhase{
				PhaseIndex:      request.PhaseIndex,
				RegionalConfigs: []SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsDesiredRegionalConfig{},
			})
			phaseIndex = len(offers[offerIndex].Phases) - 1
		}
		offers[offerIndex].Phases[phaseIndex].RegionalConfigs = append(offers[offerIndex].Phases[phaseIndex].RegionalConfigs, SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsDesiredRegionalConfig{
			RegionCode:       request.RegionCode,
			AbsoluteDiscount: request.AbsoluteDiscount,
		})
	}
	return offers
}

func desiredSubscriptionOffersForPhasePricePatch(options SubscriptionOfferBatchPatchPhasePricesOptions) []SubscriptionOfferBatchPatchPhasePricesDesiredOffer {
	byOffer := map[string]int{}
	phaseIndexesByOffer := map[string]map[int]int{}
	offers := make([]SubscriptionOfferBatchPatchPhasePricesDesiredOffer, 0)
	for _, request := range options.Requests {
		offerKey := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID)
		offerIndex, ok := byOffer[offerKey]
		if !ok {
			byOffer[offerKey] = len(offers)
			phaseIndexesByOffer[offerKey] = map[int]int{}
			offers = append(offers, SubscriptionOfferBatchPatchPhasePricesDesiredOffer{
				PackageName: options.PackageName,
				ProductID:   request.ProductID,
				BasePlanID:  request.BasePlanID,
				OfferID:     request.OfferID,
				Phases:      []SubscriptionOfferBatchPatchPhasePricesDesiredPhase{},
			})
			offerIndex = len(offers) - 1
		}
		phaseIndex, ok := phaseIndexesByOffer[offerKey][request.PhaseIndex]
		if !ok {
			phaseIndexesByOffer[offerKey][request.PhaseIndex] = len(offers[offerIndex].Phases)
			offers[offerIndex].Phases = append(offers[offerIndex].Phases, SubscriptionOfferBatchPatchPhasePricesDesiredPhase{
				PhaseIndex:      request.PhaseIndex,
				RegionalConfigs: []SubscriptionOfferBatchPatchPhasePricesDesiredRegionalConfig{},
			})
			phaseIndex = len(offers[offerIndex].Phases) - 1
		}
		offers[offerIndex].Phases[phaseIndex].RegionalConfigs = append(offers[offerIndex].Phases[phaseIndex].RegionalConfigs, SubscriptionOfferBatchPatchPhasePricesDesiredRegionalConfig{
			RegionCode: request.RegionCode,
			Price:      request.Price,
		})
	}
	return offers
}

func desiredSubscriptionOffersForPhaseFreePatch(options SubscriptionOfferBatchPatchPhaseFreeOptions) []SubscriptionOfferBatchPatchPhaseFreeDesiredOffer {
	byOffer := map[string]int{}
	phaseIndexesByOffer := map[string]map[int]int{}
	offers := make([]SubscriptionOfferBatchPatchPhaseFreeDesiredOffer, 0)
	for _, request := range options.Requests {
		offerKey := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID)
		offerIndex, ok := byOffer[offerKey]
		if !ok {
			byOffer[offerKey] = len(offers)
			phaseIndexesByOffer[offerKey] = map[int]int{}
			offers = append(offers, SubscriptionOfferBatchPatchPhaseFreeDesiredOffer{
				PackageName: options.PackageName,
				ProductID:   request.ProductID,
				BasePlanID:  request.BasePlanID,
				OfferID:     request.OfferID,
				Phases:      []SubscriptionOfferBatchPatchPhaseFreeDesiredPhase{},
			})
			offerIndex = len(offers) - 1
		}
		phaseIndex, ok := phaseIndexesByOffer[offerKey][request.PhaseIndex]
		if !ok {
			phaseIndexesByOffer[offerKey][request.PhaseIndex] = len(offers[offerIndex].Phases)
			offers[offerIndex].Phases = append(offers[offerIndex].Phases, SubscriptionOfferBatchPatchPhaseFreeDesiredPhase{
				PhaseIndex:      request.PhaseIndex,
				RegionalConfigs: []SubscriptionOfferBatchPatchPhaseFreeDesiredRegionalConfig{},
			})
			phaseIndex = len(offers[offerIndex].Phases) - 1
		}
		offers[offerIndex].Phases[phaseIndex].RegionalConfigs = append(offers[offerIndex].Phases[phaseIndex].RegionalConfigs, SubscriptionOfferBatchPatchPhaseFreeDesiredRegionalConfig{
			RegionCode: request.RegionCode,
			Free:       true,
		})
	}
	return offers
}

type SubscriptionOfferDeleteOptions struct {
	PackageName PackageName            `json:"packageName"`
	ProductID   SubscriptionProductID  `json:"productId"`
	BasePlanID  SubscriptionBasePlanID `json:"basePlanId"`
	OfferID     SubscriptionOfferID    `json:"offerId"`
	Confirm     bool                   `json:"confirm"`
	DryRun      bool                   `json:"dryRun"`
}

func (o SubscriptionOfferDeleteOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewSubscriptionProductID(o.ProductID.String()); err != nil {
		return err
	}
	if _, err := NewSubscriptionBasePlanID(o.BasePlanID.String()); err != nil {
		return err
	}
	if _, err := NewSubscriptionOfferID(o.OfferID.String()); err != nil {
		return err
	}
	if o.Confirm && o.DryRun {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.Confirm && !o.DryRun {
		return fmt.Errorf("subscription offer deletion requires --confirm or --dry-run")
	}
	return nil
}

func (o SubscriptionOfferDeleteOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live subscription offer deletion cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live subscription offer deletion requires --confirm")
	}
	return nil
}

type SubscriptionOfferDeletePlan struct {
	PackageName PackageName            `json:"packageName"`
	ProductID   SubscriptionProductID  `json:"productId"`
	BasePlanID  SubscriptionBasePlanID `json:"basePlanId"`
	OfferID     SubscriptionOfferID    `json:"offerId"`
	Confirm     bool                   `json:"confirm"`
	Steps       []string               `json:"steps"`
}

type SubscriptionOfferDeleteResult struct {
	PackageName PackageName                 `json:"packageName"`
	ProductID   SubscriptionProductID       `json:"productId"`
	BasePlanID  SubscriptionBasePlanID      `json:"basePlanId"`
	OfferID     SubscriptionOfferID         `json:"offerId"`
	DryRun      bool                        `json:"dryRun"`
	Deleted     bool                        `json:"deleted"`
	Plan        SubscriptionOfferDeletePlan `json:"plan"`
}

func DeleteSubscriptionOffer(ctx context.Context, deleter SubscriptionOfferDeleter, options SubscriptionOfferDeleteOptions) (SubscriptionOfferDeleteResult, error) {
	if err := options.Validate(); err != nil {
		return SubscriptionOfferDeleteResult{}, err
	}
	result := SubscriptionOfferDeleteResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		BasePlanID:  options.BasePlanID,
		OfferID:     options.OfferID,
		DryRun:      options.DryRun,
		Plan: SubscriptionOfferDeletePlan{
			PackageName: options.PackageName,
			ProductID:   options.ProductID,
			BasePlanID:  options.BasePlanID,
			OfferID:     options.OfferID,
			Confirm:     options.Confirm,
			Steps:       []string{"delete subscription offer"},
		},
	}
	if options.DryRun {
		return result, nil
	}
	if deleter == nil {
		return SubscriptionOfferDeleteResult{}, fmt.Errorf("subscription offer deleter is required")
	}
	if err := deleter.DeleteSubscriptionOffer(ctx, options); err != nil {
		return SubscriptionOfferDeleteResult{}, err
	}
	result.Deleted = true
	return result, nil
}

type SubscriptionOfferStateUpdateOptions struct {
	PackageName      PackageName                   `json:"packageName"`
	ProductID        SubscriptionProductID         `json:"productId"`
	BasePlanID       SubscriptionBasePlanID        `json:"basePlanId"`
	OfferID          SubscriptionOfferID           `json:"offerId"`
	Action           SubscriptionOfferStateAction  `json:"action"`
	LatencyTolerance ProductUpdateLatencyTolerance `json:"latencyTolerance"`
	Confirm          bool                          `json:"confirm"`
	DryRun           bool                          `json:"dryRun"`
}

func (o SubscriptionOfferStateUpdateOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewSubscriptionProductID(o.ProductID.String()); err != nil {
		return err
	}
	if _, err := NewSubscriptionBasePlanID(o.BasePlanID.String()); err != nil {
		return err
	}
	if _, err := NewSubscriptionOfferID(o.OfferID.String()); err != nil {
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
		return fmt.Errorf("subscription offer state update requires --confirm or --dry-run")
	}
	return nil
}

func (o SubscriptionOfferStateUpdateOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live subscription offer state update cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live subscription offer state update requires --confirm")
	}
	return nil
}

type SubscriptionOfferStateUpdatePlan struct {
	PackageName      PackageName                   `json:"packageName"`
	ProductID        SubscriptionProductID         `json:"productId"`
	BasePlanID       SubscriptionBasePlanID        `json:"basePlanId"`
	OfferID          SubscriptionOfferID           `json:"offerId"`
	Action           SubscriptionOfferStateAction  `json:"action"`
	LatencyTolerance ProductUpdateLatencyTolerance `json:"latencyTolerance"`
	Confirm          bool                          `json:"confirm"`
	Steps            []string                      `json:"steps"`
}

type SubscriptionOfferStateUpdateResult struct {
	PackageName PackageName                      `json:"packageName"`
	ProductID   SubscriptionProductID            `json:"productId"`
	BasePlanID  SubscriptionBasePlanID           `json:"basePlanId"`
	OfferID     SubscriptionOfferID              `json:"offerId"`
	Action      SubscriptionOfferStateAction     `json:"action"`
	DryRun      bool                             `json:"dryRun"`
	Applied     bool                             `json:"applied"`
	Offer       *SubscriptionOffer               `json:"offer,omitempty"`
	Plan        SubscriptionOfferStateUpdatePlan `json:"plan"`
}

type SubscriptionOfferStateUpdater interface {
	UpdateSubscriptionOfferState(ctx context.Context, options SubscriptionOfferStateUpdateOptions) (SubscriptionOffer, error)
}

func NewSubscriptionOfferStateUpdatePlan(options SubscriptionOfferStateUpdateOptions) (SubscriptionOfferStateUpdatePlan, error) {
	if err := options.Validate(); err != nil {
		return SubscriptionOfferStateUpdatePlan{}, err
	}
	return SubscriptionOfferStateUpdatePlan{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		BasePlanID:       options.BasePlanID,
		OfferID:          options.OfferID,
		Action:           options.Action,
		LatencyTolerance: options.LatencyTolerance,
		Confirm:          options.Confirm,
		Steps:            subscriptionOfferStateUpdateSteps(options),
	}, nil
}

func UpdateSubscriptionOfferState(ctx context.Context, updater SubscriptionOfferStateUpdater, options SubscriptionOfferStateUpdateOptions) (SubscriptionOfferStateUpdateResult, error) {
	plan, err := NewSubscriptionOfferStateUpdatePlan(options)
	if err != nil {
		return SubscriptionOfferStateUpdateResult{}, err
	}
	result := SubscriptionOfferStateUpdateResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		BasePlanID:  options.BasePlanID,
		OfferID:     options.OfferID,
		Action:      options.Action,
		DryRun:      options.DryRun,
		Applied:     false,
		Plan:        plan,
	}
	if options.DryRun {
		return result, nil
	}
	if updater == nil {
		return SubscriptionOfferStateUpdateResult{}, fmt.Errorf("subscription offer state updater is required")
	}
	offer, err := updater.UpdateSubscriptionOfferState(ctx, options)
	if err != nil {
		return SubscriptionOfferStateUpdateResult{}, err
	}
	result.Applied = true
	result.Offer = &offer
	return result, nil
}

func subscriptionOfferStateUpdateSteps(options SubscriptionOfferStateUpdateOptions) []string {
	if options.DryRun {
		return []string{fmt.Sprintf("plan %s subscription offer", options.Action)}
	}
	return []string{fmt.Sprintf("%s subscription offer", options.Action)}
}

func subscriptionOfferBatchStateUpdateSteps(options SubscriptionOfferBatchStateUpdateOptions) []string {
	if options.DryRun {
		return []string{fmt.Sprintf("plan batch %s subscription offers", options.Action)}
	}
	return []string{fmt.Sprintf("batch %s subscription offers", options.Action)}
}

func subscriptionOfferCreateSteps(dryRun bool) []string {
	if dryRun {
		return []string{"plan subscription offer create"}
	}
	return []string{"create subscription offer"}
}

const subscriptionOfferAvailabilityUpdateMask = "regionalConfigs"
const subscriptionOfferPhasesUpdateMask = "phases"

func subscriptionOfferBatchPatchAvailabilitySteps(options SubscriptionOfferBatchPatchAvailabilityOptions) []string {
	if options.DryRun {
		return []string{"plan subscription offer availability batch patch"}
	}
	return []string{"fetch current subscription offers", "merge regional availability", "batch patch subscription offer availability"}
}

func subscriptionOfferBatchPatchPhaseRelativeDiscountsSteps(options SubscriptionOfferBatchPatchPhaseRelativeDiscountsOptions) []string {
	if options.DryRun {
		return []string{"plan subscription offer phase relative discount batch patch"}
	}
	return []string{"fetch current subscription offers", "merge phase regional relative discounts", "batch patch subscription offer phase relative discounts"}
}

func subscriptionOfferBatchPatchPhaseAbsoluteDiscountsSteps(options SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsOptions) []string {
	if options.DryRun {
		return []string{"plan subscription offer phase absolute discount batch patch"}
	}
	return []string{"fetch current subscription offers", "merge phase regional absolute discounts", "batch patch subscription offer phase absolute discounts"}
}

func subscriptionOfferBatchPatchPhasePricesSteps(options SubscriptionOfferBatchPatchPhasePricesOptions) []string {
	if options.DryRun {
		return []string{"plan subscription offer phase price batch patch"}
	}
	return []string{"fetch current subscription offers", "merge phase regional prices", "batch patch subscription offer phase prices"}
}

func subscriptionOfferBatchPatchPhaseFreeSteps(options SubscriptionOfferBatchPatchPhaseFreeOptions) []string {
	if options.DryRun {
		return []string{"plan subscription offer phase free batch patch"}
	}
	return []string{"fetch current subscription offers", "merge phase regional free overrides", "batch patch subscription offer phase free overrides"}
}
