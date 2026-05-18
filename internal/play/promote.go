package play

import (
	"context"
	"errors"
	"fmt"
)

type TrackPromoter interface {
	InsertEdit(ctx context.Context, packageName PackageName) (Edit, error)
	PromoteTrackRelease(ctx context.Context, packageName PackageName, editID string, sourceTrack TrackName, targetTrack TrackName, releaseName string) (TrackRelease, error)
	ValidateEdit(ctx context.Context, packageName PackageName, editID string) error
	CommitEdit(ctx context.Context, packageName PackageName, editID string) (Edit, error)
	DeleteEdit(ctx context.Context, packageName PackageName, editID string) error
}

type PromoteReleaseOptions struct {
	PackageName PackageName `json:"packageName"`
	FromTrack   TrackName   `json:"fromTrack"`
	ToTrack     TrackName   `json:"toTrack"`
	ReleaseName string      `json:"releaseName,omitempty"`
	Confirm     bool        `json:"confirm"`
	DryRun      bool        `json:"dryRun"`
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
	return nil
}

type PromotePlan struct {
	PackageName PackageName `json:"packageName"`
	FromTrack   TrackName   `json:"fromTrack"`
	ToTrack     TrackName   `json:"toTrack"`
	ReleaseName string      `json:"releaseName,omitempty"`
	Confirm     bool        `json:"confirm"`
	Steps       []string    `json:"steps"`
}

func NewPromotePlan(options PromoteReleaseOptions) (PromotePlan, error) {
	if err := options.Validate(); err != nil {
		return PromotePlan{}, err
	}

	steps := []string{
		"insert edit",
		fmt.Sprintf("read %s track", options.FromTrack),
		fmt.Sprintf("append release to %s track", options.ToTrack),
		"validate edit",
	}
	if options.Confirm {
		steps = append(steps, "commit edit")
	} else {
		steps = append(steps, "delete uncommitted edit")
	}

	return PromotePlan{
		PackageName: options.PackageName,
		FromTrack:   options.FromTrack,
		ToTrack:     options.ToTrack,
		ReleaseName: options.ReleaseName,
		Confirm:     options.Confirm,
		Steps:       steps,
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

	release, err := promoter.PromoteTrackRelease(ctx, options.PackageName, edit.ID, options.FromTrack, options.ToTrack, options.ReleaseName)
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
