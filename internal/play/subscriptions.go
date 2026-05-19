package play

import (
	"context"
	"fmt"
)

type SubscriptionProductID string

func NewSubscriptionProductID(value string) (SubscriptionProductID, error) {
	if value == "" {
		return "", fmt.Errorf("subscription product ID is required")
	}
	if len(value) > 40 {
		return "", fmt.Errorf("subscription product ID cannot exceed 40 characters")
	}
	if !isValidSubscriptionProductID(value) {
		return "", fmt.Errorf("invalid subscription product ID %q", value)
	}
	return SubscriptionProductID(value), nil
}

func (s SubscriptionProductID) String() string {
	return string(s)
}

func isValidSubscriptionProductID(value string) bool {
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

func isASCIILower(character byte) bool {
	return character >= 'a' && character <= 'z'
}

type SubscriptionState string

const (
	SubscriptionStateDraft       SubscriptionState = "DRAFT"
	SubscriptionStateActive      SubscriptionState = "ACTIVE"
	SubscriptionStateInactive    SubscriptionState = "INACTIVE"
	SubscriptionStateUnspecified SubscriptionState = "STATE_UNSPECIFIED"
)

type SubscriptionBasePlanType string

const (
	SubscriptionBasePlanTypeAutoRenewing SubscriptionBasePlanType = "autoRenewing"
	SubscriptionBasePlanTypePrepaid      SubscriptionBasePlanType = "prepaid"
	SubscriptionBasePlanTypeInstallments SubscriptionBasePlanType = "installments"
)

type SubscriptionListing struct {
	LanguageCode string   `json:"languageCode"`
	Title        string   `json:"title,omitempty"`
	Description  string   `json:"description,omitempty"`
	Benefits     []string `json:"benefits,omitempty"`
}

type SubscriptionBasePlan struct {
	BasePlanID                          string                          `json:"basePlanId"`
	State                               SubscriptionState               `json:"state,omitempty"`
	Type                                SubscriptionBasePlanType        `json:"type,omitempty"`
	BillingPeriodDuration               string                          `json:"billingPeriodDuration,omitempty"`
	GracePeriodDuration                 string                          `json:"gracePeriodDuration,omitempty"`
	AccountHoldDuration                 string                          `json:"accountHoldDuration,omitempty"`
	LegacyCompatible                    bool                            `json:"legacyCompatible,omitempty"`
	LegacyCompatibleSubscriptionOfferID string                          `json:"legacyCompatibleSubscriptionOfferId,omitempty"`
	ProrationMode                       string                          `json:"prorationMode,omitempty"`
	ResubscribeState                    string                          `json:"resubscribeState,omitempty"`
	TimeExtension                       string                          `json:"timeExtension,omitempty"`
	CommittedPaymentsCount              int64                           `json:"committedPaymentsCount,omitempty"`
	RenewalType                         string                          `json:"renewalType,omitempty"`
	OfferTags                           []string                        `json:"offerTags,omitempty"`
	RegionalConfigs                     []SubscriptionRegionalConfig    `json:"regionalConfigs,omitempty"`
	OtherRegionsConfig                  *SubscriptionOtherRegionsConfig `json:"otherRegionsConfig,omitempty"`
}

type BasePlanStateAction string

const (
	BasePlanStateActionActivate   BasePlanStateAction = "activate"
	BasePlanStateActionDeactivate BasePlanStateAction = "deactivate"
)

func (a BasePlanStateAction) String() string {
	return string(a)
}

func (a BasePlanStateAction) Validate() error {
	switch a {
	case BasePlanStateActionActivate, BasePlanStateActionDeactivate:
		return nil
	default:
		return fmt.Errorf("unsupported base plan state action %q", a)
	}
}

type SubscriptionRegionalConfig struct {
	RegionCode                string `json:"regionCode"`
	NewSubscriberAvailability bool   `json:"newSubscriberAvailability,omitempty"`
	Price                     *Money `json:"price,omitempty"`
}

type SubscriptionOtherRegionsConfig struct {
	NewSubscriberAvailability bool   `json:"newSubscriberAvailability,omitempty"`
	USDPrice                  *Money `json:"usdPrice,omitempty"`
	EURPrice                  *Money `json:"eurPrice,omitempty"`
}

type Money struct {
	CurrencyCode string `json:"currencyCode,omitempty"`
	Units        int64  `json:"units,omitempty"`
	Nanos        int64  `json:"nanos,omitempty"`
}

type Subscription struct {
	PackageName              PackageName                   `json:"packageName"`
	ProductID                SubscriptionProductID         `json:"productId"`
	Archived                 bool                          `json:"archived,omitempty"`
	Listings                 []SubscriptionListing         `json:"listings"`
	BasePlans                []SubscriptionBasePlan        `json:"basePlans"`
	RestrictedCountries      []string                      `json:"restrictedCountries,omitempty"`
	TaxAndComplianceSettings *ProductTaxComplianceSettings `json:"taxAndComplianceSettings,omitempty"`
}

type SubscriptionListOptions struct {
	PackageName  PackageName `json:"packageName"`
	PageSize     int64       `json:"pageSize,omitempty"`
	PageToken    string      `json:"pageToken,omitempty"`
	ShowArchived bool        `json:"showArchived,omitempty"`
}

func (o SubscriptionListOptions) Validate() error {
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

type SubscriptionListResult struct {
	PackageName   PackageName             `json:"packageName"`
	Subscriptions []Subscription          `json:"subscriptions"`
	NextPageToken string                  `json:"nextPageToken,omitempty"`
	Options       SubscriptionListOptions `json:"options"`
}

type SubscriptionLister interface {
	ListSubscriptions(ctx context.Context, options SubscriptionListOptions) (SubscriptionListResult, error)
}

func ListSubscriptions(ctx context.Context, lister SubscriptionLister, options SubscriptionListOptions) (SubscriptionListResult, error) {
	if err := options.Validate(); err != nil {
		return SubscriptionListResult{}, err
	}
	if lister == nil {
		return SubscriptionListResult{}, fmt.Errorf("subscription lister is required")
	}
	return lister.ListSubscriptions(ctx, options)
}

type SubscriptionGetOptions struct {
	PackageName PackageName           `json:"packageName"`
	ProductID   SubscriptionProductID `json:"productId"`
}

func (o SubscriptionGetOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewSubscriptionProductID(o.ProductID.String()); err != nil {
		return err
	}
	return nil
}

type SubscriptionGetter interface {
	GetSubscription(ctx context.Context, packageName PackageName, productID SubscriptionProductID) (Subscription, error)
}

func GetSubscription(ctx context.Context, getter SubscriptionGetter, options SubscriptionGetOptions) (Subscription, error) {
	if err := options.Validate(); err != nil {
		return Subscription{}, err
	}
	if getter == nil {
		return Subscription{}, fmt.Errorf("subscription getter is required")
	}
	return getter.GetSubscription(ctx, options.PackageName, options.ProductID)
}

type SubscriptionBatchGetOptions struct {
	PackageName PackageName             `json:"packageName"`
	ProductIDs  []SubscriptionProductID `json:"productIds"`
}

func (o SubscriptionBatchGetOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if len(o.ProductIDs) == 0 {
		return fmt.Errorf("at least one subscription product ID is required")
	}
	if len(o.ProductIDs) > 100 {
		return fmt.Errorf("subscription batch-get cannot exceed 100 product IDs")
	}
	seen := map[SubscriptionProductID]struct{}{}
	for _, productID := range o.ProductIDs {
		if _, err := NewSubscriptionProductID(productID.String()); err != nil {
			return err
		}
		if _, ok := seen[productID]; ok {
			return fmt.Errorf("subscription product ID %q is duplicated", productID)
		}
		seen[productID] = struct{}{}
	}
	return nil
}

type SubscriptionBatchGetResult struct {
	PackageName   PackageName                 `json:"packageName"`
	Subscriptions []Subscription              `json:"subscriptions"`
	Options       SubscriptionBatchGetOptions `json:"options"`
}

type SubscriptionBatchGetter interface {
	BatchGetSubscriptions(ctx context.Context, options SubscriptionBatchGetOptions) (SubscriptionBatchGetResult, error)
}

type SubscriptionDeleter interface {
	DeleteSubscription(ctx context.Context, options SubscriptionDeleteOptions) error
}

func BatchGetSubscriptions(ctx context.Context, getter SubscriptionBatchGetter, options SubscriptionBatchGetOptions) (SubscriptionBatchGetResult, error) {
	if err := options.Validate(); err != nil {
		return SubscriptionBatchGetResult{}, err
	}
	if getter == nil {
		return SubscriptionBatchGetResult{}, fmt.Errorf("subscription batch getter is required")
	}
	return getter.BatchGetSubscriptions(ctx, options)
}

type SubscriptionDeleteOptions struct {
	PackageName PackageName           `json:"packageName"`
	ProductID   SubscriptionProductID `json:"productId"`
	Confirm     bool                  `json:"confirm"`
	DryRun      bool                  `json:"dryRun"`
}

func (o SubscriptionDeleteOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewSubscriptionProductID(o.ProductID.String()); err != nil {
		return err
	}
	if o.Confirm && o.DryRun {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.Confirm && !o.DryRun {
		return fmt.Errorf("subscription deletion requires --confirm or --dry-run")
	}
	return nil
}

func (o SubscriptionDeleteOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live subscription deletion cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live subscription deletion requires --confirm")
	}
	return nil
}

type SubscriptionDeletePlan struct {
	PackageName PackageName           `json:"packageName"`
	ProductID   SubscriptionProductID `json:"productId"`
	Confirm     bool                  `json:"confirm"`
	Steps       []string              `json:"steps"`
}

type SubscriptionDeleteResult struct {
	PackageName PackageName            `json:"packageName"`
	ProductID   SubscriptionProductID  `json:"productId"`
	DryRun      bool                   `json:"dryRun"`
	Deleted     bool                   `json:"deleted"`
	Plan        SubscriptionDeletePlan `json:"plan"`
}

func DeleteSubscription(ctx context.Context, deleter SubscriptionDeleter, options SubscriptionDeleteOptions) (SubscriptionDeleteResult, error) {
	if err := options.Validate(); err != nil {
		return SubscriptionDeleteResult{}, err
	}
	result := SubscriptionDeleteResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		DryRun:      options.DryRun,
		Plan: SubscriptionDeletePlan{
			PackageName: options.PackageName,
			ProductID:   options.ProductID,
			Confirm:     options.Confirm,
			Steps:       []string{"delete subscription"},
		},
	}
	if options.DryRun {
		return result, nil
	}
	if deleter == nil {
		return SubscriptionDeleteResult{}, fmt.Errorf("subscription deleter is required")
	}
	if err := deleter.DeleteSubscription(ctx, options); err != nil {
		return SubscriptionDeleteResult{}, err
	}
	result.Deleted = true
	return result, nil
}

type BasePlanStateUpdateOptions struct {
	PackageName      PackageName                   `json:"packageName"`
	ProductID        SubscriptionProductID         `json:"productId"`
	BasePlanID       SubscriptionBasePlanID        `json:"basePlanId"`
	Action           BasePlanStateAction           `json:"action"`
	LatencyTolerance ProductUpdateLatencyTolerance `json:"latencyTolerance"`
	Confirm          bool                          `json:"confirm"`
	DryRun           bool                          `json:"dryRun"`
}

func (o BasePlanStateUpdateOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewSubscriptionProductID(o.ProductID.String()); err != nil {
		return err
	}
	if _, err := NewSubscriptionBasePlanID(o.BasePlanID.String()); err != nil {
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
		return fmt.Errorf("base plan state update requires --confirm or --dry-run")
	}
	return nil
}

func (o BasePlanStateUpdateOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live base plan state update cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live base plan state update requires --confirm")
	}
	return nil
}

