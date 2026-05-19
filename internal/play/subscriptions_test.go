package play

import (
	"context"
	"reflect"
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

func TestPatchSubscriptionDryRunBuildsPlanWithoutPatcher(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := PatchSubscription(context.Background(), nil, SubscriptionPatchOptions{
		PackageName:      packageName,
		ProductID:        "premium",
		Listing:          SubscriptionListing{LanguageCode: "en-US", Title: "Premium", Description: "Full access"},
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
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected benefits validation error")
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

type fakeSubscriptionClient struct {
	listOptions   SubscriptionListOptions
	listResult    SubscriptionListResult
	batchOptions  SubscriptionBatchGetOptions
	batchResult   SubscriptionBatchGetResult
	deleteOptions SubscriptionDeleteOptions
	patchOptions  SubscriptionPatchOptions
	productID     SubscriptionProductID
	subscription  Subscription
	stateOptions  BasePlanStateUpdateOptions
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

func (c *fakeSubscriptionClient) DeleteSubscription(ctx context.Context, options SubscriptionDeleteOptions) error {
	c.deleteOptions = options
	return nil
}

func (c *fakeSubscriptionClient) UpdateBasePlanState(ctx context.Context, options BasePlanStateUpdateOptions) (Subscription, error) {
	c.stateOptions = options
	return c.subscription, nil
}

func (c *fakeSubscriptionClient) PatchSubscription(ctx context.Context, options SubscriptionPatchOptions) (Subscription, error) {
	c.patchOptions = options
	return c.subscription, nil
}
