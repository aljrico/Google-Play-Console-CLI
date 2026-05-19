package play

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
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

func TestReleaseToAPIMapsReleaseNotes(t *testing.T) {
	apiRelease := releaseToAPI(TrackRelease{
		Name:         "1.2.3",
		Status:       ReleaseStatusCompleted,
		VersionCodes: []int64{42},
		ReleaseNotes: []ReleaseNote{
			{Language: "en-US", Text: "Bug fixes."},
			{Language: "es-ES", Text: "Correcciones."},
		},
	})

	if len(apiRelease.ReleaseNotes) != 2 {
		t.Fatalf("len(ReleaseNotes) = %d, want 2", len(apiRelease.ReleaseNotes))
	}
	if apiRelease.ReleaseNotes[0].Language != "en-US" || apiRelease.ReleaseNotes[0].Text != "Bug fixes." {
		t.Fatalf("first note = %#v, want en-US bug fixes", apiRelease.ReleaseNotes[0])
	}
}

func TestReleaseFromAPIMapsReleaseNotes(t *testing.T) {
	release := releaseFromAPI(&androidpublisher.TrackRelease{
		Name:         "1.2.3",
		Status:       "completed",
		VersionCodes: googleapi.Int64s([]int64{42}),
		ReleaseNotes: []*androidpublisher.LocalizedText{
			{Language: "en-US", Text: "Bug fixes."},
		},
	})

	if len(release.ReleaseNotes) != 1 {
		t.Fatalf("len(ReleaseNotes) = %d, want 1", len(release.ReleaseNotes))
	}
	if release.ReleaseNotes[0].Language != "en-US" || release.ReleaseNotes[0].Text != "Bug fixes." {
		t.Fatalf("release note = %#v, want en-US bug fixes", release.ReleaseNotes[0])
	}
}

func TestAppendTrackReleaseSendsReleaseNotes(t *testing.T) {
	var sawUpdate bool
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/edits/edit-123/tracks/internal" {
			t.Fatalf("path = %q, want internal track endpoint", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"track":"internal","releases":[]}`))
		case http.MethodPut:
			sawUpdate = true
			var apiTrack androidpublisher.Track
			if err := json.NewDecoder(r.Body).Decode(&apiTrack); err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if len(apiTrack.Releases) != 1 || len(apiTrack.Releases[0].ReleaseNotes) != 1 {
				t.Fatalf("Track body = %#v, want release note", apiTrack)
			}
			note := apiTrack.Releases[0].ReleaseNotes[0]
			if note.Language != "en-US" || note.Text != "Bug fixes." {
				t.Fatalf("release note = %#v, want en-US bug fixes", note)
			}
			_, _ = w.Write([]byte(`{"track":"internal","releases":[]}`))
		default:
			t.Fatalf("method = %s, want GET or PUT", r.Method)
		}
	}))

	_, err := publisher.AppendTrackRelease(context.Background(), "com.example.app", "edit-123", TrackInternal, TrackRelease{
		Name:         "1.2.3",
		Status:       ReleaseStatusCompleted,
		VersionCodes: []int64{42},
		ReleaseNotes: []ReleaseNote{
			{Language: "en-US", Text: "Bug fixes."},
		},
	})
	if err != nil {
		t.Fatalf("AppendTrackRelease() error = %v", err)
	}
	if !sawUpdate {
		t.Fatal("expected track update")
	}
}

func TestPromoteTrackReleaseOverridesReleaseNotes(t *testing.T) {
	var sawUpdate bool
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/androidpublisher/v3/applications/com.example.app/edits/edit-123/tracks/internal":
			_, _ = w.Write([]byte(`{"track":"internal","releases":[{"status":"completed","versionCodes":["42"],"releaseNotes":[{"language":"en-US","text":"Internal note."}]}]}`))
		case r.Method == http.MethodGet && r.URL.Path == "/androidpublisher/v3/applications/com.example.app/edits/edit-123/tracks/production":
			_, _ = w.Write([]byte(`{"track":"production","releases":[]}`))
		case r.Method == http.MethodPut && r.URL.Path == "/androidpublisher/v3/applications/com.example.app/edits/edit-123/tracks/production":
			sawUpdate = true
			var apiTrack androidpublisher.Track
			if err := json.NewDecoder(r.Body).Decode(&apiTrack); err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if len(apiTrack.Releases) != 1 || len(apiTrack.Releases[0].ReleaseNotes) != 1 {
				t.Fatalf("Track body = %#v, want replacement release note", apiTrack)
			}
			note := apiTrack.Releases[0].ReleaseNotes[0]
			if note.Language != "en-US" || note.Text != "Production note." {
				t.Fatalf("release note = %#v, want production note", note)
			}
			_, _ = w.Write([]byte(`{"track":"production","releases":[]}`))
		default:
			t.Fatalf("method/path = %s %s, want promotion track calls", r.Method, r.URL.Path)
		}
	}))

	release, err := publisher.PromoteTrackRelease(context.Background(), "com.example.app", "edit-123", TrackInternal, TrackProduction, 42, ReleaseStatusDraft, nil, []ReleaseNote{{Language: "en-US", Text: "Production note."}})
	if err != nil {
		t.Fatalf("PromoteTrackRelease() error = %v", err)
	}
	if len(release.ReleaseNotes) != 1 || release.ReleaseNotes[0].Text != "Production note." {
		t.Fatalf("ReleaseNotes = %#v, want production note", release.ReleaseNotes)
	}
	if !sawUpdate {
		t.Fatal("expected production track update")
	}
}

func TestGetTestersUsesTrackTesterEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/edits/edit-123/testers/internal" {
			t.Fatalf("path = %q, want tester endpoint", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"googleGroups":["qa@example.com"]}`))
	}))

	testers, err := publisher.GetTesters(context.Background(), "com.example.app", "edit-123", TrackInternal)
	if err != nil {
		t.Fatalf("GetTesters() error = %v", err)
	}
	if len(testers.GoogleGroups) != 1 || testers.GoogleGroups[0] != "qa@example.com" {
		t.Fatalf("GoogleGroups = %#v, want qa@example.com", testers.GoogleGroups)
	}
}

func TestUpdateTestersUsesTrackTesterEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("method = %s, want PUT", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/edits/edit-123/testers/beta" {
			t.Fatalf("path = %q, want tester endpoint", r.URL.Path)
		}
		var apiTesters androidpublisher.Testers
		if err := json.NewDecoder(r.Body).Decode(&apiTesters); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if !reflect.DeepEqual(apiTesters.GoogleGroups, []string{"alpha@example.com", "qa@example.com"}) {
			t.Fatalf("GoogleGroups body = %#v, want sorted groups", apiTesters.GoogleGroups)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"googleGroups":["alpha@example.com","qa@example.com"]}`))
	}))

	testers, err := publisher.UpdateTesters(context.Background(), "com.example.app", "edit-123", TrackBeta, []TesterGoogleGroup{
		"qa@example.com",
		"alpha@example.com",
	})
	if err != nil {
		t.Fatalf("UpdateTesters() error = %v", err)
	}
	if !reflect.DeepEqual(testers.GoogleGroups, []TesterGoogleGroup{"alpha@example.com", "qa@example.com"}) {
		t.Fatalf("GoogleGroups = %#v, want sorted groups", testers.GoogleGroups)
	}
}

func TestUpdateTestersForceSendsEmptyGoogleGroups(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("method = %s, want PUT", r.Method)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll() error = %v", err)
		}
		if !strings.Contains(string(body), `"googleGroups":[]`) {
			t.Fatalf("body = %s, want empty googleGroups array", string(body))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"googleGroups":[]}`))
	}))

	testers, err := publisher.UpdateTesters(context.Background(), "com.example.app", "edit-123", TrackInternal, nil)
	if err != nil {
		t.Fatalf("UpdateTesters() error = %v", err)
	}
	if len(testers.GoogleGroups) != 0 {
		t.Fatalf("GoogleGroups = %#v, want empty", testers.GoogleGroups)
	}
}

func TestUploadInternalSharingBundleUsesUploadEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/upload/androidpublisher/v3/applications/internalappsharing/com.example.app/artifacts/bundle" {
			t.Fatalf("path = %q, want internal sharing bundle upload endpoint", r.URL.Path)
		}
		if r.URL.Query().Get("uploadType") == "" {
			t.Fatalf("query = %q, want uploadType", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"downloadUrl":"https://example.com/download","sha256":"abc","certificateFingerprint":"def"}`))
	}))

	artifact, err := publisher.UploadInternalSharingBundle(context.Background(), "com.example.app", writeTestFile(t, "app.aab"))
	if err != nil {
		t.Fatalf("UploadInternalSharingBundle() error = %v", err)
	}
	if artifact.DownloadURL != "https://example.com/download" || artifact.SHA256 != "abc" {
		t.Fatalf("artifact = %#v, want download URL and sha", artifact)
	}
}

func TestUploadAPKUsesEditAPKUploadEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/upload/androidpublisher/v3/applications/com.example.app/edits/edit-123/apks" {
			t.Fatalf("path = %q, want APK upload endpoint", r.URL.Path)
		}
		if r.URL.Query().Get("uploadType") == "" {
			t.Fatalf("query = %q, want uploadType", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"versionCode":43,"binary":{"sha1":"abc","sha256":"def"}}`))
	}))

	apk, err := publisher.UploadAPK(context.Background(), "com.example.app", "edit-123", writeTestFile(t, "app.apk"))
	if err != nil {
		t.Fatalf("UploadAPK() error = %v", err)
	}
	if apk.VersionCode != 43 || apk.SHA256 != "def" {
		t.Fatalf("apk = %#v, want version code and sha", apk)
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

func TestListImagesUsesListingImagesEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/edits/edit-123/listings/en-US/phoneScreenshots" {
			t.Fatalf("path = %q, want images endpoint", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"images": [{
				"id": "image-1",
				"url": "https://example.com/image.png",
				"sha1": "abc",
				"sha256": "def"
			}]
		}`))
	}))

	images, err := publisher.ListImages(context.Background(), "com.example.app", "edit-123", "en-US", ImageTypePhoneScreenshots)
	if err != nil {
		t.Fatalf("ListImages() error = %v", err)
	}
	if len(images) != 1 || images[0].ID != "image-1" || images[0].SHA256 != "def" {
		t.Fatalf("images = %#v, want mapped image", images)
	}
}

func TestUploadImageUsesListingImageUploadEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/upload/androidpublisher/v3/applications/com.example.app/edits/edit-123/listings/en-US/featureGraphic" {
			t.Fatalf("path = %q, want image upload endpoint", r.URL.Path)
		}
		if r.URL.Query().Get("uploadType") == "" {
			t.Fatalf("query = %q, want uploadType", r.URL.RawQuery)
		}
		if r.Header.Get("Content-Type") == "" {
			t.Fatal("missing content type")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"image":{"id":"image-1","url":"https://example.com/feature.png","sha256":"abc"}}`))
	}))

	image, err := publisher.UploadImage(context.Background(), "com.example.app", "edit-123", "en-US", ImageTypeFeatureGraphic, writeTestFile(t, "feature.png"))
	if err != nil {
		t.Fatalf("UploadImage() error = %v", err)
	}
	if image.ID != "image-1" || image.SHA256 != "abc" {
		t.Fatalf("image = %#v, want uploaded image", image)
	}
}

func TestUploadImageRejectsNoOpResponse(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))

	_, err := publisher.UploadImage(context.Background(), "com.example.app", "edit-123", "en-US", ImageTypeFeatureGraphic, writeTestFile(t, "feature.png"))
	if err == nil {
		t.Fatal("expected no-op upload error")
	}
	if !strings.Contains(err.Error(), "produced no image") {
		t.Fatalf("error = %v, want no-op upload context", err)
	}
}

func TestDeleteImageUsesListingImageEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/edits/edit-123/listings/en-US/phoneScreenshots/image-1" {
			t.Fatalf("path = %q, want image delete endpoint", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Fatalf("method = %s, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))

	if err := publisher.DeleteImage(context.Background(), "com.example.app", "edit-123", "en-US", ImageTypePhoneScreenshots, "image-1"); err != nil {
		t.Fatalf("DeleteImage() error = %v", err)
	}
}

func TestDeleteAllImagesUsesListingImagesEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/edits/edit-123/listings/en-US/featureGraphic" {
			t.Fatalf("path = %q, want images delete-all endpoint", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Fatalf("method = %s, want DELETE", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"deleted":[{"id":"image-1","sha256":"abc"}]}`))
	}))

	deleted, err := publisher.DeleteAllImages(context.Background(), "com.example.app", "edit-123", "en-US", ImageTypeFeatureGraphic)
	if err != nil {
		t.Fatalf("DeleteAllImages() error = %v", err)
	}
	if len(deleted) != 1 || deleted[0].ID != "image-1" || deleted[0].SHA256 != "abc" {
		t.Fatalf("deleted = %#v, want deleted image", deleted)
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

func TestGooglePublisherPatchInAppProductSendsStatusPatch(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("method = %s, want PATCH", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/inappproducts/coins_100" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		var request androidpublisher.InAppProduct
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if request.Status != "inactive" || request.Sku != "coins_100" || request.PackageName != "com.example.app" {
			t.Fatalf("request = %#v, want status patch", request)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"packageName":"com.example.app","sku":"coins_100","status":"inactive"}`)
	}))

	product, err := publisher.PatchInAppProduct(context.Background(), InAppProductPatchOptions{
		PackageName: "com.example.app",
		SKU:         "coins_100",
		Status:      ProductStatusInactive,
		Confirm:     true,
	})
	if err != nil {
		t.Fatalf("PatchInAppProduct() error = %v", err)
	}
	if product.Status != ProductStatusInactive {
		t.Fatalf("Status = %q, want inactive", product.Status)
	}
}

