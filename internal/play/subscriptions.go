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
	BasePlanID            string                   `json:"basePlanId"`
	State                 SubscriptionState        `json:"state,omitempty"`
	Type                  SubscriptionBasePlanType `json:"type,omitempty"`
	BillingPeriodDuration string                   `json:"billingPeriodDuration,omitempty"`
	GracePeriodDuration   string                   `json:"gracePeriodDuration,omitempty"`
	AccountHoldDuration   string                   `json:"accountHoldDuration,omitempty"`
	LegacyCompatible      bool                     `json:"legacyCompatible,omitempty"`
	OfferTags             []string                 `json:"offerTags,omitempty"`
	RegionalConfigCount   int                      `json:"regionalConfigCount,omitempty"`
}

type Subscription struct {
	PackageName         PackageName            `json:"packageName"`
	ProductID           SubscriptionProductID  `json:"productId"`
	Archived            bool                   `json:"archived,omitempty"`
	Listings            []SubscriptionListing  `json:"listings"`
	BasePlans           []SubscriptionBasePlan `json:"basePlans"`
	RestrictedCountries []string               `json:"restrictedCountries,omitempty"`
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
