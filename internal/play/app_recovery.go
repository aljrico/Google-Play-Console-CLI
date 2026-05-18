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
	AppRecoveryID         string                            `json:"appRecoveryId,omitempty"`
	Status                string                            `json:"status,omitempty"`
	CreateTime            string                            `json:"createTime,omitempty"`
	DeployTime            string                            `json:"deployTime,omitempty"`
	CancelTime            string                            `json:"cancelTime,omitempty"`
	LastUpdateTime        string                            `json:"lastUpdateTime,omitempty"`
	Targeting             *AppRecoveryTargeting             `json:"targeting,omitempty"`
	RemoteInAppUpdateData *AppRecoveryRemoteInAppUpdateData `json:"remoteInAppUpdateData,omitempty"`
}

type AppRecoveryTargeting struct {
	AllUsers     *AppRecoveryAllUsers     `json:"allUsers,omitempty"`
	AndroidSDKs  *AppRecoveryAndroidSDKs  `json:"androidSdks,omitempty"`
	Regions      *AppRecoveryRegions      `json:"regions,omitempty"`
	VersionList  *AppRecoveryVersionList  `json:"versionList,omitempty"`
	VersionRange *AppRecoveryVersionRange `json:"versionRange,omitempty"`
}

type AppRecoveryAllUsers struct {
	IsAllUsersRequested bool `json:"isAllUsersRequested"`
}

type AppRecoveryAndroidSDKs struct {
	SDKLevels []int64 `json:"sdkLevels,omitempty"`
}

type AppRecoveryRegions struct {
	RegionCodes []string `json:"regionCode,omitempty"`
}

type AppRecoveryVersionList struct {
	VersionCodes []int64 `json:"versionCodes,omitempty"`
}

type AppRecoveryVersionRange struct {
	VersionCodeStart string `json:"versionCodeStart,omitempty"`
	VersionCodeEnd   string `json:"versionCodeEnd,omitempty"`
}

type AppRecoveryRemoteInAppUpdateData struct {
	PerBundle []AppRecoveryRemoteInAppUpdateDataPerBundle `json:"remoteAppUpdateDataPerBundle,omitempty"`
}

type AppRecoveryRemoteInAppUpdateDataPerBundle struct {
	VersionCode          string `json:"versionCode,omitempty"`
	RecoveredDeviceCount string `json:"recoveredDeviceCount,omitempty"`
	TotalDeviceCount     string `json:"totalDeviceCount,omitempty"`
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
