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
