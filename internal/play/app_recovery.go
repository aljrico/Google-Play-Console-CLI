package play

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

type AppRecoveryID string

func NewAppRecoveryID(value string) (AppRecoveryID, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("app recovery ID is required")
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return "", fmt.Errorf("invalid app recovery ID %q", value)
	}
	return AppRecoveryID(strconv.FormatInt(id, 10)), nil
}

func (id AppRecoveryID) Int64() int64 {
	parsedID, _ := strconv.ParseInt(id.String(), 10, 64)
	return parsedID
}

func (id AppRecoveryID) String() string {
	return string(id)
}

func (id AppRecoveryID) Validate() error {
	_, err := NewAppRecoveryID(id.String())
	return err
}

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

type AppRecoveryMutator interface {
	DeployAppRecovery(ctx context.Context, options AppRecoveryMutationOptions) error
	CancelAppRecovery(ctx context.Context, options AppRecoveryMutationOptions) error
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

type AppRecoveryMutationOptions struct {
	PackageName   PackageName   `json:"packageName"`
	AppRecoveryID AppRecoveryID `json:"appRecoveryId"`
	Confirm       bool          `json:"confirm"`
	DryRun        bool          `json:"dryRun"`
}

func (o AppRecoveryMutationOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if err := o.AppRecoveryID.Validate(); err != nil {
		return err
	}
	if o.Confirm && o.DryRun {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.Confirm && !o.DryRun {
		return fmt.Errorf("app recovery mutation requires --confirm or --dry-run")
	}
	return nil
}

func (o AppRecoveryMutationOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live app recovery mutation cannot run with --dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live app recovery mutation requires --confirm")
	}
	return nil
}

type AppRecoveryMutationPlan struct {
	Action        string        `json:"action"`
	PackageName   PackageName   `json:"packageName"`
	AppRecoveryID AppRecoveryID `json:"appRecoveryId"`
	Confirm       bool          `json:"confirm"`
	Steps         []string      `json:"steps"`
}

type AppRecoveryMutationResult struct {
	Action        string                  `json:"action"`
	PackageName   PackageName             `json:"packageName"`
	AppRecoveryID AppRecoveryID           `json:"appRecoveryId"`
	DryRun        bool                    `json:"dryRun"`
	Applied       bool                    `json:"applied"`
	Plan          AppRecoveryMutationPlan `json:"plan"`
}

func DeployAppRecovery(ctx context.Context, mutator AppRecoveryMutator, options AppRecoveryMutationOptions) (AppRecoveryMutationResult, error) {
	return mutateAppRecovery(ctx, mutator, options, "deploy")
}

func CancelAppRecovery(ctx context.Context, mutator AppRecoveryMutator, options AppRecoveryMutationOptions) (AppRecoveryMutationResult, error) {
	return mutateAppRecovery(ctx, mutator, options, "cancel")
}

func mutateAppRecovery(ctx context.Context, mutator AppRecoveryMutator, options AppRecoveryMutationOptions, action string) (AppRecoveryMutationResult, error) {
	if err := options.Validate(); err != nil {
		return AppRecoveryMutationResult{}, err
	}
	result := AppRecoveryMutationResult{
		Action:        action,
		PackageName:   options.PackageName,
		AppRecoveryID: options.AppRecoveryID,
		DryRun:        options.DryRun,
		Plan: AppRecoveryMutationPlan{
			Action:        action,
			PackageName:   options.PackageName,
			AppRecoveryID: options.AppRecoveryID,
			Confirm:       options.Confirm,
			Steps:         []string{action + " app recovery action"},
		},
	}
	if options.DryRun {
		return result, nil
	}
	if mutator == nil {
		return AppRecoveryMutationResult{}, fmt.Errorf("app recovery mutator is required")
	}
	switch action {
	case "deploy":
		if err := mutator.DeployAppRecovery(ctx, options); err != nil {
			return AppRecoveryMutationResult{}, err
		}
	case "cancel":
		if err := mutator.CancelAppRecovery(ctx, options); err != nil {
			return AppRecoveryMutationResult{}, err
		}
	default:
		return AppRecoveryMutationResult{}, fmt.Errorf("unsupported app recovery action %q", action)
	}
	result.Applied = true
	return result, nil
}
