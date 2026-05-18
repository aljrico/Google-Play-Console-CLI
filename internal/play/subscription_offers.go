package play

import (
	"context"
	"fmt"
)

type SubscriptionBasePlanID string

func NewSubscriptionBasePlanID(value string) (SubscriptionBasePlanID, error) {
	if value == "" {
		return "", fmt.Errorf("subscription base plan ID is required")
	}
	return SubscriptionBasePlanID(value), nil
}

func (b SubscriptionBasePlanID) String() string {
	return string(b)
}

type SubscriptionOfferID string

func NewSubscriptionOfferID(value string) (SubscriptionOfferID, error) {
	if value == "" {
		return "", fmt.Errorf("subscription offer ID is required")
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
	RegionalConfigs    []SubscriptionOfferRegionalConfig    `json:"regionalConfigs,omitempty"`
	OtherRegionsConfig *SubscriptionOfferOtherRegionsConfig `json:"otherRegionsConfig,omitempty"`
	Phases             []SubscriptionOfferPhase             `json:"phases,omitempty"`
	Targeting          *SubscriptionOfferTargeting          `json:"targeting,omitempty"`
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
	RegionalConfigs    []SubscriptionOfferPhaseRegionalConfig    `json:"regionalConfigs,omitempty"`
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
	if _, err := NewSubscriptionProductID(o.ProductID.String()); err != nil {
		return err
	}
	if _, err := NewSubscriptionBasePlanID(o.BasePlanID.String()); err != nil {
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
