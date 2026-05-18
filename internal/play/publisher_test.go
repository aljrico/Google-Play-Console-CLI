package play

import (
	"testing"

	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/googleapi"
)

func TestUpsertTrackReleaseAppendsNewVersionCode(t *testing.T) {
	track := &androidpublisher.Track{
		Releases: []*androidpublisher.TrackRelease{
			{VersionCodes: googleapi.Int64s([]int64{41})},
		},
	}

	upsertTrackRelease(track, &androidpublisher.TrackRelease{VersionCodes: googleapi.Int64s([]int64{42})})

	if len(track.Releases) != 2 {
		t.Fatalf("len(Releases) = %d, want 2", len(track.Releases))
	}
}

func TestUpsertTrackReleaseReplacesExistingVersionCode(t *testing.T) {
	track := &androidpublisher.Track{
		Releases: []*androidpublisher.TrackRelease{
			{Name: "old", VersionCodes: googleapi.Int64s([]int64{42})},
		},
	}

	upsertTrackRelease(track, &androidpublisher.TrackRelease{Name: "new", VersionCodes: googleapi.Int64s([]int64{42})})

	if len(track.Releases) != 1 {
		t.Fatalf("len(Releases) = %d, want 1", len(track.Releases))
	}
	if track.Releases[0].Name != "new" {
		t.Fatalf("release name = %q, want new", track.Releases[0].Name)
	}
}

func TestSetReleaseStatusClearsUserFractionWhenCompleted(t *testing.T) {
	release := &androidpublisher.TrackRelease{
		Status:          ReleaseStatusInProgress.String(),
		UserFraction:    0.25,
		ForceSendFields: []string{"UserFraction"},
	}

	setReleaseStatus(release, ReleaseStatusCompleted, nil)

	if release.Status != ReleaseStatusCompleted.String() {
		t.Fatalf("Status = %q, want %q", release.Status, ReleaseStatusCompleted)
	}
	if release.UserFraction != 0 {
		t.Fatalf("UserFraction = %f, want 0", release.UserFraction)
	}
	for _, field := range release.ForceSendFields {
		if field == "UserFraction" {
			t.Fatal("UserFraction remained in ForceSendFields")
		}
	}
}

func TestListingToAPIForceSendsEmptyChangedField(t *testing.T) {
	apiListing := listingToAPI(Listing{
		Language: "en-US",
		Video:    stringValue(""),
	})

	if apiListing.Video != "" {
		t.Fatalf("Video = %q, want empty string", apiListing.Video)
	}
	if !containsField(apiListing.ForceSendFields, "Video") {
		t.Fatalf("ForceSendFields = %#v, want Video", apiListing.ForceSendFields)
	}
}

func TestAppDetailsToAPIForceSendsEmptyChangedField(t *testing.T) {
	apiDetails := appDetailsToAPI(AppDetails{ContactPhone: stringValue("")})

	if apiDetails.ContactPhone != "" {
		t.Fatalf("ContactPhone = %q, want empty string", apiDetails.ContactPhone)
	}
	if !containsField(apiDetails.ForceSendFields, "ContactPhone") {
		t.Fatalf("ForceSendFields = %#v, want ContactPhone", apiDetails.ForceSendFields)
	}
}

func TestReviewFromAPIMapsUserAndDeveloperComments(t *testing.T) {
	review := reviewFromAPI(&androidpublisher.Review{
		ReviewId:   "review-123",
		AuthorName: "A User",
		Comments: []*androidpublisher.Comment{
			{
				UserComment: &androidpublisher.UserComment{
					Text:             "Useful app.",
					StarRating:       5,
					ReviewerLanguage: "en",
					LastModified:     &androidpublisher.Timestamp{Seconds: 123},
					DeviceMetadata: &androidpublisher.DeviceMetadata{
						Manufacturer:     "Google",
						ProductName:      "Pixel",
						RamMb:            8192,
						ScreenDensityDpi: 420,
					},
				},
			},
			{
				DeveloperComment: &androidpublisher.DeveloperComment{
					Text:         "Thanks!",
					LastModified: &androidpublisher.Timestamp{Seconds: 456},
				},
			},
		},
	})

	if review.ReviewID != "review-123" {
		t.Fatalf("ReviewID = %q, want review-123", review.ReviewID)
	}
	if len(review.Comments) != 2 {
		t.Fatalf("len(Comments) = %d, want 2", len(review.Comments))
	}
	if review.Comments[0].Kind != ReviewCommentKindUser {
		t.Fatalf("first comment kind = %q, want user", review.Comments[0].Kind)
	}
	if review.Comments[0].User == nil || review.Comments[0].User.StarRating != 5 {
		t.Fatalf("first comment user = %#v, want 5-star user comment", review.Comments[0].User)
	}
	if review.Comments[0].User.DeviceMetadata == nil || review.Comments[0].User.DeviceMetadata.RAMMegabytes != 8192 {
		t.Fatalf("device metadata = %#v, want RAM metadata", review.Comments[0].User.DeviceMetadata)
	}
	if review.Comments[1].Kind != ReviewCommentKindDeveloper {
		t.Fatalf("second comment kind = %q, want developer", review.Comments[1].Kind)
	}
	if review.Comments[1].Developer == nil || review.Comments[1].Developer.LastEdited.Seconds != 456 {
		t.Fatalf("second comment developer = %#v, want timestamp 456", review.Comments[1].Developer)
	}
}

