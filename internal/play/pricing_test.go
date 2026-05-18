package play

import (
	"context"
	"reflect"
	"testing"
)

func TestNewCurrencyCodeNormalizesUppercase(t *testing.T) {
	currency, err := NewCurrencyCode(" usd ")
	if err != nil {
		t.Fatalf("NewCurrencyCode() error = %v", err)
	}
	if currency != "USD" {
		t.Fatalf("currency = %q, want USD", currency)
	}
}

func TestConvertRegionPricesPassesOptionsToConverter(t *testing.T) {
	converter := &fakeRegionPriceConverter{result: RegionPriceConversionResult{
		PackageName: "com.example.app",
		SourcePrice: Money{CurrencyCode: "USD", Units: 9, Nanos: 990000000},
		ConvertedRegionPrices: map[string]ConvertedRegionPrice{
			"US": {RegionCode: "US"},
		},
	}}
	options := RegionPriceConversionOptions{
		PackageName: "com.example.app",
		Currency:    "USD",
		Units:       9,
		Nanos:       990000000,
	}

	result, err := ConvertRegionPrices(context.Background(), converter, options)
	if err != nil {
		t.Fatalf("ConvertRegionPrices() error = %v", err)
	}
	if len(result.ConvertedRegionPrices) != 1 {
		t.Fatalf("len(ConvertedRegionPrices) = %d, want 1", len(result.ConvertedRegionPrices))
	}
	if !reflect.DeepEqual(converter.options, options) {
		t.Fatalf("options = %#v, want %#v", converter.options, options)
	}
}

func TestConvertRegionPricesRejectsInvalidOptions(t *testing.T) {
	tests := []RegionPriceConversionOptions{
		{},
		{PackageName: "bad", Currency: "USD", Units: 1},
		{PackageName: "com.example.app", Currency: "US", Units: 1},
		{PackageName: "com.example.app", Currency: "USD", Units: -1},
		{PackageName: "com.example.app", Currency: "USD", Nanos: -1},
		{PackageName: "com.example.app", Currency: "USD", Nanos: 1000000000},
		{PackageName: "com.example.app", Currency: "USD"},
	}
	for _, options := range tests {
		_, err := ConvertRegionPrices(context.Background(), nil, options)
		if err == nil {
			t.Fatalf("ConvertRegionPrices(%#v) expected validation error", options)
		}
	}
}

type fakeRegionPriceConverter struct {
	options RegionPriceConversionOptions
	result  RegionPriceConversionResult
}

func (c *fakeRegionPriceConverter) ConvertRegionPrices(ctx context.Context, options RegionPriceConversionOptions) (RegionPriceConversionResult, error) {
	c.options = options
	return c.result, nil
}
