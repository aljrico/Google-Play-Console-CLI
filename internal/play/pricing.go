package play

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

type CurrencyCode string

func NewCurrencyCode(value string) (CurrencyCode, error) {
	value = strings.ToUpper(strings.TrimSpace(value))
	if len(value) != 3 {
		return "", fmt.Errorf("currency code must be a 3-letter ISO 4217 code")
	}
	for i := 0; i < len(value); i++ {
		if !isASCIIAlpha(value[i]) {
			return "", fmt.Errorf("currency code must contain only ASCII letters")
		}
	}
	return CurrencyCode(value), nil
}

func (c CurrencyCode) String() string {
	return string(c)
}

func (c CurrencyCode) Validate() error {
	normalized, err := NewCurrencyCode(c.String())
	if err != nil {
		return err
	}
	if normalized != c {
		return fmt.Errorf("currency code must be normalized uppercase without surrounding whitespace")
	}
	return nil
}

type RegionPriceConversionOptions struct {
	PackageName PackageName  `json:"packageName"`
	Currency    CurrencyCode `json:"currency"`
	Units       int64        `json:"units"`
	Nanos       int64        `json:"nanos,omitempty"`
}

func (o RegionPriceConversionOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if err := o.Currency.Validate(); err != nil {
		return err
	}
	if o.Units < 0 {
		return fmt.Errorf("price units must be 0 or greater")
	}
	if o.Nanos < 0 || o.Nanos > 999999999 {
		return fmt.Errorf("price nanos must be between 0 and 999999999")
	}
	if o.Units == 0 && o.Nanos == 0 {
		return fmt.Errorf("price must be greater than 0")
	}
	return nil
}

type ConvertedOtherRegionsPrice struct {
	USDPrice *Money `json:"usdPrice,omitempty"`
	EURPrice *Money `json:"eurPrice,omitempty"`
}

type ConvertedRegionPrice struct {
	RegionCode string `json:"regionCode,omitempty"`
	Price      *Money `json:"price,omitempty"`
	TaxAmount  *Money `json:"taxAmount,omitempty"`
}

type RegionPriceConversionResult struct {
	PackageName                PackageName                     `json:"packageName"`
	SourcePrice                Money                           `json:"sourcePrice"`
	RegionVersion              string                          `json:"regionVersion,omitempty"`
	ConvertedOtherRegionsPrice *ConvertedOtherRegionsPrice     `json:"convertedOtherRegionsPrice,omitempty"`
	ConvertedRegionPrices      map[string]ConvertedRegionPrice `json:"convertedRegionPrices"`
}

type PricePatchBuildTarget string

const (
	PricePatchBuildTargetInAppProduct           PricePatchBuildTarget = "in-app-product"
	PricePatchBuildTargetOneTimeProduct         PricePatchBuildTarget = "one-time-product"
	PricePatchBuildTargetSubscriptionBasePlan   PricePatchBuildTarget = "subscription-base-plan"
	PricePatchBuildTargetSubscriptionOfferPhase PricePatchBuildTarget = "subscription-offer-phase"
)

type PricePatchBuildOptions struct {
	Conversion       RegionPriceConversionResult    `json:"conversion"`
	Target           PricePatchBuildTarget          `json:"target"`
	SKU              InAppProductSKU                `json:"sku,omitempty"`
	ProductID        string                         `json:"productId,omitempty"`
	PurchaseOptionID OneTimeProductPurchaseOptionID `json:"purchaseOptionId,omitempty"`
	BasePlanID       SubscriptionBasePlanID         `json:"basePlanId,omitempty"`
	OfferID          SubscriptionOfferID            `json:"offerId,omitempty"`
	PhaseIndex       int                            `json:"phaseIndex,omitempty"`
}

type PricePatchBuildResult struct {
	PackageName      PackageName                    `json:"packageName"`
	SourcePrice      Money                          `json:"sourcePrice"`
	RegionVersion    string                         `json:"regionVersion,omitempty"`
	Target           PricePatchBuildTarget          `json:"target"`
	SKU              InAppProductSKU                `json:"sku,omitempty"`
	ProductID        string                         `json:"productId,omitempty"`
	PurchaseOptionID OneTimeProductPurchaseOptionID `json:"purchaseOptionId,omitempty"`
	BasePlanID       SubscriptionBasePlanID         `json:"basePlanId,omitempty"`
	OfferID          SubscriptionOfferID            `json:"offerId,omitempty"`
	PhaseIndex       *int                           `json:"phaseIndex,omitempty"`
	PriceArgs        []string                       `json:"priceArgs"`
	SuggestedCommand []string                       `json:"suggestedCommand"`
}

type RegionPriceConverter interface {
	ConvertRegionPrices(ctx context.Context, options RegionPriceConversionOptions) (RegionPriceConversionResult, error)
}

func ConvertRegionPrices(ctx context.Context, converter RegionPriceConverter, options RegionPriceConversionOptions) (RegionPriceConversionResult, error) {
	if err := options.Validate(); err != nil {
		return RegionPriceConversionResult{}, err
	}
	if converter == nil {
		return RegionPriceConversionResult{}, fmt.Errorf("region price converter is required")
	}
	return converter.ConvertRegionPrices(ctx, options)
}

