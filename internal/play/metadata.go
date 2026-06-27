package play

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/mail"
	"net/url"
	"os"
	"sort"
	"strings"
	"unicode/utf8"
)

const (
	listingTitleMaxRunes            = 30
	listingShortDescriptionMaxRunes = 80
	listingFullDescriptionMaxRunes  = 4000
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
	PackageName             PackageName `json:"packageName"`
	FilePath                string      `json:"filePath"`
	Details                 *AppDetails `json:"details,omitempty"`
	Listings                []Listing   `json:"listings,omitempty"`
	Confirm                 bool        `json:"confirm"`
	DryRun                  bool        `json:"dryRun"`
	ChangesNotSentForReview bool        `json:"changesNotSentForReview,omitempty"`
}

func (o MetadataApplyOptions) Validate() error {
	if err := o.ValidateRequest(); err != nil {
		return err
	}
	if o.Details == nil && len(o.Listings) == 0 {
		return fmt.Errorf("metadata file must include details or at least one listing")
	}
	if o.Details != nil && !hasAppDetailsFields(*o.Details) && len(o.Listings) == 0 {
		return fmt.Errorf("metadata file must include at least one details field or listing field")
	}
	if o.Details != nil {
		if err := validateMetadataDetails(*o.Details); err != nil {
			return err
		}
	}
	seenLanguages := make(map[ListingLanguage]struct{}, len(o.Listings))
	for _, listing := range o.Listings {
		language, err := NewListingLanguage(listing.Language.String())
		if err != nil {
			return err
		}
		if !hasListingFields(listing) {
			return fmt.Errorf("listing %s must include at least one listing field", listing.Language)
		}
		if err := validateMetadataListing(listing); err != nil {
			return err
		}
		if _, ok := seenLanguages[language]; ok {
			return fmt.Errorf("duplicate listing language %s", language)
		}
		seenLanguages[language] = struct{}{}
	}
	return nil
}

func (o MetadataApplyOptions) ValidateRequest() error {
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
	return nil
}

type MetadataApplyPlan struct {
	PackageName             PackageName `json:"packageName"`
	FilePath                string      `json:"filePath"`
	Confirm                 bool        `json:"confirm"`
	ChangesNotSentForReview bool        `json:"changesNotSentForReview,omitempty"`
	Steps                   []string    `json:"steps"`
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
	UpsertListing(ctx context.Context, packageName PackageName, editID string, listing Listing) (Listing, error)
	ValidateEdit(ctx context.Context, packageName PackageName, editID string) error
	CommitEditWithOptions(ctx context.Context, packageName PackageName, editID string, opts CommitEditOptions) (Edit, error)
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
		steps = append(steps, fmt.Sprintf("update %s listing", listing.Language))
	}
	if !normalizedOptions.ChangesNotSentForReview {
		steps = append(steps, "validate edit")
	}
	if normalizedOptions.Confirm {
		if normalizedOptions.ChangesNotSentForReview {
			steps = append(steps, "commit edit (changes not sent for review)")
		} else {
			steps = append(steps, "commit edit")
		}
	} else {
		steps = append(steps, "delete uncommitted edit")
	}
	return MetadataApplyPlan{
		PackageName:             normalizedOptions.PackageName,
		FilePath:                normalizedOptions.FilePath,
		Confirm:                 normalizedOptions.Confirm,
		ChangesNotSentForReview: normalizedOptions.ChangesNotSentForReview,
		Steps:                   steps,
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
		appliedListing, err := applier.UpsertListing(ctx, normalizedOptions.PackageName, edit.ID, listing)
		if err != nil {
			return MetadataApplyResult{}, err
		}
		appliedListings = append(appliedListings, appliedListing)
	}
	result.Listings = appliedListings

	// edits.validate rejects the same enforcement state that edits.commit hits, so skip it
	// when the caller already accepts the no-auto-review path.
	if !normalizedOptions.ChangesNotSentForReview {
		if err := applier.ValidateEdit(ctx, normalizedOptions.PackageName, edit.ID); err != nil {
			return MetadataApplyResult{}, err
		}
	}
	if !normalizedOptions.Confirm {
		return result, nil
	}
	commitOpts := CommitEditOptions{ChangesNotSentForReview: normalizedOptions.ChangesNotSentForReview}
	committedEdit, err := applier.CommitEditWithOptions(ctx, normalizedOptions.PackageName, edit.ID, commitOpts)
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
	if options.Details != nil {
		details := *options.Details
		if details.DefaultLanguage != nil {
			defaultLanguage, err := NewListingLanguage(*details.DefaultLanguage)
			if err != nil {
				return MetadataApplyOptions{}, err
			}
			normalizedDefaultLanguage := defaultLanguage.String()
			details.DefaultLanguage = &normalizedDefaultLanguage
		}
		normalizedOptions.Details = &details
	}
	normalizedOptions.Listings = append([]Listing(nil), options.Listings...)
	for index, listing := range normalizedOptions.Listings {
		language, err := NewListingLanguage(listing.Language.String())
		if err != nil {
			return MetadataApplyOptions{}, err
		}
		listing.Language = language
		normalizedOptions.Listings[index] = listing
	}
	sort.Slice(normalizedOptions.Listings, func(left int, right int) bool {
		return normalizedOptions.Listings[left].Language < normalizedOptions.Listings[right].Language
	})
	return normalizedOptions, nil
}

