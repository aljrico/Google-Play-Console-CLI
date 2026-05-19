package play

import (
	"encoding/json"
	"fmt"

	"google.golang.org/api/androidpublisher/v3"
)

func DecodeSubscriptionCreateJSON(data []byte) (Subscription, error) {
	if err := validateSubscriptionCreateJSONKeys(data); err != nil {
		return Subscription{}, err
	}

	var apiSubscription androidpublisher.Subscription
	if err := decodeStrictJSON(data, &apiSubscription); err == nil {
		return subscriptionFromAPI(&apiSubscription), nil
	}

	var subscription Subscription
	if err := decodeStrictJSON(data, &subscription); err == nil {
		return subscription, nil
	}

	return Subscription{}, fmt.Errorf("subscription JSON must use Google Play API JSON or gpc subscription JSON")
}

func validateSubscriptionCreateJSONKeys(data []byte) error {
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	object, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("subscription JSON must be an object")
	}
	if err := validateSubscriptionCreateJSONUnions(object); err != nil {
		return err
	}
	return validateJSONObjectKeys("subscription", object, subscriptionJSONSchema())
}

func validateSubscriptionCreateJSONUnions(object map[string]any) error {
	rawBasePlans, ok := object["basePlans"].([]any)
	if !ok {
		return nil
	}
	for index, rawBasePlan := range rawBasePlans {
		basePlan, ok := rawBasePlan.(map[string]any)
		if !ok {
			continue
		}
		types := 0
		for _, field := range []string{"autoRenewingBasePlanType", "prepaidBasePlanType", "installmentsBasePlanType"} {
			if _, ok := basePlan[field]; ok {
				types++
			}
		}
		if types > 1 {
			return fmt.Errorf("subscription.basePlans[%d] must set only one base plan type object", index)
		}
	}
	return nil
}

func subscriptionJSONSchema() jsonObjectSchema {
	money := jsonObjectSchema{"currencyCode": {}, "units": {}, "nanos": {}}
	basePlanType := jsonObjectSchema{
		"billingPeriodDuration":               {},
		"gracePeriodDuration":                 {},
		"accountHoldDuration":                 {},
		"legacyCompatible":                    {},
		"legacyCompatibleSubscriptionOfferId": {},
		"legacyCompatibleSubscriptionOfferID": {},
		"prorationMode":                       {},
		"resubscribeState":                    {},
		"timeExtension":                       {},
		"committedPaymentsCount":              {},
		"renewalType":                         {},
	}
	return jsonObjectSchema{
		"packageName": {},
		"productId":   {},
		"archived":    {},
		"listings": {Array: &jsonSchemaValue{Object: jsonObjectSchema{
			"languageCode": {},
			"title":        {},
			"description":  {},
			"benefits":     {},
		}}},
		"basePlans": {Array: &jsonSchemaValue{Object: jsonObjectSchema{
			"basePlanId":                          {},
			"state":                               {},
			"type":                                {},
			"autoRenewingBasePlanType":            {Object: basePlanType},
			"prepaidBasePlanType":                 {Object: basePlanType},
			"installmentsBasePlanType":            {Object: basePlanType},
			"billingPeriodDuration":               {},
			"gracePeriodDuration":                 {},
			"accountHoldDuration":                 {},
			"legacyCompatible":                    {},
			"legacyCompatibleSubscriptionOfferId": {},
			"legacyCompatibleSubscriptionOfferID": {},
			"prorationMode":                       {},
			"resubscribeState":                    {},
			"timeExtension":                       {},
			"committedPaymentsCount":              {},
			"renewalType":                         {},
			"offerTags": {Array: &jsonSchemaValue{Object: jsonObjectSchema{
				"tag": {},
			}}},
			"regionalConfigs": {Array: &jsonSchemaValue{Object: jsonObjectSchema{
				"regionCode":                {},
				"newSubscriberAvailability": {},
				"price":                     {Object: money},
			}}},
			"otherRegionsConfig": {Object: jsonObjectSchema{
				"newSubscriberAvailability": {},
				"usdPrice":                  {Object: money},
				"eurPrice":                  {Object: money},
			}},
		}}},
		"restrictedCountries": {Object: jsonObjectSchema{
			"regionCodes": {},
		}},
		"restrictedPaymentCountries": {Object: jsonObjectSchema{
			"regionCodes": {},
		}},
		"taxAndComplianceSettings": {},
	}
}
