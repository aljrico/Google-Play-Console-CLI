package play

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestNewSubscriptionProductIDValidatesGoogleShape(t *testing.T) {
	valid, err := NewSubscriptionProductID("premium.monthly_1")
	if err != nil {
		t.Fatalf("NewSubscriptionProductID() error = %v", err)
	}
	if valid != "premium.monthly_1" {
		t.Fatalf("product ID = %q, want premium.monthly_1", valid)
	}

	for _, value := range []string{"", "Premium", "_premium", "premium-monthly"} {
		if _, err := NewSubscriptionProductID(value); err == nil {
			t.Fatalf("NewSubscriptionProductID(%q) succeeded, want error", value)
		}
	}
}

func TestListSubscriptionsPassesOptionsToLister(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	lister := &fakeSubscriptionClient{
		listResult: SubscriptionListResult{
			PackageName:   packageName,
			Subscriptions: []Subscription{{ProductID: "premium"}},
		},
	}

	result, err := ListSubscriptions(context.Background(), lister, SubscriptionListOptions{
		PackageName:  packageName,
		PageSize:     100,
		PageToken:    "next",
		ShowArchived: true,
	})
	if err != nil {
		t.Fatalf("ListSubscriptions() error = %v", err)
	}
	if len(result.Subscriptions) != 1 {
		t.Fatalf("len(Subscriptions) = %d, want 1", len(result.Subscriptions))
	}
	if !reflect.DeepEqual(lister.listOptions, SubscriptionListOptions{
		PackageName:  packageName,
		PageSize:     100,
		PageToken:    "next",
		ShowArchived: true,
	}) {
		t.Fatalf("listOptions = %#v", lister.listOptions)
	}
}

func TestListSubscriptionsRejectsPageSizeAboveGoogleLimit(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = ListSubscriptions(context.Background(), nil, SubscriptionListOptions{
		PackageName: packageName,
		PageSize:    1001,
	})
	if err == nil {
		t.Fatal("expected page size validation error")
	}
}

func TestGetSubscriptionPassesProductIDToGetter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	getter := &fakeSubscriptionClient{subscription: Subscription{ProductID: "premium"}}

	subscription, err := GetSubscription(context.Background(), getter, SubscriptionGetOptions{
		PackageName: packageName,
		ProductID:   "premium",
	})
	if err != nil {
		t.Fatalf("GetSubscription() error = %v", err)
	}
	if subscription.ProductID != "premium" {
		t.Fatalf("ProductID = %q, want premium", subscription.ProductID)
	}
	if getter.productID != "premium" {
		t.Fatalf("getter productID = %q, want premium", getter.productID)
	}
}

func TestBatchGetSubscriptionsPassesOptionsToGetter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	getter := &fakeSubscriptionClient{
		batchResult: SubscriptionBatchGetResult{
			PackageName:   packageName,
			Subscriptions: []Subscription{{ProductID: "premium_monthly"}},
		},
	}
	options := SubscriptionBatchGetOptions{
		PackageName: packageName,
		ProductIDs:  []SubscriptionProductID{"premium_monthly", "premium_yearly"},
	}

	result, err := BatchGetSubscriptions(context.Background(), getter, options)
	if err != nil {
		t.Fatalf("BatchGetSubscriptions() error = %v", err)
	}
	if len(result.Subscriptions) != 1 {
		t.Fatalf("len(Subscriptions) = %d, want 1", len(result.Subscriptions))
	}
	if !reflect.DeepEqual(getter.batchOptions, options) {
		t.Fatalf("batchOptions = %#v, want %#v", getter.batchOptions, options)
	}
}

func TestBatchGetSubscriptionsRejectsDuplicates(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = BatchGetSubscriptions(context.Background(), nil, SubscriptionBatchGetOptions{
		PackageName: packageName,
		ProductIDs:  []SubscriptionProductID{"premium_monthly", "premium_monthly"},
	})
	if err == nil {
		t.Fatal("expected duplicate product ID validation error")
	}
}

func TestCreateSubscriptionDryRunBuildsPlanWithoutCreator(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	subscription := validSubscriptionForCreate()
	subscription.Archived = true
	subscription.BasePlans[0].State = SubscriptionStateActive

	result, err := CreateSubscription(context.Background(), nil, SubscriptionCreateOptions{
		PackageName:    packageName,
		ProductID:      "premium",
		Subscription:   subscription,
		RegionsVersion: "2026/05",
		DryRun:         true,
	})
	if err != nil {
		t.Fatalf("CreateSubscription() error = %v", err)
	}
	if !result.DryRun || result.Created {
		t.Fatalf("result = %#v, want dry-run create", result)
	}
	if result.Desired.PackageName != packageName || result.Desired.ProductID != "premium" || result.Desired.Archived || result.Desired.BasePlans[0].State != "" {
		t.Fatalf("Desired = %#v, want normalized create subscription", result.Desired)
	}
	if result.Desired.Listings[0].LanguageCode != "en-US" {
		t.Fatalf("Desired listing language = %q, want canonical en-US", result.Desired.Listings[0].LanguageCode)
	}
}

