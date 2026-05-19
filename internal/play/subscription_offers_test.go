package play

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"testing"
)

func TestListSubscriptionOffersPassesOptionsToLister(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	lister := &fakeSubscriptionOfferClient{
		listResult: SubscriptionOfferListResult{
			PackageName: packageName,
			ProductID:   "premium",
			BasePlanID:  "monthly",
			Offers:      []SubscriptionOffer{{OfferID: "intro"}},
		},
	}

	result, err := ListSubscriptionOffers(context.Background(), lister, SubscriptionOfferListOptions{
		PackageName: packageName,
		ProductID:   "premium",
		BasePlanID:  "monthly",
		PageSize:    50,
		PageToken:   "next",
	})
	if err != nil {
		t.Fatalf("ListSubscriptionOffers() error = %v", err)
	}
	if len(result.Offers) != 1 {
		t.Fatalf("len(Offers) = %d, want 1", len(result.Offers))
	}
	if !reflect.DeepEqual(lister.listOptions, SubscriptionOfferListOptions{
		PackageName: packageName,
		ProductID:   "premium",
		BasePlanID:  "monthly",
		PageSize:    50,
		PageToken:   "next",
	}) {
		t.Fatalf("listOptions = %#v", lister.listOptions)
	}
}

func TestListSubscriptionOffersRejectsPageSizeAboveGoogleLimit(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = ListSubscriptionOffers(context.Background(), nil, SubscriptionOfferListOptions{
		PackageName: packageName,
		ProductID:   "premium",
		BasePlanID:  "monthly",
		PageSize:    1001,
	})
	if err == nil {
		t.Fatal("expected page size validation error")
	}
}

func TestListSubscriptionOffersAllowsGoogleWildcards(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	lister := &fakeSubscriptionOfferClient{}

	_, err = ListSubscriptionOffers(context.Background(), lister, SubscriptionOfferListOptions{
		PackageName: packageName,
		ProductID:   "-",
		BasePlanID:  "-",
	})
	if err != nil {
		t.Fatalf("ListSubscriptionOffers() error = %v", err)
	}
}

func TestListSubscriptionOffersRequiresBasePlanWildcardWithProductWildcard(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = ListSubscriptionOffers(context.Background(), nil, SubscriptionOfferListOptions{
		PackageName: packageName,
		ProductID:   "-",
		BasePlanID:  "monthly",
	})
	if err == nil {
		t.Fatal("expected wildcard validation error")
	}
}

func TestGetSubscriptionOfferRejectsWildcardBasePlan(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = GetSubscriptionOffer(context.Background(), nil, SubscriptionOfferGetOptions{
		PackageName: packageName,
		ProductID:   "premium",
		BasePlanID:  "-",
		OfferID:     "intro",
	})
	if err == nil {
		t.Fatal("expected base plan validation error")
	}
}

func TestGetSubscriptionOfferPassesIDsToGetter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	getter := &fakeSubscriptionOfferClient{offer: SubscriptionOffer{OfferID: "intro"}}

	offer, err := GetSubscriptionOffer(context.Background(), getter, SubscriptionOfferGetOptions{
		PackageName: packageName,
		ProductID:   "premium",
		BasePlanID:  "monthly",
		OfferID:     "intro",
	})
	if err != nil {
		t.Fatalf("GetSubscriptionOffer() error = %v", err)
	}
	if offer.OfferID != "intro" {
		t.Fatalf("OfferID = %q, want intro", offer.OfferID)
	}
	if getter.offerID != "intro" {
		t.Fatalf("getter offerID = %q, want intro", getter.offerID)
	}
}

func TestSubscriptionOfferIDValidatesGoogleShape(t *testing.T) {
	if _, err := NewSubscriptionOfferID("intro-offer-1"); err != nil {
		t.Fatalf("NewSubscriptionOfferID() error = %v", err)
	}
	for _, value := range []string{"", "Intro", "intro_offer", "intro offer", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"} {
		if _, err := NewSubscriptionOfferID(value); err == nil {
			t.Fatalf("NewSubscriptionOfferID(%q) succeeded, want error", value)
		}
	}
}