func TestGooglePublisherPatchInAppProductRejectsDryRunBeforeRequest(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))

	_, err := publisher.PatchInAppProduct(context.Background(), InAppProductPatchOptions{
		PackageName: "com.example.app",
		SKU:         "coins_100",
		Status:      ProductStatusInactive,
		DryRun:      true,
	})
	if err == nil {
		t.Fatal("expected live validation error")
	}
}

func TestGooglePublisherPatchInAppProductRequiresConfirmBeforeRequest(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))

	_, err := publisher.PatchInAppProduct(context.Background(), InAppProductPatchOptions{
		PackageName: "com.example.app",
		SKU:         "coins_100",
		Status:      ProductStatusInactive,
	})
	if err == nil {
		t.Fatal("expected live validation error")
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

func TestProductPurchaseFromAPIMapsEntitlementFields(t *testing.T) {
	purchase := productPurchaseFromAPI(ProductPurchaseOptions{
		PackageName: "com.example.app",
		ProductID:   "coins_100",
		Token:       "token-123",
	}, &androidpublisher.ProductPurchaseV2{
		OrderId:                     "GPA.123",
		PurchaseStateContext:        &androidpublisher.PurchaseStateContext{PurchaseState: "PURCHASED"},
		PurchaseCompletionTime:      "2026-05-18T10:00:00Z",
		AcknowledgementState:        "ACKNOWLEDGEMENT_STATE_PENDING",
		RegionCode:                  "US",
		ObfuscatedExternalAccountId: "account",
		TestPurchaseContext:         &androidpublisher.TestPurchaseContext{},
		ProductLineItem: []*androidpublisher.ProductLineItem{
			{
				ProductId: "coins_100",
				ProductOfferDetails: &androidpublisher.ProductOfferDetails{
					ConsumptionState:   "CONSUMPTION_STATE_YET_TO_BE_CONSUMED",
					OfferId:            "starter",
					OfferTags:          []string{"welcome"},
					Quantity:           0,
					RefundableQuantity: 1,
				},
			},
		},
	})

	if purchase.PackageName != "com.example.app" {
		t.Fatalf("PackageName = %q, want com.example.app", purchase.PackageName)
	}
	if purchase.ProductID != "coins_100" {
		t.Fatalf("ProductID = %q, want coins_100", purchase.ProductID)
	}
	if purchase.Token != "token-123" {
		t.Fatalf("Token = %q, want token-123", purchase.Token)
	}
	if purchase.PurchaseState != "PURCHASED" {
		t.Fatalf("PurchaseState = %q, want PURCHASED", purchase.PurchaseState)
	}
	if !purchase.TestPurchase {
		t.Fatal("TestPurchase = false, want true")
	}
	if len(purchase.LineItems) != 1 {
		t.Fatalf("len(LineItems) = %d, want 1", len(purchase.LineItems))
	}
	if purchase.LineItems[0].Quantity != 0 {
		t.Fatalf("Quantity = %d, want preserved zero", purchase.LineItems[0].Quantity)
	}
	if purchase.LineItems[0].ConsumptionState != "CONSUMPTION_STATE_YET_TO_BE_CONSUMED" {
		t.Fatalf("ConsumptionState = %q, want yet-to-be-consumed", purchase.LineItems[0].ConsumptionState)
	}
}

func TestGetProductPurchaseUsesProductV2TokenEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/purchases/productsv2/tokens/token-123" {
			t.Fatalf("path = %q, want productsv2 token endpoint", r.URL.Path)
		}
		if r.URL.Query().Has("productId") {
			t.Fatalf("query = %q, did not expect legacy product ID param", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"orderId": "GPA.123",
			"purchaseStateContext": {"purchaseState": "PURCHASED"},
			"productLineItem": [
				{
					"productId": "coins_100",
					"productOfferDetails": {
						"quantity": 2,
						"refundableQuantity": 1
					}
				}
			]
		}`))
	}))

	purchase, err := publisher.GetProductPurchase(context.Background(), ProductPurchaseOptions{
		PackageName: "com.example.app",
		Token:       "token-123",
	})
	if err != nil {
		t.Fatalf("GetProductPurchase() error = %v", err)
	}
	if purchase.ProductID != "coins_100" {
		t.Fatalf("ProductID = %q, want coins_100", purchase.ProductID)
	}
	if purchase.PurchaseState != "PURCHASED" {
		t.Fatalf("PurchaseState = %q, want PURCHASED", purchase.PurchaseState)
	}
	if len(purchase.LineItems) != 1 || purchase.LineItems[0].Quantity != 2 {
		t.Fatalf("LineItems = %#v, want quantity 2", purchase.LineItems)
	}
}

func TestAcknowledgeProductPurchaseUsesLegacyProductEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/purchases/products/coins_100/tokens/token-123:acknowledge" {
			t.Fatalf("path = %q, want product acknowledge endpoint", r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll() error = %v", err)
		}
		if !strings.Contains(string(body), `"developerPayload":"order-7"`) {
			t.Fatalf("body = %q, want developer payload", string(body))
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	err := publisher.AcknowledgeProductPurchase(context.Background(), ProductPurchaseMutationOptions{
		PackageName:      "com.example.app",
		ProductID:        "coins_100",
		Token:            "token-123",
		DeveloperPayload: "order-7",
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("AcknowledgeProductPurchase() error = %v", err)
	}
}

func TestAcknowledgeProductPurchaseRejectsDryRunBeforeRequest(t *testing.T) {
	requests := 0
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		t.Fatalf("unexpected request to %s", r.URL.Path)
	}))

	err := publisher.AcknowledgeProductPurchase(context.Background(), ProductPurchaseMutationOptions{
		PackageName: "com.example.app",
		ProductID:   "coins_100",
		Token:       "token-123",
		DryRun:      true,
	})
	if err == nil {
		t.Fatal("expected dry-run rejection")
	}
	if requests != 0 {
		t.Fatalf("requests = %d, want 0", requests)
	}
}

func TestConsumeProductPurchaseUsesLegacyProductEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/purchases/products/coins_100/tokens/token-123:consume" {
			t.Fatalf("path = %q, want product consume endpoint", r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll() error = %v", err)
		}
		if strings.TrimSpace(string(body)) != "" {
			t.Fatalf("body = %q, want empty body", string(body))
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	err := publisher.ConsumeProductPurchase(context.Background(), ProductPurchaseMutationOptions{
		PackageName: "com.example.app",
		ProductID:   "coins_100",
		Token:       "token-123",
		Confirm:     true,
	})
	if err != nil {
		t.Fatalf("ConsumeProductPurchase() error = %v", err)
	}
}

func TestConsumeProductPurchaseRejectsDryRunBeforeRequest(t *testing.T) {
	requests := 0
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		t.Fatalf("unexpected request to %s", r.URL.Path)
	}))

	err := publisher.ConsumeProductPurchase(context.Background(), ProductPurchaseMutationOptions{
		PackageName: "com.example.app",
		ProductID:   "coins_100",
		Token:       "token-123",
		DryRun:      true,
	})
	if err == nil {
		t.Fatal("expected dry-run rejection")
	}
	if requests != 0 {
		t.Fatalf("requests = %d, want 0", requests)
	}
}

func TestRevokeSubscriptionPurchaseUsesSubscriptionV2Endpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/purchases/subscriptionsv2/tokens/token-123:revoke" {
			t.Fatalf("path = %q, want subscription v2 revoke endpoint", r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll() error = %v", err)
		}
		if !strings.Contains(string(body), `"fullRefund":{}`) {
			t.Fatalf("body = %q, want full refund context", string(body))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))

	err := publisher.RevokeSubscriptionPurchase(context.Background(), SubscriptionPurchaseRevokeOptions{
		PackageName: "com.example.app",
		Token:       "token-123",
		RefundType:  SubscriptionRefundTypeFull,
		Confirm:     true,
	})
	if err != nil {
		t.Fatalf("RevokeSubscriptionPurchase() error = %v", err)
	}
}

func TestRevokeSubscriptionPurchaseUsesItemRefundContext(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/purchases/subscriptionsv2/tokens/token-123:revoke" {
			t.Fatalf("path = %q, want subscription v2 revoke endpoint", r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll() error = %v", err)
		}
		if !strings.Contains(string(body), `"itemBasedRefund":{"productId":"premium_addon"}`) {
			t.Fatalf("body = %q, want item refund context", string(body))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))

	err := publisher.RevokeSubscriptionPurchase(context.Background(), SubscriptionPurchaseRevokeOptions{
		PackageName:     "com.example.app",
		Token:           "token-123",
		RefundType:      SubscriptionRefundTypeItem,
		RefundProductID: "premium_addon",
		Confirm:         true,
	})
	if err != nil {
		t.Fatalf("RevokeSubscriptionPurchase() error = %v", err)
	}
}

func TestAcknowledgeSubscriptionPurchaseUsesLegacyEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/purchases/subscriptions/premium_monthly/tokens/token-123:acknowledge" {
			t.Fatalf("path = %q, want legacy subscription acknowledge endpoint", r.URL.Path)
		}
		var request androidpublisher.SubscriptionPurchasesAcknowledgeRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if request.DeveloperPayload != "handled-by-gpc" {
			t.Fatalf("DeveloperPayload = %q, want handled-by-gpc", request.DeveloperPayload)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))

	err := publisher.AcknowledgeSubscriptionPurchase(context.Background(), SubscriptionPurchaseMutationOptions{
		PackageName:      "com.example.app",
		SubscriptionID:   "premium_monthly",
		Token:            "token-123",
		Action:           SubscriptionPurchaseMutationActionAcknowledge,
		DeveloperPayload: "handled-by-gpc",
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("AcknowledgeSubscriptionPurchase() error = %v", err)
	}
}

func TestCancelSubscriptionPurchaseUsesV2Endpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/purchases/subscriptionsv2/tokens/token-123:cancel" {
			t.Fatalf("path = %q, want subscription v2 cancel endpoint", r.URL.Path)
		}
		var request rawSubscriptionCancelRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if request.CancellationContext.CancellationType != "USER_REQUESTED_STOP_RENEWALS" {
			t.Fatalf("CancellationType = %q, want user requested stop renewals", request.CancellationContext.CancellationType)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))

	err := publisher.CancelSubscriptionPurchase(context.Background(), SubscriptionPurchaseMutationOptions{
		PackageName:      "com.example.app",
		Token:            "token-123",
		Action:           SubscriptionPurchaseMutationActionCancel,
		CancellationType: SubscriptionCancellationTypeUserRequestedStopRenewals,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("CancelSubscriptionPurchase() error = %v", err)
	}
}

func TestAcknowledgeSubscriptionPurchaseRejectsMismatchedActionBeforeRequest(t *testing.T) {
	requests := 0
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		t.Fatalf("unexpected request to %s", r.URL.Path)
	}))

	err := publisher.AcknowledgeSubscriptionPurchase(context.Background(), SubscriptionPurchaseMutationOptions{
		PackageName:      "com.example.app",
		SubscriptionID:   "premium_monthly",
		Token:            "token-123",
		Action:           SubscriptionPurchaseMutationActionCancel,
		CancellationType: SubscriptionCancellationTypeUserRequestedStopRenewals,
		Confirm:          true,
	})
	if err == nil {
		t.Fatal("expected action mismatch")
	}
	if requests != 0 {
		t.Fatalf("requests = %d, want 0", requests)
	}
}

func TestCancelSubscriptionPurchaseRejectsMismatchedActionBeforeRequest(t *testing.T) {
	requests := 0
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		t.Fatalf("unexpected request to %s", r.URL.Path)
	}))

	err := publisher.CancelSubscriptionPurchase(context.Background(), SubscriptionPurchaseMutationOptions{
		PackageName:    "com.example.app",
		SubscriptionID: "premium_monthly",
		Token:          "token-123",
		Action:         SubscriptionPurchaseMutationActionAcknowledge,
		Confirm:        true,
	})
	if err == nil {
		t.Fatal("expected action mismatch")
	}
	if requests != 0 {
		t.Fatalf("requests = %d, want 0", requests)
	}
}

func TestCancelSubscriptionPurchaseRejectsDryRunBeforeRequest(t *testing.T) {
	requests := 0
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		t.Fatalf("unexpected request to %s", r.URL.Path)
	}))

	err := publisher.CancelSubscriptionPurchase(context.Background(), SubscriptionPurchaseMutationOptions{
		PackageName:      "com.example.app",
		Token:            "token-123",
		Action:           SubscriptionPurchaseMutationActionCancel,
		CancellationType: SubscriptionCancellationTypeUserRequestedStopRenewals,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected dry-run rejection")
	}
	if requests != 0 {
		t.Fatalf("requests = %d, want 0", requests)
	}
}

func TestRevokeSubscriptionPurchaseRejectsDryRunBeforeRequest(t *testing.T) {
	requests := 0
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		t.Fatalf("unexpected request to %s", r.URL.Path)
	}))

	err := publisher.RevokeSubscriptionPurchase(context.Background(), SubscriptionPurchaseRevokeOptions{
		PackageName: "com.example.app",
		Token:       "token-123",
		RefundType:  SubscriptionRefundTypeProrated,
		DryRun:      true,
	})
	if err == nil {
		t.Fatal("expected dry-run rejection")
	}
	if requests != 0 {
		t.Fatalf("requests = %d, want 0", requests)
	}
}

func TestProductPurchaseJSONPreservesZeroQuantities(t *testing.T) {
	payload, err := json.Marshal(ProductPurchase{
		PackageName: "com.example.app",
		Token:       "token-123",
		LineItems: []ProductPurchaseLineItem{
			{
				ProductID:          "coins_100",
				Quantity:           0,
				RefundableQuantity: 0,
			},
		},
	})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	output := string(payload)
	for _, want := range []string{`"quantity":0`, `"refundableQuantity":0`} {
		if !strings.Contains(output, want) {
			t.Fatalf("json = %s, want %s", output, want)
		}
	}
}

func TestListVoidedPurchasesSendsExpectedQueryParams(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/purchases/voidedpurchases" {
			t.Fatalf("path = %q, want voided purchases endpoint", r.URL.Path)
		}
		query := r.URL.Query()
		assertQueryValue(t, query, "maxResults", "25")
		assertQueryValue(t, query, "startIndex", "5")
		assertQueryValue(t, query, "token", "page")
		assertQueryValue(t, query, "type", "1")
		assertQueryValue(t, query, "includeQuantityBasedPartialRefund", "true")
		if query.Has("startTime") || query.Has("endTime") {
			t.Fatalf("query = %q, did not expect time params with token", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"tokenPagination": {"nextPageToken": "next"},
			"voidedPurchases": [
				{
					"orderId": "GPA.123",
					"purchaseToken": "token-123",
					"voidedReason": 0,
					"voidedSource": 0,
					"voidedQuantity": 0
				}
			]
		}`))
	}))

	result, err := publisher.ListVoidedPurchases(context.Background(), VoidedPurchaseListOptions{
		PackageName:                       "com.example.app",
		MaxResults:                        25,
		StartIndex:                        5,
		Token:                             "page",
		Type:                              VoidedPurchaseTypeProductsSubscriptions,
		IncludeQuantityBasedPartialRefund: true,
	})
	if err != nil {
		t.Fatalf("ListVoidedPurchases() error = %v", err)
	}
	if len(result.Purchases) != 1 {
		t.Fatalf("len(Purchases) = %d, want 1", len(result.Purchases))
	}
	if result.Pagination == nil || result.Pagination.NextPageToken != "next" {
		t.Fatalf("Pagination = %#v, want next token", result.Pagination)
	}
}

