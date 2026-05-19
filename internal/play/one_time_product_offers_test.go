package play

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"testing"
)

func TestOneTimeProductOfferIDsValidateGoogleShape(t *testing.T) {
	if _, err := NewOneTimeProductPurchaseOptionID("buy-option-1"); err != nil {
		t.Fatalf("NewOneTimeProductPurchaseOptionID() error = %v", err)
	}
	if _, err := NewOneTimeProductOfferID("intro-offer-1"); err != nil {
		t.Fatalf("NewOneTimeProductOfferID() error = %v", err)
	}

	for _, value := range []string{"", "Offer", "_offer", "offer_id"} {
		t.Run(value, func(t *testing.T) {
			if _, err := NewOneTimeProductOfferID(value); err == nil {
				t.Fatal("expected offer ID validation error")
			}
		})
	}
}

func TestListOneTimeProductOffersPassesOptionsToLister(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	lister := &fakeOneTimeProductOfferClient{
		listResult: OneTimeProductOfferListResult{
			PackageName: packageName,
			ProductID:   "coins_100",
			Offers:      []OneTimeProductOffer{{OfferID: "intro"}},
		},
	}

	result, err := ListOneTimeProductOffers(context.Background(), lister, OneTimeProductOfferListOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		PageSize:         50,
		PageToken:        "next",
	})
	if err != nil {
		t.Fatalf("ListOneTimeProductOffers() error = %v", err)
	}
	if len(result.Offers) != 1 {
		t.Fatalf("len(Offers) = %d, want 1", len(result.Offers))
	}
	if !reflect.DeepEqual(lister.listOptions, OneTimeProductOfferListOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		PageSize:         50,
		PageToken:        "next",
	}) {
		t.Fatalf("listOptions = %#v", lister.listOptions)
	}
}

func TestListOneTimeProductOffersRejectsInvalidWildcardParent(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = ListOneTimeProductOffers(context.Background(), nil, OneTimeProductOfferListOptions{
		PackageName:      packageName,
		ProductID:        "-",
		PurchaseOptionID: "buy",
	})
	if err == nil {
		t.Fatal("expected wildcard validation error")
	}
}

func TestGetOneTimeProductOfferPassesOptionsToGetter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	getter := &fakeOneTimeProductOfferClient{
		offer: OneTimeProductOffer{OfferID: "intro"},
	}

	offer, err := GetOneTimeProductOffer(context.Background(), getter, OneTimeProductOfferGetOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		OfferID:          "intro",
	})
	if err != nil {
		t.Fatalf("GetOneTimeProductOffer() error = %v", err)
	}
	if offer.OfferID != "intro" {
		t.Fatalf("OfferID = %q, want intro", offer.OfferID)
	}
	if getter.getOptions.OfferID != "intro" {
		t.Fatalf("getter OfferID = %q, want intro", getter.getOptions.OfferID)
	}
}

func TestBatchGetOneTimeProductOffersPassesOptionsToGetter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	getter := &fakeOneTimeProductOfferClient{
		batchResult: OneTimeProductOfferBatchGetResult{
			PackageName: packageName,
			ProductID:   "coins_100",
			Offers:      []OneTimeProductOffer{{OfferID: "intro"}},
		},
	}
	options := OneTimeProductOfferBatchGetOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		Requests: []OneTimeProductOfferBatchGetRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro"},
		},
	}

	result, err := BatchGetOneTimeProductOffers(context.Background(), getter, options)
	if err != nil {
		t.Fatalf("BatchGetOneTimeProductOffers() error = %v", err)
	}
	if len(result.Offers) != 1 {
		t.Fatalf("len(Offers) = %d, want 1", len(result.Offers))
	}
	if !reflect.DeepEqual(getter.batchOptions, options) {
		t.Fatalf("batchOptions = %#v, want %#v", getter.batchOptions, options)
	}
}

func TestBatchGetOneTimeProductOffersRejectsParentProductMismatch(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = BatchGetOneTimeProductOffers(context.Background(), nil, OneTimeProductOfferBatchGetOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "-",
		Requests: []OneTimeProductOfferBatchGetRequest{
			{ProductID: "coins_500", PurchaseOptionID: "buy", OfferID: "intro"},
		},
	})
	if err == nil {
		t.Fatal("expected parent product mismatch validation error")
	}
}

