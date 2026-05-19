package play

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestListInAppProductsPassesOptionsToLister(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	lister := &fakeInAppProductClient{
		listResult: InAppProductListResult{
			PackageName: packageName,
			Products:    []InAppProduct{{SKU: "coins_100"}},
		},
	}

	result, err := ListInAppProducts(context.Background(), lister, InAppProductListOptions{
		PackageName: packageName,
		Token:       "next",
	})
	if err != nil {
		t.Fatalf("ListInAppProducts() error = %v", err)
	}
	if len(result.Products) != 1 {
		t.Fatalf("len(Products) = %d, want 1", len(result.Products))
	}
	if !reflect.DeepEqual(lister.listOptions, InAppProductListOptions{PackageName: packageName, Token: "next"}) {
		t.Fatalf("listOptions = %#v", lister.listOptions)
	}
}

func TestGetInAppProductPassesSKUToGetter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	getter := &fakeInAppProductClient{
		product: InAppProduct{SKU: "coins_100"},
	}

	product, err := GetInAppProduct(context.Background(), getter, InAppProductGetOptions{
		PackageName: packageName,
		SKU:         "coins_100",
	})
	if err != nil {
		t.Fatalf("GetInAppProduct() error = %v", err)
	}
	if product.SKU != "coins_100" {
		t.Fatalf("SKU = %q, want coins_100", product.SKU)
	}
	if getter.sku != "coins_100" {
		t.Fatalf("getter sku = %q, want coins_100", getter.sku)
	}
}

func TestGetInAppProductRejectsMissingSKU(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = GetInAppProduct(context.Background(), nil, InAppProductGetOptions{PackageName: packageName})
	if err == nil {
		t.Fatal("expected SKU validation error")
	}
}

func TestBatchGetInAppProductsPassesOptionsToGetter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	getter := &fakeInAppProductClient{
		batchResult: InAppProductBatchGetResult{
			PackageName: packageName,
			Products:    []InAppProduct{{SKU: "coins_100"}},
		},
	}
	options := InAppProductBatchGetOptions{
		PackageName: packageName,
		SKUs:        []InAppProductSKU{"coins_100", "coins_500"},
	}

	result, err := BatchGetInAppProducts(context.Background(), getter, options)
	if err != nil {
		t.Fatalf("BatchGetInAppProducts() error = %v", err)
	}
	if len(result.Products) != 1 {
		t.Fatalf("len(Products) = %d, want 1", len(result.Products))
	}
	if !reflect.DeepEqual(getter.batchOptions, options) {
		t.Fatalf("batchOptions = %#v, want %#v", getter.batchOptions, options)
	}
}

func TestBatchGetInAppProductsRejectsMissingSKU(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = BatchGetInAppProducts(context.Background(), nil, InAppProductBatchGetOptions{
		PackageName: packageName,
	})
	if err == nil {
		t.Fatal("expected SKU validation error")
	}
}

func TestBatchGetInAppProductsRejectsDuplicateSKU(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = BatchGetInAppProducts(context.Background(), nil, InAppProductBatchGetOptions{
		PackageName: packageName,
		SKUs:        []InAppProductSKU{"coins_100", "coins_100"},
	})
	if err == nil {
		t.Fatal("expected duplicate SKU validation error")
	}
}

func TestDeleteInAppProductDryRunBuildsPlanWithoutDeleter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := DeleteInAppProduct(context.Background(), nil, InAppProductDeleteOptions{
		PackageName:      packageName,
		SKU:              "coins_100",
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("DeleteInAppProduct() error = %v", err)
	}
	if !result.DryRun || result.Deleted {
		t.Fatalf("result = %#v, want dry-run deletion plan", result)
	}
	wantSteps := []string{"plan in-app product deletion"}
	if !reflect.DeepEqual(result.Plan.Steps, wantSteps) {
		t.Fatalf("steps = %#v, want %#v", result.Plan.Steps, wantSteps)
	}
}

func TestDeleteInAppProductRequiresConfirmOrDryRun(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = DeleteInAppProduct(context.Background(), nil, InAppProductDeleteOptions{
		PackageName:      packageName,
		SKU:              "coins_100",
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
	})
	if err == nil {
		t.Fatal("expected confirmation validation error")
	}
}