func BuildPricePatches(options PricePatchBuildOptions) (PricePatchBuildResult, error) {
	if err := options.Validate(); err != nil {
		return PricePatchBuildResult{}, err
	}
	priceArgs, err := buildPricePatchArgs(options)
	if err != nil {
		return PricePatchBuildResult{}, err
	}
	result := PricePatchBuildResult{
		PackageName:      options.Conversion.PackageName,
		SourcePrice:      options.Conversion.SourcePrice,
		RegionVersion:    options.Conversion.RegionVersion,
		Target:           options.Target,
		SKU:              options.SKU,
		ProductID:        options.ProductID,
		PurchaseOptionID: options.PurchaseOptionID,
		BasePlanID:       options.BasePlanID,
		OfferID:          options.OfferID,
		PriceArgs:        priceArgs,
	}
	if options.Target == PricePatchBuildTargetSubscriptionOfferPhase {
		phaseIndex := options.PhaseIndex
		result.PhaseIndex = &phaseIndex
	}
	result.SuggestedCommand = buildPricePatchSuggestedCommand(result)
	return result, nil
}

func (o PricePatchBuildOptions) Validate() error {
	if err := o.Conversion.PackageName.Validate(); err != nil {
		return err
	}
	if err := validateMoney(o.Conversion.SourcePrice); err != nil {
		return fmt.Errorf("source price: %w", err)
	}
	if strings.TrimSpace(o.Conversion.RegionVersion) != o.Conversion.RegionVersion {
		return fmt.Errorf("region version cannot have leading or trailing whitespace")
	}
	if len(o.Conversion.ConvertedRegionPrices) == 0 {
		return fmt.Errorf("converted region prices are required")
	}
	seenRegions := map[string]struct{}{}
	for key, convertedPrice := range o.Conversion.ConvertedRegionPrices {
		regionCode := convertedRegionCode(key, convertedPrice)
		if !isValidRegionCode(regionCode) {
			return fmt.Errorf("converted region price %q must use a two-letter ISO 3166 region code", key)
		}
		if convertedPrice.RegionCode != "" && convertedPrice.RegionCode != key {
			return fmt.Errorf("converted region price %q regionCode %s does not match map key", key, convertedPrice.RegionCode)
		}
		if _, ok := seenRegions[regionCode]; ok {
			return fmt.Errorf("converted region price %s is duplicated", regionCode)
		}
		seenRegions[regionCode] = struct{}{}
		if convertedPrice.Price == nil {
			return fmt.Errorf("converted region price %s requires price", regionCode)
		}
		if err := validateMoney(*convertedPrice.Price); err != nil {
			return fmt.Errorf("converted region price %s: %w", regionCode, err)
		}
	}
	switch o.Target {
	case PricePatchBuildTargetInAppProduct:
		if _, err := NewInAppProductSKU(o.SKU.String()); err != nil {
			return err
		}
	case PricePatchBuildTargetOneTimeProduct:
		if o.Conversion.RegionVersion == "" {
			return fmt.Errorf("region version is required for one-time product price patches")
		}
		productID, err := NewOneTimeProductID(o.ProductID)
		if err != nil {
			return err
		}
		if productID.String() != o.ProductID {
			return fmt.Errorf("one-time product ID must be normalized")
		}
		if _, err := NewOneTimeProductPurchaseOptionID(o.PurchaseOptionID.String()); err != nil {
			return err
		}
	case PricePatchBuildTargetSubscriptionBasePlan:
		if o.Conversion.RegionVersion == "" {
			return fmt.Errorf("region version is required for subscription base plan price patches")
		}
		productID, err := NewSubscriptionProductID(o.ProductID)
		if err != nil {
			return err
		}
		if productID.String() != o.ProductID {
			return fmt.Errorf("subscription product ID must be normalized")
		}
		if _, err := NewSubscriptionBasePlanID(o.BasePlanID.String()); err != nil {
			return err
		}
	case PricePatchBuildTargetSubscriptionOfferPhase:
		if o.Conversion.RegionVersion == "" {
			return fmt.Errorf("region version is required for subscription offer phase price patches")
		}
		productID, err := NewSubscriptionProductID(o.ProductID)
		if err != nil {
			return err
		}
		if productID.String() != o.ProductID {
			return fmt.Errorf("subscription product ID must be normalized")
		}
		if _, err := NewSubscriptionBasePlanID(o.BasePlanID.String()); err != nil {
			return err
		}
		if _, err := NewSubscriptionOfferID(o.OfferID.String()); err != nil {
			return err
		}
		if o.PhaseIndex < 0 {
			return fmt.Errorf("phase index must be 0 or greater")
		}
	default:
		return fmt.Errorf("unsupported price patch target %q", o.Target)
	}
	return nil
}

