package play

import (
	"context"
	"fmt"

	"google.golang.org/api/androidpublisher/v3"
)

func (p *GooglePublisher) ConvertRegionPrices(ctx context.Context, options RegionPriceConversionOptions) (RegionPriceConversionResult, error) {
	request := &androidpublisher.ConvertRegionPricesRequest{
		Price: &androidpublisher.Money{
			CurrencyCode: options.Currency.String(),
			Units:        options.Units,
			Nanos:        options.Nanos,
		},
	}
	response, err := p.service.Monetization.ConvertRegionPrices(options.PackageName.String(), request).
		Context(ctx).
		Do()
	if err != nil {
		return RegionPriceConversionResult{}, fmt.Errorf("convert region prices for %s: %w", options.PackageName, err)
	}
	return regionPriceConversionResultFromAPI(options, response), nil
}

func regionPriceConversionResultFromAPI(options RegionPriceConversionOptions, response *androidpublisher.ConvertRegionPricesResponse) RegionPriceConversionResult {
	result := RegionPriceConversionResult{
		PackageName: options.PackageName,
		SourcePrice: Money{
			CurrencyCode: options.Currency.String(),
			Units:        options.Units,
			Nanos:        options.Nanos,
		},
		ConvertedRegionPrices: map[string]ConvertedRegionPrice{},
	}
	if response == nil {
		return result
	}
	if response.RegionVersion != nil {
		result.RegionVersion = response.RegionVersion.Version
	}
	result.ConvertedOtherRegionsPrice = convertedOtherRegionsPriceFromAPI(response.ConvertedOtherRegionsPrice)
	for regionCode, apiPrice := range response.ConvertedRegionPrices {
		result.ConvertedRegionPrices[regionCode] = convertedRegionPriceFromAPI(apiPrice)
	}
	return result
}

func convertedOtherRegionsPriceFromAPI(apiPrice *androidpublisher.ConvertedOtherRegionsPrice) *ConvertedOtherRegionsPrice {
	if apiPrice == nil {
		return nil
	}
	return &ConvertedOtherRegionsPrice{
		USDPrice: moneyFromAPI(apiPrice.UsdPrice),
		EURPrice: moneyFromAPI(apiPrice.EurPrice),
	}
}

func convertedRegionPriceFromAPI(apiPrice androidpublisher.ConvertedRegionPrice) ConvertedRegionPrice {
	return ConvertedRegionPrice{
		RegionCode: apiPrice.RegionCode,
		Price:      moneyFromAPI(apiPrice.Price),
		TaxAmount:  moneyFromAPI(apiPrice.TaxAmount),
	}
}
