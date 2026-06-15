package play

import (
	"context"
	"fmt"

	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/googleapi"
)

func (p *GooglePublisher) ListTracks(ctx context.Context, packageName PackageName, editID string) ([]Track, error) {
	response, err := p.service.Edits.Tracks.List(packageName.String(), editID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("list tracks for %s: %w", packageName, err)
	}
	tracks := make([]Track, 0, len(response.Tracks))
	for _, apiTrack := range response.Tracks {
		tracks = append(tracks, trackFromAPI(apiTrack))
	}
	return tracks, nil
}

func (p *GooglePublisher) AppendTrackRelease(ctx context.Context, packageName PackageName, editID string, trackName TrackName, release TrackRelease) (Track, error) {
	apiTrack, err := p.service.Edits.Tracks.Get(packageName.String(), editID, trackName.String()).Context(ctx).Do()
	if err != nil {
		return Track{}, fmt.Errorf("get %s track for %s: %w", trackName, packageName, err)
	}
	if apiTrack == nil {
		apiTrack = &androidpublisher.Track{Track: trackName.String()}
	}
	apiTrack.Track = trackName.String()
	upsertTrackRelease(apiTrack, releaseToAPI(release))

	updatedTrack, err := p.service.Edits.Tracks.Update(packageName.String(), editID, trackName.String(), apiTrack).Context(ctx).Do()
	if err != nil {
		return Track{}, fmt.Errorf("append release to %s track for %s: %w", trackName, packageName, err)
	}
	return trackFromAPI(updatedTrack), nil
}

func (p *GooglePublisher) PromoteTrackRelease(ctx context.Context, packageName PackageName, editID string, sourceTrack TrackName, targetTrack TrackName, versionCode int64, status ReleaseStatus, userFraction *float64, releaseNotes []ReleaseNote) (TrackRelease, error) {
	source, err := p.service.Edits.Tracks.Get(packageName.String(), editID, sourceTrack.String()).Context(ctx).Do()
	if err != nil {
		return TrackRelease{}, fmt.Errorf("get %s track for %s: %w", sourceTrack, packageName, err)
	}
	apiRelease, err := selectReleaseByVersionCode(source, versionCode)
	if err != nil {
		return TrackRelease{}, fmt.Errorf("find version code %d on %s track for %s: %w", versionCode, sourceTrack, packageName, err)
	}
	setReleaseStatus(apiRelease, status, userFraction)
	if len(releaseNotes) > 0 {
		apiRelease.ReleaseNotes = releaseNotesToAPI(releaseNotes)
	}

	target, err := p.service.Edits.Tracks.Get(packageName.String(), editID, targetTrack.String()).Context(ctx).Do()
	if err != nil {
		return TrackRelease{}, fmt.Errorf("get %s track for %s: %w", targetTrack, packageName, err)
	}
	if target == nil {
		target = &androidpublisher.Track{Track: targetTrack.String()}
	}
	target.Track = targetTrack.String()
	upsertTrackRelease(target, apiRelease)

	if _, err := p.service.Edits.Tracks.Update(packageName.String(), editID, targetTrack.String(), target).Context(ctx).Do(); err != nil {
		return TrackRelease{}, fmt.Errorf("promote release from %s to %s for %s: %w", sourceTrack, targetTrack, packageName, err)
	}
	return releaseFromAPI(apiRelease), nil
}

func (p *GooglePublisher) UpdateTrackReleaseStatus(ctx context.Context, packageName PackageName, editID string, trackName TrackName, versionCode int64, status ReleaseStatus, userFraction *float64) (TrackRelease, error) {
	apiTrack, err := p.service.Edits.Tracks.Get(packageName.String(), editID, trackName.String()).Context(ctx).Do()
	if err != nil {
		return TrackRelease{}, fmt.Errorf("get %s track for %s: %w", trackName, packageName, err)
	}
	apiRelease, err := selectReleaseByVersionCode(apiTrack, versionCode)
	if err != nil {
		return TrackRelease{}, fmt.Errorf("find version code %d on %s track for %s: %w", versionCode, trackName, packageName, err)
	}
	setReleaseStatus(apiRelease, status, userFraction)
	// A track allows only one release per exclusive status (completed, draft), so
	// moving a release into one — e.g. resuming a staged rollout to completed
	// while the prior version is still live, or completing a draft — must
	// supersede the existing release of that status instead of leaving two behind
	// (which the API rejects, e.g. "Only one completed release is allowed").
	if isExclusiveStatusRelease(apiRelease) {
		apiTrack.Releases = dropReleasesWithStatusExcept(apiTrack.Releases, apiRelease, apiRelease.Status)
	}

	if _, err := p.service.Edits.Tracks.Update(packageName.String(), editID, trackName.String(), apiTrack).Context(ctx).Do(); err != nil {
		return TrackRelease{}, fmt.Errorf("update rollout status for version code %d on %s track for %s: %w", versionCode, trackName, packageName, err)
	}
	return releaseFromAPI(apiRelease), nil
}

func trackFromAPI(apiTrack *androidpublisher.Track) Track {
	if apiTrack == nil {
		return Track{}
	}
	releases := make([]TrackRelease, 0, len(apiTrack.Releases))
	for _, apiRelease := range apiTrack.Releases {
		releases = append(releases, releaseFromAPI(apiRelease))
	}
	return Track{Name: TrackName(apiTrack.Track), Releases: releases}
}

