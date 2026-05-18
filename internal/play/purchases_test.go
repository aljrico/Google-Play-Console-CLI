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
	productOptions       ProductPurchaseOptions
	productPurchase      ProductPurchase
	acknowledgeOptions   ProductPurchaseMutationOptions
	consumeOptions       ProductPurchaseMutationOptions
	subscriptionOptions  SubscriptionPurchaseOptions
	subscriptionPurchase SubscriptionPurchase
	voidedOptions        VoidedPurchaseListOptions
	voidedResult         VoidedPurchaseListResult
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

func (c *fakePurchaseClient) ListVoidedPurchases(ctx context.Context, options VoidedPurchaseListOptions) (VoidedPurchaseListResult, error) {
	c.voidedOptions = options
	return c.voidedResult, nil
}
