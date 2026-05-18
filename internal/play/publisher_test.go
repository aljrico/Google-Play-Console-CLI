package play

import (
	"testing"

	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/googleapi"
)

func TestUpsertTrackReleaseAppendsNewVersionCode(t *testing.T) {
	track := &androidpublisher.Track{
		Releases: []*androidpublisher.TrackRelease{
			{VersionCodes: googleapi.Int64s([]int64{41})},
		},
	}

	upsertTrackRelease(track, &androidpublisher.TrackRelease{VersionCodes: googleapi.Int64s([]int64{42})})

	if len(track.Releases) != 2 {
		t.Fatalf("len(Releases) = %d, want 2", len(track.Releases))
	}
}

func TestUpsertTrackReleaseReplacesExistingVersionCode(t *testing.T) {
	track := &androidpublisher.Track{
		Releases: []*androidpublisher.TrackRelease{
			{Name: "old", VersionCodes: googleapi.Int64s([]int64{42})},
		},
	}

	upsertTrackRelease(track, &androidpublisher.TrackRelease{Name: "new", VersionCodes: googleapi.Int64s([]int64{42})})

	if len(track.Releases) != 1 {
		t.Fatalf("len(Releases) = %d, want 1", len(track.Releases))
	}
	if track.Releases[0].Name != "new" {
		t.Fatalf("release name = %q, want new", track.Releases[0].Name)
	}
}

func TestSetReleaseStatusClearsUserFractionWhenCompleted(t *testing.T) {
	release := &androidpublisher.TrackRelease{
		Status:          ReleaseStatusInProgress.String(),
		UserFraction:    0.25,
		ForceSendFields: []string{"UserFraction"},
	}

	setReleaseStatus(release, ReleaseStatusCompleted, nil)

	if release.Status != ReleaseStatusCompleted.String() {
		t.Fatalf("Status = %q, want %q", release.Status, ReleaseStatusCompleted)
	}
	if release.UserFraction != 0 {
		t.Fatalf("UserFraction = %f, want 0", release.UserFraction)
	}
	for _, field := range release.ForceSendFields {
		if field == "UserFraction" {
			t.Fatal("UserFraction remained in ForceSendFields")
		}
	}
}
