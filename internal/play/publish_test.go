package play

import (
	"context"
	"os"
	"path/filepath"
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

func TestNewPublishInternalPlanSupportsCustomTrack(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	plan, err := NewPublishInternalPlan(PublishInternalOptions{
		PackageName: packageName,
		Track:       TrackBeta,
		BundlePath:  "app-release.aab",
		Status:      ReleaseStatusCompleted,
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("NewPublishInternalPlan() error = %v", err)
	}
	if plan.Track != TrackBeta {
		t.Fatalf("Track = %q, want %q", plan.Track, TrackBeta)
	}
}

func TestNewPublishInternalPlanSupportsAPK(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	plan, err := NewPublishInternalPlan(PublishInternalOptions{
		PackageName: packageName,
		APKPath:     "app-release.apk",
		Status:      ReleaseStatusCompleted,
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("NewPublishInternalPlan() error = %v", err)
	}
	if plan.Artifact != ArtifactKindAPK || plan.APKPath != "app-release.apk" {
		t.Fatalf("plan = %#v, want APK artifact", plan)
	}
	if plan.Steps[1] != "upload APK" {
		t.Fatalf("upload step = %q, want APK upload", plan.Steps[1])
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

func TestPublishInternalChecksBundleBeforePublisher(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = PublishInternal(context.Background(), nil, PublishInternalOptions{
		PackageName: packageName,
		BundlePath:  t.TempDir() + "/missing.aab",
		Status:      ReleaseStatusCompleted,
	})
	if err == nil {
		t.Fatal("expected missing bundle error")
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

func TestPublishInternalRejectsMultipleArtifacts(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = NewPublishInternalPlan(PublishInternalOptions{
		PackageName: packageName,
		APKPath:     "app-release.apk",
		BundlePath:  "app-release.aab",
		Status:      ReleaseStatusCompleted,
	})
	if err == nil {
		t.Fatal("expected multiple artifact error")
	}
}

func TestPublishInternalValidatesAndCleansUpWithoutConfirm(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	publisher := &fakePublisher{}
	bundlePath := writeTestBundle(t)

	result, err := PublishInternal(context.Background(), publisher, PublishInternalOptions{
		PackageName: packageName,
		BundlePath:  bundlePath,
		Status:      ReleaseStatusCompleted,
	})
	if err != nil {
		t.Fatalf("PublishInternal() error = %v", err)
	}
	if result.Committed {
		t.Fatal("Committed = true, want false")
	}

	wantCalls := []string{"insert", "upload-bundle", "append-track-release", "validate", "delete"}
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
	bundlePath := writeTestBundle(t)

	result, err := PublishInternal(context.Background(), publisher, PublishInternalOptions{
		PackageName: packageName,
		BundlePath:  bundlePath,
		Status:      ReleaseStatusCompleted,
		Confirm:     true,
	})
	if err != nil {
		t.Fatalf("PublishInternal() error = %v", err)
	}
	if !result.Committed {
		t.Fatal("Committed = false, want true")
	}

	wantCalls := []string{"insert", "upload-bundle", "append-track-release", "validate", "commit"}
	if !reflect.DeepEqual(publisher.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", publisher.calls, wantCalls)
	}
}

func TestPublishInternalUploadsAPK(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	publisher := &fakePublisher{}
	apkPath := writeTestAPK(t)

	result, err := PublishInternal(context.Background(), publisher, PublishInternalOptions{
		PackageName: packageName,
		APKPath:     apkPath,
		Status:      ReleaseStatusCompleted,
	})
	if err != nil {
		t.Fatalf("PublishInternal() error = %v", err)
	}
	if result.APK == nil || result.APK.VersionCode != 43 {
		t.Fatalf("APK = %#v, want version code 43", result.APK)
	}

	wantCalls := []string{"insert", "upload-apk", "append-track-release", "validate", "delete"}
	if !reflect.DeepEqual(publisher.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", publisher.calls, wantCalls)
	}
	if publisher.appendedReleases[0].VersionCodes[0] != 43 {
		t.Fatalf("release version code = %d, want APK version code", publisher.appendedReleases[0].VersionCodes[0])
	}
}

func TestPublishInternalAppendsTrackRelease(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	publisher := &fakePublisher{}
	bundlePath := writeTestBundle(t)

	_, err = PublishInternal(context.Background(), publisher, PublishInternalOptions{
		PackageName: packageName,
		BundlePath:  bundlePath,
		ReleaseName: "1.2.3",
		Status:      ReleaseStatusCompleted,
		ReleaseNotes: []ReleaseNote{
			{Language: "en-US", Text: "Bug fixes."},
		},
	})
	if err != nil {
		t.Fatalf("PublishInternal() error = %v", err)
	}
	if len(publisher.appendedReleases) != 1 {
		t.Fatalf("len(appendedReleases) = %d, want 1", len(publisher.appendedReleases))
	}
	release := publisher.appendedReleases[0]
	if release.Name != "1.2.3" {
		t.Fatalf("release name = %q, want 1.2.3", release.Name)
	}
	if release.VersionCodes[0] != 42 {
		t.Fatalf("release version code = %d, want 42", release.VersionCodes[0])
	}
	if len(release.ReleaseNotes) != 1 || release.ReleaseNotes[0].Text != "Bug fixes." {
		t.Fatalf("release notes = %#v, want bug fixes note", release.ReleaseNotes)
	}
}

func writeTestBundle(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "app-release.aab")
	if err := os.WriteFile(path, []byte("bundle"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}

func writeTestAPK(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "app-release.apk")
	if err := os.WriteFile(path, []byte("apk"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}

type fakePublisher struct {
	calls            []string
	tracks           []Track
	appendedReleases []TrackRelease
}

func (p *fakePublisher) InsertEdit(ctx context.Context, packageName PackageName) (Edit, error) {
	p.calls = append(p.calls, "insert")
	return Edit{ID: "edit-123"}, nil
}

func (p *fakePublisher) UploadBundle(ctx context.Context, packageName PackageName, editID string, bundlePath string) (BundleArtifact, error) {
	p.calls = append(p.calls, "upload-bundle")
	return BundleArtifact{VersionCode: 42}, nil
}

func (p *fakePublisher) UploadAPK(ctx context.Context, packageName PackageName, editID string, apkPath string) (APKArtifact, error) {
	p.calls = append(p.calls, "upload-apk")
	return APKArtifact{VersionCode: 43}, nil
}

func (p *fakePublisher) UpdateTrack(ctx context.Context, packageName PackageName, editID string, track Track) (Track, error) {
	p.calls = append(p.calls, "update-track")
	return track, nil
}

func (p *fakePublisher) AppendTrackRelease(ctx context.Context, packageName PackageName, editID string, trackName TrackName, release TrackRelease) (Track, error) {
	p.calls = append(p.calls, "append-track-release")
	p.appendedReleases = append(p.appendedReleases, release)
	return Track{Name: trackName, Releases: append([]TrackRelease{}, p.appendedReleases...)}, nil
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
