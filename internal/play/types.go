package play

import (
	"fmt"
	"path/filepath"
	"strings"
)

type PackageName string

func NewPackageName(value string) (PackageName, error) {
	if value == "" {
		return "", fmt.Errorf("package name is required")
	}
	if !isValidPackageName(value) {
		return "", fmt.Errorf("invalid Android package name %q", value)
	}
	return PackageName(value), nil
}

func (p PackageName) String() string {
	return string(p)
}

func (p PackageName) Validate() error {
	_, err := NewPackageName(p.String())
	return err
}

func isValidPackageName(value string) bool {
	segments := strings.Split(value, ".")
	if len(segments) < 2 {
		return false
	}
	for _, segment := range segments {
		if segment == "" || !isASCIIAlpha(segment[0]) {
			return false
		}
		for i := 1; i < len(segment); i++ {
			character := segment[i]
			if !isASCIIAlpha(character) && !isASCIIDigit(character) && character != '_' {
				return false
			}
		}
	}
	return true
}

func isASCIIAlpha(character byte) bool {
	return (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z')
}

func isASCIIDigit(character byte) bool {
	return character >= '0' && character <= '9'
}

type TrackName string

const (
	TrackInternal   TrackName = "internal"
	TrackAlpha      TrackName = "alpha"
	TrackBeta       TrackName = "beta"
	TrackProduction TrackName = "production"
)

func (t TrackName) String() string {
	return string(t)
}

func NewTrackName(value string) (TrackName, error) {
	if value == "" {
		return "", fmt.Errorf("track name is required")
	}
	return TrackName(value), nil
}

type ReleaseStatus string

const (
	ReleaseStatusCompleted  ReleaseStatus = "completed"
	ReleaseStatusDraft      ReleaseStatus = "draft"
	ReleaseStatusHalted     ReleaseStatus = "halted"
	ReleaseStatusInProgress ReleaseStatus = "inProgress"
)

func NewReleaseStatus(value string) (ReleaseStatus, error) {
	status := ReleaseStatus(value)
	switch status {
	case ReleaseStatusCompleted, ReleaseStatusDraft, ReleaseStatusHalted, ReleaseStatusInProgress:
		return status, nil
	default:
		return "", fmt.Errorf("unsupported release status %q", value)
	}
}

func (s ReleaseStatus) String() string {
	return string(s)
}

type ArtifactKind string

const (
	ArtifactKindBundle ArtifactKind = "bundle"
)

type BundleArtifact struct {
	VersionCode int64  `json:"versionCode"`
	SHA1        string `json:"sha1"`
	SHA256      string `json:"sha256"`
}

type Edit struct {
	ID                string `json:"id"`
	ExpiryTimeSeconds string `json:"expiryTimeSeconds"`
}

type TrackRelease struct {
	Name         string        `json:"name"`
	Status       ReleaseStatus `json:"status"`
	UserFraction *float64      `json:"userFraction,omitempty"`
	VersionCodes []int64       `json:"versionCodes"`
	ReleaseNotes []ReleaseNote `json:"releaseNotes,omitempty"`
}

type Track struct {
	Name     TrackName      `json:"name"`
	Releases []TrackRelease `json:"releases"`
}

type ReleaseNote struct {
	Language ListingLanguage `json:"language"`
	Text     string          `json:"text"`
}

type PublishInternalOptions struct {
	PackageName  PackageName   `json:"packageName"`
	Track        TrackName     `json:"track"`
	BundlePath   string        `json:"bundlePath"`
	ReleaseName  string        `json:"releaseName"`
	Status       ReleaseStatus `json:"status"`
	UserFraction *float64      `json:"userFraction,omitempty"`
	ReleaseNotes []ReleaseNote `json:"releaseNotes,omitempty"`
	Confirm      bool          `json:"confirm"`
	DryRun       bool          `json:"dryRun"`
}

func (o PublishInternalOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewTrackName(targetTrackName(o).String()); err != nil {
		return err
	}
	if o.BundlePath == "" {
		return fmt.Errorf("AAB path is required")
	}
	if filepath.Ext(o.BundlePath) != ".aab" {
		return fmt.Errorf("AAB path must end with .aab")
	}
	if _, err := NewReleaseStatus(o.Status.String()); err != nil {
		return err
	}
	for _, note := range o.ReleaseNotes {
		if _, err := NewListingLanguage(note.Language.String()); err != nil {
			return err
		}
		if note.Text == "" {
			return fmt.Errorf("release note text is required")
		}
	}
	switch o.Status {
	case ReleaseStatusInProgress:
		if o.UserFraction == nil {
			return fmt.Errorf("user fraction is required when status is %s", o.Status)
		}
		if *o.UserFraction <= 0 || *o.UserFraction >= 1 {
			return fmt.Errorf("user fraction must be greater than 0 and less than 1")
		}
	case ReleaseStatusHalted:
		if o.UserFraction != nil && (*o.UserFraction <= 0 || *o.UserFraction >= 1) {
			return fmt.Errorf("user fraction must be greater than 0 and less than 1")
		}
	case ReleaseStatusCompleted, ReleaseStatusDraft:
		if o.UserFraction != nil {
			return fmt.Errorf("user fraction can only be set when status is inProgress or halted")
		}
	}
	return nil
}

type PublishPlan struct {
	PackageName  PackageName   `json:"packageName"`
	Track        TrackName     `json:"track"`
	Artifact     ArtifactKind  `json:"artifact"`
	BundlePath   string        `json:"bundlePath"`
	ReleaseName  string        `json:"releaseName,omitempty"`
	Status       ReleaseStatus `json:"status"`
	UserFraction *float64      `json:"userFraction,omitempty"`
	ReleaseNotes []ReleaseNote `json:"releaseNotes,omitempty"`
	Confirm      bool          `json:"confirm"`
	Steps        []string      `json:"steps"`
}

func NewPublishInternalPlan(options PublishInternalOptions) (PublishPlan, error) {
	if err := options.Validate(); err != nil {
		return PublishPlan{}, err
	}
	trackName := targetTrackName(options)

	steps := []string{
		"insert edit",
		"upload Android App Bundle",
		fmt.Sprintf("append release to %s track", trackName),
		"validate edit",
	}
	if options.Confirm {
		steps = append(steps, "commit edit")
	} else {
		steps = append(steps, "delete uncommitted edit")
	}

	return PublishPlan{
		PackageName:  options.PackageName,
		Track:        trackName,
		Artifact:     ArtifactKindBundle,
		BundlePath:   options.BundlePath,
		ReleaseName:  options.ReleaseName,
		Status:       options.Status,
		UserFraction: options.UserFraction,
		ReleaseNotes: options.ReleaseNotes,
		Confirm:      options.Confirm,
		Steps:        steps,
	}, nil
}

func targetTrackName(options PublishInternalOptions) TrackName {
	if options.Track == "" {
		return TrackInternal
	}
	return options.Track
}

type PublishResult struct {
	PackageName PackageName     `json:"packageName"`
	Track       TrackName       `json:"track"`
	DryRun      bool            `json:"dryRun"`
	Committed   bool            `json:"committed"`
	Edit        *Edit           `json:"edit,omitempty"`
	Bundle      *BundleArtifact `json:"bundle,omitempty"`
	Plan        PublishPlan     `json:"plan"`
}
