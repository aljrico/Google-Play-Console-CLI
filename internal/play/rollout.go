package play

import (
	"context"
	"errors"
	"fmt"
)

type RolloutController interface {
	InsertEdit(ctx context.Context, packageName PackageName) (Edit, error)
	UpdateTrackReleaseStatus(ctx context.Context, packageName PackageName, editID string, track TrackName, versionCode int64, status ReleaseStatus, userFraction *float64) (TrackRelease, error)
	ValidateEdit(ctx context.Context, packageName PackageName, editID string) error
	CommitEdit(ctx context.Context, packageName PackageName, editID string) (Edit, error)
	DeleteEdit(ctx context.Context, packageName PackageName, editID string) error
}

type RolloutAction string

const (
	RolloutActionHalt   RolloutAction = "halt"
	RolloutActionResume RolloutAction = "resume"
)

func (a RolloutAction) String() string {
	return string(a)
}

func (a RolloutAction) Title() string {
	switch a {
	case RolloutActionHalt:
		return "Halt"
	case RolloutActionResume:
		return "Resume"
	default:
		return string(a)
	}
}

type RolloutOptions struct {
	PackageName  PackageName   `json:"packageName"`
	Track        TrackName     `json:"track"`
	VersionCode  int64         `json:"versionCode"`
	Action       RolloutAction `json:"action"`
	Status       ReleaseStatus `json:"status,omitempty"`
	UserFraction *float64      `json:"userFraction,omitempty"`
	Confirm      bool          `json:"confirm"`
	DryRun       bool          `json:"dryRun"`
}

func (o RolloutOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewTrackName(o.Track.String()); err != nil {
		return err
	}
	if o.VersionCode <= 0 {
		return fmt.Errorf("version code is required")
	}
	switch o.Action {
	case RolloutActionHalt:
		if o.Status != "" && o.Status != ReleaseStatusHalted {
			return fmt.Errorf("halt status must be %s", ReleaseStatusHalted)
		}
		if o.UserFraction != nil && (*o.UserFraction <= 0 || *o.UserFraction >= 1) {
			return fmt.Errorf("user fraction must be greater than 0 and less than 1")
		}
	case RolloutActionResume:
		switch o.Status {
		case ReleaseStatusInProgress:
			if o.UserFraction == nil {
				return fmt.Errorf("user fraction is required when resuming a release to %s", ReleaseStatusInProgress)
			}
			if *o.UserFraction <= 0 || *o.UserFraction >= 1 {
				return fmt.Errorf("user fraction must be greater than 0 and less than 1")
			}
		case ReleaseStatusCompleted:
			if o.UserFraction != nil {
				return fmt.Errorf("user fraction can only be set when resuming to %s", ReleaseStatusInProgress)
			}
		case "":
			return fmt.Errorf("status is required when resuming a release")
		default:
			return fmt.Errorf("resume status must be %s or %s", ReleaseStatusInProgress, ReleaseStatusCompleted)
		}
	default:
		return fmt.Errorf("unsupported rollout action %q", o.Action)
	}
	return nil
}

func (o RolloutOptions) targetStatus() ReleaseStatus {
	if o.Action == RolloutActionHalt {
		return ReleaseStatusHalted
	}
	return o.Status
}

type RolloutPlan struct {
	PackageName  PackageName   `json:"packageName"`
	Track        TrackName     `json:"track"`
	VersionCode  int64         `json:"versionCode"`
	Action       RolloutAction `json:"action"`
	Status       ReleaseStatus `json:"status"`
	UserFraction *float64      `json:"userFraction,omitempty"`
	Confirm      bool          `json:"confirm"`
	Steps        []string      `json:"steps"`
}

func NewRolloutPlan(options RolloutOptions) (RolloutPlan, error) {
	if err := options.Validate(); err != nil {
		return RolloutPlan{}, err
	}
	status := options.targetStatus()
	steps := []string{
		"insert edit",
		fmt.Sprintf("read %s track", options.Track),
		fmt.Sprintf("set version code %d to %s", options.VersionCode, status),
		"validate edit",
	}
	if options.Confirm {
		steps = append(steps, "commit edit")
	} else {
		steps = append(steps, "delete uncommitted edit")
	}

	return RolloutPlan{
		PackageName:  options.PackageName,
		Track:        options.Track,
		VersionCode:  options.VersionCode,
		Action:       options.Action,
		Status:       status,
		UserFraction: options.UserFraction,
		Confirm:      options.Confirm,
		Steps:        steps,
	}, nil
}

type RolloutResult struct {
	PackageName PackageName   `json:"packageName"`
	Track       TrackName     `json:"track"`
	VersionCode int64         `json:"versionCode"`
	Action      RolloutAction `json:"action"`
	DryRun      bool          `json:"dryRun"`
	Committed   bool          `json:"committed"`
	Edit        *Edit         `json:"edit,omitempty"`
	Release     *TrackRelease `json:"release,omitempty"`
	Plan        RolloutPlan   `json:"plan"`
}

func UpdateRollout(ctx context.Context, controller RolloutController, options RolloutOptions) (result RolloutResult, err error) {
	plan, err := NewRolloutPlan(options)
	if err != nil {
		return RolloutResult{}, err
	}
	result = RolloutResult{
		PackageName: options.PackageName,
		Track:       options.Track,
		VersionCode: options.VersionCode,
		Action:      options.Action,
		DryRun:      options.DryRun,
		Committed:   false,
		Plan:        plan,
	}
	if options.DryRun {
		return result, nil
	}
	if controller == nil {
		return RolloutResult{}, fmt.Errorf("rollout controller is required")
	}

	edit, err := controller.InsertEdit(ctx, options.PackageName)
	if err != nil {
		return RolloutResult{}, err
	}
	result.Edit = &edit

	shouldDeleteEdit := true
	defer func() {
		if shouldDeleteEdit {
			cleanupCtx, cancel := newCleanupContext()
			defer cancel()
			if cleanupErr := controller.DeleteEdit(cleanupCtx, options.PackageName, edit.ID); cleanupErr != nil {
				err = errors.Join(err, cleanupErr)
			}
		}
	}()

	release, err := controller.UpdateTrackReleaseStatus(ctx, options.PackageName, edit.ID, options.Track, options.VersionCode, plan.Status, options.UserFraction)
	if err != nil {
		return RolloutResult{}, err
	}
	result.Release = &release

	if err := controller.ValidateEdit(ctx, options.PackageName, edit.ID); err != nil {
		return RolloutResult{}, err
	}
	if !options.Confirm {
		return result, nil
	}

	committedEdit, err := controller.CommitEdit(ctx, options.PackageName, edit.ID)
	if err != nil {
		return RolloutResult{}, err
	}
	shouldDeleteEdit = false
	result.Edit = &committedEdit
	result.Committed = true
	return result, nil
}