func TestDecodeSubscriptionCreateJSONAcceptsGooglePlayAPIJSON(t *testing.T) {
	subscription, err := DecodeSubscriptionCreateJSON([]byte(`{
		"packageName":"ignored",
		"productId":"ignored",
		"listings":[{"languageCode":"en-US","title":"Premium","description":"Full access"}],
		"basePlans":[{
			"basePlanId":"monthly",
			"state":"ACTIVE",
			"autoRenewingBasePlanType":{"billingPeriodDuration":"P1M","legacyCompatible":true},
			"offerTags":[{"tag":"public"}],
			"regionalConfigs":[{"regionCode":"US","newSubscriberAvailability":true,"price":{"currencyCode":"USD","units":"4","nanos":990000000}}],
			"otherRegionsConfig":{"newSubscriberAvailability":true,"usdPrice":{"currencyCode":"USD","units":"4"},"eurPrice":{"currencyCode":"EUR","units":"4"}}
		}]
	}`))
	if err != nil {
		t.Fatalf("DecodeSubscriptionCreateJSON() error = %v", err)
	}
	if subscription.BasePlans[0].Type != SubscriptionBasePlanTypeAutoRenewing || !subscription.BasePlans[0].LegacyCompatible {
		t.Fatalf("BasePlans = %#v, want decoded auto-renewing plan", subscription.BasePlans)
	}
	if subscription.BasePlans[0].OfferTags[0] != "public" || subscription.BasePlans[0].RegionalConfigs[0].Price.Units != 4 {
		t.Fatalf("BasePlan = %#v, want decoded API tags and price", subscription.BasePlans[0])
	}
}

func TestDecodeSubscriptionCreateJSONRejectsNestedUnknownFields(t *testing.T) {
	_, err := DecodeSubscriptionCreateJSON([]byte(`{"listings":[],"basePlans":[{"basePlanId":"monthly","autoRenewingBasePlanType":{"billingPeriodDuration":"P1M","typo":true}}]}`))
	if err == nil {
		t.Fatal("expected nested unknown field validation error")
	}
}

func TestDecodeSubscriptionCreateJSONRejectsMultipleBasePlanTypeObjects(t *testing.T) {
	_, err := DecodeSubscriptionCreateJSON([]byte(`{
		"listings":[{"languageCode":"en-US","title":"Premium"}],
		"basePlans":[{
			"basePlanId":"monthly",
			"autoRenewingBasePlanType":{"billingPeriodDuration":"P1M"},
			"prepaidBasePlanType":{"billingPeriodDuration":"P1M"}
		}]
	}`))
	if err == nil {
		t.Fatal("expected base plan type union validation error")
	}
}

