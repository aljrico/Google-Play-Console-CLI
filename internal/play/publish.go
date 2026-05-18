package play

import (
	"context"
	"errors"
	"fmt"
)

type InternalPublisher interface {
	InsertEdit(ctx context.Context, packageName PackageName) (Edit, error)
	UploadAPK(ctx context.Context, packageName PackageName, editID string, apkPath string) (APKArtifact, error)
	UploadBundle(ctx context.Context, packageName PackageName, editID string, bundlePath string) (BundleArtifact, error)
	AppendTrackRelease(ctx context.Context, packageName PackageName, editID string, trackName TrackName, release TrackRelease) (Track, error)
	ValidateEdit(ctx context.Context, packageName PackageName, editID string) error
	CommitEdit(ctx context.Context, packageName PackageName, editID string) (Edit, error)
	DeleteEdit(ctx context.Context, packageName PackageName, editID string) error
}

func PublishInternal(ctx context.Context, publisher InternalPublisher, options PublishInternalOptions) (result PublishResult, err error) {
	plan, err := NewPublishInternalPlan(options)
	if err != nil {
		return PublishResult{}, err
	}
	result = PublishResult{
		PackageName: options.PackageName,
		Track:       plan.Track,
		DryRun:      options.DryRun,
		Committed:   false,
		Plan:        plan,
	}
	if options.DryRun {
		return result, nil
	}
	if err := ValidateReadableReleaseArtifact(options); err != nil {
		return PublishResult{}, err
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
			cleanupCtx, cancel := newCleanupContext()
			defer cancel()
			if cleanupErr := publisher.DeleteEdit(cleanupCtx, options.PackageName, edit.ID); cleanupErr != nil {
				err = errors.Join(err, cleanupErr)
			}
		}
	}()

	versionCode, err := uploadReleaseArtifact(ctx, publisher, options, edit.ID, &result)
	if err != nil {
		return PublishResult{}, err
	}

	release := TrackRelease{
		Name:         options.ReleaseName,
		Status:       options.Status,
		UserFraction: options.UserFraction,
		VersionCodes: []int64{versionCode},
		ReleaseNotes: options.ReleaseNotes,
	}

	if _, err := publisher.AppendTrackRelease(ctx, options.PackageName, edit.ID, plan.Track, release); err != nil {
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

func uploadReleaseArtifact(ctx context.Context, publisher InternalPublisher, options PublishInternalOptions, editID string, result *PublishResult) (int64, error) {
	switch options.ArtifactKind() {
	case ArtifactKindAPK:
		apk, err := publisher.UploadAPK(ctx, options.PackageName, editID, options.APKPath)
		if err != nil {
			return 0, err
		}
		result.APK = &apk
		return apk.VersionCode, nil
	case ArtifactKindBundle:
		bundle, err := publisher.UploadBundle(ctx, options.PackageName, editID, options.BundlePath)
		if err != nil {
			return 0, err
		}
		result.Bundle = &bundle
		return bundle.VersionCode, nil
	default:
		return 0, fmt.Errorf("unsupported artifact kind %q", options.ArtifactKind())
	}
}
