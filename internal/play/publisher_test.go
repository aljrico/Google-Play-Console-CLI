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

func containsField(fields []string, field string) bool {
	for _, candidate := range fields {
		if candidate == field {
			return true
		}
	}
	return false
}
