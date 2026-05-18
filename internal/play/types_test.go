package play

import "testing"

func TestNewPackageNameValidatesShape(t *testing.T) {
	validValues := []string{
		"com.example.app",
		"io.gpc.cli_1",
	}
	for _, value := range validValues {
		if _, err := NewPackageName(value); err != nil {
			t.Fatalf("NewPackageName(%q) error = %v", value, err)
		}
	}

	invalidValues := []string{
		"bad",
		"com.",
		"1com.example",
		"com.example-app",
	}
	for _, value := range invalidValues {
		if _, err := NewPackageName(value); err == nil {
			t.Fatalf("NewPackageName(%q) expected error", value)
		}
	}
}

func TestPublishInternalOptionsRequiresUserFractionForStagedRollout(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	options := PublishInternalOptions{
		PackageName: packageName,
		BundlePath:  "app-release.aab",
		Status:      ReleaseStatusInProgress,
	}
	if err := options.Validate(); err == nil {
		t.Fatal("expected missing user fraction error")
	}

	userFraction := 0.25
	options.UserFraction = &userFraction
	if err := options.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestPublishInternalOptionsRejectsInvalidTypedPackageName(t *testing.T) {
	options := PublishInternalOptions{
		PackageName: PackageName("bad"),
		BundlePath:  "app-release.aab",
		Status:      ReleaseStatusCompleted,
	}
	if err := options.Validate(); err == nil {
		t.Fatal("expected package validation error")
	}
}

func TestPublishInternalOptionsRejectsUserFractionForCompletedRelease(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	userFraction := 0.25
	options := PublishInternalOptions{
		PackageName:  packageName,
		BundlePath:   "app-release.aab",
		Status:       ReleaseStatusCompleted,
		UserFraction: &userFraction,
	}
	if err := options.Validate(); err == nil {
		t.Fatal("expected user fraction error")
	}
}

func TestPublishInternalOptionsRejectsEmptyReleaseNoteText(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	options := PublishInternalOptions{
		PackageName: packageName,
		BundlePath:  "app-release.aab",
		Status:      ReleaseStatusCompleted,
		ReleaseNotes: []ReleaseNote{
			{Language: "en-US"},
		},
	}
	if err := options.Validate(); err == nil {
		t.Fatal("expected release note text error")
	}
}