func TestBatchGetOneTimeProductOffersRejectsParentPurchaseOptionMismatch(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = BatchGetOneTimeProductOffers(context.Background(), nil, OneTimeProductOfferBatchGetOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		Requests: []OneTimeProductOfferBatchGetRequest{
			{ProductID: "coins_100", PurchaseOptionID: "rent", OfferID: "intro"},
		},
	})
	if err == nil {
		t.Fatal("expected parent purchase option mismatch validation error")
	}
}

func TestBatchGetOneTimeProductOffersRejectsMoreThanGoogleLimit(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	requests := make([]OneTimeProductOfferBatchGetRequest, 101)
	for index := range requests {
		requests[index] = OneTimeProductOfferBatchGetRequest{
			ProductID:        "coins_100",
			PurchaseOptionID: "buy",
			OfferID:          OneTimeProductOfferID(fmt.Sprintf("offer-%d", index)),
		}
	}

	_, err = BatchGetOneTimeProductOffers(context.Background(), nil, OneTimeProductOfferBatchGetOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		Requests:         requests,
	})
	if err == nil {
		t.Fatal("expected batch limit validation error")
	}
}

func TestBatchGetOneTimeProductOffersRejectsDuplicates(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = BatchGetOneTimeProductOffers(context.Background(), nil, OneTimeProductOfferBatchGetOptions{
		PackageName:      packageName,
		ProductID:        "-",
		PurchaseOptionID: "-",
		Requests: []OneTimeProductOfferBatchGetRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro"},
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro"},
		},
	})
	if err == nil {
		t.Fatal("expected duplicate validation error")
	}
}

func TestCreateOneTimeProductOfferDryRunBuildsPlanWithoutCreator(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := CreateOneTimeProductOffer(context.Background(), nil, OneTimeProductOfferCreateOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		OfferID:          "intro",
		Offer:            validOneTimeProductOfferForCreate(),
		RegionsVersion:   "2026/05",
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("CreateOneTimeProductOffer() error = %v", err)
	}
	if !result.DryRun || result.Created {
		t.Fatalf("result = %#v, want dry-run create plan", result)
	}
	if result.Desired.ProductID != "coins_100" || result.Desired.PurchaseOptionID != "buy" || result.Desired.OfferID != "intro" {
		t.Fatalf("desired = %#v, want flag IDs", result.Desired)
	}
	if result.Desired.State != "" || result.Desired.RegionsVersion != nil {
		t.Fatalf("desired = %#v, want output-only fields cleared", result.Desired)
	}
}

func TestCreateOneTimeProductOfferPassesOptionsToCreator(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	creator := &fakeOneTimeProductOfferClient{offer: OneTimeProductOffer{OfferID: "intro"}}

	result, err := CreateOneTimeProductOffer(context.Background(), creator, OneTimeProductOfferCreateOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		OfferID:          "intro",
		Offer:            validOneTimeProductOfferForCreate(),
		RegionsVersion:   "2026/05",
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("CreateOneTimeProductOffer() error = %v", err)
	}
	if !result.Created {
		t.Fatal("Created = false, want true")
	}
	if creator.createOptions.OfferID != "intro" {
		t.Fatalf("createOptions = %#v, want intro", creator.createOptions)
	}
}

func TestDecodeOneTimeProductOfferCreateJSONAcceptsAPIShape(t *testing.T) {
	offer, err := DecodeOneTimeProductOfferCreateJSON([]byte(`{
		"packageName":"ignored.by.flags",
		"productId":"ignored_by_flags",
		"purchaseOptionId":"ignored",
		"offerId":"ignored",
		"state":"ACTIVE",
		"discountedOffer":{"startTime":"2026-06-01T00:00:00Z","endTime":"2026-07-01T00:00:00Z","redemptionLimit":"5"},
		"regionalPricingAndAvailabilityConfigs":[{"regionCode":"US","availability":"AVAILABLE","relativeDiscount":0.5}]
	}`))
	if err != nil {
		t.Fatalf("DecodeOneTimeProductOfferCreateJSON() error = %v", err)
	}
	if offer.Type != OneTimeProductOfferTypeDiscounted || offer.DiscountedOffer == nil {
		t.Fatalf("offer = %#v, want discounted offer", offer)
	}
	if offer.RegionalConfigs[0].RelativeDiscount != 0.5 {
		t.Fatalf("RegionalConfigs = %#v, want relative discount", offer.RegionalConfigs)
	}
}

