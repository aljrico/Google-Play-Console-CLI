package play

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
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
			{
				Name: TrackAlpha,
				Releases: []TrackRelease{
					{Name: "0.8", Status: ReleaseStatusDraft, VersionCodes: []int64{8}},
				},
			},
		},
	}

	overview, err := GetReleaseStatus(context.Background(), publisher, ReleaseStatusOptions{PackageName: packageName})
	if err != nil {
		t.Fatalf("GetReleaseStatus() error = %v", err)
	}
	if overview.Summary.TotalTrackCount != 3 {
		t.Fatalf("TotalTrackCount = %d, want 3", overview.Summary.TotalTrackCount)
	}
	if overview.Summary.VisibleTrackCount != 2 {
		t.Fatalf("VisibleTrackCount = %d, want 2", overview.Summary.VisibleTrackCount)
	}
	if overview.Summary.TotalReleaseCount != 5 {
		t.Fatalf("TotalReleaseCount = %d, want 5", overview.Summary.TotalReleaseCount)
	}
	if overview.Summary.VisibleReleaseCount != 3 {
		t.Fatalf("VisibleReleaseCount = %d, want 3", overview.Summary.VisibleReleaseCount)
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
	if overview.Summary.VisibleReleaseCount != 1 {
		t.Fatalf("VisibleReleaseCount = %d, want 1", overview.Summary.VisibleReleaseCount)
	}
	if len(overview.Tracks) != 1 {
		t.Fatalf("len(Tracks) = %d, want 1", len(overview.Tracks))
	}
}

func TestGetReleaseStatusCountsUnknownStatuses(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	publisher := &fakePublisher{
		tracks: []Track{
			{
				Name: TrackInternal,
				Releases: []TrackRelease{
					{Name: "mystery", Status: ReleaseStatus("statusUnspecified"), VersionCodes: []int64{10}},
				},
			},
		},
	}

	overview, err := GetReleaseStatus(context.Background(), publisher, ReleaseStatusOptions{PackageName: packageName})
	if err != nil {
		t.Fatalf("GetReleaseStatus() error = %v", err)
	}
	if overview.Summary.UnknownCount != 1 {
		t.Fatalf("UnknownCount = %d, want 1", overview.Summary.UnknownCount)
	}
	if overview.Summary.VisibleReleaseCount != 1 {
		t.Fatalf("VisibleReleaseCount = %d, want 1", overview.Summary.VisibleReleaseCount)
	}
}

func TestReleaseStatusOverviewJSONShape(t *testing.T) {
	payload, err := json.Marshal(ReleaseStatusOverview{
		PackageName: "com.example.app",
		Options: ReleaseStatusOptions{
			PackageName: "com.example.app",
		},
		Summary: ReleaseStatusSummary{
			TotalTrackCount:     2,
			VisibleTrackCount:   1,
			TotalReleaseCount:   2,
			VisibleReleaseCount: 1,
			UnknownCount:        0,
		},
		Tracks: []TrackReleaseStatus{
			{
				Track: TrackProduction,
				Releases: []TrackRelease{
					{Name: "1.0", Status: ReleaseStatusCompleted, VersionCodes: []int64{10}},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	output := string(payload)
	for _, want := range []string{
		`"packageName":"com.example.app"`,
		`"totalTrackCount":2`,
		`"visibleTrackCount":1`,
		`"totalReleaseCount":2`,
		`"visibleReleaseCount":1`,
		`"unknownCount":0`,
		`"tracks":[`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("json = %s, want %s", output, want)
		}
	}
}

func TestGetReleaseStatusRejectsInvalidPackage(t *testing.T) {
	_, err := GetReleaseStatus(context.Background(), nil, ReleaseStatusOptions{PackageName: "bad"})
	if err == nil {
		t.Fatal("expected package validation error")
	}
}
