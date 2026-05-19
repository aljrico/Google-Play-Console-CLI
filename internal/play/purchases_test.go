package play

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestGetProductPurchasePassesOptionsToGetter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	getter := &fakePurchaseClient{productPurchase: ProductPurchase{ProductID: "coins_100"}}

	purchase, err := GetProductPurchase(context.Background(), getter, ProductPurchaseOptions{
		PackageName: packageName,
		ProductID:   "coins_100",
		Token:       "token-123",
	})
	if err != nil {
		t.Fatalf("GetProductPurchase() error = %v", err)
	}
	if purchase.ProductID != "coins_100" {
		t.Fatalf("ProductID = %q, want coins_100", purchase.ProductID)
	}
	if getter.productOptions.Token != "token-123" {
		t.Fatalf("token = %q, want token-123", getter.productOptions.Token)
	}
}

func TestGetProductPurchaseRejectsMissingToken(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = GetProductPurchase(context.Background(), nil, ProductPurchaseOptions{
		PackageName: packageName,
		ProductID:   "coins_100",
	})
	if err == nil {
		t.Fatal("expected token validation error")
	}
}

func TestAcknowledgeProductPurchaseDryRunDoesNotCallMutator(t *testing.T) {
	result, err := AcknowledgeProductPurchase(context.Background(), nil, ProductPurchaseMutationOptions{
		PackageName: "com.example.app",
		ProductID:   "coins_100",
		Token:       "token-123",
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("AcknowledgeProductPurchase() error = %v", err)
	}
	if result.Applied {
		t.Fatalf("Applied = true, want false")
	}
	if result.Action != "acknowledge" || result.Plan.Action != "acknowledge" {
		t.Fatalf("result = %#v, want acknowledge action", result)
	}
}

func TestConsumeProductPurchasePassesOptionsToMutator(t *testing.T) {
	mutator := &fakePurchaseClient{}
	options := ProductPurchaseMutationOptions{
		PackageName: "com.example.app",
		ProductID:   "coins_100",
		Token:       "token-123",
		Confirm:     true,
	}

	result, err := ConsumeProductPurchase(context.Background(), mutator, options)
	if err != nil {
		t.Fatalf("ConsumeProductPurchase() error = %v", err)
	}
	if !result.Applied {
		t.Fatalf("Applied = false, want true")
	}
	if mutator.consumeOptions != options {
		t.Fatalf("consumeOptions = %#v, want %#v", mutator.consumeOptions, options)
	}
}

func TestProductPurchaseMutationRejectsInvalidOptions(t *testing.T) {
	tests := []ProductPurchaseMutationOptions{
		{},
		{PackageName: "bad", ProductID: "coins_100", Token: "token-123", DryRun: true},
		{PackageName: "com.example.app", Token: "token-123", DryRun: true},
		{PackageName: "com.example.app", ProductID: "coins_100", DryRun: true},
		{PackageName: "com.example.app", ProductID: "coins_100", Token: "token-123"},
		{PackageName: "com.example.app", ProductID: "coins_100", Token: "token-123", Confirm: true, DryRun: true},
	}
	for _, options := range tests {
		if _, err := AcknowledgeProductPurchase(context.Background(), nil, options); err == nil {
			t.Fatalf("AcknowledgeProductPurchase(%#v) expected validation error", options)
		}
	}
}

func TestGetSubscriptionPurchasePassesOptionsToGetter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	getter := &fakePurchaseClient{subscriptionPurchase: SubscriptionPurchase{Token: "token-123"}}

	purchase, err := GetSubscriptionPurchase(context.Background(), getter, SubscriptionPurchaseOptions{
		PackageName: packageName,
		Token:       "token-123",
	})
	if err != nil {
		t.Fatalf("GetSubscriptionPurchase() error = %v", err)
	}
	if purchase.Token != "token-123" {
		t.Fatalf("Token = %q, want token-123", purchase.Token)
	}
	if getter.subscriptionOptions.Token != "token-123" {
		t.Fatalf("token = %q, want token-123", getter.subscriptionOptions.Token)
	}
}

func TestMutateSubscriptionPurchaseDryRunDoesNotCallMutator(t *testing.T) {
	result, err := MutateSubscriptionPurchase(context.Background(), nil, SubscriptionPurchaseMutationOptions{
		PackageName:    "com.example.app",
		SubscriptionID: "premium_monthly",
		Token:          "token-123",
		Action:         SubscriptionPurchaseMutationActionAcknowledge,
		DryRun:         true,
	})
	if err != nil {
		t.Fatalf("MutateSubscriptionPurchase() error = %v", err)
	}
	if result.Applied {
		t.Fatalf("Applied = true, want false")
	}
	if result.Action != SubscriptionPurchaseMutationActionAcknowledge || result.Plan.Action != SubscriptionPurchaseMutationActionAcknowledge {
		t.Fatalf("result = %#v, want acknowledge action", result)
	}
}

