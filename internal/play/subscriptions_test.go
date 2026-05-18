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

type fakeSubscriptionClient struct {
	listOptions  SubscriptionListOptions
	listResult   SubscriptionListResult
	productID    SubscriptionProductID
	subscription Subscription
}

func (c *fakeSubscriptionClient) ListSubscriptions(ctx context.Context, options SubscriptionListOptions) (SubscriptionListResult, error) {
	c.listOptions = options
	return c.listResult, nil
}

func (c *fakeSubscriptionClient) GetSubscription(ctx context.Context, packageName PackageName, productID SubscriptionProductID) (Subscription, error) {
	c.productID = productID
	return c.subscription, nil
}
