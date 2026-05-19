package supply

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const DefaultMetadataDirectory = "fastlane/metadata/android"

var recognizedListingFiles = map[string]bool{
	"full_description.txt":  true,
	"short_description.txt": true,
	"title.txt":             true,
	"video.txt":             true,
}

type InspectOptions struct {
	Directory string `json:"directory,omitempty"`
}

func (o InspectOptions) targetDirectory() string {
	if o.Directory == "" {
		return DefaultMetadataDirectory
	}
	return o.Directory
}

type Inventory struct {
	Directory string   `json:"directory"`
	Locales   []Locale `json:"locales"`
	Summary   Summary  `json:"summary"`
}

type Summary struct {
	LocaleCount      int `json:"localeCount"`
	ListingFileCount int `json:"listingFileCount"`
	ChangelogCount   int `json:"changelogCount"`
	ImageFileCount   int `json:"imageFileCount"`
	UnknownFileCount int `json:"unknownFileCount"`
}

type Locale struct {
	Language     string     `json:"language"`
	Path         string     `json:"path"`
	ListingFiles []File     `json:"listingFiles,omitempty"`
	Changelogs   []File     `json:"changelogs,omitempty"`
	ImageSets    []ImageSet `json:"imageSets,omitempty"`
	UnknownFiles []File     `json:"unknownFiles,omitempty"`
}

type File struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Size int64  `json:"size"`
}

type ImageSet struct {
	Type  string `json:"type"`
	Path  string `json:"path"`
	Files []File `json:"files"`
}

func Inspect(ctx context.Context, options InspectOptions) (Inventory, error) {
	directory := options.targetDirectory()
	if err := validateDirectory(directory); err != nil {
		return Inventory{}, err
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return Inventory{}, fmt.Errorf("read supply metadata directory %s: %w", directory, err)
	}
	inventory := Inventory{Directory: filepath.ToSlash(directory), Locales: make([]Locale, 0)}
	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return Inventory{}, ctx.Err()
		default:
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return Inventory{}, fmt.Errorf("supply locale cannot be a symlink: %s", filepath.Join(directory, entry.Name()))
		}
		if !entry.IsDir() {
			continue
		}
		locale, err := inspectLocale(filepath.Join(directory, entry.Name()), entry.Name())
		if err != nil {
			return Inventory{}, err
		}
		inventory.Locales = append(inventory.Locales, locale)
	}
	sort.Slice(inventory.Locales, func(i, j int) bool {
		return inventory.Locales[i].Language < inventory.Locales[j].Language
	})
	inventory.Summary = summarize(inventory.Locales)
	return inventory, nil
}

func validateDirectory(directory string) error {
	if err := rejectSymlinkPathComponents(directory); err != nil {
		return err
	}
	info, err := os.Lstat(directory)
	switch {
	case err == nil:
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("supply metadata directory cannot be a symlink: %s", directory)
		}
		if !info.IsDir() {
			return fmt.Errorf("supply metadata path is not a directory: %s", directory)
		}
		return nil
	case os.IsNotExist(err):
		return fmt.Errorf("supply metadata directory does not exist: %s", directory)
	default:
		return fmt.Errorf("inspect supply metadata directory %s: %w", directory, err)
	}
}

func rejectSymlinkPathComponents(path string) error {
	cleanPath := filepath.Clean(path)
	if filepath.IsAbs(cleanPath) {
		return nil
	}
	currentPath := "."
	remainingPath := cleanPath
	separator := string(os.PathSeparator)
	for _, part := range strings.Split(remainingPath, separator) {
		if part == "" || part == "." {
			continue
		}
		currentPath = filepath.Join(currentPath, part)
		info, err := os.Lstat(currentPath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return fmt.Errorf("inspect supply path component %s: %w", currentPath, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("supply metadata path component cannot be a symlink: %s", currentPath)
		}
	}
	return nil
}

func inspectLocale(path string, language string) (Locale, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return Locale{}, fmt.Errorf("read supply locale %s: %w", path, err)
	}
	locale := Locale{Language: language, Path: filepath.ToSlash(path)}
	for _, entry := range entries {
		entryPath := filepath.Join(path, entry.Name())
		switch {
		case entry.IsDir() && entry.Name() == "changelogs":
			changelogs, unknownFiles, err := inspectChangelogs(entryPath, path)
			if err != nil {
				return Locale{}, err
			}
			locale.Changelogs = changelogs
			locale.UnknownFiles = append(locale.UnknownFiles, unknownFiles...)
		case entry.IsDir() && entry.Name() == "images":
			imageSets, unknownFiles, err := inspectImageSets(entryPath, path)
			if err != nil {
				return Locale{}, err
			}
			locale.ImageSets = imageSets
			locale.UnknownFiles = append(locale.UnknownFiles, unknownFiles...)
		case entry.IsDir():
			files, err := inspectTreeFiles(entryPath, path)
			if err != nil {
				return Locale{}, err
			}
			locale.UnknownFiles = append(locale.UnknownFiles, files...)
		case recognizedListingFiles[entry.Name()]:
			file, err := fileInfo(entryPath, entry.Name())
			if err != nil {
				return Locale{}, err
			}
			locale.ListingFiles = append(locale.ListingFiles, file)
		default:
			file, err := fileInfo(entryPath, entry.Name())
			if err != nil {
				return Locale{}, err
			}
			locale.UnknownFiles = append(locale.UnknownFiles, file)
		}
	}
	sortFiles(locale.ListingFiles)
	sortFiles(locale.Changelogs)
	sortFiles(locale.UnknownFiles)
	sort.Slice(locale.ImageSets, func(i, j int) bool {
		return locale.ImageSets[i].Type < locale.ImageSets[j].Type
	})
	return locale, nil
}

