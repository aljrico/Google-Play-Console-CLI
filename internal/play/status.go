package play

import (
	"context"
	"fmt"
)

type ReleaseStatusOptions struct {
	PackageName  PackageName `json:"packageName"`
	IncludeDraft bool        `json:"includeDraft"`
}

func (o ReleaseStatusOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	return nil
}

type ReleaseStatusSummary struct {
	TrackCount      int `json:"trackCount"`
	ReleaseCount    int `json:"releaseCount"`
	CompletedCount  int `json:"completedCount"`
	InProgressCount int `json:"inProgressCount"`
	HaltedCount     int `json:"haltedCount"`
	DraftCount      int `json:"draftCount"`
}

type TrackReleaseStatus struct {
	Track    TrackName      `json:"track"`
	Releases []TrackRelease `json:"releases"`
}

type ReleaseStatusOverview struct {
	PackageName PackageName          `json:"packageName"`
	Options     ReleaseStatusOptions `json:"options"`
	Summary     ReleaseStatusSummary `json:"summary"`
	Tracks      []TrackReleaseStatus `json:"tracks"`
}

func GetReleaseStatus(ctx context.Context, lister TrackLister, options ReleaseStatusOptions) (ReleaseStatusOverview, error) {
	if err := options.Validate(); err != nil {
		return ReleaseStatusOverview{}, err
	}
	if lister == nil {
		return ReleaseStatusOverview{}, fmt.Errorf("track lister is required")
	}
	tracks, err := ListTracks(ctx, lister, options.PackageName)
	if err != nil {
		return ReleaseStatusOverview{}, err
	}
	return releaseStatusOverviewFromTracks(options, tracks), nil
}

func releaseStatusOverviewFromTracks(options ReleaseStatusOptions, tracks []Track) ReleaseStatusOverview {
	overview := ReleaseStatusOverview{
		PackageName: options.PackageName,
		Options:     options,
		Tracks:      []TrackReleaseStatus{},
	}
	for _, track := range tracks {
		releases := visibleStatusReleases(track.Releases, options.IncludeDraft)
		if len(releases) == 0 {
			continue
		}
		overview.Tracks = append(overview.Tracks, TrackReleaseStatus{
			Track:    track.Name,
			Releases: releases,
		})
	}
	overview.Summary = summarizeReleaseStatus(overview.Tracks)
	return overview
}

func visibleStatusReleases(releases []TrackRelease, includeDraft bool) []TrackRelease {
	visible := make([]TrackRelease, 0, len(releases))
	for _, release := range releases {
		if !includeDraft && release.Status == ReleaseStatusDraft {
			continue
		}
		visible = append(visible, release)
	}
	return visible
}

func summarizeReleaseStatus(tracks []TrackReleaseStatus) ReleaseStatusSummary {
	summary := ReleaseStatusSummary{TrackCount: len(tracks)}
	for _, track := range tracks {
		for _, release := range track.Releases {
			summary.ReleaseCount++
			switch release.Status {
			case ReleaseStatusCompleted:
				summary.CompletedCount++
			case ReleaseStatusInProgress:
				summary.InProgressCount++
			case ReleaseStatusHalted:
				summary.HaltedCount++
			case ReleaseStatusDraft:
				summary.DraftCount++
			}
		}
	}
	return summary
}