func releaseToAPI(release TrackRelease) *androidpublisher.TrackRelease {
	apiRelease := &androidpublisher.TrackRelease{
		Name:            release.Name,
		Status:          release.Status.String(),
		VersionCodes:    googleapi.Int64s(release.VersionCodes),
		ForceSendFields: []string{"Status"},
	}
	if release.UserFraction != nil {
		apiRelease.UserFraction = *release.UserFraction
		apiRelease.ForceSendFields = append(apiRelease.ForceSendFields, "UserFraction")
	}
	apiRelease.ReleaseNotes = releaseNotesToAPI(release.ReleaseNotes)
	return apiRelease
}

func releaseNotesToAPI(notes []ReleaseNote) []*androidpublisher.LocalizedText {
	apiNotes := make([]*androidpublisher.LocalizedText, 0, len(notes))
	for _, note := range notes {
		apiNotes = append(apiNotes, &androidpublisher.LocalizedText{
			Language: note.Language.String(),
			Text:     note.Text,
		})
	}
	return apiNotes
}

func selectReleaseByVersionCode(track *androidpublisher.Track, versionCode int64) (*androidpublisher.TrackRelease, error) {
	if track == nil || len(track.Releases) == 0 {
		return nil, fmt.Errorf("track has no releases")
	}
	for _, release := range track.Releases {
		if hasVersionCode(release, versionCode) {
			return release, nil
		}
	}
	return nil, fmt.Errorf("version code not found")
}

func hasVersionCode(release *androidpublisher.TrackRelease, versionCode int64) bool {
	if release == nil {
		return false
	}
	for _, candidate := range release.VersionCodes {
		if candidate == versionCode {
			return true
		}
	}
	return false
}

func upsertTrackRelease(track *androidpublisher.Track, release *androidpublisher.TrackRelease) {
	// Play allows only one release per exclusive status — one completed and one
	// draft — per track. A new release in either status must therefore supersede
	// an existing one of the same status rather than be appended alongside it
	// (the API rejects the latter, e.g. "Only one draft release is allowed").
	// Staged inProgress/halted releases legitimately coexist with the prior
	// completed release, so they are left untouched.
	if isExclusiveStatusRelease(release) {
		track.Releases = dropReleasesWithStatusExcept(track.Releases, release, release.Status)
	}
	for index, existingRelease := range track.Releases {
		if releasesShareVersionCode(existingRelease, release) {
			track.Releases[index] = release
			return
		}
	}
	track.Releases = append(track.Releases, release)
}

// isExclusiveStatusRelease reports whether a track may hold at most one release
// in this release's status (completed or draft), so a new one must supersede an
// existing one of that status. inProgress/halted staged rollouts are not
// exclusive and coexist with a completed release.
func isExclusiveStatusRelease(release *androidpublisher.TrackRelease) bool {
	return release != nil &&
		(release.Status == string(ReleaseStatusCompleted) || release.Status == string(ReleaseStatusDraft))
}

func dropReleasesWithStatusExcept(releases []*androidpublisher.TrackRelease, keep *androidpublisher.TrackRelease, status string) []*androidpublisher.TrackRelease {
	retained := make([]*androidpublisher.TrackRelease, 0, len(releases))
	for _, existingRelease := range releases {
		if existingRelease != nil && existingRelease.Status == status && !releasesShareVersionCode(existingRelease, keep) {
			continue
		}
		retained = append(retained, existingRelease)
	}
	return retained
}

func releasesShareVersionCode(left *androidpublisher.TrackRelease, right *androidpublisher.TrackRelease) bool {
	if left == nil || right == nil {
		return false
	}
	for _, versionCode := range right.VersionCodes {
		if hasVersionCode(left, versionCode) {
			return true
		}
	}
	return false
}

func setReleaseStatus(release *androidpublisher.TrackRelease, status ReleaseStatus, userFraction *float64) {
	release.Status = status.String()
	release.UserFraction = 0
	release.ForceSendFields = appendUniqueField(release.ForceSendFields, "Status")
	release.NullFields = removeField(release.NullFields, "UserFraction")
	if userFraction == nil {
		release.ForceSendFields = removeField(release.ForceSendFields, "UserFraction")
		return
	}
	release.UserFraction = *userFraction
	release.ForceSendFields = appendUniqueField(release.ForceSendFields, "UserFraction")
}

func appendUniqueField(fields []string, field string) []string {
	for _, candidate := range fields {
		if candidate == field {
			return fields
		}
	}
	return append(fields, field)
}

func removeField(fields []string, field string) []string {
	result := fields[:0]
	for _, candidate := range fields {
		if candidate != field {
			result = append(result, candidate)
		}
	}
	return result
}

func releaseFromAPI(apiRelease *androidpublisher.TrackRelease) TrackRelease {
	if apiRelease == nil {
		return TrackRelease{}
	}
	return TrackRelease{
		Name:         apiRelease.Name,
		Status:       ReleaseStatus(apiRelease.Status),
		UserFraction: userFractionFromAPI(apiRelease),
		VersionCodes: []int64(apiRelease.VersionCodes),
		ReleaseNotes: releaseNotesFromAPI(apiRelease.ReleaseNotes),
	}
}

func releaseNotesFromAPI(apiNotes []*androidpublisher.LocalizedText) []ReleaseNote {
	notes := make([]ReleaseNote, 0, len(apiNotes))
	for _, apiNote := range apiNotes {
		if apiNote == nil {
			continue
		}
		notes = append(notes, ReleaseNote{
			Language: ListingLanguage(apiNote.Language),
			Text:     apiNote.Text,
		})
	}
	return notes
}
