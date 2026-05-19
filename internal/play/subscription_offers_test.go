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
	for _, value := range []string{"", "Intro", "intro_offer", "intro offer"} {
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

type fakeSubscriptionOfferClient struct {
	listOptions  SubscriptionOfferListOptions
	listResult   SubscriptionOfferListResult
	batchOptions SubscriptionOfferBatchGetOptions
	batchResult  SubscriptionOfferBatchGetResult
	offerID      SubscriptionOfferID
	offer        SubscriptionOffer
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
