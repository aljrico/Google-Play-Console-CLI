package play

import (
	"context"
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

type fakeOneTimeProductOfferClient struct {
	listOptions OneTimeProductOfferListOptions
	listResult  OneTimeProductOfferListResult
	getOptions  OneTimeProductOfferGetOptions
	offer       OneTimeProductOffer
}

func (c *fakeOneTimeProductOfferClient) ListOneTimeProductOffers(ctx context.Context, options OneTimeProductOfferListOptions) (OneTimeProductOfferListResult, error) {
	c.listOptions = options
	return c.listResult, nil
}

func (c *fakeOneTimeProductOfferClient) GetOneTimeProductOffer(ctx context.Context, options OneTimeProductOfferGetOptions) (OneTimeProductOffer, error) {
	c.getOptions = options
	return c.offer, nil
}
