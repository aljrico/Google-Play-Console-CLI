package play

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"google.golang.org/api/androidpublisher/v3"
)

func DecodeSubscriptionOfferCreateJSON(data []byte) (SubscriptionOffer, error) {
	if err := validateSubscriptionOfferCreateJSONKeys(data); err != nil {
		return SubscriptionOffer{}, err
	}

	var apiOffer androidpublisher.SubscriptionOffer
	if err := decodeStrictJSON(data, &apiOffer); err == nil {
		return subscriptionOfferFromAPI(&apiOffer), nil
	}

	var offer SubscriptionOffer
	if err := decodeStrictJSON(data, &offer); err == nil {
		return offer, nil
	}

	return SubscriptionOffer{}, fmt.Errorf("subscription offer JSON must use Google Play API JSON or playpub subscription offer JSON")
}

func decodeStrictJSON(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("JSON contains multiple values")
	}
	return nil
}

func validateSubscriptionOfferCreateJSONKeys(data []byte) error {
	var value any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("JSON contains multiple values")
	}
	object, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("subscription offer JSON must be an object")
	}
	return validateJSONObjectKeys("subscription offer", object, subscriptionOfferJSONSchema())
}

type jsonObjectSchema map[string]jsonSchemaValue

type jsonSchemaValue struct {
	Object jsonObjectSchema
	Array  *jsonSchemaValue
}

func validateJSONObjectKeys(path string, object map[string]any, schema jsonObjectSchema) error {
	for key, value := range object {
		childSchema, ok := schema[key]
		if !ok {
			return fmt.Errorf("%s contains unknown field %q", path, key)
		}
		if err := validateJSONValueKeys(path+"."+key, value, childSchema); err != nil {
			return err
		}
	}
	return nil
}

func validateJSONValueKeys(path string, value any, schema jsonSchemaValue) error {
	switch {
	case schema.Object != nil:
		object, ok := value.(map[string]any)
		if !ok {
			return nil
		}
		return validateJSONObjectKeys(path, object, schema.Object)
	case schema.Array != nil:
		items, ok := value.([]any)
		if !ok {
			return nil
		}
		for index, item := range items {
			if err := validateJSONValueKeys(fmt.Sprintf("%s[%d]", path, index), item, *schema.Array); err != nil {
				return err
			}
		}
	}
	return nil
}

func subscriptionOfferJSONSchema() jsonObjectSchema {
	emptyObject := jsonObjectSchema{}
	money := jsonObjectSchema{
		"currencyCode": {},
		"units":        {},
		"nanos":        {},
	}
	otherRegionsPrices := jsonObjectSchema{
		"usdPrice": {Object: money},
		"eurPrice": {Object: money},
	}
	targetingScope := jsonObjectSchema{
		"anySubscriptionInApp":      {Object: emptyObject},
		"thisSubscription":          {Object: emptyObject},
		"specificSubscriptionInApp": {},
	}
	return jsonObjectSchema{
		"packageName": {},
		"productId":   {},
		"basePlanId":  {},
		"offerId":     {},
		"state":       {},
		"offerTags": {Array: &jsonSchemaValue{Object: jsonObjectSchema{
			"tag": {},
		}}},
		"regionalConfigs": {Array: &jsonSchemaValue{Object: jsonObjectSchema{
			"regionCode":                {},
			"newSubscriberAvailability": {},
		}}},
		"otherRegionsConfig": {Object: jsonObjectSchema{
			"newSubscriberAvailability":             {},
			"otherRegionsNewSubscriberAvailability": {},
		}},
		"phases": {Array: &jsonSchemaValue{Object: jsonObjectSchema{
			"duration":        {},
			"recurrenceCount": {},
			"regionalConfigs": {Array: &jsonSchemaValue{Object: jsonObjectSchema{
				"regionCode":       {},
				"price":            {Object: money},
				"absoluteDiscount": {Object: money},
				"relativeDiscount": {},
				"free":             {Object: emptyObject},
			}}},
			"otherRegionsConfig": {Object: jsonObjectSchema{
				"otherRegionsPrices": {Object: otherRegionsPrices},
				"absoluteDiscounts":  {Object: otherRegionsPrices},
				"relativeDiscount":   {},
				"free":               {Object: emptyObject},
			}},
		}}},
		"targeting": {Object: jsonObjectSchema{
			"acquisition": {Object: jsonObjectSchema{
				"scope": {Object: targetingScope},
			}},
			"acquisitionRule": {Object: jsonObjectSchema{
				"scope": {Object: targetingScope},
			}},
			"upgrade": {Object: jsonObjectSchema{
				"scope":                 {Object: targetingScope},
				"billingPeriodDuration": {},
				"oncePerUser":           {},
			}},
			"upgradeRule": {Object: jsonObjectSchema{
				"scope":                 {Object: targetingScope},
				"billingPeriodDuration": {},
				"oncePerUser":           {},
			}},
		}},
	}
}
