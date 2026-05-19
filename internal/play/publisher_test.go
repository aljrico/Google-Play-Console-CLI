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
	if product.ManagedProductTaxAndComplianceSettings.IsTokenizedDigitalAsset == nil || !*product.ManagedProductTaxAndComplianceSettings.IsTokenizedDigitalAsset {
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

func TestBatchGetInAppProductsUsesRepeatedSKUs(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/inappproducts:batchGet" {
			t.Fatalf("path = %q, want in-app products batchGet endpoint", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if got := r.URL.Query()["sku"]; !reflect.DeepEqual(got, []string{"coins_100", "coins_500"}) {
			t.Fatalf("sku query = %#v, want requested SKUs", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"inappproduct":[
			{"packageName":"com.example.app","sku":"coins_100","status":"active"},
			{"packageName":"com.example.app","sku":"coins_500","status":"inactive"}
		]}`)
	}))

	result, err := publisher.BatchGetInAppProducts(context.Background(), InAppProductBatchGetOptions{
		PackageName: "com.example.app",
		SKUs:        []InAppProductSKU{"coins_100", "coins_500"},
	})
	if err != nil {
		t.Fatalf("BatchGetInAppProducts() error = %v", err)
	}
	if len(result.Products) != 2 || result.Products[0].SKU != "coins_100" || result.Products[1].SKU != "coins_500" {
		t.Fatalf("Products = %#v, want response order", result.Products)
	}
}

func TestDeleteInAppProductUsesDeleteEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/inappproducts/coins_100" {
			t.Fatalf("path = %q, want in-app product delete endpoint", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Fatalf("method = %s, want DELETE", r.Method)
		}
		if got := r.URL.Query().Get("latencyTolerance"); got != "PRODUCT_UPDATE_LATENCY_TOLERANCE_LATENCY_TOLERANT" {
			t.Fatalf("latencyTolerance = %q, want tolerant", got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	err := publisher.DeleteInAppProduct(context.Background(), InAppProductDeleteOptions{
		PackageName:      "com.example.app",
		SKU:              "coins_100",
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("DeleteInAppProduct() error = %v", err)
	}
}

func TestDeleteInAppProductRejectsDryRunBeforeRequest(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))

	err := publisher.DeleteInAppProduct(context.Background(), InAppProductDeleteOptions{
		PackageName:      "com.example.app",
		SKU:              "coins_100",
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected live validation error")
	}
}

func TestBatchDeleteInAppProductsUsesBatchDeleteEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/inappproducts:batchDelete" {
			t.Fatalf("path = %q, want in-app products batch-delete endpoint", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		var request androidpublisher.InappproductsBatchDeleteRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if len(request.Requests) != 2 {
			t.Fatalf("len(Requests) = %d, want 2", len(request.Requests))
		}
		if request.Requests[1].Sku != "coins_500" || request.Requests[1].PackageName != "com.example.app" {
			t.Fatalf("Requests[1] = %#v, want coins_500 request", request.Requests[1])
		}
		if request.Requests[0].LatencyTolerance != "PRODUCT_UPDATE_LATENCY_TOLERANCE_LATENCY_TOLERANT" {
			t.Fatalf("LatencyTolerance = %q, want tolerant", request.Requests[0].LatencyTolerance)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	err := publisher.BatchDeleteInAppProducts(context.Background(), InAppProductBatchDeleteOptions{
		PackageName:      "com.example.app",
		SKUs:             []InAppProductSKU{"coins_100", "coins_500"},
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("BatchDeleteInAppProducts() error = %v", err)
	}
}

func TestBatchDeleteInAppProductsRejectsDryRunBeforeRequest(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))

	err := publisher.BatchDeleteInAppProducts(context.Background(), InAppProductBatchDeleteOptions{
		PackageName:      "com.example.app",
		SKUs:             []InAppProductSKU{"coins_100"},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected live validation error")
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
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll() error = %v", err)
		}
		var request androidpublisher.InAppProduct
		if err := json.Unmarshal(body, &request); err != nil {
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

func TestGooglePublisherPatchInAppProductSendsPriceAndListingPatch(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("method = %s, want PATCH", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/inappproducts/coins_100" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("autoConvertMissingPrices"); got != "true" {
			t.Fatalf("autoConvertMissingPrices = %q, want true", got)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll() error = %v", err)
		}
		var request androidpublisher.InAppProduct
		if err := json.Unmarshal(body, &request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if request.DefaultPrice == nil || request.DefaultPrice.PriceMicros != "2990000" {
			t.Fatalf("DefaultPrice = %#v, want 2990000 micros", request.DefaultPrice)
		}
		if request.DefaultLanguage != "" {
			t.Fatalf("DefaultLanguage = %q, want omitted", request.DefaultLanguage)
		}
		if request.Listings["en-US"].Description != "A better coin pack." {
			t.Fatalf("request = %#v, want default listing patch", request)
		}
		if request.Status != "" {
			t.Fatalf("Status = %q, want omitted", request.Status)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"packageName":"com.example.app","sku":"coins_100","defaultPrice":{"currency":"USD","priceMicros":"2990000"},"listings":{"en-US":{"title":"100 coins","description":"A better coin pack."}}}`)
	}))
	price := ProductPrice{Currency: "USD", PriceMicros: "2990000"}

	product, err := publisher.PatchInAppProduct(context.Background(), InAppProductPatchOptions{
		PackageName:     "com.example.app",
		SKU:             "coins_100",
		ListingLanguage: "en-US",
		DefaultPrice:    &price,
		Listing:         &InAppProductListing{Title: "100 coins", Description: "A better coin pack."},
		Confirm:         true,
	})
	if err != nil {
		t.Fatalf("PatchInAppProduct() error = %v", err)
	}
	if product.DefaultPrice == nil || product.DefaultPrice.PriceMicros != "2990000" {
		t.Fatalf("product = %#v, want patched price", product)
	}
}

func TestGooglePublisherPatchInAppProductSendsRegionalPricesPatch(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("method = %s, want PATCH", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/inappproducts/coins_100" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("autoConvertMissingPrices"); got != "true" {
			t.Fatalf("autoConvertMissingPrices = %q, want true", got)
		}
		var request androidpublisher.InAppProduct
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if request.Prices["US"].PriceMicros != "2990000" || request.Prices["BR"].Currency != "BRL" {
			t.Fatalf("Prices = %#v, want US and BR regional prices", request.Prices)
		}
		if request.DefaultPrice != nil {
			t.Fatalf("DefaultPrice = %#v, want omitted", request.DefaultPrice)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"packageName":"com.example.app","sku":"coins_100","prices":{"US":{"currency":"USD","priceMicros":"2990000"},"BR":{"currency":"BRL","priceMicros":"9990000"}}}`)
	}))
	usPrice, err := NewRegionalProductPrice("US:USD:2990000")
	if err != nil {
		t.Fatalf("NewRegionalProductPrice() error = %v", err)
	}
	brPrice, err := NewRegionalProductPrice("BR:BRL:9990000")
	if err != nil {
		t.Fatalf("NewRegionalProductPrice() error = %v", err)
	}

	product, err := publisher.PatchInAppProduct(context.Background(), InAppProductPatchOptions{
		PackageName:    "com.example.app",
		SKU:            "coins_100",
		RegionalPrices: []RegionalProductPrice{usPrice, brPrice},
		Confirm:        true,
	})
	if err != nil {
		t.Fatalf("PatchInAppProduct() error = %v", err)
	}
	if product.Prices["BR"].PriceMicros != "9990000" {
		t.Fatalf("product = %#v, want patched regional prices", product)
	}
}

func TestGooglePublisherPatchInAppProductSendsTaxCompliancePatch(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("method = %s, want PATCH", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/inappproducts/coins_100" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("autoConvertMissingPrices"); got != "" {
			t.Fatalf("autoConvertMissingPrices = %q, want omitted", got)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll() error = %v", err)
		}
		var request androidpublisher.InAppProduct
		if err := json.Unmarshal(body, &request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		settings := request.ManagedProductTaxesAndComplianceSettings
		if settings == nil {
			t.Fatal("ManagedProductTaxesAndComplianceSettings = nil, want settings")
		}
		if settings.EeaWithdrawalRightType != "WITHDRAWAL_RIGHT_SERVICE" {
			t.Fatalf("settings = %#v, want withdrawal type", settings)
		}
		if settings.IsTokenizedDigitalAsset {
			t.Fatal("IsTokenizedDigitalAsset = true, want explicit false")
		}
		if settings.TaxRateInfoByRegionCode["FR"].TaxTier != "TAX_TIER_NEWS_1" {
			t.Fatalf("TaxRateInfoByRegionCode = %#v, want FR tax tier", settings.TaxRateInfoByRegionCode)
		}
		if !settings.TaxRateInfoByRegionCode["US"].EligibleForStreamingServiceTaxRate || settings.TaxRateInfoByRegionCode["US"].StreamingTaxType != "STREAMING_TAX_TYPE_TELCO_VIDEO_SALES" {
			t.Fatalf("TaxRateInfoByRegionCode = %#v, want US streaming tax", settings.TaxRateInfoByRegionCode)
		}
		if !strings.Contains(string(body), `"isTokenizedDigitalAsset":false`) {
			t.Fatalf("body = %s, want explicit false tokenized field", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"packageName":"com.example.app","sku":"coins_100","managedProductTaxesAndComplianceSettings":{"eeaWithdrawalRightType":"WITHDRAWAL_RIGHT_SERVICE","isTokenizedDigitalAsset":false,"taxRateInfoByRegionCode":{"FR":{"taxTier":"TAX_TIER_NEWS_1"},"US":{"eligibleForStreamingServiceTaxRate":true,"streamingTaxType":"STREAMING_TAX_TYPE_TELCO_VIDEO_SALES"}}}}`)
	}))
	tokenizedDigitalAsset := false

	product, err := publisher.PatchInAppProduct(context.Background(), InAppProductPatchOptions{
		PackageName: "com.example.app",
		SKU:         "coins_100",
		TaxComplianceSettings: &ProductTaxComplianceSettings{
			EEAWithdrawalRightType:  "WITHDRAWAL_RIGHT_SERVICE",
			IsTokenizedDigitalAsset: &tokenizedDigitalAsset,
			TaxRateInfoByRegionCode: map[string]RegionalTaxRateInfo{
				"FR": {TaxTier: "TAX_TIER_NEWS_1"},
				"US": {EligibleForStreamingServiceTaxRate: true, StreamingTaxType: "STREAMING_TAX_TYPE_TELCO_VIDEO_SALES"},
			},
		},
		Confirm: true,
	})
	if err != nil {
		t.Fatalf("PatchInAppProduct() error = %v", err)
	}
	if product.ManagedProductTaxAndComplianceSettings == nil || product.ManagedProductTaxAndComplianceSettings.EEAWithdrawalRightType != "WITHDRAWAL_RIGHT_SERVICE" {
		t.Fatalf("product = %#v, want tax settings", product)
	}
	if product.ManagedProductTaxAndComplianceSettings.TaxRateInfoByRegionCode["FR"].TaxTier != "TAX_TIER_NEWS_1" {
		t.Fatalf("TaxRateInfoByRegionCode = %#v, want FR tax tier", product.ManagedProductTaxAndComplianceSettings.TaxRateInfoByRegionCode)
	}
	if product.ManagedProductTaxAndComplianceSettings.TaxRateInfoByRegionCode["US"].StreamingTaxType != "STREAMING_TAX_TYPE_TELCO_VIDEO_SALES" {
		t.Fatalf("TaxRateInfoByRegionCode = %#v, want US streaming tax", product.ManagedProductTaxAndComplianceSettings.TaxRateInfoByRegionCode)
	}
}

func TestGooglePublisherCreateInAppProductSendsManagedProduct(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/inappproducts" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("autoConvertMissingPrices"); got != "true" {
			t.Fatalf("autoConvertMissingPrices = %q, want true", got)
		}
		var request androidpublisher.InAppProduct
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if request.PurchaseType != "managedUser" || request.Status != "inactive" || request.Sku != "coins_100" {
			t.Fatalf("request = %#v, want managed product", request)
		}
		if request.DefaultPrice == nil || request.DefaultPrice.Currency != "USD" || request.DefaultPrice.PriceMicros != "1990000" {
			t.Fatalf("DefaultPrice = %#v, want USD 1990000", request.DefaultPrice)
		}
		if request.DefaultLanguage != "en-US" || request.Listings["en-US"].Title != "100 coins" {
			t.Fatalf("listing = %#v language = %q, want default listing", request.Listings, request.DefaultLanguage)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"packageName":"com.example.app","sku":"coins_100","status":"inactive","purchaseType":"managedUser","defaultLanguage":"en-US","defaultPrice":{"currency":"USD","priceMicros":"1990000"}}`)
	}))

	product, err := publisher.CreateInAppProduct(context.Background(), InAppProductCreateOptions{
		PackageName:     "com.example.app",
		SKU:             "coins_100",
		Status:          ProductStatusInactive,
		DefaultLanguage: "en-US",
		DefaultPrice:    ProductPrice{Currency: "USD", PriceMicros: "1990000"},
		Listing:         InAppProductListing{Title: "100 coins", Description: "A small coin pack."},
		Confirm:         true,
	})
	if err != nil {
		t.Fatalf("CreateInAppProduct() error = %v", err)
	}
	if product.SKU != "coins_100" || product.PurchaseType != ProductPurchaseTypeManagedUser {
		t.Fatalf("product = %#v, want managed product", product)
	}
}

func TestGooglePublisherCreateInAppProductRejectsDryRunBeforeRequest(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))

	_, err := publisher.CreateInAppProduct(context.Background(), InAppProductCreateOptions{
		PackageName:     "com.example.app",
		SKU:             "coins_100",
		Status:          ProductStatusInactive,
		DefaultLanguage: "en-US",
		DefaultPrice:    ProductPrice{Currency: "USD", PriceMicros: "1990000"},
		Listing:         InAppProductListing{Title: "100 coins", Description: "A small coin pack."},
		DryRun:          true,
	})
	if err == nil {
		t.Fatal("expected live validation error")
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
	if subscription.TaxAndComplianceSettings == nil || subscription.TaxAndComplianceSettings.IsTokenizedDigitalAsset == nil || !*subscription.TaxAndComplianceSettings.IsTokenizedDigitalAsset {
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

func TestGooglePublisherPatchSubscriptionMergesListingsBeforePatch(t *testing.T) {
	var requestCount int
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		switch requestCount {
		case 1:
			if r.Method != http.MethodGet {
				t.Fatalf("method = %s, want GET", r.Method)
			}
			if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions/premium" {
				t.Fatalf("path = %q, want subscription get endpoint", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{
				"packageName": "com.example.app",
				"productId": "premium",
				"listings": [
					{"languageCode": "en-US", "title": "Premium", "description": "Old English", "benefits": ["Old benefit"]},
					{"languageCode": "es-ES", "title": "Premium ES", "description": "Spanish", "benefits": ["Beneficio anterior"]}
				]
			}`)
		case 2:
			if r.Method != http.MethodPatch {
				t.Fatalf("method = %s, want PATCH", r.Method)
			}
			if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions/premium" {
				t.Fatalf("path = %q, want subscription patch endpoint", r.URL.Path)
			}
			if got := r.URL.Query().Get("updateMask"); got != "listings" {
				t.Fatalf("updateMask = %q, want listings", got)
			}
			if got := r.URL.Query().Get("regionsVersion.version"); got != "2022/02" {
				t.Fatalf("regionsVersion.version = %q, want 2022/02", got)
			}
			if got := r.URL.Query().Get("latencyTolerance"); got != "PRODUCT_UPDATE_LATENCY_TOLERANCE_LATENCY_TOLERANT" {
				t.Fatalf("latencyTolerance = %q, want tolerant", got)
			}
			var request androidpublisher.Subscription
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if len(request.Listings) != 2 {
				t.Fatalf("len(Listings) = %d, want merged listings", len(request.Listings))
			}
			if request.Listings[0].LanguageCode != "en-US" || request.Listings[0].Title != "Premium Plus" {
				t.Fatalf("Listings[0] = %#v, want patched English title", request.Listings[0])
			}
			if request.Listings[0].Description != "Old English" || !reflect.DeepEqual(request.Listings[0].Benefits, []string{"Old benefit"}) {
				t.Fatalf("Listings[0] = %#v, want preserved omitted fields", request.Listings[0])
			}
			if request.Listings[1].LanguageCode != "es-ES" || request.Listings[1].Description != "Spanish" {
				t.Fatalf("Listings[1] = %#v, want preserved Spanish listing", request.Listings[1])
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{
				"packageName": "com.example.app",
				"productId": "premium",
				"listings": [
					{"languageCode": "en-US", "title": "Premium Plus", "description": "Old English", "benefits": ["Old benefit"]},
					{"languageCode": "es-ES", "title": "Premium ES", "description": "Spanish", "benefits": ["Beneficio anterior"]}
				]
			}`)
		default:
			t.Fatalf("unexpected request %d: %s %s", requestCount, r.Method, r.URL.Path)
		}
	}))

	subscription, err := publisher.PatchSubscription(context.Background(), SubscriptionPatchOptions{
		PackageName: "com.example.app",
		ProductID:   "premium",
		Listing: SubscriptionListing{
			LanguageCode: "en-US",
			Title:        "Premium Plus",
		},
		RegionsVersion:   "2022/02",
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("PatchSubscription() error = %v", err)
	}
	if requestCount != 2 {
		t.Fatalf("requestCount = %d, want get plus patch", requestCount)
	}
	if len(subscription.Listings) != 2 || subscription.Listings[0].Title != "Premium Plus" || subscription.Listings[0].Description != "Old English" {
		t.Fatalf("Listings = %#v, want patched listing", subscription.Listings)
	}
}

func TestGooglePublisherBatchPatchSubscriptionListingsUsesBatchUpdateEndpoint(t *testing.T) {
	var requestCount int
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		switch requestCount {
		case 1:
			if r.Method != http.MethodGet {
				t.Fatalf("method = %s, want GET", r.Method)
			}
			if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions/premium" {
				t.Fatalf("path = %q, want premium get endpoint", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{
				"packageName": "com.example.app",
				"productId": "premium",
				"listings": [
					{"languageCode": "en-US", "title": "Premium", "description": "Old English", "benefits": ["Old benefit"]},
					{"languageCode": "es-ES", "title": "Premium ES", "description": "Spanish", "benefits": ["Beneficio anterior"]}
				]
			}`)
		case 2:
			if r.Method != http.MethodGet {
				t.Fatalf("method = %s, want GET", r.Method)
			}
			if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions/vip" {
				t.Fatalf("path = %q, want vip get endpoint", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{
				"packageName": "com.example.app",
				"productId": "vip",
				"listings": [{"languageCode": "en-US", "title": "VIP", "description": "Old VIP"}]
			}`)
		case 3:
			if r.Method != http.MethodPost {
				t.Fatalf("method = %s, want POST", r.Method)
			}
			if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions:batchUpdate" {
				t.Fatalf("path = %q, want subscriptions batchUpdate endpoint", r.URL.Path)
			}
			var request androidpublisher.BatchUpdateSubscriptionsRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if len(request.Requests) != 2 {
				t.Fatalf("len(Requests) = %d, want 2", len(request.Requests))
			}
			first := request.Requests[0]
			if first.UpdateMask != "listings" {
				t.Fatalf("UpdateMask = %q, want listings", first.UpdateMask)
			}
			if first.RegionsVersion == nil || first.RegionsVersion.Version != "2026/05" {
				t.Fatalf("RegionsVersion = %#v, want 2026/05", first.RegionsVersion)
			}
			if first.LatencyTolerance != "PRODUCT_UPDATE_LATENCY_TOLERANCE_LATENCY_TOLERANT" {
				t.Fatalf("LatencyTolerance = %q, want tolerant", first.LatencyTolerance)
			}
			if first.Subscription.ProductId != "premium" || len(first.Subscription.Listings) != 2 {
				t.Fatalf("first subscription = %#v, want merged premium listings", first.Subscription)
			}
			if first.Subscription.Listings[0].Title != "Premium Plus" || first.Subscription.Listings[0].Description != "Full access" {
				t.Fatalf("premium English listing = %#v, want patched listing", first.Subscription.Listings[0])
			}
			if !reflect.DeepEqual(first.Subscription.Listings[0].Benefits, []string{"Old benefit"}) {
				t.Fatalf("premium English listing = %#v, want preserved benefits", first.Subscription.Listings[0])
			}
			if first.Subscription.Listings[1].LanguageCode != "es-ES" || first.Subscription.Listings[1].Title != "Premium Mas" {
				t.Fatalf("premium listings = %#v, want patched Spanish listing", first.Subscription.Listings)
			}
			if !reflect.DeepEqual(first.Subscription.Listings[1].Benefits, []string{"Beneficio anterior"}) {
				t.Fatalf("premium Spanish listing = %#v, want preserved benefits", first.Subscription.Listings[1])
			}
			second := request.Requests[1]
			if second.Subscription.ProductId != "vip" || second.Subscription.Listings[0].Title != "VIP Plus" {
				t.Fatalf("second subscription = %#v, want patched VIP listing", second.Subscription)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"subscriptions":[
				{"packageName":"com.example.app","productId":"premium","listings":[{"languageCode":"en-US","title":"Premium Plus","description":"Full access"},{"languageCode":"es-ES","title":"Premium Mas","description":"Acceso completo"}]},
				{"packageName":"com.example.app","productId":"vip","listings":[{"languageCode":"en-US","title":"VIP Plus","description":"All access"}]}
			]}`)
		default:
			t.Fatalf("unexpected request %d: %s %s", requestCount, r.Method, r.URL.Path)
		}
	}))

	result, err := publisher.BatchPatchSubscriptionListings(context.Background(), SubscriptionBatchPatchListingsOptions{
		PackageName: "com.example.app",
		Requests: []SubscriptionBatchPatchListingRequest{
			{ProductID: "premium", Listing: SubscriptionListing{LanguageCode: "en-US", Title: "Premium Plus", Description: "Full access"}},
			{ProductID: "premium", Listing: SubscriptionListing{LanguageCode: "es-ES", Title: "Premium Mas", Description: "Acceso completo"}},
			{ProductID: "vip", Listing: SubscriptionListing{LanguageCode: "en-US", Title: "VIP Plus", Description: "All access"}},
		},
		RegionsVersion:   "2026/05",
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("BatchPatchSubscriptionListings() error = %v", err)
	}
	if requestCount != 3 {
		t.Fatalf("requestCount = %d, want two gets plus batch update", requestCount)
	}
	if len(result.Subscriptions) != 2 || result.Subscriptions[0].ProductID != "premium" || result.Subscriptions[1].ProductID != "vip" {
		t.Fatalf("Subscriptions = %#v, want updated subscriptions", result.Subscriptions)
	}
}