func inspectImageSets(path string, localePath string) ([]ImageSet, []File, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, nil, fmt.Errorf("read supply image directory %s: %w", path, err)
	}
	imageSets := make([]ImageSet, 0)
	unknownFiles := make([]File, 0)
	for _, entry := range entries {
		entryPath := filepath.Join(path, entry.Name())
		if entry.Type()&os.ModeSymlink != 0 {
			return nil, nil, fmt.Errorf("supply image set cannot be a symlink: %s", entryPath)
		}
		if !entry.IsDir() {
			file, err := relativeFileInfo(entryPath, localePath)
			if err != nil {
				return nil, nil, err
			}
			unknownFiles = append(unknownFiles, file)
			continue
		}
		files, err := inspectTreeFiles(entryPath, entryPath)
		if err != nil {
			return nil, nil, err
		}
		imageSets = append(imageSets, ImageSet{Type: entry.Name(), Path: filepath.ToSlash(entryPath), Files: files})
	}
	sortFiles(unknownFiles)
	return imageSets, unknownFiles, nil
}

func inspectChangelogs(path string, localePath string) ([]File, []File, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, nil, fmt.Errorf("read supply directory %s: %w", path, err)
	}
	files := make([]File, 0)
	unknownFiles := make([]File, 0)
	for _, entry := range entries {
		entryPath := filepath.Join(path, entry.Name())
		if entry.Type()&os.ModeSymlink != 0 {
			return nil, nil, fmt.Errorf("supply changelog file cannot be a symlink: %s", entryPath)
		}
		if entry.IsDir() {
			dirFiles, err := inspectTreeFiles(entryPath, localePath)
			if err != nil {
				return nil, nil, err
			}
			unknownFiles = append(unknownFiles, dirFiles...)
			continue
		}
		file, err := fileInfo(entryPath, entry.Name())
		if err != nil {
			return nil, nil, err
		}
		files = append(files, file)
	}
	sortFiles(files)
	sortFiles(unknownFiles)
	return files, unknownFiles, nil
}

func inspectTreeFiles(path string, nameRoot string) ([]File, error) {
	files := make([]File, 0)
	err := filepath.WalkDir(path, func(currentPath string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walk supply path %s: %w", currentPath, walkErr)
		}
		if entry.IsDir() {
			return nil
		}
		file, err := relativeFileInfo(currentPath, nameRoot)
		if err != nil {
			return err
		}
		files = append(files, file)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sortFiles(files)
	return files, nil
}

func relativeFileInfo(path string, nameRoot string) (File, error) {
	name, err := filepath.Rel(nameRoot, path)
	if err != nil {
		return File{}, fmt.Errorf("relativize supply file %s: %w", path, err)
	}
	return fileInfo(path, filepath.ToSlash(name))
}

func fileInfo(path string, name string) (File, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return File{}, fmt.Errorf("inspect supply file %s: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return File{}, fmt.Errorf("supply file cannot be a symlink: %s", path)
	}
	if info.IsDir() {
		return File{}, fmt.Errorf("supply file is a directory: %s", path)
	}
	return File{Name: name, Path: filepath.ToSlash(path), Size: info.Size()}, nil
}

func sortFiles(files []File) {
	sort.Slice(files, func(i, j int) bool {
		if files[i].Name == files[j].Name {
			return strings.Compare(files[i].Path, files[j].Path) < 0
		}
		return strings.Compare(files[i].Name, files[j].Name) < 0
	})
}

func summarize(locales []Locale) Summary {
	summary := Summary{LocaleCount: len(locales)}
	for _, locale := range locales {
		summary.ListingFileCount += len(locale.ListingFiles)
		summary.ChangelogCount += len(locale.Changelogs)
		summary.UnknownFileCount += len(locale.UnknownFiles)
		for _, imageSet := range locale.ImageSets {
			summary.ImageFileCount += len(imageSet.Files)
		}
	}
	return summary
}