func TestDecodeOneTimeProductOfferCreateJSONRejectsNestedUnknownField(t *testing.T) {
	_, err := DecodeOneTimeProductOfferCreateJSON([]byte(`{
		"discountedOffer":{"startTime":"2026-06-01T00:00:00Z","bad":true},
		"regionalConfigs":[{"regionCode":"US","availability":"available","relativeDiscount":0.5}]
	}`))
	if err == nil {
		t.Fatal("expected unknown field error")
	}
}

func TestDecodeOneTimeProductOfferCreateJSONRejectsMultipleOfferTypes(t *testing.T) {
	_, err := DecodeOneTimeProductOfferCreateJSON([]byte(`{
		"discountedOffer":{"startTime":"2026-06-01T00:00:00Z"},
		"preOrderOffer":{"startTime":"2026-06-01T00:00:00Z"},
		"regionalConfigs":[{"regionCode":"US","availability":"available","relativeDiscount":0.5}]
	}`))
	if err == nil {
		t.Fatal("expected offer type union validation error")
	}
}

func TestCreateOneTimeProductOfferRejectsInvalidBody(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	tests := []struct {
		name  string
		offer OneTimeProductOffer
	}{
		{name: "missing type", offer: OneTimeProductOffer{RegionalConfigs: validOneTimeProductOfferForCreate().RegionalConfigs}},
		{name: "missing regions", offer: OneTimeProductOffer{Type: OneTimeProductOfferTypeDiscounted, DiscountedOffer: &OneTimeProductDiscountedOffer{}}},
		{
			name: "bad relative discount",
			offer: func() OneTimeProductOffer {
				offer := validOneTimeProductOfferForCreate()
				offer.RegionalConfigs[0].RelativeDiscount = 1
				return offer
			}(),
		},
		{
			name: "multiple price modes",
			offer: func() OneTimeProductOffer {
				offer := validOneTimeProductOfferForCreate()
				offer.RegionalConfigs[0].NoOverride = true
				return offer
			}(),
		},
		{
			name: "bad preorder behavior",
			offer: OneTimeProductOffer{
				Type: OneTimeProductOfferTypePreOrder,
				PreOrderOffer: &OneTimeProductPreOrderOffer{
					StartTime:           "2026-06-01T00:00:00Z",
					EndTime:             "2026-07-01T00:00:00Z",
					ReleaseTime:         "2026-08-01T00:00:00Z",
					PriceChangeBehavior: "NOPE",
				},
				RegionalConfigs: validOneTimeProductOfferForCreate().RegionalConfigs,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := CreateOneTimeProductOffer(context.Background(), nil, OneTimeProductOfferCreateOptions{
				PackageName:      packageName,
				ProductID:        "coins_100",
				PurchaseOptionID: "buy",
				OfferID:          "intro",
				Offer:            test.offer,
				RegionsVersion:   "2026/05",
				LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
				DryRun:           true,
			})
			if err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestBatchDeleteOneTimeProductOffersDryRunBuildsPlanWithoutDeleter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := BatchDeleteOneTimeProductOffers(context.Background(), nil, OneTimeProductOfferBatchDeleteOptions{
		PackageName:      packageName,
		ProductID:        "-",
		PurchaseOptionID: "-",
		Requests: []OneTimeProductOfferBatchDeleteRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro"},
			{ProductID: "coins_500", PurchaseOptionID: "rent", OfferID: "preorder"},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("BatchDeleteOneTimeProductOffers() error = %v", err)
	}
	if !result.DryRun || result.Deleted {
		t.Fatalf("result = %#v, want dry-run batch deletion", result)
	}
}

func TestBatchDeleteOneTimeProductOffersRejectsDuplicates(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = BatchDeleteOneTimeProductOffers(context.Background(), nil, OneTimeProductOfferBatchDeleteOptions{
		PackageName:      packageName,
		ProductID:        "-",
		PurchaseOptionID: "-",
		Requests: []OneTimeProductOfferBatchDeleteRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro"},
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro"},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected duplicate validation error")
	}
}

func TestBatchDeleteOneTimeProductOffersRejectsOverbroadWildcardParent(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = BatchDeleteOneTimeProductOffers(context.Background(), nil, OneTimeProductOfferBatchDeleteOptions{
		PackageName:      packageName,
		ProductID:        "-",
		PurchaseOptionID: "-",
		Requests: []OneTimeProductOfferBatchDeleteRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro"},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected overbroad wildcard parent validation error")
	}
}

func TestBatchDeleteOneTimeProductOffersRejectsConcreteParentForMultiplePurchaseOptions(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = BatchDeleteOneTimeProductOffers(context.Background(), nil, OneTimeProductOfferBatchDeleteOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		Requests: []OneTimeProductOfferBatchDeleteRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro"},
			{ProductID: "coins_100", PurchaseOptionID: "rent", OfferID: "preorder"},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected multi-purchase-option parent validation error")
	}
}

func TestBatchDeleteOneTimeProductOffersPassesOptionsToDeleter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	deleter := &fakeOneTimeProductOfferClient{}
	options := OneTimeProductOfferBatchDeleteOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		Requests: []OneTimeProductOfferBatchDeleteRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro"},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	}

	result, err := BatchDeleteOneTimeProductOffers(context.Background(), deleter, options)
	if err != nil {
		t.Fatalf("BatchDeleteOneTimeProductOffers() error = %v", err)
	}
	if !result.Deleted {
		t.Fatal("Deleted = false, want true")
	}
	if !reflect.DeepEqual(deleter.batchDeleteOptions, options) {
		t.Fatalf("batchDeleteOptions = %#v, want %#v", deleter.batchDeleteOptions, options)
	}
}

func TestBatchUpdateOneTimeProductOfferStatesDryRunBuildsPlanWithoutUpdater(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := BatchUpdateOneTimeProductOfferStates(context.Background(), nil, OneTimeProductOfferBatchStateUpdateOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "-",
		Requests: []OneTimeProductOfferBatchDeleteRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro"},
			{ProductID: "coins_100", PurchaseOptionID: "rent", OfferID: "preorder"},
		},
		Action:           OneTimeProductOfferStateActionDeactivate,
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("BatchUpdateOneTimeProductOfferStates() error = %v", err)
	}
	if !result.DryRun || result.Applied {
		t.Fatalf("result = %#v, want dry-run batch state update", result)
	}
	if result.Action != OneTimeProductOfferStateActionDeactivate {
		t.Fatalf("Action = %q, want deactivate", result.Action)
	}
}

func TestBatchUpdateOneTimeProductOfferStatesRejectsOverbroadWildcardParent(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = BatchUpdateOneTimeProductOfferStates(context.Background(), nil, OneTimeProductOfferBatchStateUpdateOptions{
		PackageName:      packageName,
		ProductID:        "-",
		PurchaseOptionID: "-",
		Requests: []OneTimeProductOfferBatchDeleteRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro"},
		},
		Action:           OneTimeProductOfferStateActionActivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected overbroad wildcard parent validation error")
	}
}

func TestBatchUpdateOneTimeProductOfferStatesPassesOptionsToUpdater(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	updater := &fakeOneTimeProductOfferClient{
		batchStateResult: OneTimeProductOfferBatchStateUpdateResult{
			PackageName:      packageName,
			ProductID:        "coins_100",
			PurchaseOptionID: "buy",
			Offers:           []OneTimeProductOffer{{OfferID: "intro", State: "ACTIVE"}},
		},
	}
	options := OneTimeProductOfferBatchStateUpdateOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		Requests: []OneTimeProductOfferBatchDeleteRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro"},
		},
		Action:           OneTimeProductOfferStateActionActivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	}

	result, err := BatchUpdateOneTimeProductOfferStates(context.Background(), updater, options)
	if err != nil {
		t.Fatalf("BatchUpdateOneTimeProductOfferStates() error = %v", err)
	}
	if !result.Applied || len(result.Offers) != 1 {
		t.Fatalf("result = %#v, want applied offer", result)
	}
	if !reflect.DeepEqual(updater.batchStateOptions, options) {
		t.Fatalf("batchStateOptions = %#v, want %#v", updater.batchStateOptions, options)
	}
}

