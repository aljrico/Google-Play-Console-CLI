package play

import (
	"context"
	"fmt"
)

type AppRecoveryListOptions struct {
	PackageName PackageName `json:"packageName"`
	VersionCode int64       `json:"versionCode"`
}

func (o AppRecoveryListOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if o.VersionCode <= 0 {
		return fmt.Errorf("version code must be greater than 0")
	}
	return nil
}

type AppRecoveryAction struct {
	ID             int64  `json:"id"`
	Status         string `json:"status,omitempty"`
	CreateTime     string `json:"createTime,omitempty"`
	DeployTime     string `json:"deployTime,omitempty"`
	CancelTime     string `json:"cancelTime,omitempty"`
	LastUpdateTime string `json:"lastUpdateTime,omitempty"`
}

type AppRecoveryListResult struct {
	PackageName PackageName         `json:"packageName"`
	VersionCode int64               `json:"versionCode"`
	Actions     []AppRecoveryAction `json:"actions"`
}

type AppRecoveryLister interface {
	ListAppRecoveries(ctx context.Context, options AppRecoveryListOptions) (AppRecoveryListResult, error)
}

func ListAppRecoveries(ctx context.Context, lister AppRecoveryLister, options AppRecoveryListOptions) (AppRecoveryListResult, error) {
	if err := options.Validate(); err != nil {
		return AppRecoveryListResult{}, err
	}
	if lister == nil {
		return AppRecoveryListResult{}, fmt.Errorf("app recovery lister is required")
	}
	return lister.ListAppRecoveries(ctx, options)
}