type BasePlanStateUpdatePlan struct {
	PackageName      PackageName                   `json:"packageName"`
	ProductID        SubscriptionProductID         `json:"productId"`
	BasePlanID       SubscriptionBasePlanID        `json:"basePlanId"`
	Action           BasePlanStateAction           `json:"action"`
	LatencyTolerance ProductUpdateLatencyTolerance `json:"latencyTolerance"`
	Confirm          bool                          `json:"confirm"`
	Steps            []string                      `json:"steps"`
}

type BasePlanStateUpdateResult struct {
	PackageName  PackageName             `json:"packageName"`
	ProductID    SubscriptionProductID   `json:"productId"`
	BasePlanID   SubscriptionBasePlanID  `json:"basePlanId"`
	Action       BasePlanStateAction     `json:"action"`
	DryRun       bool                    `json:"dryRun"`
	Applied      bool                    `json:"applied"`
	Subscription *Subscription           `json:"subscription,omitempty"`
	Plan         BasePlanStateUpdatePlan `json:"plan"`
}

type BasePlanStateUpdater interface {
	UpdateBasePlanState(ctx context.Context, options BasePlanStateUpdateOptions) (Subscription, error)
}

func NewBasePlanStateUpdatePlan(options BasePlanStateUpdateOptions) (BasePlanStateUpdatePlan, error) {
	if err := options.Validate(); err != nil {
		return BasePlanStateUpdatePlan{}, err
	}
	return BasePlanStateUpdatePlan{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		BasePlanID:       options.BasePlanID,
		Action:           options.Action,
		LatencyTolerance: options.LatencyTolerance,
		Confirm:          options.Confirm,
		Steps:            basePlanStateUpdateSteps(options),
	}, nil
}

func UpdateBasePlanState(ctx context.Context, updater BasePlanStateUpdater, options BasePlanStateUpdateOptions) (BasePlanStateUpdateResult, error) {
	plan, err := NewBasePlanStateUpdatePlan(options)
	if err != nil {
		return BasePlanStateUpdateResult{}, err
	}
	result := BasePlanStateUpdateResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		BasePlanID:  options.BasePlanID,
		Action:      options.Action,
		DryRun:      options.DryRun,
		Applied:     false,
		Plan:        plan,
	}
	if options.DryRun {
		return result, nil
	}
	if updater == nil {
		return BasePlanStateUpdateResult{}, fmt.Errorf("base plan state updater is required")
	}
	subscription, err := updater.UpdateBasePlanState(ctx, options)
	if err != nil {
		return BasePlanStateUpdateResult{}, err
	}
	result.Applied = true
	result.Subscription = &subscription
	return result, nil
}

func basePlanStateUpdateSteps(options BasePlanStateUpdateOptions) []string {
	if options.DryRun {
		return []string{fmt.Sprintf("plan %s base plan", options.Action)}
	}
	return []string{fmt.Sprintf("%s base plan", options.Action)}
}
