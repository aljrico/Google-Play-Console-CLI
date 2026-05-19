package play

import (
	"google.golang.org/api/androidpublisher/v3"
)

func productUpdateLatencyToleranceToAPI(latencyTolerance ProductUpdateLatencyTolerance) string {
	switch latencyTolerance {
	case ProductUpdateLatencyToleranceTolerant:
		return "PRODUCT_UPDATE_LATENCY_TOLERANCE_LATENCY_TOLERANT"
	default:
		return "PRODUCT_UPDATE_LATENCY_TOLERANCE_LATENCY_SENSITIVE"
	}
}

func offerTagsToAPI(tags []string) []*androidpublisher.OfferTag {
	apiTags := make([]*androidpublisher.OfferTag, 0, len(tags))
	for _, tag := range tags {
		apiTags = append(apiTags, &androidpublisher.OfferTag{Tag: tag})
	}
	return apiTags
}

func moneyToAPI(money *Money) *androidpublisher.Money {
	if money == nil {
		return nil
	}
	return &androidpublisher.Money{
		CurrencyCode: money.CurrencyCode,
		Units:        money.Units,
		Nanos:        money.Nanos,
	}
}

func moneyFromAPI(apiMoney *androidpublisher.Money) *Money {
	if apiMoney == nil {
		return nil
	}
	return &Money{
		CurrencyCode: apiMoney.CurrencyCode,
		Units:        apiMoney.Units,
		Nanos:        apiMoney.Nanos,
	}
}

func offerTagsFromAPI(apiOfferTags []*androidpublisher.OfferTag) []string {
	if len(apiOfferTags) == 0 {
		return nil
	}
	tags := make([]string, 0, len(apiOfferTags))
	for _, apiOfferTag := range apiOfferTags {
		if apiOfferTag == nil {
			continue
		}
		tags = append(tags, apiOfferTag.Tag)
	}
	return tags
}

func restrictedCountriesToAPI(countries []string) *androidpublisher.RestrictedPaymentCountries {
	if len(countries) == 0 {
		return nil
	}
	return &androidpublisher.RestrictedPaymentCountries{RegionCodes: append([]string(nil), countries...)}
}

func stringPointer(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
