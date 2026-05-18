package play

import (
	"context"
	"errors"
	"fmt"
)

type AppDetails struct {
	DefaultLanguage *string `json:"defaultLanguage,omitempty"`
	ContactWebsite  *string `json:"contactWebsite,omitempty"`
	ContactEmail    *string `json:"contactEmail,omitempty"`
	ContactPhone    *string `json:"contactPhone,omitempty"`
}

type DetailsReader interface {
	InsertEdit(ctx context.Context, packageName PackageName) (Edit, error)
	GetAppDetails(ctx context.Context, packageName PackageName, editID string) (AppDetails, error)
	DeleteEdit(ctx context.Context, packageName PackageName, editID string) error
}

func GetAppDetails(ctx context.Context, reader DetailsReader, packageName PackageName) (details AppDetails, err error) {
	if err := packageName.Validate(); err != nil {
		return AppDetails{}, err
	}
	if reader == nil {
		return AppDetails{}, fmt.Errorf("details reader is required")
	}

	edit, err := reader.InsertEdit(ctx, packageName)
	if err != nil {
		return AppDetails{}, err
	}
	defer func() {
		cleanupCtx, cancel := newCleanupContext()
		defer cancel()
		if cleanupErr := reader.DeleteEdit(cleanupCtx, packageName, edit.ID); cleanupErr != nil {
			err = errors.Join(err, cleanupErr)
		}
	}()

	return reader.GetAppDetails(ctx, packageName, edit.ID)
}

type DetailsUpdater interface {
	InsertEdit(ctx context.Context, packageName PackageName) (Edit, error)
	PatchAppDetails(ctx context.Context, packageName PackageName, editID string, details AppDetails) (AppDetails, error)
	ValidateEdit(ctx context.Context, packageName PackageName, editID string) error
	CommitEdit(ctx context.Context, packageName PackageName, editID string) (Edit, error)
	DeleteEdit(ctx context.Context, packageName PackageName, editID string) error
}

type UpdateDetailsOptions struct {
	PackageName PackageName `json:"packageName"`
	Details     AppDetails  `json:"details"`
	Confirm     bool        `json:"confirm"`
	DryRun      bool        `json:"dryRun"`
}

func (o UpdateDetailsOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if o.Details.DefaultLanguage == nil && o.Details.ContactWebsite == nil && o.Details.ContactEmail == nil && o.Details.ContactPhone == nil {
		return fmt.Errorf("at least one details field is required")
	}
	return nil
}

type UpdateDetailsPlan struct {
	PackageName PackageName `json:"packageName"`
	Confirm     bool        `json:"confirm"`
	Steps       []string    `json:"steps"`
}

type UpdateDetailsResult struct {
	PackageName PackageName       `json:"packageName"`
	DryRun      bool              `json:"dryRun"`
	Committed   bool              `json:"committed"`
	Edit        *Edit             `json:"edit,omitempty"`
	Details     AppDetails        `json:"details"`
	Plan        UpdateDetailsPlan `json:"plan"`
}

func NewUpdateDetailsPlan(options UpdateDetailsOptions) (UpdateDetailsPlan, error) {
	if err := options.Validate(); err != nil {
		return UpdateDetailsPlan{}, err
	}
	steps := []string{
		"insert edit",
		"patch app details",
		"validate edit",
	}
	if options.Confirm {
		steps = append(steps, "commit edit")
	} else {
		steps = append(steps, "delete uncommitted edit")
	}
	return UpdateDetailsPlan{PackageName: options.PackageName, Confirm: options.Confirm, Steps: steps}, nil
}

func UpdateAppDetails(ctx context.Context, updater DetailsUpdater, options UpdateDetailsOptions) (result UpdateDetailsResult, err error) {
	plan, err := NewUpdateDetailsPlan(options)
	if err != nil {
		return UpdateDetailsResult{}, err
	}
	result = UpdateDetailsResult{
		PackageName: options.PackageName,
		DryRun:      options.DryRun,
		Committed:   false,
		Details:     options.Details,
		Plan:        plan,
	}
	if options.DryRun {
		return result, nil
	}
	if updater == nil {
		return UpdateDetailsResult{}, fmt.Errorf("details updater is required")
	}

	edit, err := updater.InsertEdit(ctx, options.PackageName)
	if err != nil {
		return UpdateDetailsResult{}, err
	}
	result.Edit = &edit

	shouldDeleteEdit := true
	defer func() {
		if shouldDeleteEdit {
			cleanupCtx, cancel := newCleanupContext()
			defer cancel()
			if cleanupErr := updater.DeleteEdit(cleanupCtx, options.PackageName, edit.ID); cleanupErr != nil {
				err = errors.Join(err, cleanupErr)
			}
		}
	}()

	details, err := updater.PatchAppDetails(ctx, options.PackageName, edit.ID, options.Details)
	if err != nil {
		return UpdateDetailsResult{}, err
	}
	result.Details = details

	if err := updater.ValidateEdit(ctx, options.PackageName, edit.ID); err != nil {
		return UpdateDetailsResult{}, err
	}
	if !options.Confirm {
		return result, nil
	}

	committedEdit, err := updater.CommitEdit(ctx, options.PackageName, edit.ID)
	if err != nil {
		return UpdateDetailsResult{}, err
	}
	shouldDeleteEdit = false
	result.Edit = &committedEdit
	result.Committed = true
	return result, nil
}
