package play

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestParseMetadataFileRejectsUnknownFields(t *testing.T) {
	_, err := ParseMetadataFile("metadata.json", []byte(`{"details":{"contactEmail":"support@example.com"},"unexpected":true}`))
	if err == nil {
		t.Fatal("ParseMetadataFile() error = nil, want unknown field error")
	}
	if !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("error = %v, want unknown field", err)
	}
}

func TestParseMetadataFileRejectsTrailingJSON(t *testing.T) {
	_, err := ParseMetadataFile("metadata.json", []byte(`{"details":{"contactEmail":"support@example.com"}} {}`))
	if err == nil {
		t.Fatal("ParseMetadataFile() error = nil, want trailing JSON error")
	}
	if !strings.Contains(err.Error(), "trailing JSON value") {
		t.Fatalf("error = %v, want trailing JSON", err)
	}
}

func TestMetadataApplyDryRunSortsListingsAndDoesNotRequireApplier(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := ApplyMetadata(context.Background(), nil, MetadataApplyOptions{
		PackageName: packageName,
		FilePath:    "metadata.json",
		Listings: []Listing{
			{Language: "es-ES", Title: stringValue("Ejemplo")},
			{Language: "en-US", Title: stringValue("Example")},
		},
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("ApplyMetadata() error = %v", err)
	}
	if !result.DryRun {
		t.Fatal("DryRun = false, want true")
	}
	wantSteps := []string{"insert edit", "patch en-US listing", "patch es-ES listing", "validate edit", "delete uncommitted edit"}
	if !reflect.DeepEqual(result.Plan.Steps, wantSteps) {
		t.Fatalf("steps = %#v, want %#v", result.Plan.Steps, wantSteps)
	}
}

func TestMetadataApplyNormalizesLanguageTags(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := ApplyMetadata(context.Background(), nil, MetadataApplyOptions{
		PackageName: packageName,
		FilePath:    "metadata.json",
		Details:     &AppDetails{DefaultLanguage: stringValue("en-us")},
		Listings:    []Listing{{Language: "es-es", Title: stringValue("Ejemplo")}},
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("ApplyMetadata() error = %v", err)
	}
	if result.Details == nil || result.Details.DefaultLanguage == nil || *result.Details.DefaultLanguage != "en-US" {
		t.Fatalf("DefaultLanguage = %#v, want en-US", result.Details)
	}
	if len(result.Listings) != 1 || result.Listings[0].Language != "es-ES" {
		t.Fatalf("Listings = %#v, want es-ES", result.Listings)
	}
}

func TestMetadataApplyRejectsInvalidDetails(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	tests := []struct {
		name    string
		details AppDetails
		want    string
	}{
		{name: "language", details: AppDetails{DefaultLanguage: stringValue("zz-ZZ")}, want: "invalid defaultLanguage"},
		{name: "website", details: AppDetails{ContactWebsite: stringValue("ftp://example.com")}, want: "contactWebsite must use http or https"},
		{name: "email", details: AppDetails{ContactEmail: stringValue("not-email")}, want: "contactEmail must be a valid email address"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewMetadataApplyPlan(MetadataApplyOptions{
				PackageName: packageName,
				FilePath:    "metadata.json",
				Details:     &tt.details,
				DryRun:      true,
			})
			if err == nil {
				t.Fatal("NewMetadataApplyPlan() error = nil, want validation error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %s", err, tt.want)
			}
		})
	}
}

func TestMetadataApplyRejectsInvalidListingValues(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	tests := []struct {
		name    string
		listing Listing
		want    string
	}{
		{name: "empty title", listing: Listing{Language: "en-US", Title: stringValue("")}, want: "title must be non-empty"},
		{name: "long title", listing: Listing{Language: "en-US", Title: stringValue(strings.Repeat("a", listingTitleMaxRunes+1))}, want: "title must be at most 30 characters"},
		{name: "long short description", listing: Listing{Language: "en-US", ShortDescription: stringValue(strings.Repeat("a", listingShortDescriptionMaxRunes+1))}, want: "shortDescription must be at most 80 characters"},
		{name: "long full description", listing: Listing{Language: "en-US", FullDescription: stringValue(strings.Repeat("a", listingFullDescriptionMaxRunes+1))}, want: "fullDescription must be at most 4000 characters"},
		{name: "bad video scheme", listing: Listing{Language: "en-US", Video: stringValue("ftp://youtube.com/watch?v=abc")}, want: "video must use http or https"},
		{name: "bad video host", listing: Listing{Language: "en-US", Video: stringValue("https://example.com/watch?v=abc")}, want: "video must use a youtube.com or youtu.be URL"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewMetadataApplyPlan(MetadataApplyOptions{
				PackageName: packageName,
				FilePath:    "metadata.json",
				Listings:    []Listing{tt.listing},
				DryRun:      true,
			})
			if err == nil {
				t.Fatal("NewMetadataApplyPlan() error = nil, want validation error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %s", err, tt.want)
			}
		})
	}
}