func TestBatchPatchOneTimeProductOfferAvailabilityDryRunBuildsPlanWithoutPatcher(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := BatchPatchOneTimeProductOfferAvailability(context.Background(), nil, OneTimeProductOfferBatchPatchAvailabilityOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		RegionsVersion:   "2026/05",
		Requests: []OneTimeProductOfferAvailabilityPatchRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro", RegionCode: "US", Availability: OneTimeProductOfferAvailabilityNoLongerAvailable},
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro", RegionCode: "FR", Availability: OneTimeProductOfferAvailabilityAvailable},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("BatchPatchOneTimeProductOfferAvailability() error = %v", err)
	}
	if !result.DryRun || result.Applied {
		t.Fatalf("result = %#v, want dry-run availability patch", result)
	}
	if result.Plan.UpdateMask != oneTimeProductOfferAvailabilityUpdateMask {
		t.Fatalf("UpdateMask = %q, want %q", result.Plan.UpdateMask, oneTimeProductOfferAvailabilityUpdateMask)
	}
	if len(result.Desired) != 1 || len(result.Desired[0].RegionalConfigs) != 2 {
		t.Fatalf("Desired = %#v, want one offer with two regional configs", result.Desired)
	}
}

func TestBatchPatchOneTimeProductOfferAvailabilityRejectsDuplicateOfferRegion(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = NewOneTimeProductOfferBatchPatchAvailabilityPlan(OneTimeProductOfferBatchPatchAvailabilityOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		RegionsVersion:   "2026/05",
		Requests: []OneTimeProductOfferAvailabilityPatchRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro", RegionCode: "US", Availability: OneTimeProductOfferAvailabilityAvailable},
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro", RegionCode: "US", Availability: OneTimeProductOfferAvailabilityNoLongerAvailable},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected duplicate offer region validation error")
	}
}

