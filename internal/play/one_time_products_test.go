package play

import (
	"context"
	"fmt"
	"reflect"
	"testing"
)

func TestNewOneTimeProductIDValidatesGoogleShape(t *testing.T) {
	valid, err := NewOneTimeProductID("coins_100.v2")
	if err != nil {
		t.Fatalf("NewOneTimeProductID() error = %v", err)
	}
	if valid != "coins_100.v2" {
		t.Fatalf("product ID = %q, want coins_100.v2", valid)
	}

	for _, value := range []string{"", "Coins", "_coins", "coins-100"} {
		t.Run(value, func(t *testing.T) {
			if _, err := NewOneTimeProductID(value); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestListOneTimeProductsPassesOptionsToLister(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	lister := &fakeOneTimeProductClient{
		listResult: OneTimeProductListResult{
			PackageName: packageName,
			Products:    []OneTimeProduct{{ProductID: "coins_100"}},
		},
	}

	result, err := ListOneTimeProducts(context.Background(), lister, OneTimeProductListOptions{
		PackageName: packageName,
		PageSize:    50,
		PageToken:   "next",
	})
	if err != nil {
		t.Fatalf("ListOneTimeProducts() error = %v", err)
	}
	if len(result.Products) != 1 {
		t.Fatalf("len(Products) = %d, want 1", len(result.Products))
	}
	if !reflect.DeepEqual(lister.listOptions, OneTimeProductListOptions{PackageName: packageName, PageSize: 50, PageToken: "next"}) {
		t.Fatalf("listOptions = %#v", lister.listOptions)
	}
}

func TestListOneTimeProductsRejectsInvalidPageSize(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = ListOneTimeProducts(context.Background(), nil, OneTimeProductListOptions{
		PackageName: packageName,
		PageSize:    1001,
	})
	if err == nil {
		t.Fatal("expected page size validation error")
	}
}

func TestGetOneTimeProductPassesProductIDToGetter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	getter := &fakeOneTimeProductClient{
		product: OneTimeProduct{ProductID: "coins_100"},
	}

	product, err := GetOneTimeProduct(context.Background(), getter, OneTimeProductGetOptions{
		PackageName: packageName,
		ProductID:   "coins_100",
	})
	if err != nil {
		t.Fatalf("GetOneTimeProduct() error = %v", err)
	}
	if product.ProductID != "coins_100" {
		t.Fatalf("ProductID = %q, want coins_100", product.ProductID)
	}
	if getter.productID != "coins_100" {
		t.Fatalf("getter productID = %q, want coins_100", getter.productID)
	}
}

func TestGetOneTimeProductRejectsMissingProductID(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = GetOneTimeProduct(context.Background(), nil, OneTimeProductGetOptions{PackageName: packageName})
	if err == nil {
		t.Fatal("expected product ID validation error")
	}
}

func TestBatchGetOneTimeProductsPassesOptionsToGetter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	getter := &fakeOneTimeProductClient{
		batchResult: OneTimeProductBatchGetResult{
			PackageName: packageName,
			Products:    []OneTimeProduct{{ProductID: "coins_100"}, {ProductID: "coins_500"}},
		},
	}
	options := OneTimeProductBatchGetOptions{
		PackageName: packageName,
		ProductIDs:  []OneTimeProductID{"coins_100", "coins_500"},
	}

	result, err := BatchGetOneTimeProducts(context.Background(), getter, options)
	if err != nil {
		t.Fatalf("BatchGetOneTimeProducts() error = %v", err)
	}
	if len(result.Products) != 2 {
		t.Fatalf("len(Products) = %d, want 2", len(result.Products))
	}
	if !reflect.DeepEqual(getter.batchOptions, options) {
		t.Fatalf("batchOptions = %#v, want %#v", getter.batchOptions, options)
	}
}

func TestBatchGetOneTimeProductsRejectsDuplicates(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = BatchGetOneTimeProducts(context.Background(), nil, OneTimeProductBatchGetOptions{
		PackageName: packageName,
		ProductIDs:  []OneTimeProductID{"coins_100", "coins_100"},
	})
	if err == nil {
		t.Fatal("expected duplicate product ID validation error")
	}
}

func TestDeleteOneTimeProductDryRunBuildsPlanWithoutDeleter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := DeleteOneTimeProduct(context.Background(), nil, OneTimeProductDeleteOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("DeleteOneTimeProduct() error = %v", err)
	}
	if !result.DryRun || result.Deleted {
		t.Fatalf("result = %#v, want dry-run deletion plan", result)
	}
	if !reflect.DeepEqual(result.Plan.Steps, []string{"delete one-time product"}) {
		t.Fatalf("steps = %#v, want delete step", result.Plan.Steps)
	}
}

func TestDeleteOneTimeProductRequiresConfirmOrDryRun(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = DeleteOneTimeProduct(context.Background(), nil, OneTimeProductDeleteOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
	})
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
}