func TestMetadataApplyUsesSingleEditAndCommitsWhenConfirmed(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	applier := &fakeMetadataApplier{}

	result, err := ApplyMetadata(context.Background(), applier, MetadataApplyOptions{
		PackageName: packageName,
		FilePath:    "metadata.json",
		Details:     &AppDetails{ContactEmail: stringValue("support@example.com")},
		Listings: []Listing{
			{Language: "es-ES", Title: stringValue("Ejemplo")},
			{Language: "en-US", Title: stringValue("Example")},
		},
		Confirm: true,
	})
	if err != nil {
		t.Fatalf("ApplyMetadata() error = %v", err)
	}
	if !result.Committed {
		t.Fatal("Committed = false, want true")
	}
	wantCalls := []string{"insert", "patch-details", "patch-listing:en-US", "patch-listing:es-ES", "validate", "commit"}
	if !reflect.DeepEqual(applier.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", applier.calls, wantCalls)
	}
}

func TestMetadataApplyCleansUpWhenListingPatchFails(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	applier := &fakeMetadataApplier{patchListingErr: errors.New("patch failed")}

	_, err = ApplyMetadata(context.Background(), applier, MetadataApplyOptions{
		PackageName: packageName,
		FilePath:    "metadata.json",
		Listings:    []Listing{{Language: "en-US", Title: stringValue("Example")}},
		Confirm:     true,
	})
	if err == nil {
		t.Fatal("ApplyMetadata() error = nil, want patch error")
	}
	if !strings.Contains(err.Error(), "patch failed") {
		t.Fatalf("error = %v, want patch failure", err)
	}
	wantCalls := []string{"insert", "patch-listing:en-US", "delete"}
	if !reflect.DeepEqual(applier.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", applier.calls, wantCalls)
	}
}

func TestMetadataApplyRejectsDuplicateListingLanguages(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = NewMetadataApplyPlan(MetadataApplyOptions{
		PackageName: packageName,
		FilePath:    "metadata.json",
		Listings: []Listing{
			{Language: "en-US", Title: stringValue("Example")},
			{Language: "en-us", ShortDescription: stringValue("Duplicate")},
		},
		DryRun: true,
	})
	if err == nil {
		t.Fatal("NewMetadataApplyPlan() error = nil, want duplicate language error")
	}
	if !strings.Contains(err.Error(), "duplicate listing language en-US") {
		t.Fatalf("error = %v, want duplicate language", err)
	}
}

func TestMetadataApplyRequiresConfirmOrDryRun(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = NewMetadataApplyPlan(MetadataApplyOptions{
		PackageName: packageName,
		FilePath:    "metadata.json",
		Details:     &AppDetails{ContactEmail: stringValue("support@example.com")},
	})
	if err == nil {
		t.Fatal("NewMetadataApplyPlan() error = nil, want confirmation gate error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
}

type fakeMetadataApplier struct {
	calls           []string
	patchListingErr error
}

func (a *fakeMetadataApplier) InsertEdit(ctx context.Context, packageName PackageName) (Edit, error) {
	a.calls = append(a.calls, "insert")
	return Edit{ID: "edit-123"}, nil
}

func (a *fakeMetadataApplier) PatchAppDetails(ctx context.Context, packageName PackageName, editID string, details AppDetails) (AppDetails, error) {
	a.calls = append(a.calls, "patch-details")
	return details, nil
}

func (a *fakeMetadataApplier) PatchListing(ctx context.Context, packageName PackageName, editID string, listing Listing) (Listing, error) {
	a.calls = append(a.calls, "patch-listing:"+listing.Language.String())
	if a.patchListingErr != nil {
		return Listing{}, a.patchListingErr
	}
	return listing, nil
}

func (a *fakeMetadataApplier) ValidateEdit(ctx context.Context, packageName PackageName, editID string) error {
	a.calls = append(a.calls, "validate")
	return nil
}

func (a *fakeMetadataApplier) CommitEdit(ctx context.Context, packageName PackageName, editID string) (Edit, error) {
	a.calls = append(a.calls, "commit")
	return Edit{ID: editID}, nil
}

func (a *fakeMetadataApplier) DeleteEdit(ctx context.Context, packageName PackageName, editID string) error {
	a.calls = append(a.calls, "delete")
	return nil
}
