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
		{PackageName: "com.example.app", Currency: " usd ", Units: 1},
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

func TestBuildPricePatchesBuildsSubscriptionOfferPhaseArgs(t *testing.T) {
	result, err := BuildPricePatches(PricePatchBuildOptions{
		Conversion: RegionPriceConversionResult{
			PackageName:   "com.example.app",
			SourcePrice:   Money{CurrencyCode: "USD", Units: 9, Nanos: 990000000},
			RegionVersion: "2026/05",
			ConvertedRegionPrices: map[string]ConvertedRegionPrice{
				"US": {Price: &Money{CurrencyCode: "USD", Units: 9, Nanos: 990000000}},
				"BR": {RegionCode: "BR", Price: &Money{CurrencyCode: "BRL", Units: 19}},
			},
		},
		Target:     PricePatchBuildTargetSubscriptionOfferPhase,
		ProductID:  "premium",
		BasePlanID: "monthly",
		OfferID:    "intro",
		PhaseIndex: 0,
	})
	if err != nil {
		t.Fatalf("BuildPricePatches() error = %v", err)
	}
	wantArgs := []string{
		"premium/monthly/intro/0/BR:BRL:19",
		"premium/monthly/intro/0/US:USD:9:990000000",
	}
	if !reflect.DeepEqual(result.PriceArgs, wantArgs) {
		t.Fatalf("PriceArgs = %#v, want %#v", result.PriceArgs, wantArgs)
	}
	wantCommand := []string{
		"gpc", "subscription-offers", "batch-patch-phase-prices",
		"--package", "com.example.app",
		"--regions-version", "2026/05",
		"--price", "premium/monthly/intro/0/BR:BRL:19",
		"--price", "premium/monthly/intro/0/US:USD:9:990000000",
		"--dry-run",
	}
	if !reflect.DeepEqual(result.SuggestedCommand, wantCommand) {
		t.Fatalf("SuggestedCommand = %#v, want %#v", result.SuggestedCommand, wantCommand)
	}
	if result.PhaseIndex == nil || *result.PhaseIndex != 0 {
		t.Fatalf("PhaseIndex = %#v, want 0", result.PhaseIndex)
	}
}

func TestBuildPricePatchesBuildsInAppProductMicrosArgs(t *testing.T) {
	result, err := BuildPricePatches(PricePatchBuildOptions{
		Conversion: RegionPriceConversionResult{
			PackageName: "com.example.app",
			SourcePrice: Money{CurrencyCode: "USD", Units: 1},
			ConvertedRegionPrices: map[string]ConvertedRegionPrice{
				"US": {Price: &Money{CurrencyCode: "USD", Units: 1, Nanos: 990000000}},
			},
		},
		Target: PricePatchBuildTargetInAppProduct,
		SKU:    "coins_100",
	})
	if err != nil {
		t.Fatalf("BuildPricePatches() error = %v", err)
	}
	if !reflect.DeepEqual(result.PriceArgs, []string{"US:USD:1990000"}) {
		t.Fatalf("PriceArgs = %#v, want micros", result.PriceArgs)
	}
}

func TestBuildPricePatchesRejectsInvalidOptions(t *testing.T) {
	validConversion := RegionPriceConversionResult{
		PackageName: "com.example.app",
		SourcePrice: Money{CurrencyCode: "USD", Units: 1},
		ConvertedRegionPrices: map[string]ConvertedRegionPrice{
			"US": {Price: &Money{CurrencyCode: "USD", Units: 1}},
		},
	}
	tests := []PricePatchBuildOptions{
		{},
		{Conversion: validConversion, Target: "wat"},
		{Conversion: validConversion, Target: PricePatchBuildTargetInAppProduct},
		{Conversion: validConversion, Target: PricePatchBuildTargetOneTimeProduct, ProductID: "coins_100", PurchaseOptionID: "buy"},
		{Conversion: validConversion, Target: PricePatchBuildTargetOneTimeProduct, ProductID: "coins_100"},
		{Conversion: validConversion, Target: PricePatchBuildTargetSubscriptionBasePlan, ProductID: "premium"},
		{Conversion: validConversion, Target: PricePatchBuildTargetSubscriptionOfferPhase, ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: -1},
		{Conversion: RegionPriceConversionResult{
			PackageName: "com.example.app",
			SourcePrice: Money{CurrencyCode: "USD", Units: 1},
			ConvertedRegionPrices: map[string]ConvertedRegionPrice{
				"US": {Price: &Money{CurrencyCode: "USD", Units: 1, Nanos: 1}},
			},
		}, Target: PricePatchBuildTargetInAppProduct, SKU: "coins_100"},
		{Conversion: RegionPriceConversionResult{
			PackageName: "com.example.app",
			SourcePrice: Money{CurrencyCode: "USD", Units: 1},
			ConvertedRegionPrices: map[string]ConvertedRegionPrice{
				"US": {Price: &Money{CurrencyCode: "USD", Units: 1}},
				"XX": {RegionCode: "US", Price: &Money{CurrencyCode: "USD", Units: 2}},
			},
		}, Target: PricePatchBuildTargetInAppProduct, SKU: "coins_100"},
		{Conversion: RegionPriceConversionResult{
			PackageName: "com.example.app",
			SourcePrice: Money{CurrencyCode: "USD", Units: 1},
			ConvertedRegionPrices: map[string]ConvertedRegionPrice{
				"BR": {RegionCode: "US", Price: &Money{CurrencyCode: "USD", Units: 1}},
			},
		}, Target: PricePatchBuildTargetInAppProduct, SKU: "coins_100"},
	}
	for _, options := range tests {
		if _, err := BuildPricePatches(options); err == nil {
			t.Fatalf("BuildPricePatches(%#v) error = nil, want validation error", options)
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
