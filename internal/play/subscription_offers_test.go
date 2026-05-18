package play

import (
	"context"
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

type fakeSubscriptionOfferClient struct {
	listOptions SubscriptionOfferListOptions
	listResult  SubscriptionOfferListResult
	offerID     SubscriptionOfferID
	offer       SubscriptionOffer
}

func (c *fakeSubscriptionOfferClient) ListSubscriptionOffers(ctx context.Context, options SubscriptionOfferListOptions) (SubscriptionOfferListResult, error) {
	c.listOptions = options
	return c.listResult, nil
}

func (c *fakeSubscriptionOfferClient) GetSubscriptionOffer(ctx context.Context, packageName PackageName, productID SubscriptionProductID, basePlanID SubscriptionBasePlanID, offerID SubscriptionOfferID) (SubscriptionOffer, error) {
	c.offerID = offerID
	return c.offer, nil
}
