package play

import (
	"context"
	"errors"
	"fmt"
)

type TrackLister interface {
	InsertEdit(ctx context.Context, packageName PackageName) (Edit, error)
	ListTracks(ctx context.Context, packageName PackageName, editID string) ([]Track, error)
	DeleteEdit(ctx context.Context, packageName PackageName, editID string) error
}

func ListTracks(ctx context.Context, publisher TrackLister, packageName PackageName) (tracks []Track, err error) {
	if packageName == "" {
		return nil, fmt.Errorf("package name is required")
	}
	if publisher == nil {
		return nil, fmt.Errorf("publisher is required")
	}

	edit, err := publisher.InsertEdit(ctx, packageName)
	if err != nil {
		return nil, err
	}

	defer func() {
		if cleanupErr := publisher.DeleteEdit(ctx, packageName, edit.ID); cleanupErr != nil {
			err = errors.Join(err, cleanupErr)
		}
	}()

	tracks, err = publisher.ListTracks(ctx, packageName, edit.ID)
	if err != nil {
		return nil, err
	}
	return tracks, nil
}