func TestReviewListResultFromAPIMapsPagination(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result := reviewListResultFromAPI(ReviewListOptions{PackageName: packageName}, &androidpublisher.ReviewsListResponse{
		PageInfo: &androidpublisher.PageInfo{
			ResultPerPage: 10,
			StartIndex:    20,
			TotalResults:  30,
		},
		TokenPagination: &androidpublisher.TokenPagination{
			NextPageToken:     "next",
			PreviousPageToken: "previous",
		},
	})

	if result.PageInfo == nil || result.PageInfo.TotalResults != 30 {
		t.Fatalf("PageInfo = %#v, want total results", result.PageInfo)
	}
	if result.Pagination == nil || result.Pagination.NextPageToken != "next" {
		t.Fatalf("Pagination = %#v, want next token", result.Pagination)
	}
}

func TestInAppProductFromAPIMapsCatalogFields(t *testing.T) {
	product := inAppProductFromAPI(&androidpublisher.InAppProduct{
		PackageName:     "com.example.app",
		Sku:             "coins_100",
		Status:          "active",
		PurchaseType:    "managedUser",
		DefaultLanguage: "en-US",
		DefaultPrice:    &androidpublisher.Price{Currency: "USD", PriceMicros: "1990000"},
		Prices: map[string]androidpublisher.Price{
			"US": {Currency: "USD", PriceMicros: "1990000"},
		},
		Listings: map[string]androidpublisher.InAppProductListing{
			"en-US": {Title: "100 coins", Description: "A small pack."},
		},
		ManagedProductTaxesAndComplianceSettings: &androidpublisher.ManagedProductTaxAndComplianceSettings{
			EeaWithdrawalRightType:  "WITHDRAWAL_RIGHT_DIGITAL_CONTENT",
			IsTokenizedDigitalAsset: true,
			TaxRateInfoByRegionCode: map[string]androidpublisher.RegionalTaxRateInfo{
				"US": {TaxTier: "TAX_TIER_NEWS_1", EligibleForStreamingServiceTaxRate: true},
			},
		},
	})

	if product.PackageName != "com.example.app" {
		t.Fatalf("PackageName = %q, want com.example.app", product.PackageName)
	}
	if product.SKU != "coins_100" {
		t.Fatalf("SKU = %q, want coins_100", product.SKU)
	}
	if product.DefaultPrice == nil || product.DefaultPrice.PriceMicros != "1990000" {
		t.Fatalf("DefaultPrice = %#v, want 1990000 micros", product.DefaultPrice)
	}
	if product.Prices["US"].Currency != "USD" {
		t.Fatalf("US price = %#v, want USD", product.Prices["US"])
	}
	if product.Listings["en-US"].Title != "100 coins" {
		t.Fatalf("listing = %#v, want title", product.Listings["en-US"])
	}
	if product.ManagedProductTaxAndComplianceSettings == nil {
		t.Fatal("ManagedProductTaxAndComplianceSettings = nil, want settings")
	}
	if !product.ManagedProductTaxAndComplianceSettings.IsTokenizedDigitalAsset {
		t.Fatal("IsTokenizedDigitalAsset = false, want true")
	}
	if product.ManagedProductTaxAndComplianceSettings.TaxRateInfoByRegionCode["US"].TaxTier != "TAX_TIER_NEWS_1" {
		t.Fatalf("TaxRateInfoByRegionCode = %#v, want US tax tier", product.ManagedProductTaxAndComplianceSettings.TaxRateInfoByRegionCode)
	}
}

