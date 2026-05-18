package play

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestListImagesUsesTemporaryEdit(t *testing.T) {
	reader := &fakeImageReader{
		images: []StoreImage{{ID: "image-1", URL: "https://example.com/image.png"}},
	}
	options := ImageListOptions{
		PackageName: "com.example.app",
		Language:    "en-US",
		Type:        ImageTypePhoneScreenshots,
	}

	result, err := ListImages(context.Background(), reader, options)
	if err != nil {
		t.Fatalf("ListImages() error = %v", err)
	}
	if len(result.Images) != 1 {
		t.Fatalf("len(Images) = %d, want 1", len(result.Images))
	}
	wantCalls := []string{"insert", "list", "delete"}
	if !reflect.DeepEqual(reader.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", reader.calls, wantCalls)
	}
	if reader.language != "en-US" || reader.imageType != ImageTypePhoneScreenshots {
		t.Fatalf("language/type = %q/%q, want en-US/phoneScreenshots", reader.language, reader.imageType)
	}
}

func TestListImagesRejectsInvalidOptions(t *testing.T) {
	tests := []ImageListOptions{
		{},
		{PackageName: "bad", Language: "en-US", Type: ImageTypeIcon},
		{PackageName: "com.example.app", Type: ImageTypeIcon},
		{PackageName: "com.example.app", Language: "en-US", Type: "bad"},
	}
	for _, options := range tests {
		_, err := ListImages(context.Background(), nil, options)
		if err == nil {
			t.Fatalf("ListImages(%#v) expected validation error", options)
		}
	}
}

