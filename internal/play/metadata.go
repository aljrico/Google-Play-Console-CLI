package play

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
)

type MetadataFile struct {
	Details  *AppDetails `json:"details,omitempty"`
	Listings []Listing   `json:"listings,omitempty"`
}

func LoadMetadataFile(path string) (MetadataFile, error) {
	if path == "" {
		return MetadataFile{}, fmt.Errorf("metadata file path is required")
	}
	if err := ValidateReadableFile(path); err != nil {
		return MetadataFile{}, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return MetadataFile{}, fmt.Errorf("read metadata file %s: %w", path, err)
	}
	return ParseMetadataFile(path, content)
}

func ParseMetadataFile(path string, content []byte) (MetadataFile, error) {
	var metadata MetadataFile
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&metadata); err != nil {
		return MetadataFile{}, fmt.Errorf("parse metadata file %s: %w", path, err)
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return MetadataFile{}, fmt.Errorf("parse metadata file %s: trailing JSON value", path)
	}
	return metadata, nil
}

type MetadataApplyOptions struct {
	PackageName PackageName `json:"packageName"`
	FilePath    string      `json:"filePath"`
	Details     *AppDetails `json:"details,omitempty"`
	Listings    []Listing   `json:"listings,omitempty"`
	Confirm     bool        `json:"confirm"`
	DryRun      bool        `json:"dryRun"`
}