func TestGooglePublisherPatchSubscriptionRejectsDryRunBeforeRequest(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))

	_, err := publisher.PatchSubscription(context.Background(), SubscriptionPatchOptions{
		PackageName:      "com.example.app",
		ProductID:        "premium",
		Listing:          SubscriptionListing{LanguageCode: "en-US", Title: "Premium"},
		RegionsVersion:   "2022/02",
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected live validation error")
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

func TestCreateAppRecoveryUsesCreateEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/appRecoveries" {
			t.Fatalf("path = %q, want app recovery create endpoint", r.URL.Path)
		}
		var request androidpublisher.CreateDraftAppRecoveryRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if request.RemoteInAppUpdate == nil || !request.RemoteInAppUpdate.IsRemoteInAppUpdateRequested {
			t.Fatalf("RemoteInAppUpdate = %#v, want requested", request.RemoteInAppUpdate)
		}
		if request.Targeting == nil || request.Targeting.VersionList == nil || !reflect.DeepEqual([]int64(request.Targeting.VersionList.VersionCodes), []int64{42, 43}) {
			t.Fatalf("Targeting = %#v, want version list", request.Targeting)
		}
		if request.Targeting.Regions == nil || !reflect.DeepEqual(request.Targeting.Regions.RegionCode, []string{"US", "BR"}) {
			t.Fatalf("Regions = %#v, want US and BR", request.Targeting.Regions)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"appRecoveryId":"7","status":"RECOVERY_STATUS_DRAFT","targeting":{"versionList":{"versionCodes":["42","43"]},"regions":{"regionCode":["US","BR"]}}}`))
	}))

	action, err := publisher.CreateAppRecovery(context.Background(), AppRecoveryCreateOptions{
		PackageName:  "com.example.app",
		VersionCodes: []int64{42, 43},
		RegionCodes:  []string{"US", "BR"},
		Confirm:      true,
	})
	if err != nil {
		t.Fatalf("CreateAppRecovery() error = %v", err)
	}
	if action.AppRecoveryID != "7" || action.Status != "RECOVERY_STATUS_DRAFT" {
		t.Fatalf("action = %#v, want draft recovery 7", action)
	}
}

func TestCreateAppRecoveryRejectsDryRunBeforeRequest(t *testing.T) {
	requests := 0
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		t.Fatalf("unexpected request to %s", r.URL.Path)
	}))

	_, err := publisher.CreateAppRecovery(context.Background(), AppRecoveryCreateOptions{
		PackageName:  "com.example.app",
		VersionCodes: []int64{42},
		DryRun:       true,
	})
	if err == nil {
		t.Fatal("expected dry-run rejection")
	}
	if requests != 0 {
		t.Fatalf("requests = %d, want 0", requests)
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

func TestAddAppRecoveryTargetingUsesAddTargetingEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/appRecoveries/7:addTargeting" {
			t.Fatalf("path = %q, want addTargeting endpoint", r.URL.Path)
		}
		var request androidpublisher.AddTargetingRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if request.TargetingUpdate == nil || request.TargetingUpdate.Regions == nil {
			t.Fatalf("TargetingUpdate = %#v, want regions", request.TargetingUpdate)
		}
		if request.TargetingUpdate.AllUsers != nil || request.TargetingUpdate.AndroidSdks != nil {
			t.Fatalf("TargetingUpdate = %#v, want only one union criterion", request.TargetingUpdate)
		}
		if !reflect.DeepEqual(request.TargetingUpdate.Regions.RegionCode, []string{"US", "BR"}) {
			t.Fatalf("RegionCode = %#v, want US and BR", request.TargetingUpdate.Regions.RegionCode)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))

	if err := publisher.AddAppRecoveryTargeting(context.Background(), AppRecoveryTargetingUpdateOptions{
		PackageName:   "com.example.app",
		AppRecoveryID: "7",
		RegionCodes:   []string{"US", "BR"},
		Confirm:       true,
	}); err != nil {
		t.Fatalf("AddAppRecoveryTargeting() error = %v", err)
	}
}

func TestAddAppRecoveryTargetingRejectsDryRunBeforeRequest(t *testing.T) {
	requests := 0
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		t.Fatalf("unexpected request to %s", r.URL.Path)
	}))

	err := publisher.AddAppRecoveryTargeting(context.Background(), AppRecoveryTargetingUpdateOptions{
		PackageName:   "com.example.app",
		AppRecoveryID: "7",
		RegionCodes:   []string{"US"},
		DryRun:        true,
	})
	if err == nil {
		t.Fatal("expected dry-run rejection")
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

func TestBatchGetOneTimeProductsUsesRepeatedProductIDs(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/oneTimeProducts:batchGet" {
			t.Fatalf("path = %q, want one-time products batch-get endpoint", r.URL.Path)
		}
		if got := r.URL.Query()["productIds"]; !reflect.DeepEqual(got, []string{"coins_100", "coins_500"}) {
			t.Fatalf("productIds = %#v, want repeated ids", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"oneTimeProducts": [
				{"packageName":"com.example.app","productId":"coins_100","listings":[{"languageCode":"en-US","title":"100 coins"}]},
				{"packageName":"com.example.app","productId":"coins_500","listings":[{"languageCode":"en-US","title":"500 coins"}]}
			]
		}`)
	}))

	result, err := publisher.BatchGetOneTimeProducts(context.Background(), OneTimeProductBatchGetOptions{
		PackageName: "com.example.app",
		ProductIDs:  []OneTimeProductID{"coins_100", "coins_500"},
	})
	if err != nil {
		t.Fatalf("BatchGetOneTimeProducts() error = %v", err)
	}
	if len(result.Products) != 2 || result.Products[1].ProductID != "coins_500" {
		t.Fatalf("Products = %#v, want requested products", result.Products)
	}
}

func TestGooglePublisherPatchOneTimeProductMergesListing(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/coins_100" {
				t.Fatalf("path = %q, want one-time product get endpoint", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{
				"packageName":"com.example.app",
				"productId":"coins_100",
				"listings":[
					{"languageCode":"en-US","title":"Old title","description":"Old description"},
					{"languageCode":"es-ES","title":"Cien monedas","description":"Comprar monedas"}
				]
			}`)
		case http.MethodPatch:
			if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/onetimeproducts/coins_100" {
				t.Fatalf("path = %q, want one-time product patch endpoint", r.URL.Path)
			}
			assertQueryValue(t, r.URL.Query(), "updateMask", oneTimeProductPatchUpdateMask)
			assertQueryValue(t, r.URL.Query(), "regionsVersion.version", "2026/05")
			assertQueryValue(t, r.URL.Query(), "latencyTolerance", "PRODUCT_UPDATE_LATENCY_TOLERANCE_LATENCY_TOLERANT")
			var request androidpublisher.OneTimeProduct
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if len(request.Listings) != 2 {
				t.Fatalf("len(Listings) = %d, want merged listings", len(request.Listings))
			}
			if request.Listings[0].Title != "Old title" || request.Listings[0].Description != "Fresh description" {
				t.Fatalf("first listing = %#v, want preserved title and patched description", request.Listings[0])
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"packageName":"com.example.app","productId":"coins_100","listings":[{"languageCode":"en-US","title":"Old title","description":"Fresh description"}]}`)
		default:
			t.Fatalf("method = %s, want GET or PATCH", r.Method)
		}
	}))

	product, err := publisher.PatchOneTimeProduct(context.Background(), OneTimeProductPatchOptions{
		PackageName: "com.example.app",
		ProductID:   "coins_100",
		Listing: OneTimeProductListing{
			LanguageCode: "en-US",
			Description:  "Fresh description",
		},
		DescriptionSet:   true,
		RegionsVersion:   "2026/05",
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("PatchOneTimeProduct() error = %v", err)
	}
	if product.Listings[0].Description != "Fresh description" {
		t.Fatalf("Listings = %#v, want patched description", product.Listings)
	}
}

