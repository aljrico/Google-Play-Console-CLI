package play

import (
	"context"
	"fmt"
	"strings"
)

type SubscriptionBasePlanID string

const SubscriptionOfferWildcardID = "-"

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

type SubscriptionOfferBatchGetRequest struct {
	ProductID  SubscriptionProductID  `json:"productId"`
	BasePlanID SubscriptionBasePlanID `json:"basePlanId"`
	OfferID    SubscriptionOfferID    `json:"offerId"`
}

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

type SubscriptionOfferBatchGetOptions struct {
	PackageName PackageName                        `json:"packageName"`
	ProductID   SubscriptionProductID              `json:"productId"`
	BasePlanID  SubscriptionBasePlanID             `json:"basePlanId"`
	Requests    []SubscriptionOfferBatchGetRequest `json:"requests"`
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

type SubscriptionOfferBatchGetResult struct {
	PackageName PackageName                      `json:"packageName"`
	ProductID   SubscriptionProductID            `json:"productId"`
	BasePlanID  SubscriptionBasePlanID           `json:"basePlanId"`
	Offers      []SubscriptionOffer              `json:"offers"`
	Options     SubscriptionOfferBatchGetOptions `json:"options"`
}

type SubscriptionOfferBatchGetter interface {
	BatchGetSubscriptionOffers(ctx context.Context, options SubscriptionOfferBatchGetOptions) (SubscriptionOfferBatchGetResult, error)
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
