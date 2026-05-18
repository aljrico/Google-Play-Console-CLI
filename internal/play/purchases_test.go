package play

import (
	"context"
	"testing"
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

type fakePurchaseClient struct {
	productOptions       ProductPurchaseOptions
	productPurchase      ProductPurchase
	subscriptionOptions  SubscriptionPurchaseOptions
	subscriptionPurchase SubscriptionPurchase
}

func (c *fakePurchaseClient) GetProductPurchase(ctx context.Context, options ProductPurchaseOptions) (ProductPurchase, error) {
	c.productOptions = options
	return c.productPurchase, nil
}

func (c *fakePurchaseClient) GetSubscriptionPurchase(ctx context.Context, options SubscriptionPurchaseOptions) (SubscriptionPurchase, error) {
	c.subscriptionOptions = options
	return c.subscriptionPurchase, nil
}