func TestBatchPatchOneTimeProductOfferAvailabilityPassesOptionsToPatcher(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	patcher := &fakeOneTimeProductOfferClient{
		batchAvailabilityResult: OneTimeProductOfferBatchPatchAvailabilityResult{
			Offers: []OneTimeProductOffer{{OfferID: "intro"}},
		},
	}
	options := OneTimeProductOfferBatchPatchAvailabilityOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		RegionsVersion:   "2026/05",
		Requests: []OneTimeProductOfferAvailabilityPatchRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro", RegionCode: "US", Availability: OneTimeProductOfferAvailabilityNoLongerAvailable},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	}

	result, err := BatchPatchOneTimeProductOfferAvailability(context.Background(), patcher, options)
	if err != nil {
		t.Fatalf("BatchPatchOneTimeProductOfferAvailability() error = %v", err)
	}
	if !result.Applied || len(result.Offers) != 1 {
		t.Fatalf("result = %#v, want applied offer", result)
	}
	if !reflect.DeepEqual(patcher.batchAvailabilityOptions, options) {
		t.Fatalf("batchAvailabilityOptions = %#v, want %#v", patcher.batchAvailabilityOptions, options)
	}
}

func TestBatchPatchOneTimeProductOfferRelativeDiscountsDryRunBuildsPlanWithoutPatcher(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := BatchPatchOneTimeProductOfferRelativeDiscounts(context.Background(), nil, OneTimeProductOfferBatchPatchRelativeDiscountsOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		RegionsVersion:   "2026/05",
		Requests: []OneTimeProductOfferRelativeDiscountPatchRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro", RegionCode: "US", RelativeDiscount: 0.5},
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro", RegionCode: "FR", RelativeDiscount: 0.25},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("BatchPatchOneTimeProductOfferRelativeDiscounts() error = %v", err)
	}
	if !result.DryRun || result.Applied {
		t.Fatalf("result = %#v, want dry-run relative discount patch", result)
	}
	if result.Plan.UpdateMask != oneTimeProductOfferRegionalConfigsUpdateMask {
		t.Fatalf("UpdateMask = %q, want %q", result.Plan.UpdateMask, oneTimeProductOfferRegionalConfigsUpdateMask)
	}
	if len(result.Desired) != 1 || len(result.Desired[0].RegionalConfigs) != 2 {
		t.Fatalf("Desired = %#v, want one offer with two regional configs", result.Desired)
	}
}

