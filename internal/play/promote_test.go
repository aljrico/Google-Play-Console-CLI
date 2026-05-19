package play

import (
	"context"
	"reflect"
	"testing"
)

func TestPromoteReleaseDryRunDoesNotRequirePromoter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := PromoteRelease(context.Background(), nil, PromoteReleaseOptions{
		PackageName: packageName,
		FromTrack:   TrackInternal,
		ToTrack:     TrackProduction,
		VersionCode: 42,
		Status:      ReleaseStatusDraft,
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("PromoteRelease() error = %v", err)
	}
	if !result.DryRun {
		t.Fatal("DryRun = false, want true")
	}
}

func TestPromoteReleaseDryRunIncludesReleaseNotes(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := PromoteRelease(context.Background(), nil, PromoteReleaseOptions{
		PackageName: packageName,
		FromTrack:   TrackInternal,
		ToTrack:     TrackProduction,
		VersionCode: 42,
		Status:      ReleaseStatusDraft,
		ReleaseNotes: []ReleaseNote{
			{Language: "en-US", Text: "Production rollout."},
		},
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("PromoteRelease() error = %v", err)
	}
	if len(result.Plan.ReleaseNotes) != 1 || result.Plan.ReleaseNotes[0].Text != "Production rollout." {
		t.Fatalf("ReleaseNotes = %#v, want production note", result.Plan.ReleaseNotes)
	}
	wantSteps := []string{"insert edit", "read internal track", "copy version code 42 to production track as draft", "replace release notes", "validate edit", "delete uncommitted edit"}
	if !reflect.DeepEqual(result.Plan.Steps, wantSteps) {
		t.Fatalf("steps = %#v, want %#v", result.Plan.Steps, wantSteps)
	}
}

func TestPromoteReleaseRejectsSameTrack(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = NewPromotePlan(PromoteReleaseOptions{
		PackageName: packageName,
		FromTrack:   TrackInternal,
		ToTrack:     TrackInternal,
		VersionCode: 42,
		Status:      ReleaseStatusDraft,
	})
	if err == nil {
		t.Fatal("expected same-track error")
	}
}

func TestPromoteReleaseValidatesAndCleansUpWithoutConfirm(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	promoter := &fakePromoter{}

	result, err := PromoteRelease(context.Background(), promoter, PromoteReleaseOptions{
		PackageName: packageName,
		FromTrack:   TrackInternal,
		ToTrack:     TrackProduction,
		VersionCode: 42,
		Status:      ReleaseStatusDraft,
	})
	if err != nil {
		t.Fatalf("PromoteRelease() error = %v", err)
	}
	if result.Committed {
		t.Fatal("Committed = true, want false")
	}

	wantCalls := []string{"insert", "promote", "validate", "delete"}
	if !reflect.DeepEqual(promoter.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", promoter.calls, wantCalls)
	}
}

func TestPromoteReleaseCommitsWithConfirm(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	promoter := &fakePromoter{}

	result, err := PromoteRelease(context.Background(), promoter, PromoteReleaseOptions{
		PackageName: packageName,
		FromTrack:   TrackInternal,
		ToTrack:     TrackProduction,
		VersionCode: 42,
		Status:      ReleaseStatusDraft,
		Confirm:     true,
	})
	if err != nil {
		t.Fatalf("PromoteRelease() error = %v", err)
	}
	if !result.Committed {
		t.Fatal("Committed = false, want true")
	}

	wantCalls := []string{"insert", "promote", "validate", "commit"}
	if !reflect.DeepEqual(promoter.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", promoter.calls, wantCalls)
	}
}

func TestPromoteReleaseRequiresVersionCode(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = NewPromotePlan(PromoteReleaseOptions{
		PackageName: packageName,
		FromTrack:   TrackInternal,
		ToTrack:     TrackProduction,
		Status:      ReleaseStatusDraft,
	})
	if err == nil {
		t.Fatal("expected version-code error")
	}
}

func TestPromoteReleaseRequiresUserFractionForStagedTarget(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = NewPromotePlan(PromoteReleaseOptions{
		PackageName: packageName,
		FromTrack:   TrackInternal,
		ToTrack:     TrackProduction,
		VersionCode: 42,
		Status:      ReleaseStatusInProgress,
	})
	if err == nil {
		t.Fatal("expected user-fraction error")
	}
}

type fakePromoter struct {
	calls        []string
	releaseNotes []ReleaseNote
}

func (p *fakePromoter) InsertEdit(ctx context.Context, packageName PackageName) (Edit, error) {
	p.calls = append(p.calls, "insert")
	return Edit{ID: "edit-123"}, nil
}

func (p *fakePromoter) PromoteTrackRelease(ctx context.Context, packageName PackageName, editID string, sourceTrack TrackName, targetTrack TrackName, versionCode int64, status ReleaseStatus, userFraction *float64, releaseNotes []ReleaseNote) (TrackRelease, error) {
	p.calls = append(p.calls, "promote")
	p.releaseNotes = releaseNotes
	return TrackRelease{Status: status, VersionCodes: []int64{versionCode}, ReleaseNotes: releaseNotes}, nil
}

func (p *fakePromoter) ValidateEdit(ctx context.Context, packageName PackageName, editID string) error {
	p.calls = append(p.calls, "validate")
	return nil
}

func (p *fakePromoter) CommitEdit(ctx context.Context, packageName PackageName, editID string) (Edit, error) {
	p.calls = append(p.calls, "commit")
	return Edit{ID: editID}, nil
}

func (p *fakePromoter) DeleteEdit(ctx context.Context, packageName PackageName, editID string) error {
	p.calls = append(p.calls, "delete")
	return nil
}
