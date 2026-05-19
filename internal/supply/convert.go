package supply

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
)

type ConvertOptions struct {
	Directory string `json:"directory,omitempty"`
}

func Convert(ctx context.Context, options ConvertOptions) (play.MetadataFile, error) {
	inventory, err := Inspect(ctx, InspectOptions{Directory: options.Directory})
	if err != nil {
		return play.MetadataFile{}, err
	}
	metadata := play.MetadataFile{Listings: make([]play.Listing, 0, len(inventory.Locales))}
	for _, locale := range inventory.Locales {
		select {
		case <-ctx.Done():
			return play.MetadataFile{}, ctx.Err()
		default:
		}
		listing, ok, err := convertLocale(locale)
		if err != nil {
			return play.MetadataFile{}, err
		}
		if ok {
			metadata.Listings = append(metadata.Listings, listing)
		}
	}
	return metadata, nil
}

func convertLocale(locale Locale) (play.Listing, bool, error) {
	language, err := play.NewListingLanguage(locale.Language)
	if err != nil {
		return play.Listing{}, false, fmt.Errorf("convert supply locale %s: %w", locale.Language, err)
	}
	listing := play.Listing{Language: language}
	for _, file := range locale.ListingFiles {
		value, err := readSupplyTextFile(file.Path)
		if err != nil {
			return play.Listing{}, false, err
		}
		switch file.Name {
		case "title.txt":
			listing.Title = &value
		case "short_description.txt":
			listing.ShortDescription = &value
		case "full_description.txt":
			listing.FullDescription = &value
		case "video.txt":
			listing.Video = &value
		}
	}
	if listing.Title == nil && listing.ShortDescription == nil && listing.FullDescription == nil && listing.Video == nil {
		return play.Listing{}, false, nil
	}
	return listing, true, nil
}

func readSupplyTextFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read supply metadata file %s: %w", path, err)
	}
	return strings.TrimRight(string(content), "\r\n"), nil
}