func TestBatchPatchOneTimeProductOfferRelativeDiscountsRejectsInvalidDiscount(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = NewOneTimeProductOfferBatchPatchRelativeDiscountsPlan(OneTimeProductOfferBatchPatchRelativeDiscountsOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		RegionsVersion:   "2026/05",
		Requests: []OneTimeProductOfferRelativeDiscountPatchRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro", RegionCode: "US", RelativeDiscount: math.NaN()},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected relative discount validation error")
	}
}

func TestBatchPatchOneTimeProductOfferRelativeDiscountsRejectsDuplicateOfferRegion(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = NewOneTimeProductOfferBatchPatchRelativeDiscountsPlan(OneTimeProductOfferBatchPatchRelativeDiscountsOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		RegionsVersion:   "2026/05",
		Requests: []OneTimeProductOfferRelativeDiscountPatchRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro", RegionCode: "US", RelativeDiscount: 0.5},
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro", RegionCode: "US", RelativeDiscount: 0.25},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected duplicate offer region validation error")
	}
}

func TestBatchPatchOneTimeProductOfferRelativeDiscountsPassesOptionsToPatcher(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	patcher := &fakeOneTimeProductOfferClient{
		batchRelativeDiscountResult: OneTimeProductOfferBatchPatchRelativeDiscountsResult{
			Offers: []OneTimeProductOffer{{OfferID: "intro"}},
		},
	}
	options := OneTimeProductOfferBatchPatchRelativeDiscountsOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		RegionsVersion:   "2026/05",
		Requests: []OneTimeProductOfferRelativeDiscountPatchRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro", RegionCode: "US", RelativeDiscount: 0.5},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	}

	result, err := BatchPatchOneTimeProductOfferRelativeDiscounts(context.Background(), patcher, options)
	if err != nil {
		t.Fatalf("BatchPatchOneTimeProductOfferRelativeDiscounts() error = %v", err)
	}
	if !result.Applied || len(result.Offers) != 1 {
		t.Fatalf("result = %#v, want applied offer", result)
	}
	if !reflect.DeepEqual(patcher.batchRelativeDiscountOptions, options) {
		t.Fatalf("batchRelativeDiscountOptions = %#v, want %#v", patcher.batchRelativeDiscountOptions, options)
	}
}

func TestBatchPatchOneTimeProductOfferAbsoluteDiscountsDryRunBuildsPlanWithoutPatcher(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := BatchPatchOneTimeProductOfferAbsoluteDiscounts(context.Background(), nil, OneTimeProductOfferBatchPatchAbsoluteDiscountsOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		RegionsVersion:   "2026/05",
		Requests: []OneTimeProductOfferAbsoluteDiscountPatchRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro", RegionCode: "US", AbsoluteDiscount: Money{CurrencyCode: "USD", Units: 1}},
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro", RegionCode: "FR", AbsoluteDiscount: Money{CurrencyCode: "EUR", Nanos: 500000000}},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("BatchPatchOneTimeProductOfferAbsoluteDiscounts() error = %v", err)
	}
	if !result.DryRun || result.Applied {
		t.Fatalf("result = %#v, want dry-run absolute discount patch", result)
	}
	if result.Plan.UpdateMask != oneTimeProductOfferRegionalConfigsUpdateMask {
		t.Fatalf("UpdateMask = %q, want %q", result.Plan.UpdateMask, oneTimeProductOfferRegionalConfigsUpdateMask)
	}
	if len(result.Desired) != 1 || len(result.Desired[0].RegionalConfigs) != 2 {
		t.Fatalf("Desired = %#v, want one offer with two regional configs", result.Desired)
	}
}

func TestBatchPatchOneTimeProductOfferAbsoluteDiscountsRejectsDuplicateOfferRegion(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = NewOneTimeProductOfferBatchPatchAbsoluteDiscountsPlan(OneTimeProductOfferBatchPatchAbsoluteDiscountsOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		RegionsVersion:   "2026/05",
		Requests: []OneTimeProductOfferAbsoluteDiscountPatchRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro", RegionCode: "US", AbsoluteDiscount: Money{CurrencyCode: "USD", Units: 1}},
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro", RegionCode: "US", AbsoluteDiscount: Money{CurrencyCode: "USD", Units: 2}},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected duplicate offer region validation error")
	}
}

