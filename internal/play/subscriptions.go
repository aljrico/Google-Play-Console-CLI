package play

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
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

type SubscriptionCreateOptions struct {
	PackageName    PackageName           `json:"packageName"`
	ProductID      SubscriptionProductID `json:"productId"`
	Subscription   Subscription          `json:"subscription"`
	RegionsVersion string                `json:"regionsVersion"`
	Confirm        bool                  `json:"confirm"`
	DryRun         bool                  `json:"dryRun"`
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

func (o SubscriptionCreateOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewSubscriptionProductID(o.ProductID.String()); err != nil {
		return err
	}
	if strings.TrimSpace(o.RegionsVersion) == "" {
		return fmt.Errorf("regions version is required")
	}
	if strings.TrimSpace(o.RegionsVersion) != o.RegionsVersion {
		return fmt.Errorf("regions version cannot have leading or trailing whitespace")
	}
	if err := validateSubscriptionForCreate(subscriptionCreateDesiredSubscription(o)); err != nil {
		return err
	}
	if o.Confirm && o.DryRun {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.Confirm && !o.DryRun {
		return fmt.Errorf("subscription create requires --confirm or --dry-run")
	}
	return nil
}

func (o SubscriptionCreateOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live subscription create cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live subscription create requires --confirm")
	}
	return nil
}

func validateSubscriptionForCreate(subscription Subscription) error {
	if len(subscription.Listings) == 0 {
		return fmt.Errorf("subscription create requires at least one listing")
	}
	seenListings := map[string]struct{}{}
	for _, listing := range subscription.Listings {
		language, err := NewListingLanguage(listing.LanguageCode)
		if err != nil {
			return err
		}
		if err := validateSubscriptionListing(SubscriptionListing{
			LanguageCode: language.String(),
			Title:        listing.Title,
			Description:  listing.Description,
			Benefits:     listing.Benefits,
		}); err != nil {
			return err
		}
		if _, ok := seenListings[language.String()]; ok {
			return fmt.Errorf("subscription create listing %s is duplicated", language)
		}
		seenListings[language.String()] = struct{}{}
	}
	if len(subscription.BasePlans) == 0 {
		return fmt.Errorf("subscription create requires at least one base plan")
	}
	if err := validateSubscriptionRestrictedCountriesForCreate(subscription.RestrictedCountries); err != nil {
		return err
	}
	if subscription.TaxAndComplianceSettings != nil {
		if err := validateProductTaxComplianceSettings(*subscription.TaxAndComplianceSettings); err != nil {
			return err
		}
	}
	seenBasePlans := map[string]struct{}{}
	legacyCompatibleBasePlanCount := 0
	for _, basePlan := range subscription.BasePlans {
		if _, err := NewSubscriptionBasePlanID(basePlan.BasePlanID); err != nil {
			return err
		}
		if _, ok := seenBasePlans[basePlan.BasePlanID]; ok {
			return fmt.Errorf("subscription create base plan %s is duplicated", basePlan.BasePlanID)
		}
		seenBasePlans[basePlan.BasePlanID] = struct{}{}
		if err := validateSubscriptionBasePlanForCreate(basePlan); err != nil {
			return fmt.Errorf("subscription create base plan %s: %w", basePlan.BasePlanID, err)
		}
		if basePlan.LegacyCompatible {
			legacyCompatibleBasePlanCount++
			if legacyCompatibleBasePlanCount > 1 {
				return fmt.Errorf("subscription create can set at most one legacy-compatible base plan")
			}
		}
	}
	return nil
}

func validateSubscriptionBasePlanForCreate(basePlan SubscriptionBasePlan) error {
	if len(basePlan.OfferTags) > 20 {
		return fmt.Errorf("supports at most 20 offer tags")
	}
	seenTags := map[string]struct{}{}
	for _, tag := range basePlan.OfferTags {
		if err := validateSubscriptionOfferTag(tag); err != nil {
			return err
		}
		if _, ok := seenTags[tag]; ok {
			return fmt.Errorf("offer tag %q is duplicated", tag)
		}
		seenTags[tag] = struct{}{}
	}
	switch basePlan.Type {
	case SubscriptionBasePlanTypeAutoRenewing:
		if basePlan.TimeExtension != "" || basePlan.CommittedPaymentsCount != 0 || basePlan.RenewalType != "" {
			return fmt.Errorf("auto-renewing base plans cannot set prepaid or installment fields")
		}
		if err := validateBasePlanBillingPeriod("auto-renewing billing period duration", basePlan.BillingPeriodDuration); err != nil {
			return err
		}
		if err := validateOptionalBasePlanGracePeriod("auto-renewing grace period duration", basePlan.GracePeriodDuration, basePlan.BillingPeriodDuration); err != nil {
			return err
		}
		if err := validateOptionalISODayDurationRange("auto-renewing account hold duration", basePlan.AccountHoldDuration, 0, 60); err != nil {
			return err
		}
		if err := validateOptionalGraceAndHoldSum("auto-renewing", basePlan.GracePeriodDuration, basePlan.AccountHoldDuration); err != nil {
			return err
		}
		if basePlan.LegacyCompatibleSubscriptionOfferID != "" {
			if _, err := NewSubscriptionOfferID(basePlan.LegacyCompatibleSubscriptionOfferID); err != nil {
				return err
			}
		}
		if err := validateBasePlanProrationMode(basePlan.ProrationMode); err != nil {
			return err
		}
		if err := validateBasePlanResubscribeState(basePlan.ResubscribeState); err != nil {
			return err
		}
	case SubscriptionBasePlanTypePrepaid:
		if basePlan.GracePeriodDuration != "" || basePlan.AccountHoldDuration != "" || basePlan.LegacyCompatible || basePlan.LegacyCompatibleSubscriptionOfferID != "" || basePlan.ProrationMode != "" || basePlan.ResubscribeState != "" || basePlan.CommittedPaymentsCount != 0 || basePlan.RenewalType != "" {
			return fmt.Errorf("prepaid base plans cannot set auto-renewing or installment fields")
		}
		if err := validateBasePlanBillingPeriod("prepaid billing period duration", basePlan.BillingPeriodDuration); err != nil {
			return err
		}
		if err := validateBasePlanTimeExtension(basePlan.TimeExtension); err != nil {
			return err
		}
	case SubscriptionBasePlanTypeInstallments:
		if basePlan.TimeExtension != "" || basePlan.LegacyCompatible || basePlan.LegacyCompatibleSubscriptionOfferID != "" {
			return fmt.Errorf("installments base plans cannot set prepaid or legacy-compatible auto-renewing fields")
		}
		if err := validateBasePlanBillingPeriod("installments billing period duration", basePlan.BillingPeriodDuration); err != nil {
			return err
		}
		if err := validateOptionalBasePlanGracePeriod("installments grace period duration", basePlan.GracePeriodDuration, basePlan.BillingPeriodDuration); err != nil {
			return err
		}
		if err := validateOptionalISODayDurationRange("installments account hold duration", basePlan.AccountHoldDuration, 0, 60); err != nil {
			return err
		}
		if err := validateOptionalGraceAndHoldSum("installments", basePlan.GracePeriodDuration, basePlan.AccountHoldDuration); err != nil {
			return err
		}
		if basePlan.CommittedPaymentsCount <= 0 {
			return fmt.Errorf("committed payments count must be greater than 0")
		}
		if err := validateBasePlanProrationMode(basePlan.ProrationMode); err != nil {
			return err
		}
		if err := validateBasePlanResubscribeState(basePlan.ResubscribeState); err != nil {
			return err
		}
		if err := validateBasePlanRenewalType(basePlan.RenewalType); err != nil {
			return err
		}
	default:
		return fmt.Errorf("base plan type must be auto-renewing, prepaid, or installments")
	}
	if len(basePlan.RegionalConfigs) == 0 {
		return fmt.Errorf("requires at least one regional config")
	}
	seenRegions := map[string]struct{}{}
	for _, config := range basePlan.RegionalConfigs {
		if !isValidRegionCode(config.RegionCode) {
			return fmt.Errorf("regional config region must be a two-letter ISO 3166 code")
		}
		if _, ok := seenRegions[config.RegionCode]; ok {
			return fmt.Errorf("regional config %s is duplicated", config.RegionCode)
		}
		seenRegions[config.RegionCode] = struct{}{}
		if config.NewSubscriberAvailability && config.Price == nil {
			return fmt.Errorf("regional config %s requires price when available to new subscribers", config.RegionCode)
		}
		if config.Price != nil {
			if err := validateMoney(*config.Price); err != nil {
				return fmt.Errorf("regional config %s: %w", config.RegionCode, err)
			}
		}
	}
	if basePlan.OtherRegionsConfig != nil {
		if err := validateSubscriptionOtherRegionsConfigForCreate(*basePlan.OtherRegionsConfig); err != nil {
			return err
		}
	}
	return nil
}

func validateOptionalISODuration(fieldName, value string) error {
	if value == "" {
		return nil
	}
	return validateISODuration(fieldName, value)
}

func validateBasePlanBillingPeriod(fieldName, value string) error {
	if err := validateISODuration(fieldName, value); err != nil {
		return err
	}
	switch value {
	case "P1W", "P4W", "P1M", "P3M", "P6M", "P1Y":
		return nil
	default:
		return fmt.Errorf("%s must be one of P1W, P4W, P1M, P3M, P6M, or P1Y", fieldName)
	}
}

func validateOptionalISODayDurationRange(fieldName, value string, minimumDays, maximumDays int) error {
	if value == "" {
		return nil
	}
	days, err := parseISODayDuration(fieldName, value)
	if err != nil {
		return err
	}
	if days < minimumDays || days > maximumDays {
		return fmt.Errorf("%s must be between P%dD and P%dD", fieldName, minimumDays, maximumDays)
	}
	return nil
}

func validateOptionalBasePlanGracePeriod(fieldName, gracePeriodDuration, billingPeriodDuration string) error {
	maximumDays := maximumGracePeriodDaysForBillingPeriod(billingPeriodDuration)
	if maximumDays == 0 {
		maximumDays = 30
	}
	return validateOptionalISODayDurationRange(fieldName, gracePeriodDuration, 0, maximumDays)
}

func maximumGracePeriodDaysForBillingPeriod(billingPeriodDuration string) int {
	switch billingPeriodDuration {
	case "P1W":
		return 7
	case "P4W":
		return 28
	case "P1M", "P3M", "P6M", "P1Y":
		return 30
	default:
		return 0
	}
}

func validateOptionalGraceAndHoldSum(planType, gracePeriodDuration, accountHoldDuration string) error {
	if gracePeriodDuration == "" || accountHoldDuration == "" {
		return nil
	}
	graceDays, err := parseISODayDuration(planType+" grace period duration", gracePeriodDuration)
	if err != nil {
		return err
	}
	holdDays, err := parseISODayDuration(planType+" account hold duration", accountHoldDuration)
	if err != nil {
		return err
	}
	total := graceDays + holdDays
	if total < 30 || total > 60 {
		return fmt.Errorf("%s grace period duration plus account hold duration must be between P30D and P60D", planType)
	}
	return nil
}

func parseISODayDuration(fieldName, value string) (int, error) {
	if strings.TrimSpace(value) != value || !strings.HasPrefix(value, "P") || !strings.HasSuffix(value, "D") || strings.Contains(value, "T") {
		return 0, fmt.Errorf("%s must be an ISO 8601 day duration", fieldName)
	}
	rawDays := strings.TrimSuffix(strings.TrimPrefix(value, "P"), "D")
	if rawDays == "" || strings.ContainsAny(rawDays, "YMW.,") {
		return 0, fmt.Errorf("%s must be an ISO 8601 day duration", fieldName)
	}
	days, err := strconv.Atoi(rawDays)
	if err != nil {
		return 0, fmt.Errorf("%s must be an ISO 8601 day duration", fieldName)
	}
	return days, nil
}

func validateBasePlanProrationMode(value string) error {
	if value == "" {
		return nil
	}
	switch value {
	case "SUBSCRIPTION_PRORATION_MODE_CHARGE_ON_NEXT_BILLING_DATE", "SUBSCRIPTION_PRORATION_MODE_CHARGE_FULL_PRICE_IMMEDIATELY":
		return nil
	default:
		return fmt.Errorf("unsupported proration mode %q", value)
	}
}

func validateBasePlanResubscribeState(value string) error {
	if value == "" {
		return nil
	}
	switch value {
	case "RESUBSCRIBE_STATE_ACTIVE", "RESUBSCRIBE_STATE_INACTIVE":
		return nil
	default:
		return fmt.Errorf("unsupported resubscribe state %q", value)
	}
}

func validateBasePlanTimeExtension(value string) error {
	if value == "" {
		return nil
	}
	switch value {
	case "TIME_EXTENSION_ACTIVE", "TIME_EXTENSION_INACTIVE":
		return nil
	default:
		return fmt.Errorf("unsupported time extension %q", value)
	}
}

func validateBasePlanRenewalType(value string) error {
	switch value {
	case "RENEWAL_TYPE_RENEWS_WITHOUT_COMMITMENT", "RENEWAL_TYPE_RENEWS_WITH_COMMITMENT":
		return nil
	case "":
		return fmt.Errorf("installments renewal type is required")
	default:
		return fmt.Errorf("unsupported renewal type %q", value)
	}
}

func validateSubscriptionRestrictedCountriesForCreate(countries []string) error {
	seen := map[string]struct{}{}
	for _, country := range countries {
		if !isValidRegionCode(country) {
			return fmt.Errorf("restricted country must be a two-letter ISO 3166 code")
		}
		if _, ok := seen[country]; ok {
			return fmt.Errorf("restricted country %s is duplicated", country)
		}
		seen[country] = struct{}{}
	}
	return nil
}

func validateSubscriptionOtherRegionsConfigForCreate(config SubscriptionOtherRegionsConfig) error {
	if config.USDPrice == nil || config.EURPrice == nil {
		return fmt.Errorf("other regions config requires usdPrice and eurPrice")
	}
	if config.USDPrice.CurrencyCode != "USD" {
		return fmt.Errorf("other regions usdPrice currencyCode must be USD")
	}
	if config.EURPrice.CurrencyCode != "EUR" {
		return fmt.Errorf("other regions eurPrice currencyCode must be EUR")
	}
	if err := validateMoney(*config.USDPrice); err != nil {
		return err
	}
	if err := validateMoney(*config.EURPrice); err != nil {
		return err
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

type SubscriptionCreator interface {
	CreateSubscription(ctx context.Context, options SubscriptionCreateOptions) (Subscription, error)
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

type SubscriptionCreatePlan struct {
	Action         string                `json:"action"`
	PackageName    PackageName           `json:"packageName"`
	ProductID      SubscriptionProductID `json:"productId"`
	RegionsVersion string                `json:"regionsVersion"`
	Confirm        bool                  `json:"confirm"`
	Steps          []string              `json:"steps"`
}

type SubscriptionCreateResult struct {
	Action       string                 `json:"action"`
	DryRun       bool                   `json:"dryRun"`
	Created      bool                   `json:"created"`
	Subscription *Subscription          `json:"subscription,omitempty"`
	Desired      Subscription           `json:"desiredSubscription"`
	Plan         SubscriptionCreatePlan `json:"plan"`
}

func CreateSubscription(ctx context.Context, creator SubscriptionCreator, options SubscriptionCreateOptions) (SubscriptionCreateResult, error) {
	if err := options.Validate(); err != nil {
		return SubscriptionCreateResult{}, err
	}
	desired := subscriptionCreateDesiredSubscription(options)
	result := SubscriptionCreateResult{
		Action:  "create",
		DryRun:  options.DryRun,
		Desired: desired,
		Plan: SubscriptionCreatePlan{
			Action:         "create",
			PackageName:    options.PackageName,
			ProductID:      options.ProductID,
			RegionsVersion: options.RegionsVersion,
			Confirm:        options.Confirm,
			Steps:          subscriptionCreateSteps(options.DryRun),
		},
	}
	if options.DryRun {
		return result, nil
	}
	if creator == nil {
		return SubscriptionCreateResult{}, fmt.Errorf("subscription creator is required")
	}
	subscription, err := creator.CreateSubscription(ctx, options)
	if err != nil {
		return SubscriptionCreateResult{}, err
	}
	result.Created = true
	result.Subscription = &subscription
	return result, nil
}

func subscriptionCreateDesiredSubscription(options SubscriptionCreateOptions) Subscription {
	subscription := options.Subscription
	subscription.PackageName = options.PackageName
	subscription.ProductID = options.ProductID
	subscription.Archived = false
	for index := range subscription.Listings {
		if language, err := NewListingLanguage(subscription.Listings[index].LanguageCode); err == nil {
			subscription.Listings[index].LanguageCode = language.String()
		}
	}
	for index := range subscription.BasePlans {
		subscription.BasePlans[index].State = ""
	}
	return subscription
}

type SubscriptionDeleteOptions struct {
	PackageName PackageName           `json:"packageName"`
	ProductID   SubscriptionProductID `json:"productId"`
	Confirm     bool                  `json:"confirm"`
	DryRun      bool                  `json:"dryRun"`
}

type BasePlanDeleteOptions struct {
	PackageName PackageName            `json:"packageName"`
	ProductID   SubscriptionProductID  `json:"productId"`
	BasePlanID  SubscriptionBasePlanID `json:"basePlanId"`
	Confirm     bool                   `json:"confirm"`
	DryRun      bool                   `json:"dryRun"`
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

func (o BasePlanDeleteOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewSubscriptionProductID(o.ProductID.String()); err != nil {
		return err
	}
	if _, err := NewSubscriptionBasePlanID(o.BasePlanID.String()); err != nil {
		return err
	}
	if o.Confirm && o.DryRun {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.Confirm && !o.DryRun {
		return fmt.Errorf("base plan deletion requires --confirm or --dry-run")
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

func (o BasePlanDeleteOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live base plan deletion cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live base plan deletion requires --confirm")
	}
	return nil
}

type SubscriptionDeletePlan struct {
	PackageName PackageName           `json:"packageName"`
	ProductID   SubscriptionProductID `json:"productId"`
	Confirm     bool                  `json:"confirm"`
	Steps       []string              `json:"steps"`
}

type BasePlanDeletePlan struct {
	PackageName PackageName            `json:"packageName"`
	ProductID   SubscriptionProductID  `json:"productId"`
	BasePlanID  SubscriptionBasePlanID `json:"basePlanId"`
	Confirm     bool                   `json:"confirm"`
	Steps       []string               `json:"steps"`
}

type SubscriptionDeleteResult struct {
	PackageName PackageName            `json:"packageName"`
	ProductID   SubscriptionProductID  `json:"productId"`
	DryRun      bool                   `json:"dryRun"`
	Deleted     bool                   `json:"deleted"`
	Plan        SubscriptionDeletePlan `json:"plan"`
}

type BasePlanDeleteResult struct {
	PackageName PackageName            `json:"packageName"`
	ProductID   SubscriptionProductID  `json:"productId"`
	BasePlanID  SubscriptionBasePlanID `json:"basePlanId"`
	DryRun      bool                   `json:"dryRun"`
	Deleted     bool                   `json:"deleted"`
	Plan        BasePlanDeletePlan     `json:"plan"`
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

type BasePlanDeleter interface {
	DeleteBasePlan(ctx context.Context, options BasePlanDeleteOptions) error
}

func DeleteBasePlan(ctx context.Context, deleter BasePlanDeleter, options BasePlanDeleteOptions) (BasePlanDeleteResult, error) {
	if err := options.Validate(); err != nil {
		return BasePlanDeleteResult{}, err
	}
	result := BasePlanDeleteResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		BasePlanID:  options.BasePlanID,
		DryRun:      options.DryRun,
		Plan: BasePlanDeletePlan{
			PackageName: options.PackageName,
			ProductID:   options.ProductID,
			BasePlanID:  options.BasePlanID,
			Confirm:     options.Confirm,
			Steps:       []string{"delete base plan"},
		},
	}
	if options.DryRun {
		return result, nil
	}
	if deleter == nil {
		return BasePlanDeleteResult{}, fmt.Errorf("base plan deleter is required")
	}
	if err := deleter.DeleteBasePlan(ctx, options); err != nil {
		return BasePlanDeleteResult{}, err
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

type BasePlanBatchStateUpdateRequest struct {
	ProductID  SubscriptionProductID  `json:"productId"`
	BasePlanID SubscriptionBasePlanID `json:"basePlanId"`
}

type BasePlanBatchStateUpdateOptions struct {
	PackageName      PackageName                       `json:"packageName"`
	ProductID        SubscriptionProductID             `json:"productId"`
	Requests         []BasePlanBatchStateUpdateRequest `json:"requests"`
	Action           BasePlanStateAction               `json:"action"`
	LatencyTolerance ProductUpdateLatencyTolerance     `json:"latencyTolerance"`
	Confirm          bool                              `json:"confirm"`
	DryRun           bool                              `json:"dryRun"`
}

type BasePlanPriceIncreaseType string

const (
	BasePlanPriceIncreaseTypeOptIn  BasePlanPriceIncreaseType = "optIn"
	BasePlanPriceIncreaseTypeOptOut BasePlanPriceIncreaseType = "optOut"
)

func (t BasePlanPriceIncreaseType) String() string {
	return string(t)
}

func (t BasePlanPriceIncreaseType) Validate() error {
	switch t {
	case "", BasePlanPriceIncreaseTypeOptIn, BasePlanPriceIncreaseTypeOptOut:
		return nil
	default:
		return fmt.Errorf("unsupported base plan price increase type %q", t)
	}
}

type BasePlanPriceMigrationConfig struct {
	RegionCode                    string                    `json:"regionCode"`
	OldestAllowedPriceVersionTime string                    `json:"oldestAllowedPriceVersionTime"`
	PriceIncreaseType             BasePlanPriceIncreaseType `json:"priceIncreaseType,omitempty"`
}

type BasePlanPriceMigrationRequest struct {
	ProductID  SubscriptionProductID          `json:"productId"`
	BasePlanID SubscriptionBasePlanID         `json:"basePlanId"`
	Regions    []BasePlanPriceMigrationConfig `json:"regions"`
}

type BasePlanPricePatchRequest struct {
	ProductID  SubscriptionProductID  `json:"productId"`
	BasePlanID SubscriptionBasePlanID `json:"basePlanId"`
	RegionCode string                 `json:"regionCode"`
	Price      Money                  `json:"price"`
}

type BasePlanBatchPriceMigrationOptions struct {
	PackageName      PackageName                     `json:"packageName"`
	ProductID        SubscriptionProductID           `json:"productId"`
	RegionsVersion   string                          `json:"regionsVersion"`
	Requests         []BasePlanPriceMigrationRequest `json:"requests"`
	LatencyTolerance ProductUpdateLatencyTolerance   `json:"latencyTolerance"`
	Confirm          bool                            `json:"confirm"`
	DryRun           bool                            `json:"dryRun"`
}

type BasePlanBatchPatchPriceOptions struct {
	PackageName      PackageName                   `json:"packageName"`
	ProductID        SubscriptionProductID         `json:"productId"`
	RegionsVersion   string                        `json:"regionsVersion"`
	Requests         []BasePlanPricePatchRequest   `json:"requests"`
	LatencyTolerance ProductUpdateLatencyTolerance `json:"latencyTolerance"`
	Confirm          bool                          `json:"confirm"`
	DryRun           bool                          `json:"dryRun"`
}

type SubscriptionPatchOptions struct {
	PackageName      PackageName                   `json:"packageName"`
	ProductID        SubscriptionProductID         `json:"productId"`
	Listing          SubscriptionListing           `json:"listing"`
	DescriptionSet   bool                          `json:"descriptionSet,omitempty"`
	BenefitsSet      bool                          `json:"benefitsSet,omitempty"`
	RegionsVersion   string                        `json:"regionsVersion"`
	LatencyTolerance ProductUpdateLatencyTolerance `json:"latencyTolerance"`
	Confirm          bool                          `json:"confirm"`
	DryRun           bool                          `json:"dryRun"`
}

type SubscriptionBatchPatchListingRequest struct {
	ProductID SubscriptionProductID `json:"productId"`
	Listing   SubscriptionListing   `json:"listing"`
}

type SubscriptionBatchPatchListingsOptions struct {
	PackageName      PackageName                            `json:"packageName"`
	Requests         []SubscriptionBatchPatchListingRequest `json:"requests"`
	RegionsVersion   string                                 `json:"regionsVersion"`
	LatencyTolerance ProductUpdateLatencyTolerance          `json:"latencyTolerance"`
	Confirm          bool                                   `json:"confirm"`
	DryRun           bool                                   `json:"dryRun"`
}

func (o SubscriptionPatchOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewSubscriptionProductID(o.ProductID.String()); err != nil {
		return err
	}
	if err := validateSubscriptionListing(o.Listing); err != nil {
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
		return fmt.Errorf("subscription patch requires --confirm or --dry-run")
	}
	return nil
}

func (o SubscriptionPatchOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live subscription patch cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live subscription patch requires --confirm")
	}
	return nil
}

func (o SubscriptionBatchPatchListingsOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if len(o.Requests) == 0 {
		return fmt.Errorf("at least one subscription listing patch is required")
	}
	seenListings := map[string]struct{}{}
	seenProducts := map[SubscriptionProductID]struct{}{}
	for _, request := range o.Requests {
		if _, err := NewSubscriptionProductID(request.ProductID.String()); err != nil {
			return err
		}
		if err := validateSubscriptionListing(request.Listing); err != nil {
			return err
		}
		key := subscriptionBatchPatchListingKey(request.ProductID, request.Listing.LanguageCode)
		if _, ok := seenListings[key]; ok {
			return fmt.Errorf("subscription listing %s is duplicated", key)
		}
		seenListings[key] = struct{}{}
		seenProducts[request.ProductID] = struct{}{}
	}
	if len(seenProducts) > 100 {
		return fmt.Errorf("subscription listing batch patch cannot exceed 100 subscriptions")
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
		return fmt.Errorf("subscription listing batch patch requires --confirm or --dry-run")
	}
	return nil
}

func (o SubscriptionBatchPatchListingsOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live subscription listing batch patch cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live subscription listing batch patch requires --confirm")
	}
	return nil
}

func subscriptionBatchPatchListingKey(productID SubscriptionProductID, languageCode string) string {
	return productID.String() + "/" + languageCode
}

func validateSubscriptionListing(listing SubscriptionListing) error {
	if _, err := NewListingLanguage(listing.LanguageCode); err != nil {
		return err
	}
	if strings.TrimSpace(listing.Title) == "" {
		return fmt.Errorf("subscription listing title is required")
	}
	if strings.TrimSpace(listing.Title) != listing.Title {
		return fmt.Errorf("subscription listing title cannot have leading or trailing whitespace")
	}
	if strings.TrimSpace(listing.Description) != listing.Description {
		return fmt.Errorf("subscription listing description cannot have leading or trailing whitespace")
	}
	if len(listing.Description) > 80 {
		return fmt.Errorf("subscription listing description cannot exceed 80 characters")
	}
	if len(listing.Benefits) > 4 {
		return fmt.Errorf("subscription listing cannot have more than 4 benefits")
	}
	for index, benefit := range listing.Benefits {
		if strings.TrimSpace(benefit) == "" {
			return fmt.Errorf("subscription listing benefit %d is required", index+1)
		}
		if strings.TrimSpace(benefit) != benefit {
			return fmt.Errorf("subscription listing benefit %d cannot have leading or trailing whitespace", index+1)
		}
	}
	return nil
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

func (o BasePlanBatchStateUpdateOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewSubscriptionBasePlanBatchProductID(o.ProductID.String()); err != nil {
		return err
	}
	if len(o.Requests) == 0 {
		return fmt.Errorf("at least one subscription base plan is required")
	}
	if len(o.Requests) > 100 {
		return fmt.Errorf("subscription base plan batch state update cannot exceed 100 base plans")
	}
	seen := map[string]struct{}{}
	seenProducts := map[SubscriptionProductID]struct{}{}
	for _, request := range o.Requests {
		if _, err := NewSubscriptionProductID(request.ProductID.String()); err != nil {
			return err
		}
		if _, err := NewSubscriptionBasePlanID(request.BasePlanID.String()); err != nil {
			return err
		}
		if o.ProductID.String() != SubscriptionOfferWildcardID && request.ProductID != o.ProductID {
			return fmt.Errorf("subscription base plan %s/%s does not match parent product ID %s", request.ProductID, request.BasePlanID, o.ProductID)
		}
		key := subscriptionBasePlanBatchStateUpdateKey(request.ProductID, request.BasePlanID)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("subscription base plan %s is duplicated", key)
		}
		seen[key] = struct{}{}
		seenProducts[request.ProductID] = struct{}{}
	}
	if len(seenProducts) == 1 && o.ProductID.String() == SubscriptionOfferWildcardID {
		return fmt.Errorf("single-product base plan batch state update requires parent product ID, not %q", SubscriptionOfferWildcardID)
	}
	if len(seenProducts) > 1 && o.ProductID.String() != SubscriptionOfferWildcardID {
		return fmt.Errorf("multi-product base plan batch state update requires parent product ID %q", SubscriptionOfferWildcardID)
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
		return fmt.Errorf("base plan batch state update requires --confirm or --dry-run")
	}
	return nil
}

func NewSubscriptionBasePlanBatchProductID(value string) (SubscriptionProductID, error) {
	if value == SubscriptionOfferWildcardID {
		return SubscriptionProductID(value), nil
	}
	return NewSubscriptionProductID(value)
}

func subscriptionBasePlanBatchStateUpdateKey(productID SubscriptionProductID, basePlanID SubscriptionBasePlanID) string {
	return productID.String() + "/" + basePlanID.String()
}

func (o BasePlanBatchPriceMigrationOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewSubscriptionBasePlanBatchProductID(o.ProductID.String()); err != nil {
		return err
	}
	if strings.TrimSpace(o.RegionsVersion) == "" {
		return fmt.Errorf("regions version is required")
	}
	if strings.TrimSpace(o.RegionsVersion) != o.RegionsVersion {
		return fmt.Errorf("regions version cannot have leading or trailing whitespace")
	}
	if len(o.Requests) == 0 {
		return fmt.Errorf("at least one subscription base plan price migration is required")
	}
	if len(o.Requests) > 100 {
		return fmt.Errorf("subscription base plan price migration cannot exceed 100 base plans")
	}
	seenBasePlans := map[string]struct{}{}
	seenProducts := map[SubscriptionProductID]struct{}{}
	for _, request := range o.Requests {
		if _, err := NewSubscriptionProductID(request.ProductID.String()); err != nil {
			return err
		}
		if _, err := NewSubscriptionBasePlanID(request.BasePlanID.String()); err != nil {
			return err
		}
		if o.ProductID.String() != SubscriptionOfferWildcardID && request.ProductID != o.ProductID {
			return fmt.Errorf("subscription base plan %s/%s does not match parent product ID %s", request.ProductID, request.BasePlanID, o.ProductID)
		}
		key := subscriptionBasePlanBatchStateUpdateKey(request.ProductID, request.BasePlanID)
		if _, ok := seenBasePlans[key]; ok {
			return fmt.Errorf("subscription base plan %s is duplicated", key)
		}
		if len(request.Regions) == 0 {
			return fmt.Errorf("subscription base plan %s requires at least one regional price migration", key)
		}
		seenRegions := map[string]struct{}{}
		for _, region := range request.Regions {
			if err := region.Validate(); err != nil {
				return err
			}
			if _, ok := seenRegions[region.RegionCode]; ok {
				return fmt.Errorf("subscription base plan %s has duplicate region %s", key, region.RegionCode)
			}
			seenRegions[region.RegionCode] = struct{}{}
		}
		seenBasePlans[key] = struct{}{}
		seenProducts[request.ProductID] = struct{}{}
	}
	if len(seenProducts) == 1 && o.ProductID.String() == SubscriptionOfferWildcardID {
		return fmt.Errorf("single-product base plan price migration requires parent product ID, not %q", SubscriptionOfferWildcardID)
	}
	if len(seenProducts) > 1 && o.ProductID.String() != SubscriptionOfferWildcardID {
		return fmt.Errorf("multi-product base plan price migration requires parent product ID %q", SubscriptionOfferWildcardID)
	}
	if _, err := NewProductUpdateLatencyTolerance(o.LatencyTolerance.String()); err != nil {
		return err
	}
	if o.Confirm && o.DryRun {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.Confirm && !o.DryRun {
		return fmt.Errorf("base plan price migration requires --confirm or --dry-run")
	}
	return nil
}

func (o BasePlanBatchPriceMigrationOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live base plan price migration cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live base plan price migration requires --confirm")
	}
	return nil
}

func (o BasePlanBatchPatchPriceOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewSubscriptionBasePlanBatchProductID(o.ProductID.String()); err != nil {
		return err
	}
	if strings.TrimSpace(o.RegionsVersion) == "" {
		return fmt.Errorf("regions version is required")
	}
	if strings.TrimSpace(o.RegionsVersion) != o.RegionsVersion {
		return fmt.Errorf("regions version cannot have leading or trailing whitespace")
	}
	if len(o.Requests) == 0 {
		return fmt.Errorf("at least one subscription base plan price patch is required")
	}
	seenRegions := map[string]struct{}{}
	seenProducts := map[SubscriptionProductID]struct{}{}
	for _, request := range o.Requests {
		if _, err := NewSubscriptionProductID(request.ProductID.String()); err != nil {
			return err
		}
		if _, err := NewSubscriptionBasePlanID(request.BasePlanID.String()); err != nil {
			return err
		}
		if o.ProductID.String() != SubscriptionOfferWildcardID && request.ProductID != o.ProductID {
			return fmt.Errorf("subscription base plan %s/%s does not match parent product ID %s", request.ProductID, request.BasePlanID, o.ProductID)
		}
		if !isValidRegionCode(request.RegionCode) {
			return fmt.Errorf("subscription base plan price region must be a two-letter ISO 3166 code")
		}
		if err := validateMoney(request.Price); err != nil {
			return fmt.Errorf("subscription base plan price for %s/%s/%s: %w", request.ProductID, request.BasePlanID, request.RegionCode, err)
		}
		key := subscriptionBasePlanPricePatchKey(request.ProductID, request.BasePlanID, request.RegionCode)
		if _, ok := seenRegions[key]; ok {
			return fmt.Errorf("subscription base plan price %s is duplicated", key)
		}
		seenRegions[key] = struct{}{}
		seenProducts[request.ProductID] = struct{}{}
	}
	if len(seenProducts) > 100 {
		return fmt.Errorf("subscription base plan price batch patch cannot exceed 100 subscriptions")
	}
	if len(seenProducts) == 1 && o.ProductID.String() == SubscriptionOfferWildcardID {
		return fmt.Errorf("single-product base plan price patch requires parent product ID, not %q", SubscriptionOfferWildcardID)
	}
	if len(seenProducts) > 1 && o.ProductID.String() != SubscriptionOfferWildcardID {
		return fmt.Errorf("multi-product base plan price patch requires parent product ID %q", SubscriptionOfferWildcardID)
	}
	if _, err := NewProductUpdateLatencyTolerance(o.LatencyTolerance.String()); err != nil {
		return err
	}
	if o.Confirm && o.DryRun {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.Confirm && !o.DryRun {
		return fmt.Errorf("base plan price batch patch requires --confirm or --dry-run")
	}
	return nil
}

func (o BasePlanBatchPatchPriceOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live base plan price batch patch cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live base plan price batch patch requires --confirm")
	}
	return nil
}

func subscriptionBasePlanPricePatchKey(productID SubscriptionProductID, basePlanID SubscriptionBasePlanID, regionCode string) string {
	return productID.String() + "/" + basePlanID.String() + "/" + regionCode
}

func (c BasePlanPriceMigrationConfig) Validate() error {
	if !isValidRegionCode(c.RegionCode) {
		return fmt.Errorf("invalid region code %q", c.RegionCode)
	}
	if strings.TrimSpace(c.OldestAllowedPriceVersionTime) == "" {
		return fmt.Errorf("oldest allowed price version time is required")
	}
	if strings.TrimSpace(c.OldestAllowedPriceVersionTime) != c.OldestAllowedPriceVersionTime {
		return fmt.Errorf("oldest allowed price version time cannot have leading or trailing whitespace")
	}
	if _, err := time.Parse(time.RFC3339, c.OldestAllowedPriceVersionTime); err != nil {
		return fmt.Errorf("oldest allowed price version time must be RFC3339: %w", err)
	}
	return c.PriceIncreaseType.Validate()
}

func (o BasePlanBatchStateUpdateOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live base plan batch state update cannot be a dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live base plan batch state update requires --confirm")
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

type BasePlanBatchStateUpdatePlan struct {
	PackageName      PackageName                       `json:"packageName"`
	ProductID        SubscriptionProductID             `json:"productId"`
	Requests         []BasePlanBatchStateUpdateRequest `json:"requests"`
	Action           BasePlanStateAction               `json:"action"`
	LatencyTolerance ProductUpdateLatencyTolerance     `json:"latencyTolerance"`
	Confirm          bool                              `json:"confirm"`
	Steps            []string                          `json:"steps"`
}

type BasePlanBatchStateUpdateResult struct {
	PackageName   PackageName                       `json:"packageName"`
	ProductID     SubscriptionProductID             `json:"productId"`
	Requests      []BasePlanBatchStateUpdateRequest `json:"requests"`
	Action        BasePlanStateAction               `json:"action"`
	DryRun        bool                              `json:"dryRun"`
	Applied       bool                              `json:"applied"`
	Subscriptions []Subscription                    `json:"subscriptions,omitempty"`
	Plan          BasePlanBatchStateUpdatePlan      `json:"plan"`
}

type BasePlanBatchPriceMigrationPlan struct {
	PackageName      PackageName                     `json:"packageName"`
	ProductID        SubscriptionProductID           `json:"productId"`
	RegionsVersion   string                          `json:"regionsVersion"`
	Requests         []BasePlanPriceMigrationRequest `json:"requests"`
	LatencyTolerance ProductUpdateLatencyTolerance   `json:"latencyTolerance"`
	Confirm          bool                            `json:"confirm"`
	Steps            []string                        `json:"steps"`
}

type BasePlanPriceMigrationResponse struct {
	ProductID  SubscriptionProductID  `json:"productId"`
	BasePlanID SubscriptionBasePlanID `json:"basePlanId"`
}

type BasePlanBatchPriceMigrationResult struct {
	PackageName PackageName                      `json:"packageName"`
	ProductID   SubscriptionProductID            `json:"productId"`
	DryRun      bool                             `json:"dryRun"`
	Applied     bool                             `json:"applied"`
	Responses   []BasePlanPriceMigrationResponse `json:"responses,omitempty"`
	Plan        BasePlanBatchPriceMigrationPlan  `json:"plan"`
}

type BasePlanBatchPatchPricePlan struct {
	PackageName      PackageName                   `json:"packageName"`
	ProductID        SubscriptionProductID         `json:"productId"`
	RegionsVersion   string                        `json:"regionsVersion"`
	Requests         []BasePlanPricePatchRequest   `json:"requests"`
	UpdateMask       string                        `json:"updateMask"`
	LatencyTolerance ProductUpdateLatencyTolerance `json:"latencyTolerance"`
	Confirm          bool                          `json:"confirm"`
	Steps            []string                      `json:"steps"`
}

type BasePlanBatchPatchPriceDesiredSubscription struct {
	PackageName PackageName                          `json:"packageName"`
	ProductID   SubscriptionProductID                `json:"productId"`
	BasePlans   []BasePlanBatchPatchPriceDesiredPlan `json:"basePlans"`
}

type BasePlanBatchPatchPriceDesiredPlan struct {
	BasePlanID      SubscriptionBasePlanID                 `json:"basePlanId"`
	RegionalConfigs []BasePlanBatchPatchPriceDesiredRegion `json:"regionalConfigs"`
}

type BasePlanBatchPatchPriceDesiredRegion struct {
	RegionCode string `json:"regionCode"`
	Price      Money  `json:"price"`
}

type BasePlanBatchPatchPriceResult struct {
	PackageName   PackageName                                  `json:"packageName"`
	ProductID     SubscriptionProductID                        `json:"productId"`
	Requests      []BasePlanPricePatchRequest                  `json:"requests"`
	DryRun        bool                                         `json:"dryRun"`
	Applied       bool                                         `json:"applied"`
	Desired       []BasePlanBatchPatchPriceDesiredSubscription `json:"desiredSubscriptions"`
	Subscriptions []Subscription                               `json:"subscriptions,omitempty"`
	Plan          BasePlanBatchPatchPricePlan                  `json:"plan"`
}

type SubscriptionPatchPlan struct {
	PackageName      PackageName                   `json:"packageName"`
	ProductID        SubscriptionProductID         `json:"productId"`
	Listing          SubscriptionListing           `json:"listing"`
	DescriptionSet   bool                          `json:"descriptionSet,omitempty"`
	BenefitsSet      bool                          `json:"benefitsSet,omitempty"`
	UpdateMask       string                        `json:"updateMask"`
	RegionsVersion   string                        `json:"regionsVersion"`
	LatencyTolerance ProductUpdateLatencyTolerance `json:"latencyTolerance"`
	Confirm          bool                          `json:"confirm"`
	Steps            []string                      `json:"steps"`
}

type SubscriptionBatchPatchListingsPlan struct {
	PackageName      PackageName                            `json:"packageName"`
	Requests         []SubscriptionBatchPatchListingRequest `json:"requests"`
	UpdateMask       string                                 `json:"updateMask"`
	RegionsVersion   string                                 `json:"regionsVersion"`
	LatencyTolerance ProductUpdateLatencyTolerance          `json:"latencyTolerance"`
	Confirm          bool                                   `json:"confirm"`
	Steps            []string                               `json:"steps"`
}

type SubscriptionPatchResult struct {
	PackageName  PackageName           `json:"packageName"`
	ProductID    SubscriptionProductID `json:"productId"`
	DryRun       bool                  `json:"dryRun"`
	Applied      bool                  `json:"applied"`
	Desired      Subscription          `json:"desiredSubscription"`
	Subscription *Subscription         `json:"subscription,omitempty"`
	Plan         SubscriptionPatchPlan `json:"plan"`
}

type SubscriptionBatchPatchListingsResult struct {
	PackageName   PackageName                        `json:"packageName"`
	DryRun        bool                               `json:"dryRun"`
	Applied       bool                               `json:"applied"`
	Desired       []Subscription                     `json:"desiredSubscriptions"`
	Subscriptions []Subscription                     `json:"subscriptions,omitempty"`
	Plan          SubscriptionBatchPatchListingsPlan `json:"plan"`
}

type BasePlanStateUpdater interface {
	UpdateBasePlanState(ctx context.Context, options BasePlanStateUpdateOptions) (Subscription, error)
}

type BasePlanBatchStateUpdater interface {
	BatchUpdateBasePlanStates(ctx context.Context, options BasePlanBatchStateUpdateOptions) (BasePlanBatchStateUpdateResult, error)
}

type BasePlanBatchPriceMigrator interface {
	BatchMigrateBasePlanPrices(ctx context.Context, options BasePlanBatchPriceMigrationOptions) (BasePlanBatchPriceMigrationResult, error)
}

type BasePlanBatchPricePatcher interface {
	BatchPatchBasePlanPrices(ctx context.Context, options BasePlanBatchPatchPriceOptions) (BasePlanBatchPatchPriceResult, error)
}

type SubscriptionPatcher interface {
	PatchSubscription(ctx context.Context, options SubscriptionPatchOptions) (Subscription, error)
}

type SubscriptionBatchListingsPatcher interface {
	BatchPatchSubscriptionListings(ctx context.Context, options SubscriptionBatchPatchListingsOptions) (SubscriptionBatchPatchListingsResult, error)
}

func NewSubscriptionPatchPlan(options SubscriptionPatchOptions) (SubscriptionPatchPlan, error) {
	if err := options.Validate(); err != nil {
		return SubscriptionPatchPlan{}, err
	}
	return SubscriptionPatchPlan{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		Listing:          options.Listing,
		DescriptionSet:   options.DescriptionSet,
		BenefitsSet:      options.BenefitsSet,
		UpdateMask:       subscriptionPatchUpdateMask,
		RegionsVersion:   options.RegionsVersion,
		LatencyTolerance: options.LatencyTolerance,
		Confirm:          options.Confirm,
		Steps:            subscriptionPatchSteps(options),
	}, nil
}

func NewSubscriptionBatchPatchListingsPlan(options SubscriptionBatchPatchListingsOptions) (SubscriptionBatchPatchListingsPlan, error) {
	if err := options.Validate(); err != nil {
		return SubscriptionBatchPatchListingsPlan{}, err
	}
	return SubscriptionBatchPatchListingsPlan{
		PackageName:      options.PackageName,
		Requests:         options.Requests,
		UpdateMask:       subscriptionPatchUpdateMask,
		RegionsVersion:   options.RegionsVersion,
		LatencyTolerance: options.LatencyTolerance,
		Confirm:          options.Confirm,
		Steps:            subscriptionBatchPatchListingsSteps(options),
	}, nil
}

func PatchSubscription(ctx context.Context, patcher SubscriptionPatcher, options SubscriptionPatchOptions) (SubscriptionPatchResult, error) {
	plan, err := NewSubscriptionPatchPlan(options)
	if err != nil {
		return SubscriptionPatchResult{}, err
	}
	result := SubscriptionPatchResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		DryRun:      options.DryRun,
		Applied:     false,
		Desired: Subscription{
			PackageName: options.PackageName,
			ProductID:   options.ProductID,
			Listings:    []SubscriptionListing{options.Listing},
			BasePlans:   []SubscriptionBasePlan{},
		},
		Plan: plan,
	}
	if options.DryRun {
		return result, nil
	}
	if patcher == nil {
		return SubscriptionPatchResult{}, fmt.Errorf("subscription patcher is required")
	}
	subscription, err := patcher.PatchSubscription(ctx, options)
	if err != nil {
		return SubscriptionPatchResult{}, err
	}
	result.Applied = true
	result.Subscription = &subscription
	return result, nil
}

func BatchPatchSubscriptionListings(ctx context.Context, patcher SubscriptionBatchListingsPatcher, options SubscriptionBatchPatchListingsOptions) (SubscriptionBatchPatchListingsResult, error) {
	plan, err := NewSubscriptionBatchPatchListingsPlan(options)
	if err != nil {
		return SubscriptionBatchPatchListingsResult{}, err
	}
	result := SubscriptionBatchPatchListingsResult{
		PackageName: options.PackageName,
		DryRun:      options.DryRun,
		Applied:     false,
		Desired:     desiredSubscriptionsForBatchListingPatch(options),
		Plan:        plan,
	}
	if options.DryRun {
		return result, nil
	}
	if patcher == nil {
		return SubscriptionBatchPatchListingsResult{}, fmt.Errorf("subscription listing batch patcher is required")
	}
	liveResult, err := patcher.BatchPatchSubscriptionListings(ctx, options)
	if err != nil {
		return SubscriptionBatchPatchListingsResult{}, err
	}
	liveResult.PackageName = options.PackageName
	liveResult.DryRun = false
	liveResult.Applied = true
	liveResult.Desired = result.Desired
	liveResult.Plan = plan
	return liveResult, nil
}

func desiredSubscriptionsForBatchListingPatch(options SubscriptionBatchPatchListingsOptions) []Subscription {
	byProduct := map[SubscriptionProductID]int{}
	subscriptions := make([]Subscription, 0)
	for _, request := range options.Requests {
		index, ok := byProduct[request.ProductID]
		if !ok {
			byProduct[request.ProductID] = len(subscriptions)
			subscriptions = append(subscriptions, Subscription{
				PackageName: options.PackageName,
				ProductID:   request.ProductID,
				Listings:    []SubscriptionListing{},
				BasePlans:   []SubscriptionBasePlan{},
			})
			index = len(subscriptions) - 1
		}
		subscriptions[index].Listings = append(subscriptions[index].Listings, request.Listing)
	}
	return subscriptions
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

func NewBasePlanBatchStateUpdatePlan(options BasePlanBatchStateUpdateOptions) (BasePlanBatchStateUpdatePlan, error) {
	if err := options.Validate(); err != nil {
		return BasePlanBatchStateUpdatePlan{}, err
	}
	return BasePlanBatchStateUpdatePlan{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		Requests:         options.Requests,
		Action:           options.Action,
		LatencyTolerance: options.LatencyTolerance,
		Confirm:          options.Confirm,
		Steps:            basePlanBatchStateUpdateSteps(options),
	}, nil
}

func BatchUpdateBasePlanStates(ctx context.Context, updater BasePlanBatchStateUpdater, options BasePlanBatchStateUpdateOptions) (BasePlanBatchStateUpdateResult, error) {
	plan, err := NewBasePlanBatchStateUpdatePlan(options)
	if err != nil {
		return BasePlanBatchStateUpdateResult{}, err
	}
	result := BasePlanBatchStateUpdateResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		Requests:    options.Requests,
		Action:      options.Action,
		DryRun:      options.DryRun,
		Applied:     false,
		Plan:        plan,
	}
	if options.DryRun {
		return result, nil
	}
	if updater == nil {
		return BasePlanBatchStateUpdateResult{}, fmt.Errorf("base plan batch state updater is required")
	}
	liveResult, err := updater.BatchUpdateBasePlanStates(ctx, options)
	if err != nil {
		return BasePlanBatchStateUpdateResult{}, err
	}
	liveResult.PackageName = options.PackageName
	liveResult.ProductID = options.ProductID
	liveResult.Requests = options.Requests
	liveResult.Action = options.Action
	liveResult.DryRun = false
	liveResult.Applied = true
	liveResult.Plan = plan
	return liveResult, nil
}

func NewBasePlanBatchPriceMigrationPlan(options BasePlanBatchPriceMigrationOptions) (BasePlanBatchPriceMigrationPlan, error) {
	if err := options.Validate(); err != nil {
		return BasePlanBatchPriceMigrationPlan{}, err
	}
	return BasePlanBatchPriceMigrationPlan{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		RegionsVersion:   options.RegionsVersion,
		Requests:         options.Requests,
		LatencyTolerance: options.LatencyTolerance,
		Confirm:          options.Confirm,
		Steps:            basePlanBatchPriceMigrationSteps(options),
	}, nil
}

func BatchMigrateBasePlanPrices(ctx context.Context, migrator BasePlanBatchPriceMigrator, options BasePlanBatchPriceMigrationOptions) (BasePlanBatchPriceMigrationResult, error) {
	plan, err := NewBasePlanBatchPriceMigrationPlan(options)
	if err != nil {
		return BasePlanBatchPriceMigrationResult{}, err
	}
	result := BasePlanBatchPriceMigrationResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		DryRun:      options.DryRun,
		Applied:     false,
		Plan:        plan,
	}
	if options.DryRun {
		return result, nil
	}
	if migrator == nil {
		return BasePlanBatchPriceMigrationResult{}, fmt.Errorf("base plan price migrator is required")
	}
	liveResult, err := migrator.BatchMigrateBasePlanPrices(ctx, options)
	if err != nil {
		return BasePlanBatchPriceMigrationResult{}, err
	}
	liveResult.PackageName = options.PackageName
	liveResult.ProductID = options.ProductID
	liveResult.DryRun = false
	liveResult.Applied = true
	liveResult.Plan = plan
	return liveResult, nil
}

func NewBasePlanBatchPatchPricePlan(options BasePlanBatchPatchPriceOptions) (BasePlanBatchPatchPricePlan, error) {
	if err := options.Validate(); err != nil {
		return BasePlanBatchPatchPricePlan{}, err
	}
	return BasePlanBatchPatchPricePlan{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		RegionsVersion:   options.RegionsVersion,
		Requests:         append([]BasePlanPricePatchRequest(nil), options.Requests...),
		UpdateMask:       subscriptionBasePlanPatchUpdateMask,
		LatencyTolerance: options.LatencyTolerance,
		Confirm:          options.Confirm,
		Steps:            basePlanBatchPatchPriceSteps(options),
	}, nil
}

func BatchPatchBasePlanPrices(ctx context.Context, patcher BasePlanBatchPricePatcher, options BasePlanBatchPatchPriceOptions) (BasePlanBatchPatchPriceResult, error) {
	plan, err := NewBasePlanBatchPatchPricePlan(options)
	if err != nil {
		return BasePlanBatchPatchPriceResult{}, err
	}
	requests := append([]BasePlanPricePatchRequest(nil), options.Requests...)
	result := BasePlanBatchPatchPriceResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		Requests:    requests,
		DryRun:      options.DryRun,
		Desired:     desiredSubscriptionsForBasePlanPricePatch(options),
		Plan:        plan,
	}
	if options.DryRun {
		return result, nil
	}
	if patcher == nil {
		return BasePlanBatchPatchPriceResult{}, fmt.Errorf("base plan price batch patcher is required")
	}
	updated, err := patcher.BatchPatchBasePlanPrices(ctx, options)
	if err != nil {
		return BasePlanBatchPatchPriceResult{}, err
	}
	updated.PackageName = options.PackageName
	updated.ProductID = options.ProductID
	updated.Requests = requests
	updated.DryRun = false
	updated.Applied = true
	updated.Desired = result.Desired
	updated.Plan = plan
	return updated, nil
}

func desiredSubscriptionsForBasePlanPricePatch(options BasePlanBatchPatchPriceOptions) []BasePlanBatchPatchPriceDesiredSubscription {
	byProduct := map[SubscriptionProductID]int{}
	byBasePlan := map[string]int{}
	subscriptions := make([]BasePlanBatchPatchPriceDesiredSubscription, 0)
	for _, request := range options.Requests {
		productIndex, ok := byProduct[request.ProductID]
		if !ok {
			byProduct[request.ProductID] = len(subscriptions)
			subscriptions = append(subscriptions, BasePlanBatchPatchPriceDesiredSubscription{
				PackageName: options.PackageName,
				ProductID:   request.ProductID,
				BasePlans:   []BasePlanBatchPatchPriceDesiredPlan{},
			})
			productIndex = len(subscriptions) - 1
		}
		basePlanKey := request.ProductID.String() + "/" + request.BasePlanID.String()
		basePlanIndex, ok := byBasePlan[basePlanKey]
		if !ok {
			byBasePlan[basePlanKey] = len(subscriptions[productIndex].BasePlans)
			subscriptions[productIndex].BasePlans = append(subscriptions[productIndex].BasePlans, BasePlanBatchPatchPriceDesiredPlan{
				BasePlanID:      request.BasePlanID,
				RegionalConfigs: []BasePlanBatchPatchPriceDesiredRegion{},
			})
			basePlanIndex = len(subscriptions[productIndex].BasePlans) - 1
		}
		subscriptions[productIndex].BasePlans[basePlanIndex].RegionalConfigs = append(subscriptions[productIndex].BasePlans[basePlanIndex].RegionalConfigs, BasePlanBatchPatchPriceDesiredRegion{
			RegionCode: request.RegionCode,
			Price:      request.Price,
		})
	}
	return subscriptions
}

func basePlanStateUpdateSteps(options BasePlanStateUpdateOptions) []string {
	if options.DryRun {
		return []string{fmt.Sprintf("plan %s base plan", options.Action)}
	}
	return []string{fmt.Sprintf("%s base plan", options.Action)}
}

func basePlanBatchStateUpdateSteps(options BasePlanBatchStateUpdateOptions) []string {
	if options.DryRun {
		return []string{fmt.Sprintf("plan batch %s base plans", options.Action)}
	}
	return []string{fmt.Sprintf("batch %s base plans", options.Action)}
}

func basePlanBatchPriceMigrationSteps(options BasePlanBatchPriceMigrationOptions) []string {
	if options.DryRun {
		return []string{"plan batch base plan price migration"}
	}
	return []string{"batch migrate base plan prices"}
}

func basePlanBatchPatchPriceSteps(options BasePlanBatchPatchPriceOptions) []string {
	if options.DryRun {
		return []string{"plan base plan price batch patch"}
	}
	return []string{"fetch current subscriptions", "merge base plan regional prices", "batch patch base plan prices"}
}

func subscriptionCreateSteps(dryRun bool) []string {
	if dryRun {
		return []string{"plan subscription create"}
	}
	return []string{"create subscription"}
}

const subscriptionPatchUpdateMask = "listings"
const subscriptionBasePlanPatchUpdateMask = "basePlans"

func subscriptionPatchSteps(options SubscriptionPatchOptions) []string {
	if options.DryRun {
		return []string{"plan subscription listing patch"}
	}
	return []string{"fetch current subscription", "merge localized listing", "patch subscription listings"}
}

func subscriptionBatchPatchListingsSteps(options SubscriptionBatchPatchListingsOptions) []string {
	if options.DryRun {
		return []string{"plan subscription listing batch patch"}
	}
	return []string{"fetch current subscriptions", "merge localized listings", "batch patch subscription listings"}
}
