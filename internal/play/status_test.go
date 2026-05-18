package play

import (
	"context"
	"reflect"
	"testing"
)

func TestGetReleaseStatusSummarizesVisibleReleases(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	userFraction := 0.25
	publisher := &fakePublisher{
		tracks: []Track{
			{
				Name: TrackProduction,
				Releases: []TrackRelease{
					{Name: "1.0", Status: ReleaseStatusCompleted, VersionCodes: []int64{10}},
					{Name: "1.1", Status: ReleaseStatusInProgress, UserFraction: &userFraction, VersionCodes: []int64{11}},
					{Name: "1.2", Status: ReleaseStatusDraft, VersionCodes: []int64{12}},
				},
			},
			{
				Name: TrackBeta,
				Releases: []TrackRelease{
					{Name: "0.9", Status: ReleaseStatusHalted, VersionCodes: []int64{9}},
				},
			},
		},
	}

	overview, err := GetReleaseStatus(context.Background(), publisher, ReleaseStatusOptions{PackageName: packageName})
	if err != nil {
		t.Fatalf("GetReleaseStatus() error = %v", err)
	}
	if overview.Summary.TrackCount != 2 {
		t.Fatalf("TrackCount = %d, want 2", overview.Summary.TrackCount)
	}
	if overview.Summary.ReleaseCount != 3 {
		t.Fatalf("ReleaseCount = %d, want 3", overview.Summary.ReleaseCount)
	}
	if overview.Summary.CompletedCount != 1 || overview.Summary.InProgressCount != 1 || overview.Summary.HaltedCount != 1 {
		t.Fatalf("Summary = %#v, want one completed, in-progress, and halted", overview.Summary)
	}
	if overview.Summary.DraftCount != 0 {
		t.Fatalf("DraftCount = %d, want 0", overview.Summary.DraftCount)
	}
	if len(overview.Tracks[0].Releases) != 2 {
		t.Fatalf("production releases = %d, want draft filtered out", len(overview.Tracks[0].Releases))
	}

	wantCalls := []string{"insert", "list-tracks", "delete"}
	if !reflect.DeepEqual(publisher.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", publisher.calls, wantCalls)
	}
}

func TestGetReleaseStatusCanIncludeDraftReleases(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	publisher := &fakePublisher{
		tracks: []Track{
			{
				Name: TrackInternal,
				Releases: []TrackRelease{
					{Name: "1.0", Status: ReleaseStatusDraft, VersionCodes: []int64{10}},
				},
			},
		},
	}

	overview, err := GetReleaseStatus(context.Background(), publisher, ReleaseStatusOptions{
		PackageName:  packageName,
		IncludeDraft: true,
	})
	if err != nil {
		t.Fatalf("GetReleaseStatus() error = %v", err)
	}
	if overview.Summary.DraftCount != 1 {
		t.Fatalf("DraftCount = %d, want 1", overview.Summary.DraftCount)
	}
	if len(overview.Tracks) != 1 {
		t.Fatalf("len(Tracks) = %d, want 1", len(overview.Tracks))
	}
}

func TestGetReleaseStatusRejectsInvalidPackage(t *testing.T) {
	_, err := GetReleaseStatus(context.Background(), nil, ReleaseStatusOptions{PackageName: "bad"})
	if err == nil {
		t.Fatal("expected package validation error")
	}
}
