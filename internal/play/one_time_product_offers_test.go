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

func TestUpdateOneTimeProductOfferStateDryRunBuildsPlanWithoutUpdater(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := UpdateOneTimeProductOfferState(context.Background(), nil, OneTimeProductOfferStateUpdateOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		OfferID:          "intro",
		Action:           OneTimeProductOfferStateActionCancel,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("UpdateOneTimeProductOfferState() error = %v", err)
	}
	if result.Applied {
		t.Fatal("Applied = true, want dry-run result")
	}
	if !reflect.DeepEqual(result.Plan.Steps, []string{"plan cancel one-time product offer"}) {
		t.Fatalf("Steps = %#v", result.Plan.Steps)
	}
}

func TestUpdateOneTimeProductOfferStateRequiresConfirmOrDryRun(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = NewOneTimeProductOfferStateUpdatePlan(OneTimeProductOfferStateUpdateOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		OfferID:          "intro",
		Action:           OneTimeProductOfferStateActionDeactivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
	})
	if err == nil {
		t.Fatal("expected confirmation validation error")
	}
}

func TestUpdateOneTimeProductOfferStatePassesOptionsToUpdater(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	updater := &fakeOneTimeProductOfferClient{
		offer: OneTimeProductOffer{OfferID: "intro"},
	}

	result, err := UpdateOneTimeProductOfferState(context.Background(), updater, OneTimeProductOfferStateUpdateOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		OfferID:          "intro",
		Action:           OneTimeProductOfferStateActionActivate,
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("UpdateOneTimeProductOfferState() error = %v", err)
	}
	if !result.Applied {
		t.Fatal("Applied = false, want applied result")
	}
	if updater.stateOptions.Action != OneTimeProductOfferStateActionActivate {
		t.Fatalf("Action = %q, want activate", updater.stateOptions.Action)
	}
	if updater.stateOptions.LatencyTolerance != ProductUpdateLatencyToleranceTolerant {
		t.Fatalf("LatencyTolerance = %q, want tolerant", updater.stateOptions.LatencyTolerance)
	}
}

type fakeOneTimeProductOfferClient struct {
	listOptions  OneTimeProductOfferListOptions
	listResult   OneTimeProductOfferListResult
	getOptions   OneTimeProductOfferGetOptions
	stateOptions OneTimeProductOfferStateUpdateOptions
	offer        OneTimeProductOffer
}

func (c *fakeOneTimeProductOfferClient) ListOneTimeProductOffers(ctx context.Context, options OneTimeProductOfferListOptions) (OneTimeProductOfferListResult, error) {
	c.listOptions = options
	return c.listResult, nil
}

func (c *fakeOneTimeProductOfferClient) GetOneTimeProductOffer(ctx context.Context, options OneTimeProductOfferGetOptions) (OneTimeProductOffer, error) {
	c.getOptions = options
	return c.offer, nil
}

func (c *fakeOneTimeProductOfferClient) UpdateOneTimeProductOfferState(ctx context.Context, options OneTimeProductOfferStateUpdateOptions) (OneTimeProductOffer, error) {
	c.stateOptions = options
	return c.offer, nil
}
