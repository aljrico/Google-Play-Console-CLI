package play

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"sort"
	"strings"

	"google.golang.org/api/androidpublisher/v3"
)

type TesterGoogleGroup string

func NewTesterGoogleGroup(value string) (TesterGoogleGroup, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("tester Google Group email is required")
	}
	address, err := mail.ParseAddress(value)
	if err != nil || address.Address != value || address.Name != "" {
		return "", fmt.Errorf("invalid tester Google Group email %q", value)
	}
	return TesterGoogleGroup(value), nil
}

func (g TesterGoogleGroup) String() string {
	return string(g)
}

func (g TesterGoogleGroup) Validate() error {
	_, err := NewTesterGoogleGroup(g.String())
	return err
}

type TrackTesters struct {
	PackageName  PackageName         `json:"packageName"`
	Track        TrackName           `json:"track"`
	GoogleGroups []TesterGoogleGroup `json:"googleGroups"`
}

type TestersGetOptions struct {
	PackageName PackageName `json:"packageName"`
	Track       TrackName   `json:"track"`
}

func (o TestersGetOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewTrackName(o.Track.String()); err != nil {
		return err
	}
	return nil
}

type TestersUpdateOptions struct {
	PackageName  PackageName         `json:"packageName"`
	Track        TrackName           `json:"track"`
	GoogleGroups []TesterGoogleGroup `json:"googleGroups"`
	Clear        bool                `json:"clear"`
	Confirm      bool                `json:"confirm"`
	DryRun       bool                `json:"dryRun"`
}