func TestGooglePublisherCreateOneTimeProductUsesAllowMissingPatch(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/coins_100" {
				t.Fatalf("path = %q, want one-time product get endpoint", r.URL.Path)
			}
			http.Error(w, `{"error":{"code":404,"message":"not found"}}`, http.StatusNotFound)
			return
		case http.MethodPatch:
		default:
			t.Fatalf("method = %s, want GET or PATCH", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/onetimeproducts/coins_100" {
			t.Fatalf("path = %q, want one-time product patch endpoint", r.URL.Path)
		}
		assertQueryValue(t, r.URL.Query(), "allowMissing", "true")
		assertQueryValue(t, r.URL.Query(), "updateMask", oneTimeProductCreateUpdateMask)
		assertQueryValue(t, r.URL.Query(), "regionsVersion.version", "2026/05")
		var request androidpublisher.OneTimeProduct
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if request.PackageName != "com.example.app" || request.ProductId != "coins_100" {
			t.Fatalf("request = %#v, want package/product IDs from flags", request)
		}
		if len(request.PurchaseOptions) != 1 || request.PurchaseOptions[0].BuyOption == nil {
			t.Fatalf("purchase options = %#v, want buy option", request.PurchaseOptions)
		}
		if request.PurchaseOptions[0].State != "" {
			t.Fatalf("state = %q, want omitted output-only state", request.PurchaseOptions[0].State)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"packageName":"com.example.app","productId":"coins_100","listings":[{"languageCode":"en-US","title":"100 coins","description":"Buy coins."}],"purchaseOptions":[{"purchaseOptionId":"buy","buyOption":{"legacyCompatible":true}}]}`)
	}))

	product, err := publisher.CreateOneTimeProduct(context.Background(), OneTimeProductCreateOptions{
		PackageName:      "com.example.app",
		ProductID:        "coins_100",
		Product:          validOneTimeProductForCreate(),
		RegionsVersion:   "2026/05",
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("CreateOneTimeProduct() error = %v", err)
	}
	if product.ProductID != "coins_100" {
		t.Fatalf("ProductID = %q, want coins_100", product.ProductID)
	}
}

func TestGooglePublisherCreateOneTimeProductRejectsExistingProduct(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want only GET", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"packageName":"com.example.app","productId":"coins_100"}`)
	}))

	_, err := publisher.CreateOneTimeProduct(context.Background(), OneTimeProductCreateOptions{
		PackageName:      "com.example.app",
		ProductID:        "coins_100",
		Product:          validOneTimeProductForCreate(),
		RegionsVersion:   "2026/05",
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err == nil {
		t.Fatal("expected existing product error")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("error = %v, want already exists", err)
	}
}

func TestGooglePublisherBatchPatchOneTimeProductListingsUsesBatchUpdateEndpoint(t *testing.T) {
	var getCount int
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getCount++
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Path {
			case "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/coins_100":
				_, _ = io.WriteString(w, `{"packageName":"com.example.app","productId":"coins_100","listings":[{"languageCode":"en-US","title":"Old title","description":"Old description"}]}`)
			case "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/coins_500":
				_, _ = io.WriteString(w, `{"packageName":"com.example.app","productId":"coins_500","listings":[{"languageCode":"pt-BR","title":"500 moedas","description":"Comprar moedas"}]}`)
			default:
				t.Fatalf("path = %q, want one-time product get endpoint", r.URL.Path)
			}
		case http.MethodPost:
			if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/oneTimeProducts:batchUpdate" {
				t.Fatalf("path = %q, want one-time product batch update endpoint", r.URL.Path)
			}
			var request androidpublisher.BatchUpdateOneTimeProductsRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if len(request.Requests) != 2 {
				t.Fatalf("len(Requests) = %d, want two product update requests", len(request.Requests))
			}
			first := request.Requests[0]
			if first.UpdateMask != oneTimeProductPatchUpdateMask || first.RegionsVersion.Version != "2026/05" {
				t.Fatalf("first request = %#v, want listing update mask and regions version", first)
			}
			if first.LatencyTolerance != "PRODUCT_UPDATE_LATENCY_TOLERANCE_LATENCY_TOLERANT" {
				t.Fatalf("LatencyTolerance = %q, want tolerant", first.LatencyTolerance)
			}
			if len(first.OneTimeProduct.Listings) != 2 {
				t.Fatalf("first listings = %#v, want merged en-US and es-ES", first.OneTimeProduct.Listings)
			}
			if first.OneTimeProduct.Listings[0].Description != "Fresh description" {
				t.Fatalf("first listing = %#v, want patched description", first.OneTimeProduct.Listings[0])
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"oneTimeProducts":[{"packageName":"com.example.app","productId":"coins_100","listings":[{"languageCode":"en-US","title":"Old title","description":"Fresh description"}]},{"packageName":"com.example.app","productId":"coins_500","listings":[{"languageCode":"pt-BR","title":"500 moedas","description":"Comprar moedas"},{"languageCode":"es-ES","title":"500 monedas","description":"Comprar monedas"}]}]}`)
		default:
			t.Fatalf("method = %s, want GET or POST", r.Method)
		}
	}))

	result, err := publisher.BatchPatchOneTimeProductListings(context.Background(), OneTimeProductBatchPatchListingsOptions{
		PackageName: "com.example.app",
		Requests: []OneTimeProductBatchPatchListingRequest{
			{
				ProductID: "coins_100",
				Listing:   OneTimeProductListing{LanguageCode: "en-US", Title: "Old title", Description: "Fresh description"},
			},
			{
				ProductID: "coins_100",
				Listing:   OneTimeProductListing{LanguageCode: "es-ES", Title: "100 monedas", Description: "Comprar monedas"},
			},
			{
				ProductID: "coins_500",
				Listing:   OneTimeProductListing{LanguageCode: "es-ES", Title: "500 monedas", Description: "Comprar monedas"},
			},
		},
		RegionsVersion:   "2026/05",
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("BatchPatchOneTimeProductListings() error = %v", err)
	}
	if getCount != 2 {
		t.Fatalf("getCount = %d, want one preflight get per product", getCount)
	}
	if len(result.Products) != 2 || result.Products[0].Listings[0].Description != "Fresh description" {
		t.Fatalf("result = %#v, want patched products", result)
	}
}

func TestGooglePublisherDeleteOneTimeProductUsesMonetizationEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/coins_100" {
			t.Fatalf("path = %q, want one-time product delete endpoint", r.URL.Path)
		}
		assertQueryValue(t, r.URL.Query(), "latencyTolerance", "PRODUCT_UPDATE_LATENCY_TOLERANCE_LATENCY_TOLERANT")
		w.WriteHeader(http.StatusNoContent)
	}))

	err := publisher.DeleteOneTimeProduct(context.Background(), OneTimeProductDeleteOptions{
		PackageName:      "com.example.app",
		ProductID:        "coins_100",
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("DeleteOneTimeProduct() error = %v", err)
	}
}

func TestGooglePublisherDeleteOneTimeProductRejectsDryRunBeforeRequest(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))

	err := publisher.DeleteOneTimeProduct(context.Background(), OneTimeProductDeleteOptions{
		PackageName:      "com.example.app",
		ProductID:        "coins_100",
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected live validation error")
	}
}

func TestGooglePublisherBatchDeleteOneTimeProductsUsesMonetizationEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/oneTimeProducts:batchDelete" {
			t.Fatalf("path = %q, want one-time products batch-delete endpoint", r.URL.Path)
		}
		var request androidpublisher.BatchDeleteOneTimeProductsRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if len(request.Requests) != 2 {
			t.Fatalf("len(Requests) = %d, want 2", len(request.Requests))
		}
		if request.Requests[0].PackageName != "com.example.app" || request.Requests[0].ProductId != "coins_100" {
			t.Fatalf("first request = %#v, want identifiers", request.Requests[0])
		}
		if request.Requests[0].LatencyTolerance != "PRODUCT_UPDATE_LATENCY_TOLERANCE_LATENCY_TOLERANT" {
			t.Fatalf("LatencyTolerance = %q, want tolerant enum", request.Requests[0].LatencyTolerance)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	err := publisher.BatchDeleteOneTimeProducts(context.Background(), OneTimeProductBatchDeleteOptions{
		PackageName:      "com.example.app",
		ProductIDs:       []OneTimeProductID{"coins_100", "coins_500"},
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("BatchDeleteOneTimeProducts() error = %v", err)
	}
}

func TestGooglePublisherBatchDeleteOneTimeProductsRejectsDryRunBeforeRequest(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))

	err := publisher.BatchDeleteOneTimeProducts(context.Background(), OneTimeProductBatchDeleteOptions{
		PackageName:      "com.example.app",
		ProductIDs:       []OneTimeProductID{"coins_100", "coins_500"},
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected live validation error")
	}
}

func TestGooglePublisherBatchDeletePurchaseOptionsUsesBatchDeleteEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/-/purchaseOptions:batchDelete" {
			t.Fatalf("path = %q, want purchase options batch-delete endpoint", r.URL.Path)
		}
		var request androidpublisher.BatchDeletePurchaseOptionsRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if len(request.Requests) != 2 {
			t.Fatalf("len(Requests) = %d, want 2", len(request.Requests))
		}
		first := request.Requests[0]
		if first.PackageName != "com.example.app" || first.ProductId != "coins_100" || first.PurchaseOptionId != "buy" {
			t.Fatalf("first request = %#v, want identifiers", first)
		}
		if !first.Force {
			t.Fatal("Force = false, want true")
		}
		if first.LatencyTolerance != "PRODUCT_UPDATE_LATENCY_TOLERANCE_LATENCY_TOLERANT" {
			t.Fatalf("LatencyTolerance = %q, want tolerant enum", first.LatencyTolerance)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	err := publisher.BatchDeletePurchaseOptions(context.Background(), PurchaseOptionBatchDeleteOptions{
		PackageName:     "com.example.app",
		ParentProductID: OneTimeProductBatchParentProductID(OneTimeProductWildcardID),
		Requests: []PurchaseOptionBatchDeleteRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy"},
			{ProductID: "coins_500", PurchaseOptionID: "rent"},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		Force:            true,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("BatchDeletePurchaseOptions() error = %v", err)
	}
}

func TestGooglePublisherBatchDeletePurchaseOptionsUsesConcreteParentEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/coins_100/purchaseOptions:batchDelete" {
			t.Fatalf("path = %q, want concrete purchase options batch-delete endpoint", r.URL.Path)
		}
		var request androidpublisher.BatchDeletePurchaseOptionsRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if len(request.Requests) != 1 || request.Requests[0].ProductId != "coins_100" {
			t.Fatalf("Requests = %#v, want concrete product request", request.Requests)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	err := publisher.BatchDeletePurchaseOptions(context.Background(), PurchaseOptionBatchDeleteOptions{
		PackageName:      "com.example.app",
		ParentProductID:  "coins_100",
		Requests:         []PurchaseOptionBatchDeleteRequest{{ProductID: "coins_100", PurchaseOptionID: "buy"}},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("BatchDeletePurchaseOptions() error = %v", err)
	}
}

func TestGooglePublisherBatchDeletePurchaseOptionsRejectsDryRunBeforeRequest(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))

	err := publisher.BatchDeletePurchaseOptions(context.Background(), PurchaseOptionBatchDeleteOptions{
		PackageName:      "com.example.app",
		ParentProductID:  "coins_100",
		Requests:         []PurchaseOptionBatchDeleteRequest{{ProductID: "coins_100", PurchaseOptionID: "buy"}},
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected live validation error")
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

func TestGooglePublisherCreateOneTimeProductOfferUsesAllowMissingBatchUpdate(t *testing.T) {
	var calls []string
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.URL.Path)
		switch r.URL.Path {
		case "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/coins_100/purchaseOptions/buy/offers:batchGet":
			if r.Method != http.MethodPost {
				t.Fatalf("method = %s, want POST batchGet", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"oneTimeProductOffers":[]}`)
		case "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/coins_100/purchaseOptions/buy/offers:batchUpdate":
			if r.Method != http.MethodPost {
				t.Fatalf("method = %s, want POST batchUpdate", r.Method)
			}
			var request androidpublisher.BatchUpdateOneTimeProductOffersRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if len(request.Requests) != 1 {
				t.Fatalf("len(Requests) = %d, want 1", len(request.Requests))
			}
			update := request.Requests[0]
			if !update.AllowMissing {
				t.Fatal("AllowMissing = false, want true")
			}
			if update.UpdateMask != oneTimeProductOfferCreateUpdateMask {
				t.Fatalf("UpdateMask = %q, want create mask", update.UpdateMask)
			}
			if update.RegionsVersion == nil || update.RegionsVersion.Version != "2026/05" {
				t.Fatalf("RegionsVersion = %#v, want 2026/05", update.RegionsVersion)
			}
			if update.OneTimeProductOffer.ProductId != "coins_100" || update.OneTimeProductOffer.PurchaseOptionId != "buy" || update.OneTimeProductOffer.OfferId != "intro" {
				t.Fatalf("offer IDs = %#v, want create IDs", update.OneTimeProductOffer)
			}
			if update.OneTimeProductOffer.State != "" {
				t.Fatalf("State = %q, want omitted output-only state", update.OneTimeProductOffer.State)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"oneTimeProductOffers":[{"packageName":"com.example.app","productId":"coins_100","purchaseOptionId":"buy","offerId":"intro","discountedOffer":{"startTime":"2026-06-01T00:00:00Z"},"regionalPricingAndAvailabilityConfigs":[{"regionCode":"US","availability":"AVAILABLE","relativeDiscount":0.5}]}]}`)
		default:
			t.Fatalf("path = %q, want offer batchGet or batchUpdate", r.URL.Path)
		}
	}))

	offer, err := publisher.CreateOneTimeProductOffer(context.Background(), OneTimeProductOfferCreateOptions{
		PackageName:      "com.example.app",
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		OfferID:          "intro",
		Offer:            validOneTimeProductOfferForCreate(),
		RegionsVersion:   "2026/05",
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("CreateOneTimeProductOffer() error = %v", err)
	}
	if offer.OfferID != "intro" {
		t.Fatalf("OfferID = %q, want intro", offer.OfferID)
	}
	if len(calls) != 2 {
		t.Fatalf("calls = %#v, want preflight then create", calls)
	}
}

func TestGooglePublisherCreateOneTimeProductOfferRejectsExistingOffer(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/coins_100/purchaseOptions/buy/offers:batchGet" {
			t.Fatalf("path = %q, want offer batchGet", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"oneTimeProductOffers":[{"packageName":"com.example.app","productId":"coins_100","purchaseOptionId":"buy","offerId":"intro"}]}`)
	}))

	_, err := publisher.CreateOneTimeProductOffer(context.Background(), OneTimeProductOfferCreateOptions{
		PackageName:      "com.example.app",
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		OfferID:          "intro",
		Offer:            validOneTimeProductOfferForCreate(),
		RegionsVersion:   "2026/05",
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err == nil {
		t.Fatal("expected existing offer error")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("error = %v, want already exists", err)
	}
}

func TestBatchPatchPurchaseOptionAvailabilityMergesAndBatchUpdates(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/coins_100":
			if r.Method != http.MethodGet {
				t.Fatalf("method = %s, want GET product", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"packageName":"com.example.app","productId":"coins_100","listings":[{"languageCode":"en-US","title":"Coins","description":"Coins"}],"purchaseOptions":[{"purchaseOptionId":"buy","buyOption":{"legacyCompatible":true},"offerTags":[{"tag":"storefront"}],"newRegionsConfig":{"availability":"AVAILABLE","usdPrice":{"currencyCode":"USD","units":"1"},"eurPrice":{"currencyCode":"EUR","units":"1"}},"taxAndComplianceSettings":{"withdrawalRightType":"WITHDRAWAL_RIGHT_SERVICE"},"regionalPricingAndAvailabilityConfigs":[{"regionCode":"US","availability":"AVAILABLE","price":{"currencyCode":"USD","units":"1","nanos":990000000}},{"regionCode":"FR","availability":"AVAILABLE","price":{"currencyCode":"EUR","units":"1","nanos":990000000}}]},{"purchaseOptionId":"rent","rentOption":{"rentalPeriod":"P7D","expirationPeriod":"P30D"},"regionalPricingAndAvailabilityConfigs":[{"regionCode":"US","availability":"AVAILABLE_IF_RELEASED","price":{"currencyCode":"USD","units":"2"}}]}]}`)
		case "/androidpublisher/v3/applications/com.example.app/oneTimeProducts:batchUpdate":
			if r.Method != http.MethodPost {
				t.Fatalf("method = %s, want POST batchUpdate", r.Method)
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("ReadAll() error = %v", err)
			}
			var request androidpublisher.BatchUpdateOneTimeProductsRequest
			if err := json.Unmarshal(body, &request); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if len(request.Requests) != 1 {
				t.Fatalf("len(Requests) = %d, want 1", len(request.Requests))
			}
			update := request.Requests[0]
			if update.UpdateMask != "purchaseOptions" {
				t.Fatalf("UpdateMask = %q, want purchaseOptions", update.UpdateMask)
			}
			if update.RegionsVersion == nil || update.RegionsVersion.Version != "2026/05" {
				t.Fatalf("RegionsVersion = %#v, want 2026/05", update.RegionsVersion)
			}
			if update.LatencyTolerance != "PRODUCT_UPDATE_LATENCY_TOLERANCE_LATENCY_TOLERANT" {
				t.Fatalf("LatencyTolerance = %q, want tolerant", update.LatencyTolerance)
			}
			options := update.OneTimeProduct.PurchaseOptions
			if len(options) != 2 {
				t.Fatalf("len(PurchaseOptions) = %d, want preserved options", len(options))
			}
			if options[0].PurchaseOptionId != "buy" || options[0].BuyOption == nil || !options[0].BuyOption.LegacyCompatible {
				t.Fatalf("first option = %#v, want preserved buy option", options[0])
			}
			if len(options[0].OfferTags) != 1 || options[0].OfferTags[0].Tag != "storefront" {
				t.Fatalf("OfferTags = %#v, want preserved tag", options[0].OfferTags)
			}
			if options[0].NewRegionsConfig == nil || options[0].NewRegionsConfig.Availability != "AVAILABLE" || options[0].NewRegionsConfig.UsdPrice == nil {
				t.Fatalf("NewRegionsConfig = %#v, want preserved new regions config", options[0].NewRegionsConfig)
			}
			if options[0].TaxAndComplianceSettings == nil || options[0].TaxAndComplianceSettings.WithdrawalRightType != "WITHDRAWAL_RIGHT_SERVICE" {
				t.Fatalf("TaxAndComplianceSettings = %#v, want preserved withdrawal right", options[0].TaxAndComplianceSettings)
			}
			configs := options[0].RegionalPricingAndAvailabilityConfigs
			if len(configs) != 2 {
				t.Fatalf("len(RegionalPricingAndAvailabilityConfigs) = %d, want preserved regions", len(configs))
			}
			if configs[0].RegionCode != "US" || configs[0].Availability != "NO_LONGER_AVAILABLE" || configs[0].Price == nil || configs[0].Price.CurrencyCode != "USD" {
				t.Fatalf("first config = %#v, want patched availability with preserved price", configs[0])
			}
			if options[1].PurchaseOptionId != "rent" || options[1].RentOption == nil {
				t.Fatalf("second option = %#v, want preserved rent option", options[1])
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"oneTimeProducts":[{"packageName":"com.example.app","productId":"coins_100","purchaseOptions":[{"purchaseOptionId":"buy","buyOption":{"legacyCompatible":true},"regionalPricingAndAvailabilityConfigs":[{"regionCode":"US","availability":"NO_LONGER_AVAILABLE","price":{"currencyCode":"USD","units":"1","nanos":990000000}}]}]}]}`)
		default:
			t.Fatalf("unexpected path = %q", r.URL.Path)
		}
	}))

	result, err := publisher.BatchPatchPurchaseOptionAvailability(context.Background(), PurchaseOptionBatchPatchAvailabilityOptions{
		PackageName:    "com.example.app",
		RegionsVersion: "2026/05",
		Requests: []PurchaseOptionAvailabilityPatchRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", RegionCode: "US", Availability: PurchaseOptionAvailabilityNoLongerAvailable},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("BatchPatchPurchaseOptionAvailability() error = %v", err)
	}
	if !result.Applied || len(result.Products) != 1 || result.Products[0].PurchaseOptions[0].RegionalConfigs[0].Availability != "NO_LONGER_AVAILABLE" {
		t.Fatalf("result = %#v, want applied product with Google availability enum", result)
	}
}

func TestBatchPatchPurchaseOptionAvailabilityRejectsMissingRegion(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/coins_100" {
			t.Fatalf("path = %q, want product get only", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"packageName":"com.example.app","productId":"coins_100","purchaseOptions":[{"purchaseOptionId":"buy","buyOption":{},"regionalPricingAndAvailabilityConfigs":[{"regionCode":"US","availability":"AVAILABLE","price":{"currencyCode":"USD","units":"1"}}]}]}`)
	}))

	_, err := publisher.BatchPatchPurchaseOptionAvailability(context.Background(), PurchaseOptionBatchPatchAvailabilityOptions{
		PackageName:    "com.example.app",
		RegionsVersion: "2026/05",
		Requests: []PurchaseOptionAvailabilityPatchRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", RegionCode: "FR", Availability: PurchaseOptionAvailabilityAvailable},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err == nil {
		t.Fatal("expected missing region validation error")
	}
	if !strings.Contains(err.Error(), "not already configured") {
		t.Fatalf("error = %v, want configured region message", err)
	}
}

func TestBatchPatchPurchaseOptionAvailabilityRejectsInvalidNoLongerAvailableTransition(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/coins_100" {
			t.Fatalf("path = %q, want product get only", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"packageName":"com.example.app","productId":"coins_100","purchaseOptions":[{"purchaseOptionId":"buy","buyOption":{},"regionalPricingAndAvailabilityConfigs":[{"regionCode":"US","availability":"AVAILABLE_IF_RELEASED","price":{"currencyCode":"USD","units":"1"}}]}]}`)
	}))

	_, err := publisher.BatchPatchPurchaseOptionAvailability(context.Background(), PurchaseOptionBatchPatchAvailabilityOptions{
		PackageName:    "com.example.app",
		RegionsVersion: "2026/05",
		Requests: []PurchaseOptionAvailabilityPatchRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", RegionCode: "US", Availability: PurchaseOptionAvailabilityNoLongerAvailable},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err == nil {
		t.Fatal("expected invalid no-longer-available transition error")
	}
	for _, want := range []string{"coins_100", "buy", "US", "AVAILABLE_IF_RELEASED"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error = %v, want %s", err, want)
		}
	}
}