func (o MetadataApplyOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if o.FilePath == "" {
		return fmt.Errorf("metadata file path is required")
	}
	if o.Confirm && o.DryRun {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.Confirm && !o.DryRun {
		return fmt.Errorf("metadata apply requires --confirm or --dry-run")
	}
	if o.Details == nil && len(o.Listings) == 0 {
		return fmt.Errorf("metadata file must include details or at least one listing")
	}
	if o.Details != nil && !hasAppDetailsFields(*o.Details) && len(o.Listings) == 0 {
		return fmt.Errorf("metadata file must include at least one details field or listing field")
	}
	seenLanguages := make(map[ListingLanguage]struct{}, len(o.Listings))
	for _, listing := range o.Listings {
		if _, err := NewListingLanguage(listing.Language.String()); err != nil {
			return err
		}
		if !hasListingFields(listing) {
			return fmt.Errorf("listing %s must include at least one listing field", listing.Language)
		}
		if _, ok := seenLanguages[listing.Language]; ok {
			return fmt.Errorf("duplicate listing language %s", listing.Language)
		}
		seenLanguages[listing.Language] = struct{}{}
	}
	return nil
}

type MetadataApplyPlan struct {
	PackageName PackageName `json:"packageName"`
	FilePath    string      `json:"filePath"`
	Confirm     bool        `json:"confirm"`
	Steps       []string    `json:"steps"`
}

type MetadataApplyResult struct {
	PackageName PackageName       `json:"packageName"`
	FilePath    string            `json:"filePath"`
	DryRun      bool              `json:"dryRun"`
	Committed   bool              `json:"committed"`
	Edit        *Edit             `json:"edit,omitempty"`
	Details     *AppDetails       `json:"details,omitempty"`
	Listings    []Listing         `json:"listings,omitempty"`
	Plan        MetadataApplyPlan `json:"plan"`
}

type MetadataApplier interface {
	InsertEdit(ctx context.Context, packageName PackageName) (Edit, error)
	PatchAppDetails(ctx context.Context, packageName PackageName, editID string, details AppDetails) (AppDetails, error)
	PatchListing(ctx context.Context, packageName PackageName, editID string, listing Listing) (Listing, error)
	ValidateEdit(ctx context.Context, packageName PackageName, editID string) error
	CommitEdit(ctx context.Context, packageName PackageName, editID string) (Edit, error)
	DeleteEdit(ctx context.Context, packageName PackageName, editID string) error
}

func NewMetadataApplyPlan(options MetadataApplyOptions) (MetadataApplyPlan, error) {
	normalizedOptions, err := NormalizeMetadataApplyOptions(options)
	if err != nil {
		return MetadataApplyPlan{}, err
	}
	steps := []string{"insert edit"}
	if normalizedOptions.Details != nil && hasAppDetailsFields(*normalizedOptions.Details) {
		steps = append(steps, "patch app details")
	}
	for _, listing := range normalizedOptions.Listings {
		steps = append(steps, fmt.Sprintf("patch %s listing", listing.Language))
	}
	steps = append(steps, "validate edit")
	if normalizedOptions.Confirm {
		steps = append(steps, "commit edit")
	} else {
		steps = append(steps, "delete uncommitted edit")
	}
	return MetadataApplyPlan{
		PackageName: normalizedOptions.PackageName,
		FilePath:    normalizedOptions.FilePath,
		Confirm:     normalizedOptions.Confirm,
		Steps:       steps,
	}, nil
}

func ApplyMetadata(ctx context.Context, applier MetadataApplier, options MetadataApplyOptions) (result MetadataApplyResult, err error) {
	normalizedOptions, err := NormalizeMetadataApplyOptions(options)
	if err != nil {
		return MetadataApplyResult{}, err
	}
	plan, err := NewMetadataApplyPlan(normalizedOptions)
	if err != nil {
		return MetadataApplyResult{}, err
	}
	result = MetadataApplyResult{
		PackageName: normalizedOptions.PackageName,
		FilePath:    normalizedOptions.FilePath,
		DryRun:      normalizedOptions.DryRun,
		Committed:   false,
		Details:     normalizedOptions.Details,
		Listings:    normalizedOptions.Listings,
		Plan:        plan,
	}
	if normalizedOptions.DryRun {
		return result, nil
	}
	if applier == nil {
		return MetadataApplyResult{}, fmt.Errorf("metadata applier is required")
	}

	edit, err := applier.InsertEdit(ctx, normalizedOptions.PackageName)
	if err != nil {
		return MetadataApplyResult{}, err
	}
	result.Edit = &edit

	shouldDeleteEdit := true
	defer func() {
		if shouldDeleteEdit {
			cleanupCtx, cancel := newCleanupContext()
			defer cancel()
			if cleanupErr := applier.DeleteEdit(cleanupCtx, normalizedOptions.PackageName, edit.ID); cleanupErr != nil {
				err = errors.Join(err, cleanupErr)
			}
		}
	}()

	if normalizedOptions.Details != nil && hasAppDetailsFields(*normalizedOptions.Details) {
		details, err := applier.PatchAppDetails(ctx, normalizedOptions.PackageName, edit.ID, *normalizedOptions.Details)
		if err != nil {
			return MetadataApplyResult{}, err
		}
		result.Details = &details
	}
	appliedListings := make([]Listing, 0, len(normalizedOptions.Listings))
	for _, listing := range normalizedOptions.Listings {
		appliedListing, err := applier.PatchListing(ctx, normalizedOptions.PackageName, edit.ID, listing)
		if err != nil {
			return MetadataApplyResult{}, err
		}
		appliedListings = append(appliedListings, appliedListing)
	}
	result.Listings = appliedListings

	if err := applier.ValidateEdit(ctx, normalizedOptions.PackageName, edit.ID); err != nil {
		return MetadataApplyResult{}, err
	}
	if !normalizedOptions.Confirm {
		return result, nil
	}
	committedEdit, err := applier.CommitEdit(ctx, normalizedOptions.PackageName, edit.ID)
	if err != nil {
		return MetadataApplyResult{}, err
	}
	shouldDeleteEdit = false
	result.Edit = &committedEdit
	result.Committed = true
	return result, nil
}

func NormalizeMetadataApplyOptions(options MetadataApplyOptions) (MetadataApplyOptions, error) {
	if err := options.Validate(); err != nil {
		return MetadataApplyOptions{}, err
	}
	normalizedOptions := options
	normalizedOptions.Listings = append([]Listing(nil), options.Listings...)
	sort.Slice(normalizedOptions.Listings, func(left int, right int) bool {
		return normalizedOptions.Listings[left].Language < normalizedOptions.Listings[right].Language
	})
	return normalizedOptions, nil
}

func hasAppDetailsFields(details AppDetails) bool {
	return details.DefaultLanguage != nil || details.ContactWebsite != nil || details.ContactEmail != nil || details.ContactPhone != nil
}

func hasListingFields(listing Listing) bool {
	return listing.Title != nil || listing.ShortDescription != nil || listing.FullDescription != nil || listing.Video != nil
}
