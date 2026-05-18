package play

import (
	"context"
	"errors"
	"fmt"
)

type ListingLanguage string

func NewListingLanguage(value string) (ListingLanguage, error) {
	if value == "" {
		return "", fmt.Errorf("language is required")
	}
	return ListingLanguage(value), nil
}

func (l ListingLanguage) String() string {
	return string(l)
}

type Listing struct {
	Language         ListingLanguage `json:"language"`
	Title            *string         `json:"title,omitempty"`
	ShortDescription *string         `json:"shortDescription,omitempty"`
	FullDescription  *string         `json:"fullDescription,omitempty"`
	Video            *string         `json:"video,omitempty"`
}

type ListingReader interface {
	InsertEdit(ctx context.Context, packageName PackageName) (Edit, error)
	ListListings(ctx context.Context, packageName PackageName, editID string) ([]Listing, error)
	GetListing(ctx context.Context, packageName PackageName, editID string, language ListingLanguage) (Listing, error)
	DeleteEdit(ctx context.Context, packageName PackageName, editID string) error
}

func ListListings(ctx context.Context, reader ListingReader, packageName PackageName) (listings []Listing, err error) {
	if err := packageName.Validate(); err != nil {
		return nil, err
	}
	if reader == nil {
		return nil, fmt.Errorf("listing reader is required")
	}

	edit, err := reader.InsertEdit(ctx, packageName)
	if err != nil {
		return nil, err
	}
	defer func() {
		cleanupCtx, cancel := newCleanupContext()
		defer cancel()
		if cleanupErr := reader.DeleteEdit(cleanupCtx, packageName, edit.ID); cleanupErr != nil {
			err = errors.Join(err, cleanupErr)
		}
	}()

	return reader.ListListings(ctx, packageName, edit.ID)
}

func GetListing(ctx context.Context, reader ListingReader, packageName PackageName, language ListingLanguage) (listing Listing, err error) {
	if err := packageName.Validate(); err != nil {
		return Listing{}, err
	}
	if _, err := NewListingLanguage(language.String()); err != nil {
		return Listing{}, err
	}
	if reader == nil {
		return Listing{}, fmt.Errorf("listing reader is required")
	}

	edit, err := reader.InsertEdit(ctx, packageName)
	if err != nil {
		return Listing{}, err
	}
	defer func() {
		cleanupCtx, cancel := newCleanupContext()
		defer cancel()
		if cleanupErr := reader.DeleteEdit(cleanupCtx, packageName, edit.ID); cleanupErr != nil {
			err = errors.Join(err, cleanupErr)
		}
	}()

	return reader.GetListing(ctx, packageName, edit.ID, language)
}

type ListingUpdater interface {
	InsertEdit(ctx context.Context, packageName PackageName) (Edit, error)
	PatchListing(ctx context.Context, packageName PackageName, editID string, listing Listing) (Listing, error)
	ValidateEdit(ctx context.Context, packageName PackageName, editID string) error
	CommitEdit(ctx context.Context, packageName PackageName, editID string) (Edit, error)
	DeleteEdit(ctx context.Context, packageName PackageName, editID string) error
}

type UpdateListingOptions struct {
	PackageName PackageName `json:"packageName"`
	Listing     Listing     `json:"listing"`
	Confirm     bool        `json:"confirm"`
	DryRun      bool        `json:"dryRun"`
}

func (o UpdateListingOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewListingLanguage(o.Listing.Language.String()); err != nil {
		return err
	}
	if o.Listing.Title == nil && o.Listing.ShortDescription == nil && o.Listing.FullDescription == nil && o.Listing.Video == nil {
		return fmt.Errorf("at least one listing field is required")
	}
	return nil
}

type UpdateListingPlan struct {
	PackageName PackageName     `json:"packageName"`
	Language    ListingLanguage `json:"language"`
	Confirm     bool            `json:"confirm"`
	Steps       []string        `json:"steps"`
}

type UpdateListingResult struct {
	PackageName PackageName       `json:"packageName"`
	Language    ListingLanguage   `json:"language"`
	DryRun      bool              `json:"dryRun"`
	Committed   bool              `json:"committed"`
	Edit        *Edit             `json:"edit,omitempty"`
	Listing     Listing           `json:"listing"`
	Plan        UpdateListingPlan `json:"plan"`
}

func NewUpdateListingPlan(options UpdateListingOptions) (UpdateListingPlan, error) {
	if err := options.Validate(); err != nil {
		return UpdateListingPlan{}, err
	}
	steps := []string{
		"insert edit",
		fmt.Sprintf("patch %s listing", options.Listing.Language),
		"validate edit",
	}
	if options.Confirm {
		steps = append(steps, "commit edit")
	} else {
		steps = append(steps, "delete uncommitted edit")
	}

	return UpdateListingPlan{
		PackageName: options.PackageName,
		Language:    options.Listing.Language,
		Confirm:     options.Confirm,
		Steps:       steps,
	}, nil
}

func UpdateListing(ctx context.Context, updater ListingUpdater, options UpdateListingOptions) (result UpdateListingResult, err error) {
	plan, err := NewUpdateListingPlan(options)
	if err != nil {
		return UpdateListingResult{}, err
	}
	result = UpdateListingResult{
		PackageName: options.PackageName,
		Language:    options.Listing.Language,
		DryRun:      options.DryRun,
		Committed:   false,
		Listing:     options.Listing,
		Plan:        plan,
	}
	if options.DryRun {
		return result, nil
	}
	if updater == nil {
		return UpdateListingResult{}, fmt.Errorf("listing updater is required")
	}

	edit, err := updater.InsertEdit(ctx, options.PackageName)
	if err != nil {
		return UpdateListingResult{}, err
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

	listing, err := updater.PatchListing(ctx, options.PackageName, edit.ID, options.Listing)
	if err != nil {
		return UpdateListingResult{}, err
	}
	result.Listing = listing

	if err := updater.ValidateEdit(ctx, options.PackageName, edit.ID); err != nil {
		return UpdateListingResult{}, err
	}
	if !options.Confirm {
		return result, nil
	}

	committedEdit, err := updater.CommitEdit(ctx, options.PackageName, edit.ID)
	if err != nil {
		return UpdateListingResult{}, err
	}
	shouldDeleteEdit = false
	result.Edit = &committedEdit
	result.Committed = true
	return result, nil
}
