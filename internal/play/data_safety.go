package play

import (
	"context"
	"fmt"
)

type DataSafetyUpdateOptions struct {
	PackageName  PackageName `json:"packageName"`
	CSVPath      string      `json:"csvPath,omitempty"`
	SafetyLabels string      `json:"safetyLabels,omitempty"`
	Confirm      bool        `json:"confirm"`
	DryRun       bool        `json:"dryRun"`
}

func (o DataSafetyUpdateOptions) Validate() error {
	if err := o.ValidateRequest(); err != nil {
		return err
	}
	if o.SafetyLabels == "" {
		return fmt.Errorf("data safety CSV content is required")
	}
	return nil
}

func (o DataSafetyUpdateOptions) ValidateRequest() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if o.CSVPath == "" {
		return fmt.Errorf("data safety CSV path is required")
	}
	if !o.DryRun && !o.Confirm {
		return fmt.Errorf("data safety update requires --confirm or --dry-run")
	}
	return nil
}

type DataSafetyUpdatePlan struct {
	PackageName PackageName `json:"packageName"`
	CSVPath     string      `json:"csvPath"`
	Confirm     bool        `json:"confirm"`
	Steps       []string    `json:"steps"`
}

type DataSafetyUpdateResult struct {
	PackageName PackageName          `json:"packageName"`
	CSVPath     string               `json:"csvPath"`
	DryRun      bool                 `json:"dryRun"`
	Applied     bool                 `json:"applied"`
	Plan        DataSafetyUpdatePlan `json:"plan"`
}

type DataSafetyUpdater interface {
	UpdateDataSafety(ctx context.Context, packageName PackageName, safetyLabels string) error
}

func NewDataSafetyUpdatePlan(options DataSafetyUpdateOptions) (DataSafetyUpdatePlan, error) {
	if err := options.Validate(); err != nil {
		return DataSafetyUpdatePlan{}, err
	}
	steps := []string{"upload data safety CSV"}
	if options.DryRun {
		steps = []string{"plan data safety CSV upload"}
	}
	return DataSafetyUpdatePlan{
		PackageName: options.PackageName,
		CSVPath:     options.CSVPath,
		Confirm:     options.Confirm,
		Steps:       steps,
	}, nil
}

func UpdateDataSafety(ctx context.Context, updater DataSafetyUpdater, options DataSafetyUpdateOptions) (DataSafetyUpdateResult, error) {
	plan, err := NewDataSafetyUpdatePlan(options)
	if err != nil {
		return DataSafetyUpdateResult{}, err
	}
	result := DataSafetyUpdateResult{
		PackageName: options.PackageName,
		CSVPath:     options.CSVPath,
		DryRun:      options.DryRun,
		Applied:     false,
		Plan:        plan,
	}
	if options.DryRun {
		return result, nil
	}
	if updater == nil {
		return DataSafetyUpdateResult{}, fmt.Errorf("data safety updater is required")
	}
	if err := updater.UpdateDataSafety(ctx, options.PackageName, options.SafetyLabels); err != nil {
		return DataSafetyUpdateResult{}, err
	}
	result.Applied = true
	return result, nil
}
