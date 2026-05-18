package play

import (
	"context"
	"reflect"
	"testing"
)

func TestListListingsUsesTemporaryEdit(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	reader := &fakeListingClient{
		listings: []Listing{{Language: "en-US", Title: "Example"}},
	}

	listings, err := ListListings(context.Background(), reader, packageName)
	if err != nil {
		t.Fatalf("ListListings() error = %v", err)
	}
	if len(listings) != 1 {
		t.Fatalf("len(listings) = %d, want 1", len(listings))
	}
	wantCalls := []string{"insert", "list", "delete"}
	if !reflect.DeepEqual(reader.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", reader.calls, wantCalls)
	}
}

func TestGetListingUsesTemporaryEdit(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	reader := &fakeListingClient{
		listing: Listing{Language: "en-US", Title: "Example"},
	}

	listing, err := GetListing(context.Background(), reader, packageName, "en-US")
	if err != nil {
		t.Fatalf("GetListing() error = %v", err)
	}
	if listing.Title != "Example" {
		t.Fatalf("Title = %q, want Example", listing.Title)
	}
	wantCalls := []string{"insert", "get", "delete"}
	if !reflect.DeepEqual(reader.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", reader.calls, wantCalls)
	}
}

func TestUpdateListingDryRunDoesNotRequireUpdater(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := UpdateListing(context.Background(), nil, UpdateListingOptions{
		PackageName: packageName,
		Listing:     Listing{Language: "en-US", Title: "Example"},
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("UpdateListing() error = %v", err)
	}
	if !result.DryRun {
		t.Fatal("DryRun = false, want true")
	}
}

func TestUpdateListingRequiresAField(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = NewUpdateListingPlan(UpdateListingOptions{
		PackageName: packageName,
		Listing:     Listing{Language: "en-US"},
	})
	if err == nil {
		t.Fatal("expected field validation error")
	}
}

func TestUpdateListingValidatesAndCleansUpWithoutConfirm(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	updater := &fakeListingClient{}

	result, err := UpdateListing(context.Background(), updater, UpdateListingOptions{
		PackageName: packageName,
		Listing:     Listing{Language: "en-US", Title: "Example"},
	})
	if err != nil {
		t.Fatalf("UpdateListing() error = %v", err)
	}
	if result.Committed {
		t.Fatal("Committed = true, want false")
	}
	wantCalls := []string{"insert", "update", "validate", "delete"}
	if !reflect.DeepEqual(updater.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", updater.calls, wantCalls)
	}
}

func TestUpdateListingCommitsWithConfirm(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	updater := &fakeListingClient{}

	result, err := UpdateListing(context.Background(), updater, UpdateListingOptions{
		PackageName: packageName,
		Listing:     Listing{Language: "en-US", Title: "Example"},
		Confirm:     true,
	})
	if err != nil {
		t.Fatalf("UpdateListing() error = %v", err)
	}
	if !result.Committed {
		t.Fatal("Committed = false, want true")
	}
	wantCalls := []string{"insert", "update", "validate", "commit"}
	if !reflect.DeepEqual(updater.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", updater.calls, wantCalls)
	}
}

type fakeListingClient struct {
	calls    []string
	listing  Listing
	listings []Listing
}

func (c *fakeListingClient) InsertEdit(ctx context.Context, packageName PackageName) (Edit, error) {
	c.calls = append(c.calls, "insert")
	return Edit{ID: "edit-123"}, nil
}

func (c *fakeListingClient) ListListings(ctx context.Context, packageName PackageName, editID string) ([]Listing, error) {
	c.calls = append(c.calls, "list")
	return c.listings, nil
}

func (c *fakeListingClient) GetListing(ctx context.Context, packageName PackageName, editID string, language ListingLanguage) (Listing, error) {
	c.calls = append(c.calls, "get")
	return c.listing, nil
}

func (c *fakeListingClient) UpdateListing(ctx context.Context, packageName PackageName, editID string, listing Listing) (Listing, error) {
	c.calls = append(c.calls, "update")
	return listing, nil
}

func (c *fakeListingClient) ValidateEdit(ctx context.Context, packageName PackageName, editID string) error {
	c.calls = append(c.calls, "validate")
	return nil
}

func (c *fakeListingClient) CommitEdit(ctx context.Context, packageName PackageName, editID string) (Edit, error) {
	c.calls = append(c.calls, "commit")
	return Edit{ID: editID}, nil
}

func (c *fakeListingClient) DeleteEdit(ctx context.Context, packageName PackageName, editID string) error {
	c.calls = append(c.calls, "delete")
	return nil
}