func TestInAppProductListResultFromAPIMapsPagination(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result := inAppProductListResultFromAPI(InAppProductListOptions{PackageName: packageName}, &androidpublisher.InappproductsListResponse{
		Inappproduct: []*androidpublisher.InAppProduct{
			{PackageName: "com.example.app", Sku: "coins_100"},
		},
		TokenPagination: &androidpublisher.TokenPagination{
			NextPageToken: "next",
		},
	})

	if len(result.Products) != 1 {
		t.Fatalf("len(Products) = %d, want 1", len(result.Products))
	}
	if result.Pagination == nil || result.Pagination.NextPageToken != "next" {
		t.Fatalf("Pagination = %#v, want next token", result.Pagination)
	}
}

func TestSubscriptionFromAPIMapsListingsAndBasePlans(t *testing.T) {
	subscription := subscriptionFromAPI(&androidpublisher.Subscription{
		PackageName: "com.example.app",
		ProductId:   "premium",
		Listings: []*androidpublisher.SubscriptionListing{
			{LanguageCode: "en-US", Title: "Premium", Description: "All features"},
		},
		BasePlans: []*androidpublisher.BasePlan{
			{
				BasePlanId: "monthly",
				State:      "ACTIVE",
				AutoRenewingBasePlanType: &androidpublisher.AutoRenewingBasePlanType{
					BillingPeriodDuration:               "P1M",
					GracePeriodDuration:                 "P7D",
					AccountHoldDuration:                 "P30D",
					LegacyCompatible:                    true,
					LegacyCompatibleSubscriptionOfferId: "intro",
					ProrationMode:                       "SUBSCRIPTION_PRORATION_MODE_CHARGE_ON_NEXT_BILLING_DATE",
					ResubscribeState:                    "RESUBSCRIBE_STATE_ACTIVE",
				},
				OfferTags: []*androidpublisher.OfferTag{{Tag: "public"}},
				RegionalConfigs: []*androidpublisher.RegionalBasePlanConfig{
					{
						RegionCode:                "US",
						NewSubscriberAvailability: true,
						Price:                     &androidpublisher.Money{CurrencyCode: "USD", Units: 9, Nanos: 990000000},
					},
				},
				OtherRegionsConfig: &androidpublisher.OtherRegionsBasePlanConfig{
					NewSubscriberAvailability: true,
					UsdPrice:                  &androidpublisher.Money{CurrencyCode: "USD", Units: 9, Nanos: 990000000},
					EurPrice:                  &androidpublisher.Money{CurrencyCode: "EUR", Units: 8, Nanos: 990000000},
				},
			},
			{
				BasePlanId:          "prepaid",
				State:               "DRAFT",
				PrepaidBasePlanType: &androidpublisher.PrepaidBasePlanType{BillingPeriodDuration: "P1M", TimeExtension: "TIME_EXTENSION_ACTIVE"},
			},
			{
				BasePlanId: "installments",
				State:      "ACTIVE",
				InstallmentsBasePlanType: &androidpublisher.InstallmentsBasePlanType{
					BillingPeriodDuration:  "P1M",
					CommittedPaymentsCount: 12,
					RenewalType:            "RENEWAL_TYPE_RENEWS_WITHOUT_COMMITMENT",
				},
			},
		},
		RestrictedPaymentCountries: &androidpublisher.RestrictedPaymentCountries{
			RegionCodes: []string{"BR", "IN"},
		},
		TaxAndComplianceSettings: &androidpublisher.SubscriptionTaxAndComplianceSettings{
			EeaWithdrawalRightType:  "WITHDRAWAL_RIGHT_SERVICE",
			IsTokenizedDigitalAsset: true,
		},
	})

	if subscription.ProductID != "premium" {
		t.Fatalf("ProductID = %q, want premium", subscription.ProductID)
	}
	if len(subscription.Listings) != 1 || subscription.Listings[0].Title != "Premium" {
		t.Fatalf("Listings = %#v, want Premium listing", subscription.Listings)
	}
	if len(subscription.BasePlans) != 3 {
		t.Fatalf("len(BasePlans) = %d, want 3", len(subscription.BasePlans))
	}
	if subscription.BasePlans[0].Type != SubscriptionBasePlanTypeAutoRenewing {
		t.Fatalf("first base plan type = %q, want autoRenewing", subscription.BasePlans[0].Type)
	}
	if subscription.BasePlans[0].BillingPeriodDuration != "P1M" {
		t.Fatalf("billing duration = %q, want P1M", subscription.BasePlans[0].BillingPeriodDuration)
	}
	if subscription.BasePlans[0].OfferTags[0] != "public" {
		t.Fatalf("OfferTags = %#v, want public", subscription.BasePlans[0].OfferTags)
	}
	if len(subscription.BasePlans[0].RegionalConfigs) != 1 {
		t.Fatalf("RegionalConfigs = %#v, want 1 config", subscription.BasePlans[0].RegionalConfigs)
	}
	if subscription.BasePlans[0].RegionalConfigs[0].Price == nil || subscription.BasePlans[0].RegionalConfigs[0].Price.CurrencyCode != "USD" {
		t.Fatalf("RegionalConfigs = %#v, want USD price", subscription.BasePlans[0].RegionalConfigs)
	}
	if subscription.BasePlans[0].OtherRegionsConfig == nil || subscription.BasePlans[0].OtherRegionsConfig.USDPrice.Units != 9 {
		t.Fatalf("OtherRegionsConfig = %#v, want USD price", subscription.BasePlans[0].OtherRegionsConfig)
	}
	if subscription.BasePlans[0].LegacyCompatibleSubscriptionOfferID != "intro" {
		t.Fatalf("LegacyCompatibleSubscriptionOfferID = %q, want intro", subscription.BasePlans[0].LegacyCompatibleSubscriptionOfferID)
	}
	if subscription.BasePlans[1].Type != SubscriptionBasePlanTypePrepaid {
		t.Fatalf("second base plan type = %q, want prepaid", subscription.BasePlans[1].Type)
	}
	if subscription.BasePlans[1].TimeExtension != "TIME_EXTENSION_ACTIVE" {
		t.Fatalf("TimeExtension = %q, want TIME_EXTENSION_ACTIVE", subscription.BasePlans[1].TimeExtension)
	}
	if subscription.BasePlans[2].Type != SubscriptionBasePlanTypeInstallments {
		t.Fatalf("third base plan type = %q, want installments", subscription.BasePlans[2].Type)
	}
	if subscription.BasePlans[2].CommittedPaymentsCount != 12 {
		t.Fatalf("CommittedPaymentsCount = %d, want 12", subscription.BasePlans[2].CommittedPaymentsCount)
	}
	if len(subscription.RestrictedCountries) != 2 {
		t.Fatalf("RestrictedCountries = %#v, want two countries", subscription.RestrictedCountries)
	}
	if subscription.TaxAndComplianceSettings == nil || !subscription.TaxAndComplianceSettings.IsTokenizedDigitalAsset {
		t.Fatalf("TaxAndComplianceSettings = %#v, want tokenized settings", subscription.TaxAndComplianceSettings)
	}
}