func TestBatchPatchPurchaseOptionPricesMergesAndBatchUpdates(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/coins_100":
			if r.Method != http.MethodGet {
				t.Fatalf("method = %s, want GET product", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"packageName":"com.example.app","productId":"coins_100","purchaseOptions":[{"purchaseOptionId":"buy","buyOption":{"legacyCompatible":true},"offerTags":[{"tag":"storefront"}],"newRegionsConfig":{"availability":"AVAILABLE","usdPrice":{"currencyCode":"USD","units":"1"},"eurPrice":{"currencyCode":"EUR","units":"1"}},"taxAndComplianceSettings":{"withdrawalRightType":"WITHDRAWAL_RIGHT_SERVICE"},"regionalPricingAndAvailabilityConfigs":[{"regionCode":"US","availability":"AVAILABLE","price":{"currencyCode":"USD","units":"1","nanos":990000000}},{"regionCode":"FR","availability":"AVAILABLE","price":{"currencyCode":"EUR","units":"1","nanos":990000000}}]},{"purchaseOptionId":"rent","rentOption":{"rentalPeriod":"P7D","expirationPeriod":"P30D"},"regionalPricingAndAvailabilityConfigs":[{"regionCode":"US","availability":"AVAILABLE_IF_RELEASED","price":{"currencyCode":"USD","units":"2"}}]}]}`)
		case "/androidpublisher/v3/applications/com.example.app/oneTimeProducts:batchUpdate":
			if r.Method != http.MethodPost {
				t.Fatalf("method = %s, want POST batchUpdate", r.Method)
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("ReadAll() error = %v", err)
			}
			var request androidpublisher.BatchUpdateOneTimeProductsRequest
			if err := json.Unmarshal(body, &request); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if len(request.Requests) != 1 {
				t.Fatalf("len(Requests) = %d, want 1", len(request.Requests))
			}
			update := request.Requests[0]
			if update.UpdateMask != "purchaseOptions" {
				t.Fatalf("UpdateMask = %q, want purchaseOptions", update.UpdateMask)
			}
			if update.RegionsVersion == nil || update.RegionsVersion.Version != "2026/05" {
				t.Fatalf("RegionsVersion = %#v, want 2026/05", update.RegionsVersion)
			}
			options := update.OneTimeProduct.PurchaseOptions
			if len(options) != 2 {
				t.Fatalf("len(PurchaseOptions) = %d, want preserved options", len(options))
			}
			if len(options[0].OfferTags) != 1 || options[0].NewRegionsConfig == nil || options[0].TaxAndComplianceSettings == nil {
				t.Fatalf("first option = %#v, want preserved tags/new regions/tax settings", options[0])
			}
			configs := options[0].RegionalPricingAndAvailabilityConfigs
			if len(configs) != 2 {
				t.Fatalf("len(RegionalPricingAndAvailabilityConfigs) = %d, want preserved regions", len(configs))
			}
			if configs[0].RegionCode != "US" || configs[0].Availability != "AVAILABLE" || configs[0].Price == nil || configs[0].Price.CurrencyCode != "USD" || configs[0].Price.Units != 3 || configs[0].Price.Nanos != 490000000 {
				t.Fatalf("first config = %#v, want patched price with preserved availability", configs[0])
			}
			if configs[1].Price == nil || configs[1].Price.CurrencyCode != "EUR" || configs[1].Price.Units != 1 {
				t.Fatalf("second config = %#v, want preserved EUR price", configs[1])
			}
			if options[1].PurchaseOptionId != "rent" || options[1].RentOption == nil {
				t.Fatalf("second option = %#v, want preserved rent option", options[1])
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"oneTimeProducts":[{"packageName":"com.example.app","productId":"coins_100","purchaseOptions":[{"purchaseOptionId":"buy","buyOption":{"legacyCompatible":true},"regionalPricingAndAvailabilityConfigs":[{"regionCode":"US","availability":"AVAILABLE","price":{"currencyCode":"USD","units":"3","nanos":490000000}}]}]}]}`)
		default:
			t.Fatalf("unexpected path = %q", r.URL.Path)
		}
	}))

	result, err := publisher.BatchPatchPurchaseOptionPrices(context.Background(), PurchaseOptionBatchPatchPriceOptions{
		PackageName:    "com.example.app",
		RegionsVersion: "2026/05",
		Requests: []PurchaseOptionPricePatchRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", RegionCode: "US", Price: Money{CurrencyCode: "USD", Units: 3, Nanos: 490000000}},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("BatchPatchPurchaseOptionPrices() error = %v", err)
	}
	if !result.Applied || len(result.Products) != 1 || result.Products[0].PurchaseOptions[0].RegionalConfigs[0].Price.Units != 3 {
		t.Fatalf("result = %#v, want applied product with patched price", result)
	}
}

func TestBatchPatchPurchaseOptionPricesRejectsMissingRegion(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/coins_100" {
			t.Fatalf("path = %q, want product get only", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"packageName":"com.example.app","productId":"coins_100","purchaseOptions":[{"purchaseOptionId":"buy","buyOption":{},"regionalPricingAndAvailabilityConfigs":[{"regionCode":"US","availability":"AVAILABLE","price":{"currencyCode":"USD","units":"1"}}]}]}`)
	}))

	_, err := publisher.BatchPatchPurchaseOptionPrices(context.Background(), PurchaseOptionBatchPatchPriceOptions{
		PackageName:    "com.example.app",
		RegionsVersion: "2026/05",
		Requests: []PurchaseOptionPricePatchRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", RegionCode: "FR", Price: Money{CurrencyCode: "EUR", Units: 2}},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err == nil {
		t.Fatal("expected missing region validation error")
	}
	if !strings.Contains(err.Error(), "not already configured") {
		t.Fatalf("error = %v, want configured region message", err)
	}
}

func TestBatchGetOneTimeProductOffersUsesBatchGetEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/-/purchaseOptions/-/offers:batchGet" {
			t.Fatalf("path = %q, want one-time product offers batchGet endpoint", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		var request androidpublisher.BatchGetOneTimeProductOffersRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if len(request.Requests) != 2 {
			t.Fatalf("len(Requests) = %d, want 2", len(request.Requests))
		}
		if request.Requests[1].ProductId != "coins_500" || request.Requests[1].PurchaseOptionId != "buy" || request.Requests[1].OfferId != "preorder" {
			t.Fatalf("Requests[1] = %#v", request.Requests[1])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"oneTimeProductOffers":[{"packageName":"com.example.app","productId":"coins_100","purchaseOptionId":"buy","offerId":"intro"}]}`)
	}))

	result, err := publisher.BatchGetOneTimeProductOffers(context.Background(), OneTimeProductOfferBatchGetOptions{
		PackageName:      "com.example.app",
		ProductID:        "-",
		PurchaseOptionID: "-",
		Requests: []OneTimeProductOfferBatchGetRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro"},
			{ProductID: "coins_500", PurchaseOptionID: "buy", OfferID: "preorder"},
		},
	})
	if err != nil {
		t.Fatalf("BatchGetOneTimeProductOffers() error = %v", err)
	}
	if len(result.Offers) != 1 || result.Offers[0].OfferID != "intro" {
		t.Fatalf("Offers = %#v, want intro offer", result.Offers)
	}
}

func TestBatchGetOneTimeProductOffersOrdersOutputByRequestOrder(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"oneTimeProductOffers":[
			{"packageName":"com.example.app","productId":"coins_500","purchaseOptionId":"buy","offerId":"preorder"},
			{"packageName":"com.example.app","productId":"coins_100","purchaseOptionId":"buy","offerId":"intro"}
		]}`)
	}))

	result, err := publisher.BatchGetOneTimeProductOffers(context.Background(), OneTimeProductOfferBatchGetOptions{
		PackageName:      "com.example.app",
		ProductID:        "-",
		PurchaseOptionID: "-",
		Requests: []OneTimeProductOfferBatchGetRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro"},
			{ProductID: "coins_500", PurchaseOptionID: "buy", OfferID: "preorder"},
		},
	})
	if err != nil {
		t.Fatalf("BatchGetOneTimeProductOffers() error = %v", err)
	}
	if len(result.Offers) != 2 || result.Offers[0].OfferID != "intro" || result.Offers[1].OfferID != "preorder" {
		t.Fatalf("Offers = %#v, want request order", result.Offers)
	}
}

func TestBatchDeleteOneTimeProductOffersUsesBatchDeleteEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/-/purchaseOptions/-/offers:batchDelete" {
			t.Fatalf("path = %q, want one-time product offers batchDelete endpoint", r.URL.Path)
		}
		var request androidpublisher.BatchDeleteOneTimeProductOffersRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if len(request.Requests) != 2 {
			t.Fatalf("len(Requests) = %d, want 2", len(request.Requests))
		}
		if request.Requests[1].ProductId != "coins_500" || request.Requests[1].PurchaseOptionId != "rent" || request.Requests[1].OfferId != "preorder" {
			t.Fatalf("Requests[1] = %#v", request.Requests[1])
		}
		if request.Requests[0].LatencyTolerance != "PRODUCT_UPDATE_LATENCY_TOLERANCE_LATENCY_TOLERANT" {
			t.Fatalf("LatencyTolerance = %q, want tolerant", request.Requests[0].LatencyTolerance)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	err := publisher.BatchDeleteOneTimeProductOffers(context.Background(), OneTimeProductOfferBatchDeleteOptions{
		PackageName:      "com.example.app",
		ProductID:        "-",
		PurchaseOptionID: "-",
		Requests: []OneTimeProductOfferBatchDeleteRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro"},
			{ProductID: "coins_500", PurchaseOptionID: "rent", OfferID: "preorder"},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("BatchDeleteOneTimeProductOffers() error = %v", err)
	}
}

func TestBatchDeleteOneTimeProductOffersRejectsDryRunBeforeRequest(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))

	err := publisher.BatchDeleteOneTimeProductOffers(context.Background(), OneTimeProductOfferBatchDeleteOptions{
		PackageName:      "com.example.app",
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		Requests: []OneTimeProductOfferBatchDeleteRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro"},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected live validation error")
	}
}

func TestBatchUpdateOneTimeProductOfferStatesUsesBatchUpdateEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/coins_100/purchaseOptions/-/offers:batchUpdateStates" {
			t.Fatalf("path = %q, want one-time product offers batchUpdateStates endpoint", r.URL.Path)
		}
		var request androidpublisher.BatchUpdateOneTimeProductOfferStatesRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if len(request.Requests) != 2 {
			t.Fatalf("len(Requests) = %d, want 2", len(request.Requests))
		}
		deactivate := request.Requests[0].DeactivateOneTimeProductOfferRequest
		if deactivate == nil || deactivate.ProductId != "coins_100" || deactivate.PurchaseOptionId != "buy" || deactivate.OfferId != "intro" {
			t.Fatalf("first request = %#v, want deactivate intro", request.Requests[0])
		}
		if deactivate.LatencyTolerance != "PRODUCT_UPDATE_LATENCY_TOLERANCE_LATENCY_TOLERANT" {
			t.Fatalf("LatencyTolerance = %q, want tolerant", deactivate.LatencyTolerance)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"oneTimeProductOffers":[{"packageName":"com.example.app","productId":"coins_100","purchaseOptionId":"buy","offerId":"intro","state":"INACTIVE"}]}`)
	}))

	result, err := publisher.BatchUpdateOneTimeProductOfferStates(context.Background(), OneTimeProductOfferBatchStateUpdateOptions{
		PackageName:      "com.example.app",
		ProductID:        "coins_100",
		PurchaseOptionID: "-",
		Requests: []OneTimeProductOfferBatchDeleteRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro"},
			{ProductID: "coins_100", PurchaseOptionID: "rent", OfferID: "preorder"},
		},
		Action:           OneTimeProductOfferStateActionDeactivate,
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("BatchUpdateOneTimeProductOfferStates() error = %v", err)
	}
	if len(result.Offers) != 1 || result.Offers[0].OfferID != "intro" {
		t.Fatalf("Offers = %#v, want intro", result.Offers)
	}
}

func TestBatchPatchOneTimeProductOfferAvailabilityMergesAndBatchUpdates(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/coins_100/purchaseOptions/buy/offers:batchGet":
			if r.Method != http.MethodPost {
				t.Fatalf("method = %s, want POST batchGet", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"oneTimeProductOffers":[{"packageName":"com.example.app","productId":"coins_100","purchaseOptionId":"buy","offerId":"intro","regionalPricingAndAvailabilityConfigs":[{"regionCode":"US","availability":"AVAILABLE","relativeDiscount":0.5},{"regionCode":"FR","availability":"AVAILABLE","noOverride":{}},{"regionCode":"DE","availability":"AVAILABLE","noOverride":{}}]}]}`)
		case "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/coins_100/purchaseOptions/buy/offers:batchUpdate":
			if r.Method != http.MethodPost {
				t.Fatalf("method = %s, want POST batchUpdate", r.Method)
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("ReadAll() error = %v", err)
			}
			var request androidpublisher.BatchUpdateOneTimeProductOffersRequest
			if err := json.Unmarshal(body, &request); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if len(request.Requests) != 1 {
				t.Fatalf("len(Requests) = %d, want 1", len(request.Requests))
			}
			update := request.Requests[0]
			if update.UpdateMask != "regionalPricingAndAvailabilityConfigs" {
				t.Fatalf("UpdateMask = %q, want regionalPricingAndAvailabilityConfigs", update.UpdateMask)
			}
			if update.RegionsVersion == nil || update.RegionsVersion.Version != "2026/05" {
				t.Fatalf("RegionsVersion = %#v, want 2026/05", update.RegionsVersion)
			}
			if update.LatencyTolerance != "PRODUCT_UPDATE_LATENCY_TOLERANCE_LATENCY_TOLERANT" {
				t.Fatalf("LatencyTolerance = %q, want tolerant", update.LatencyTolerance)
			}
			configs := update.OneTimeProductOffer.RegionalPricingAndAvailabilityConfigs
			if len(configs) != 3 {
				t.Fatalf("len(RegionalPricingAndAvailabilityConfigs) = %d, want preserved plus added configs", len(configs))
			}
			if configs[0].RegionCode != "US" || configs[0].Availability != "NO_LONGER_AVAILABLE" || configs[0].RelativeDiscount != 0.5 {
				t.Fatalf("first config = %#v, want US no-longer-available with preserved discount", configs[0])
			}
			if configs[1].RegionCode != "FR" || configs[1].Availability != "AVAILABLE" || configs[1].NoOverride == nil {
				t.Fatalf("second config = %#v, want FR available with preserved noOverride", configs[1])
			}
			if configs[2].RegionCode != "DE" || configs[2].Availability != "AVAILABLE" || configs[2].NoOverride == nil {
				t.Fatalf("third config = %#v, want preserved DE noOverride", configs[2])
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"oneTimeProductOffers":[{"packageName":"com.example.app","productId":"coins_100","purchaseOptionId":"buy","offerId":"intro","regionalPricingAndAvailabilityConfigs":[{"regionCode":"US","availability":"NO_LONGER_AVAILABLE"},{"regionCode":"FR","availability":"AVAILABLE"}]}]}`)
		default:
			t.Fatalf("unexpected path = %q", r.URL.Path)
		}
	}))

	result, err := publisher.BatchPatchOneTimeProductOfferAvailability(context.Background(), OneTimeProductOfferBatchPatchAvailabilityOptions{
		PackageName:      "com.example.app",
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		RegionsVersion:   "2026/05",
		Requests: []OneTimeProductOfferAvailabilityPatchRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro", RegionCode: "US", Availability: OneTimeProductOfferAvailabilityNoLongerAvailable},
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro", RegionCode: "FR", Availability: OneTimeProductOfferAvailabilityAvailable},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("BatchPatchOneTimeProductOfferAvailability() error = %v", err)
	}
	if !result.Applied || len(result.Offers) != 1 || result.Offers[0].RegionalConfigs[0].Availability != "NO_LONGER_AVAILABLE" {
		t.Fatalf("result = %#v, want applied offer with Google availability enum", result)
	}
}

