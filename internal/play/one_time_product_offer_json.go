package play

import (
	"encoding/json"
	"fmt"

	"google.golang.org/api/androidpublisher/v3"
)

func DecodeOneTimeProductOfferCreateJSON(data []byte) (OneTimeProductOffer, error) {
	if err := validateOneTimeProductOfferCreateJSONKeys(data); err != nil {
		return OneTimeProductOffer{}, err
	}

	var apiOffer androidpublisher.OneTimeProductOffer
	if err := decodeStrictJSON(data, &apiOffer); err == nil {
		return oneTimeProductOfferFromAPI(&apiOffer), nil
	}

	var offer OneTimeProductOffer
	if err := decodeStrictJSON(data, &offer); err == nil {
		return offer, nil
	}

	return OneTimeProductOffer{}, fmt.Errorf("one-time product offer JSON must use Google Play API JSON or playpub one-time product offer JSON")
}

func validateOneTimeProductOfferCreateJSONKeys(data []byte) error {
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	object, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("one-time product offer JSON must be an object")
	}
	if err := validateOneTimeProductOfferCreateJSONUnions(object); err != nil {
		return err
	}
	return validateJSONObjectKeys("one-time product offer", object, oneTimeProductOfferJSONSchema())
}

func validateOneTimeProductOfferCreateJSONUnions(object map[string]any) error {
	types := 0
	for _, field := range []string{"discountedOffer", "preOrderOffer"} {
		if _, ok := object[field]; ok {
			types++
		}
	}
	if types > 1 {
		return fmt.Errorf("one-time product offer must set only one offer type object")
	}
	return nil
}

func oneTimeProductOfferJSONSchema() jsonObjectSchema {
	money := jsonObjectSchema{"currencyCode": {}, "units": {}, "nanos": {}}
	emptyObject := jsonObjectSchema{}
	return jsonObjectSchema{
		"packageName":      {},
		"productId":        {},
		"purchaseOptionId": {},
		"offerId":          {},
		"state":            {},
		"type":             {},
		"regionsVersion":   {},
		"offerTags": {Array: &jsonSchemaValue{Object: jsonObjectSchema{
			"tag": {},
		}}},
		"discountedOffer": {Object: jsonObjectSchema{
			"startTime":       {},
			"endTime":         {},
			"redemptionLimit": {},
		}},
		"preOrderOffer": {Object: jsonObjectSchema{
			"startTime":           {},
			"endTime":             {},
			"releaseTime":         {},
			"priceChangeBehavior": {},
		}},
		"regionalConfigs": {Array: &jsonSchemaValue{Object: jsonObjectSchema{
			"regionCode":       {},
			"availability":     {},
			"absoluteDiscount": {Object: money},
			"relativeDiscount": {},
			"noOverride":       {},
		}}},
		"regionalPricingAndAvailabilityConfigs": {Array: &jsonSchemaValue{Object: jsonObjectSchema{
			"regionCode":       {},
			"availability":     {},
			"absoluteDiscount": {Object: money},
			"relativeDiscount": {},
			"noOverride":       {Object: emptyObject},
		}}},
	}
}
