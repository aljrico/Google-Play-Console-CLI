package play

import (
	"context"
	"reflect"
	"testing"
)

func TestValidateInsertsValidatesAndDeletesEdit(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	validator := &fakeEditValidator{}

	result, err := Validate(context.Background(), validator, ValidateOptions{PackageName: packageName})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if !result.Valid {
		t.Fatal("Valid = false, want true")
	}
	if !result.Deleted {
		t.Fatal("Deleted = false, want true")
	}
	if result.Edit.ID != "edit-123" {
		t.Fatalf("Edit.ID = %q, want edit-123", result.Edit.ID)
	}
	wantCalls := []string{"insert", "validate", "delete"}
	if !reflect.DeepEqual(validator.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", validator.calls, wantCalls)
	}
}

func TestValidateRejectsInvalidPackage(t *testing.T) {
	_, err := Validate(context.Background(), nil, ValidateOptions{PackageName: "bad"})
	if err == nil {
		t.Fatal("expected package validation error")
	}
}

type fakeEditValidator struct {
	calls []string
}

func (v *fakeEditValidator) InsertEdit(ctx context.Context, packageName PackageName) (Edit, error) {
	v.calls = append(v.calls, "insert")
	return Edit{ID: "edit-123"}, nil
}

func (v *fakeEditValidator) ValidateEdit(ctx context.Context, packageName PackageName, editID string) error {
	v.calls = append(v.calls, "validate")
	return nil
}

func (v *fakeEditValidator) DeleteEdit(ctx context.Context, packageName PackageName, editID string) error {
	v.calls = append(v.calls, "delete")
	return nil
}