func buildPricePatchArgs(options PricePatchBuildOptions) ([]string, error) {
	convertedPrices := sortedConvertedRegionPrices(options.Conversion.ConvertedRegionPrices)
	priceArgs := make([]string, 0, len(convertedPrices))
	for _, convertedPrice := range convertedPrices {
		regionCode := convertedPrice.RegionCode
		price := *convertedPrice.Price
		priceValue, err := formatPricePatchValue(options.Target, price)
		if err != nil {
			return nil, fmt.Errorf("converted region price %s: %w", regionCode, err)
		}
		switch options.Target {
		case PricePatchBuildTargetInAppProduct:
			priceArgs = append(priceArgs, fmt.Sprintf("%s:%s", regionCode, priceValue))
		case PricePatchBuildTargetOneTimeProduct:
			priceArgs = append(priceArgs, fmt.Sprintf("%s/%s/%s:%s", options.ProductID, options.PurchaseOptionID, regionCode, priceValue))
		case PricePatchBuildTargetSubscriptionBasePlan:
			priceArgs = append(priceArgs, fmt.Sprintf("%s/%s/%s:%s", options.ProductID, options.BasePlanID, regionCode, priceValue))
		case PricePatchBuildTargetSubscriptionOfferPhase:
			priceArgs = append(priceArgs, fmt.Sprintf("%s/%s/%s/%d/%s:%s", options.ProductID, options.BasePlanID, options.OfferID, options.PhaseIndex, regionCode, priceValue))
		}
	}
	return priceArgs, nil
}

func buildPricePatchSuggestedCommand(result PricePatchBuildResult) []string {
	switch result.Target {
	case PricePatchBuildTargetInAppProduct:
		command := []string{"gpc", "in-app-products", "patch", "--package", result.PackageName.String(), "--sku", result.SKU.String()}
		for _, priceArg := range result.PriceArgs {
			command = append(command, "--regional-price", priceArg)
		}
		return append(command, "--dry-run")
	case PricePatchBuildTargetOneTimeProduct:
		command := []string{"gpc", "one-time-products", "purchase-option", "batch-patch-prices", "--package", result.PackageName.String()}
		if result.RegionVersion != "" {
			command = append(command, "--regions-version", result.RegionVersion)
		}
		for _, priceArg := range result.PriceArgs {
			command = append(command, "--price", priceArg)
		}
		return append(command, "--dry-run")
	case PricePatchBuildTargetSubscriptionBasePlan:
		command := []string{"gpc", "subscriptions", "base-plan", "batch-patch-prices", "--package", result.PackageName.String()}
		if result.RegionVersion != "" {
			command = append(command, "--regions-version", result.RegionVersion)
		}
		for _, priceArg := range result.PriceArgs {
			command = append(command, "--price", priceArg)
		}
		return append(command, "--dry-run")
	case PricePatchBuildTargetSubscriptionOfferPhase:
		command := []string{"gpc", "subscription-offers", "batch-patch-phase-prices", "--package", result.PackageName.String()}
		if result.RegionVersion != "" {
			command = append(command, "--regions-version", result.RegionVersion)
		}
		for _, priceArg := range result.PriceArgs {
			command = append(command, "--price", priceArg)
		}
		return append(command, "--dry-run")
	default:
		return nil
	}
}

func sortedConvertedRegionPrices(prices map[string]ConvertedRegionPrice) []ConvertedRegionPrice {
	convertedPrices := make([]ConvertedRegionPrice, 0, len(prices))
	for key, convertedPrice := range prices {
		convertedPrice.RegionCode = convertedRegionCode(key, convertedPrice)
		convertedPrices = append(convertedPrices, convertedPrice)
	}
	sort.Slice(convertedPrices, func(i, j int) bool {
		return convertedPrices[i].RegionCode < convertedPrices[j].RegionCode
	})
	return convertedPrices
}

func convertedRegionCode(key string, price ConvertedRegionPrice) string {
	if price.RegionCode != "" {
		return price.RegionCode
	}
	return key
}

func formatPricePatchValue(target PricePatchBuildTarget, price Money) (string, error) {
	if target == PricePatchBuildTargetInAppProduct {
		micros, err := priceMicros(price)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s:%d", price.CurrencyCode, micros), nil
	}
	if price.Nanos == 0 {
		return fmt.Sprintf("%s:%d", price.CurrencyCode, price.Units), nil
	}
	return fmt.Sprintf("%s:%d:%d", price.CurrencyCode, price.Units, price.Nanos), nil
}

func priceMicros(price Money) (int64, error) {
	if price.Nanos%1000 != 0 {
		return 0, fmt.Errorf("price nanos must be divisible by 1000 to build legacy micros")
	}
	const maxInt64 = int64(1<<63 - 1)
	microsNanos := price.Nanos / 1000
	if price.Units > (maxInt64-microsNanos)/1000000 {
		return 0, fmt.Errorf("price micros overflow")
	}
	micros := price.Units*1000000 + microsNanos
	if micros <= 0 {
		return 0, fmt.Errorf("price micros must be greater than 0")
	}
	return micros, nil
}