func TestBatchPatchOneTimeProductOfferAbsoluteDiscountsPassesOptionsToPatcher(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	patcher := &fakeOneTimeProductOfferClient{
		batchAbsoluteDiscountResult: OneTimeProductOfferBatchPatchAbsoluteDiscountsResult{
			Offers: []OneTimeProductOffer{{OfferID: "intro"}},
		},
	}
	options := OneTimeProductOfferBatchPatchAbsoluteDiscountsOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		RegionsVersion:   "2026/05",
		Requests: []OneTimeProductOfferAbsoluteDiscountPatchRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro", RegionCode: "US", AbsoluteDiscount: Money{CurrencyCode: "USD", Units: 1}},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	}

	result, err := BatchPatchOneTimeProductOfferAbsoluteDiscounts(context.Background(), patcher, options)
	if err != nil {
		t.Fatalf("BatchPatchOneTimeProductOfferAbsoluteDiscounts() error = %v", err)
	}
	if !result.Applied || len(result.Offers) != 1 {
		t.Fatalf("result = %#v, want applied offer", result)
	}
	if !reflect.DeepEqual(patcher.batchAbsoluteDiscountOptions, options) {
		t.Fatalf("batchAbsoluteDiscountOptions = %#v, want %#v", patcher.batchAbsoluteDiscountOptions, options)
	}
}

func TestUpdateOneTimeProductOfferStateDryRunBuildsPlanWithoutUpdater(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := UpdateOneTimeProductOfferState(context.Background(), nil, OneTimeProductOfferStateUpdateOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		OfferID:          "intro",
		Action:           OneTimeProductOfferStateActionCancel,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("UpdateOneTimeProductOfferState() error = %v", err)
	}
	if result.Applied {
		t.Fatal("Applied = true, want dry-run result")
	}
	if !reflect.DeepEqual(result.Plan.Steps, []string{"plan cancel one-time product offer"}) {
		t.Fatalf("Steps = %#v", result.Plan.Steps)
	}
}

func TestUpdateOneTimeProductOfferStateRequiresConfirmOrDryRun(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = NewOneTimeProductOfferStateUpdatePlan(OneTimeProductOfferStateUpdateOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		OfferID:          "intro",
		Action:           OneTimeProductOfferStateActionDeactivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
	})
	if err == nil {
		t.Fatal("expected confirmation validation error")
	}
}

func TestUpdateOneTimeProductOfferStatePassesOptionsToUpdater(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	updater := &fakeOneTimeProductOfferClient{
		offer: OneTimeProductOffer{OfferID: "intro"},
	}

	result, err := UpdateOneTimeProductOfferState(context.Background(), updater, OneTimeProductOfferStateUpdateOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		OfferID:          "intro",
		Action:           OneTimeProductOfferStateActionActivate,
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("UpdateOneTimeProductOfferState() error = %v", err)
	}
	if !result.Applied {
		t.Fatal("Applied = false, want applied result")
	}
	if updater.stateOptions.Action != OneTimeProductOfferStateActionActivate {
		t.Fatalf("Action = %q, want activate", updater.stateOptions.Action)
	}
	if updater.stateOptions.LatencyTolerance != ProductUpdateLatencyToleranceTolerant {
		t.Fatalf("LatencyTolerance = %q, want tolerant", updater.stateOptions.LatencyTolerance)
	}
}

func validOneTimeProductOfferForCreate() OneTimeProductOffer {
	return OneTimeProductOffer{
		Type:      OneTimeProductOfferTypeDiscounted,
		OfferTags: []string{"public"},
		DiscountedOffer: &OneTimeProductDiscountedOffer{
			StartTime:       "2026-06-01T00:00:00Z",
			EndTime:         "2026-07-01T00:00:00Z",
			RedemptionLimit: 5,
		},
		RegionalConfigs: []OneTimeProductOfferRegion{{
			RegionCode:       "US",
			Availability:     "available",
			RelativeDiscount: 0.5,
		}},
	}
}