func (o TestersUpdateOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewTrackName(o.Track.String()); err != nil {
		return err
	}
	if o.Clear && len(o.GoogleGroups) > 0 {
		return fmt.Errorf("--clear cannot be combined with --google-group")
	}
	if !o.Clear && len(o.GoogleGroups) == 0 {
		return fmt.Errorf("at least one --google-group is required unless --clear is set")
	}
	if err := validateTesterGoogleGroups(o.GoogleGroups); err != nil {
		return err
	}
	if o.DryRun && o.Confirm {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.DryRun && !o.Confirm {
		return fmt.Errorf("testers update requires --confirm or --dry-run")
	}
	return nil
}

type TestersUpdatePlan struct {
	PackageName  PackageName         `json:"packageName"`
	Track        TrackName           `json:"track"`
	GoogleGroups []TesterGoogleGroup `json:"googleGroups"`
	Clear        bool                `json:"clear"`
	Confirm      bool                `json:"confirm"`
	Steps        []string            `json:"steps"`
}

type TestersUpdateResult struct {
	PackageName PackageName       `json:"packageName"`
	Track       TrackName         `json:"track"`
	DryRun      bool              `json:"dryRun"`
	Committed   bool              `json:"committed"`
	Edit        *Edit             `json:"edit,omitempty"`
	Testers     *TrackTesters     `json:"testers,omitempty"`
	Desired     TrackTesters      `json:"desiredTesters"`
	Plan        TestersUpdatePlan `json:"plan"`
}

type TestersReader interface {
	InsertEdit(ctx context.Context, packageName PackageName) (Edit, error)
	GetTesters(ctx context.Context, packageName PackageName, editID string, track TrackName) (TrackTesters, error)
	DeleteEdit(ctx context.Context, packageName PackageName, editID string) error
}

type TestersUpdater interface {
	InsertEdit(ctx context.Context, packageName PackageName) (Edit, error)
	UpdateTesters(ctx context.Context, packageName PackageName, editID string, track TrackName, googleGroups []TesterGoogleGroup) (TrackTesters, error)
	ValidateEdit(ctx context.Context, packageName PackageName, editID string) error
	CommitEdit(ctx context.Context, packageName PackageName, editID string) (Edit, error)
	DeleteEdit(ctx context.Context, packageName PackageName, editID string) error
}

func GetTesters(ctx context.Context, reader TestersReader, options TestersGetOptions) (result TrackTesters, err error) {
	if err := options.Validate(); err != nil {
		return TrackTesters{}, err
	}
	if reader == nil {
		return TrackTesters{}, fmt.Errorf("testers reader is required")
	}
	edit, err := reader.InsertEdit(ctx, options.PackageName)
	if err != nil {
		return TrackTesters{}, err
	}
	defer func() {
		cleanupCtx, cancel := newCleanupContext()
		defer cancel()
		if cleanupErr := reader.DeleteEdit(cleanupCtx, options.PackageName, edit.ID); cleanupErr != nil {
			err = errors.Join(err, cleanupErr)
		}
	}()

	testers, err := reader.GetTesters(ctx, options.PackageName, edit.ID, options.Track)
	if err != nil {
		return TrackTesters{}, err
	}
	return testers, nil
}

func UpdateTesters(ctx context.Context, updater TestersUpdater, options TestersUpdateOptions) (result TestersUpdateResult, err error) {
	plan, err := NewTestersUpdatePlan(options)
	if err != nil {
		return TestersUpdateResult{}, err
	}
	desired := TrackTesters{
		PackageName:  options.PackageName,
		Track:        options.Track,
		GoogleGroups: normalizedTesterGoogleGroups(options.GoogleGroups),
	}
	result = TestersUpdateResult{
		PackageName: options.PackageName,
		Track:       options.Track,
		DryRun:      options.DryRun,
		Committed:   false,
		Desired:     desired,
		Plan:        plan,
	}
	if options.DryRun {
		return result, nil
	}
	if updater == nil {
		return TestersUpdateResult{}, fmt.Errorf("testers updater is required")
	}

	edit, err := updater.InsertEdit(ctx, options.PackageName)
	if err != nil {
		return TestersUpdateResult{}, err
	}
	result.Edit = &edit

	shouldDeleteEdit := true
	defer func() {
		if shouldDeleteEdit {
			cleanupCtx, cancel := newCleanupContext()
			defer cancel()
			if cleanupErr := updater.DeleteEdit(cleanupCtx, options.PackageName, edit.ID); cleanupErr != nil {
				err = errors.Join(err, cleanupErr)
			}
		}
	}()

	testers, err := updater.UpdateTesters(ctx, options.PackageName, edit.ID, options.Track, desired.GoogleGroups)
	if err != nil {
		return TestersUpdateResult{}, err
	}
	result.Testers = &testers
	if err := updater.ValidateEdit(ctx, options.PackageName, edit.ID); err != nil {
		return TestersUpdateResult{}, err
	}
	committedEdit, err := updater.CommitEdit(ctx, options.PackageName, edit.ID)
	if err != nil {
		return TestersUpdateResult{}, err
	}
	shouldDeleteEdit = false
	result.Edit = &committedEdit
	result.Committed = true
	return result, nil
}

func NewTestersUpdatePlan(options TestersUpdateOptions) (TestersUpdatePlan, error) {
	if err := options.Validate(); err != nil {
		return TestersUpdatePlan{}, err
	}
	steps := []string{"insert edit", fmt.Sprintf("replace %s tester Google Groups", options.Track), "validate edit"}
	if options.DryRun {
		steps = []string{fmt.Sprintf("plan %s tester Google Group replacement", options.Track)}
	} else {
		steps = append(steps, "commit edit")
	}
	return TestersUpdatePlan{
		PackageName:  options.PackageName,
		Track:        options.Track,
		GoogleGroups: normalizedTesterGoogleGroups(options.GoogleGroups),
		Clear:        options.Clear,
		Confirm:      options.Confirm,
		Steps:        steps,
	}, nil
}

func validateTesterGoogleGroups(googleGroups []TesterGoogleGroup) error {
	seen := map[TesterGoogleGroup]struct{}{}
	for _, googleGroup := range googleGroups {
		canonicalGoogleGroup, err := NewTesterGoogleGroup(googleGroup.String())
		if err != nil {
			return err
		}
		normalized := TesterGoogleGroup(strings.ToLower(canonicalGoogleGroup.String()))
		if _, ok := seen[normalized]; ok {
			return fmt.Errorf("tester Google Group %q is duplicated", googleGroup)
		}
		seen[normalized] = struct{}{}
	}
	return nil
}

func normalizedTesterGoogleGroups(googleGroups []TesterGoogleGroup) []TesterGoogleGroup {
	normalized := make([]TesterGoogleGroup, 0, len(googleGroups))
	for _, googleGroup := range googleGroups {
		normalized = append(normalized, TesterGoogleGroup(strings.ToLower(strings.TrimSpace(googleGroup.String()))))
	}
	sort.Slice(normalized, func(i int, j int) bool {
		return normalized[i] < normalized[j]
	})
	return normalized
}

func testersFromAPI(packageName PackageName, track TrackName, apiTesters *androidpublisher.Testers) TrackTesters {
	if apiTesters == nil {
		return TrackTesters{PackageName: packageName, Track: track, GoogleGroups: []TesterGoogleGroup{}}
	}
	googleGroups := make([]TesterGoogleGroup, 0, len(apiTesters.GoogleGroups))
	for _, googleGroup := range apiTesters.GoogleGroups {
		if googleGroup == "" {
			continue
		}
		googleGroups = append(googleGroups, TesterGoogleGroup(googleGroup))
	}
	return TrackTesters{
		PackageName:  packageName,
		Track:        track,
		GoogleGroups: normalizedTesterGoogleGroups(googleGroups),
	}
}

func testersToAPI(googleGroups []TesterGoogleGroup) *androidpublisher.Testers {
	apiGoogleGroups := make([]string, 0, len(googleGroups))
	for _, googleGroup := range normalizedTesterGoogleGroups(googleGroups) {
		apiGoogleGroups = append(apiGoogleGroups, googleGroup.String())
	}
	return &androidpublisher.Testers{
		GoogleGroups:    apiGoogleGroups,
		ForceSendFields: []string{"GoogleGroups"},
	}
}