func TestDeleteInAppProductPassesOptionsToDeleter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	deleter := &fakeInAppProductClient{
		product: InAppProduct{SKU: "coins_100", PurchaseType: ProductPurchaseTypeManagedUser},
	}
	options := InAppProductDeleteOptions{
		PackageName:      packageName,
		SKU:              "coins_100",
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	}

	result, err := DeleteInAppProduct(context.Background(), deleter, options)
	if err != nil {
		t.Fatalf("DeleteInAppProduct() error = %v", err)
	}
	if !result.Deleted {
		t.Fatal("Deleted = false, want true")
	}
	if !reflect.DeepEqual(deleter.deleteOptions, options) {
		t.Fatalf("deleteOptions = %#v, want %#v", deleter.deleteOptions, options)
	}
}

func TestDeleteInAppProductRejectsLegacySubscription(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	deleter := &fakeInAppProductClient{
		product: InAppProduct{SKU: "premium", PurchaseType: ProductPurchaseTypeSubscription},
	}

	_, err = DeleteInAppProduct(context.Background(), deleter, InAppProductDeleteOptions{
		PackageName:      packageName,
		SKU:              "premium",
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err == nil {
		t.Fatal("expected legacy subscription rejection")
	}
	if deleter.deleteOptions.SKU != "" {
		t.Fatalf("deleteOptions = %#v, did not expect delete after preflight", deleter.deleteOptions)
	}
}

func TestDeleteInAppProductRejectsUnspecifiedPurchaseType(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	deleter := &fakeInAppProductClient{
		product: InAppProduct{SKU: "coins_100", PurchaseType: ProductPurchaseTypeUnspecified},
	}

	_, err = DeleteInAppProduct(context.Background(), deleter, InAppProductDeleteOptions{
		PackageName:      packageName,
		SKU:              "coins_100",
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err == nil {
		t.Fatal("expected unspecified purchase type rejection")
	}
	if deleter.deleteOptions.SKU != "" {
		t.Fatalf("deleteOptions = %#v, did not expect delete after preflight", deleter.deleteOptions)
	}
}

func TestBatchDeleteInAppProductsDryRunBuildsPlanWithoutDeleter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := BatchDeleteInAppProducts(context.Background(), nil, InAppProductBatchDeleteOptions{
		PackageName:      packageName,
		SKUs:             []InAppProductSKU{"coins_100", "coins_500"},
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("BatchDeleteInAppProducts() error = %v", err)
	}
	if !result.DryRun || result.Deleted {
		t.Fatalf("result = %#v, want dry-run batch deletion plan", result)
	}
	wantSteps := []string{"plan in-app product batch deletion"}
	if !reflect.DeepEqual(result.Plan.Steps, wantSteps) {
		t.Fatalf("steps = %#v, want %#v", result.Plan.Steps, wantSteps)
	}
}

func TestBatchDeleteInAppProductsRejectsDuplicates(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = BatchDeleteInAppProducts(context.Background(), nil, InAppProductBatchDeleteOptions{
		PackageName:      packageName,
		SKUs:             []InAppProductSKU{"coins_100", "coins_100"},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected duplicate SKU validation error")
	}
}

func TestBatchDeleteInAppProductsRejectsMoreThanGoogleLimit(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	skus := make([]InAppProductSKU, 101)
	for index := range skus {
		skus[index] = InAppProductSKU(fmt.Sprintf("coins_%d", index))
	}

	_, err = BatchDeleteInAppProducts(context.Background(), nil, InAppProductBatchDeleteOptions{
		PackageName:      packageName,
		SKUs:             skus,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected batch limit validation error")
	}
}

func TestBatchDeleteInAppProductsPassesOptionsToDeleter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	deleter := &fakeInAppProductClient{
		batchResult: InAppProductBatchGetResult{
			PackageName: packageName,
			Products: []InAppProduct{
				{SKU: "coins_100", PurchaseType: ProductPurchaseTypeManagedUser},
				{SKU: "coins_500", PurchaseType: ProductPurchaseTypeManagedUser},
			},
		},
	}
	options := InAppProductBatchDeleteOptions{
		PackageName:      packageName,
		SKUs:             []InAppProductSKU{"coins_100", "coins_500"},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	}

	result, err := BatchDeleteInAppProducts(context.Background(), deleter, options)
	if err != nil {
		t.Fatalf("BatchDeleteInAppProducts() error = %v", err)
	}
	if !result.Deleted {
		t.Fatal("Deleted = false, want true")
	}
	if !reflect.DeepEqual(deleter.batchDeleteOptions, options) {
		t.Fatalf("batchDeleteOptions = %#v, want %#v", deleter.batchDeleteOptions, options)
	}
}

func TestBatchDeleteInAppProductsRejectsMissingPreflightProduct(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	deleter := &fakeInAppProductClient{
		batchResult: InAppProductBatchGetResult{
			PackageName: packageName,
			Products: []InAppProduct{
				{SKU: "coins_100", PurchaseType: ProductPurchaseTypeManagedUser},
			},
		},
	}

	_, err = BatchDeleteInAppProducts(context.Background(), deleter, InAppProductBatchDeleteOptions{
		PackageName:      packageName,
		SKUs:             []InAppProductSKU{"coins_100", "coins_500"},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err == nil {
		t.Fatal("expected missing preflight product validation error")
	}
	if len(deleter.batchDeleteOptions.SKUs) != 0 {
		t.Fatalf("batchDeleteOptions = %#v, did not expect delete after preflight", deleter.batchDeleteOptions)
	}
}

func TestBatchDeleteInAppProductsRejectsNonManagedPreflightProduct(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	deleter := &fakeInAppProductClient{
		batchResult: InAppProductBatchGetResult{
			PackageName: packageName,
			Products: []InAppProduct{
				{SKU: "coins_100", PurchaseType: ProductPurchaseTypeManagedUser},
				{SKU: "premium", PurchaseType: ProductPurchaseTypeSubscription},
			},
		},
	}

	_, err = BatchDeleteInAppProducts(context.Background(), deleter, InAppProductBatchDeleteOptions{
		PackageName:      packageName,
		SKUs:             []InAppProductSKU{"coins_100", "premium"},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err == nil {
		t.Fatal("expected non-managed preflight product validation error")
	}
	if len(deleter.batchDeleteOptions.SKUs) != 0 {
		t.Fatalf("batchDeleteOptions = %#v, did not expect delete after preflight", deleter.batchDeleteOptions)
	}
}

func TestNewProductPriceParsesCurrencyAndMicros(t *testing.T) {
	price, err := NewProductPrice("usd:1990000")
	if err != nil {
		t.Fatalf("NewProductPrice() error = %v", err)
	}
	if price.Currency != "USD" || price.PriceMicros != "1990000" {
		t.Fatalf("price = %#v, want USD 1990000", price)
	}
}

func TestNewProductPriceRejectsInvalidMicros(t *testing.T) {
	_, err := NewProductPrice("USD:0")
	if err == nil {
		t.Fatal("expected positive micros validation error")
	}
}

func TestNewRegionalProductPriceParsesRegionCurrencyAndMicros(t *testing.T) {
	price, err := NewRegionalProductPrice("us:usd:2990000")
	if err != nil {
		t.Fatalf("NewRegionalProductPrice() error = %v", err)
	}
	if price.RegionCode != "US" || price.Price.Currency != "USD" || price.Price.PriceMicros != "2990000" {
		t.Fatalf("price = %#v, want US USD 2990000", price)
	}
}

func TestNewRegionalProductPriceRejectsInvalidRegion(t *testing.T) {
	_, err := NewRegionalProductPrice("USA:USD:2990000")
	if err == nil {
		t.Fatal("expected region validation error")
	}
}

func TestCreateInAppProductDryRunBuildsManagedProductPlan(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	price, err := NewProductPrice("USD:1990000")
	if err != nil {
		t.Fatalf("NewProductPrice() error = %v", err)
	}

	result, err := CreateInAppProduct(context.Background(), nil, InAppProductCreateOptions{
		PackageName:     packageName,
		SKU:             "coins_100",
		Status:          ProductStatusInactive,
		DefaultLanguage: "en-US",
		DefaultPrice:    price,
		Listing:         InAppProductListing{Title: "100 coins", Description: "A small coin pack."},
		DryRun:          true,
	})
	if err != nil {
		t.Fatalf("CreateInAppProduct() error = %v", err)
	}
	if result.Created || !result.DryRun || result.Desired.PurchaseType != ProductPurchaseTypeManagedUser {
		t.Fatalf("result = %#v, want dry-run managed product creation", result)
	}
	if result.Desired.DefaultPrice == nil || result.Desired.DefaultPrice.PriceMicros != "1990000" {
		t.Fatalf("DefaultPrice = %#v, want 1990000 micros", result.Desired.DefaultPrice)
	}
	if result.Desired.Listings["en-US"].Title != "100 coins" {
		t.Fatalf("Listings = %#v, want default listing", result.Desired.Listings)
	}
	if !result.Plan.AutoConvertMissingPrices {
		t.Fatal("AutoConvertMissingPrices = false, want true")
	}
}

func TestCreateInAppProductRequiresConfirmOrDryRun(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	price, err := NewProductPrice("USD:1990000")
	if err != nil {
		t.Fatalf("NewProductPrice() error = %v", err)
	}

	_, err = CreateInAppProduct(context.Background(), nil, InAppProductCreateOptions{
		PackageName:     packageName,
		SKU:             "coins_100",
		Status:          ProductStatusInactive,
		DefaultLanguage: "en-US",
		DefaultPrice:    price,
		Listing:         InAppProductListing{Title: "100 coins", Description: "A small coin pack."},
	})
	if err == nil {
		t.Fatal("expected confirmation validation error")
	}
}

func TestCreateInAppProductRejectsMissingListing(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	price, err := NewProductPrice("USD:1990000")
	if err != nil {
		t.Fatalf("NewProductPrice() error = %v", err)
	}

	_, err = CreateInAppProduct(context.Background(), nil, InAppProductCreateOptions{
		PackageName:     packageName,
		SKU:             "coins_100",
		Status:          ProductStatusInactive,
		DefaultLanguage: "en-US",
		DefaultPrice:    price,
		DryRun:          true,
	})
	if err == nil {
		t.Fatal("expected listing validation error")
	}
}

func TestCreateInAppProductRejectsInvalidManagedProductSKU(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	price, err := NewProductPrice("USD:1990000")
	if err != nil {
		t.Fatalf("NewProductPrice() error = %v", err)
	}

	_, err = CreateInAppProduct(context.Background(), nil, InAppProductCreateOptions{
		PackageName:     packageName,
		SKU:             "Coins-100",
		Status:          ProductStatusInactive,
		DefaultLanguage: "en-US",
		DefaultPrice:    price,
		Listing:         InAppProductListing{Title: "100 coins", Description: "A small coin pack."},
		DryRun:          true,
	})
	if err == nil {
		t.Fatal("expected SKU validation error")
	}
}

func TestCreateInAppProductRejectsReservedTestSKU(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	price, err := NewProductPrice("USD:1990000")
	if err != nil {
		t.Fatalf("NewProductPrice() error = %v", err)
	}

	_, err = CreateInAppProduct(context.Background(), nil, InAppProductCreateOptions{
		PackageName:     packageName,
		SKU:             "android.test.purchased",
		Status:          ProductStatusInactive,
		DefaultLanguage: "en-US",
		DefaultPrice:    price,
		Listing:         InAppProductListing{Title: "100 coins", Description: "A small coin pack."},
		DryRun:          true,
	})
	if err == nil {
		t.Fatal("expected reserved SKU validation error")
	}
}

func TestCreateInAppProductRejectsOverlongListing(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	price, err := NewProductPrice("USD:1990000")
	if err != nil {
		t.Fatalf("NewProductPrice() error = %v", err)
	}

	_, err = CreateInAppProduct(context.Background(), nil, InAppProductCreateOptions{
		PackageName:     packageName,
		SKU:             "coins_100",
		Status:          ProductStatusInactive,
		DefaultLanguage: "en-US",
		DefaultPrice:    price,
		Listing:         InAppProductListing{Title: strings.Repeat("x", 56), Description: "A small coin pack."},
		DryRun:          true,
	})
	if err == nil {
		t.Fatal("expected listing title length validation error")
	}
}

func TestCreateInAppProductPassesOptionsToCreator(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	price, err := NewProductPrice("USD:1990000")
	if err != nil {
		t.Fatalf("NewProductPrice() error = %v", err)
	}
	creator := &fakeInAppProductClient{product: InAppProduct{SKU: "coins_100", Status: ProductStatusInactive}}

	result, err := CreateInAppProduct(context.Background(), creator, InAppProductCreateOptions{
		PackageName:     packageName,
		SKU:             "coins_100",
		Status:          ProductStatusInactive,
		DefaultLanguage: "en-US",
		DefaultPrice:    price,
		Listing:         InAppProductListing{Title: "100 coins", Description: "A small coin pack."},
		Confirm:         true,
	})
	if err != nil {
		t.Fatalf("CreateInAppProduct() error = %v", err)
	}
	if !result.Created || result.Product == nil || result.Product.SKU != "coins_100" {
		t.Fatalf("result = %#v, want created product", result)
	}
	if creator.createOptions.DefaultPrice.PriceMicros != "1990000" {
		t.Fatalf("createOptions = %#v, want price", creator.createOptions)
	}
}

func TestPatchInAppProductDryRunBuildsPlanWithoutPatcher(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := PatchInAppProduct(context.Background(), nil, InAppProductPatchOptions{
		PackageName: packageName,
		SKU:         "coins_100",
		Status:      ProductStatusActive,
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("PatchInAppProduct() error = %v", err)
	}
	if result.Applied || !result.DryRun || result.Desired.Status != ProductStatusActive {
		t.Fatalf("result = %#v, want dry-run active patch", result)
	}
}

func TestPatchInAppProductDryRunBuildsPriceAndListingPlan(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	price, err := NewProductPrice("USD:2990000")
	if err != nil {
		t.Fatalf("NewProductPrice() error = %v", err)
	}

	result, err := PatchInAppProduct(context.Background(), nil, InAppProductPatchOptions{
		PackageName:     packageName,
		SKU:             "coins_100",
		ListingLanguage: "en-US",
		DefaultPrice:    &price,
		Listing:         &InAppProductListing{Title: "100 coins", Description: "A better coin pack."},
		DryRun:          true,
	})
	if err != nil {
		t.Fatalf("PatchInAppProduct() error = %v", err)
	}
	if result.Desired.DefaultPrice == nil || result.Desired.DefaultPrice.PriceMicros != "2990000" {
		t.Fatalf("DefaultPrice = %#v, want 2990000 micros", result.Desired.DefaultPrice)
	}
	if result.Desired.Listings["en-US"].Description != "A better coin pack." {
		t.Fatalf("Listings = %#v, want patched listing", result.Desired.Listings)
	}
	if !result.Plan.AutoConvertMissingPrices {
		t.Fatal("AutoConvertMissingPrices = false, want true")
	}
}

func TestPatchInAppProductDryRunBuildsRegionalPricePlan(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	regionalPrice, err := NewRegionalProductPrice("US:USD:2990000")
	if err != nil {
		t.Fatalf("NewRegionalProductPrice() error = %v", err)
	}

	result, err := PatchInAppProduct(context.Background(), nil, InAppProductPatchOptions{
		PackageName:    packageName,
		SKU:            "coins_100",
		RegionalPrices: []RegionalProductPrice{regionalPrice},
		DryRun:         true,
	})
	if err != nil {
		t.Fatalf("PatchInAppProduct() error = %v", err)
	}
	if result.Desired.Prices["US"].PriceMicros != "2990000" {
		t.Fatalf("Prices = %#v, want US 2990000 micros", result.Desired.Prices)
	}
	if len(result.Plan.RegionalPrices) != 1 || !result.Plan.AutoConvertMissingPrices {
		t.Fatalf("Plan = %#v, want one regional price with auto-conversion", result.Plan)
	}
}

func TestPatchInAppProductRequiresConfirmOrDryRun(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = PatchInAppProduct(context.Background(), nil, InAppProductPatchOptions{
		PackageName: packageName,
		SKU:         "coins_100",
		Status:      ProductStatusActive,
	})
	if err == nil {
		t.Fatal("expected confirmation validation error")
	}
}

func TestPatchInAppProductRequiresMutation(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = PatchInAppProduct(context.Background(), nil, InAppProductPatchOptions{
		PackageName: packageName,
		SKU:         "coins_100",
		DryRun:      true,
	})
	if err == nil {
		t.Fatal("expected mutation validation error")
	}
}

func TestPatchInAppProductListingRequiresListingLanguage(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = PatchInAppProduct(context.Background(), nil, InAppProductPatchOptions{
		PackageName: packageName,
		SKU:         "coins_100",
		Listing:     &InAppProductListing{Title: "100 coins", Description: "A small coin pack."},
		DryRun:      true,
	})
	if err == nil {
		t.Fatal("expected listing language validation error")
	}
}

func TestPatchInAppProductRejectsPartialListing(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = PatchInAppProduct(context.Background(), nil, InAppProductPatchOptions{
		PackageName:     packageName,
		SKU:             "coins_100",
		ListingLanguage: "en-US",
		Listing:         &InAppProductListing{Title: "100 coins"},
		DryRun:          true,
	})
	if err == nil {
		t.Fatal("expected full listing validation error")
	}
}

func TestPatchInAppProductRejectsConfirmAndDryRun(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = PatchInAppProduct(context.Background(), nil, InAppProductPatchOptions{
		PackageName: packageName,
		SKU:         "coins_100",
		Status:      ProductStatusActive,
		Confirm:     true,
		DryRun:      true,
	})
	if err == nil {
		t.Fatal("expected mutually exclusive flag validation error")
	}
}

func TestPatchInAppProductRejectsDuplicateRegionalPrice(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	usPrice, err := NewRegionalProductPrice("US:USD:2990000")
	if err != nil {
		t.Fatalf("NewRegionalProductPrice() error = %v", err)
	}

	_, err = PatchInAppProduct(context.Background(), nil, InAppProductPatchOptions{
		PackageName:    packageName,
		SKU:            "coins_100",
		RegionalPrices: []RegionalProductPrice{usPrice, usPrice},
		DryRun:         true,
	})
	if err == nil {
		t.Fatal("expected duplicate regional price validation error")
	}
}

func TestPatchInAppProductRejectsNonCanonicalStatus(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = PatchInAppProduct(context.Background(), nil, InAppProductPatchOptions{
		PackageName: packageName,
		SKU:         "coins_100",
		Status:      ProductStatus(" inactive "),
		DryRun:      true,
	})
	if err == nil {
		t.Fatal("expected status validation error")
	}
}

func TestPatchInAppProductPassesOptionsToPatcher(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	patcher := &fakeInAppProductClient{product: InAppProduct{SKU: "coins_100", Status: ProductStatusInactive}}

	result, err := PatchInAppProduct(context.Background(), patcher, InAppProductPatchOptions{
		PackageName: packageName,
		SKU:         "coins_100",
		Status:      ProductStatusInactive,
		Confirm:     true,
	})
	if err != nil {
		t.Fatalf("PatchInAppProduct() error = %v", err)
	}
	if !result.Applied || result.Product == nil || result.Product.Status != ProductStatusInactive {
		t.Fatalf("result = %#v, want applied inactive product", result)
	}
	if patcher.patchOptions.Status != ProductStatusInactive {
		t.Fatalf("patchOptions = %#v", patcher.patchOptions)
	}
}

func TestPatchInAppProductRejectsLegacySubscription(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	patcher := &fakeInAppProductClient{
		product: InAppProduct{SKU: "premium", PurchaseType: ProductPurchaseTypeSubscription},
	}

	_, err = PatchInAppProduct(context.Background(), patcher, InAppProductPatchOptions{
		PackageName: packageName,
		SKU:         "premium",
		Status:      ProductStatusInactive,
		Confirm:     true,
	})
	if err == nil {
		t.Fatal("expected legacy subscription rejection")
	}
	if patcher.patchOptions.SKU != "" {
		t.Fatalf("patchOptions = %#v, did not expect patch after preflight", patcher.patchOptions)
	}
}

type fakeInAppProductClient struct {
	listOptions        InAppProductListOptions
	listResult         InAppProductListResult
	batchOptions       InAppProductBatchGetOptions
	batchResult        InAppProductBatchGetResult
	batchDeleteOptions InAppProductBatchDeleteOptions
	createOptions      InAppProductCreateOptions
	deleteOptions      InAppProductDeleteOptions
	patchOptions       InAppProductPatchOptions
	sku                InAppProductSKU
	product            InAppProduct
}

func (c *fakeInAppProductClient) ListInAppProducts(ctx context.Context, options InAppProductListOptions) (InAppProductListResult, error) {
	c.listOptions = options
	return c.listResult, nil
}

func (c *fakeInAppProductClient) GetInAppProduct(ctx context.Context, packageName PackageName, sku InAppProductSKU) (InAppProduct, error) {
	c.sku = sku
	return c.product, nil
}

func (c *fakeInAppProductClient) BatchGetInAppProducts(ctx context.Context, options InAppProductBatchGetOptions) (InAppProductBatchGetResult, error) {
	c.batchOptions = options
	return c.batchResult, nil
}

func (c *fakeInAppProductClient) CreateInAppProduct(ctx context.Context, options InAppProductCreateOptions) (InAppProduct, error) {
	c.createOptions = options
	return c.product, nil
}

func (c *fakeInAppProductClient) DeleteInAppProduct(ctx context.Context, options InAppProductDeleteOptions) error {
	c.deleteOptions = options
	return nil
}

func (c *fakeInAppProductClient) BatchDeleteInAppProducts(ctx context.Context, options InAppProductBatchDeleteOptions) error {
	c.batchDeleteOptions = options
	return nil
}

func (c *fakeInAppProductClient) PatchInAppProduct(ctx context.Context, options InAppProductPatchOptions) (InAppProduct, error) {
	c.patchOptions = options
	return c.product, nil
}