func TestListUsersSendsExpectedQueryParams(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/developers/1234567890/users" {
			t.Fatalf("path = %q, want users endpoint", r.URL.Path)
		}
		query := r.URL.Query()
		assertQueryValue(t, query, "pageSize", "25")
		assertQueryValue(t, query, "pageToken", "page")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"nextPageToken": "next",
			"users": [
				{
					"name": "developers/1234567890/users/user@example.com",
					"email": "user@example.com",
					"accessState": "ACCESS_GRANTED",
					"developerAccountPermissions": ["CAN_VIEW_NON_FINANCIAL_DATA_GLOBAL"],
					"partial": true,
					"grants": [
						{
							"name": "developers/1234567890/users/user@example.com/grants/com.example.app",
							"packageName": "com.example.app",
							"appLevelPermissions": ["CAN_REPLY_TO_REVIEWS"]
						}
					]
				}
			]
		}`))
	}))

	result, err := publisher.ListUsers(context.Background(), UserListOptions{
		Developer: "1234567890",
		PageSize:  25,
		PageToken: "page",
	})
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	if result.NextPageToken != "next" {
		t.Fatalf("NextPageToken = %q, want next", result.NextPageToken)
	}
	if len(result.Users) != 1 {
		t.Fatalf("len(Users) = %d, want 1", len(result.Users))
	}
	user := result.Users[0]
	if user.Email != "user@example.com" || !user.Partial {
		t.Fatalf("user = %#v, want partial user@example.com", user)
	}
	if len(user.Grants) != 1 || user.Grants[0].PackageName != "com.example.app" {
		t.Fatalf("grants = %#v, want com.example.app grant", user.Grants)
	}
}

func TestListUsersValidatesOptionsBeforeRequest(t *testing.T) {
	requests := 0
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		t.Fatalf("unexpected request to %s", r.URL.Path)
	}))

	_, err := publisher.ListUsers(context.Background(), UserListOptions{Developer: "1234567890", PageSize: -2})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if requests != 0 {
		t.Fatalf("requests = %d, want 0", requests)
	}
}

func TestCreateUserCanonicalizesDeveloperResource(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/developers/1234567890/users" {
			t.Fatalf("path = %q, want canonical users endpoint", r.URL.Path)
		}
		var request androidpublisher.User
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if request.Name != "developers/1234567890/users/user@example.com" {
			t.Fatalf("request name = %q, want canonical user resource", request.Name)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"developers/1234567890/users/user@example.com","email":"user@example.com"}`))
	}))

	_, err := publisher.CreateUser(context.Background(), UserCreateOptions{
		Developer:   "developers/1234567890",
		UserEmail:   "user@example.com",
		Permissions: []UserPermission{UserPermissionViewNonFinancialDataGlobal},
		Confirm:     true,
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
}

func TestCreateUserUsesUsersEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/developers/1234567890/users" {
			t.Fatalf("path = %q, want users endpoint", r.URL.Path)
		}
		var request androidpublisher.User
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if request.Name != "developers/1234567890/users/user@example.com" || request.Email != "user@example.com" {
			t.Fatalf("request = %#v, want user identity", request)
		}
		if len(request.DeveloperAccountPermissions) != 1 || request.DeveloperAccountPermissions[0] != "CAN_VIEW_NON_FINANCIAL_DATA_GLOBAL" {
			t.Fatalf("permissions = %#v, want non-financial global", request.DeveloperAccountPermissions)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"name": "developers/1234567890/users/user@example.com",
			"email": "user@example.com",
			"developerAccountPermissions": ["CAN_VIEW_NON_FINANCIAL_DATA_GLOBAL"]
		}`))
	}))

	user, err := publisher.CreateUser(context.Background(), UserCreateOptions{
		Developer:   "1234567890",
		UserEmail:   "user@example.com",
		Permissions: []UserPermission{UserPermissionViewNonFinancialDataGlobal},
		Confirm:     true,
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if user.Email != "user@example.com" || len(user.DeveloperAccountPermissions) != 1 {
		t.Fatalf("user = %#v, want mapped user", user)
	}
}

func TestCreateUserRejectsDryRunBeforeRequest(t *testing.T) {
	requests := 0
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		t.Fatalf("unexpected request to %s", r.URL.Path)
	}))

	_, err := publisher.CreateUser(context.Background(), UserCreateOptions{
		Developer:   "1234567890",
		UserEmail:   "user@example.com",
		Permissions: []UserPermission{UserPermissionViewNonFinancialDataGlobal},
		DryRun:      true,
	})
	if err == nil {
		t.Fatal("expected dry-run rejection")
	}
	if requests != 0 {
		t.Fatalf("requests = %d, want 0", requests)
	}
}

func TestPatchUserUsesUserEndpointAndUpdateMask(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("method = %s, want PATCH", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/developers/1234567890/users/user@example.com" {
			t.Fatalf("path = %q, want user endpoint", r.URL.Path)
		}
		assertQueryValue(t, r.URL.Query(), "updateMask", "developerAccountPermissions,expirationTime")
		var request androidpublisher.User
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if request.Name != "developers/1234567890/users/user@example.com" || request.ExpirationTime != "2027-01-02T03:04:05Z" {
			t.Fatalf("request = %#v, want name and expiration", request)
		}
		if len(request.DeveloperAccountPermissions) != 1 || request.DeveloperAccountPermissions[0] != "CAN_REPLY_TO_REVIEWS_GLOBAL" {
			t.Fatalf("permissions = %#v, want reply permission", request.DeveloperAccountPermissions)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"name": "developers/1234567890/users/user@example.com",
			"email": "user@example.com",
			"developerAccountPermissions": ["CAN_REPLY_TO_REVIEWS_GLOBAL"],
			"expirationTime": "2027-01-02T03:04:05Z"
		}`))
	}))

	user, err := publisher.PatchUser(context.Background(), UserPatchOptions{
		Name:           "developers/1234567890/users/user@example.com",
		Permissions:    []UserPermission{UserPermissionReplyToReviewsGlobal},
		ExpirationTime: "2027-01-02T03:04:05Z",
		Confirm:        true,
	})
	if err != nil {
		t.Fatalf("PatchUser() error = %v", err)
	}
	if user.ExpirationTime != "2027-01-02T03:04:05Z" || len(user.DeveloperAccountPermissions) != 1 {
		t.Fatalf("user = %#v, want expiration and permission", user)
	}
}

func TestPatchUserRejectsDryRunBeforeRequest(t *testing.T) {
	requests := 0
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		t.Fatalf("unexpected request to %s", r.URL.Path)
	}))

	_, err := publisher.PatchUser(context.Background(), UserPatchOptions{
		Name:           "developers/1234567890/users/user@example.com",
		ExpirationTime: "2027-01-02T03:04:05Z",
		DryRun:         true,
	})
	if err == nil {
		t.Fatal("expected dry-run rejection")
	}
	if requests != 0 {
		t.Fatalf("requests = %d, want 0", requests)
	}
}

func TestDeleteUserUsesUserEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/developers/1234567890/users/user@example.com" {
			t.Fatalf("path = %q, want user endpoint", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	err := publisher.DeleteUser(context.Background(), UserDeleteOptions{
		Name:    "developers/1234567890/users/user@example.com",
		Confirm: true,
	})
	if err != nil {
		t.Fatalf("DeleteUser() error = %v", err)
	}
}

func TestDeleteUserRejectsDryRunBeforeRequest(t *testing.T) {
	requests := 0
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		t.Fatalf("unexpected request to %s", r.URL.Path)
	}))

	err := publisher.DeleteUser(context.Background(), UserDeleteOptions{
		Name:   "developers/1234567890/users/user@example.com",
		DryRun: true,
	})
	if err == nil {
		t.Fatal("expected dry-run rejection")
	}
	if requests != 0 {
		t.Fatalf("requests = %d, want 0", requests)
	}
}