func TestBatchPatchOneTimeProductOfferAvailabilityRejectsMissingRegion(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/coins_100/purchaseOptions/buy/offers:batchGet" {
			t.Fatalf("path = %q, want batchGet only", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"oneTimeProductOffers":[{"packageName":"com.example.app","productId":"coins_100","purchaseOptionId":"buy","offerId":"intro","regionalPricingAndAvailabilityConfigs":[{"regionCode":"US","availability":"AVAILABLE","relativeDiscount":0.5}]}]}`)
	}))

	_, err := publisher.BatchPatchOneTimeProductOfferAvailability(context.Background(), OneTimeProductOfferBatchPatchAvailabilityOptions{
		PackageName:      "com.example.app",
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		RegionsVersion:   "2026/05",
		Requests: []OneTimeProductOfferAvailabilityPatchRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro", RegionCode: "FR", Availability: OneTimeProductOfferAvailabilityAvailable},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err == nil {
		t.Fatal("expected missing region validation error")
	}
	if !strings.Contains(err.Error(), "region that is not already configured") {
		t.Fatalf("error = %v, want missing configured region message", err)
	}
}

func TestBatchPatchOneTimeProductOfferRelativeDiscountsMergesAndBatchUpdates(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/coins_100/purchaseOptions/buy/offers:batchGet":
			if r.Method != http.MethodPost {
				t.Fatalf("method = %s, want POST batchGet", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"oneTimeProductOffers":[{"packageName":"com.example.app","productId":"coins_100","purchaseOptionId":"buy","offerId":"intro","regionalPricingAndAvailabilityConfigs":[{"regionCode":"US","availability":"AVAILABLE","relativeDiscount":0.25},{"regionCode":"FR","availability":"AVAILABLE","noOverride":{}},{"regionCode":"DE","availability":"AVAILABLE","absoluteDiscount":{"currencyCode":"EUR","units":"1","nanos":500000000}}]}]}`)
		case "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/coins_100/purchaseOptions/buy/offers:batchUpdate":
			if r.Method != http.MethodPost {
				t.Fatalf("method = %s, want POST batchUpdate", r.Method)
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("ReadAll() error = %v", err)
			}
			var request androidpublisher.BatchUpdateOneTimeProductOffersRequest
			if err := json.Unmarshal(body, &request); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if len(request.Requests) != 1 {
				t.Fatalf("len(Requests) = %d, want 1", len(request.Requests))
			}
			update := request.Requests[0]
			if update.UpdateMask != "regionalPricingAndAvailabilityConfigs" {
				t.Fatalf("UpdateMask = %q, want regionalPricingAndAvailabilityConfigs", update.UpdateMask)
			}
			if update.RegionsVersion == nil || update.RegionsVersion.Version != "2026/05" {
				t.Fatalf("RegionsVersion = %#v, want 2026/05", update.RegionsVersion)
			}
			if update.LatencyTolerance != "PRODUCT_UPDATE_LATENCY_TOLERANCE_LATENCY_TOLERANT" {
				t.Fatalf("LatencyTolerance = %q, want tolerant", update.LatencyTolerance)
			}
			configs := update.OneTimeProductOffer.RegionalPricingAndAvailabilityConfigs
			if len(configs) != 3 {
				t.Fatalf("len(RegionalPricingAndAvailabilityConfigs) = %d, want preserved configs", len(configs))
			}
			if configs[0].RegionCode != "US" || configs[0].Availability != "AVAILABLE" || configs[0].RelativeDiscount != 0.5 || configs[0].NoOverride != nil || configs[0].AbsoluteDiscount != nil {
				t.Fatalf("first config = %#v, want US relative discount with old price mode cleared", configs[0])
			}
			if configs[1].RegionCode != "FR" || configs[1].Availability != "AVAILABLE" || configs[1].RelativeDiscount != 0.75 || configs[1].NoOverride != nil || configs[1].AbsoluteDiscount != nil {
				t.Fatalf("second config = %#v, want FR relative discount with noOverride cleared", configs[1])
			}
			if configs[2].RegionCode != "DE" || configs[2].Availability != "AVAILABLE" || configs[2].AbsoluteDiscount == nil {
				t.Fatalf("third config = %#v, want preserved DE absolute discount", configs[2])
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"oneTimeProductOffers":[{"packageName":"com.example.app","productId":"coins_100","purchaseOptionId":"buy","offerId":"intro","regionalPricingAndAvailabilityConfigs":[{"regionCode":"US","availability":"AVAILABLE","relativeDiscount":0.5},{"regionCode":"FR","availability":"AVAILABLE","relativeDiscount":0.75}]}]}`)
		default:
			t.Fatalf("unexpected path = %q", r.URL.Path)
		}
	}))

	result, err := publisher.BatchPatchOneTimeProductOfferRelativeDiscounts(context.Background(), OneTimeProductOfferBatchPatchRelativeDiscountsOptions{
		PackageName:      "com.example.app",
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		RegionsVersion:   "2026/05",
		Requests: []OneTimeProductOfferRelativeDiscountPatchRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro", RegionCode: "US", RelativeDiscount: 0.5},
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro", RegionCode: "FR", RelativeDiscount: 0.75},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("BatchPatchOneTimeProductOfferRelativeDiscounts() error = %v", err)
	}
	if !result.Applied || len(result.Offers) != 1 || result.Offers[0].RegionalConfigs[0].RelativeDiscount != 0.5 {
		t.Fatalf("result = %#v, want applied offer with relative discount", result)
	}
}

func TestBatchPatchOneTimeProductOfferRelativeDiscountsRejectsMissingRegion(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/coins_100/purchaseOptions/buy/offers:batchGet" {
			t.Fatalf("path = %q, want batchGet only", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"oneTimeProductOffers":[{"packageName":"com.example.app","productId":"coins_100","purchaseOptionId":"buy","offerId":"intro","regionalPricingAndAvailabilityConfigs":[{"regionCode":"US","availability":"AVAILABLE","relativeDiscount":0.5}]}]}`)
	}))

	_, err := publisher.BatchPatchOneTimeProductOfferRelativeDiscounts(context.Background(), OneTimeProductOfferBatchPatchRelativeDiscountsOptions{
		PackageName:      "com.example.app",
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		RegionsVersion:   "2026/05",
		Requests: []OneTimeProductOfferRelativeDiscountPatchRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro", RegionCode: "FR", RelativeDiscount: 0.5},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err == nil {
		t.Fatal("expected missing region validation error")
	}
	if !strings.Contains(err.Error(), "region that is not already configured") {
		t.Fatalf("error = %v, want missing configured region message", err)
	}
}

func TestBatchPatchOneTimeProductOfferAbsoluteDiscountsMergesAndBatchUpdates(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/coins_100/purchaseOptions/buy/offers:batchGet":
			if r.Method != http.MethodPost {
				t.Fatalf("method = %s, want POST batchGet", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"oneTimeProductOffers":[{"packageName":"com.example.app","productId":"coins_100","purchaseOptionId":"buy","offerId":"intro","regionalPricingAndAvailabilityConfigs":[{"regionCode":"US","availability":"AVAILABLE","relativeDiscount":0.25},{"regionCode":"FR","availability":"AVAILABLE","noOverride":{}},{"regionCode":"DE","availability":"AVAILABLE","absoluteDiscount":{"currencyCode":"EUR","units":"1","nanos":500000000}}]}]}`)
		case "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/coins_100/purchaseOptions/buy/offers:batchUpdate":
			if r.Method != http.MethodPost {
				t.Fatalf("method = %s, want POST batchUpdate", r.Method)
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("ReadAll() error = %v", err)
			}
			rawBody := string(body)
			if strings.Contains(rawBody, "relativeDiscount") {
				t.Fatalf("body = %s, did not expect relativeDiscount in absolute discount patch request", rawBody)
			}
			if strings.Contains(rawBody, "noOverride") {
				t.Fatalf("body = %s, did not expect noOverride in absolute discount patch request", rawBody)
			}
			var request androidpublisher.BatchUpdateOneTimeProductOffersRequest
			if err := json.Unmarshal(body, &request); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if len(request.Requests) != 1 {
				t.Fatalf("len(Requests) = %d, want 1", len(request.Requests))
			}
			update := request.Requests[0]
			if update.UpdateMask != "regionalPricingAndAvailabilityConfigs" {
				t.Fatalf("UpdateMask = %q, want regionalPricingAndAvailabilityConfigs", update.UpdateMask)
			}
			configs := update.OneTimeProductOffer.RegionalPricingAndAvailabilityConfigs
			if len(configs) != 3 {
				t.Fatalf("len(RegionalPricingAndAvailabilityConfigs) = %d, want preserved configs", len(configs))
			}
			if configs[0].RegionCode != "US" || configs[0].Availability != "AVAILABLE" || configs[0].AbsoluteDiscount == nil || configs[0].AbsoluteDiscount.CurrencyCode != "USD" || configs[0].AbsoluteDiscount.Units != 1 || configs[0].RelativeDiscount != 0 || configs[0].NoOverride != nil {
				t.Fatalf("first config = %#v, want US absolute discount with old price mode cleared", configs[0])
			}
			if configs[1].RegionCode != "FR" || configs[1].Availability != "AVAILABLE" || configs[1].AbsoluteDiscount == nil || configs[1].AbsoluteDiscount.CurrencyCode != "EUR" || configs[1].NoOverride != nil {
				t.Fatalf("second config = %#v, want FR absolute discount with noOverride cleared", configs[1])
			}
			if configs[2].RegionCode != "DE" || configs[2].Availability != "AVAILABLE" || configs[2].AbsoluteDiscount == nil {
				t.Fatalf("third config = %#v, want preserved DE absolute discount", configs[2])
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"oneTimeProductOffers":[{"packageName":"com.example.app","productId":"coins_100","purchaseOptionId":"buy","offerId":"intro","regionalPricingAndAvailabilityConfigs":[{"regionCode":"US","availability":"AVAILABLE","absoluteDiscount":{"currencyCode":"USD","units":"1"}},{"regionCode":"FR","availability":"AVAILABLE","absoluteDiscount":{"currencyCode":"EUR","nanos":500000000}}]}]}`)
		default:
			t.Fatalf("unexpected path = %q", r.URL.Path)
		}
	}))

	result, err := publisher.BatchPatchOneTimeProductOfferAbsoluteDiscounts(context.Background(), OneTimeProductOfferBatchPatchAbsoluteDiscountsOptions{
		PackageName:      "com.example.app",
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		RegionsVersion:   "2026/05",
		Requests: []OneTimeProductOfferAbsoluteDiscountPatchRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro", RegionCode: "US", AbsoluteDiscount: Money{CurrencyCode: "USD", Units: 1}},
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro", RegionCode: "FR", AbsoluteDiscount: Money{CurrencyCode: "EUR", Nanos: 500000000}},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("BatchPatchOneTimeProductOfferAbsoluteDiscounts() error = %v", err)
	}
	if !result.Applied || len(result.Offers) != 1 || result.Offers[0].RegionalConfigs[0].AbsoluteDiscount == nil {
		t.Fatalf("result = %#v, want applied offer with absolute discount", result)
	}
}

func TestBatchPatchOneTimeProductOfferAbsoluteDiscountsRejectsMissingRegion(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/coins_100/purchaseOptions/buy/offers:batchGet" {
			t.Fatalf("path = %q, want batchGet only", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"oneTimeProductOffers":[{"packageName":"com.example.app","productId":"coins_100","purchaseOptionId":"buy","offerId":"intro","regionalPricingAndAvailabilityConfigs":[{"regionCode":"US","availability":"AVAILABLE","relativeDiscount":0.5}]}]}`)
	}))

	_, err := publisher.BatchPatchOneTimeProductOfferAbsoluteDiscounts(context.Background(), OneTimeProductOfferBatchPatchAbsoluteDiscountsOptions{
		PackageName:      "com.example.app",
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		RegionsVersion:   "2026/05",
		Requests: []OneTimeProductOfferAbsoluteDiscountPatchRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro", RegionCode: "FR", AbsoluteDiscount: Money{CurrencyCode: "EUR", Units: 1}},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err == nil {
		t.Fatal("expected missing region validation error")
	}
	if !strings.Contains(err.Error(), "region that is not already configured") {
		t.Fatalf("error = %v, want missing configured region message", err)
	}
}

func TestBatchPatchOneTimeProductOfferNoOverridesMergesAndBatchUpdates(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/coins_100/purchaseOptions/buy/offers:batchGet":
			if r.Method != http.MethodPost {
				t.Fatalf("method = %s, want POST batchGet", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"oneTimeProductOffers":[{"packageName":"com.example.app","productId":"coins_100","purchaseOptionId":"buy","offerId":"intro","regionalPricingAndAvailabilityConfigs":[{"regionCode":"US","availability":"AVAILABLE","relativeDiscount":0.25},{"regionCode":"FR","availability":"AVAILABLE","absoluteDiscount":{"currencyCode":"EUR","units":"1","nanos":500000000}}]}]}`)
		case "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/coins_100/purchaseOptions/buy/offers:batchUpdate":
			if r.Method != http.MethodPost {
				t.Fatalf("method = %s, want POST batchUpdate", r.Method)
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("ReadAll() error = %v", err)
			}
			rawBody := string(body)
			if strings.Contains(rawBody, "relativeDiscount") {
				t.Fatalf("body = %s, did not expect relativeDiscount in no-override patch request", rawBody)
			}
			if strings.Contains(rawBody, "absoluteDiscount") {
				t.Fatalf("body = %s, did not expect absoluteDiscount in no-override patch request", rawBody)
			}
			var request androidpublisher.BatchUpdateOneTimeProductOffersRequest
			if err := json.Unmarshal(body, &request); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if len(request.Requests) != 1 {
				t.Fatalf("len(Requests) = %d, want 1", len(request.Requests))
			}
			update := request.Requests[0]
			if update.UpdateMask != "regionalPricingAndAvailabilityConfigs" {
				t.Fatalf("UpdateMask = %q, want regionalPricingAndAvailabilityConfigs", update.UpdateMask)
			}
			configs := update.OneTimeProductOffer.RegionalPricingAndAvailabilityConfigs
			if len(configs) != 2 {
				t.Fatalf("len(RegionalPricingAndAvailabilityConfigs) = %d, want preserved configs", len(configs))
			}
			if configs[0].RegionCode != "US" || configs[0].Availability != "AVAILABLE" || configs[0].NoOverride == nil || configs[0].RelativeDiscount != 0 || configs[0].AbsoluteDiscount != nil {
				t.Fatalf("first config = %#v, want US noOverride with old price mode cleared", configs[0])
			}
			if configs[1].RegionCode != "FR" || configs[1].Availability != "AVAILABLE" || configs[1].NoOverride == nil || configs[1].AbsoluteDiscount != nil {
				t.Fatalf("second config = %#v, want FR noOverride with absolute discount cleared", configs[1])
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"oneTimeProductOffers":[{"packageName":"com.example.app","productId":"coins_100","purchaseOptionId":"buy","offerId":"intro","regionalPricingAndAvailabilityConfigs":[{"regionCode":"US","availability":"AVAILABLE","noOverride":{}},{"regionCode":"FR","availability":"AVAILABLE","noOverride":{}}]}]}`)
		default:
			t.Fatalf("unexpected path = %q", r.URL.Path)
		}
	}))

	result, err := publisher.BatchPatchOneTimeProductOfferNoOverrides(context.Background(), OneTimeProductOfferBatchPatchNoOverridesOptions{
		PackageName:      "com.example.app",
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		RegionsVersion:   "2026/05",
		Requests: []OneTimeProductOfferNoOverridePatchRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro", RegionCode: "US", NoOverride: true},
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro", RegionCode: "FR", NoOverride: true},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("BatchPatchOneTimeProductOfferNoOverrides() error = %v", err)
	}
	if !result.Applied || len(result.Offers) != 1 || !result.Offers[0].RegionalConfigs[0].NoOverride {
		t.Fatalf("result = %#v, want applied offer with no override", result)
	}
}

func TestBatchPatchOneTimeProductOfferNoOverridesRejectsMissingRegion(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/oneTimeProducts/coins_100/purchaseOptions/buy/offers:batchGet" {
			t.Fatalf("path = %q, want batchGet only", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"oneTimeProductOffers":[{"packageName":"com.example.app","productId":"coins_100","purchaseOptionId":"buy","offerId":"intro","regionalPricingAndAvailabilityConfigs":[{"regionCode":"US","availability":"AVAILABLE","relativeDiscount":0.5}]}]}`)
	}))

	_, err := publisher.BatchPatchOneTimeProductOfferNoOverrides(context.Background(), OneTimeProductOfferBatchPatchNoOverridesOptions{
		PackageName:      "com.example.app",
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		RegionsVersion:   "2026/05",
		Requests: []OneTimeProductOfferNoOverridePatchRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro", RegionCode: "FR", NoOverride: true},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err == nil {
		t.Fatal("expected missing region validation error")
	}
	if !strings.Contains(err.Error(), "region that is not already configured") {
		t.Fatalf("error = %v, want missing configured region message", err)
	}
}

