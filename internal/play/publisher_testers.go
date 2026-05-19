package play

import (
	"context"
	"fmt"
)

func (p *GooglePublisher) GetTesters(ctx context.Context, packageName PackageName, editID string, track TrackName) (TrackTesters, error) {
	testers, err := p.service.Edits.Testers.Get(packageName.String(), editID, track.String()).Context(ctx).Do()
	if err != nil {
		return TrackTesters{}, fmt.Errorf("get %s testers for %s: %w", track, packageName, err)
	}
	return testersFromAPI(packageName, track, testers), nil
}

func (p *GooglePublisher) UpdateTesters(ctx context.Context, packageName PackageName, editID string, track TrackName, googleGroups []TesterGoogleGroup) (TrackTesters, error) {
	apiTesters := testersToAPI(googleGroups)
	updatedTesters, err := p.service.Edits.Testers.Update(packageName.String(), editID, track.String(), apiTesters).Context(ctx).Do()
	if err != nil {
		return TrackTesters{}, fmt.Errorf("update %s testers for %s: %w", track, packageName, err)
	}
	return testersFromAPI(packageName, track, updatedTesters), nil
}
