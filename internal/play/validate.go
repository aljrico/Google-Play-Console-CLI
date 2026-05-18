package play

import (
	"context"
	"errors"
	"fmt"
)

type EditValidator interface {
	InsertEdit(ctx context.Context, packageName PackageName) (Edit, error)
	ValidateEdit(ctx context.Context, packageName PackageName, editID string) error
	DeleteEdit(ctx context.Context, packageName PackageName, editID string) error
}

type ValidateOptions struct {
	PackageName PackageName `json:"packageName"`
}

func (o ValidateOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	return nil
}

type ValidateResult struct {
	PackageName PackageName `json:"packageName"`
	Edit        Edit        `json:"edit"`
	Valid       bool        `json:"valid"`
	Deleted     bool        `json:"deleted"`
}

func Validate(ctx context.Context, validator EditValidator, options ValidateOptions) (result ValidateResult, err error) {
	if err := options.Validate(); err != nil {
		return ValidateResult{}, err
	}
	if validator == nil {
		return ValidateResult{}, fmt.Errorf("edit validator is required")
	}
	edit, err := validator.InsertEdit(ctx, options.PackageName)
	if err != nil {
		return ValidateResult{}, err
	}
	result = ValidateResult{
		PackageName: options.PackageName,
		Edit:        edit,
	}
	defer func() {
		cleanupCtx, cancel := newCleanupContext()
		defer cancel()
		if cleanupErr := validator.DeleteEdit(cleanupCtx, options.PackageName, edit.ID); cleanupErr != nil {
			err = errors.Join(err, cleanupErr)
			return
		}
		result.Deleted = true
	}()

	if err := validator.ValidateEdit(ctx, options.PackageName, edit.ID); err != nil {
		return result, err
	}
	result.Valid = true
	return result, nil
}
