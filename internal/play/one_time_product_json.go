package play

import (
	"encoding/json"
	"fmt"

	"google.golang.org/api/androidpublisher/v3"
)

func DecodeOneTimeProductCreateJSON(data []byte) (OneTimeProduct, error) {
	if err := validateOneTimeProductCreateJSONKeys(data); err != nil {
		return OneTimeProduct{}, err
	}

	var apiProduct androidpublisher.OneTimeProduct
	if err := decodeStrictJSON(data, &apiProduct); err == nil {
		return oneTimeProductFromGeneratedAPI(&apiProduct)
	}

	var product OneTimeProduct
	if err := decodeStrictJSON(data, &product); err == nil {
		return product, nil
	}

	return OneTimeProduct{}, fmt.Errorf("one-time product JSON must use Google Play API JSON or gpc one-time product JSON")
}

func validateOneTimeProductCreateJSONKeys(data []byte) error {
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	object, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("one-time product JSON must be an object")
	}
	if err := validateOneTimeProductCreateJSONUnions(object); err != nil {
		return err
	}
	if hasNestedJSONKey(object, "regionalProductAgeRatingInfos") {
		return fmt.Errorf("one-time product regionalProductAgeRatingInfos is not supported by the pinned Google API client; use a newer gpc build before setting regional age ratings")
	}
	return validateJSONObjectKeys("one-time product", object, oneTimeProductJSONSchema())
}

func validateOneTimeProductCreateJSONUnions(object map[string]any) error {
	rawOptions, ok := object["purchaseOptions"].([]any)
	if !ok {
		return nil
	}
	for index, rawOption := range rawOptions {
		option, ok := rawOption.(map[string]any)
		if !ok {
			continue
		}
		types := 0
		for _, field := range []string{"buyOption", "rentOption"} {
			if _, ok := option[field]; ok {
				types++
			}
		}
		if types > 1 {
			return fmt.Errorf("one-time product.purchaseOptions[%d] must set only one purchase option type object", index)
		}
	}
	return nil
}

func hasNestedJSONKey(value any, key string) bool {
	switch typed := value.(type) {
	case map[string]any:
		for childKey, childValue := range typed {
			if childKey == key || hasNestedJSONKey(childValue, key) {
				return true
			}
		}
	case []any:
		for _, childValue := range typed {
			if hasNestedJSONKey(childValue, key) {
				return true
			}
		}
	}
	return false
}

func oneTimeProductJSONSchema() jsonObjectSchema {
	money := jsonObjectSchema{"currencyCode": {}, "units": {}, "nanos": {}}
	regionalTaxConfig := jsonObjectSchema{
		"regionCode":                         {},
		"eligibleForStreamingServiceTaxRate": {},
		"streamingTaxType":                   {},
		"taxTier":                            {},
	}
	return jsonObjectSchema{
		"packageName": {},
		"productId":   {},
		"listings": {Array: &jsonSchemaValue{Object: jsonObjectSchema{
			"languageCode": {},
			"title":        {},
			"description":  {},
		}}},
		"offerTags": {Array: &jsonSchemaValue{Object: jsonObjectSchema{
			"tag": {},
		}}},
		"purchaseOptions": {Array: &jsonSchemaValue{Object: jsonObjectSchema{
			"purchaseOptionId":     {},
			"state":                {},
			"type":                 {},
			"buyOption":            {Object: jsonObjectSchema{"legacyCompatible": {}, "multiQuantityEnabled": {}}},
			"rentOption":           {Object: jsonObjectSchema{"rentalPeriod": {}, "expirationPeriod": {}}},
			"legacyCompatible":     {},
			"multiQuantityEnabled": {},
			"rentalPeriod":         {},
			"expirationPeriod":     {},
			"offerTags": {Array: &jsonSchemaValue{Object: jsonObjectSchema{
				"tag": {},
			}}},
			"regionalConfigs": {Array: &jsonSchemaValue{Object: jsonObjectSchema{
				"regionCode":   {},
				"availability": {},
				"price":        {Object: money},
			}}},
			"regionalPricingAndAvailabilityConfigs": {Array: &jsonSchemaValue{Object: jsonObjectSchema{
				"regionCode":   {},
				"availability": {},
				"price":        {Object: money},
			}}},
			"newRegionsConfig": {Object: jsonObjectSchema{
				"availability": {},
				"usdPrice":     {Object: money},
				"eurPrice":     {Object: money},
			}},
			"taxAndComplianceSettings": {Object: jsonObjectSchema{
				"withdrawalRightType": {},
			}},
		}}},
		"regionsVersion": {},
		"restrictedCountries": {Object: jsonObjectSchema{
			"regionCodes": {},
		}},
		"restrictedPaymentCountries": {Object: jsonObjectSchema{
			"regionCodes": {},
		}},
		"taxAndComplianceSettings": {Object: jsonObjectSchema{
			"isTokenizedDigitalAsset": {},
			"productTaxCategoryCode":  {},
			"regionalAgeRatings": {Array: &jsonSchemaValue{Object: jsonObjectSchema{
				"regionCode":           {},
				"productAgeRatingTier": {},
			}}},
			"regionalProductAgeRatingInfos": {Array: &jsonSchemaValue{Object: jsonObjectSchema{
				"regionCode":           {},
				"productAgeRatingTier": {},
			}}},
			"regionalTaxConfigs": {Array: &jsonSchemaValue{Object: regionalTaxConfig}},
		}},
	}
}