func TestBatchUpdateOneTimeProductOfferStatesRejectsDryRunBeforeRequest(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))

	_, err := publisher.BatchUpdateOneTimeProductOfferStates(context.Background(), OneTimeProductOfferBatchStateUpdateOptions{
		PackageName:      "com.example.app",
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		Requests: []OneTimeProductOfferBatchDeleteRequest{
			{ProductID: "coins_100", PurchaseOptionID: "buy", OfferID: "intro"},
		},
		Action:           OneTimeProductOfferStateActionActivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected live validation error")
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

func TestDeleteSubscriptionOfferUsesDeleteEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions/premium/basePlans/monthly/offers/intro" {
			t.Fatalf("path = %q, want subscription offer delete endpoint", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	err := publisher.DeleteSubscriptionOffer(context.Background(), SubscriptionOfferDeleteOptions{
		PackageName: "com.example.app",
		ProductID:   "premium",
		BasePlanID:  "monthly",
		OfferID:     "intro",
		Confirm:     true,
	})
	if err != nil {
		t.Fatalf("DeleteSubscriptionOffer() error = %v", err)
	}
}

func TestDeleteSubscriptionOfferRejectsDryRunBeforeRequest(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))

	err := publisher.DeleteSubscriptionOffer(context.Background(), SubscriptionOfferDeleteOptions{
		PackageName: "com.example.app",
		ProductID:   "premium",
		BasePlanID:  "monthly",
		OfferID:     "intro",
		DryRun:      true,
	})
	if err == nil {
		t.Fatal("expected live validation error")
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

func TestBatchUpdateSubscriptionOfferStatesUsesBatchUpdateEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions/premium/basePlans/-/offers:batchUpdateStates" {
			t.Fatalf("path = %q, want subscription offers batchUpdateStates endpoint", r.URL.Path)
		}
		var request androidpublisher.BatchUpdateSubscriptionOfferStatesRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if len(request.Requests) != 2 {
			t.Fatalf("len(Requests) = %d, want 2", len(request.Requests))
		}
		deactivate := request.Requests[0].DeactivateSubscriptionOfferRequest
		if deactivate == nil || deactivate.ProductId != "premium" || deactivate.BasePlanId != "monthly" || deactivate.OfferId != "intro" {
			t.Fatalf("first request = %#v, want deactivate intro", request.Requests[0])
		}
		if deactivate.LatencyTolerance != "PRODUCT_UPDATE_LATENCY_TOLERANCE_LATENCY_TOLERANT" {
			t.Fatalf("LatencyTolerance = %q, want tolerant", deactivate.LatencyTolerance)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"subscriptionOffers":[{"packageName":"com.example.app","productId":"premium","basePlanId":"annual","offerId":"winback","state":"INACTIVE"},{"packageName":"com.example.app","productId":"premium","basePlanId":"monthly","offerId":"intro","state":"INACTIVE"}]}`)
	}))

	result, err := publisher.BatchUpdateSubscriptionOfferStates(context.Background(), SubscriptionOfferBatchStateUpdateOptions{
		PackageName: "com.example.app",
		ProductID:   "premium",
		BasePlanID:  "-",
		Requests: []SubscriptionOfferBatchMutationRequest{
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro"},
			{ProductID: "premium", BasePlanID: "annual", OfferID: "winback"},
		},
		Action:           SubscriptionOfferStateActionDeactivate,
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("BatchUpdateSubscriptionOfferStates() error = %v", err)
	}
	if len(result.Offers) != 2 {
		t.Fatalf("len(Offers) = %d, want 2: %#v", len(result.Offers), result.Offers)
	}
	if result.Offers[0].OfferID != "intro" || result.Offers[1].OfferID != "winback" {
		t.Fatalf("Offers = %#v, want request order intro then winback", result.Offers)
	}
}

func TestBatchPatchSubscriptionOfferAvailabilityMergesAndBatchUpdates(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions/premium/basePlans/monthly/offers/intro" {
				t.Fatalf("path = %q, want subscription offer get endpoint", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{
				"packageName":"com.example.app",
				"productId":"premium",
				"basePlanId":"monthly",
				"offerId":"intro",
				"regionalConfigs":[
					{"regionCode":"US","newSubscriberAvailability":true},
					{"regionCode":"DE","newSubscriberAvailability":true}
				]
			}`)
		case http.MethodPost:
			if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions/premium/basePlans/monthly/offers:batchUpdate" {
				t.Fatalf("path = %q, want subscription offer batchUpdate endpoint", r.URL.Path)
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("ReadAll() error = %v", err)
			}
			if !strings.Contains(string(body), `"newSubscriberAvailability":false`) {
				t.Fatalf("body = %s, want explicit false availability", body)
			}
			var request androidpublisher.BatchUpdateSubscriptionOffersRequest
			if err := json.Unmarshal(body, &request); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if len(request.Requests) != 1 {
				t.Fatalf("len(Requests) = %d, want 1", len(request.Requests))
			}
			update := request.Requests[0]
			if update.UpdateMask != "regionalConfigs" {
				t.Fatalf("UpdateMask = %q, want regionalConfigs", update.UpdateMask)
			}
			if update.RegionsVersion == nil || update.RegionsVersion.Version != "2026/05" {
				t.Fatalf("RegionsVersion = %#v, want 2026/05", update.RegionsVersion)
			}
			if update.LatencyTolerance != "PRODUCT_UPDATE_LATENCY_TOLERANCE_LATENCY_TOLERANT" {
				t.Fatalf("LatencyTolerance = %q, want tolerant", update.LatencyTolerance)
			}
			configs := update.SubscriptionOffer.RegionalConfigs
			if len(configs) != 3 {
				t.Fatalf("len(RegionalConfigs) = %d, want preserved plus added configs", len(configs))
			}
			if configs[0].RegionCode != "US" || configs[0].NewSubscriberAvailability {
				t.Fatalf("first config = %#v, want US false patch", configs[0])
			}
			if configs[1].RegionCode != "DE" || !configs[1].NewSubscriberAvailability {
				t.Fatalf("second config = %#v, want preserved DE true", configs[1])
			}
			if configs[2].RegionCode != "FR" || !configs[2].NewSubscriberAvailability {
				t.Fatalf("third config = %#v, want added FR true", configs[2])
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"subscriptionOffers":[{"packageName":"com.example.app","productId":"premium","basePlanId":"monthly","offerId":"intro","regionalConfigs":[{"regionCode":"US","newSubscriberAvailability":false},{"regionCode":"FR","newSubscriberAvailability":true}]}]}`)
		default:
			t.Fatalf("method = %s, want GET or POST", r.Method)
		}
	}))

	result, err := publisher.BatchPatchSubscriptionOfferAvailability(context.Background(), SubscriptionOfferBatchPatchAvailabilityOptions{
		PackageName:    "com.example.app",
		ProductID:      "premium",
		BasePlanID:     "monthly",
		RegionsVersion: "2026/05",
		Requests: []SubscriptionOfferAvailabilityPatchRequest{
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", RegionCode: "US", Availability: false},
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", RegionCode: "FR", Availability: true},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("BatchPatchSubscriptionOfferAvailability() error = %v", err)
	}
	if !result.Applied || len(result.Offers) != 1 || result.Offers[0].OfferID != "intro" {
		t.Fatalf("result = %#v, want applied intro offer", result)
	}
}

func TestBatchPatchSubscriptionOfferPhaseRelativeDiscountsMergesAndBatchUpdates(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions/premium/basePlans/monthly/offers/intro" {
				t.Fatalf("path = %q, want subscription offer get endpoint", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"packageName":"com.example.app","productId":"premium","basePlanId":"monthly","offerId":"intro","regionalConfigs":[{"regionCode":"US","newSubscriberAvailability":true},{"regionCode":"FR","newSubscriberAvailability":true}],"phases":[{"duration":"P1M","recurrenceCount":1,"regionalConfigs":[{"regionCode":"US","price":{"currencyCode":"USD","units":"1"}},{"regionCode":"FR","absoluteDiscount":{"currencyCode":"EUR","units":"2"}}],"otherRegionsConfig":{"relativeDiscount":0.5}},{"duration":"P2M","recurrenceCount":2,"regionalConfigs":[{"regionCode":"US","free":{}},{"regionCode":"FR","relativeDiscount":0.25}]}]}`)
		case http.MethodPost:
			if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions/premium/basePlans/monthly/offers:batchUpdate" {
				t.Fatalf("path = %q, want subscription offer batchUpdate endpoint", r.URL.Path)
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("ReadAll() error = %v", err)
			}
			var request androidpublisher.BatchUpdateSubscriptionOffersRequest
			if err := json.Unmarshal(body, &request); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if len(request.Requests) != 1 {
				t.Fatalf("len(Requests) = %d, want 1", len(request.Requests))
			}
			update := request.Requests[0]
			if update.UpdateMask != "phases" {
				t.Fatalf("UpdateMask = %q, want phases", update.UpdateMask)
			}
			if update.RegionsVersion == nil || update.RegionsVersion.Version != "2026/05" {
				t.Fatalf("RegionsVersion = %#v, want 2026/05", update.RegionsVersion)
			}
			if update.LatencyTolerance != "PRODUCT_UPDATE_LATENCY_TOLERANCE_LATENCY_TOLERANT" {
				t.Fatalf("LatencyTolerance = %q, want tolerant", update.LatencyTolerance)
			}
			phases := update.SubscriptionOffer.Phases
			if len(phases) != 2 {
				t.Fatalf("len(Phases) = %d, want preserved phases", len(phases))
			}
			firstUS := phases[0].RegionalConfigs[0]
			if firstUS.RegionCode != "US" || firstUS.RelativeDiscount != 0.75 || firstUS.Price != nil || firstUS.AbsoluteDiscount != nil || firstUS.Free != nil {
				t.Fatalf("first US phase config = %#v, want relative discount with old price modes cleared", firstUS)
			}
			firstFR := phases[0].RegionalConfigs[1]
			if firstFR.RegionCode != "FR" || firstFR.AbsoluteDiscount == nil || firstFR.RelativeDiscount != 0 {
				t.Fatalf("first FR phase config = %#v, want preserved absolute discount", firstFR)
			}
			secondFR := phases[1].RegionalConfigs[1]
			if secondFR.RegionCode != "FR" || secondFR.RelativeDiscount != 0.5 {
				t.Fatalf("second FR phase config = %#v, want patched relative discount", secondFR)
			}
			if phases[0].OtherRegionsConfig == nil || phases[0].OtherRegionsConfig.RelativeDiscount != 0.5 {
				t.Fatalf("OtherRegionsConfig = %#v, want preserved relative discount", phases[0].OtherRegionsConfig)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"subscriptionOffers":[{"packageName":"com.example.app","productId":"premium","basePlanId":"monthly","offerId":"intro","phases":[{"duration":"P1M","recurrenceCount":1,"regionalConfigs":[{"regionCode":"US","relativeDiscount":0.75},{"regionCode":"FR","absoluteDiscount":{"currencyCode":"EUR","units":"2"}}]},{"duration":"P2M","recurrenceCount":2,"regionalConfigs":[{"regionCode":"US","free":{}},{"regionCode":"FR","relativeDiscount":0.5}]}]}]}`)
		default:
			t.Fatalf("method = %s, want GET or POST", r.Method)
		}
	}))

	result, err := publisher.BatchPatchSubscriptionOfferPhaseRelativeDiscounts(context.Background(), SubscriptionOfferBatchPatchPhaseRelativeDiscountsOptions{
		PackageName:    "com.example.app",
		ProductID:      "premium",
		BasePlanID:     "monthly",
		RegionsVersion: "2026/05",
		Requests: []SubscriptionOfferPhaseRelativeDiscountPatchRequest{
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: 0, RegionCode: "US", RelativeDiscount: 0.75},
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: 1, RegionCode: "FR", RelativeDiscount: 0.5},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("BatchPatchSubscriptionOfferPhaseRelativeDiscounts() error = %v", err)
	}
	if !result.Applied || len(result.Offers) != 1 || result.Offers[0].Phases[0].RegionalConfigs[0].RelativeDiscount != 0.75 {
		t.Fatalf("result = %#v, want applied offer with relative discount", result)
	}
}

func TestBatchPatchSubscriptionOfferPhaseRelativeDiscountsRejectsMissingPhaseOrRegion(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions/premium/basePlans/monthly/offers/intro" {
			t.Fatalf("path = %q, want get only", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"packageName":"com.example.app","productId":"premium","basePlanId":"monthly","offerId":"intro","phases":[{"duration":"P1M","recurrenceCount":1,"regionalConfigs":[{"regionCode":"US","relativeDiscount":0.75}]}]}`)
	}))

	_, err := publisher.BatchPatchSubscriptionOfferPhaseRelativeDiscounts(context.Background(), SubscriptionOfferBatchPatchPhaseRelativeDiscountsOptions{
		PackageName:    "com.example.app",
		ProductID:      "premium",
		BasePlanID:     "monthly",
		RegionsVersion: "2026/05",
		Requests: []SubscriptionOfferPhaseRelativeDiscountPatchRequest{
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: 1, RegionCode: "FR", RelativeDiscount: 0.5},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err == nil {
		t.Fatal("expected missing phase or region validation error")
	}
	if !strings.Contains(err.Error(), "phase or region that is not already configured") {
		t.Fatalf("error = %v, want missing phase or region message", err)
	}
}

func TestBatchPatchSubscriptionOfferPhaseAbsoluteDiscountsMergesAndBatchUpdates(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions/premium/basePlans/monthly/offers/intro" {
				t.Fatalf("path = %q, want subscription offer get endpoint", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"packageName":"com.example.app","productId":"premium","basePlanId":"monthly","offerId":"intro","phases":[{"duration":"P1M","recurrenceCount":1,"regionalConfigs":[{"regionCode":"US","relativeDiscount":0.75},{"regionCode":"FR","absoluteDiscount":{"currencyCode":"EUR","units":"2"}}],"otherRegionsConfig":{"absoluteDiscounts":{"usdPrice":{"currencyCode":"USD","units":"1"},"eurPrice":{"currencyCode":"EUR","units":"1"}}}},{"duration":"P2M","recurrenceCount":2,"regionalConfigs":[{"regionCode":"US","free":{}},{"regionCode":"FR","price":{"currencyCode":"EUR","units":"3"}}]}]}`)
		case http.MethodPost:
			if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions/premium/basePlans/monthly/offers:batchUpdate" {
				t.Fatalf("path = %q, want subscription offer batchUpdate endpoint", r.URL.Path)
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("ReadAll() error = %v", err)
			}
			rawBody := string(body)
			if strings.Contains(rawBody, `"relativeDiscount"`) {
				t.Fatalf("body = %s, did not expect relativeDiscount in absolute discount patch request", rawBody)
			}
			var request androidpublisher.BatchUpdateSubscriptionOffersRequest
			if err := json.Unmarshal(body, &request); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if len(request.Requests) != 1 {
				t.Fatalf("len(Requests) = %d, want 1", len(request.Requests))
			}
			update := request.Requests[0]
			if update.UpdateMask != "phases" {
				t.Fatalf("UpdateMask = %q, want phases", update.UpdateMask)
			}
			phases := update.SubscriptionOffer.Phases
			if len(phases) != 2 {
				t.Fatalf("len(Phases) = %d, want preserved phases", len(phases))
			}
			firstUS := phases[0].RegionalConfigs[0]
			if firstUS.RegionCode != "US" || firstUS.AbsoluteDiscount == nil || firstUS.AbsoluteDiscount.CurrencyCode != "USD" || firstUS.AbsoluteDiscount.Units != 1 || firstUS.RelativeDiscount != 0 || firstUS.Price != nil || firstUS.Free != nil {
				t.Fatalf("first US phase config = %#v, want absolute discount with old price modes cleared", firstUS)
			}
			firstFR := phases[0].RegionalConfigs[1]
			if firstFR.RegionCode != "FR" || firstFR.AbsoluteDiscount == nil || firstFR.AbsoluteDiscount.CurrencyCode != "EUR" {
				t.Fatalf("first FR phase config = %#v, want preserved absolute discount", firstFR)
			}
			if phases[0].OtherRegionsConfig == nil || phases[0].OtherRegionsConfig.AbsoluteDiscounts == nil || phases[0].OtherRegionsConfig.AbsoluteDiscounts.UsdPrice == nil {
				t.Fatalf("OtherRegionsConfig = %#v, want preserved other-regions absolute discounts", phases[0].OtherRegionsConfig)
			}
			secondFR := phases[1].RegionalConfigs[1]
			if secondFR.RegionCode != "FR" || secondFR.AbsoluteDiscount == nil || secondFR.AbsoluteDiscount.CurrencyCode != "EUR" || secondFR.AbsoluteDiscount.Nanos != 500000000 || secondFR.Price != nil {
				t.Fatalf("second FR phase config = %#v, want patched absolute discount", secondFR)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"subscriptionOffers":[{"packageName":"com.example.app","productId":"premium","basePlanId":"monthly","offerId":"intro","phases":[{"duration":"P1M","recurrenceCount":1,"regionalConfigs":[{"regionCode":"US","absoluteDiscount":{"currencyCode":"USD","units":"1"}},{"regionCode":"FR","absoluteDiscount":{"currencyCode":"EUR","units":"2"}}]},{"duration":"P2M","recurrenceCount":2,"regionalConfigs":[{"regionCode":"US","free":{}},{"regionCode":"FR","absoluteDiscount":{"currencyCode":"EUR","nanos":500000000}}]}]}]}`)
		default:
			t.Fatalf("method = %s, want GET or POST", r.Method)
		}
	}))

	result, err := publisher.BatchPatchSubscriptionOfferPhaseAbsoluteDiscounts(context.Background(), SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsOptions{
		PackageName:    "com.example.app",
		ProductID:      "premium",
		BasePlanID:     "monthly",
		RegionsVersion: "2026/05",
		Requests: []SubscriptionOfferPhaseAbsoluteDiscountPatchRequest{
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: 0, RegionCode: "US", AbsoluteDiscount: Money{CurrencyCode: "USD", Units: 1}},
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: 1, RegionCode: "FR", AbsoluteDiscount: Money{CurrencyCode: "EUR", Nanos: 500000000}},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("BatchPatchSubscriptionOfferPhaseAbsoluteDiscounts() error = %v", err)
	}
	if !result.Applied || len(result.Offers) != 1 || result.Offers[0].Phases[0].RegionalConfigs[0].AbsoluteDiscount == nil {
		t.Fatalf("result = %#v, want applied offer with absolute discount", result)
	}
}

