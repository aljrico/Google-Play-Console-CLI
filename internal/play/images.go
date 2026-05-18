package play

import (
	"context"
	"errors"
	"fmt"
)

type ImageType string

const (
	ImageTypeFeatureGraphic       ImageType = "featureGraphic"
	ImageTypeIcon                 ImageType = "icon"
	ImageTypePhoneScreenshots     ImageType = "phoneScreenshots"
	ImageTypeSevenInchScreenshots ImageType = "sevenInchScreenshots"
	ImageTypeTenInchScreenshots   ImageType = "tenInchScreenshots"
	ImageTypeTVBanner             ImageType = "tvBanner"
	ImageTypeTVScreenshots        ImageType = "tvScreenshots"
	ImageTypeWearScreenshots      ImageType = "wearScreenshots"
)

func NewImageType(value string) (ImageType, error) {
	imageType := ImageType(value)
	switch imageType {
	case ImageTypeFeatureGraphic,
		ImageTypeIcon,
		ImageTypePhoneScreenshots,
		ImageTypeSevenInchScreenshots,
		ImageTypeTenInchScreenshots,
		ImageTypeTVBanner,
		ImageTypeTVScreenshots,
		ImageTypeWearScreenshots:
		return imageType, nil
	default:
		return "", fmt.Errorf("unsupported image type %q", value)
	}
}

func (t ImageType) String() string {
	return string(t)
}

func (t ImageType) Validate() error {
	_, err := NewImageType(t.String())
	return err
}

type StoreImage struct {
	ID     string `json:"id,omitempty"`
	URL    string `json:"url,omitempty"`
	SHA1   string `json:"sha1,omitempty"`
	SHA256 string `json:"sha256,omitempty"`
}

type ImageListOptions struct {
	PackageName PackageName     `json:"packageName"`
	Language    ListingLanguage `json:"language"`
	Type        ImageType       `json:"type"`
}

func (o ImageListOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewListingLanguage(o.Language.String()); err != nil {
		return err
	}
	return o.Type.Validate()
}

type ImageListResult struct {
	PackageName PackageName     `json:"packageName"`
	Language    ListingLanguage `json:"language"`
	Type        ImageType       `json:"type"`
	Images      []StoreImage    `json:"images"`
}

type ImageReader interface {
	InsertEdit(ctx context.Context, packageName PackageName) (Edit, error)
	ListImages(ctx context.Context, packageName PackageName, editID string, language ListingLanguage, imageType ImageType) ([]StoreImage, error)
	DeleteEdit(ctx context.Context, packageName PackageName, editID string) error
}

type ImageDeleter interface {
	InsertEdit(ctx context.Context, packageName PackageName) (Edit, error)
	DeleteImage(ctx context.Context, packageName PackageName, editID string, language ListingLanguage, imageType ImageType, imageID string) error
	DeleteAllImages(ctx context.Context, packageName PackageName, editID string, language ListingLanguage, imageType ImageType) ([]StoreImage, error)
	ValidateEdit(ctx context.Context, packageName PackageName, editID string) error
	CommitEdit(ctx context.Context, packageName PackageName, editID string) (Edit, error)
	DeleteEdit(ctx context.Context, packageName PackageName, editID string) error
}

func ListImages(ctx context.Context, reader ImageReader, options ImageListOptions) (result ImageListResult, err error) {
	if err := options.Validate(); err != nil {
		return ImageListResult{}, err
	}
	if reader == nil {
		return ImageListResult{}, fmt.Errorf("image reader is required")
	}

	edit, err := reader.InsertEdit(ctx, options.PackageName)
	if err != nil {
		return ImageListResult{}, err
	}
	defer func() {
		cleanupCtx, cancel := newCleanupContext()
		defer cancel()
		if cleanupErr := reader.DeleteEdit(cleanupCtx, options.PackageName, edit.ID); cleanupErr != nil {
			err = errors.Join(err, cleanupErr)
		}
	}()

	images, err := reader.ListImages(ctx, options.PackageName, edit.ID, options.Language, options.Type)
	if err != nil {
		return ImageListResult{}, err
	}
	return ImageListResult{
		PackageName: options.PackageName,
		Language:    options.Language,
		Type:        options.Type,
		Images:      images,
	}, nil
}