func TestSubscriptionListResultFromAPIMapsNextPageToken(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result := subscriptionListResultFromAPI(SubscriptionListOptions{PackageName: packageName}, &androidpublisher.ListSubscriptionsResponse{
		NextPageToken: "next",
		Subscriptions: []*androidpublisher.Subscription{
			{PackageName: "com.example.app", ProductId: "premium"},
		},
	})

	if len(result.Subscriptions) != 1 {
		t.Fatalf("len(Subscriptions) = %d, want 1", len(result.Subscriptions))
	}
	if result.NextPageToken != "next" {
		t.Fatalf("NextPageToken = %q, want next", result.NextPageToken)
	}
}

func TestSubscriptionOfferFromAPIMapsCatalogFields(t *testing.T) {
	offer := subscriptionOfferFromAPI(&androidpublisher.SubscriptionOffer{
		PackageName: "com.example.app",
		ProductId:   "premium",
		BasePlanId:  "monthly",
		OfferId:     "intro",
		State:       "ACTIVE",
		OfferTags:   []*androidpublisher.OfferTag{{Tag: "public"}},
		RegionalConfigs: []*androidpublisher.RegionalSubscriptionOfferConfig{
			{RegionCode: "US", NewSubscriberAvailability: true},
		},
		OtherRegionsConfig: &androidpublisher.OtherRegionsSubscriptionOfferConfig{OtherRegionsNewSubscriberAvailability: true},
		Phases: []*androidpublisher.SubscriptionOfferPhase{
			{
				Duration:        "P1M",
				RecurrenceCount: 1,
				RegionalConfigs: []*androidpublisher.RegionalSubscriptionOfferPhaseConfig{
					{
						RegionCode:       "US",
						AbsoluteDiscount: &androidpublisher.Money{CurrencyCode: "USD", Units: 4, Nanos: 990000000},
					},
					{
						RegionCode: "BR",
						Free:       &androidpublisher.RegionalSubscriptionOfferPhaseFreePriceOverride{},
					},
				},
				OtherRegionsConfig: &androidpublisher.OtherRegionsSubscriptionOfferPhaseConfig{
					RelativeDiscount: 0.5,
				},
			},
		},
		Targeting: &androidpublisher.SubscriptionOfferTargeting{
			AcquisitionRule: &androidpublisher.AcquisitionTargetingRule{
				Scope: &androidpublisher.TargetingRuleScope{ThisSubscription: &androidpublisher.TargetingRuleScopeThisSubscription{}},
			},
			UpgradeRule: &androidpublisher.UpgradeTargetingRule{
				BillingPeriodDuration: "P1M",
				OncePerUser:           true,
				Scope:                 &androidpublisher.TargetingRuleScope{SpecificSubscriptionInApp: "premium"},
			},
		},
	})

	if offer.OfferID != "intro" {
		t.Fatalf("OfferID = %q, want intro", offer.OfferID)
	}
	if offer.OfferTags[0] != "public" {
		t.Fatalf("OfferTags = %#v, want public", offer.OfferTags)
	}
	if len(offer.RegionalConfigs) != 1 || !offer.RegionalConfigs[0].NewSubscriberAvailability {
		t.Fatalf("RegionalConfigs = %#v, want available US config", offer.RegionalConfigs)
	}
	if offer.OtherRegionsConfig == nil || !offer.OtherRegionsConfig.NewSubscriberAvailability {
		t.Fatalf("OtherRegionsConfig = %#v, want available other regions", offer.OtherRegionsConfig)
	}
	if len(offer.Phases) != 1 || len(offer.Phases[0].RegionalConfigs) != 2 {
		t.Fatalf("Phases = %#v, want phase configs", offer.Phases)
	}
	if offer.Phases[0].RegionalConfigs[0].AbsoluteDiscount == nil || offer.Phases[0].RegionalConfigs[0].AbsoluteDiscount.Units != 4 {
		t.Fatalf("AbsoluteDiscount = %#v, want four units", offer.Phases[0].RegionalConfigs[0].AbsoluteDiscount)
	}
	if !offer.Phases[0].RegionalConfigs[1].Free {
		t.Fatalf("second regional config = %#v, want free", offer.Phases[0].RegionalConfigs[1])
	}
	if offer.Targeting == nil || offer.Targeting.Acquisition.Scope == nil || !offer.Targeting.Acquisition.Scope.ThisSubscription {
		t.Fatalf("Targeting = %#v, want acquisition this-subscription scope", offer.Targeting)
	}
	if offer.Targeting.Upgrade == nil || !offer.Targeting.Upgrade.OncePerUser {
		t.Fatalf("Upgrade targeting = %#v, want once per user", offer.Targeting.Upgrade)
	}
}

func TestSubscriptionOfferListResultFromAPIMapsNextPageToken(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result := subscriptionOfferListResultFromAPI(SubscriptionOfferListOptions{
		PackageName: packageName,
		ProductID:   "premium",
		BasePlanID:  "monthly",
	}, &androidpublisher.ListSubscriptionOffersResponse{
		NextPageToken: "next",
		SubscriptionOffers: []*androidpublisher.SubscriptionOffer{
			{PackageName: "com.example.app", ProductId: "premium", BasePlanId: "monthly", OfferId: "intro"},
		},
	})

	if len(result.Offers) != 1 {
		t.Fatalf("len(Offers) = %d, want 1", len(result.Offers))
	}
	if result.NextPageToken != "next" {
		t.Fatalf("NextPageToken = %q, want next", result.NextPageToken)
	}
}

func containsField(fields []string, field string) bool {
	for _, candidate := range fields {
		if candidate == field {
			return true
		}
	}
	return false
}
