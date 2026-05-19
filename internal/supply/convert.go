package supply

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
)

type ConvertOptions struct {
	Directory string `json:"directory,omitempty"`
}

type ConvertChangelogsOptions struct {
	Directory   string `json:"directory,omitempty"`
	VersionCode int64  `json:"versionCode,omitempty"`
}

type ConvertImagesOptions struct {
	Directory string `json:"directory,omitempty"`
	Language  string `json:"language,omitempty"`
	Type      string `json:"type,omitempty"`
}

type ChangelogMigration struct {
	Directory  string             `json:"directory"`
	Changelogs []ReleaseChangelog `json:"changelogs"`
}

type ReleaseChangelog struct {
	VersionCode     int64              `json:"versionCode"`
	ReleaseNotes    []play.ReleaseNote `json:"releaseNotes"`
	ReleaseNoteArgs []string           `json:"releaseNoteArgs"`
}

type ImageMigration struct {
	Directory string        `json:"directory"`
	Images    []SupplyImage `json:"images"`
}

type SupplyImage struct {
	Language   play.ListingLanguage `json:"language"`
	Type       play.ImageType       `json:"type"`
	FileName   string               `json:"fileName"`
	Path       string               `json:"path"`
	UploadArgs []string             `json:"uploadArgs"`
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

func ConvertChangelogs(ctx context.Context, options ConvertChangelogsOptions) (ChangelogMigration, error) {
	if options.VersionCode < 0 {
		return ChangelogMigration{}, fmt.Errorf("version code cannot be negative")
	}
	inventory, err := Inspect(ctx, InspectOptions{Directory: options.Directory})
	if err != nil {
		return ChangelogMigration{}, err
	}
	changelogsByVersionCode := make(map[int64][]play.ReleaseNote)
	for _, locale := range inventory.Locales {
		select {
		case <-ctx.Done():
			return ChangelogMigration{}, ctx.Err()
		default:
		}
		for _, file := range locale.Changelogs {
			versionCode, err := changelogVersionCode(file.Name)
			if err != nil {
				return ChangelogMigration{}, err
			}
			if options.VersionCode > 0 && versionCode != options.VersionCode {
				continue
			}
			language, err := play.NewListingLanguage(locale.Language)
			if err != nil {
				return ChangelogMigration{}, fmt.Errorf("convert supply changelog locale %s: %w", locale.Language, err)
			}
			text, err := readSupplyTextFile(file.Path)
			if err != nil {
				return ChangelogMigration{}, err
			}
			note := play.ReleaseNote{Language: language, Text: text}
			changelogsByVersionCode[versionCode] = append(changelogsByVersionCode[versionCode], note)
		}
	}
	changelogs := make([]ReleaseChangelog, 0, len(changelogsByVersionCode))
	for versionCode, releaseNotes := range changelogsByVersionCode {
		sort.Slice(releaseNotes, func(i, j int) bool {
			return releaseNotes[i].Language < releaseNotes[j].Language
		})
		if err := play.ValidateReleaseNotes(releaseNotes); err != nil {
			return ChangelogMigration{}, fmt.Errorf("convert supply changelog %d: %w", versionCode, err)
		}
		changelogs = append(changelogs, ReleaseChangelog{
			VersionCode:     versionCode,
			ReleaseNotes:    releaseNotes,
			ReleaseNoteArgs: releaseNoteArgs(releaseNotes),
		})
	}
	sort.Slice(changelogs, func(i, j int) bool {
		return changelogs[i].VersionCode < changelogs[j].VersionCode
	})
	return ChangelogMigration{Directory: inventory.Directory, Changelogs: changelogs}, nil
}

func ConvertImages(ctx context.Context, options ConvertImagesOptions) (ImageMigration, error) {
	filterLanguage, hasLanguageFilter, err := optionalListingLanguage(options.Language)
	if err != nil {
		return ImageMigration{}, err
	}
	filterType, hasTypeFilter, err := optionalImageType(options.Type)
	if err != nil {
		return ImageMigration{}, err
	}
	inventory, err := Inspect(ctx, InspectOptions{Directory: options.Directory})
	if err != nil {
		return ImageMigration{}, err
	}
	images := make([]SupplyImage, 0)
	for _, locale := range inventory.Locales {
		select {
		case <-ctx.Done():
			return ImageMigration{}, ctx.Err()
		default:
		}
		language, err := play.NewListingLanguage(locale.Language)
		if err != nil {
			return ImageMigration{}, fmt.Errorf("convert supply image locale %s: %w", locale.Language, err)
		}
		if hasLanguageFilter && language != filterLanguage {
			continue
		}
		for _, imageSet := range locale.ImageSets {
			if hasTypeFilter && imageSet.Type != filterType.String() {
				continue
			}
			imageType, err := play.NewImageType(imageSet.Type)
			if err != nil {
				return ImageMigration{}, fmt.Errorf("convert supply image set %s: %w", imageSet.Type, err)
			}
			for _, file := range imageSet.Files {
				if err := play.ValidateReadableImageFile(file.Path); err != nil {
					return ImageMigration{}, err
				}
				images = append(images, SupplyImage{
					Language:   language,
					Type:       imageType,
					FileName:   file.Name,
					Path:       file.Path,
					UploadArgs: imageUploadArgs(language, imageType, file.Path),
				})
			}
		}
	}
	sort.Slice(images, func(i, j int) bool {
		if images[i].Language != images[j].Language {
			return images[i].Language < images[j].Language
		}
		if images[i].Type != images[j].Type {
			return images[i].Type < images[j].Type
		}
		return images[i].Path < images[j].Path
	})
	return ImageMigration{Directory: inventory.Directory, Images: images}, nil
}

func optionalListingLanguage(value string) (play.ListingLanguage, bool, error) {
	if value == "" {
		return "", false, nil
	}
	language, err := play.NewListingLanguage(value)
	if err != nil {
		return "", false, err
	}
	return language, true, nil
}

func optionalImageType(value string) (play.ImageType, bool, error) {
	if value == "" {
		return "", false, nil
	}
	imageType, err := play.NewImageType(value)
	if err != nil {
		return "", false, err
	}
	return imageType, true, nil
}

func changelogVersionCode(name string) (int64, error) {
	if filepath.Ext(name) != ".txt" {
		return 0, fmt.Errorf("supply changelog file %s must be named VERSION_CODE.txt", name)
	}
	versionCodeText := strings.TrimSuffix(name, ".txt")
	versionCode, err := strconv.ParseInt(versionCodeText, 10, 64)
	if err != nil || versionCode <= 0 {
		return 0, fmt.Errorf("supply changelog file %s must be named VERSION_CODE.txt", name)
	}
	return versionCode, nil
}

func releaseNoteArgs(notes []play.ReleaseNote) []string {
	args := make([]string, 0, len(notes))
	for _, note := range notes {
		args = append(args, fmt.Sprintf("%s=%s", note.Language, note.Text))
	}
	return args
}

func imageUploadArgs(language play.ListingLanguage, imageType play.ImageType, path string) []string {
	return []string{"--language", language.String(), "--type", imageType.String(), "--file", path}
}

func readSupplyTextFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read supply metadata file %s: %w", path, err)
	}
	return strings.TrimRight(string(content), "\r\n"), nil
}