func TestListAppRecoveriesSendsVersionCode(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/appRecoveries" {
			t.Fatalf("path = %q, want app recoveries endpoint", r.URL.Path)
		}
		assertQueryValue(t, r.URL.Query(), "versionCode", "42")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"recoveryActions": [
				{
					"appRecoveryId": "7",
					"status": "RECOVERY_STATUS_ACTIVE",
					"createTime": "2026-05-18T10:00:00Z",
					"lastUpdateTime": "2026-05-18T11:00:00Z",
					"targeting": {
						"regions": {"regionCode": ["US", "ES"]},
						"versionRange": {"versionCodeStart": "40", "versionCodeEnd": "42"}
					},
					"remoteInAppUpdateData": {
						"remoteAppUpdateDataPerBundle": [{
							"versionCode": "42",
							"recoveredDeviceCount": "12",
							"totalDeviceCount": "30"
						}]
					}
				}
			]
		}`))
	}))

	result, err := publisher.ListAppRecoveries(context.Background(), AppRecoveryListOptions{
		PackageName: "com.example.app",
		VersionCode: 42,
	})
	if err != nil {
		t.Fatalf("ListAppRecoveries() error = %v", err)
	}
	if len(result.Actions) != 1 {
		t.Fatalf("len(Actions) = %d, want 1", len(result.Actions))
	}
	if result.Actions[0].AppRecoveryID != "7" || result.Actions[0].Status != "RECOVERY_STATUS_ACTIVE" {
		t.Fatalf("action = %#v, want active ID 7", result.Actions[0])
	}
	if result.Actions[0].Targeting == nil || result.Actions[0].Targeting.Regions == nil || len(result.Actions[0].Targeting.Regions.RegionCodes) != 2 {
		t.Fatalf("targeting = %#v, want region targeting", result.Actions[0].Targeting)
	}
	if result.Actions[0].Targeting.VersionRange == nil || result.Actions[0].Targeting.VersionRange.VersionCodeStart != "40" || result.Actions[0].Targeting.VersionRange.VersionCodeEnd != "42" {
		t.Fatalf("version range = %#v, want 40-42", result.Actions[0].Targeting.VersionRange)
	}
	if result.Actions[0].RemoteInAppUpdateData == nil || len(result.Actions[0].RemoteInAppUpdateData.PerBundle) != 1 {
		t.Fatalf("remote in-app update data = %#v, want one bundle", result.Actions[0].RemoteInAppUpdateData)
	}
	if result.Actions[0].RemoteInAppUpdateData.PerBundle[0].RecoveredDeviceCount != "12" {
		t.Fatalf("remote in-app update data = %#v, want recovered device count", result.Actions[0].RemoteInAppUpdateData)
	}
}

func TestGetOrderUsesOrderEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/orders/GPA.123" {
			t.Fatalf("path = %q, want order endpoint", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"orderId": "GPA.123",
			"purchaseToken": "token-123",
			"state": "PROCESSED",
			"total": {"currencyCode": "USD", "units": "9", "nanos": 990000000},
			"orderDetails": {"taxInclusive": true},
			"orderHistory": {
				"processedEvent": {"eventTime": "2026-05-18T10:00:00Z"},
				"partialRefundEvents": [{
					"createTime": "2026-05-18T11:00:00Z",
					"processTime": "2026-05-18T11:01:00Z",
					"state": "PROCESSED_SUCCESSFULLY",
					"refundDetails": {"total": {"currencyCode": "USD", "units": "2"}}
				}],
				"refundEvent": {
					"eventTime": "2026-05-18T12:00:00Z",
					"refundReason": "OTHER",
					"refundDetails": {"total": {"currencyCode": "USD", "units": "7"}}
				}
			},
			"pointsDetails": {
				"pointsOfferId": "points-offer",
				"pointsSpent": "100",
				"pointsDiscountRateMicros": "500000",
				"pointsCouponValue": {"currencyCode": "USD", "units": "2"}
			},
			"buyerAddress": {"buyerCountry": "US"},
			"lineItems": [{
				"productId": "premium",
				"productTitle": "Premium",
				"total": {"currencyCode": "USD", "units": "9", "nanos": 990000000},
				"subscriptionDetails": {
					"basePlanId": "monthly",
					"offerId": "intro",
					"offerPhase": "INTRODUCTORY",
					"servicePeriodStartTime": "2026-05-18T00:00:00Z",
					"servicePeriodEndTime": "2026-06-18T00:00:00Z"
				}
			}]
		}`))
	}))

	result, err := publisher.GetOrder(context.Background(), OrderGetOptions{
		PackageName: "com.example.app",
		OrderID:     "GPA.123",
	})
	if err != nil {
		t.Fatalf("GetOrder() error = %v", err)
	}
	if result.Order.OrderID != "GPA.123" || result.Order.Total.Units != 9 {
		t.Fatalf("order = %#v, want GPA.123 total 9", result.Order)
	}
	if result.Order.BuyerAddress == nil || result.Order.BuyerAddress.Country != "US" {
		t.Fatalf("buyer address = %#v, want US", result.Order.BuyerAddress)
	}
	if len(result.Order.LineItems) != 1 || result.Order.LineItems[0].ProductID != "premium" {
		t.Fatalf("line items = %#v, want premium line item", result.Order.LineItems)
	}
	if result.Order.OrderDetails == nil || !result.Order.OrderDetails.TaxInclusive {
		t.Fatalf("order details = %#v, want tax-inclusive details", result.Order.OrderDetails)
	}
	if result.Order.OrderHistory == nil || result.Order.OrderHistory.ProcessedEvent.EventTime == "" {
		t.Fatalf("order history = %#v, want processed event", result.Order.OrderHistory)
	}
	if len(result.Order.OrderHistory.PartialRefundEvents) != 1 || result.Order.OrderHistory.PartialRefundEvents[0].RefundDetails.Total.Units != 2 {
		t.Fatalf("partial refunds = %#v, want one $2 refund", result.Order.OrderHistory.PartialRefundEvents)
	}
	if result.Order.OrderHistory.RefundEvent == nil || result.Order.OrderHistory.RefundEvent.RefundReason != "OTHER" {
		t.Fatalf("refund event = %#v, want OTHER refund", result.Order.OrderHistory.RefundEvent)
	}
	if result.Order.PointsDetails == nil || result.Order.PointsDetails.PointsSpent != 100 {
		t.Fatalf("points details = %#v, want 100 points", result.Order.PointsDetails)
	}
	if result.Order.LineItems[0].SubscriptionDetails == nil || result.Order.LineItems[0].SubscriptionDetails.BasePlanID != "monthly" {
		t.Fatalf("subscription details = %#v, want monthly base plan", result.Order.LineItems[0].SubscriptionDetails)
	}
}

func TestBatchGetOrdersUsesRepeatedOrderIDs(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/orders:batchGet" {
			t.Fatalf("path = %q, want orders batch endpoint", r.URL.Path)
		}
		got := r.URL.Query()["orderIds"]
		want := []string{"GPA.123", "GPA.456"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("orderIds = %#v, want %#v", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"orders": [
				{"orderId": "GPA.123", "state": "PROCESSED"},
				{"orderId": "GPA.456", "state": "REFUNDED"}
			]
		}`))
	}))

	result, err := publisher.BatchGetOrders(context.Background(), OrderBatchGetOptions{
		PackageName: "com.example.app",
		OrderIDs:    []OrderID{"GPA.123", "GPA.456"},
	})
	if err != nil {
		t.Fatalf("BatchGetOrders() error = %v", err)
	}
	if len(result.Orders) != 2 || result.Orders[1].State != "REFUNDED" {
		t.Fatalf("orders = %#v, want two orders", result.Orders)
	}
}

func TestRefundOrderUsesRefundEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/orders/GPA.123:refund" {
			t.Fatalf("path = %q, want order refund endpoint", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		assertQueryValue(t, r.URL.Query(), "revoke", "true")
		w.WriteHeader(http.StatusOK)
	}))

	err := publisher.RefundOrder(context.Background(), OrderRefundOptions{
		PackageName: "com.example.app",
		OrderID:     "GPA.123",
		Revoke:      true,
		Confirm:     true,
	})
	if err != nil {
		t.Fatalf("RefundOrder() error = %v", err)
	}
}

func TestRefundOrderOmitsRevokeByDefault(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/orders/GPA.123:refund" {
			t.Fatalf("path = %q, want order refund endpoint", r.URL.Path)
		}
		if _, ok := r.URL.Query()["revoke"]; ok {
			t.Fatalf("query = %q, did not expect revoke", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusOK)
	}))

	err := publisher.RefundOrder(context.Background(), OrderRefundOptions{
		PackageName: "com.example.app",
		OrderID:     "GPA.123",
		Confirm:     true,
	})
	if err != nil {
		t.Fatalf("RefundOrder() error = %v", err)
	}
}

func TestConvertRegionPricesUsesPricingEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/pricing:convertRegionPrices" {
			t.Fatalf("path = %q, want pricing conversion endpoint", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		var request androidpublisher.ConvertRegionPricesRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if request.Price == nil || request.Price.CurrencyCode != "USD" || request.Price.Units != 9 || request.Price.Nanos != 990000000 {
			t.Fatalf("request price = %#v, want USD 9.99", request.Price)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"regionVersion": {"version": "2026/05"},
			"convertedOtherRegionsPrice": {
				"usdPrice": {"currencyCode": "USD", "units": "9", "nanos": 990000000},
				"eurPrice": {"currencyCode": "EUR", "units": "8", "nanos": 990000000}
			},
			"convertedRegionPrices": {
				"US": {
					"regionCode": "US",
					"price": {"currencyCode": "USD", "units": "9", "nanos": 990000000},
					"taxAmount": {"currencyCode": "USD", "units": "0", "nanos": 700000000}
				}
			}
		}`))
	}))

	result, err := publisher.ConvertRegionPrices(context.Background(), RegionPriceConversionOptions{
		PackageName: "com.example.app",
		Currency:    "USD",
		Units:       9,
		Nanos:       990000000,
	})
	if err != nil {
		t.Fatalf("ConvertRegionPrices() error = %v", err)
	}
	if result.RegionVersion != "2026/05" {
		t.Fatalf("RegionVersion = %q, want 2026/05", result.RegionVersion)
	}
	if result.SourcePrice.CurrencyCode != "USD" || result.SourcePrice.Units != 9 {
		t.Fatalf("SourcePrice = %#v, want USD 9", result.SourcePrice)
	}
	if result.ConvertedOtherRegionsPrice == nil || result.ConvertedOtherRegionsPrice.EURPrice.Units != 8 {
		t.Fatalf("ConvertedOtherRegionsPrice = %#v, want EUR price", result.ConvertedOtherRegionsPrice)
	}
	if result.ConvertedRegionPrices["US"].TaxAmount.Nanos != 700000000 {
		t.Fatalf("US converted price = %#v, want tax nanos", result.ConvertedRegionPrices["US"])
	}
}

func TestUpdateDataSafetyUsesApplicationsEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/dataSafety" {
			t.Fatalf("path = %q, want data safety endpoint", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		var request androidpublisher.SafetyLabelsUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if request.SafetyLabels != "question,answer\n" {
			t.Fatalf("SafetyLabels = %q, want CSV content", request.SafetyLabels)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))

	if err := publisher.UpdateDataSafety(context.Background(), "com.example.app", "question,answer\n"); err != nil {
		t.Fatalf("UpdateDataSafety() error = %v", err)
	}
}

func TestCreateGrantUsesGrantEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/developers/123/users/user@example.com/grants" {
			t.Fatalf("path = %q, want grants endpoint", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		var request androidpublisher.Grant
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if request.PackageName != "com.example.app" || len(request.AppLevelPermissions) != 1 || request.AppLevelPermissions[0] != "CAN_VIEW_NON_FINANCIAL_DATA" {
			t.Fatalf("request = %#v, want package and permission", request)
		}
		if request.Name != "developers/123/users/user@example.com/grants/com.example.app" {
			t.Fatalf("request name = %q, want full grant resource", request.Name)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"name": "developers/123/users/user@example.com/grants/com.example.app",
			"packageName": "com.example.app",
			"appLevelPermissions": ["CAN_VIEW_NON_FINANCIAL_DATA"]
		}`))
	}))

	grant, err := publisher.CreateGrant(context.Background(), GrantCreateOptions{
		Developer:   "123",
		UserEmail:   "user@example.com",
		PackageName: "com.example.app",
		Permissions: []GrantPermission{GrantPermissionViewNonFinancialData},
		Confirm:     true,
	})
	if err != nil {
		t.Fatalf("CreateGrant() error = %v", err)
	}
	if grant.Name != "developers/123/users/user@example.com/grants/com.example.app" || len(grant.Permissions) != 1 {
		t.Fatalf("grant = %#v, want mapped grant", grant)
	}
}

func TestPatchGrantUsesUpdateMask(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/developers/123/users/user@example.com/grants/com.example.app" {
			t.Fatalf("path = %q, want grant resource endpoint", r.URL.Path)
		}
		if r.Method != http.MethodPatch {
			t.Fatalf("method = %s, want PATCH", r.Method)
		}
		assertQueryValue(t, r.URL.Query(), "updateMask", "appLevelPermissions")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"name": "developers/123/users/user@example.com/grants/com.example.app",
			"packageName": "com.example.app",
			"appLevelPermissions": ["CAN_REPLY_TO_REVIEWS"]
		}`))
	}))

	grant, err := publisher.PatchGrant(context.Background(), GrantPatchOptions{
		Name:        "developers/123/users/user@example.com/grants/com.example.app",
		Permissions: []GrantPermission{GrantPermissionReplyToReviews},
		Confirm:     true,
	})
	if err != nil {
		t.Fatalf("PatchGrant() error = %v", err)
	}
	if len(grant.Permissions) != 1 || grant.Permissions[0] != GrantPermissionReplyToReviews {
		t.Fatalf("grant = %#v, want reply permission", grant)
	}
}

func TestDeleteGrantUsesGrantEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/developers/123/users/user@example.com/grants/com.example.app" {
			t.Fatalf("path = %q, want grant resource endpoint", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Fatalf("method = %s, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	if err := publisher.DeleteGrant(context.Background(), GrantDeleteOptions{
		Name:    "developers/123/users/user@example.com/grants/com.example.app",
		Confirm: true,
	}); err != nil {
		t.Fatalf("DeleteGrant() error = %v", err)
	}
}

