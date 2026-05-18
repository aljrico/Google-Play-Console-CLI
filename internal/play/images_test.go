package play

import (
	"context"
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