func TestBatchGetSubscriptionOffersPassesOptionsToGetter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	getter := &fakeSubscriptionOfferClient{
		batchResult: SubscriptionOfferBatchGetResult{
			PackageName: packageName,
			Offers:      []SubscriptionOffer{{OfferID: "intro"}},
		},
	}
	options := SubscriptionOfferBatchGetOptions{
		PackageName: packageName,
		ProductID:   "-",
		BasePlanID:  "-",
		Requests: []SubscriptionOfferBatchGetRequest{
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro"},
		},
	}

	result, err := BatchGetSubscriptionOffers(context.Background(), getter, options)
	if err != nil {
		t.Fatalf("BatchGetSubscriptionOffers() error = %v", err)
	}
	if len(result.Offers) != 1 {
		t.Fatalf("len(Offers) = %d, want 1", len(result.Offers))
	}
	if !reflect.DeepEqual(getter.batchOptions, options) {
		t.Fatalf("batchOptions = %#v, want %#v", getter.batchOptions, options)
	}
}

func TestBatchGetSubscriptionOffersRejectsParentProductMismatch(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = BatchGetSubscriptionOffers(context.Background(), nil, SubscriptionOfferBatchGetOptions{
		PackageName: packageName,
		ProductID:   "premium",
		BasePlanID:  "monthly",
		Requests: []SubscriptionOfferBatchGetRequest{
			{ProductID: "other", BasePlanID: "monthly", OfferID: "intro"},
		},
	})
	if err == nil {
		t.Fatal("expected parent product mismatch validation error")
	}
}

func TestBatchGetSubscriptionOffersRejectsParentBasePlanMismatch(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = BatchGetSubscriptionOffers(context.Background(), nil, SubscriptionOfferBatchGetOptions{
		PackageName: packageName,
		ProductID:   "premium",
		BasePlanID:  "monthly",
		Requests: []SubscriptionOfferBatchGetRequest{
			{ProductID: "premium", BasePlanID: "annual", OfferID: "intro"},
		},
	})
	if err == nil {
		t.Fatal("expected parent base plan mismatch validation error")
	}
}

func TestBatchGetSubscriptionOffersRejectsDuplicates(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = BatchGetSubscriptionOffers(context.Background(), nil, SubscriptionOfferBatchGetOptions{
		PackageName: packageName,
		ProductID:   "-",
		BasePlanID:  "-",
		Requests: []SubscriptionOfferBatchGetRequest{
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro"},
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro"},
		},
	})
	if err == nil {
		t.Fatal("expected duplicate offer validation error")
	}
}

func TestNewSubscriptionOfferBatchGetRequestParsesPath(t *testing.T) {
	request, err := NewSubscriptionOfferBatchGetRequest("premium/monthly/intro")
	if err != nil {
		t.Fatalf("NewSubscriptionOfferBatchGetRequest() error = %v", err)
	}
	if request.ProductID != "premium" || request.BasePlanID != "monthly" || request.OfferID != "intro" {
		t.Fatalf("request = %#v", request)
	}
}