func TestCreateSubscriptionRejectsInvalidBody(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	tests := []struct {
		name         string
		subscription Subscription
	}{
		{name: "missing listings", subscription: Subscription{BasePlans: validSubscriptionForCreate().BasePlans}},
		{name: "missing base plans", subscription: Subscription{Listings: validSubscriptionForCreate().Listings}},
		{
			name: "bad tag",
			subscription: func() Subscription {
				subscription := validSubscriptionForCreate()
				subscription.BasePlans[0].OfferTags = []string{"Bad"}
				return subscription
			}(),
		},
		{
			name: "bad duration",
			subscription: func() Subscription {
				subscription := validSubscriptionForCreate()
				subscription.BasePlans[0].BillingPeriodDuration = "P1MT"
				return subscription
			}(),
		},
		{
			name: "overlong description",
			subscription: func() Subscription {
				subscription := validSubscriptionForCreate()
				subscription.Listings[0].Description = strings.Repeat("x", 81)
				return subscription
			}(),
		},
		{
			name: "too many benefits",
			subscription: func() Subscription {
				subscription := validSubscriptionForCreate()
				subscription.Listings[0].Benefits = []string{"a", "b", "c", "d", "e"}
				return subscription
			}(),
		},
		{
			name: "bad installment renewal type",
			subscription: func() Subscription {
				subscription := validSubscriptionForCreate()
				subscription.BasePlans[0].Type = SubscriptionBasePlanTypeInstallments
				subscription.BasePlans[0].CommittedPaymentsCount = 3
				subscription.BasePlans[0].RenewalType = ""
				return subscription
			}(),
		},
		{
			name: "bad proration mode",
			subscription: func() Subscription {
				subscription := validSubscriptionForCreate()
				subscription.BasePlans[0].ProrationMode = "NOPE"
				return subscription
			}(),
		},
		{
			name: "bad grace period unit",
			subscription: func() Subscription {
				subscription := validSubscriptionForCreate()
				subscription.BasePlans[0].GracePeriodDuration = "P1M"
				return subscription
			}(),
		},
		{
			name: "mixed day duration",
			subscription: func() Subscription {
				subscription := validSubscriptionForCreate()
				subscription.BasePlans[0].GracePeriodDuration = "P1M1D"
				return subscription
			}(),
		},
		{
			name: "account hold above range",
			subscription: func() Subscription {
				subscription := validSubscriptionForCreate()
				subscription.BasePlans[0].AccountHoldDuration = "P61D"
				return subscription
			}(),
		},
		{
			name: "grace and hold sum below range",
			subscription: func() Subscription {
				subscription := validSubscriptionForCreate()
				subscription.BasePlans[0].GracePeriodDuration = "P1D"
				subscription.BasePlans[0].AccountHoldDuration = "P1D"
				return subscription
			}(),
		},
		{
			name: "unsupported billing period",
			subscription: func() Subscription {
				subscription := validSubscriptionForCreate()
				subscription.BasePlans[0].BillingPeriodDuration = "P2M"
				return subscription
			}(),
		},
		{
			name: "mixed billing period duration",
			subscription: func() Subscription {
				subscription := validSubscriptionForCreate()
				subscription.BasePlans[0].BillingPeriodDuration = "P1M1D"
				return subscription
			}(),
		},
		{
			name: "weekly grace period above billing period",
			subscription: func() Subscription {
				subscription := validSubscriptionForCreate()
				subscription.BasePlans[0].BillingPeriodDuration = "P1W"
				subscription.BasePlans[0].GracePeriodDuration = "P30D"
				return subscription
			}(),
		},
		{
			name: "bad base plan ID",
			subscription: func() Subscription {
				subscription := validSubscriptionForCreate()
				subscription.BasePlans[0].BasePlanID = "monthly-"
				return subscription
			}(),
		},
		{
			name: "bad legacy compatible offer ID",
			subscription: func() Subscription {
				subscription := validSubscriptionForCreate()
				subscription.BasePlans[0].LegacyCompatibleSubscriptionOfferID = "intro-"
				return subscription
			}(),
		},
		{
			name: "multiple legacy compatible base plans",
			subscription: func() Subscription {
				subscription := validSubscriptionForCreate()
				secondBasePlan := subscription.BasePlans[0]
				secondBasePlan.BasePlanID = "annual"
				secondBasePlan.BillingPeriodDuration = "P1Y"
				subscription.BasePlans[0].LegacyCompatible = true
				secondBasePlan.LegacyCompatible = true
				subscription.BasePlans = append(subscription.BasePlans, secondBasePlan)
				return subscription
			}(),
		},
		{
			name: "ignored prepaid field on auto renewing",
			subscription: func() Subscription {
				subscription := validSubscriptionForCreate()
				subscription.BasePlans[0].TimeExtension = "TIME_EXTENSION_ACTIVE"
				return subscription
			}(),
		},
		{
			name: "ignored auto renewing field on prepaid",
			subscription: func() Subscription {
				subscription := validSubscriptionForCreate()
				subscription.BasePlans[0].Type = SubscriptionBasePlanTypePrepaid
				subscription.BasePlans[0].GracePeriodDuration = "P7D"
				return subscription
			}(),
		},
		{
			name: "bad restricted country",
			subscription: func() Subscription {
				subscription := validSubscriptionForCreate()
				subscription.RestrictedCountries = []string{"USA"}
				return subscription
			}(),
		},
		{
			name: "empty tax settings",
			subscription: func() Subscription {
				subscription := validSubscriptionForCreate()
				subscription.TaxAndComplianceSettings = &ProductTaxComplianceSettings{}
				return subscription
			}(),
		},
		{
			name: "available region missing price",
			subscription: func() Subscription {
				subscription := validSubscriptionForCreate()
				subscription.BasePlans[0].RegionalConfigs[0].Price = nil
				return subscription
			}(),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := CreateSubscription(context.Background(), nil, SubscriptionCreateOptions{
				PackageName:    packageName,
				ProductID:      "premium",
				Subscription:   test.subscription,
				RegionsVersion: "2026/05",
				DryRun:         true,
			})
			if err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestDeleteSubscriptionDryRunBuildsPlanWithoutDeleter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := DeleteSubscription(context.Background(), nil, SubscriptionDeleteOptions{
		PackageName: packageName,
		ProductID:   "premium_monthly",
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("DeleteSubscription() error = %v", err)
	}
	if !result.DryRun || result.Deleted {
		t.Fatalf("result = %#v, want dry-run deletion plan", result)
	}
	if !reflect.DeepEqual(result.Plan.Steps, []string{"delete subscription"}) {
		t.Fatalf("steps = %#v, want delete step", result.Plan.Steps)
	}
}

func TestDeleteSubscriptionRequiresConfirmOrDryRun(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = DeleteSubscription(context.Background(), nil, SubscriptionDeleteOptions{
		PackageName: packageName,
		ProductID:   "premium_monthly",
	})
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
}

func TestDeleteSubscriptionPassesOptionsToDeleter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	deleter := &fakeSubscriptionClient{}

	result, err := DeleteSubscription(context.Background(), deleter, SubscriptionDeleteOptions{
		PackageName: packageName,
		ProductID:   "premium_monthly",
		Confirm:     true,
	})
	if err != nil {
		t.Fatalf("DeleteSubscription() error = %v", err)
	}
	if !result.Deleted {
		t.Fatal("Deleted = false, want true")
	}
	if deleter.deleteOptions.ProductID != "premium_monthly" {
		t.Fatalf("deleteOptions = %#v, want premium_monthly", deleter.deleteOptions)
	}
}

func TestDeleteBasePlanDryRunBuildsPlanWithoutDeleter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := DeleteBasePlan(context.Background(), nil, BasePlanDeleteOptions{
		PackageName: packageName,
		ProductID:   "premium",
		BasePlanID:  "monthly",
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("DeleteBasePlan() error = %v", err)
	}
	if !result.DryRun || result.Deleted {
		t.Fatalf("result = %#v, want dry-run deletion plan", result)
	}
	if !reflect.DeepEqual(result.Plan.Steps, []string{"delete base plan"}) {
		t.Fatalf("steps = %#v, want delete step", result.Plan.Steps)
	}
}

func TestDeleteBasePlanRequiresConfirmOrDryRun(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = DeleteBasePlan(context.Background(), nil, BasePlanDeleteOptions{
		PackageName: packageName,
		ProductID:   "premium",
		BasePlanID:  "monthly",
	})
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
}

func TestDeleteBasePlanPassesOptionsToDeleter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	deleter := &fakeSubscriptionClient{}

	result, err := DeleteBasePlan(context.Background(), deleter, BasePlanDeleteOptions{
		PackageName: packageName,
		ProductID:   "premium",
		BasePlanID:  "monthly",
		Confirm:     true,
	})
	if err != nil {
		t.Fatalf("DeleteBasePlan() error = %v", err)
	}
	if !result.Deleted {
		t.Fatal("Deleted = false, want true")
	}
	if deleter.basePlanDeleteOptions.BasePlanID != "monthly" {
		t.Fatalf("basePlanDeleteOptions = %#v, want monthly", deleter.basePlanDeleteOptions)
	}
}

func TestUpdateBasePlanStateDryRunBuildsPlanWithoutUpdater(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := UpdateBasePlanState(context.Background(), nil, BasePlanStateUpdateOptions{
		PackageName:      packageName,
		ProductID:        "premium",
		BasePlanID:       "monthly",
		Action:           BasePlanStateActionActivate,
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("UpdateBasePlanState() error = %v", err)
	}
	if !result.DryRun {
		t.Fatal("DryRun = false, want true")
	}
	wantSteps := []string{"plan activate base plan"}
	if !reflect.DeepEqual(result.Plan.Steps, wantSteps) {
		t.Fatalf("steps = %#v, want %#v", result.Plan.Steps, wantSteps)
	}
}

func TestUpdateBasePlanStateRequiresConfirmOrDryRun(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = NewBasePlanStateUpdatePlan(BasePlanStateUpdateOptions{
		PackageName:      packageName,
		ProductID:        "premium",
		BasePlanID:       "monthly",
		Action:           BasePlanStateActionDeactivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
	})
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
}

func TestUpdateBasePlanStatePassesOptionsToUpdater(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	updater := &fakeSubscriptionClient{
		subscription: Subscription{ProductID: "premium"},
	}

	result, err := UpdateBasePlanState(context.Background(), updater, BasePlanStateUpdateOptions{
		PackageName:      packageName,
		ProductID:        "premium",
		BasePlanID:       "monthly",
		Action:           BasePlanStateActionDeactivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("UpdateBasePlanState() error = %v", err)
	}
	if !result.Applied {
		t.Fatal("Applied = false, want true")
	}
	if updater.stateOptions.Action != BasePlanStateActionDeactivate {
		t.Fatalf("Action = %q, want deactivate", updater.stateOptions.Action)
	}
}

func TestBatchUpdateBasePlanStatesDryRunBuildsPlanWithoutUpdater(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := BatchUpdateBasePlanStates(context.Background(), nil, BasePlanBatchStateUpdateOptions{
		PackageName: packageName,
		ProductID:   "premium",
		Requests: []BasePlanBatchStateUpdateRequest{
			{ProductID: "premium", BasePlanID: "monthly"},
			{ProductID: "premium", BasePlanID: "annual"},
		},
		Action:           BasePlanStateActionActivate,
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("BatchUpdateBasePlanStates() error = %v", err)
	}
	if !result.DryRun {
		t.Fatal("DryRun = false, want true")
	}
	wantSteps := []string{"plan batch activate base plans"}
	if !reflect.DeepEqual(result.Plan.Steps, wantSteps) {
		t.Fatalf("steps = %#v, want %#v", result.Plan.Steps, wantSteps)
	}
}

func TestBatchUpdateBasePlanStatesRejectsDuplicateBasePlan(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = NewBasePlanBatchStateUpdatePlan(BasePlanBatchStateUpdateOptions{
		PackageName: packageName,
		ProductID:   "premium",
		Requests: []BasePlanBatchStateUpdateRequest{
			{ProductID: "premium", BasePlanID: "monthly"},
			{ProductID: "premium", BasePlanID: "monthly"},
		},
		Action:           BasePlanStateActionDeactivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected duplicate validation error")
	}
}

func TestBatchUpdateBasePlanStatesAllowsMultipleProductsWithWildcardParent(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = NewBasePlanBatchStateUpdatePlan(BasePlanBatchStateUpdateOptions{
		PackageName: packageName,
		ProductID:   "-",
		Requests: []BasePlanBatchStateUpdateRequest{
			{ProductID: "premium", BasePlanID: "monthly"},
			{ProductID: "vip", BasePlanID: "annual"},
		},
		Action:           BasePlanStateActionDeactivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("NewBasePlanBatchStateUpdatePlan() error = %v", err)
	}
}

func TestBatchUpdateBasePlanStatesRejectsConcreteParentMismatch(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = NewBasePlanBatchStateUpdatePlan(BasePlanBatchStateUpdateOptions{
		PackageName: packageName,
		ProductID:   "premium",
		Requests: []BasePlanBatchStateUpdateRequest{
			{ProductID: "vip", BasePlanID: "monthly"},
		},
		Action:           BasePlanStateActionDeactivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected parent mismatch validation error")
	}
}

func TestBatchUpdateBasePlanStatesPassesOptionsToUpdater(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	updater := &fakeSubscriptionClient{
		batchStateResult: BasePlanBatchStateUpdateResult{
			Subscriptions: []Subscription{{ProductID: "premium"}},
		},
	}

	result, err := BatchUpdateBasePlanStates(context.Background(), updater, BasePlanBatchStateUpdateOptions{
		PackageName: packageName,
		ProductID:   "premium",
		Requests: []BasePlanBatchStateUpdateRequest{
			{ProductID: "premium", BasePlanID: "monthly"},
		},
		Action:           BasePlanStateActionDeactivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("BatchUpdateBasePlanStates() error = %v", err)
	}
	if !result.Applied {
		t.Fatal("Applied = false, want true")
	}
	if updater.batchStateOptions.Action != BasePlanStateActionDeactivate {
		t.Fatalf("Action = %q, want deactivate", updater.batchStateOptions.Action)
	}
}

func TestBatchMigrateBasePlanPricesDryRunBuildsPlanWithoutMigrator(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := BatchMigrateBasePlanPrices(context.Background(), nil, BasePlanBatchPriceMigrationOptions{
		PackageName:    packageName,
		ProductID:      "premium",
		RegionsVersion: "2026/05",
		Requests: []BasePlanPriceMigrationRequest{
			{
				ProductID:  "premium",
				BasePlanID: "monthly",
				Regions: []BasePlanPriceMigrationConfig{
					{RegionCode: "US", OldestAllowedPriceVersionTime: "2026-05-01T00:00:00Z", PriceIncreaseType: BasePlanPriceIncreaseTypeOptOut},
				},
			},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("BatchMigrateBasePlanPrices() error = %v", err)
	}
	if !result.DryRun {
		t.Fatal("DryRun = false, want true")
	}
	wantSteps := []string{"plan batch base plan price migration"}
	if !reflect.DeepEqual(result.Plan.Steps, wantSteps) {
		t.Fatalf("steps = %#v, want %#v", result.Plan.Steps, wantSteps)
	}
}

func TestBatchMigrateBasePlanPricesRejectsDuplicateRegion(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = NewBasePlanBatchPriceMigrationPlan(BasePlanBatchPriceMigrationOptions{
		PackageName:    packageName,
		ProductID:      "premium",
		RegionsVersion: "2026/05",
		Requests: []BasePlanPriceMigrationRequest{
			{
				ProductID:  "premium",
				BasePlanID: "monthly",
				Regions: []BasePlanPriceMigrationConfig{
					{RegionCode: "US", OldestAllowedPriceVersionTime: "2026-05-01T00:00:00Z"},
					{RegionCode: "US", OldestAllowedPriceVersionTime: "2026-05-01T00:00:00Z"},
				},
			},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected duplicate region validation error")
	}
}

func TestBatchMigrateBasePlanPricesPassesOptionsToMigrator(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	migrator := &fakeSubscriptionClient{
		priceMigrationResult: BasePlanBatchPriceMigrationResult{
			Responses: []BasePlanPriceMigrationResponse{{ProductID: "premium", BasePlanID: "monthly"}},
		},
	}

	result, err := BatchMigrateBasePlanPrices(context.Background(), migrator, BasePlanBatchPriceMigrationOptions{
		PackageName:    packageName,
		ProductID:      "premium",
		RegionsVersion: "2026/05",
		Requests: []BasePlanPriceMigrationRequest{
			{
				ProductID:  "premium",
				BasePlanID: "monthly",
				Regions: []BasePlanPriceMigrationConfig{
					{RegionCode: "US", OldestAllowedPriceVersionTime: "2026-05-01T00:00:00Z"},
				},
			},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("BatchMigrateBasePlanPrices() error = %v", err)
	}
	if !result.Applied {
		t.Fatal("Applied = false, want true")
	}
	if migrator.priceMigrationOptions.RegionsVersion != "2026/05" {
		t.Fatalf("RegionsVersion = %q, want 2026/05", migrator.priceMigrationOptions.RegionsVersion)
	}
}

func TestPatchSubscriptionDryRunBuildsPlanWithoutPatcher(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := PatchSubscription(context.Background(), nil, SubscriptionPatchOptions{
		PackageName:      packageName,
		ProductID:        "premium",
		Listing:          SubscriptionListing{LanguageCode: "en-US", Title: "Premium", Description: "Full access"},
		DescriptionSet:   true,
		RegionsVersion:   "2022/02",
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("PatchSubscription() error = %v", err)
	}
	if !result.DryRun || result.Applied {
		t.Fatalf("result = %#v, want dry-run patch plan", result)
	}
	if result.Plan.UpdateMask != "listings" {
		t.Fatalf("UpdateMask = %q, want listings", result.Plan.UpdateMask)
	}
	if result.Plan.RegionsVersion != "2022/02" {
		t.Fatalf("RegionsVersion = %q, want 2022/02", result.Plan.RegionsVersion)
	}
	if result.Desired.BasePlans == nil {
		t.Fatal("Desired.BasePlans = nil, want empty slice")
	}
	wantSteps := []string{"plan subscription listing patch"}
	if !reflect.DeepEqual(result.Plan.Steps, wantSteps) {
		t.Fatalf("steps = %#v, want %#v", result.Plan.Steps, wantSteps)
	}
}

func TestPatchSubscriptionRequiresConfirmOrDryRun(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = PatchSubscription(context.Background(), nil, SubscriptionPatchOptions{
		PackageName:      packageName,
		ProductID:        "premium",
		Listing:          SubscriptionListing{LanguageCode: "en-US", Title: "Premium"},
		RegionsVersion:   "2022/02",
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
	})
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
}

func TestPatchSubscriptionRejectsTooManyBenefits(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = NewSubscriptionPatchPlan(SubscriptionPatchOptions{
		PackageName: packageName,
		ProductID:   "premium",
		Listing: SubscriptionListing{
			LanguageCode: "en-US",
			Title:        "Premium",
			Benefits:     []string{"one", "two", "three", "four", "five"},
		},
		BenefitsSet:      true,
		RegionsVersion:   "2022/02",
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected benefits validation error")
	}
}

func TestPatchSubscriptionRejectsDescriptionAboveGoogleLimit(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = NewSubscriptionPatchPlan(SubscriptionPatchOptions{
		PackageName: packageName,
		ProductID:   "premium",
		Listing: SubscriptionListing{
			LanguageCode: "en-US",
			Title:        "Premium",
			Description:  "This subscription description is intentionally longer than eighty characters for validation.",
		},
		DescriptionSet:   true,
		RegionsVersion:   "2022/02",
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected description validation error")
	}
}

func TestPatchSubscriptionPassesOptionsToPatcher(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	patcher := &fakeSubscriptionClient{
		subscription: Subscription{ProductID: "premium"},
	}
	options := SubscriptionPatchOptions{
		PackageName:      packageName,
		ProductID:        "premium",
		Listing:          SubscriptionListing{LanguageCode: "en-US", Title: "Premium", Description: "Full access"},
		DescriptionSet:   true,
		RegionsVersion:   "2022/02",
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	}

	result, err := PatchSubscription(context.Background(), patcher, options)
	if err != nil {
		t.Fatalf("PatchSubscription() error = %v", err)
	}
	if !result.Applied {
		t.Fatal("Applied = false, want true")
	}
	if !reflect.DeepEqual(patcher.patchOptions, options) {
		t.Fatalf("patchOptions = %#v, want %#v", patcher.patchOptions, options)
	}
}

func TestBatchPatchSubscriptionListingsDryRunBuildsPlanWithoutPatcher(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := BatchPatchSubscriptionListings(context.Background(), nil, SubscriptionBatchPatchListingsOptions{
		PackageName: packageName,
		Requests: []SubscriptionBatchPatchListingRequest{
			{ProductID: "premium", Listing: SubscriptionListing{LanguageCode: "en-US", Title: "Premium", Description: "Full access"}},
			{ProductID: "vip", Listing: SubscriptionListing{LanguageCode: "es-ES", Title: "VIP", Description: "Acceso completo"}},
		},
		RegionsVersion:   "2026/05",
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("BatchPatchSubscriptionListings() error = %v", err)
	}
	if !result.DryRun || result.Applied {
		t.Fatalf("result = %#v, want dry-run batch patch", result)
	}
	if len(result.Desired) != 2 {
		t.Fatalf("len(Desired) = %d, want 2", len(result.Desired))
	}
	wantSteps := []string{"plan subscription listing batch patch"}
	if !reflect.DeepEqual(result.Plan.Steps, wantSteps) {
		t.Fatalf("steps = %#v, want %#v", result.Plan.Steps, wantSteps)
	}
}

func TestBatchPatchSubscriptionListingsRejectsDuplicateProductLanguage(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = NewSubscriptionBatchPatchListingsPlan(SubscriptionBatchPatchListingsOptions{
		PackageName: packageName,
		Requests: []SubscriptionBatchPatchListingRequest{
			{ProductID: "premium", Listing: SubscriptionListing{LanguageCode: "en-US", Title: "Premium", Description: "Full access"}},
			{ProductID: "premium", Listing: SubscriptionListing{LanguageCode: "en-US", Title: "Premium 2", Description: "Full access"}},
		},
		RegionsVersion:   "2026/05",
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected duplicate listing validation error")
	}
}

func TestBatchPatchSubscriptionListingsAllowsMoreThanOneHundredListingsAcrossFewerSubscriptions(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	requests := make([]SubscriptionBatchPatchListingRequest, 0, 102)
	for productIndex := 0; productIndex < 51; productIndex++ {
		productID := SubscriptionProductID(fmt.Sprintf("premium%d", productIndex))
		requests = append(requests,
			SubscriptionBatchPatchListingRequest{
				ProductID: productID,
				Listing:   SubscriptionListing{LanguageCode: "en-US", Title: "Premium", Description: "Full access"},
			},
			SubscriptionBatchPatchListingRequest{
				ProductID: productID,
				Listing:   SubscriptionListing{LanguageCode: "es-ES", Title: "Premium", Description: "Acceso completo"},
			},
		)
	}

	_, err = NewSubscriptionBatchPatchListingsPlan(SubscriptionBatchPatchListingsOptions{
		PackageName:      packageName,
		Requests:         requests,
		RegionsVersion:   "2026/05",
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("NewSubscriptionBatchPatchListingsPlan() error = %v", err)
	}
}

func TestBatchPatchSubscriptionListingsPassesOptionsToPatcher(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	patcher := &fakeSubscriptionClient{
		batchPatchListingsResult: SubscriptionBatchPatchListingsResult{
			Subscriptions: []Subscription{{ProductID: "premium"}},
		},
	}
	options := SubscriptionBatchPatchListingsOptions{
		PackageName: packageName,
		Requests: []SubscriptionBatchPatchListingRequest{
			{ProductID: "premium", Listing: SubscriptionListing{LanguageCode: "en-US", Title: "Premium", Description: "Full access"}},
		},
		RegionsVersion:   "2026/05",
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	}

	result, err := BatchPatchSubscriptionListings(context.Background(), patcher, options)
	if err != nil {
		t.Fatalf("BatchPatchSubscriptionListings() error = %v", err)
	}
	if !result.Applied {
		t.Fatal("Applied = false, want true")
	}
	if !reflect.DeepEqual(patcher.batchPatchListingsOptions, options) {
		t.Fatalf("batchPatchListingsOptions = %#v, want %#v", patcher.batchPatchListingsOptions, options)
	}
}

type fakeSubscriptionClient struct {
	listOptions               SubscriptionListOptions
	listResult                SubscriptionListResult
	batchOptions              SubscriptionBatchGetOptions
	batchResult               SubscriptionBatchGetResult
	createOptions             SubscriptionCreateOptions
	deleteOptions             SubscriptionDeleteOptions
	patchOptions              SubscriptionPatchOptions
	batchPatchListingsOptions SubscriptionBatchPatchListingsOptions
	batchPatchListingsResult  SubscriptionBatchPatchListingsResult
	productID                 SubscriptionProductID
	subscription              Subscription
	basePlanDeleteOptions     BasePlanDeleteOptions
	stateOptions              BasePlanStateUpdateOptions
	batchStateOptions         BasePlanBatchStateUpdateOptions
	batchStateResult          BasePlanBatchStateUpdateResult
	priceMigrationOptions     BasePlanBatchPriceMigrationOptions
	priceMigrationResult      BasePlanBatchPriceMigrationResult
}

func (c *fakeSubscriptionClient) ListSubscriptions(ctx context.Context, options SubscriptionListOptions) (SubscriptionListResult, error) {
	c.listOptions = options
	return c.listResult, nil
}

func (c *fakeSubscriptionClient) GetSubscription(ctx context.Context, packageName PackageName, productID SubscriptionProductID) (Subscription, error) {
	c.productID = productID
	return c.subscription, nil
}

func (c *fakeSubscriptionClient) BatchGetSubscriptions(ctx context.Context, options SubscriptionBatchGetOptions) (SubscriptionBatchGetResult, error) {
	c.batchOptions = options
	return c.batchResult, nil
}

func (c *fakeSubscriptionClient) CreateSubscription(ctx context.Context, options SubscriptionCreateOptions) (Subscription, error) {
	c.createOptions = options
	return c.subscription, nil
}

func (c *fakeSubscriptionClient) DeleteSubscription(ctx context.Context, options SubscriptionDeleteOptions) error {
	c.deleteOptions = options
	return nil
}

func (c *fakeSubscriptionClient) DeleteBasePlan(ctx context.Context, options BasePlanDeleteOptions) error {
	c.basePlanDeleteOptions = options
	return nil
}

func (c *fakeSubscriptionClient) UpdateBasePlanState(ctx context.Context, options BasePlanStateUpdateOptions) (Subscription, error) {
	c.stateOptions = options
	return c.subscription, nil
}

func (c *fakeSubscriptionClient) BatchUpdateBasePlanStates(ctx context.Context, options BasePlanBatchStateUpdateOptions) (BasePlanBatchStateUpdateResult, error) {
	c.batchStateOptions = options
	return c.batchStateResult, nil
}

func (c *fakeSubscriptionClient) BatchMigrateBasePlanPrices(ctx context.Context, options BasePlanBatchPriceMigrationOptions) (BasePlanBatchPriceMigrationResult, error) {
	c.priceMigrationOptions = options
	return c.priceMigrationResult, nil
}

func (c *fakeSubscriptionClient) PatchSubscription(ctx context.Context, options SubscriptionPatchOptions) (Subscription, error) {
	c.patchOptions = options
	return c.subscription, nil
}

func (c *fakeSubscriptionClient) BatchPatchSubscriptionListings(ctx context.Context, options SubscriptionBatchPatchListingsOptions) (SubscriptionBatchPatchListingsResult, error) {
	c.batchPatchListingsOptions = options
	return c.batchPatchListingsResult, nil
}

func validSubscriptionForCreate() Subscription {
	return Subscription{
		Listings: []SubscriptionListing{{LanguageCode: "en-US", Title: "Premium", Description: "Full access"}},
		BasePlans: []SubscriptionBasePlan{{
			BasePlanID:            "monthly",
			Type:                  SubscriptionBasePlanTypeAutoRenewing,
			BillingPeriodDuration: "P1M",
			GracePeriodDuration:   "P7D",
			OfferTags:             []string{"public"},
			RegionalConfigs:       []SubscriptionRegionalConfig{{RegionCode: "US", NewSubscriberAvailability: true, Price: &Money{CurrencyCode: "USD", Units: 4, Nanos: 990000000}}},
			OtherRegionsConfig:    &SubscriptionOtherRegionsConfig{NewSubscriberAvailability: true, USDPrice: &Money{CurrencyCode: "USD", Units: 4}, EURPrice: &Money{CurrencyCode: "EUR", Units: 4}},
		}},
	}
}