func TestBatchPatchSubscriptionOfferPhaseAbsoluteDiscountsRejectsMissingPhaseOrRegion(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions/premium/basePlans/monthly/offers/intro" {
			t.Fatalf("path = %q, want get only", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"packageName":"com.example.app","productId":"premium","basePlanId":"monthly","offerId":"intro","phases":[{"duration":"P1M","recurrenceCount":1,"regionalConfigs":[{"regionCode":"US","relativeDiscount":0.75}]}]}`)
	}))

	_, err := publisher.BatchPatchSubscriptionOfferPhaseAbsoluteDiscounts(context.Background(), SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsOptions{
		PackageName:    "com.example.app",
		ProductID:      "premium",
		BasePlanID:     "monthly",
		RegionsVersion: "2026/05",
		Requests: []SubscriptionOfferPhaseAbsoluteDiscountPatchRequest{
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: 1, RegionCode: "FR", AbsoluteDiscount: Money{CurrencyCode: "EUR", Units: 1}},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err == nil {
		t.Fatal("expected missing phase or region validation error")
	}
	if !strings.Contains(err.Error(), "phase or region that is not already configured") {
		t.Fatalf("error = %v, want missing phase or region message", err)
	}
}

func TestBatchPatchSubscriptionOfferPhasePricesMergesAndBatchUpdates(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions/premium/basePlans/monthly/offers/intro" {
				t.Fatalf("path = %q, want subscription offer get endpoint", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"packageName":"com.example.app","productId":"premium","basePlanId":"monthly","offerId":"intro","phases":[{"duration":"P1M","recurrenceCount":1,"regionalConfigs":[{"regionCode":"US","relativeDiscount":0.75},{"regionCode":"FR","absoluteDiscount":{"currencyCode":"EUR","units":"2"}}],"otherRegionsConfig":{"absoluteDiscounts":{"usdPrice":{"currencyCode":"USD","units":"1"},"eurPrice":{"currencyCode":"EUR","units":"1"}}}},{"duration":"P2M","recurrenceCount":2,"regionalConfigs":[{"regionCode":"US","free":{}},{"regionCode":"FR","absoluteDiscount":{"currencyCode":"EUR","units":"3"}}]}]}`)
		case http.MethodPost:
			if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions/premium/basePlans/monthly/offers:batchUpdate" {
				t.Fatalf("path = %q, want subscription offer batchUpdate endpoint", r.URL.Path)
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("ReadAll() error = %v", err)
			}
			var request androidpublisher.BatchUpdateSubscriptionOffersRequest
			if err := json.Unmarshal(body, &request); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if len(request.Requests) != 1 {
				t.Fatalf("len(Requests) = %d, want 1", len(request.Requests))
			}
			update := request.Requests[0]
			if update.UpdateMask != "phases" {
				t.Fatalf("UpdateMask = %q, want phases", update.UpdateMask)
			}
			if update.RegionsVersion == nil || update.RegionsVersion.Version != "2026/05" {
				t.Fatalf("RegionsVersion = %#v, want 2026/05", update.RegionsVersion)
			}
			if update.LatencyTolerance != "PRODUCT_UPDATE_LATENCY_TOLERANCE_LATENCY_TOLERANT" {
				t.Fatalf("LatencyTolerance = %q, want tolerant", update.LatencyTolerance)
			}
			phases := update.SubscriptionOffer.Phases
			if len(phases) != 2 {
				t.Fatalf("len(Phases) = %d, want preserved phases", len(phases))
			}
			firstUS := phases[0].RegionalConfigs[0]
			if firstUS.RegionCode != "US" || firstUS.Price == nil || firstUS.Price.CurrencyCode != "USD" || firstUS.Price.Units != 1 || firstUS.RelativeDiscount != 0 || firstUS.AbsoluteDiscount != nil || firstUS.Free != nil {
				t.Fatalf("first US phase config = %#v, want price with old price modes cleared", firstUS)
			}
			firstFR := phases[0].RegionalConfigs[1]
			if firstFR.RegionCode != "FR" || firstFR.AbsoluteDiscount == nil || firstFR.AbsoluteDiscount.CurrencyCode != "EUR" {
				t.Fatalf("first FR phase config = %#v, want preserved absolute discount", firstFR)
			}
			if phases[0].OtherRegionsConfig == nil || phases[0].OtherRegionsConfig.AbsoluteDiscounts == nil || phases[0].OtherRegionsConfig.AbsoluteDiscounts.UsdPrice == nil {
				t.Fatalf("OtherRegionsConfig = %#v, want preserved other-regions absolute discounts", phases[0].OtherRegionsConfig)
			}
			secondUS := phases[1].RegionalConfigs[0]
			if secondUS.RegionCode != "US" || secondUS.Free == nil {
				t.Fatalf("second US phase config = %#v, want preserved free config", secondUS)
			}
			secondFR := phases[1].RegionalConfigs[1]
			if secondFR.RegionCode != "FR" || secondFR.Price == nil || secondFR.Price.CurrencyCode != "EUR" || secondFR.Price.Nanos != 500000000 || secondFR.AbsoluteDiscount != nil || secondFR.Free != nil {
				t.Fatalf("second FR phase config = %#v, want patched price", secondFR)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"subscriptionOffers":[{"packageName":"com.example.app","productId":"premium","basePlanId":"monthly","offerId":"intro","phases":[{"duration":"P1M","recurrenceCount":1,"regionalConfigs":[{"regionCode":"US","price":{"currencyCode":"USD","units":"1"}},{"regionCode":"FR","absoluteDiscount":{"currencyCode":"EUR","units":"2"}}]},{"duration":"P2M","recurrenceCount":2,"regionalConfigs":[{"regionCode":"US","free":{}},{"regionCode":"FR","price":{"currencyCode":"EUR","nanos":500000000}}]}]}]}`)
		default:
			t.Fatalf("method = %s, want GET or POST", r.Method)
		}
	}))

	result, err := publisher.BatchPatchSubscriptionOfferPhasePrices(context.Background(), SubscriptionOfferBatchPatchPhasePricesOptions{
		PackageName:    "com.example.app",
		ProductID:      "premium",
		BasePlanID:     "monthly",
		RegionsVersion: "2026/05",
		Requests: []SubscriptionOfferPhasePricePatchRequest{
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: 0, RegionCode: "US", Price: Money{CurrencyCode: "USD", Units: 1}},
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: 1, RegionCode: "FR", Price: Money{CurrencyCode: "EUR", Nanos: 500000000}},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("BatchPatchSubscriptionOfferPhasePrices() error = %v", err)
	}
	if !result.Applied || len(result.Offers) != 1 || result.Offers[0].Phases[0].RegionalConfigs[0].Price == nil {
		t.Fatalf("result = %#v, want applied offer with price", result)
	}
}

func TestBatchPatchSubscriptionOfferPhasePricesRejectsMissingPhaseOrRegion(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions/premium/basePlans/monthly/offers/intro" {
			t.Fatalf("path = %q, want get only", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"packageName":"com.example.app","productId":"premium","basePlanId":"monthly","offerId":"intro","phases":[{"duration":"P1M","recurrenceCount":1,"regionalConfigs":[{"regionCode":"US","relativeDiscount":0.75}]}]}`)
	}))

	_, err := publisher.BatchPatchSubscriptionOfferPhasePrices(context.Background(), SubscriptionOfferBatchPatchPhasePricesOptions{
		PackageName:    "com.example.app",
		ProductID:      "premium",
		BasePlanID:     "monthly",
		RegionsVersion: "2026/05",
		Requests: []SubscriptionOfferPhasePricePatchRequest{
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: 1, RegionCode: "FR", Price: Money{CurrencyCode: "EUR", Units: 1}},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err == nil {
		t.Fatal("expected missing phase or region validation error")
	}
	if !strings.Contains(err.Error(), "phase or region that is not already configured") {
		t.Fatalf("error = %v, want missing phase or region validation", err)
	}
}

func TestBatchPatchSubscriptionOfferPhaseFreeMergesAndBatchUpdates(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions/premium/basePlans/monthly/offers/intro" {
				t.Fatalf("path = %q, want subscription offer get endpoint", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"packageName":"com.example.app","productId":"premium","basePlanId":"monthly","offerId":"intro","phases":[{"duration":"P1M","recurrenceCount":1,"regionalConfigs":[{"regionCode":"US","price":{"currencyCode":"USD","units":"1"}},{"regionCode":"FR","absoluteDiscount":{"currencyCode":"EUR","units":"2"}}],"otherRegionsConfig":{"absoluteDiscounts":{"usdPrice":{"currencyCode":"USD","units":"1"},"eurPrice":{"currencyCode":"EUR","units":"1"}}}},{"duration":"P2M","recurrenceCount":2,"regionalConfigs":[{"regionCode":"US","relativeDiscount":0.75},{"regionCode":"FR","price":{"currencyCode":"EUR","nanos":500000000}}]}]}`)
		case http.MethodPost:
			if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions/premium/basePlans/monthly/offers:batchUpdate" {
				t.Fatalf("path = %q, want subscription offer batchUpdate endpoint", r.URL.Path)
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("ReadAll() error = %v", err)
			}
			var request androidpublisher.BatchUpdateSubscriptionOffersRequest
			if err := json.Unmarshal(body, &request); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if len(request.Requests) != 1 {
				t.Fatalf("len(Requests) = %d, want 1", len(request.Requests))
			}
			update := request.Requests[0]
			if update.UpdateMask != "phases" {
				t.Fatalf("UpdateMask = %q, want phases", update.UpdateMask)
			}
			if update.RegionsVersion == nil || update.RegionsVersion.Version != "2026/05" {
				t.Fatalf("RegionsVersion = %#v, want 2026/05", update.RegionsVersion)
			}
			if update.LatencyTolerance != "PRODUCT_UPDATE_LATENCY_TOLERANCE_LATENCY_TOLERANT" {
				t.Fatalf("LatencyTolerance = %q, want tolerant", update.LatencyTolerance)
			}
			phases := update.SubscriptionOffer.Phases
			firstUS := phases[0].RegionalConfigs[0]
			if firstUS.RegionCode != "US" || firstUS.Free == nil || firstUS.Price != nil || firstUS.AbsoluteDiscount != nil || firstUS.RelativeDiscount != 0 {
				t.Fatalf("first US phase config = %#v, want free with old price modes cleared", firstUS)
			}
			firstFR := phases[0].RegionalConfigs[1]
			if firstFR.RegionCode != "FR" || firstFR.AbsoluteDiscount == nil || firstFR.AbsoluteDiscount.CurrencyCode != "EUR" {
				t.Fatalf("first FR phase config = %#v, want preserved absolute discount", firstFR)
			}
			if phases[0].OtherRegionsConfig == nil || phases[0].OtherRegionsConfig.AbsoluteDiscounts == nil || phases[0].OtherRegionsConfig.AbsoluteDiscounts.UsdPrice == nil {
				t.Fatalf("OtherRegionsConfig = %#v, want preserved other-regions absolute discounts", phases[0].OtherRegionsConfig)
			}
			secondUS := phases[1].RegionalConfigs[0]
			if secondUS.RegionCode != "US" || secondUS.RelativeDiscount != 0.75 || secondUS.Free != nil {
				t.Fatalf("second US phase config = %#v, want preserved relative discount", secondUS)
			}
			secondFR := phases[1].RegionalConfigs[1]
			if secondFR.RegionCode != "FR" || secondFR.Free == nil || secondFR.Price != nil || secondFR.AbsoluteDiscount != nil || secondFR.RelativeDiscount != 0 {
				t.Fatalf("second FR phase config = %#v, want patched free config", secondFR)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"subscriptionOffers":[{"packageName":"com.example.app","productId":"premium","basePlanId":"monthly","offerId":"intro","phases":[{"duration":"P1M","recurrenceCount":1,"regionalConfigs":[{"regionCode":"US","free":{}},{"regionCode":"FR","absoluteDiscount":{"currencyCode":"EUR","units":"2"}}]},{"duration":"P2M","recurrenceCount":2,"regionalConfigs":[{"regionCode":"US","relativeDiscount":0.75},{"regionCode":"FR","free":{}}]}]}]}`)
		default:
			t.Fatalf("method = %s, want GET or POST", r.Method)
		}
	}))

	result, err := publisher.BatchPatchSubscriptionOfferPhaseFree(context.Background(), SubscriptionOfferBatchPatchPhaseFreeOptions{
		PackageName:    "com.example.app",
		ProductID:      "premium",
		BasePlanID:     "monthly",
		RegionsVersion: "2026/05",
		Requests: []SubscriptionOfferPhaseFreePatchRequest{
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: 0, RegionCode: "US"},
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: 1, RegionCode: "FR"},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("BatchPatchSubscriptionOfferPhaseFree() error = %v", err)
	}
	if !result.Applied || len(result.Offers) != 1 || !result.Offers[0].Phases[0].RegionalConfigs[0].Free {
		t.Fatalf("result = %#v, want applied offer with free phase config", result)
	}
}

func TestBatchPatchSubscriptionOfferPhaseFreeRejectsMissingPhaseOrRegion(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions/premium/basePlans/monthly/offers/intro" {
			t.Fatalf("path = %q, want get only", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"packageName":"com.example.app","productId":"premium","basePlanId":"monthly","offerId":"intro","phases":[{"duration":"P1M","recurrenceCount":1,"regionalConfigs":[{"regionCode":"US","relativeDiscount":0.75}]}]}`)
	}))

	_, err := publisher.BatchPatchSubscriptionOfferPhaseFree(context.Background(), SubscriptionOfferBatchPatchPhaseFreeOptions{
		PackageName:    "com.example.app",
		ProductID:      "premium",
		BasePlanID:     "monthly",
		RegionsVersion: "2026/05",
		Requests: []SubscriptionOfferPhaseFreePatchRequest{
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro", PhaseIndex: 1, RegionCode: "FR"},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err == nil {
		t.Fatal("expected missing phase or region validation error")
	}
	if !strings.Contains(err.Error(), "phase or region that is not already configured") {
		t.Fatalf("error = %v, want missing phase or region validation", err)
	}
}

func TestCreateSubscriptionSendsTypedSubscriptionBody(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions" {
			t.Fatalf("path = %q, want subscription create endpoint", r.URL.Path)
		}
		if got := r.URL.Query().Get("productId"); got != "premium" {
			t.Fatalf("productId query = %q, want premium", got)
		}
		if got := r.URL.Query().Get("regionsVersion.version"); got != "2026/05" {
			t.Fatalf("regionsVersion.version query = %q, want 2026/05", got)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll() error = %v", err)
		}
		var request androidpublisher.Subscription
		if err := json.Unmarshal(body, &request); err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}
		if request.PackageName != "com.example.app" || request.ProductId != "premium" {
			t.Fatalf("request IDs = %#v, want create IDs", request)
		}
		if len(request.Listings) != 1 || request.Listings[0].Title != "Premium" {
			t.Fatalf("Listings = %#v, want listing", request.Listings)
		}
		if len(request.BasePlans) != 1 || request.BasePlans[0].AutoRenewingBasePlanType == nil || request.BasePlans[0].RegionalConfigs[0].Price == nil {
			t.Fatalf("BasePlans = %#v, want typed base plan", request.BasePlans)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"packageName":"com.example.app","productId":"premium","listings":[{"languageCode":"en-US","title":"Premium"}],"basePlans":[{"basePlanId":"monthly","state":"DRAFT","autoRenewingBasePlanType":{"billingPeriodDuration":"P1M"},"regionalConfigs":[{"regionCode":"US","newSubscriberAvailability":true,"price":{"currencyCode":"USD","units":"4","nanos":990000000}}]}]}`)
	}))

	result, err := publisher.CreateSubscription(context.Background(), SubscriptionCreateOptions{
		PackageName:    "com.example.app",
		ProductID:      "premium",
		Subscription:   validSubscriptionForCreate(),
		RegionsVersion: "2026/05",
		Confirm:        true,
	})
	if err != nil {
		t.Fatalf("CreateSubscription() error = %v", err)
	}
	if result.ProductID != "premium" || len(result.BasePlans) != 1 || result.BasePlans[0].State != SubscriptionStateDraft {
		t.Fatalf("result = %#v, want created draft subscription", result)
	}
}