func TestUpdateSubscriptionOfferStateDryRunBuildsPlanWithoutUpdater(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := UpdateSubscriptionOfferState(context.Background(), nil, SubscriptionOfferStateUpdateOptions{
		PackageName:      packageName,
		ProductID:        "premium",
		BasePlanID:       "monthly",
		OfferID:          "intro",
		Action:           SubscriptionOfferStateActionActivate,
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("UpdateSubscriptionOfferState() error = %v", err)
	}
	if !result.DryRun {
		t.Fatal("DryRun = false, want true")
	}
	wantSteps := []string{"plan activate subscription offer"}
	if !reflect.DeepEqual(result.Plan.Steps, wantSteps) {
		t.Fatalf("steps = %#v, want %#v", result.Plan.Steps, wantSteps)
	}
}

func TestBatchUpdateSubscriptionOfferStatesDryRunBuildsPlanWithoutUpdater(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := BatchUpdateSubscriptionOfferStates(context.Background(), nil, SubscriptionOfferBatchStateUpdateOptions{
		PackageName: packageName,
		ProductID:   "premium",
		BasePlanID:  "-",
		Requests: []SubscriptionOfferBatchMutationRequest{
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro"},
			{ProductID: "premium", BasePlanID: "annual", OfferID: "winback"},
		},
		Action:           SubscriptionOfferStateActionDeactivate,
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("BatchUpdateSubscriptionOfferStates() error = %v", err)
	}
	if !result.DryRun || result.Applied {
		t.Fatalf("result = %#v, want dry-run batch state update", result)
	}
	if result.Action != SubscriptionOfferStateActionDeactivate {
		t.Fatalf("Action = %q, want deactivate", result.Action)
	}
}

func TestBatchUpdateSubscriptionOfferStatesRejectsOverbroadWildcardParent(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = BatchUpdateSubscriptionOfferStates(context.Background(), nil, SubscriptionOfferBatchStateUpdateOptions{
		PackageName: packageName,
		ProductID:   "-",
		BasePlanID:  "-",
		Requests: []SubscriptionOfferBatchMutationRequest{
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro"},
		},
		Action:           SubscriptionOfferStateActionActivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected overbroad wildcard parent validation error")
	}
}

func TestBatchUpdateSubscriptionOfferStatesAllowsSharedBasePlanAcrossProducts(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = BatchUpdateSubscriptionOfferStates(context.Background(), nil, SubscriptionOfferBatchStateUpdateOptions{
		PackageName: packageName,
		ProductID:   "-",
		BasePlanID:  "monthly",
		Requests: []SubscriptionOfferBatchMutationRequest{
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro"},
			{ProductID: "vip", BasePlanID: "monthly", OfferID: "winback"},
		},
		Action:           SubscriptionOfferStateActionActivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("BatchUpdateSubscriptionOfferStates() error = %v", err)
	}
}

func TestBatchUpdateSubscriptionOfferStatesPassesOptionsToUpdater(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	updater := &fakeSubscriptionOfferClient{
		batchStateResult: SubscriptionOfferBatchStateUpdateResult{
			Offers: []SubscriptionOffer{{OfferID: "intro", State: SubscriptionOfferStateActive}},
		},
	}
	options := SubscriptionOfferBatchStateUpdateOptions{
		PackageName: packageName,
		ProductID:   "premium",
		BasePlanID:  "monthly",
		Requests: []SubscriptionOfferBatchMutationRequest{
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro"},
		},
		Action:           SubscriptionOfferStateActionActivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	}

	result, err := BatchUpdateSubscriptionOfferStates(context.Background(), updater, options)
	if err != nil {
		t.Fatalf("BatchUpdateSubscriptionOfferStates() error = %v", err)
	}
	if !result.Applied || len(result.Offers) != 1 {
		t.Fatalf("result = %#v, want applied offer", result)
	}
	if !reflect.DeepEqual(updater.batchStateOptions, options) {
		t.Fatalf("batchStateOptions = %#v, want %#v", updater.batchStateOptions, options)
	}
}

func TestBatchPatchSubscriptionOfferAvailabilityDryRunBuildsPlanWithoutPatcher(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := BatchPatchSubscriptionOfferAvailability(context.Background(), nil, SubscriptionOfferBatchPatchAvailabilityOptions{
		PackageName:    packageName,
		ProductID:      "premium",
		BasePlanID:     "monthly",
		RegionsVersion: "2026/05",
		Requests: []SubscriptionOfferAvailabilityPatchRequest{
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", RegionCode: "US", Availability: false},
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", RegionCode: "FR", Availability: true},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("BatchPatchSubscriptionOfferAvailability() error = %v", err)
	}
	if !result.DryRun || result.Applied {
		t.Fatalf("result = %#v, want dry-run availability patch", result)
	}
	if result.Plan.UpdateMask != subscriptionOfferAvailabilityUpdateMask {
		t.Fatalf("UpdateMask = %q, want %q", result.Plan.UpdateMask, subscriptionOfferAvailabilityUpdateMask)
	}
	if len(result.Desired) != 1 || len(result.Desired[0].RegionalConfigs) != 2 {
		t.Fatalf("Desired = %#v, want one offer with two regional configs", result.Desired)
	}
	if result.Desired[0].RegionalConfigs[0].NewSubscriberAvailability {
		t.Fatalf("first availability = true, want explicit false")
	}
}

func TestBatchPatchSubscriptionOfferAvailabilityRejectsDuplicateOfferRegion(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = NewSubscriptionOfferBatchPatchAvailabilityPlan(SubscriptionOfferBatchPatchAvailabilityOptions{
		PackageName:    packageName,
		ProductID:      "premium",
		BasePlanID:     "monthly",
		RegionsVersion: "2026/05",
		Requests: []SubscriptionOfferAvailabilityPatchRequest{
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", RegionCode: "US", Availability: true},
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", RegionCode: "US", Availability: false},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected duplicate offer region validation error")
	}
}

func TestBatchPatchSubscriptionOfferAvailabilityCountsDistinctOffersForBatchLimit(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	requests := []SubscriptionOfferAvailabilityPatchRequest{
		{ProductID: "premium0", BasePlanID: "monthly", OfferID: "intro", RegionCode: "US", Availability: true},
		{ProductID: "premium0", BasePlanID: "monthly", OfferID: "intro", RegionCode: "FR", Availability: false},
	}
	for index := 1; index < 100; index++ {
		requests = append(requests, SubscriptionOfferAvailabilityPatchRequest{
			ProductID:    SubscriptionProductID(fmt.Sprintf("premium%d", index)),
			BasePlanID:   "monthly",
			OfferID:      "intro",
			RegionCode:   "US",
			Availability: true,
		})
	}
	_, err = NewSubscriptionOfferBatchPatchAvailabilityPlan(SubscriptionOfferBatchPatchAvailabilityOptions{
		PackageName:      packageName,
		ProductID:        "-",
		BasePlanID:       "monthly",
		RegionsVersion:   "2026/05",
		Requests:         requests,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("NewSubscriptionOfferBatchPatchAvailabilityPlan() error = %v, want 101 patches across 100 offers allowed", err)
	}

	requests = requests[:0]
	for index := 0; index < 101; index++ {
		requests = append(requests, SubscriptionOfferAvailabilityPatchRequest{
			ProductID:    SubscriptionProductID(fmt.Sprintf("premium%d", index)),
			BasePlanID:   "monthly",
			OfferID:      "intro",
			RegionCode:   "US",
			Availability: true,
		})
	}
	_, err = NewSubscriptionOfferBatchPatchAvailabilityPlan(SubscriptionOfferBatchPatchAvailabilityOptions{
		PackageName:      packageName,
		ProductID:        "-",
		BasePlanID:       "monthly",
		RegionsVersion:   "2026/05",
		Requests:         requests,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected over 100 distinct offers validation error")
	}
}

func TestBatchPatchSubscriptionOfferAvailabilityPassesOptionsToPatcher(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	patcher := &fakeSubscriptionOfferClient{
		batchAvailabilityResult: SubscriptionOfferBatchPatchAvailabilityResult{
			Offers: []SubscriptionOffer{{OfferID: "intro"}},
		},
	}
	options := SubscriptionOfferBatchPatchAvailabilityOptions{
		PackageName:    packageName,
		ProductID:      "premium",
		BasePlanID:     "monthly",
		RegionsVersion: "2026/05",
		Requests: []SubscriptionOfferAvailabilityPatchRequest{
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", RegionCode: "US", Availability: false},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	}

	result, err := BatchPatchSubscriptionOfferAvailability(context.Background(), patcher, options)
	if err != nil {
		t.Fatalf("BatchPatchSubscriptionOfferAvailability() error = %v", err)
	}
	if !result.Applied || len(result.Offers) != 1 {
		t.Fatalf("result = %#v, want applied offer", result)
	}
	if !reflect.DeepEqual(patcher.batchAvailabilityOptions, options) {
		t.Fatalf("batchAvailabilityOptions = %#v, want %#v", patcher.batchAvailabilityOptions, options)
	}
}

func TestBatchPatchSubscriptionOfferPhaseRelativeDiscountsDryRunBuildsPlanWithoutPatcher(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := BatchPatchSubscriptionOfferPhaseRelativeDiscounts(context.Background(), nil, SubscriptionOfferBatchPatchPhaseRelativeDiscountsOptions{
		PackageName:    packageName,
		ProductID:      "premium",
		BasePlanID:     "monthly",
		RegionsVersion: "2026/05",
		Requests: []SubscriptionOfferPhaseRelativeDiscountPatchRequest{
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: 0, RegionCode: "US", RelativeDiscount: 0.75},
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: 1, RegionCode: "FR", RelativeDiscount: 0.5},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("BatchPatchSubscriptionOfferPhaseRelativeDiscounts() error = %v", err)
	}
	if !result.DryRun || result.Applied {
		t.Fatalf("result = %#v, want dry-run phase relative discount patch", result)
	}
	if result.Plan.UpdateMask != subscriptionOfferPhasesUpdateMask {
		t.Fatalf("UpdateMask = %q, want %q", result.Plan.UpdateMask, subscriptionOfferPhasesUpdateMask)
	}
	if len(result.Desired) != 1 || len(result.Desired[0].Phases) != 2 {
		t.Fatalf("Desired = %#v, want one offer with two phase patches", result.Desired)
	}
}

func TestBatchPatchSubscriptionOfferPhaseRelativeDiscountsRejectsInvalidPatch(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	tests := []struct {
		name    string
		request SubscriptionOfferPhaseRelativeDiscountPatchRequest
	}{
		{
			name:    "negative phase index",
			request: SubscriptionOfferPhaseRelativeDiscountPatchRequest{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: -1, RegionCode: "US", RelativeDiscount: 0.75},
		},
		{
			name:    "zero discount",
			request: SubscriptionOfferPhaseRelativeDiscountPatchRequest{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: 0, RegionCode: "US", RelativeDiscount: 0},
		},
		{
			name:    "one discount",
			request: SubscriptionOfferPhaseRelativeDiscountPatchRequest{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: 0, RegionCode: "US", RelativeDiscount: 1},
		},
		{
			name:    "negative discount",
			request: SubscriptionOfferPhaseRelativeDiscountPatchRequest{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: 0, RegionCode: "US", RelativeDiscount: -0.1},
		},
		{
			name:    "nan discount",
			request: SubscriptionOfferPhaseRelativeDiscountPatchRequest{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: 0, RegionCode: "US", RelativeDiscount: math.NaN()},
		},
		{
			name:    "infinite discount",
			request: SubscriptionOfferPhaseRelativeDiscountPatchRequest{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: 0, RegionCode: "US", RelativeDiscount: math.Inf(1)},
		},
		{
			name:    "invalid region",
			request: SubscriptionOfferPhaseRelativeDiscountPatchRequest{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: 0, RegionCode: "USA", RelativeDiscount: 0.75},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err = NewSubscriptionOfferBatchPatchPhaseRelativeDiscountsPlan(SubscriptionOfferBatchPatchPhaseRelativeDiscountsOptions{
				PackageName:      packageName,
				ProductID:        "premium",
				BasePlanID:       "monthly",
				RegionsVersion:   "2026/05",
				Requests:         []SubscriptionOfferPhaseRelativeDiscountPatchRequest{test.request},
				LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
				DryRun:           true,
			})
			if err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestBatchPatchSubscriptionOfferPhaseRelativeDiscountsRejectsDuplicateOfferPhaseRegion(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = NewSubscriptionOfferBatchPatchPhaseRelativeDiscountsPlan(SubscriptionOfferBatchPatchPhaseRelativeDiscountsOptions{
		PackageName:    packageName,
		ProductID:      "premium",
		BasePlanID:     "monthly",
		RegionsVersion: "2026/05",
		Requests: []SubscriptionOfferPhaseRelativeDiscountPatchRequest{
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: 0, RegionCode: "US", RelativeDiscount: 0.75},
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: 0, RegionCode: "US", RelativeDiscount: 0.5},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected duplicate offer phase region validation error")
	}
}

func TestBatchPatchSubscriptionOfferPhaseRelativeDiscountsPassesOptionsToPatcher(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	patcher := &fakeSubscriptionOfferClient{
		batchPhaseRelativeDiscountResult: SubscriptionOfferBatchPatchPhaseRelativeDiscountsResult{
			Offers: []SubscriptionOffer{{OfferID: "intro"}},
		},
	}
	options := SubscriptionOfferBatchPatchPhaseRelativeDiscountsOptions{
		PackageName:    packageName,
		ProductID:      "premium",
		BasePlanID:     "monthly",
		RegionsVersion: "2026/05",
		Requests: []SubscriptionOfferPhaseRelativeDiscountPatchRequest{
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: 0, RegionCode: "US", RelativeDiscount: 0.75},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	}

	result, err := BatchPatchSubscriptionOfferPhaseRelativeDiscounts(context.Background(), patcher, options)
	if err != nil {
		t.Fatalf("BatchPatchSubscriptionOfferPhaseRelativeDiscounts() error = %v", err)
	}
	if !result.Applied || len(result.Offers) != 1 {
		t.Fatalf("result = %#v, want applied offer", result)
	}
	if !reflect.DeepEqual(patcher.batchPhaseRelativeDiscountOptions, options) {
		t.Fatalf("batchPhaseRelativeDiscountOptions = %#v, want %#v", patcher.batchPhaseRelativeDiscountOptions, options)
	}
}

func TestBatchPatchSubscriptionOfferPhaseAbsoluteDiscountsDryRunBuildsPlanWithoutPatcher(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := BatchPatchSubscriptionOfferPhaseAbsoluteDiscounts(context.Background(), nil, SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsOptions{
		PackageName:    packageName,
		ProductID:      "premium",
		BasePlanID:     "monthly",
		RegionsVersion: "2026/05",
		Requests: []SubscriptionOfferPhaseAbsoluteDiscountPatchRequest{
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: 0, RegionCode: "US", AbsoluteDiscount: Money{CurrencyCode: "USD", Units: 1}},
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: 1, RegionCode: "FR", AbsoluteDiscount: Money{CurrencyCode: "EUR", Nanos: 500000000}},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("BatchPatchSubscriptionOfferPhaseAbsoluteDiscounts() error = %v", err)
	}
	if !result.DryRun || result.Applied {
		t.Fatalf("result = %#v, want dry-run phase absolute discount patch", result)
	}
	if result.Plan.UpdateMask != subscriptionOfferPhasesUpdateMask {
		t.Fatalf("UpdateMask = %q, want %q", result.Plan.UpdateMask, subscriptionOfferPhasesUpdateMask)
	}
	if len(result.Desired) != 1 || len(result.Desired[0].Phases) != 2 {
		t.Fatalf("Desired = %#v, want one offer with two phase patches", result.Desired)
	}
}

func TestBatchPatchSubscriptionOfferPhaseAbsoluteDiscountsRejectsInvalidPatch(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	tests := []struct {
		name    string
		request SubscriptionOfferPhaseAbsoluteDiscountPatchRequest
	}{
		{
			name:    "negative phase index",
			request: SubscriptionOfferPhaseAbsoluteDiscountPatchRequest{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: -1, RegionCode: "US", AbsoluteDiscount: Money{CurrencyCode: "USD", Units: 1}},
		},
		{
			name:    "invalid region",
			request: SubscriptionOfferPhaseAbsoluteDiscountPatchRequest{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: 0, RegionCode: "USA", AbsoluteDiscount: Money{CurrencyCode: "USD", Units: 1}},
		},
		{
			name:    "zero money",
			request: SubscriptionOfferPhaseAbsoluteDiscountPatchRequest{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: 0, RegionCode: "US", AbsoluteDiscount: Money{CurrencyCode: "USD"}},
		},
		{
			name:    "negative units",
			request: SubscriptionOfferPhaseAbsoluteDiscountPatchRequest{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: 0, RegionCode: "US", AbsoluteDiscount: Money{CurrencyCode: "USD", Units: -1}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err = NewSubscriptionOfferBatchPatchPhaseAbsoluteDiscountsPlan(SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsOptions{
				PackageName:      packageName,
				ProductID:        "premium",
				BasePlanID:       "monthly",
				RegionsVersion:   "2026/05",
				Requests:         []SubscriptionOfferPhaseAbsoluteDiscountPatchRequest{test.request},
				LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
				DryRun:           true,
			})
			if err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestBatchPatchSubscriptionOfferPhaseAbsoluteDiscountsRejectsDuplicateOfferPhaseRegion(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = NewSubscriptionOfferBatchPatchPhaseAbsoluteDiscountsPlan(SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsOptions{
		PackageName:    packageName,
		ProductID:      "premium",
		BasePlanID:     "monthly",
		RegionsVersion: "2026/05",
		Requests: []SubscriptionOfferPhaseAbsoluteDiscountPatchRequest{
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: 0, RegionCode: "US", AbsoluteDiscount: Money{CurrencyCode: "USD", Units: 1}},
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: 0, RegionCode: "US", AbsoluteDiscount: Money{CurrencyCode: "USD", Units: 2}},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected duplicate offer phase region validation error")
	}
}

func TestBatchPatchSubscriptionOfferPhaseAbsoluteDiscountsPassesOptionsToPatcher(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	patcher := &fakeSubscriptionOfferClient{
		batchPhaseAbsoluteDiscountResult: SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsResult{
			Offers: []SubscriptionOffer{{OfferID: "intro"}},
		},
	}
	options := SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsOptions{
		PackageName:    packageName,
		ProductID:      "premium",
		BasePlanID:     "monthly",
		RegionsVersion: "2026/05",
		Requests: []SubscriptionOfferPhaseAbsoluteDiscountPatchRequest{
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: 0, RegionCode: "US", AbsoluteDiscount: Money{CurrencyCode: "USD", Units: 1}},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	}

	result, err := BatchPatchSubscriptionOfferPhaseAbsoluteDiscounts(context.Background(), patcher, options)
	if err != nil {
		t.Fatalf("BatchPatchSubscriptionOfferPhaseAbsoluteDiscounts() error = %v", err)
	}
	if !result.Applied || len(result.Offers) != 1 {
		t.Fatalf("result = %#v, want applied offer", result)
	}
	if !reflect.DeepEqual(patcher.batchPhaseAbsoluteDiscountOptions, options) {
		t.Fatalf("batchPhaseAbsoluteDiscountOptions = %#v, want %#v", patcher.batchPhaseAbsoluteDiscountOptions, options)
	}
}

func TestDeleteSubscriptionOfferDryRunBuildsPlanWithoutDeleter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := DeleteSubscriptionOffer(context.Background(), nil, SubscriptionOfferDeleteOptions{
		PackageName: packageName,
		ProductID:   "premium",
		BasePlanID:  "monthly",
		OfferID:     "intro",
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("DeleteSubscriptionOffer() error = %v", err)
	}
	if !result.DryRun || result.Deleted {
		t.Fatalf("result = %#v, want dry-run deletion plan", result)
	}
	if !reflect.DeepEqual(result.Plan.Steps, []string{"delete subscription offer"}) {
		t.Fatalf("steps = %#v, want delete step", result.Plan.Steps)
	}
}

func TestDeleteSubscriptionOfferRequiresConfirmOrDryRun(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = DeleteSubscriptionOffer(context.Background(), nil, SubscriptionOfferDeleteOptions{
		PackageName: packageName,
		ProductID:   "premium",
		BasePlanID:  "monthly",
		OfferID:     "intro",
	})
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
}

func TestDeleteSubscriptionOfferPassesOptionsToDeleter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	deleter := &fakeSubscriptionOfferClient{}

	result, err := DeleteSubscriptionOffer(context.Background(), deleter, SubscriptionOfferDeleteOptions{
		PackageName: packageName,
		ProductID:   "premium",
		BasePlanID:  "monthly",
		OfferID:     "intro",
		Confirm:     true,
	})
	if err != nil {
		t.Fatalf("DeleteSubscriptionOffer() error = %v", err)
	}
	if !result.Deleted {
		t.Fatal("Deleted = false, want true")
	}
	if deleter.deleteOptions.OfferID != "intro" {
		t.Fatalf("deleteOptions = %#v, want intro", deleter.deleteOptions)
	}
}

func TestUpdateSubscriptionOfferStateRequiresConfirmOrDryRun(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = NewSubscriptionOfferStateUpdatePlan(SubscriptionOfferStateUpdateOptions{
		PackageName:      packageName,
		ProductID:        "premium",
		BasePlanID:       "monthly",
		OfferID:          "intro",
		Action:           SubscriptionOfferStateActionDeactivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
	})
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
}

func TestUpdateSubscriptionOfferStatePassesOptionsToUpdater(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	updater := &fakeSubscriptionOfferClient{offer: SubscriptionOffer{OfferID: "intro"}}

	result, err := UpdateSubscriptionOfferState(context.Background(), updater, SubscriptionOfferStateUpdateOptions{
		PackageName:      packageName,
		ProductID:        "premium",
		BasePlanID:       "monthly",
		OfferID:          "intro",
		Action:           SubscriptionOfferStateActionDeactivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("UpdateSubscriptionOfferState() error = %v", err)
	}
	if !result.Applied {
		t.Fatal("Applied = false, want true")
	}
	if updater.stateOptions.Action != SubscriptionOfferStateActionDeactivate {
		t.Fatalf("Action = %q, want deactivate", updater.stateOptions.Action)
	}
}

type fakeSubscriptionOfferClient struct {
	listOptions                       SubscriptionOfferListOptions
	listResult                        SubscriptionOfferListResult
	batchOptions                      SubscriptionOfferBatchGetOptions
	batchResult                       SubscriptionOfferBatchGetResult
	batchStateOptions                 SubscriptionOfferBatchStateUpdateOptions
	batchStateResult                  SubscriptionOfferBatchStateUpdateResult
	batchAvailabilityOptions          SubscriptionOfferBatchPatchAvailabilityOptions
	batchAvailabilityResult           SubscriptionOfferBatchPatchAvailabilityResult
	batchPhaseRelativeDiscountOptions SubscriptionOfferBatchPatchPhaseRelativeDiscountsOptions
	batchPhaseRelativeDiscountResult  SubscriptionOfferBatchPatchPhaseRelativeDiscountsResult
	batchPhaseAbsoluteDiscountOptions SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsOptions
	batchPhaseAbsoluteDiscountResult  SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsResult
	deleteOptions                     SubscriptionOfferDeleteOptions
	offerID                           SubscriptionOfferID
	offer                             SubscriptionOffer
	stateOptions                      SubscriptionOfferStateUpdateOptions
}

func (c *fakeSubscriptionOfferClient) ListSubscriptionOffers(ctx context.Context, options SubscriptionOfferListOptions) (SubscriptionOfferListResult, error) {
	c.listOptions = options
	return c.listResult, nil
}

func (c *fakeSubscriptionOfferClient) GetSubscriptionOffer(ctx context.Context, packageName PackageName, productID SubscriptionProductID, basePlanID SubscriptionBasePlanID, offerID SubscriptionOfferID) (SubscriptionOffer, error) {
	c.offerID = offerID
	return c.offer, nil
}

func (c *fakeSubscriptionOfferClient) BatchGetSubscriptionOffers(ctx context.Context, options SubscriptionOfferBatchGetOptions) (SubscriptionOfferBatchGetResult, error) {
	c.batchOptions = options
	return c.batchResult, nil
}

func (c *fakeSubscriptionOfferClient) BatchUpdateSubscriptionOfferStates(ctx context.Context, options SubscriptionOfferBatchStateUpdateOptions) (SubscriptionOfferBatchStateUpdateResult, error) {
	c.batchStateOptions = options
	return c.batchStateResult, nil
}

func (c *fakeSubscriptionOfferClient) BatchPatchSubscriptionOfferAvailability(ctx context.Context, options SubscriptionOfferBatchPatchAvailabilityOptions) (SubscriptionOfferBatchPatchAvailabilityResult, error) {
	c.batchAvailabilityOptions = options
	return c.batchAvailabilityResult, nil
}

func (c *fakeSubscriptionOfferClient) BatchPatchSubscriptionOfferPhaseRelativeDiscounts(ctx context.Context, options SubscriptionOfferBatchPatchPhaseRelativeDiscountsOptions) (SubscriptionOfferBatchPatchPhaseRelativeDiscountsResult, error) {
	c.batchPhaseRelativeDiscountOptions = options
	return c.batchPhaseRelativeDiscountResult, nil
}

func (c *fakeSubscriptionOfferClient) BatchPatchSubscriptionOfferPhaseAbsoluteDiscounts(ctx context.Context, options SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsOptions) (SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsResult, error) {
	c.batchPhaseAbsoluteDiscountOptions = options
	return c.batchPhaseAbsoluteDiscountResult, nil
}

func (c *fakeSubscriptionOfferClient) DeleteSubscriptionOffer(ctx context.Context, options SubscriptionOfferDeleteOptions) error {
	c.deleteOptions = options
	return nil
}

func (c *fakeSubscriptionOfferClient) UpdateSubscriptionOfferState(ctx context.Context, options SubscriptionOfferStateUpdateOptions) (SubscriptionOffer, error) {
	c.stateOptions = options
	return c.offer, nil
}