type fakeOneTimeProductOfferClient struct {
	listOptions                  OneTimeProductOfferListOptions
	listResult                   OneTimeProductOfferListResult
	batchOptions                 OneTimeProductOfferBatchGetOptions
	batchResult                  OneTimeProductOfferBatchGetResult
	createOptions                OneTimeProductOfferCreateOptions
	batchDeleteOptions           OneTimeProductOfferBatchDeleteOptions
	batchStateOptions            OneTimeProductOfferBatchStateUpdateOptions
	batchStateResult             OneTimeProductOfferBatchStateUpdateResult
	batchAvailabilityOptions     OneTimeProductOfferBatchPatchAvailabilityOptions
	batchAvailabilityResult      OneTimeProductOfferBatchPatchAvailabilityResult
	batchRelativeDiscountOptions OneTimeProductOfferBatchPatchRelativeDiscountsOptions
	batchRelativeDiscountResult  OneTimeProductOfferBatchPatchRelativeDiscountsResult
	batchAbsoluteDiscountOptions OneTimeProductOfferBatchPatchAbsoluteDiscountsOptions
	batchAbsoluteDiscountResult  OneTimeProductOfferBatchPatchAbsoluteDiscountsResult
	getOptions                   OneTimeProductOfferGetOptions
	stateOptions                 OneTimeProductOfferStateUpdateOptions
	offer                        OneTimeProductOffer
}

func (c *fakeOneTimeProductOfferClient) ListOneTimeProductOffers(ctx context.Context, options OneTimeProductOfferListOptions) (OneTimeProductOfferListResult, error) {
	c.listOptions = options
	return c.listResult, nil
}

func (c *fakeOneTimeProductOfferClient) GetOneTimeProductOffer(ctx context.Context, options OneTimeProductOfferGetOptions) (OneTimeProductOffer, error) {
	c.getOptions = options
	return c.offer, nil
}

func (c *fakeOneTimeProductOfferClient) BatchGetOneTimeProductOffers(ctx context.Context, options OneTimeProductOfferBatchGetOptions) (OneTimeProductOfferBatchGetResult, error) {
	c.batchOptions = options
	return c.batchResult, nil
}

func (c *fakeOneTimeProductOfferClient) CreateOneTimeProductOffer(ctx context.Context, options OneTimeProductOfferCreateOptions) (OneTimeProductOffer, error) {
	c.createOptions = options
	return c.offer, nil
}

func (c *fakeOneTimeProductOfferClient) BatchDeleteOneTimeProductOffers(ctx context.Context, options OneTimeProductOfferBatchDeleteOptions) error {
	c.batchDeleteOptions = options
	return nil
}

func (c *fakeOneTimeProductOfferClient) BatchUpdateOneTimeProductOfferStates(ctx context.Context, options OneTimeProductOfferBatchStateUpdateOptions) (OneTimeProductOfferBatchStateUpdateResult, error) {
	c.batchStateOptions = options
	return c.batchStateResult, nil
}

func (c *fakeOneTimeProductOfferClient) BatchPatchOneTimeProductOfferAvailability(ctx context.Context, options OneTimeProductOfferBatchPatchAvailabilityOptions) (OneTimeProductOfferBatchPatchAvailabilityResult, error) {
	c.batchAvailabilityOptions = options
	return c.batchAvailabilityResult, nil
}

func (c *fakeOneTimeProductOfferClient) BatchPatchOneTimeProductOfferRelativeDiscounts(ctx context.Context, options OneTimeProductOfferBatchPatchRelativeDiscountsOptions) (OneTimeProductOfferBatchPatchRelativeDiscountsResult, error) {
	c.batchRelativeDiscountOptions = options
	return c.batchRelativeDiscountResult, nil
}

func (c *fakeOneTimeProductOfferClient) BatchPatchOneTimeProductOfferAbsoluteDiscounts(ctx context.Context, options OneTimeProductOfferBatchPatchAbsoluteDiscountsOptions) (OneTimeProductOfferBatchPatchAbsoluteDiscountsResult, error) {
	c.batchAbsoluteDiscountOptions = options
	return c.batchAbsoluteDiscountResult, nil
}

func (c *fakeOneTimeProductOfferClient) UpdateOneTimeProductOfferState(ctx context.Context, options OneTimeProductOfferStateUpdateOptions) (OneTimeProductOffer, error) {
	c.stateOptions = options
	return c.offer, nil
}
