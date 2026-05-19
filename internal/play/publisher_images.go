package play

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/googleapi"
)

func (p *GooglePublisher) ListImages(ctx context.Context, packageName PackageName, editID string, language ListingLanguage, imageType ImageType) ([]StoreImage, error) {
	response, err := p.service.Edits.Images.List(packageName.String(), editID, language.String(), imageType.String()).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("list %s images for %s %s listing: %w", imageType, packageName, language, err)
	}
	images := make([]StoreImage, 0, len(response.Images))
	for _, apiImage := range response.Images {
		if apiImage == nil {
			continue
		}
		images = append(images, imageFromAPI(apiImage))
	}
	return images, nil
}

func (p *GooglePublisher) UploadImage(ctx context.Context, packageName PackageName, editID string, language ListingLanguage, imageType ImageType, path string) (StoreImage, error) {
	file, err := os.Open(path)
	if err != nil {
		return StoreImage{}, fmt.Errorf("open image %s: %w", path, err)
	}
	defer file.Close()

	response, err := p.service.Edits.Images.Upload(packageName.String(), editID, language.String(), imageType.String()).
		Media(file, googleapi.ContentType(ImageContentType(path))).
		Context(ctx).
		Do()
	if err != nil {
		return StoreImage{}, fmt.Errorf("upload %s image %s for %s %s listing: %w", imageType, path, packageName, language, err)
	}
	if response == nil || response.Image == nil {
		return StoreImage{}, fmt.Errorf("upload %s image %s for %s %s listing produced no image; verify the listing language is supported", imageType, path, packageName, language)
	}
	return imageFromAPI(response.Image), nil
}

func (p *GooglePublisher) DeleteImage(ctx context.Context, packageName PackageName, editID string, language ListingLanguage, imageType ImageType, imageID string) error {
	if err := p.service.Edits.Images.Delete(packageName.String(), editID, language.String(), imageType.String(), imageID).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete %s image %s for %s %s listing: %w", imageType, imageID, packageName, language, err)
	}
	return nil
}

func (p *GooglePublisher) DeleteAllImages(ctx context.Context, packageName PackageName, editID string, language ListingLanguage, imageType ImageType) ([]StoreImage, error) {
	response, err := p.service.Edits.Images.Deleteall(packageName.String(), editID, language.String(), imageType.String()).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("delete all %s images for %s %s listing: %w", imageType, packageName, language, err)
	}
	if response == nil {
		return []StoreImage{}, nil
	}
	images := make([]StoreImage, 0, len(response.Deleted))
	for _, apiImage := range response.Deleted {
		if apiImage == nil {
			continue
		}
		images = append(images, imageFromAPI(apiImage))
	}
	return images, nil
}

func imageFromAPI(apiImage *androidpublisher.Image) StoreImage {
	if apiImage == nil {
		return StoreImage{}
	}
	return StoreImage{
		ID:     apiImage.Id,
		URL:    apiImage.Url,
		SHA1:   apiImage.Sha1,
		SHA256: apiImage.Sha256,
	}
}