type ImageDeleteOptions struct {
	PackageName PackageName     `json:"packageName"`
	Language    ListingLanguage `json:"language"`
	Type        ImageType       `json:"type"`
	ImageID     string          `json:"imageId,omitempty"`
	All         bool            `json:"all"`
	Confirm     bool            `json:"confirm"`
	DryRun      bool            `json:"dryRun"`
}

func (o ImageDeleteOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewListingLanguage(o.Language.String()); err != nil {
		return err
	}
	if err := o.Type.Validate(); err != nil {
		return err
	}
	if o.All {
		if o.ImageID != "" {
			return fmt.Errorf("image ID cannot be set when deleting all images")
		}
		return nil
	}
	if o.ImageID == "" {
		return fmt.Errorf("image ID is required")
	}
	return nil
}

type ImageDeletePlan struct {
	PackageName PackageName     `json:"packageName"`
	Language    ListingLanguage `json:"language"`
	Type        ImageType       `json:"type"`
	ImageID     string          `json:"imageId,omitempty"`
	All         bool            `json:"all"`
	Confirm     bool            `json:"confirm"`
	Steps       []string        `json:"steps"`
}

type ImageDeleteResult struct {
	PackageName PackageName     `json:"packageName"`
	Language    ListingLanguage `json:"language"`
	Type        ImageType       `json:"type"`
	ImageID     string          `json:"imageId,omitempty"`
	All         bool            `json:"all"`
	DryRun      bool            `json:"dryRun"`
	Committed   bool            `json:"committed"`
	Edit        *Edit           `json:"edit,omitempty"`
	Deleted     []StoreImage    `json:"deleted,omitempty"`
	Plan        ImageDeletePlan `json:"plan"`
}

func NewImageDeletePlan(options ImageDeleteOptions) (ImageDeletePlan, error) {
	if err := options.Validate(); err != nil {
		return ImageDeletePlan{}, err
	}
	target := fmt.Sprintf("all %s images", options.Type)
	if !options.All {
		target = fmt.Sprintf("%s image %s", options.Type, options.ImageID)
	}
	steps := []string{
		"insert edit",
		fmt.Sprintf("delete %s for %s", target, options.Language),
		"validate edit",
	}
	if options.Confirm {
		steps = append(steps, "commit edit")
	} else {
		steps = append(steps, "delete uncommitted edit")
	}
	return ImageDeletePlan{
		PackageName: options.PackageName,
		Language:    options.Language,
		Type:        options.Type,
		ImageID:     options.ImageID,
		All:         options.All,
		Confirm:     options.Confirm,
		Steps:       steps,
	}, nil
}

func DeleteImages(ctx context.Context, deleter ImageDeleter, options ImageDeleteOptions) (result ImageDeleteResult, err error) {
	plan, err := NewImageDeletePlan(options)
	if err != nil {
		return ImageDeleteResult{}, err
	}
	result = ImageDeleteResult{
		PackageName: options.PackageName,
		Language:    options.Language,
		Type:        options.Type,
		ImageID:     options.ImageID,
		All:         options.All,
		DryRun:      options.DryRun,
		Committed:   false,
		Plan:        plan,
	}
	if options.DryRun {
		return result, nil
	}
	if deleter == nil {
		return ImageDeleteResult{}, fmt.Errorf("image deleter is required")
	}

	edit, err := deleter.InsertEdit(ctx, options.PackageName)
	if err != nil {
		return ImageDeleteResult{}, err
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
		deleted, deleteErr := deleter.DeleteAllImages(ctx, options.PackageName, edit.ID, options.Language, options.Type)
		err = deleteErr
		result.Deleted = deleted
	} else {
		err = deleter.DeleteImage(ctx, options.PackageName, edit.ID, options.Language, options.Type, options.ImageID)
	}
	if err != nil {
		return ImageDeleteResult{}, err
	}
	if err := deleter.ValidateEdit(ctx, options.PackageName, edit.ID); err != nil {
		return ImageDeleteResult{}, err
	}
	if !options.Confirm {
		return result, nil
	}

	committedEdit, err := deleter.CommitEdit(ctx, options.PackageName, edit.ID)
	if err != nil {
		return ImageDeleteResult{}, err
	}
	shouldDeleteEdit = false
	result.Edit = &committedEdit
	result.Committed = true
	return result, nil
}