func validateMetadataDetails(details AppDetails) error {
	if details.DefaultLanguage != nil {
		if _, err := NewListingLanguage(*details.DefaultLanguage); err != nil {
			return fmt.Errorf("invalid defaultLanguage: %w", err)
		}
	}
	if details.ContactWebsite != nil {
		if err := validateHTTPSURL("contactWebsite", *details.ContactWebsite); err != nil {
			return err
		}
	}
	if details.ContactEmail != nil {
		if strings.TrimSpace(*details.ContactEmail) != *details.ContactEmail || *details.ContactEmail == "" {
			return fmt.Errorf("contactEmail must be a valid email address")
		}
		address, err := mail.ParseAddress(*details.ContactEmail)
		if err != nil || address.Address != *details.ContactEmail {
			return fmt.Errorf("contactEmail must be a valid email address")
		}
	}
	if details.ContactPhone != nil && strings.TrimSpace(*details.ContactPhone) != *details.ContactPhone {
		return fmt.Errorf("contactPhone cannot have leading or trailing whitespace")
	}
	return nil
}

func validateMetadataListing(listing Listing) error {
	if listing.Title != nil {
		if err := validateRequiredTextField("title", *listing.Title, listingTitleMaxRunes); err != nil {
			return fmt.Errorf("listing %s %w", listing.Language, err)
		}
	}
	if listing.ShortDescription != nil {
		if err := validateRequiredTextField("shortDescription", *listing.ShortDescription, listingShortDescriptionMaxRunes); err != nil {
			return fmt.Errorf("listing %s %w", listing.Language, err)
		}
	}
	if listing.FullDescription != nil {
		if err := validateRequiredTextField("fullDescription", *listing.FullDescription, listingFullDescriptionMaxRunes); err != nil {
			return fmt.Errorf("listing %s %w", listing.Language, err)
		}
	}
	if listing.Video != nil && *listing.Video != "" {
		if err := validateHTTPSURL("video", *listing.Video); err != nil {
			return fmt.Errorf("listing %s %w", listing.Language, err)
		}
		if !isYouTubeHost(*listing.Video) {
			return fmt.Errorf("listing %s video must use a youtube.com or youtu.be URL", listing.Language)
		}
	}
	return nil
}

func validateRequiredTextField(name string, value string, maxRunes int) error {
	if strings.TrimSpace(value) != value || value == "" {
		return fmt.Errorf("%s must be non-empty and cannot have leading or trailing whitespace", name)
	}
	if utf8.RuneCountInString(value) > maxRunes {
		return fmt.Errorf("%s must be at most %d characters", name, maxRunes)
	}
	return nil
}

func validateHTTPSURL(name string, value string) error {
	if strings.TrimSpace(value) != value || value == "" {
		return fmt.Errorf("%s must be a valid http or https URL", name)
	}
	parsedURL, err := url.ParseRequestURI(value)
	if err != nil {
		return fmt.Errorf("%s must be a valid http or https URL", name)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("%s must use http or https", name)
	}
	if parsedURL.Host == "" {
		return fmt.Errorf("%s must include a host", name)
	}
	return nil
}

func isYouTubeHost(value string) bool {
	parsedURL, err := url.Parse(value)
	if err != nil {
		return false
	}
	host := strings.ToLower(parsedURL.Hostname())
	return host == "youtu.be" || host == "youtube.com" || strings.HasSuffix(host, ".youtube.com")
}

func hasAppDetailsFields(details AppDetails) bool {
	return details.DefaultLanguage != nil || details.ContactWebsite != nil || details.ContactEmail != nil || details.ContactPhone != nil
}

func hasListingFields(listing Listing) bool {
	return listing.Title != nil || listing.ShortDescription != nil || listing.FullDescription != nil || listing.Video != nil
}
