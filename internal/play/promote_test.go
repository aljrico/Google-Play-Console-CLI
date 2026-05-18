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
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("PromoteRelease() error = %v", err)
	}
	if !result.DryRun {
		t.Fatal("DryRun = false, want true")
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

type fakePromoter struct {
	calls []string
}

func (p *fakePromoter) InsertEdit(ctx context.Context, packageName PackageName) (Edit, error) {
	p.calls = append(p.calls, "insert")
	return Edit{ID: "edit-123"}, nil
}

func (p *fakePromoter) PromoteTrackRelease(ctx context.Context, packageName PackageName, editID string, sourceTrack TrackName, targetTrack TrackName, releaseName string) (TrackRelease, error) {
	p.calls = append(p.calls, "promote")
	return TrackRelease{Name: releaseName, Status: ReleaseStatusCompleted, VersionCodes: []int64{42}}, nil
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
