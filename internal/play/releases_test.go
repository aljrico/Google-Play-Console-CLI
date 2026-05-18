package play

import (
	"context"
	"reflect"
	"testing"
)

func TestListTrackReleases(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	publisher := &fakePublisher{
		tracks: []Track{
			{
				Name: TrackInternal,
				Releases: []TrackRelease{
					{Name: "1.2.3", Status: ReleaseStatusCompleted, VersionCodes: []int64{42}},
				},
			},
		},
	}

	releases, err := ListTrackReleases(context.Background(), publisher, packageName, TrackInternal)
	if err != nil {
		t.Fatalf("ListTrackReleases() error = %v", err)
	}
	if len(releases) != 1 {
		t.Fatalf("len(releases) = %d, want 1", len(releases))
	}

	wantCalls := []string{"insert", "list-tracks", "delete"}
	if !reflect.DeepEqual(publisher.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", publisher.calls, wantCalls)
	}
}