func TestUploadImageDryRunDoesNotRequireUploader(t *testing.T) {
	result, err := UploadImage(context.Background(), nil, ImageUploadOptions{
		PackageName: "com.example.app",
		Language:    "en-US",
		Type:        ImageTypeFeatureGraphic,
		Path:        "feature.png",
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("UploadImage() error = %v", err)
	}
	if !result.DryRun || result.Path != "feature.png" {
		t.Fatalf("result = %#v, want dry-run image upload", result)
	}
}

func TestUploadImageCommitsWithConfirm(t *testing.T) {
	uploader := &fakeImageUploader{image: StoreImage{ID: "image-1"}}
	path := writeTestImage(t, "feature.png")
	result, err := UploadImage(context.Background(), uploader, ImageUploadOptions{
		PackageName: "com.example.app",
		Language:    "en-US",
		Type:        ImageTypeFeatureGraphic,
		Path:        path,
		Confirm:     true,
	})
	if err != nil {
		t.Fatalf("UploadImage() error = %v", err)
	}
	if !result.Committed || result.Image == nil || result.Image.ID != "image-1" {
		t.Fatalf("result = %#v, want committed upload result", result)
	}
	wantCalls := []string{"insert", "upload", "validate", "commit"}
	if !reflect.DeepEqual(uploader.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", uploader.calls, wantCalls)
	}
}

func TestUploadImageRejectsInvalidOptions(t *testing.T) {
	tests := []ImageUploadOptions{
		{},
		{PackageName: "bad", Language: "en-US", Type: ImageTypeIcon, Path: "icon.png", DryRun: true},
		{PackageName: "com.example.app", Type: ImageTypeIcon, Path: "icon.png", DryRun: true},
		{PackageName: "com.example.app", Language: "en-US", Type: "bad", Path: "icon.png", DryRun: true},
		{PackageName: "com.example.app", Language: "en-US", Type: ImageTypeIcon, DryRun: true},
		{PackageName: "com.example.app", Language: "en-US", Type: ImageTypeIcon, Path: "icon.gif", DryRun: true},
	}
	for _, options := range tests {
		_, err := UploadImage(context.Background(), nil, options)
		if err == nil {
			t.Fatalf("UploadImage(%#v) expected validation error", options)
		}
	}
}

func TestDeleteImagesDryRunDoesNotRequireDeleter(t *testing.T) {
	result, err := DeleteImages(context.Background(), nil, ImageDeleteOptions{
		PackageName: "com.example.app",
		Language:    "en-US",
		Type:        ImageTypePhoneScreenshots,
		ImageID:     "image-1",
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("DeleteImages() error = %v", err)
	}
	if !result.DryRun || result.Committed {
		t.Fatalf("result = %#v, want dry-run not committed", result)
	}
	if result.Plan.ImageID != "image-1" || result.Plan.All {
		t.Fatalf("plan = %#v, want single image deletion", result.Plan)
	}
	if result.Deleted == nil || len(result.Deleted) != 0 {
		t.Fatalf("deleted = %#v, want stable empty list", result.Deleted)
	}
}

func TestDeleteAllImagesCommitsWithConfirm(t *testing.T) {
	deleter := &fakeImageDeleter{}
	result, err := DeleteImages(context.Background(), deleter, ImageDeleteOptions{
		PackageName: "com.example.app",
		Language:    "en-US",
		Type:        ImageTypeFeatureGraphic,
		All:         true,
		Confirm:     true,
	})
	if err != nil {
		t.Fatalf("DeleteImages() error = %v", err)
	}
	if !result.Committed {
		t.Fatalf("result = %#v, want committed", result)
	}
	if len(result.Deleted) != 1 || result.Deleted[0].ID != "image-1" {
		t.Fatalf("deleted = %#v, want deleted image", result.Deleted)
	}
	wantCalls := []string{"insert", "delete-all", "validate", "commit"}
	if !reflect.DeepEqual(deleter.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", deleter.calls, wantCalls)
	}
}

func TestDeleteImagesRejectsInvalidOptions(t *testing.T) {
	tests := []ImageDeleteOptions{
		{},
		{PackageName: "bad", Language: "en-US", Type: ImageTypeIcon, ImageID: "image-1", DryRun: true},
		{PackageName: "com.example.app", Type: ImageTypeIcon, ImageID: "image-1", DryRun: true},
		{PackageName: "com.example.app", Language: "en-US", Type: "bad", ImageID: "image-1", DryRun: true},
		{PackageName: "com.example.app", Language: "en-US", Type: ImageTypeIcon, DryRun: true},
		{PackageName: "com.example.app", Language: "en-US", Type: ImageTypeIcon, ImageID: "image-1", All: true, DryRun: true},
	}
	for _, options := range tests {
		_, err := DeleteImages(context.Background(), nil, options)
		if err == nil {
			t.Fatalf("DeleteImages(%#v) expected validation error", options)
		}
	}
}

type fakeImageReader struct {
	calls     []string
	language  ListingLanguage
	imageType ImageType
	images    []StoreImage
}

func (r *fakeImageReader) InsertEdit(ctx context.Context, packageName PackageName) (Edit, error) {
	r.calls = append(r.calls, "insert")
	return Edit{ID: "edit-123"}, nil
}

func (r *fakeImageReader) ListImages(ctx context.Context, packageName PackageName, editID string, language ListingLanguage, imageType ImageType) ([]StoreImage, error) {
	r.calls = append(r.calls, "list")
	r.language = language
	r.imageType = imageType
	return r.images, nil
}

func (r *fakeImageReader) DeleteEdit(ctx context.Context, packageName PackageName, editID string) error {
	r.calls = append(r.calls, "delete")
	return nil
}

func writeTestImage(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte("image"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}

type fakeImageDeleter struct {
	calls []string
}

type fakeImageUploader struct {
	calls []string
	image StoreImage
}

func (u *fakeImageUploader) InsertEdit(ctx context.Context, packageName PackageName) (Edit, error) {
	u.calls = append(u.calls, "insert")
	return Edit{ID: "edit-123"}, nil
}

func (u *fakeImageUploader) UploadImage(ctx context.Context, packageName PackageName, editID string, language ListingLanguage, imageType ImageType, path string) (StoreImage, error) {
	u.calls = append(u.calls, "upload")
	return u.image, nil
}

func (u *fakeImageUploader) ValidateEdit(ctx context.Context, packageName PackageName, editID string) error {
	u.calls = append(u.calls, "validate")
	return nil
}

func (u *fakeImageUploader) CommitEdit(ctx context.Context, packageName PackageName, editID string) (Edit, error) {
	u.calls = append(u.calls, "commit")
	return Edit{ID: editID}, nil
}

func (u *fakeImageUploader) DeleteEdit(ctx context.Context, packageName PackageName, editID string) error {
	u.calls = append(u.calls, "delete-edit")
	return nil
}

func (d *fakeImageDeleter) InsertEdit(ctx context.Context, packageName PackageName) (Edit, error) {
	d.calls = append(d.calls, "insert")
	return Edit{ID: "edit-123"}, nil
}

func (d *fakeImageDeleter) DeleteImage(ctx context.Context, packageName PackageName, editID string, language ListingLanguage, imageType ImageType, imageID string) error {
	d.calls = append(d.calls, "delete-one")
	return nil
}

func (d *fakeImageDeleter) DeleteAllImages(ctx context.Context, packageName PackageName, editID string, language ListingLanguage, imageType ImageType) ([]StoreImage, error) {
	d.calls = append(d.calls, "delete-all")
	return []StoreImage{{ID: "image-1"}}, nil
}

func (d *fakeImageDeleter) ValidateEdit(ctx context.Context, packageName PackageName, editID string) error {
	d.calls = append(d.calls, "validate")
	return nil
}

func (d *fakeImageDeleter) CommitEdit(ctx context.Context, packageName PackageName, editID string) (Edit, error) {
	d.calls = append(d.calls, "commit")
	return Edit{ID: editID}, nil
}

func (d *fakeImageDeleter) DeleteEdit(ctx context.Context, packageName PackageName, editID string) error {
	d.calls = append(d.calls, "delete-edit")
	return nil
}
