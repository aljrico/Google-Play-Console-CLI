package play

import (
	"context"
	"fmt"
)

func ListTrackReleases(ctx context.Context, publisher Publisher, packageName PackageName, track TrackName) ([]TrackRelease, error) {
	if packageName == "" {
		return nil, fmt.Errorf("package name is required")
	}
	if track == "" {
		return nil, fmt.Errorf("track name is required")
	}
	if publisher == nil {
		return nil, fmt.Errorf("publisher is required")
	}
	tracks, err := ListTracks(ctx, publisher, packageName)
	if err != nil {
		return nil, err
	}
	for _, candidate := range tracks {
		if candidate.Name == track {
			return candidate.Releases, nil
		}
	}
	return []TrackRelease{}, nil
}