func TestListDeviceTierConfigsUsesApplicationsEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/deviceTierConfigs" {
			t.Fatalf("path = %q, want device tier configs endpoint", r.URL.Path)
		}
		assertQueryValue(t, r.URL.Query(), "pageSize", "25")
		assertQueryValue(t, r.URL.Query(), "pageToken", "page")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"nextPageToken": "next",
			"deviceTierConfigs": [{
				"deviceTierConfigId": "7",
				"deviceGroups": [{
					"name": "premium",
					"deviceSelectors": [{"includedDeviceIds": [{"buildBrand": "google", "buildDevice": "panther"}]}]
				}],
				"deviceTierSet": {"deviceTiers": [{"level": 1, "deviceGroupNames": ["premium"]}]},
				"userCountrySets": [{"name": "latam", "countryCodes": ["BR", "MX"]}]
			}]
		}`))
	}))

	result, err := publisher.ListDeviceTierConfigs(context.Background(), DeviceTierConfigListOptions{
		PackageName: "com.example.app",
		PageSize:    25,
		PageToken:   "page",
	})
	if err != nil {
		t.Fatalf("ListDeviceTierConfigs() error = %v", err)
	}
	if result.NextPageToken != "next" || len(result.Configs) != 1 {
		t.Fatalf("result = %#v, want next token and one config", result)
	}
	if result.Configs[0].ID != "7" {
		t.Fatalf("ID = %q, want 7", result.Configs[0].ID)
	}
	if len(result.Configs[0].DeviceGroups) != 1 || !strings.Contains(string(result.Configs[0].DeviceGroups[0].DeviceSelectors[0]), `"includedDeviceIds"`) {
		t.Fatalf("device groups = %#v, want selector payload", result.Configs[0].DeviceGroups)
	}
	if result.Configs[0].DeviceTierSet == nil || result.Configs[0].DeviceTierSet.DeviceTiers[0].Level != 1 {
		t.Fatalf("device tier set = %#v, want level 1", result.Configs[0].DeviceTierSet)
	}
	if len(result.Configs[0].UserCountrySets) != 1 || result.Configs[0].UserCountrySets[0].Name != "latam" {
		t.Fatalf("user country sets = %#v, want latam", result.Configs[0].UserCountrySets)
	}
}

func TestGetDeviceTierConfigUsesApplicationsEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/deviceTierConfigs/7" {
			t.Fatalf("path = %q, want device tier config endpoint", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"deviceTierConfigId": "7"}`))
	}))

	result, err := publisher.GetDeviceTierConfig(context.Background(), DeviceTierConfigGetOptions{
		PackageName:        "com.example.app",
		DeviceTierConfigID: 7,
	})
	if err != nil {
		t.Fatalf("GetDeviceTierConfig() error = %v", err)
	}
	if result.Config.ID != "7" {
		t.Fatalf("ID = %q, want 7", result.Config.ID)
	}
}

func TestDeployAppRecoveryUsesDeployEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/appRecoveries/7:deploy" {
			t.Fatalf("path = %q, want deploy endpoint", r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll() error = %v", err)
		}
		if strings.TrimSpace(string(body)) != "{}" {
			t.Fatalf("body = %q, want empty JSON object", string(body))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))

	if err := publisher.DeployAppRecovery(context.Background(), AppRecoveryMutationOptions{
		PackageName:   "com.example.app",
		AppRecoveryID: "7",
		Confirm:       true,
	}); err != nil {
		t.Fatalf("DeployAppRecovery() error = %v", err)
	}
}

func TestDeployAppRecoveryRejectsDryRunBeforeRequest(t *testing.T) {
	requests := 0
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		t.Fatalf("unexpected request to %s", r.URL.Path)
	}))

	err := publisher.DeployAppRecovery(context.Background(), AppRecoveryMutationOptions{
		PackageName:   "com.example.app",
		AppRecoveryID: "7",
		DryRun:        true,
	})
	if err == nil {
		t.Fatal("expected dry-run rejection")
	}
	if !strings.Contains(err.Error(), "--dry-run") {
		t.Fatalf("error = %v, want dry-run validation", err)
	}
	if requests != 0 {
		t.Fatalf("requests = %d, want 0", requests)
	}
}

func TestCancelAppRecoveryUsesCancelEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/appRecoveries/7:cancel" {
			t.Fatalf("path = %q, want cancel endpoint", r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll() error = %v", err)
		}
		if strings.TrimSpace(string(body)) != "{}" {
			t.Fatalf("body = %q, want empty JSON object", string(body))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))

	if err := publisher.CancelAppRecovery(context.Background(), AppRecoveryMutationOptions{
		PackageName:   "com.example.app",
		AppRecoveryID: "7",
		Confirm:       true,
	}); err != nil {
		t.Fatalf("CancelAppRecovery() error = %v", err)
	}
}

func TestCancelAppRecoveryRejectsDryRunBeforeRequest(t *testing.T) {
	requests := 0
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		t.Fatalf("unexpected request to %s", r.URL.Path)
	}))

	err := publisher.CancelAppRecovery(context.Background(), AppRecoveryMutationOptions{
		PackageName:   "com.example.app",
		AppRecoveryID: "7",
		DryRun:        true,
	})
	if err == nil {
		t.Fatal("expected dry-run rejection")
	}
	if requests != 0 {
		t.Fatalf("requests = %d, want 0", requests)
	}
}

func TestListGeneratedAPKsUsesVersionCodeEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/generatedApks/42" {
			t.Fatalf("path = %q, want generated APKs endpoint", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"generatedApks": [{
				"certificateSha256Hash": "abc123",
				"generatedSplitApks": [{
					"downloadId": "split-download",
					"moduleName": "base",
					"splitId": "config.en",
					"variantId": 1
				}],
				"generatedStandaloneApks": [{
					"downloadId": "standalone-download",
					"variantId": 2
				}],
				"generatedAssetPackSlices": [{
					"downloadId": "asset-download",
					"moduleName": "assets",
					"sliceId": "slice-a",
					"version": "3"
				}],
				"generatedRecoveryModules": [{
					"downloadId": "recovery-download",
					"moduleName": "base",
					"recoveryId": "7",
					"recoveryStatus": "RECOVERY_STATUS_ACTIVE"
				}],
				"generatedUniversalApk": {"downloadId": "universal-download"},
				"targetingInfo": {
					"packageName": "com.example.app",
					"variant": [{
						"variantNumber": 0,
						"targeting": {
							"abiTargeting": {"value": [{"alias": "ARM64_V8A"}]},
							"sdkVersionTargeting": {"value": [{"min": 23}]}
						},
						"apkSet": [{
							"moduleMetadata": {"name": "base"},
							"apkDescription": [{
								"path": "split-download.apk",
								"targeting": {
									"screenDensityTargeting": {"value": [{"densityAlias": "XXHDPI"}]}
								}
							}]
						}]
					}],
					"assetSliceSet": [{
						"assetModuleMetadata": {"name": "assets", "deliveryType": "FAST_FOLLOW"},
						"apkDescription": [{"path": "asset-download.apk"}]
					}]
				}
			}]
		}`))
	}))

	result, err := publisher.ListGeneratedAPKs(context.Background(), GeneratedAPKListOptions{
		PackageName: "com.example.app",
		VersionCode: 42,
	})
	if err != nil {
		t.Fatalf("ListGeneratedAPKs() error = %v", err)
	}
	if len(result.SigningKeys) != 1 {
		t.Fatalf("len(SigningKeys) = %d, want 1", len(result.SigningKeys))
	}
	signingKey := result.SigningKeys[0]
	if signingKey.CertificateSHA256Hash != "abc123" || signingKey.TargetingPackageName != "com.example.app" {
		t.Fatalf("signing key = %#v, want hash and targeting package", signingKey)
	}
	if signingKey.TargetingInfo == nil || len(signingKey.TargetingInfo.Variants) != 1 || signingKey.TargetingInfo.Variants[0].VariantNumber != 0 {
		t.Fatalf("targeting info = %#v, want variantNumber 0", signingKey.TargetingInfo)
	}
	if len(signingKey.TargetingInfo.Variants[0].ModuleNames) != 1 || signingKey.TargetingInfo.Variants[0].ModuleNames[0] != "base" {
		t.Fatalf("targeting variants = %#v, want base module", signingKey.TargetingInfo.Variants)
	}
	if !strings.Contains(string(signingKey.TargetingInfo.Variants[0].Targeting), `"abiTargeting"`) || !strings.Contains(string(signingKey.TargetingInfo.Variants[0].Targeting), `"sdkVersionTargeting"`) {
		t.Fatalf("variant targeting = %s, want ABI and SDK targeting", string(signingKey.TargetingInfo.Variants[0].Targeting))
	}
	if len(signingKey.TargetingInfo.Variants[0].APKs) != 1 || !strings.Contains(string(signingKey.TargetingInfo.Variants[0].APKs[0].Targeting), `"screenDensityTargeting"`) {
		t.Fatalf("APK targeting = %#v, want screen density targeting", signingKey.TargetingInfo.Variants[0].APKs)
	}
	if len(signingKey.TargetingInfo.AssetSliceSets) != 1 || signingKey.TargetingInfo.AssetSliceSets[0].DeliveryType != "FAST_FOLLOW" {
		t.Fatalf("asset slice targeting = %#v, want FAST_FOLLOW", signingKey.TargetingInfo.AssetSliceSets)
	}
	if len(signingKey.SplitAPKs) != 1 || signingKey.SplitAPKs[0].DownloadID != "split-download" {
		t.Fatalf("split APKs = %#v, want split download", signingKey.SplitAPKs)
	}
	if signingKey.UniversalAPK == nil || signingKey.UniversalAPK.DownloadID != "universal-download" {
		t.Fatalf("universal APK = %#v, want universal download", signingKey.UniversalAPK)
	}
	if len(signingKey.AssetPackSlices) != 1 || signingKey.AssetPackSlices[0].Version != 3 {
		t.Fatalf("asset pack slices = %#v, want version 3", signingKey.AssetPackSlices)
	}
	if len(signingKey.RecoveryModules) != 1 || signingKey.RecoveryModules[0].RecoveryID != "7" {
		t.Fatalf("recovery modules = %#v, want recovery ID 7", signingKey.RecoveryModules)
	}
}

func TestDownloadGeneratedAPKUsesDownloadEndpoint(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "split.apk")
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/generatedApks/42/downloads/split-download:download" {
			t.Fatalf("path = %q, want generated APK download endpoint", r.URL.Path)
		}
		if r.URL.Query().Get("alt") != "media" {
			t.Fatalf("alt = %q, want media", r.URL.Query().Get("alt"))
		}
		w.Header().Set("Content-Type", "application/vnd.android.package-archive")
		_, _ = w.Write([]byte("apk-bytes"))
	}))

	result, err := publisher.DownloadGeneratedAPK(context.Background(), GeneratedAPKDownloadOptions{
		PackageName: "com.example.app",
		VersionCode: 42,
		DownloadID:  "split-download",
		OutputPath:  outputPath,
	})
	if err != nil {
		t.Fatalf("DownloadGeneratedAPK() error = %v", err)
	}
	if !result.Downloaded || result.BytesWritten != int64(len("apk-bytes")) {
		t.Fatalf("result = %#v, want downloaded bytes", result)
	}
	contents, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(contents) != "apk-bytes" {
		t.Fatalf("contents = %q, want apk-bytes", string(contents))
	}
}

func TestDownloadGeneratedAPKRejectsDryRunBeforeRequest(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "split.apk")
	requests := 0
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		t.Fatalf("unexpected request to %s", r.URL.Path)
	}))

	_, err := publisher.DownloadGeneratedAPK(context.Background(), GeneratedAPKDownloadOptions{
		PackageName: "com.example.app",
		VersionCode: 42,
		DownloadID:  "split-download",
		OutputPath:  outputPath,
		DryRun:      true,
	})
	if err == nil {
		t.Fatal("expected dry-run rejection")
	}
	if requests != 0 {
		t.Fatalf("requests = %d, want 0", requests)
	}
	if _, statErr := os.Stat(outputPath); !os.IsNotExist(statErr) {
		t.Fatalf("output stat error = %v, want not exist", statErr)
	}
}

func TestDownloadGeneratedAPKRejectsExistingFileBeforeRequest(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "split.apk")
	if err := os.WriteFile(outputPath, []byte("existing"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	requests := 0
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		t.Fatalf("unexpected request to %s", r.URL.Path)
	}))

	_, err := publisher.DownloadGeneratedAPK(context.Background(), GeneratedAPKDownloadOptions{
		PackageName: "com.example.app",
		VersionCode: 42,
		DownloadID:  "split-download",
		OutputPath:  outputPath,
	})
	if err == nil {
		t.Fatal("expected existing output error")
	}
	if requests != 0 {
		t.Fatalf("requests = %d, want 0", requests)
	}
}

func TestDownloadGeneratedAPKForceReplacesExistingFile(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "split.apk")
	if err := os.WriteFile(outputPath, []byte("existing"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.android.package-archive")
		_, _ = w.Write([]byte("replacement"))
	}))

	result, err := publisher.DownloadGeneratedAPK(context.Background(), GeneratedAPKDownloadOptions{
		PackageName: "com.example.app",
		VersionCode: 42,
		DownloadID:  "split-download",
		OutputPath:  outputPath,
		Force:       true,
	})
	if err != nil {
		t.Fatalf("DownloadGeneratedAPK() error = %v", err)
	}
	if result.BytesWritten != int64(len("replacement")) {
		t.Fatalf("BytesWritten = %d, want replacement length", result.BytesWritten)
	}
	contents, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(contents) != "replacement" {
		t.Fatalf("contents = %q, want replacement", string(contents))
	}
}

func TestDownloadGeneratedAPKRemovesTempFileOnBodyError(t *testing.T) {
	dir := t.TempDir()
	outputPath := filepath.Join(dir, "split.apk")
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.android.package-archive")
		w.Header().Set("Content-Length", "20")
		_, _ = w.Write([]byte("partial"))
	}))

	_, err := publisher.DownloadGeneratedAPK(context.Background(), GeneratedAPKDownloadOptions{
		PackageName: "com.example.app",
		VersionCode: 42,
		DownloadID:  "split-download",
		OutputPath:  outputPath,
	})
	if err == nil {
		t.Fatal("expected body read error")
	}
	if _, statErr := os.Stat(outputPath); !os.IsNotExist(statErr) {
		t.Fatalf("output stat error = %v, want not exist", statErr)
	}
	matches, globErr := filepath.Glob(filepath.Join(dir, ".split.apk.tmp-*"))
	if globErr != nil {
		t.Fatalf("Glob() error = %v", globErr)
	}
	if len(matches) != 0 {
		t.Fatalf("temporary files = %v, want none", matches)
	}
}

func TestListSystemAPKVariantsUsesVersionCodeEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/systemApks/42/variants" {
			t.Fatalf("path = %q, want system APK variants endpoint", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"variants": [{
				"variantId": 7,
				"deviceSpec": {"supportedAbis": ["arm64-v8a"]},
				"options": {"rotated": true}
			}]
		}`))
	}))

	result, err := publisher.ListSystemAPKVariants(context.Background(), SystemAPKVariantListOptions{
		PackageName: "com.example.app",
		VersionCode: 42,
	})
	if err != nil {
		t.Fatalf("ListSystemAPKVariants() error = %v", err)
	}
	if len(result.Variants) != 1 || result.Variants[0].VariantID != 7 {
		t.Fatalf("variants = %#v, want variant 7", result.Variants)
	}
	if result.Variants[0].DeviceSpec == nil || len(result.Variants[0].DeviceSpec.SupportedABIs) != 1 || result.Variants[0].DeviceSpec.SupportedABIs[0] != "arm64-v8a" {
		t.Fatalf("device spec = %#v, want ABI preserved", result.Variants[0].DeviceSpec)
	}
	if result.Variants[0].Options == nil || !result.Variants[0].Options.Rotated {
		t.Fatalf("options = %#v, want rotated option", result.Variants[0].Options)
	}
}

