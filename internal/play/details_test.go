package play

import (
	"context"
	"reflect"
	"testing"
)

func TestGetAppDetailsUsesTemporaryEdit(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	reader := &fakeDetailsClient{details: AppDetails{DefaultLanguage: stringValue("en-US")}}

	details, err := GetAppDetails(context.Background(), reader, packageName)
	if err != nil {
		t.Fatalf("GetAppDetails() error = %v", err)
	}
	if details.DefaultLanguage == nil || *details.DefaultLanguage != "en-US" {
		t.Fatalf("DefaultLanguage = %v, want en-US", details.DefaultLanguage)
	}
	wantCalls := []string{"insert", "get", "delete"}
	if !reflect.DeepEqual(reader.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", reader.calls, wantCalls)
	}
}

func TestUpdateAppDetailsDryRunDoesNotRequireUpdater(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := UpdateAppDetails(context.Background(), nil, UpdateDetailsOptions{
		PackageName: packageName,
		Details:     AppDetails{ContactEmail: stringValue("support@example.com")},
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("UpdateAppDetails() error = %v", err)
	}
	if !result.DryRun {
		t.Fatal("DryRun = false, want true")
	}
}

func TestUpdateAppDetailsRequiresAField(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = NewUpdateDetailsPlan(UpdateDetailsOptions{PackageName: packageName})
	if err == nil {
		t.Fatal("expected field validation error")
	}
}

func TestUpdateAppDetailsValidatesAndCleansUpWithoutConfirm(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	updater := &fakeDetailsClient{}

	result, err := UpdateAppDetails(context.Background(), updater, UpdateDetailsOptions{
		PackageName: packageName,
		Details:     AppDetails{ContactEmail: stringValue("support@example.com")},
	})
	if err != nil {
		t.Fatalf("UpdateAppDetails() error = %v", err)
	}
	if result.Committed {
		t.Fatal("Committed = true, want false")
	}
	wantCalls := []string{"insert", "patch", "validate", "delete"}
	if !reflect.DeepEqual(updater.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", updater.calls, wantCalls)
	}
}

func TestUpdateAppDetailsCommitsWithConfirm(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	updater := &fakeDetailsClient{}

	result, err := UpdateAppDetails(context.Background(), updater, UpdateDetailsOptions{
		PackageName: packageName,
		Details:     AppDetails{ContactEmail: stringValue("support@example.com")},
		Confirm:     true,
	})
	if err != nil {
		t.Fatalf("UpdateAppDetails() error = %v", err)
	}
	if !result.Committed {
		t.Fatal("Committed = false, want true")
	}
	wantCalls := []string{"insert", "patch", "validate", "commit"}
	if !reflect.DeepEqual(updater.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", updater.calls, wantCalls)
	}
}

type fakeDetailsClient struct {
	calls   []string
	details AppDetails
}

func (c *fakeDetailsClient) InsertEdit(ctx context.Context, packageName PackageName) (Edit, error) {
	c.calls = append(c.calls, "insert")
	return Edit{ID: "edit-123"}, nil
}

func (c *fakeDetailsClient) GetAppDetails(ctx context.Context, packageName PackageName, editID string) (AppDetails, error) {
	c.calls = append(c.calls, "get")
	return c.details, nil
}

func (c *fakeDetailsClient) PatchAppDetails(ctx context.Context, packageName PackageName, editID string, details AppDetails) (AppDetails, error) {
	c.calls = append(c.calls, "patch")
	return details, nil
}

func (c *fakeDetailsClient) ValidateEdit(ctx context.Context, packageName PackageName, editID string) error {
	c.calls = append(c.calls, "validate")
	return nil
}

func (c *fakeDetailsClient) CommitEdit(ctx context.Context, packageName PackageName, editID string) (Edit, error) {
	c.calls = append(c.calls, "commit")
	return Edit{ID: editID}, nil
}

func (c *fakeDetailsClient) DeleteEdit(ctx context.Context, packageName PackageName, editID string) error {
	c.calls = append(c.calls, "delete")
	return nil
}
