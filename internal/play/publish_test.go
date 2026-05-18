package play

import (
	"context"
	"reflect"
	"testing"
)

func TestNewPublishInternalPlan(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	plan, err := NewPublishInternalPlan(PublishInternalOptions{
		PackageName: packageName,
		BundlePath:  "app-release.aab",
		ReleaseName: "1.2.3",
		Status:      ReleaseStatusCompleted,
		Confirm:     true,
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("NewPublishInternalPlan() error = %v", err)
	}
	if plan.Track != TrackInternal {
		t.Fatalf("Track = %q, want %q", plan.Track, TrackInternal)
	}
	if plan.Steps[len(plan.Steps)-1] != "commit edit" {
		t.Fatalf("last step = %q, want commit edit", plan.Steps[len(plan.Steps)-1])
	}
}

func TestPublishInternalDryRunDoesNotRequirePublisher(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := PublishInternal(context.Background(), nil, PublishInternalOptions{
		PackageName: packageName,
		BundlePath:  "app-release.aab",
		Status:      ReleaseStatusCompleted,
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("PublishInternal() error = %v", err)
	}
	if !result.DryRun {
		t.Fatal("DryRun = false, want true")
	}
	if result.Committed {
		t.Fatal("Committed = true, want false")
	}
}

func TestPublishInternalRejectsNonBundlePath(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = NewPublishInternalPlan(PublishInternalOptions{
		PackageName: packageName,
		BundlePath:  "app-release.apk",
		Status:      ReleaseStatusCompleted,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPublishInternalValidatesAndCleansUpWithoutConfirm(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	publisher := &fakePublisher{}

	result, err := PublishInternal(context.Background(), publisher, PublishInternalOptions{
		PackageName: packageName,
		BundlePath:  "app-release.aab",
		Status:      ReleaseStatusCompleted,
	})
	if err != nil {
		t.Fatalf("PublishInternal() error = %v", err)
	}
	if result.Committed {
		t.Fatal("Committed = true, want false")
	}

	wantCalls := []string{"insert", "upload-bundle", "update-track", "validate", "delete"}
	if !reflect.DeepEqual(publisher.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", publisher.calls, wantCalls)
	}
}

func TestPublishInternalCommitsWithConfirm(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	publisher := &fakePublisher{}

	result, err := PublishInternal(context.Background(), publisher, PublishInternalOptions{
		PackageName: packageName,
		BundlePath:  "app-release.aab",
		Status:      ReleaseStatusCompleted,
		Confirm:     true,
	})
	if err != nil {
		t.Fatalf("PublishInternal() error = %v", err)
	}
	if !result.Committed {
		t.Fatal("Committed = false, want true")
	}

	wantCalls := []string{"insert", "upload-bundle", "update-track", "validate", "commit"}
	if !reflect.DeepEqual(publisher.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", publisher.calls, wantCalls)
	}
}

type fakePublisher struct {
	calls  []string
	tracks []Track
}

func (p *fakePublisher) InsertEdit(ctx context.Context, packageName PackageName) (Edit, error) {
	p.calls = append(p.calls, "insert")
	return Edit{ID: "edit-123"}, nil
}

func (p *fakePublisher) UploadBundle(ctx context.Context, packageName PackageName, editID string, bundlePath string) (BundleArtifact, error) {
	p.calls = append(p.calls, "upload-bundle")
	return BundleArtifact{VersionCode: 42}, nil
}

func (p *fakePublisher) UpdateTrack(ctx context.Context, packageName PackageName, editID string, track TrackName, release TrackRelease) (Track, error) {
	p.calls = append(p.calls, "update-track")
	return Track{Name: track, Releases: []TrackRelease{release}}, nil
}

func (p *fakePublisher) ValidateEdit(ctx context.Context, packageName PackageName, editID string) error {
	p.calls = append(p.calls, "validate")
	return nil
}

func (p *fakePublisher) CommitEdit(ctx context.Context, packageName PackageName, editID string) (Edit, error) {
	p.calls = append(p.calls, "commit")
	return Edit{ID: editID}, nil
}

func (p *fakePublisher) DeleteEdit(ctx context.Context, packageName PackageName, editID string) error {
	p.calls = append(p.calls, "delete")
	return nil
}

func (p *fakePublisher) ListTracks(ctx context.Context, packageName PackageName, editID string) ([]Track, error) {
	p.calls = append(p.calls, "list-tracks")
	return p.tracks, nil
}
