package play

import (
	"context"
	"errors"
	"fmt"
)

type TrackPromoter interface {
	InsertEdit(ctx context.Context, packageName PackageName) (Edit, error)
	PromoteTrackRelease(ctx context.Context, packageName PackageName, editID string, sourceTrack TrackName, targetTrack TrackName, versionCode int64, status ReleaseStatus, userFraction *float64) (TrackRelease, error)
	ValidateEdit(ctx context.Context, packageName PackageName, editID string) error
	CommitEdit(ctx context.Context, packageName PackageName, editID string) (Edit, error)
	DeleteEdit(ctx context.Context, packageName PackageName, editID string) error
}

type PromoteReleaseOptions struct {
	PackageName  PackageName   `json:"packageName"`
	FromTrack    TrackName     `json:"fromTrack"`
	ToTrack      TrackName     `json:"toTrack"`
	VersionCode  int64         `json:"versionCode"`
	Status       ReleaseStatus `json:"status"`
	UserFraction *float64      `json:"userFraction,omitempty"`
	Confirm      bool          `json:"confirm"`
	DryRun       bool          `json:"dryRun"`
}

func (o PromoteReleaseOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewTrackName(o.FromTrack.String()); err != nil {
		return fmt.Errorf("from track: %w", err)
	}
	if _, err := NewTrackName(o.ToTrack.String()); err != nil {
		return fmt.Errorf("to track: %w", err)
	}
	if o.FromTrack == o.ToTrack {
		return fmt.Errorf("from track and to track must be different")
	}
	if o.VersionCode <= 0 {
		return fmt.Errorf("version code is required")
	}
	if _, err := NewReleaseStatus(o.Status.String()); err != nil {
		return err
	}
	switch o.Status {
	case ReleaseStatusInProgress:
		if o.UserFraction == nil {
			return fmt.Errorf("user fraction is required when status is %s", o.Status)
		}
		if *o.UserFraction <= 0 || *o.UserFraction >= 1 {
			return fmt.Errorf("user fraction must be greater than 0 and less than 1")
		}
	case ReleaseStatusHalted:
		if o.UserFraction != nil && (*o.UserFraction <= 0 || *o.UserFraction >= 1) {
			return fmt.Errorf("user fraction must be greater than 0 and less than 1")
		}
	case ReleaseStatusCompleted, ReleaseStatusDraft:
		if o.UserFraction != nil {
			return fmt.Errorf("user fraction can only be set when status is inProgress or halted")
		}
	}
	return nil
}

type PromotePlan struct {
	PackageName  PackageName   `json:"packageName"`
	FromTrack    TrackName     `json:"fromTrack"`
	ToTrack      TrackName     `json:"toTrack"`
	VersionCode  int64         `json:"versionCode"`
	Status       ReleaseStatus `json:"status"`
	UserFraction *float64      `json:"userFraction,omitempty"`
	Confirm      bool          `json:"confirm"`
	Steps        []string      `json:"steps"`
}

func NewPromotePlan(options PromoteReleaseOptions) (PromotePlan, error) {
	if err := options.Validate(); err != nil {
		return PromotePlan{}, err
	}

	steps := []string{
		"insert edit",
		fmt.Sprintf("read %s track", options.FromTrack),
		fmt.Sprintf("copy version code %d to %s track as %s", options.VersionCode, options.ToTrack, options.Status),
		"validate edit",
	}
	if options.Confirm {
		steps = append(steps, "commit edit")
	} else {
		steps = append(steps, "delete uncommitted edit")
	}

	return PromotePlan{
		PackageName:  options.PackageName,
		FromTrack:    options.FromTrack,
		ToTrack:      options.ToTrack,
		VersionCode:  options.VersionCode,
		Status:       options.Status,
		UserFraction: options.UserFraction,
		Confirm:      options.Confirm,
		Steps:        steps,
	}, nil
}

type PromoteResult struct {
	PackageName PackageName   `json:"packageName"`
	FromTrack   TrackName     `json:"fromTrack"`
	ToTrack     TrackName     `json:"toTrack"`
	DryRun      bool          `json:"dryRun"`
	Committed   bool          `json:"committed"`
	Edit        *Edit         `json:"edit,omitempty"`
	Release     *TrackRelease `json:"release,omitempty"`
	Plan        PromotePlan   `json:"plan"`
}

func PromoteRelease(ctx context.Context, promoter TrackPromoter, options PromoteReleaseOptions) (result PromoteResult, err error) {
	plan, err := NewPromotePlan(options)
	if err != nil {
		return PromoteResult{}, err
	}
	result = PromoteResult{
		PackageName: options.PackageName,
		FromTrack:   options.FromTrack,
		ToTrack:     options.ToTrack,
		DryRun:      options.DryRun,
		Committed:   false,
		Plan:        plan,
	}
	if options.DryRun {
		return result, nil
	}
	if promoter == nil {
		return PromoteResult{}, fmt.Errorf("promoter is required")
	}

	edit, err := promoter.InsertEdit(ctx, options.PackageName)
	if err != nil {
		return PromoteResult{}, err
	}
	result.Edit = &edit

	shouldDeleteEdit := true
	defer func() {
		if shouldDeleteEdit {
			cleanupCtx, cancel := newCleanupContext()
			defer cancel()
			if cleanupErr := promoter.DeleteEdit(cleanupCtx, options.PackageName, edit.ID); cleanupErr != nil {
				err = errors.Join(err, cleanupErr)
			}
		}
	}()

	release, err := promoter.PromoteTrackRelease(ctx, options.PackageName, edit.ID, options.FromTrack, options.ToTrack, options.VersionCode, options.Status, options.UserFraction)
	if err != nil {
		return PromoteResult{}, err
	}
	result.Release = &release

	if err := promoter.ValidateEdit(ctx, options.PackageName, edit.ID); err != nil {
		return PromoteResult{}, err
	}
	if !options.Confirm {
		return result, nil
	}

	committedEdit, err := promoter.CommitEdit(ctx, options.PackageName, edit.ID)
	if err != nil {
		return PromoteResult{}, err
	}
	shouldDeleteEdit = false
	result.Edit = &committedEdit
	result.Committed = true
	return result, nil
}
