package play

import (
	"context"
	"reflect"
	"testing"
)

func TestListTracksUsesTemporaryEdit(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	publisher := &fakePublisher{
		tracks: []Track{
			{Name: TrackInternal},
			{Name: TrackProduction},
		},
	}

	tracks, err := ListTracks(context.Background(), publisher, packageName)
	if err != nil {
		t.Fatalf("ListTracks() error = %v", err)
	}
	if len(tracks) != 2 {
		t.Fatalf("len(tracks) = %d, want 2", len(tracks))
	}

	wantCalls := []string{"insert", "list-tracks", "delete"}
	if !reflect.DeepEqual(publisher.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", publisher.calls, wantCalls)
	}
}