func TestVoidedPurchaseListResultFromAPIMapsPurchasesAndPagination(t *testing.T) {
	result := voidedPurchaseListResultFromAPI(VoidedPurchaseListOptions{
		PackageName:                       "com.example.app",
		MaxResults:                        25,
		StartIndex:                        5,
		Token:                             "page",
		StartTimeMillis:                   1700000000000,
		EndTimeMillis:                     1700001000000,
		Type:                              VoidedPurchaseTypeProductsSubscriptions,
		IncludeQuantityBasedPartialRefund: true,
	}, &androidpublisher.VoidedPurchasesListResponse{
		PageInfo:        &androidpublisher.PageInfo{ResultPerPage: 25, StartIndex: 5, TotalResults: 30},
		TokenPagination: &androidpublisher.TokenPagination{NextPageToken: "next", PreviousPageToken: "prev"},
		VoidedPurchases: []*androidpublisher.VoidedPurchase{
			{
				OrderId:            "GPA.123",
				PurchaseToken:      "token-123",
				PurchaseTimeMillis: 1700000000000,
				VoidedTimeMillis:   1700000100000,
				VoidedReason:       0,
				VoidedSource:       0,
				VoidedQuantity:     0,
			},
		},
	})

	if result.PackageName != "com.example.app" {
		t.Fatalf("PackageName = %q, want com.example.app", result.PackageName)
	}
	if len(result.Purchases) != 1 {
		t.Fatalf("len(Purchases) = %d, want 1", len(result.Purchases))
	}
	purchase := result.Purchases[0]
	if purchase.OrderID != "GPA.123" {
		t.Fatalf("OrderID = %q, want GPA.123", purchase.OrderID)
	}
	if purchase.VoidedReason != 0 {
		t.Fatalf("VoidedReason = %d, want 0", purchase.VoidedReason)
	}
	if result.PageInfo == nil || result.PageInfo.TotalResults != 30 {
		t.Fatalf("PageInfo = %#v, want total results 30", result.PageInfo)
	}
	if result.Pagination == nil || result.Pagination.NextPageToken != "next" {
		t.Fatalf("Pagination = %#v, want next token", result.Pagination)
	}
}

func TestVoidedPurchaseJSONPreservesZeroReasonSourceAndQuantity(t *testing.T) {
	payload, err := json.Marshal(VoidedPurchase{
		OrderID:        "GPA.123",
		VoidedReason:   0,
		VoidedSource:   0,
		VoidedQuantity: 0,
	})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	output := string(payload)
	for _, want := range []string{`"voidedReason":0`, `"voidedSource":0`, `"voidedQuantity":0`} {
		if !strings.Contains(output, want) {
			t.Fatalf("json = %s, want %s", output, want)
		}
	}
}

func TestListOneTimeProductsUsesMonetizationEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/oneTimeProducts" {
			t.Fatalf("path = %q, want one-time products list endpoint", r.URL.Path)
		}
		assertQueryValue(t, r.URL.Query(), "pageSize", "50")
		assertQueryValue(t, r.URL.Query(), "pageToken", "next")
		_, _ = w.Write([]byte(`{
			"nextPageToken": "later",
			"oneTimeProducts": [
				{
					"packageName": "com.example.app",
					"productId": "coins_100",
					"listings": [{"languageCode": "en-US", "title": "Coins", "description": "A pile of coins"}],
					"offerTags": [{"tag": "coins"}],
					"regionsVersion": {"version": "2026/05"},
					"purchaseOptions": [
						{
							"purchaseOptionId": "buy",
							"state": "ACTIVE",
							"buyOption": {"legacyCompatible": true, "multiQuantityEnabled": true},
							"regionalPricingAndAvailabilityConfigs": [
								{
									"regionCode": "US",
									"availability": "AVAILABLE",
									"price": {"currencyCode": "USD", "units": "4", "nanos": 990000000}
								}
							],
							"newRegionsConfig": {
								"availability": "AVAILABLE",
								"usdPrice": {"currencyCode": "USD", "units": "4"},
								"eurPrice": {"currencyCode": "EUR", "units": "4"}
							},
							"taxAndComplianceSettings": {"withdrawalRightType": "WITHDRAWAL_RIGHT_DIGITAL_CONTENT"}
						}
					],
					"restrictedPaymentCountries": {"regionCodes": ["US"]},
					"taxAndComplianceSettings": {
						"isTokenizedDigitalAsset": true,
						"productTaxCategoryCode": "P2",
						"regionalProductAgeRatingInfos": [
							{"regionCode": "US", "productAgeRatingTier": "PRODUCT_AGE_RATING_TIER_THIRTEEN_AND_ABOVE"}
						],
						"regionalTaxConfigs": [
							{"regionCode": "US", "eligibleForStreamingServiceTaxRate": true, "streamingTaxType": "STREAMING_TAX_TYPE_TELCO_VIDEO_SALES", "taxTier": "TAX_TIER_NEWS_1"}
						]
					}
				}
			]
		}`))
	}))

	result, err := publisher.ListOneTimeProducts(context.Background(), OneTimeProductListOptions{
		PackageName: "com.example.app",
		PageSize:    50,
		PageToken:   "next",
	})
	if err != nil {
		t.Fatalf("ListOneTimeProducts() error = %v", err)
	}
	if result.NextPageToken != "later" {
		t.Fatalf("NextPageToken = %q, want later", result.NextPageToken)
	}
	if len(result.Products) != 1 {
		t.Fatalf("len(Products) = %d, want 1", len(result.Products))
	}
	product := result.Products[0]
	if product.ProductID != "coins_100" {
		t.Fatalf("ProductID = %q, want coins_100", product.ProductID)
	}
	if len(product.PurchaseOptions) != 1 {
		t.Fatalf("len(PurchaseOptions) = %d, want 1", len(product.PurchaseOptions))
	}
	option := product.PurchaseOptions[0]
	if option.Type != OneTimeProductPurchaseOptionTypeBuy || !option.LegacyCompatible || !option.MultiQuantityEnabled {
		t.Fatalf("purchase option = %#v, want buy option flags", option)
	}
	if option.RegionalConfigs[0].Price == nil || option.RegionalConfigs[0].Price.Units != 4 {
		t.Fatalf("regional price = %#v, want four units", option.RegionalConfigs[0].Price)
	}
	if product.RegionsVersion == nil || product.RegionsVersion.Version != "2026/05" {
		t.Fatalf("RegionsVersion = %#v, want 2026/05", product.RegionsVersion)
	}
	if product.TaxAndComplianceSettings == nil || !product.TaxAndComplianceSettings.IsTokenizedDigitalAsset {
		t.Fatalf("TaxAndComplianceSettings = %#v, want tokenized asset", product.TaxAndComplianceSettings)
	}
	if product.TaxAndComplianceSettings.ProductTaxCategoryCode != "P2" {
		t.Fatalf("ProductTaxCategoryCode = %q, want P2", product.TaxAndComplianceSettings.ProductTaxCategoryCode)
	}
	if len(product.TaxAndComplianceSettings.RegionalAgeRatings) != 1 {
		t.Fatalf("len(RegionalAgeRatings) = %d, want 1", len(product.TaxAndComplianceSettings.RegionalAgeRatings))
	}
	if product.TaxAndComplianceSettings.RegionalAgeRatings[0].ProductAgeRatingTier != "PRODUCT_AGE_RATING_TIER_THIRTEEN_AND_ABOVE" {
		t.Fatalf("RegionalAgeRatings = %#v, want 13+ tier", product.TaxAndComplianceSettings.RegionalAgeRatings)
	}
}

func TestGetOneTimeProductUsesMonetizationEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/rental_1" {
			t.Fatalf("path = %q, want one-time product get endpoint", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{
			"packageName": "com.example.app",
			"productId": "rental_1",
			"purchaseOptions": [
				{
					"purchaseOptionId": "rent",
					"state": "DRAFT",
					"rentOption": {"rentalPeriod": "P30D", "expirationPeriod": "P2D"}
				}
			]
		}`))
	}))

	product, err := publisher.GetOneTimeProduct(context.Background(), "com.example.app", "rental_1")
	if err != nil {
		t.Fatalf("GetOneTimeProduct() error = %v", err)
	}
	if product.ProductID != "rental_1" {
		t.Fatalf("ProductID = %q, want rental_1", product.ProductID)
	}
	if len(product.PurchaseOptions) != 1 {
		t.Fatalf("len(PurchaseOptions) = %d, want 1", len(product.PurchaseOptions))
	}
	if product.PurchaseOptions[0].Type != OneTimeProductPurchaseOptionTypeRent {
		t.Fatalf("Type = %q, want rent", product.PurchaseOptions[0].Type)
	}
	if product.PurchaseOptions[0].RentalPeriod != "P30D" {
		t.Fatalf("RentalPeriod = %q, want P30D", product.PurchaseOptions[0].RentalPeriod)
	}
}

func TestUpdatePurchaseOptionStateUsesBatchUpdateStatesEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/coins_100/purchaseOptions:batchUpdateStates" {
			t.Fatalf("path = %q, want purchase option state endpoint", r.URL.Path)
		}
		var request androidpublisher.BatchUpdatePurchaseOptionStatesRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if len(request.Requests) != 1 || request.Requests[0].DeactivatePurchaseOptionRequest == nil {
			t.Fatalf("request = %#v, want deactivate request", request)
		}
		deactivate := request.Requests[0].DeactivatePurchaseOptionRequest
		if deactivate.PackageName != "com.example.app" || deactivate.ProductId != "coins_100" || deactivate.PurchaseOptionId != "buy" {
			t.Fatalf("deactivate request = %#v, want identifiers", deactivate)
		}
		if deactivate.LatencyTolerance != "PRODUCT_UPDATE_LATENCY_TOLERANCE_LATENCY_TOLERANT" {
			t.Fatalf("LatencyTolerance = %q, want tolerant", deactivate.LatencyTolerance)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"oneTimeProducts": [
				{
					"packageName": "com.example.app",
					"productId": "coins_100",
					"purchaseOptions": [{"purchaseOptionId": "buy", "state": "INACTIVE", "buyOption": {}}]
				}
			]
		}`)
	}))

	product, err := publisher.UpdatePurchaseOptionState(context.Background(), PurchaseOptionStateUpdateOptions{
		PackageName:      "com.example.app",
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		Action:           PurchaseOptionStateActionDeactivate,
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("UpdatePurchaseOptionState() error = %v", err)
	}
	if len(product.PurchaseOptions) != 1 || product.PurchaseOptions[0].State != "INACTIVE" {
		t.Fatalf("PurchaseOptions = %#v, want inactive", product.PurchaseOptions)
	}
}

