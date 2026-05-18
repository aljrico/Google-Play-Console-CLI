package play

import (
	"context"
	"fmt"
)

func PublishInternal(ctx context.Context, publisher Publisher, options PublishInternalOptions) (result PublishResult, err error) {
	plan, err := NewPublishInternalPlan(options)
	if err != nil {
		return PublishResult{}, err
	}
	result = PublishResult{
		PackageName: options.PackageName,
		Track:       TrackInternal,
		DryRun:      options.DryRun,
		Committed:   false,
		Plan:        plan,
	}
	if options.DryRun {
		return result, nil
	}
	if publisher == nil {
		return PublishResult{}, fmt.Errorf("publisher is required")
	}

	edit, err := publisher.InsertEdit(ctx, options.PackageName)
	if err != nil {
		return PublishResult{}, err
	}
	result.Edit = &edit

	shouldDeleteEdit := true
	defer func() {
		if shouldDeleteEdit {
			if cleanupErr := publisher.DeleteEdit(ctx, options.PackageName, edit.ID); cleanupErr != nil && err == nil {
				err = cleanupErr
			}
		}
	}()

	bundle, err := publisher.UploadBundle(ctx, options.PackageName, edit.ID, options.BundlePath)
	if err != nil {
		return PublishResult{}, err
	}
	result.Bundle = &bundle

	release := TrackRelease{
		Name:         options.ReleaseName,
		Status:       options.Status,
		VersionCodes: []int64{bundle.VersionCode},
	}
	if _, err := publisher.UpdateTrack(ctx, options.PackageName, edit.ID, TrackInternal, release); err != nil {
		return PublishResult{}, err
	}
	if err := publisher.ValidateEdit(ctx, options.PackageName, edit.ID); err != nil {
		return PublishResult{}, err
	}
	if !options.Confirm {
		return result, nil
	}

	committedEdit, err := publisher.CommitEdit(ctx, options.PackageName, edit.ID)
	if err != nil {
		return PublishResult{}, err
	}
	shouldDeleteEdit = false
	result.Edit = &committedEdit
	result.Committed = true
	return result, nil
}