func TestDeleteOneTimeProductPassesOptionsToDeleter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	deleter := &fakeOneTimeProductClient{}

	result, err := DeleteOneTimeProduct(context.Background(), deleter, OneTimeProductDeleteOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("DeleteOneTimeProduct() error = %v", err)
	}
	if !result.Deleted {
		t.Fatal("Deleted = false, want true")
	}
	if deleter.deleteOptions.ProductID != "coins_100" {
		t.Fatalf("deleteOptions = %#v, want coins_100", deleter.deleteOptions)
	}
}

func TestPatchOneTimeProductDryRunBuildsPlanWithoutPatcher(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := PatchOneTimeProduct(context.Background(), nil, OneTimeProductPatchOptions{
		PackageName: packageName,
		ProductID:   "coins_100",
		Listing: OneTimeProductListing{
			LanguageCode: "en-US",
			Title:        "100 coins",
			Description:  "Buy a stack of coins.",
		},
		DescriptionSet:   true,
		RegionsVersion:   "2022/02",
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("PatchOneTimeProduct() error = %v", err)
	}
	if !result.DryRun || result.Applied {
		t.Fatalf("result = %#v, want dry-run patch plan", result)
	}
	if result.Plan.UpdateMask != oneTimeProductPatchUpdateMask {
		t.Fatalf("UpdateMask = %q, want %q", result.Plan.UpdateMask, oneTimeProductPatchUpdateMask)
	}
}

func TestPatchOneTimeProductRejectsLongListingText(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = PatchOneTimeProduct(context.Background(), nil, OneTimeProductPatchOptions{
		PackageName: packageName,
		ProductID:   "coins_100",
		Listing: OneTimeProductListing{
			LanguageCode: "en-US",
			Title:        "this title is intentionally much longer than fifty five characters",
		},
		RegionsVersion:   "2022/02",
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected listing length validation error")
	}
}

func TestPatchOneTimeProductCountsListingCharacters(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = PatchOneTimeProduct(context.Background(), nil, OneTimeProductPatchOptions{
		PackageName: packageName,
		ProductID:   "coins_100",
		Listing: OneTimeProductListing{
			LanguageCode: "ja-JP",
			Title:        "コインコインコインコインコインコインコインコインコインコインコインコインコインコインコインコインコインコイン",
		},
		TitleSet:         true,
		RegionsVersion:   "2022/02",
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("PatchOneTimeProduct() error = %v", err)
	}
}

func TestPatchOneTimeProductPassesOptionsToPatcher(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	patcher := &fakeOneTimeProductClient{
		product: OneTimeProduct{
			ProductID: "coins_100",
			Listings: []OneTimeProductListing{
				{LanguageCode: "en-US", Title: "100 coins"},
			},
		},
	}
	options := OneTimeProductPatchOptions{
		PackageName: packageName,
		ProductID:   "coins_100",
		Listing: OneTimeProductListing{
			LanguageCode: "en-US",
			Title:        "100 coins",
		},
		TitleSet:         true,
		RegionsVersion:   "2022/02",
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	}

	result, err := PatchOneTimeProduct(context.Background(), patcher, options)
	if err != nil {
		t.Fatalf("PatchOneTimeProduct() error = %v", err)
	}
	if !result.Applied || result.Product == nil {
		t.Fatalf("result = %#v, want applied product", result)
	}
	if !reflect.DeepEqual(patcher.patchOptions, options) {
		t.Fatalf("patchOptions = %#v, want %#v", patcher.patchOptions, options)
	}
}

func TestBatchPatchOneTimeProductListingsDryRunBuildsPlan(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := BatchPatchOneTimeProductListings(context.Background(), nil, OneTimeProductBatchPatchListingsOptions{
		PackageName: packageName,
		Requests: []OneTimeProductBatchPatchListingRequest{
			{
				ProductID: "coins_100",
				Listing:   OneTimeProductListing{LanguageCode: "en-US", Title: "100 coins", Description: "Buy coins."},
			},
			{
				ProductID: "coins_500",
				Listing:   OneTimeProductListing{LanguageCode: "es-ES", Title: "500 monedas", Description: "Compra monedas."},
			},
		},
		RegionsVersion:   "2026/05",
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("BatchPatchOneTimeProductListings() error = %v", err)
	}
	if result.Applied || !result.DryRun || len(result.Desired) != 2 {
		t.Fatalf("result = %#v, want dry-run desired products", result)
	}
	if result.Plan.UpdateMask != oneTimeProductPatchUpdateMask || result.Plan.LatencyTolerance != ProductUpdateLatencyToleranceTolerant {
		t.Fatalf("Plan = %#v, want listing update mask and latency", result.Plan)
	}
}

func TestBatchPatchOneTimeProductListingsRejectsDuplicateProductLanguage(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	request := OneTimeProductBatchPatchListingRequest{
		ProductID: "coins_100",
		Listing:   OneTimeProductListing{LanguageCode: "en-US", Title: "100 coins", Description: "Buy coins."},
	}

	_, err = BatchPatchOneTimeProductListings(context.Background(), nil, OneTimeProductBatchPatchListingsOptions{
		PackageName:    packageName,
		Requests:       []OneTimeProductBatchPatchListingRequest{request, request},
		RegionsVersion: "2026/05",
		DryRun:         true,
	})
	if err == nil {
		t.Fatal("expected duplicate listing validation error")
	}
}

func TestBatchPatchOneTimeProductListingsRejectsMoreThanOneHundredProducts(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	requests := make([]OneTimeProductBatchPatchListingRequest, 0, 101)
	for i := range 101 {
		requests = append(requests, OneTimeProductBatchPatchListingRequest{
			ProductID: OneTimeProductID(fmt.Sprintf("coins_%03d", i)),
			Listing:   OneTimeProductListing{LanguageCode: "en-US", Title: "Coins", Description: "Buy coins."},
		})
	}

	_, err = BatchPatchOneTimeProductListings(context.Background(), nil, OneTimeProductBatchPatchListingsOptions{
		PackageName:    packageName,
		Requests:       requests,
		RegionsVersion: "2026/05",
		DryRun:         true,
	})
	if err == nil {
		t.Fatal("expected product count validation error")
	}
}

func TestBatchPatchOneTimeProductListingsAllowsMoreThanOneHundredListingsAcrossOneHundredProducts(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	requests := make([]OneTimeProductBatchPatchListingRequest, 0, 101)
	for i := range 100 {
		requests = append(requests, OneTimeProductBatchPatchListingRequest{
			ProductID: OneTimeProductID(fmt.Sprintf("coins_%03d", i)),
			Listing:   OneTimeProductListing{LanguageCode: "en-US", Title: "Coins", Description: "Buy coins."},
		})
	}
	requests = append(requests, OneTimeProductBatchPatchListingRequest{
		ProductID: "coins_000",
		Listing:   OneTimeProductListing{LanguageCode: "es-ES", Title: "Monedas", Description: "Compra monedas."},
	})

	result, err := BatchPatchOneTimeProductListings(context.Background(), nil, OneTimeProductBatchPatchListingsOptions{
		PackageName:    packageName,
		Requests:       requests,
		RegionsVersion: "2026/05",
		DryRun:         true,
	})
	if err != nil {
		t.Fatalf("BatchPatchOneTimeProductListings() error = %v", err)
	}
	if len(result.Desired) != 100 || len(result.Desired[0].Listings) != 2 {
		t.Fatalf("Desired = %#v, want 100 products and two listings on first product", result.Desired)
	}
}

func TestBatchPatchOneTimeProductListingsPassesOptionsToPatcher(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	patcher := &fakeOneTimeProductClient{
		batchPatchResult: OneTimeProductBatchPatchListingsResult{
			Products: []OneTimeProduct{{ProductID: "coins_100"}},
		},
	}

	result, err := BatchPatchOneTimeProductListings(context.Background(), patcher, OneTimeProductBatchPatchListingsOptions{
		PackageName: packageName,
		Requests: []OneTimeProductBatchPatchListingRequest{{
			ProductID: "coins_100",
			Listing:   OneTimeProductListing{LanguageCode: "en-US", Title: "100 coins", Description: "Buy coins."},
		}},
		RegionsVersion: "2026/05",
		Confirm:        true,
	})
	if err != nil {
		t.Fatalf("BatchPatchOneTimeProductListings() error = %v", err)
	}
	if !result.Applied || len(result.Products) != 1 {
		t.Fatalf("result = %#v, want applied products", result)
	}
	if patcher.batchPatchOptions.Requests[0].ProductID != "coins_100" {
		t.Fatalf("batchPatchOptions = %#v, want request passed to patcher", patcher.batchPatchOptions)
	}
}

func TestBatchDeleteOneTimeProductsDryRunBuildsPlanWithoutDeleter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := BatchDeleteOneTimeProducts(context.Background(), nil, OneTimeProductBatchDeleteOptions{
		PackageName:      packageName,
		ProductIDs:       []OneTimeProductID{"coins_100", "coins_500"},
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("BatchDeleteOneTimeProducts() error = %v", err)
	}
	if !result.DryRun || result.Deleted {
		t.Fatalf("result = %#v, want dry-run batch deletion plan", result)
	}
	if len(result.ProductIDs) != 2 || result.ProductIDs[1] != "coins_500" {
		t.Fatalf("ProductIDs = %#v, want requested product IDs", result.ProductIDs)
	}
}

func TestBatchDeleteOneTimeProductsRejectsDuplicates(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = BatchDeleteOneTimeProducts(context.Background(), nil, OneTimeProductBatchDeleteOptions{
		PackageName:      packageName,
		ProductIDs:       []OneTimeProductID{"coins_100", "coins_100"},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected duplicate product ID validation error")
	}
}

func TestBatchDeleteOneTimeProductsPassesOptionsToDeleter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	deleter := &fakeOneTimeProductClient{}

	result, err := BatchDeleteOneTimeProducts(context.Background(), deleter, OneTimeProductBatchDeleteOptions{
		PackageName:      packageName,
		ProductIDs:       []OneTimeProductID{"coins_100", "coins_500"},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("BatchDeleteOneTimeProducts() error = %v", err)
	}
	if !result.Deleted {
		t.Fatal("Deleted = false, want true")
	}
	if len(deleter.batchDeleteOptions.ProductIDs) != 2 {
		t.Fatalf("batchDeleteOptions = %#v, want two products", deleter.batchDeleteOptions)
	}
}

func TestBatchDeletePurchaseOptionsDryRunBuildsPlanWithoutDeleter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := BatchDeletePurchaseOptions(context.Background(), nil, PurchaseOptionBatchDeleteOptions{
		PackageName:     packageName,
		ParentProductID: OneTimeProductBatchParentProductID(OneTimeProductWildcardID),
		Requests: []PurchaseOptionBatchDeleteRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy"},
			{ProductID: "coins_500", PurchaseOptionID: "rent"},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		Force:            true,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("BatchDeletePurchaseOptions() error = %v", err)
	}
	if !result.DryRun || result.Deleted || !result.Force {
		t.Fatalf("result = %#v, want forced dry-run batch deletion plan", result)
	}
	if len(result.Requests) != 2 || result.Requests[1].ProductID != "coins_500" {
		t.Fatalf("Requests = %#v, want requested purchase options", result.Requests)
	}
}

func TestBatchDeletePurchaseOptionsRejectsDuplicatePurchaseOption(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = BatchDeletePurchaseOptions(context.Background(), nil, PurchaseOptionBatchDeleteOptions{
		PackageName:     packageName,
		ParentProductID: OneTimeProductBatchParentProductID(OneTimeProductWildcardID),
		Requests: []PurchaseOptionBatchDeleteRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy"},
			{ProductID: "coins_100", PurchaseOptionID: "buy"},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected duplicate purchase option validation error")
	}
}

func TestBatchDeletePurchaseOptionsRejectsRepeatedProductID(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = BatchDeletePurchaseOptions(context.Background(), nil, PurchaseOptionBatchDeleteOptions{
		PackageName:     packageName,
		ParentProductID: OneTimeProductBatchParentProductID(OneTimeProductWildcardID),
		Requests: []PurchaseOptionBatchDeleteRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy"},
			{ProductID: "coins_100", PurchaseOptionID: "rent"},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected repeated product validation error")
	}
}

func TestBatchDeletePurchaseOptionsRejectsParentMismatch(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = BatchDeletePurchaseOptions(context.Background(), nil, PurchaseOptionBatchDeleteOptions{
		PackageName:      packageName,
		ParentProductID:  "coins_100",
		Requests:         []PurchaseOptionBatchDeleteRequest{{ProductID: "coins_500", PurchaseOptionID: "buy"}},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected parent mismatch validation error")
	}
}

func TestBatchDeletePurchaseOptionsRejectsSingleProductWildcardParent(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = BatchDeletePurchaseOptions(context.Background(), nil, PurchaseOptionBatchDeleteOptions{
		PackageName:      packageName,
		ParentProductID:  OneTimeProductBatchParentProductID(OneTimeProductWildcardID),
		Requests:         []PurchaseOptionBatchDeleteRequest{{ProductID: "coins_100", PurchaseOptionID: "buy"}},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected single-product wildcard parent validation error")
	}
}

func TestBatchDeletePurchaseOptionsPassesOptionsToDeleter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	deleter := &fakeOneTimeProductClient{}
	options := PurchaseOptionBatchDeleteOptions{
		PackageName:     packageName,
		ParentProductID: OneTimeProductBatchParentProductID(OneTimeProductWildcardID),
		Requests: []PurchaseOptionBatchDeleteRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy"},
			{ProductID: "coins_500", PurchaseOptionID: "rent"},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Force:            true,
		Confirm:          true,
	}

	result, err := BatchDeletePurchaseOptions(context.Background(), deleter, options)
	if err != nil {
		t.Fatalf("BatchDeletePurchaseOptions() error = %v", err)
	}
	if !result.Deleted {
		t.Fatal("Deleted = false, want true")
	}
	if !reflect.DeepEqual(deleter.purchaseOptionBatchDeleteOptions, options) {
		t.Fatalf("purchaseOptionBatchDeleteOptions = %#v, want %#v", deleter.purchaseOptionBatchDeleteOptions, options)
	}
}

func TestBatchPatchPurchaseOptionAvailabilityDryRunBuildsPlanWithoutPatcher(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := BatchPatchPurchaseOptionAvailability(context.Background(), nil, PurchaseOptionBatchPatchAvailabilityOptions{
		PackageName: packageName,
		Requests: []PurchaseOptionAvailabilityPatchRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", RegionCode: "US", Availability: PurchaseOptionAvailabilityNoLongerAvailable},
			{ProductID: "coins_100", PurchaseOptionID: "buy", RegionCode: "FR", Availability: PurchaseOptionAvailabilityAvailable},
		},
		RegionsVersion:   "2026/05",
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("BatchPatchPurchaseOptionAvailability() error = %v", err)
	}
	if !result.DryRun || result.Applied {
		t.Fatalf("result = %#v, want dry-run availability patch", result)
	}
	if result.Plan.UpdateMask != "purchaseOptions" {
		t.Fatalf("UpdateMask = %q, want purchaseOptions", result.Plan.UpdateMask)
	}
	if len(result.Desired) != 1 || len(result.Desired[0].PurchaseOptions) != 1 || len(result.Desired[0].PurchaseOptions[0].RegionalConfigs) != 2 {
		t.Fatalf("Desired = %#v, want grouped purchase option regions", result.Desired)
	}
}

func TestBatchPatchPurchaseOptionAvailabilityRejectsDuplicateRegion(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = BatchPatchPurchaseOptionAvailability(context.Background(), nil, PurchaseOptionBatchPatchAvailabilityOptions{
		PackageName: packageName,
		Requests: []PurchaseOptionAvailabilityPatchRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", RegionCode: "US", Availability: PurchaseOptionAvailabilityAvailable},
			{ProductID: "coins_100", PurchaseOptionID: "buy", RegionCode: "US", Availability: PurchaseOptionAvailabilityNoLongerAvailable},
		},
		RegionsVersion:   "2026/05",
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected duplicate purchase option availability validation error")
	}
}

func TestBatchPatchPurchaseOptionAvailabilityAllowsMoreThanHundredRegionsForOneProduct(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	requests := make([]PurchaseOptionAvailabilityPatchRequest, 0, 101)
	for index := 0; index < 101; index++ {
		requests = append(requests, PurchaseOptionAvailabilityPatchRequest{
			ProductID:        "coins_100",
			PurchaseOptionID: "buy",
			RegionCode:       testRegionCode(index),
			Availability:     PurchaseOptionAvailabilityAvailable,
		})
	}

	_, err = NewPurchaseOptionBatchPatchAvailabilityPlan(PurchaseOptionBatchPatchAvailabilityOptions{
		PackageName:      packageName,
		Requests:         requests,
		RegionsVersion:   "2026/05",
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("NewPurchaseOptionBatchPatchAvailabilityPlan() error = %v", err)
	}
}

func TestBatchPatchPurchaseOptionAvailabilityRejectsMoreThanHundredProducts(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	requests := make([]PurchaseOptionAvailabilityPatchRequest, 0, 101)
	for index := 0; index < 101; index++ {
		requests = append(requests, PurchaseOptionAvailabilityPatchRequest{
			ProductID:        OneTimeProductID(fmt.Sprintf("coins_%d", index)),
			PurchaseOptionID: "buy",
			RegionCode:       "US",
			Availability:     PurchaseOptionAvailabilityAvailable,
		})
	}

	_, err = NewPurchaseOptionBatchPatchAvailabilityPlan(PurchaseOptionBatchPatchAvailabilityOptions{
		PackageName:      packageName,
		Requests:         requests,
		RegionsVersion:   "2026/05",
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected product count validation error")
	}
}

func TestBatchPatchPurchaseOptionAvailabilityPassesOptionsToPatcher(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	patcher := &fakeOneTimeProductClient{
		purchaseOptionAvailabilityResult: PurchaseOptionBatchPatchAvailabilityResult{
			Products: []OneTimeProduct{{ProductID: "coins_100"}},
		},
	}
	options := PurchaseOptionBatchPatchAvailabilityOptions{
		PackageName: packageName,
		Requests: []PurchaseOptionAvailabilityPatchRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", RegionCode: "US", Availability: PurchaseOptionAvailabilityNoLongerAvailable},
		},
		RegionsVersion:   "2026/05",
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	}

	result, err := BatchPatchPurchaseOptionAvailability(context.Background(), patcher, options)
	if err != nil {
		t.Fatalf("BatchPatchPurchaseOptionAvailability() error = %v", err)
	}
	if !result.Applied || len(result.Products) != 1 {
		t.Fatalf("result = %#v, want applied product", result)
	}
	if !reflect.DeepEqual(patcher.purchaseOptionAvailabilityOptions, options) {
		t.Fatalf("purchaseOptionAvailabilityOptions = %#v, want %#v", patcher.purchaseOptionAvailabilityOptions, options)
	}
}

func TestUpdatePurchaseOptionStateDryRunBuildsPlanWithoutUpdater(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := UpdatePurchaseOptionState(context.Background(), nil, PurchaseOptionStateUpdateOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		Action:           PurchaseOptionStateActionActivate,
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("UpdatePurchaseOptionState() error = %v", err)
	}
	if !result.DryRun {
		t.Fatal("DryRun = false, want true")
	}
	wantSteps := []string{"plan activate purchase option"}
	if !reflect.DeepEqual(result.Plan.Steps, wantSteps) {
		t.Fatalf("steps = %#v, want %#v", result.Plan.Steps, wantSteps)
	}
}

func TestUpdatePurchaseOptionStateRequiresConfirmOrDryRun(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = NewPurchaseOptionStateUpdatePlan(PurchaseOptionStateUpdateOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		Action:           PurchaseOptionStateActionDeactivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
	})
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
}

func TestUpdatePurchaseOptionStatePassesOptionsToUpdater(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	updater := &fakeOneTimeProductClient{
		product: OneTimeProduct{ProductID: "coins_100"},
	}

	result, err := UpdatePurchaseOptionState(context.Background(), updater, PurchaseOptionStateUpdateOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		Action:           PurchaseOptionStateActionDeactivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("UpdatePurchaseOptionState() error = %v", err)
	}
	if !result.Applied {
		t.Fatal("Applied = false, want true")
	}
	if updater.stateOptions.Action != PurchaseOptionStateActionDeactivate {
		t.Fatalf("Action = %q, want deactivate", updater.stateOptions.Action)
	}
}

func testRegionCode(index int) string {
	first := byte('A' + (index / 26))
	second := byte('A' + (index % 26))
	return string([]byte{first, second})
}

type fakeOneTimeProductClient struct {
	listOptions                       OneTimeProductListOptions
	listResult                        OneTimeProductListResult
	batchOptions                      OneTimeProductBatchGetOptions
	batchResult                       OneTimeProductBatchGetResult
	deleteOptions                     OneTimeProductDeleteOptions
	patchOptions                      OneTimeProductPatchOptions
	batchPatchOptions                 OneTimeProductBatchPatchListingsOptions
	batchPatchResult                  OneTimeProductBatchPatchListingsResult
	batchDeleteOptions                OneTimeProductBatchDeleteOptions
	purchaseOptionBatchDeleteOptions  PurchaseOptionBatchDeleteOptions
	purchaseOptionAvailabilityOptions PurchaseOptionBatchPatchAvailabilityOptions
	purchaseOptionAvailabilityResult  PurchaseOptionBatchPatchAvailabilityResult
	productID                         OneTimeProductID
	product                           OneTimeProduct
	stateOptions                      PurchaseOptionStateUpdateOptions
}

func (c *fakeOneTimeProductClient) ListOneTimeProducts(ctx context.Context, options OneTimeProductListOptions) (OneTimeProductListResult, error) {
	c.listOptions = options
	return c.listResult, nil
}

func (c *fakeOneTimeProductClient) GetOneTimeProduct(ctx context.Context, packageName PackageName, productID OneTimeProductID) (OneTimeProduct, error) {
	c.productID = productID
	return c.product, nil
}

func (c *fakeOneTimeProductClient) BatchGetOneTimeProducts(ctx context.Context, options OneTimeProductBatchGetOptions) (OneTimeProductBatchGetResult, error) {
	c.batchOptions = options
	return c.batchResult, nil
}

func (c *fakeOneTimeProductClient) DeleteOneTimeProduct(ctx context.Context, options OneTimeProductDeleteOptions) error {
	c.deleteOptions = options
	return nil
}

func (c *fakeOneTimeProductClient) PatchOneTimeProduct(ctx context.Context, options OneTimeProductPatchOptions) (OneTimeProduct, error) {
	c.patchOptions = options
	return c.product, nil
}

func (c *fakeOneTimeProductClient) BatchPatchOneTimeProductListings(ctx context.Context, options OneTimeProductBatchPatchListingsOptions) (OneTimeProductBatchPatchListingsResult, error) {
	c.batchPatchOptions = options
	return c.batchPatchResult, nil
}

func (c *fakeOneTimeProductClient) BatchDeleteOneTimeProducts(ctx context.Context, options OneTimeProductBatchDeleteOptions) error {
	c.batchDeleteOptions = options
	return nil
}

func (c *fakeOneTimeProductClient) BatchDeletePurchaseOptions(ctx context.Context, options PurchaseOptionBatchDeleteOptions) error {
	c.purchaseOptionBatchDeleteOptions = options
	return nil
}

func (c *fakeOneTimeProductClient) BatchPatchPurchaseOptionAvailability(ctx context.Context, options PurchaseOptionBatchPatchAvailabilityOptions) (PurchaseOptionBatchPatchAvailabilityResult, error) {
	c.purchaseOptionAvailabilityOptions = options
	return c.purchaseOptionAvailabilityResult, nil
}

func (c *fakeOneTimeProductClient) UpdatePurchaseOptionState(ctx context.Context, options PurchaseOptionStateUpdateOptions) (OneTimeProduct, error) {
	c.stateOptions = options
	return c.product, nil
}