func TestUpdatePurchaseOptionStateRejectsDryRunBeforeRequest(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))

	_, err := publisher.UpdatePurchaseOptionState(context.Background(), PurchaseOptionStateUpdateOptions{
		PackageName:      "com.example.app",
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		Action:           PurchaseOptionStateActionActivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected live validation error")
	}
	if !strings.Contains(err.Error(), "cannot be a dry-run") {
		t.Fatalf("error = %v, want dry-run validation", err)
	}
}

func TestUpdatePurchaseOptionStateRequiresConfirmBeforeRequest(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))

	_, err := publisher.UpdatePurchaseOptionState(context.Background(), PurchaseOptionStateUpdateOptions{
		PackageName:      "com.example.app",
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		Action:           PurchaseOptionStateActionActivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
	})
	if err == nil {
		t.Fatal("expected live confirmation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm") {
		t.Fatalf("error = %v, want confirm validation", err)
	}
}

func TestUpdatePurchaseOptionStateUsesActivateRequest(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request androidpublisher.BatchUpdatePurchaseOptionStatesRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if len(request.Requests) != 1 || request.Requests[0].ActivatePurchaseOptionRequest == nil {
			t.Fatalf("request = %#v, want activate request", request)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"oneTimeProducts":[{"packageName":"com.example.app","productId":"coins_100","purchaseOptions":[{"purchaseOptionId":"buy","state":"ACTIVE","buyOption":{}}]}]}`)
	}))

	product, err := publisher.UpdatePurchaseOptionState(context.Background(), PurchaseOptionStateUpdateOptions{
		PackageName:      "com.example.app",
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		Action:           PurchaseOptionStateActionActivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("UpdatePurchaseOptionState() error = %v", err)
	}
	if len(product.PurchaseOptions) != 1 || product.PurchaseOptions[0].State != "ACTIVE" {
		t.Fatalf("PurchaseOptions = %#v, want active", product.PurchaseOptions)
	}
}

func TestListOneTimeProductOffersUsesMonetizationEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/coins_100/purchaseOptions/buy/offers" {
			t.Fatalf("path = %q, want one-time product offers list endpoint", r.URL.Path)
		}
		assertQueryValue(t, r.URL.Query(), "pageSize", "50")
		assertQueryValue(t, r.URL.Query(), "pageToken", "next")
		_, _ = w.Write([]byte(`{
			"nextPageToken": "later",
			"oneTimeProductOffers": [
				{
					"packageName": "com.example.app",
					"productId": "coins_100",
					"purchaseOptionId": "buy",
					"offerId": "intro",
					"state": "ACTIVE",
					"offerTags": [{"tag": "welcome"}],
					"regionsVersion": {"version": "2026/05"},
					"discountedOffer": {
						"startTime": "2026-05-01T00:00:00Z",
						"endTime": "2026-06-01T00:00:00Z",
						"redemptionLimit": "10"
					},
					"regionalPricingAndAvailabilityConfigs": [
						{
							"regionCode": "US",
							"availability": "AVAILABLE",
							"absoluteDiscount": {"currencyCode": "USD", "units": "1"},
							"relativeDiscount": 0.5
						},
						{
							"regionCode": "CA",
							"availability": "AVAILABLE",
							"noOverride": {}
						}
					]
				}
			]
		}`))
	}))

	result, err := publisher.ListOneTimeProductOffers(context.Background(), OneTimeProductOfferListOptions{
		PackageName:      "com.example.app",
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		PageSize:         50,
		PageToken:        "next",
	})
	if err != nil {
		t.Fatalf("ListOneTimeProductOffers() error = %v", err)
	}
	if result.NextPageToken != "later" {
		t.Fatalf("NextPageToken = %q, want later", result.NextPageToken)
	}
	if len(result.Offers) != 1 {
		t.Fatalf("len(Offers) = %d, want 1", len(result.Offers))
	}
	offer := result.Offers[0]
	if offer.OfferID != "intro" {
		t.Fatalf("OfferID = %q, want intro", offer.OfferID)
	}
	if offer.Type != OneTimeProductOfferTypeDiscounted {
		t.Fatalf("Type = %q, want discounted", offer.Type)
	}
	if offer.DiscountedOffer == nil || offer.DiscountedOffer.RedemptionLimit != 10 {
		t.Fatalf("DiscountedOffer = %#v, want redemption limit 10", offer.DiscountedOffer)
	}
	if offer.RegionsVersion == nil || offer.RegionsVersion.Version != "2026/05" {
		t.Fatalf("RegionsVersion = %#v, want 2026/05", offer.RegionsVersion)
	}
	if offer.RegionalConfigs[0].AbsoluteDiscount == nil || offer.RegionalConfigs[0].AbsoluteDiscount.Units != 1 {
		t.Fatalf("AbsoluteDiscount = %#v, want one unit", offer.RegionalConfigs[0].AbsoluteDiscount)
	}
	if !offer.RegionalConfigs[1].NoOverride {
		t.Fatalf("second regional config = %#v, want no override", offer.RegionalConfigs[1])
	}
}

func TestGetOneTimeProductOfferUsesBatchGetEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/coins_100/purchaseOptions/buy/offers:batchGet" {
			t.Fatalf("path = %q, want one-time product offers batchGet endpoint", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		var body struct {
			Requests []struct {
				PackageName      string `json:"packageName"`
				ProductID        string `json:"productId"`
				PurchaseOptionID string `json:"purchaseOptionId"`
				OfferID          string `json:"offerId"`
			} `json:"requests"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if len(body.Requests) != 1 || body.Requests[0].OfferID != "preorder" {
			t.Fatalf("body = %#v, want one preorder request", body)
		}
		_, _ = w.Write([]byte(`{
			"oneTimeProductOffers": [
				{
					"packageName": "com.example.app",
					"productId": "coins_100",
					"purchaseOptionId": "buy",
					"offerId": "preorder",
					"state": "DRAFT",
					"preOrderOffer": {
						"startTime": "2026-05-01T00:00:00Z",
						"endTime": "2026-06-01T00:00:00Z",
						"releaseTime": "2026-06-15T00:00:00Z",
						"priceChangeBehavior": "PRE_ORDER_PRICE_CHANGE_BEHAVIOR_NEW_ORDERS_ONLY"
					}
				}
			]
		}`))
	}))

	offer, err := publisher.GetOneTimeProductOffer(context.Background(), OneTimeProductOfferGetOptions{
		PackageName:      "com.example.app",
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		OfferID:          "preorder",
	})
	if err != nil {
		t.Fatalf("GetOneTimeProductOffer() error = %v", err)
	}
	if offer.Type != OneTimeProductOfferTypePreOrder {
		t.Fatalf("Type = %q, want preOrder", offer.Type)
	}
	if offer.PreOrderOffer == nil || offer.PreOrderOffer.ReleaseTime != "2026-06-15T00:00:00Z" {
		t.Fatalf("PreOrderOffer = %#v, want release time", offer.PreOrderOffer)
	}
}

func TestUpdateOneTimeProductOfferStateUsesDeactivateEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/coins_100/purchaseOptions/buy/offers/intro:deactivate" {
			t.Fatalf("path = %q, want one-time product offer deactivate endpoint", r.URL.Path)
		}
		var request androidpublisher.DeactivateOneTimeProductOfferRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if request.PackageName != "com.example.app" || request.ProductId != "coins_100" || request.PurchaseOptionId != "buy" || request.OfferId != "intro" {
			t.Fatalf("request = %#v, want identifiers", request)
		}
		if request.LatencyTolerance != "PRODUCT_UPDATE_LATENCY_TOLERANCE_LATENCY_TOLERANT" {
			t.Fatalf("LatencyTolerance = %q, want tolerant", request.LatencyTolerance)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"packageName":"com.example.app","productId":"coins_100","purchaseOptionId":"buy","offerId":"intro","state":"INACTIVE"}`)
	}))

	offer, err := publisher.UpdateOneTimeProductOfferState(context.Background(), OneTimeProductOfferStateUpdateOptions{
		PackageName:      "com.example.app",
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		OfferID:          "intro",
		Action:           OneTimeProductOfferStateActionDeactivate,
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("UpdateOneTimeProductOfferState() error = %v", err)
	}
	if offer.State != "INACTIVE" {
		t.Fatalf("State = %q, want inactive", offer.State)
	}
}

func TestUpdateOneTimeProductOfferStateRejectsDryRunBeforeRequest(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))

	_, err := publisher.UpdateOneTimeProductOfferState(context.Background(), OneTimeProductOfferStateUpdateOptions{
		PackageName:      "com.example.app",
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		OfferID:          "intro",
		Action:           OneTimeProductOfferStateActionActivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected live validation error")
	}
	if !strings.Contains(err.Error(), "cannot be a dry-run") {
		t.Fatalf("error = %v, want dry-run validation", err)
	}
}

func TestUpdateOneTimeProductOfferStateRequiresConfirmBeforeRequest(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))

	_, err := publisher.UpdateOneTimeProductOfferState(context.Background(), OneTimeProductOfferStateUpdateOptions{
		PackageName:      "com.example.app",
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		OfferID:          "intro",
		Action:           OneTimeProductOfferStateActionActivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
	})
	if err == nil {
		t.Fatal("expected live confirmation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm") {
		t.Fatalf("error = %v, want confirm validation", err)
	}
}

func TestUpdateOneTimeProductOfferStateUsesActivateEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/coins_100/purchaseOptions/buy/offers/intro:activate" {
			t.Fatalf("path = %q, want one-time product offer activate endpoint", r.URL.Path)
		}
		var request androidpublisher.ActivateOneTimeProductOfferRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if request.OfferId != "intro" {
			t.Fatalf("request = %#v, want intro offer", request)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"packageName":"com.example.app","productId":"coins_100","purchaseOptionId":"buy","offerId":"intro","state":"ACTIVE"}`)
	}))

	offer, err := publisher.UpdateOneTimeProductOfferState(context.Background(), OneTimeProductOfferStateUpdateOptions{
		PackageName:      "com.example.app",
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		OfferID:          "intro",
		Action:           OneTimeProductOfferStateActionActivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("UpdateOneTimeProductOfferState() error = %v", err)
	}
	if offer.State != "ACTIVE" {
		t.Fatalf("State = %q, want active", offer.State)
	}
}

func TestUpdateOneTimeProductOfferStateUsesCancelEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/coins_100/purchaseOptions/buy/offers/preorder:cancel" {
			t.Fatalf("path = %q, want one-time product offer cancel endpoint", r.URL.Path)
		}
		var request androidpublisher.CancelOneTimeProductOfferRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if request.OfferId != "preorder" {
			t.Fatalf("request = %#v, want preorder offer", request)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"packageName":"com.example.app","productId":"coins_100","purchaseOptionId":"buy","offerId":"preorder","state":"CANCELLED"}`)
	}))

	offer, err := publisher.UpdateOneTimeProductOfferState(context.Background(), OneTimeProductOfferStateUpdateOptions{
		PackageName:      "com.example.app",
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		OfferID:          "preorder",
		Action:           OneTimeProductOfferStateActionCancel,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("UpdateOneTimeProductOfferState() error = %v", err)
	}
	if offer.State != "CANCELLED" {
		t.Fatalf("State = %q, want cancelled", offer.State)
	}
}

func TestBatchGetSubscriptionOffersUsesBatchGetEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions/-/basePlans/-/offers:batchGet" {
			t.Fatalf("path = %q, want subscription offers batchGet endpoint", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		var request androidpublisher.BatchGetSubscriptionOffersRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if len(request.Requests) != 2 {
			t.Fatalf("len(Requests) = %d, want 2", len(request.Requests))
		}
		if request.Requests[1].ProductId != "premium" || request.Requests[1].BasePlanId != "annual" || request.Requests[1].OfferId != "winback" {
			t.Fatalf("Requests[1] = %#v", request.Requests[1])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"subscriptionOffers":[{"packageName":"com.example.app","productId":"premium","basePlanId":"monthly","offerId":"intro"}]}`)
	}))

	result, err := publisher.BatchGetSubscriptionOffers(context.Background(), SubscriptionOfferBatchGetOptions{
		PackageName: "com.example.app",
		ProductID:   "-",
		BasePlanID:  "-",
		Requests: []SubscriptionOfferBatchGetRequest{
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro"},
			{ProductID: "premium", BasePlanID: "annual", OfferID: "winback"},
		},
	})
	if err != nil {
		t.Fatalf("BatchGetSubscriptionOffers() error = %v", err)
	}
	if len(result.Offers) != 1 || result.Offers[0].OfferID != "intro" {
		t.Fatalf("Offers = %#v, want intro offer", result.Offers)
	}
}

