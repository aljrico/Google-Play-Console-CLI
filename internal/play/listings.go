package play

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/text/language"
)

type ListingLanguage string

func NewListingLanguage(value string) (ListingLanguage, error) {
	if value == "" {
		return "", fmt.Errorf("language is required")
	}
	if strings.TrimSpace(value) != value {
		return "", fmt.Errorf("language cannot have leading or trailing whitespace")
	}
	tag, err := language.Parse(value)
	if err != nil || tag.String() == "und" {
		return "", fmt.Errorf("invalid BCP-47 language %q", value)
	}
	return ListingLanguage(tag.String()), nil
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
	UpsertListing(ctx context.Context, packageName PackageName, editID string, listing Listing) (Listing, error)
	ValidateEdit(ctx context.Context, packageName PackageName, editID string) error
	CommitEdit(ctx context.Context, packageName PackageName, editID string) (Edit, error)
	DeleteEdit(ctx context.Context, packageName PackageName, editID string) error
}

type ListingDeleter interface {
	InsertEdit(ctx context.Context, packageName PackageName) (Edit, error)
	DeleteListing(ctx context.Context, packageName PackageName, editID string, language ListingLanguage) error
	DeleteAllListings(ctx context.Context, packageName PackageName, editID string) error
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
		fmt.Sprintf("update %s listing", options.Listing.Language),
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

	listing, err := updater.UpsertListing(ctx, options.PackageName, edit.ID, options.Listing)
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

type DeleteListingOptions struct {
	PackageName PackageName     `json:"packageName"`
	Language    ListingLanguage `json:"language,omitempty"`
	All         bool            `json:"all"`
	Confirm     bool            `json:"confirm"`
	DryRun      bool            `json:"dryRun"`
}

func (o DeleteListingOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if o.All {
		if o.Language != "" {
			return fmt.Errorf("language cannot be set when deleting all listings")
		}
		return nil
	}
	if _, err := NewListingLanguage(o.Language.String()); err != nil {
		return err
	}
	return nil
}

type DeleteListingPlan struct {
	PackageName PackageName     `json:"packageName"`
	Language    ListingLanguage `json:"language,omitempty"`
	All         bool            `json:"all"`
	Confirm     bool            `json:"confirm"`
	Steps       []string        `json:"steps"`
}

type DeleteListingResult struct {
	PackageName PackageName       `json:"packageName"`
	Language    ListingLanguage   `json:"language,omitempty"`
	All         bool              `json:"all"`
	DryRun      bool              `json:"dryRun"`
	Committed   bool              `json:"committed"`
	Edit        *Edit             `json:"edit,omitempty"`
	Plan        DeleteListingPlan `json:"plan"`
}

func NewDeleteListingPlan(options DeleteListingOptions) (DeleteListingPlan, error) {
	if err := options.Validate(); err != nil {
		return DeleteListingPlan{}, err
	}
	target := "all listings"
	if !options.All {
		target = fmt.Sprintf("%s listing", options.Language)
	}
	steps := []string{
		"insert edit",
		fmt.Sprintf("delete %s", target),
		"validate edit",
	}
	if options.Confirm {
		steps = append(steps, "commit edit")
	} else {
		steps = append(steps, "delete uncommitted edit")
	}
	return DeleteListingPlan{
		PackageName: options.PackageName,
		Language:    options.Language,
		All:         options.All,
		Confirm:     options.Confirm,
		Steps:       steps,
	}, nil
}

func DeleteListing(ctx context.Context, deleter ListingDeleter, options DeleteListingOptions) (result DeleteListingResult, err error) {
	plan, err := NewDeleteListingPlan(options)
	if err != nil {
		return DeleteListingResult{}, err
	}
	result = DeleteListingResult{
		PackageName: options.PackageName,
		Language:    options.Language,
		All:         options.All,
		DryRun:      options.DryRun,
		Committed:   false,
		Plan:        plan,
	}
	if options.DryRun {
		return result, nil
	}
	if deleter == nil {
		return DeleteListingResult{}, fmt.Errorf("listing deleter is required")
	}

	edit, err := deleter.InsertEdit(ctx, options.PackageName)
	if err != nil {
		return DeleteListingResult{}, err
	}
	result.Edit = &edit

	shouldDeleteEdit := true
	defer func() {
		if shouldDeleteEdit {
			cleanupCtx, cancel := newCleanupContext()
			defer cancel()
			if cleanupErr := deleter.DeleteEdit(cleanupCtx, options.PackageName, edit.ID); cleanupErr != nil {
				err = errors.Join(err, cleanupErr)
			}
		}
	}()

	if options.All {
		err = deleter.DeleteAllListings(ctx, options.PackageName, edit.ID)
	} else {
		err = deleter.DeleteListing(ctx, options.PackageName, edit.ID, options.Language)
	}
	if err != nil {
		return DeleteListingResult{}, err
	}

	if err := deleter.ValidateEdit(ctx, options.PackageName, edit.ID); err != nil {
		return DeleteListingResult{}, err
	}
	if !options.Confirm {
		return result, nil
	}

	committedEdit, err := deleter.CommitEdit(ctx, options.PackageName, edit.ID)
	if err != nil {
		return DeleteListingResult{}, err
	}
	shouldDeleteEdit = false
	result.Edit = &committedEdit
	result.Committed = true
	return result, nil
}