func TestCreateSubscriptionOfferSendsTypedOfferBody(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions/premium/basePlans/monthly/offers" {
			t.Fatalf("path = %q, want subscription offer create endpoint", r.URL.Path)
		}
		if got := r.URL.Query().Get("offerId"); got != "intro" {
			t.Fatalf("offerId query = %q, want intro", got)
		}
		if got := r.URL.Query().Get("regionsVersion.version"); got != "2026/05" {
			t.Fatalf("regionsVersion.version query = %q, want 2026/05", got)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll() error = %v", err)
		}
		var request androidpublisher.SubscriptionOffer
		if err := json.Unmarshal(body, &request); err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}
		if request.PackageName != "com.example.app" || request.ProductId != "premium" || request.BasePlanId != "monthly" || request.OfferId != "intro" {
			t.Fatalf("request IDs = %#v, want create IDs", request)
		}
		if len(request.RegionalConfigs) != 2 || !request.RegionalConfigs[0].NewSubscriberAvailability {
			t.Fatalf("RegionalConfigs = %#v, want availability configs", request.RegionalConfigs)
		}
		if len(request.Phases) != 1 || len(request.Phases[0].RegionalConfigs) != 2 || request.Phases[0].RegionalConfigs[0].Price == nil {
			t.Fatalf("Phases = %#v, want typed phase prices", request.Phases)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"packageName":"com.example.app","productId":"premium","basePlanId":"monthly","offerId":"intro","state":"DRAFT","regionalConfigs":[{"regionCode":"US","newSubscriberAvailability":true},{"regionCode":"FR","newSubscriberAvailability":true}],"phases":[{"duration":"P1M","recurrenceCount":1,"regionalConfigs":[{"regionCode":"US","price":{"currencyCode":"USD","units":"1"}},{"regionCode":"FR","price":{"currencyCode":"EUR","nanos":990000000}}]}]}`)
	}))

	result, err := publisher.CreateSubscriptionOffer(context.Background(), SubscriptionOfferCreateOptions{
		PackageName:    "com.example.app",
		ProductID:      "premium",
		BasePlanID:     "monthly",
		OfferID:        "intro",
		Offer:          validSubscriptionOfferForCreate(),
		RegionsVersion: "2026/05",
		Confirm:        true,
	})
	if err != nil {
		t.Fatalf("CreateSubscriptionOffer() error = %v", err)
	}
	if result.State != SubscriptionOfferStateDraft || len(result.Phases) != 1 || result.Phases[0].RegionalConfigs[0].Price == nil {
		t.Fatalf("result = %#v, want created draft offer", result)
	}
}

func TestBatchUpdateSubscriptionOfferStatesRejectsDryRunBeforeRequest(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))

	_, err := publisher.BatchUpdateSubscriptionOfferStates(context.Background(), SubscriptionOfferBatchStateUpdateOptions{
		PackageName: "com.example.app",
		ProductID:   "premium",
		BasePlanID:  "monthly",
		Requests: []SubscriptionOfferBatchMutationRequest{
			{ProductID: "premium", BasePlanID: "monthly", OfferID: "intro"},
		},
		Action:           SubscriptionOfferStateActionActivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		DryRun:           true,
	})
	if err == nil {
		t.Fatal("expected live validation error")
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

func TestDeleteSubscriptionUsesMonetizationEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions/premium_monthly" {
			t.Fatalf("path = %q, want subscription delete endpoint", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	err := publisher.DeleteSubscription(context.Background(), SubscriptionDeleteOptions{
		PackageName: "com.example.app",
		ProductID:   "premium_monthly",
		Confirm:     true,
	})
	if err != nil {
		t.Fatalf("DeleteSubscription() error = %v", err)
	}
}

func TestDeleteSubscriptionRejectsDryRunBeforeRequest(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))

	err := publisher.DeleteSubscription(context.Background(), SubscriptionDeleteOptions{
		PackageName: "com.example.app",
		ProductID:   "premium_monthly",
		DryRun:      true,
	})
	if err == nil {
		t.Fatal("expected live validation error")
	}
}

func TestDeleteBasePlanUsesDeleteEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions/premium/basePlans/monthly" {
			t.Fatalf("path = %q, want base plan delete endpoint", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	err := publisher.DeleteBasePlan(context.Background(), BasePlanDeleteOptions{
		PackageName: "com.example.app",
		ProductID:   "premium",
		BasePlanID:  "monthly",
		Confirm:     true,
	})
	if err != nil {
		t.Fatalf("DeleteBasePlan() error = %v", err)
	}
}

func TestDeleteBasePlanRejectsDryRunBeforeRequest(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))

	err := publisher.DeleteBasePlan(context.Background(), BasePlanDeleteOptions{
		PackageName: "com.example.app",
		ProductID:   "premium",
		BasePlanID:  "monthly",
		DryRun:      true,
	})
	if err == nil {
		t.Fatal("expected live validation error")
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

func TestBatchUpdateBasePlanStatesUsesBatchUpdateEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions/premium/basePlans:batchUpdateStates" {
			t.Fatalf("path = %q, want base plans batchUpdateStates endpoint", r.URL.Path)
		}
		var request androidpublisher.BatchUpdateBasePlanStatesRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if len(request.Requests) != 2 {
			t.Fatalf("len(Requests) = %d, want 2", len(request.Requests))
		}
		deactivate := request.Requests[0].DeactivateBasePlanRequest
		if deactivate == nil || deactivate.ProductId != "premium" || deactivate.BasePlanId != "monthly" {
			t.Fatalf("first request = %#v, want deactivate monthly", request.Requests[0])
		}
		if deactivate.LatencyTolerance != "PRODUCT_UPDATE_LATENCY_TOLERANCE_LATENCY_TOLERANT" {
			t.Fatalf("LatencyTolerance = %q, want tolerant", deactivate.LatencyTolerance)
		}
		if request.Requests[1].DeactivateBasePlanRequest == nil || request.Requests[1].DeactivateBasePlanRequest.BasePlanId != "annual" {
			t.Fatalf("second request = %#v, want deactivate annual", request.Requests[1])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"subscriptions":[
			{"packageName":"com.example.app","productId":"premium","basePlans":[{"basePlanId":"monthly","state":"INACTIVE"}]},
			{"packageName":"com.example.app","productId":"premium","basePlans":[{"basePlanId":"annual","state":"INACTIVE"}]}
		]}`)
	}))

	result, err := publisher.BatchUpdateBasePlanStates(context.Background(), BasePlanBatchStateUpdateOptions{
		PackageName: "com.example.app",
		ProductID:   "premium",
		Requests: []BasePlanBatchStateUpdateRequest{
			{ProductID: "premium", BasePlanID: "monthly"},
			{ProductID: "premium", BasePlanID: "annual"},
		},
		Action:           BasePlanStateActionDeactivate,
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("BatchUpdateBasePlanStates() error = %v", err)
	}
	if len(result.Subscriptions) != 2 {
		t.Fatalf("len(Subscriptions) = %d, want 2", len(result.Subscriptions))
	}
	if result.Subscriptions[0].BasePlans[0].BasePlanID != "monthly" || result.Subscriptions[1].BasePlans[0].BasePlanID != "annual" {
		t.Fatalf("Subscriptions = %#v, want response order", result.Subscriptions)
	}
}

func TestBatchUpdateBasePlanStatesUsesActivateUnionForWildcardProduct(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions/-/basePlans:batchUpdateStates" {
			t.Fatalf("path = %q, want wildcard base plans batchUpdateStates endpoint", r.URL.Path)
		}
		var request androidpublisher.BatchUpdateBasePlanStatesRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if len(request.Requests) != 2 {
			t.Fatalf("len(Requests) = %d, want 2", len(request.Requests))
		}
		activate := request.Requests[0].ActivateBasePlanRequest
		if activate == nil || request.Requests[0].DeactivateBasePlanRequest != nil {
			t.Fatalf("first request = %#v, want activate union only", request.Requests[0])
		}
		if activate.PackageName != "com.example.app" || activate.ProductId != "premium" || activate.BasePlanId != "monthly" {
			t.Fatalf("activate = %#v, want first request identifiers", activate)
		}
		second := request.Requests[1].ActivateBasePlanRequest
		if second == nil || second.ProductId != "vip" || second.BasePlanId != "annual" {
			t.Fatalf("second request = %#v, want activate vip/annual", request.Requests[1])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"subscriptions":[]}`)
	}))

	_, err := publisher.BatchUpdateBasePlanStates(context.Background(), BasePlanBatchStateUpdateOptions{
		PackageName: "com.example.app",
		ProductID:   "-",
		Requests: []BasePlanBatchStateUpdateRequest{
			{ProductID: "premium", BasePlanID: "monthly"},
			{ProductID: "vip", BasePlanID: "annual"},
		},
		Action:           BasePlanStateActionActivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("BatchUpdateBasePlanStates() error = %v", err)
	}
}

func TestBatchUpdateBasePlanStatesRejectsDryRunBeforeRequest(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))

	_, err := publisher.BatchUpdateBasePlanStates(context.Background(), BasePlanBatchStateUpdateOptions{
		PackageName: "com.example.app",
		ProductID:   "premium",
		Requests: []BasePlanBatchStateUpdateRequest{
			{ProductID: "premium", BasePlanID: "monthly"},
		},
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

func TestBatchMigrateBasePlanPricesUsesBatchMigrateEndpoint(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions/-/basePlans:batchMigratePrices" {
			t.Fatalf("path = %q, want base plan batchMigratePrices endpoint", r.URL.Path)
		}
		var request androidpublisher.BatchMigrateBasePlanPricesRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if len(request.Requests) != 2 {
			t.Fatalf("len(Requests) = %d, want 2", len(request.Requests))
		}
		first := request.Requests[0]
		if first.PackageName != "com.example.app" || first.ProductId != "premium" || first.BasePlanId != "monthly" {
			t.Fatalf("first request = %#v, want premium/monthly", first)
		}
		if first.RegionsVersion == nil || first.RegionsVersion.Version != "2026/05" {
			t.Fatalf("RegionsVersion = %#v, want 2026/05", first.RegionsVersion)
		}
		if first.LatencyTolerance != "PRODUCT_UPDATE_LATENCY_TOLERANCE_LATENCY_TOLERANT" {
			t.Fatalf("LatencyTolerance = %q, want tolerant", first.LatencyTolerance)
		}
		if len(first.RegionalPriceMigrations) != 2 {
			t.Fatalf("len(RegionalPriceMigrations) = %d, want 2", len(first.RegionalPriceMigrations))
		}
		if first.RegionalPriceMigrations[0].RegionCode != "US" || first.RegionalPriceMigrations[0].PriceIncreaseType != "PRICE_INCREASE_TYPE_OPT_OUT" {
			t.Fatalf("first regional migration = %#v, want US opt-out", first.RegionalPriceMigrations[0])
		}
		if request.Requests[1].ProductId != "vip" || request.Requests[1].BasePlanId != "annual" {
			t.Fatalf("second request = %#v, want vip/annual", request.Requests[1])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"responses":[{},{}]}`)
	}))

	result, err := publisher.BatchMigrateBasePlanPrices(context.Background(), BasePlanBatchPriceMigrationOptions{
		PackageName:    "com.example.app",
		ProductID:      "-",
		RegionsVersion: "2026/05",
		Requests: []BasePlanPriceMigrationRequest{
			{
				ProductID:  "premium",
				BasePlanID: "monthly",
				Regions: []BasePlanPriceMigrationConfig{
					{RegionCode: "US", OldestAllowedPriceVersionTime: "2026-05-01T00:00:00Z", PriceIncreaseType: BasePlanPriceIncreaseTypeOptOut},
					{RegionCode: "BR", OldestAllowedPriceVersionTime: "2026-05-01T00:00:00Z", PriceIncreaseType: BasePlanPriceIncreaseTypeOptOut},
				},
			},
			{
				ProductID:  "vip",
				BasePlanID: "annual",
				Regions: []BasePlanPriceMigrationConfig{
					{RegionCode: "US", OldestAllowedPriceVersionTime: "2026-05-01T00:00:00Z", PriceIncreaseType: BasePlanPriceIncreaseTypeOptOut},
				},
			},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("BatchMigrateBasePlanPrices() error = %v", err)
	}
	if len(result.Responses) != 2 {
		t.Fatalf("len(Responses) = %d, want 2", len(result.Responses))
	}
	if result.Responses[0].ProductID != "premium" || result.Responses[1].ProductID != "vip" {
		t.Fatalf("Responses = %#v, want request identities", result.Responses)
	}
}

func TestBatchMigrateBasePlanPricesRejectsDryRunBeforeRequest(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))

	_, err := publisher.BatchMigrateBasePlanPrices(context.Background(), BasePlanBatchPriceMigrationOptions{
		PackageName:    "com.example.app",
		ProductID:      "premium",
		RegionsVersion: "2026/05",
		Requests: []BasePlanPriceMigrationRequest{
			{
				ProductID:  "premium",
				BasePlanID: "monthly",
				Regions: []BasePlanPriceMigrationConfig{
					{RegionCode: "US", OldestAllowedPriceVersionTime: "2026-05-01T00:00:00Z"},
				},
			},
		},
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

func TestBatchPatchBasePlanPricesFetchesMergesAndBatchUpdates(t *testing.T) {
	var seenGet bool
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/androidpublisher/v3/applications/com.example.app/subscriptions/premium":
			seenGet = true
			_, _ = io.WriteString(w, `{"packageName":"com.example.app","productId":"premium","listings":[{"languageCode":"en-US","title":"Premium"}],"basePlans":[{"basePlanId":"monthly","state":"ACTIVE","autoRenewingBasePlanType":{"billingPeriodDuration":"P1M"},"regionalConfigs":[{"regionCode":"US","newSubscriberAvailability":true,"price":{"currencyCode":"USD","units":"2"}},{"regionCode":"BR","newSubscriberAvailability":true,"price":{"currencyCode":"BRL","units":"9"}}]},{"basePlanId":"annual","state":"ACTIVE","autoRenewingBasePlanType":{"billingPeriodDuration":"P1Y"},"regionalConfigs":[{"regionCode":"US","newSubscriberAvailability":true,"price":{"currencyCode":"USD","units":"20"}}]}]}`)
		case "/androidpublisher/v3/applications/com.example.app/subscriptions:batchUpdate":
			if !seenGet {
				t.Fatal("batch update happened before get")
			}
			var request androidpublisher.BatchUpdateSubscriptionsRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if len(request.Requests) != 1 {
				t.Fatalf("len(Requests) = %d, want 1", len(request.Requests))
			}
			update := request.Requests[0]
			if update.UpdateMask != subscriptionBasePlanPatchUpdateMask {
				t.Fatalf("UpdateMask = %q, want %q", update.UpdateMask, subscriptionBasePlanPatchUpdateMask)
			}
			if update.RegionsVersion == nil || update.RegionsVersion.Version != "2026/05" {
				t.Fatalf("RegionsVersion = %#v, want 2026/05", update.RegionsVersion)
			}
			if update.LatencyTolerance != "PRODUCT_UPDATE_LATENCY_TOLERANCE_LATENCY_TOLERANT" {
				t.Fatalf("LatencyTolerance = %q, want tolerant", update.LatencyTolerance)
			}
			if update.Subscription == nil || len(update.Subscription.BasePlans) != 2 {
				t.Fatalf("Subscription = %#v, want preserved base plans", update.Subscription)
			}
			monthly := update.Subscription.BasePlans[0]
			if monthly.BasePlanId != "monthly" || len(monthly.RegionalConfigs) != 2 {
				t.Fatalf("monthly = %#v, want two regions", monthly)
			}
			if monthly.RegionalConfigs[0].Price == nil || monthly.RegionalConfigs[0].Price.Units != 4 || monthly.RegionalConfigs[0].Price.Nanos != 990000000 {
				t.Fatalf("US price = %#v, want patched 4.99", monthly.RegionalConfigs[0].Price)
			}
			if monthly.RegionalConfigs[1].Price == nil || monthly.RegionalConfigs[1].Price.Units != 9 {
				t.Fatalf("BR price = %#v, want preserved 9", monthly.RegionalConfigs[1].Price)
			}
			_, _ = io.WriteString(w, `{"subscriptions":[{"packageName":"com.example.app","productId":"premium","basePlans":[{"basePlanId":"monthly","state":"ACTIVE","autoRenewingBasePlanType":{"billingPeriodDuration":"P1M"},"regionalConfigs":[{"regionCode":"US","newSubscriberAvailability":true,"price":{"currencyCode":"USD","units":"4","nanos":990000000}}]}]}]}`)
		default:
			t.Fatalf("unexpected path = %q", r.URL.Path)
		}
	}))

	result, err := publisher.BatchPatchBasePlanPrices(context.Background(), BasePlanBatchPatchPriceOptions{
		PackageName:    "com.example.app",
		ProductID:      "premium",
		RegionsVersion: "2026/05",
		Requests: []BasePlanPricePatchRequest{
			{ProductID: "premium", BasePlanID: "monthly", RegionCode: "US", Price: Money{CurrencyCode: "USD", Units: 4, Nanos: 990000000}},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("BatchPatchBasePlanPrices() error = %v", err)
	}
	if !result.Applied || len(result.Subscriptions) != 1 || result.Subscriptions[0].BasePlans[0].RegionalConfigs[0].Price.Units != 4 {
		t.Fatalf("result = %#v, want applied subscription with patched price", result)
	}
}

func TestBatchPatchBasePlanPricesRejectsMissingRegion(t *testing.T) {
	publisher := newTestPublisher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/androidpublisher/v3/applications/com.example.app/subscriptions/premium" {
			t.Fatalf("path = %q, want subscription get only", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"packageName":"com.example.app","productId":"premium","basePlans":[{"basePlanId":"monthly","autoRenewingBasePlanType":{"billingPeriodDuration":"P1M"},"regionalConfigs":[{"regionCode":"US","newSubscriberAvailability":true,"price":{"currencyCode":"USD","units":"2"}}]}]}`)
	}))

	_, err := publisher.BatchPatchBasePlanPrices(context.Background(), BasePlanBatchPatchPriceOptions{
		PackageName:    "com.example.app",
		ProductID:      "premium",
		RegionsVersion: "2026/05",
		Requests: []BasePlanPricePatchRequest{
			{ProductID: "premium", BasePlanID: "monthly", RegionCode: "FR", Price: Money{CurrencyCode: "EUR", Units: 2}},
		},
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err == nil {
		t.Fatal("expected missing region validation error")
	}
	if !strings.Contains(err.Error(), "not already configured") {
		t.Fatalf("error = %v, want configured region message", err)
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