func TestMutateSubscriptionPurchaseCancelDryRunIncludesCancellationType(t *testing.T) {
	result, err := MutateSubscriptionPurchase(context.Background(), nil, SubscriptionPurchaseMutationOptions{
		PackageName:      "com.example.app",
		Token:            "token-123",
		Action:           SubscriptionPurchaseMutationActionCancel,
		CancellationType: SubscriptionCancellationTypeDeveloperRequestedStopPayments,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("MutateSubscriptionPurchase() error = %v", err)
	}
	if result.CancellationType != SubscriptionCancellationTypeDeveloperRequestedStopPayments {
		t.Fatalf("CancellationType = %q, want developer requested stop payments", result.CancellationType)
	}
	if !reflect.DeepEqual(result.Plan.Steps, []string{"cancel subscription purchase", "use v2 cancellation type developerRequestedStopPayments"}) {
		t.Fatalf("Steps = %#v", result.Plan.Steps)
	}
}

func TestMutateSubscriptionPurchasePassesOptionsToMutator(t *testing.T) {
	mutator := &fakePurchaseClient{}
	options := SubscriptionPurchaseMutationOptions{
		PackageName:      "com.example.app",
		Token:            "token-123",
		Action:           SubscriptionPurchaseMutationActionCancel,
		CancellationType: SubscriptionCancellationTypeUserRequestedStopRenewals,
		Confirm:          true,
	}

	result, err := MutateSubscriptionPurchase(context.Background(), mutator, options)
	if err != nil {
		t.Fatalf("MutateSubscriptionPurchase() error = %v", err)
	}
	if !result.Applied {
		t.Fatalf("Applied = false, want true")
	}
	if mutator.subscriptionCancelOptions != options {
		t.Fatalf("subscriptionCancelOptions = %#v, want %#v", mutator.subscriptionCancelOptions, options)
	}
}

func TestSubscriptionPurchaseMutationRejectsInvalidOptions(t *testing.T) {
	tests := []SubscriptionPurchaseMutationOptions{
		{},
		{PackageName: "bad", SubscriptionID: "premium_monthly", Token: "token-123", Action: SubscriptionPurchaseMutationActionAcknowledge, DryRun: true},
		{PackageName: "com.example.app", Token: "token-123", Action: SubscriptionPurchaseMutationActionAcknowledge, DryRun: true},
		{PackageName: "com.example.app", SubscriptionID: "premium_monthly", Action: SubscriptionPurchaseMutationActionAcknowledge, DryRun: true},
		{PackageName: "com.example.app", SubscriptionID: "premium_monthly", Token: "token-123", DryRun: true},
		{PackageName: "com.example.app", SubscriptionID: "premium_monthly", Token: "token-123", Action: SubscriptionPurchaseMutationActionCancel, DeveloperPayload: "payload", DryRun: true},
		{PackageName: "com.example.app", SubscriptionID: "premium_monthly", Token: "token-123", Action: SubscriptionPurchaseMutationActionCancel, CancellationType: SubscriptionCancellationTypeUserRequestedStopRenewals, DryRun: true},
		{PackageName: "com.example.app", Token: "token-123", Action: SubscriptionPurchaseMutationActionCancel, DryRun: true},
		{PackageName: "com.example.app", SubscriptionID: "premium_monthly", Token: "token-123", Action: SubscriptionPurchaseMutationActionAcknowledge},
		{PackageName: "com.example.app", SubscriptionID: "premium_monthly", Token: "token-123", Action: SubscriptionPurchaseMutationActionAcknowledge, Confirm: true, DryRun: true},
	}
	for _, options := range tests {
		if _, err := MutateSubscriptionPurchase(context.Background(), nil, options); err == nil {
			t.Fatalf("MutateSubscriptionPurchase(%#v) expected validation error", options)
		}
	}
}

func TestRevokeSubscriptionPurchaseDryRunDoesNotCallRevoker(t *testing.T) {
	result, err := RevokeSubscriptionPurchase(context.Background(), nil, SubscriptionPurchaseRevokeOptions{
		PackageName: "com.example.app",
		Token:       "token-123",
		RefundType:  SubscriptionRefundTypeFull,
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("RevokeSubscriptionPurchase() error = %v", err)
	}
	if result.Applied {
		t.Fatalf("Applied = true, want false")
	}
	if result.RefundType != SubscriptionRefundTypeFull || len(result.Plan.Steps) != 2 {
		t.Fatalf("result = %#v, want full refund plan", result)
	}
}

func TestRevokeSubscriptionPurchaseItemRefundRequiresProductID(t *testing.T) {
	_, err := RevokeSubscriptionPurchase(context.Background(), nil, SubscriptionPurchaseRevokeOptions{
		PackageName: "com.example.app",
		Token:       "token-123",
		RefundType:  SubscriptionRefundTypeItem,
		DryRun:      true,
	})
	if err == nil {
		t.Fatal("expected item refund product ID validation error")
	}
}

func TestRevokeSubscriptionPurchaseItemRefundPlanIncludesProductID(t *testing.T) {
	result, err := RevokeSubscriptionPurchase(context.Background(), nil, SubscriptionPurchaseRevokeOptions{
		PackageName:     "com.example.app",
		Token:           "token-123",
		RefundType:      SubscriptionRefundTypeItem,
		RefundProductID: "premium_addon",
		DryRun:          true,
	})
	if err != nil {
		t.Fatalf("RevokeSubscriptionPurchase() error = %v", err)
	}
	if result.RefundProductID != "premium_addon" {
		t.Fatalf("RefundProductID = %q, want premium_addon", result.RefundProductID)
	}
	if !reflect.DeepEqual(result.Plan.Steps, []string{"revoke subscription purchase", "item refund", "refund subscription item premium_addon"}) {
		t.Fatalf("Steps = %#v", result.Plan.Steps)
	}
}

func TestRevokeSubscriptionPurchasePassesOptionsToRevoker(t *testing.T) {
	revoker := &fakePurchaseClient{}
	options := SubscriptionPurchaseRevokeOptions{
		PackageName: "com.example.app",
		Token:       "token-123",
		RefundType:  SubscriptionRefundTypeProrated,
		Confirm:     true,
	}

	result, err := RevokeSubscriptionPurchase(context.Background(), revoker, options)
	if err != nil {
		t.Fatalf("RevokeSubscriptionPurchase() error = %v", err)
	}
	if !result.Applied {
		t.Fatalf("Applied = false, want true")
	}
	if revoker.subscriptionRevokeOptions != options {
		t.Fatalf("subscriptionRevokeOptions = %#v, want %#v", revoker.subscriptionRevokeOptions, options)
	}
}

func TestRevokeSubscriptionPurchaseRejectsInvalidOptions(t *testing.T) {
	tests := []SubscriptionPurchaseRevokeOptions{
		{},
		{PackageName: "bad", Token: "token-123", RefundType: SubscriptionRefundTypeFull, DryRun: true},
		{PackageName: "com.example.app", RefundType: SubscriptionRefundTypeFull, DryRun: true},
		{PackageName: "com.example.app", Token: "token-123", DryRun: true},
		{PackageName: "com.example.app", Token: "token-123", RefundType: "partial", DryRun: true},
		{PackageName: "com.example.app", Token: "token-123", RefundType: SubscriptionRefundTypeFull, RefundProductID: "premium_addon", DryRun: true},
		{PackageName: "com.example.app", Token: "token-123", RefundType: SubscriptionRefundTypeFull},
		{PackageName: "com.example.app", Token: "token-123", RefundType: SubscriptionRefundTypeFull, Confirm: true, DryRun: true},
	}
	for _, options := range tests {
		if _, err := RevokeSubscriptionPurchase(context.Background(), nil, options); err == nil {
			t.Fatalf("RevokeSubscriptionPurchase(%#v) expected validation error", options)
		}
	}
}

func TestListVoidedPurchasesPassesOptionsToLister(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	lister := &fakePurchaseClient{voidedResult: VoidedPurchaseListResult{
		PackageName: packageName,
		Purchases:   []VoidedPurchase{{OrderID: "GPA.123"}},
	}}
	now := time.Now()
	options := VoidedPurchaseListOptions{
		PackageName:                       packageName,
		MaxResults:                        25,
		StartIndex:                        5,
		StartTimeMillis:                   now.Add(-time.Hour).UnixMilli(),
		EndTimeMillis:                     now.UnixMilli(),
		Type:                              VoidedPurchaseTypeProductsSubscriptions,
		IncludeQuantityBasedPartialRefund: true,
	}

	result, err := ListVoidedPurchases(context.Background(), lister, options)
	if err != nil {
		t.Fatalf("ListVoidedPurchases() error = %v", err)
	}
	if len(result.Purchases) != 1 {
		t.Fatalf("len(Purchases) = %d, want 1", len(result.Purchases))
	}
	if !reflect.DeepEqual(lister.voidedOptions, options) {
		t.Fatalf("options = %#v, want %#v", lister.voidedOptions, options)
	}
}

func TestListVoidedPurchasesRejectsInvalidOptions(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	now := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		options VoidedPurchaseListOptions
	}{
		{
			name:    "negative max results",
			options: VoidedPurchaseListOptions{PackageName: packageName, MaxResults: -1},
		},
		{
			name:    "negative start index",
			options: VoidedPurchaseListOptions{PackageName: packageName, StartIndex: -1},
		},
		{
			name:    "negative start time",
			options: VoidedPurchaseListOptions{PackageName: packageName, StartTimeMillis: -1},
		},
		{
			name:    "token with start time",
			options: VoidedPurchaseListOptions{PackageName: packageName, Token: "page", StartTimeMillis: now.UnixMilli()},
		},
		{
			name:    "token with end time",
			options: VoidedPurchaseListOptions{PackageName: packageName, Token: "page", EndTimeMillis: now.UnixMilli()},
		},
		{
			name:    "start after end",
			options: VoidedPurchaseListOptions{PackageName: packageName, StartTimeMillis: now.UnixMilli(), EndTimeMillis: now.Add(-time.Hour).UnixMilli()},
		},
		{
			name:    "start older than window",
			options: VoidedPurchaseListOptions{PackageName: packageName, StartTimeMillis: now.Add(-voidedPurchaseWindow - time.Millisecond).UnixMilli()},
		},
		{
			name:    "future end time",
			options: VoidedPurchaseListOptions{PackageName: packageName, EndTimeMillis: now.Add(time.Millisecond).UnixMilli()},
		},
		{
			name:    "invalid type",
			options: VoidedPurchaseListOptions{PackageName: packageName, Type: 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.options.ValidateAt(now)
			if err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestVoidedPurchaseListOptionsAcceptsValidTimeWindow(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	now := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)

	err = VoidedPurchaseListOptions{
		PackageName:     packageName,
		StartTimeMillis: now.Add(-voidedPurchaseWindow).UnixMilli(),
		EndTimeMillis:   now.UnixMilli(),
	}.ValidateAt(now)
	if err != nil {
		t.Fatalf("ValidateAt() error = %v", err)
	}
}

type fakePurchaseClient struct {
	productOptions                 ProductPurchaseOptions
	productPurchase                ProductPurchase
	acknowledgeOptions             ProductPurchaseMutationOptions
	consumeOptions                 ProductPurchaseMutationOptions
	subscriptionOptions            SubscriptionPurchaseOptions
	subscriptionPurchase           SubscriptionPurchase
	subscriptionAcknowledgeOptions SubscriptionPurchaseMutationOptions
	subscriptionCancelOptions      SubscriptionPurchaseMutationOptions
	subscriptionRevokeOptions      SubscriptionPurchaseRevokeOptions
	voidedOptions                  VoidedPurchaseListOptions
	voidedResult                   VoidedPurchaseListResult
}

func (c *fakePurchaseClient) GetProductPurchase(ctx context.Context, options ProductPurchaseOptions) (ProductPurchase, error) {
	c.productOptions = options
	return c.productPurchase, nil
}

func (c *fakePurchaseClient) AcknowledgeProductPurchase(ctx context.Context, options ProductPurchaseMutationOptions) error {
	c.acknowledgeOptions = options
	return nil
}

func (c *fakePurchaseClient) ConsumeProductPurchase(ctx context.Context, options ProductPurchaseMutationOptions) error {
	c.consumeOptions = options
	return nil
}

func (c *fakePurchaseClient) GetSubscriptionPurchase(ctx context.Context, options SubscriptionPurchaseOptions) (SubscriptionPurchase, error) {
	c.subscriptionOptions = options
	return c.subscriptionPurchase, nil
}

func (c *fakePurchaseClient) AcknowledgeSubscriptionPurchase(ctx context.Context, options SubscriptionPurchaseMutationOptions) error {
	c.subscriptionAcknowledgeOptions = options
	return nil
}

func (c *fakePurchaseClient) CancelSubscriptionPurchase(ctx context.Context, options SubscriptionPurchaseMutationOptions) error {
	c.subscriptionCancelOptions = options
	return nil
}

func (c *fakePurchaseClient) RevokeSubscriptionPurchase(ctx context.Context, options SubscriptionPurchaseRevokeOptions) error {
	c.subscriptionRevokeOptions = options
	return nil
}

func (c *fakePurchaseClient) ListVoidedPurchases(ctx context.Context, options VoidedPurchaseListOptions) (VoidedPurchaseListResult, error) {
	c.voidedOptions = options
	return c.voidedResult, nil
}