func TestBatchGetSubscriptionOffersOrdersOutputByRequestOrder(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"subscriptionOffers":[
			{"packageName":"com.example.app","productId":"premium","basePlanId":"annual","offerId":"winback"},
			{"packageName":"com.example.app","productId":"premium","basePlanId":"monthly","offerId":"intro"}
		]}`)
	}))

	result, err := publisher.BatchGetSubscriptionOffers(context.Background(), SubscriptionOfferBatchGetOptions{
		PackageName: "com.example.app",
		ProductID:   "premium",
		BasePlanID:  "-",
		Requests: []SubscriptionOfferBatchGetRequest{
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro"},
			{ProductID: "premium", BasePlanID: "annual", OfferID: "winback"},
		},
	})
	if err != nil {
		t.Fatalf("BatchGetSubscriptionOffers() error = %v", err)
	}
	if len(result.Offers) != 2 || result.Offers[0].OfferID != "intro" || result.Offers[1].OfferID != "winback" {
		t.Fatalf("Offers = %#v, want request order", result.Offers)
	}
}

func TestUpdateSubscriptionOfferStateUsesDeactivateEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions/premium/basePlans/monthly/offers/intro:deactivate" {
			t.Fatalf("path = %q, want subscription offer deactivate endpoint", r.URL.Path)
		}
		var request androidpublisher.DeactivateSubscriptionOfferRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if request.PackageName != "com.example.app" || request.ProductId != "premium" || request.BasePlanId != "monthly" || request.OfferId != "intro" {
			t.Fatalf("request = %#v, want identifiers", request)
		}
		if request.LatencyTolerance != "PRODUCT_UPDATE_LATENCY_TOLERANCE_LATENCY_TOLERANT" {
			t.Fatalf("LatencyTolerance = %q, want tolerant", request.LatencyTolerance)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"packageName":"com.example.app","productId":"premium","basePlanId":"monthly","offerId":"intro","state":"INACTIVE"}`)
	}))

	offer, err := publisher.UpdateSubscriptionOfferState(context.Background(), SubscriptionOfferStateUpdateOptions{
		PackageName:      "com.example.app",
		ProductID:        "premium",
		BasePlanID:       "monthly",
		OfferID:          "intro",
		Action:           SubscriptionOfferStateActionDeactivate,
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("UpdateSubscriptionOfferState() error = %v", err)
	}
	if offer.State != SubscriptionOfferStateInactive {
		t.Fatalf("State = %q, want inactive", offer.State)
	}
}

func TestUpdateSubscriptionOfferStateRejectsDryRunBeforeRequest(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))

	_, err := publisher.UpdateSubscriptionOfferState(context.Background(), SubscriptionOfferStateUpdateOptions{
		PackageName:      "com.example.app",
		ProductID:        "premium",
		BasePlanID:       "monthly",
		OfferID:          "intro",
		Action:           SubscriptionOfferStateActionActivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected live validation error")
	}
	if !strings.Contains(err.Error(), "cannot be a dry-run") {
		t.Fatalf("error = %v, want dry-run validation", err)
	}
}

func TestUpdateSubscriptionOfferStateRequiresConfirmBeforeRequest(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))

	_, err := publisher.UpdateSubscriptionOfferState(context.Background(), SubscriptionOfferStateUpdateOptions{
		PackageName:      "com.example.app",
		ProductID:        "premium",
		BasePlanID:       "monthly",
		OfferID:          "intro",
		Action:           SubscriptionOfferStateActionActivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
	})
	if err == nil {
		t.Fatal("expected live confirmation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm") {
		t.Fatalf("error = %v, want confirm validation", err)
	}
}

func TestUpdateSubscriptionOfferStateUsesActivateEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions/premium/basePlans/monthly/offers/intro:activate" {
			t.Fatalf("path = %q, want subscription offer activate endpoint", r.URL.Path)
		}
		var request androidpublisher.ActivateSubscriptionOfferRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if request.OfferId != "intro" {
			t.Fatalf("request = %#v, want intro offer", request)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"packageName":"com.example.app","productId":"premium","basePlanId":"monthly","offerId":"intro","state":"ACTIVE"}`)
	}))

	offer, err := publisher.UpdateSubscriptionOfferState(context.Background(), SubscriptionOfferStateUpdateOptions{
		PackageName:      "com.example.app",
		ProductID:        "premium",
		BasePlanID:       "monthly",
		OfferID:          "intro",
		Action:           SubscriptionOfferStateActionActivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("UpdateSubscriptionOfferState() error = %v", err)
	}
	if offer.State != SubscriptionOfferStateActive {
		t.Fatalf("State = %q, want active", offer.State)
	}
}

func TestBatchGetSubscriptionsUsesRepeatedProductIDs(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions:batchGet" {
			t.Fatalf("path = %q, want subscriptions batchGet endpoint", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		values := r.URL.Query()["productIds"]
		if !reflect.DeepEqual(values, []string{"premium_monthly", "premium_yearly"}) {
			t.Fatalf("productIds = %#v", values)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"subscriptions":[{"packageName":"com.example.app","productId":"premium_monthly"}]}`)
	}))

	result, err := publisher.BatchGetSubscriptions(context.Background(), SubscriptionBatchGetOptions{
		PackageName: "com.example.app",
		ProductIDs:  []SubscriptionProductID{"premium_monthly", "premium_yearly"},
	})
	if err != nil {
		t.Fatalf("BatchGetSubscriptions() error = %v", err)
	}
	if len(result.Subscriptions) != 1 || result.Subscriptions[0].ProductID != "premium_monthly" {
		t.Fatalf("Subscriptions = %#v, want premium_monthly", result.Subscriptions)
	}
}

func TestUpdateBasePlanStateUsesDeactivateEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions/premium/basePlans/monthly:deactivate" {
			t.Fatalf("path = %q, want base plan deactivate endpoint", r.URL.Path)
		}
		var request androidpublisher.DeactivateBasePlanRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if request.PackageName != "com.example.app" || request.ProductId != "premium" || request.BasePlanId != "monthly" {
			t.Fatalf("request = %#v, want identifiers", request)
		}
		if request.LatencyTolerance != "PRODUCT_UPDATE_LATENCY_TOLERANCE_LATENCY_TOLERANT" {
			t.Fatalf("LatencyTolerance = %q, want tolerant", request.LatencyTolerance)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"packageName": "com.example.app",
			"productId": "premium",
			"basePlans": [{"basePlanId": "monthly", "state": "INACTIVE"}]
		}`)
	}))

	subscription, err := publisher.UpdateBasePlanState(context.Background(), BasePlanStateUpdateOptions{
		PackageName:      "com.example.app",
		ProductID:        "premium",
		BasePlanID:       "monthly",
		Action:           BasePlanStateActionDeactivate,
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("UpdateBasePlanState() error = %v", err)
	}
	if len(subscription.BasePlans) != 1 || subscription.BasePlans[0].State != SubscriptionStateInactive {
		t.Fatalf("BasePlans = %#v, want inactive", subscription.BasePlans)
	}
}

func TestUpdateBasePlanStateRejectsDryRunBeforeRequest(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))

	_, err := publisher.UpdateBasePlanState(context.Background(), BasePlanStateUpdateOptions{
		PackageName:      "com.example.app",
		ProductID:        "premium",
		BasePlanID:       "monthly",
		Action:           BasePlanStateActionActivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected live validation error")
	}
	if !strings.Contains(err.Error(), "cannot be a dry-run") {
		t.Fatalf("error = %v, want dry-run validation", err)
	}
}

func TestUpdateBasePlanStateRequiresConfirmBeforeRequest(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))

	_, err := publisher.UpdateBasePlanState(context.Background(), BasePlanStateUpdateOptions{
		PackageName:      "com.example.app",
		ProductID:        "premium",
		BasePlanID:       "monthly",
		Action:           BasePlanStateActionActivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
	})
	if err == nil {
		t.Fatal("expected live confirmation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm") {
		t.Fatalf("error = %v, want confirm validation", err)
	}
}

func TestUpdateBasePlanStateUsesActivateEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions/premium/basePlans/monthly:activate" {
			t.Fatalf("path = %q, want base plan activate endpoint", r.URL.Path)
		}
		var request androidpublisher.ActivateBasePlanRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if request.BasePlanId != "monthly" {
			t.Fatalf("request = %#v, want monthly base plan", request)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"packageName":"com.example.app","productId":"premium","basePlans":[{"basePlanId":"monthly","state":"ACTIVE"}]}`)
	}))

	subscription, err := publisher.UpdateBasePlanState(context.Background(), BasePlanStateUpdateOptions{
		PackageName:      "com.example.app",
		ProductID:        "premium",
		BasePlanID:       "monthly",
		Action:           BasePlanStateActionActivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("UpdateBasePlanState() error = %v", err)
	}
	if len(subscription.BasePlans) != 1 || subscription.BasePlans[0].State != SubscriptionStateActive {
		t.Fatalf("BasePlans = %#v, want active", subscription.BasePlans)
	}
}

func newTestPublisher(t *testing.T, handler http.Handler) GooglePublisher {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	service, err := androidpublisher.NewService(
		context.Background(),
		option.WithHTTPClient(server.Client()),
		option.WithEndpoint(server.URL+"/"),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	return GooglePublisher{service: service, httpClient: server.Client(), basePath: server.URL + "/"}
}

func assertQueryValue(t *testing.T, query map[string][]string, key string, want string) {
	t.Helper()
	values := query[key]
	if len(values) != 1 || values[0] != want {
		t.Fatalf("query[%s] = %#v, want %q", key, values, want)
	}
}

func TestSubscriptionPurchaseFromAPIMapsLineItems(t *testing.T) {
	purchase := subscriptionPurchaseFromAPI("com.example.app", "token-123", &androidpublisher.SubscriptionPurchaseV2{
		SubscriptionState:    "SUBSCRIPTION_STATE_ACTIVE",
		AcknowledgementState: "ACKNOWLEDGEMENT_STATE_ACKNOWLEDGED",
		LatestOrderId:        "GPA.123",
		LinkedPurchaseToken:  "old-token",
		RegionCode:           "US",
		StartTime:            "2026-05-18T10:00:00Z",
		ExternalAccountIdentifiers: &androidpublisher.ExternalAccountIdentifiers{
			ExternalAccountId:           "external",
			ObfuscatedExternalAccountId: "account",
		},
		TestPurchase: &androidpublisher.TestPurchase{},
		LineItems: []*androidpublisher.SubscriptionPurchaseLineItem{
			{
				ProductId:               "premium",
				ExpiryTime:              "2026-06-18T10:00:00Z",
				LatestSuccessfulOrderId: "GPA.456",
				OfferDetails: &androidpublisher.OfferDetails{
					BasePlanId: "monthly",
					OfferId:    "intro",
					OfferTags:  []string{"public"},
				},
				AutoRenewingPlan: &androidpublisher.AutoRenewingPlan{
					AutoRenewEnabled: true,
					RecurringPrice:   &androidpublisher.Money{CurrencyCode: "USD", Units: 9, Nanos: 990000000},
				},
			},
		},
	})

	if purchase.Token != "token-123" {
		t.Fatalf("Token = %q, want token-123", purchase.Token)
	}
	if purchase.SubscriptionState != "SUBSCRIPTION_STATE_ACTIVE" {
		t.Fatalf("SubscriptionState = %q, want active", purchase.SubscriptionState)
	}
	if !purchase.TestPurchase {
		t.Fatal("TestPurchase = false, want true")
	}
	if len(purchase.LineItems) != 1 {
		t.Fatalf("len(LineItems) = %d, want 1", len(purchase.LineItems))
	}
	if purchase.LineItems[0].BasePlanID != "monthly" {
		t.Fatalf("BasePlanID = %q, want monthly", purchase.LineItems[0].BasePlanID)
	}
	if purchase.LineItems[0].AutoRenewEnabled == nil || !*purchase.LineItems[0].AutoRenewEnabled {
		t.Fatalf("AutoRenewEnabled = %v, want true", purchase.LineItems[0].AutoRenewEnabled)
	}
	if purchase.LineItems[0].RecurringPrice == nil || purchase.LineItems[0].RecurringPrice.Units != 9 {
		t.Fatalf("RecurringPrice = %#v, want 9 units", purchase.LineItems[0].RecurringPrice)
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
