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

type AppRecoveryTargetingAdder interface {
	AddAppRecoveryTargeting(ctx context.Context, options AppRecoveryTargetingUpdateOptions) error
}

type AppRecoveryCreator interface {
	CreateAppRecovery(ctx context.Context, options AppRecoveryCreateOptions) (AppRecoveryAction, error)
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

type AppRecoveryTargetingUpdateOptions struct {
	PackageName   PackageName   `json:"packageName"`
	AppRecoveryID AppRecoveryID `json:"appRecoveryId"`
	AllUsers      bool          `json:"allUsers,omitempty"`
	SDKLevels     []int64       `json:"sdkLevels,omitempty"`
	RegionCodes   []string      `json:"regionCodes,omitempty"`
	Confirm       bool          `json:"confirm"`
	DryRun        bool          `json:"dryRun"`
}

func (o AppRecoveryTargetingUpdateOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if err := o.AppRecoveryID.Validate(); err != nil {
		return err
	}
	targetingCriteriaCount := 0
	if o.AllUsers {
		targetingCriteriaCount++
	}
	if len(o.SDKLevels) > 0 {
		targetingCriteriaCount++
	}
	if len(o.RegionCodes) > 0 {
		targetingCriteriaCount++
	}
	if targetingCriteriaCount == 0 {
		return fmt.Errorf("app recovery targeting requires --all-users, --sdk-level, or --region")
	}
	if targetingCriteriaCount > 1 {
		return fmt.Errorf("app recovery targeting accepts exactly one of --all-users, --sdk-level, or --region")
	}
	for _, sdkLevel := range o.SDKLevels {
		if sdkLevel <= 0 {
			return fmt.Errorf("SDK level must be greater than 0")
		}
	}
	for _, regionCode := range o.RegionCodes {
		if !isValidAppRecoveryRegionCode(regionCode) {
			return fmt.Errorf("invalid region code %q", regionCode)
		}
	}
	if o.Confirm && o.DryRun {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.Confirm && !o.DryRun {
		return fmt.Errorf("app recovery targeting update requires --confirm or --dry-run")
	}
	return nil
}

func (o AppRecoveryTargetingUpdateOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live app recovery targeting update cannot run with --dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live app recovery targeting update requires --confirm")
	}
	return nil
}

type AppRecoveryTargetingUpdatePlan struct {
	PackageName   PackageName   `json:"packageName"`
	AppRecoveryID AppRecoveryID `json:"appRecoveryId"`
	AllUsers      bool          `json:"allUsers,omitempty"`
	SDKLevels     []int64       `json:"sdkLevels,omitempty"`
	RegionCodes   []string      `json:"regionCodes,omitempty"`
	Confirm       bool          `json:"confirm"`
	Steps         []string      `json:"steps"`
}

type AppRecoveryTargetingUpdateResult struct {
	PackageName   PackageName                    `json:"packageName"`
	AppRecoveryID AppRecoveryID                  `json:"appRecoveryId"`
	DryRun        bool                           `json:"dryRun"`
	Applied       bool                           `json:"applied"`
	Plan          AppRecoveryTargetingUpdatePlan `json:"plan"`
}

func AddAppRecoveryTargeting(ctx context.Context, adder AppRecoveryTargetingAdder, options AppRecoveryTargetingUpdateOptions) (AppRecoveryTargetingUpdateResult, error) {
	if err := options.Validate(); err != nil {
		return AppRecoveryTargetingUpdateResult{}, err
	}
	result := AppRecoveryTargetingUpdateResult{
		PackageName:   options.PackageName,
		AppRecoveryID: options.AppRecoveryID,
		DryRun:        options.DryRun,
		Plan: AppRecoveryTargetingUpdatePlan{
			PackageName:   options.PackageName,
			AppRecoveryID: options.AppRecoveryID,
			AllUsers:      options.AllUsers,
			SDKLevels:     append([]int64(nil), options.SDKLevels...),
			RegionCodes:   append([]string(nil), options.RegionCodes...),
			Confirm:       options.Confirm,
			Steps:         appRecoveryTargetingUpdateSteps(options),
		},
	}
	if options.DryRun {
		return result, nil
	}
	if adder == nil {
		return AppRecoveryTargetingUpdateResult{}, fmt.Errorf("app recovery targeting adder is required")
	}
	if err := adder.AddAppRecoveryTargeting(ctx, options); err != nil {
		return AppRecoveryTargetingUpdateResult{}, err
	}
	result.Applied = true
	return result, nil
}

func appRecoveryTargetingUpdateSteps(options AppRecoveryTargetingUpdateOptions) []string {
	steps := []string{"add app recovery targeting"}
	if options.AllUsers {
		steps = append(steps, "target all users")
	}
	if len(options.SDKLevels) > 0 {
		steps = append(steps, "target android sdk levels")
	}
	if len(options.RegionCodes) > 0 {
		steps = append(steps, "target regions")
	}
	return steps
}

func isValidAppRecoveryRegionCode(value string) bool {
	if len(value) != 2 {
		return false
	}
	return value[0] >= 'A' && value[0] <= 'Z' && value[1] >= 'A' && value[1] <= 'Z'
}

type AppRecoveryCreateOptions struct {
	PackageName      PackageName `json:"packageName"`
	VersionCodes     []int64     `json:"versionCodes,omitempty"`
	VersionCodeStart int64       `json:"versionCodeStart,omitempty"`
	VersionCodeEnd   int64       `json:"versionCodeEnd,omitempty"`
	AllUsers         bool        `json:"allUsers,omitempty"`
	SDKLevels        []int64     `json:"sdkLevels,omitempty"`
	RegionCodes      []string    `json:"regionCodes,omitempty"`
	Confirm          bool        `json:"confirm"`
	DryRun           bool        `json:"dryRun"`
}

func (o AppRecoveryCreateOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if !o.hasVersionTargeting() && !o.hasAudienceTargeting() {
		return fmt.Errorf("app recovery create requires at least one targeting criterion")
	}
	if len(o.VersionCodes) > 0 && (o.VersionCodeStart != 0 || o.VersionCodeEnd != 0) {
		return fmt.Errorf("--version-code cannot be used with --version-code-start or --version-code-end")
	}
	if (o.VersionCodeStart == 0) != (o.VersionCodeEnd == 0) {
		return fmt.Errorf("version range requires both --version-code-start and --version-code-end")
	}
	if o.VersionCodeStart != 0 && o.VersionCodeStart > o.VersionCodeEnd {
		return fmt.Errorf("version code start cannot exceed version code end")
	}
	if o.AllUsers && (len(o.SDKLevels) > 0 || len(o.RegionCodes) > 0) {
		return fmt.Errorf("--all-users cannot be combined with --sdk-level or --region")
	}
	if o.VersionCodeStart < 0 || o.VersionCodeEnd < 0 {
		return fmt.Errorf("version code range values must be greater than 0")
	}
	if o.VersionCodeStart == 0 && o.VersionCodeEnd != 0 {
		return fmt.Errorf("version code start must be greater than 0")
	}
	if o.VersionCodeEnd == 0 && o.VersionCodeStart != 0 {
		return fmt.Errorf("version code end must be greater than 0")
	}
	seenVersionCodes := make(map[int64]struct{}, len(o.VersionCodes))
	for _, versionCode := range o.VersionCodes {
		if versionCode <= 0 {
			return fmt.Errorf("version code must be greater than 0")
		}
		if _, ok := seenVersionCodes[versionCode]; ok {
			return fmt.Errorf("duplicate version code %d", versionCode)
		}
		seenVersionCodes[versionCode] = struct{}{}
	}
	for _, sdkLevel := range o.SDKLevels {
		if sdkLevel <= 0 {
			return fmt.Errorf("SDK level must be greater than 0")
		}
	}
	for _, regionCode := range o.RegionCodes {
		if !isValidAppRecoveryRegionCode(regionCode) {
			return fmt.Errorf("invalid region code %q", regionCode)
		}
	}
	if o.Confirm && o.DryRun {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.Confirm && !o.DryRun {
		return fmt.Errorf("app recovery create requires --confirm or --dry-run")
	}
	return nil
}

func (o AppRecoveryCreateOptions) hasVersionTargeting() bool {
	return len(o.VersionCodes) > 0 || o.VersionCodeStart > 0 || o.VersionCodeEnd > 0
}

func (o AppRecoveryCreateOptions) hasAudienceTargeting() bool {
	return o.AllUsers || len(o.SDKLevels) > 0 || len(o.RegionCodes) > 0
}

func (o AppRecoveryCreateOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live app recovery create cannot run with --dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live app recovery create requires --confirm")
	}
	return nil
}

type AppRecoveryCreatePlan struct {
	PackageName      PackageName `json:"packageName"`
	VersionCodes     []int64     `json:"versionCodes,omitempty"`
	VersionCodeStart int64       `json:"versionCodeStart,omitempty"`
	VersionCodeEnd   int64       `json:"versionCodeEnd,omitempty"`
	AllUsers         bool        `json:"allUsers,omitempty"`
	SDKLevels        []int64     `json:"sdkLevels,omitempty"`
	RegionCodes      []string    `json:"regionCodes,omitempty"`
	Confirm          bool        `json:"confirm"`
	Steps            []string    `json:"steps"`
}

type AppRecoveryCreateResult struct {
	PackageName PackageName           `json:"packageName"`
	DryRun      bool                  `json:"dryRun"`
	Created     bool                  `json:"created"`
	Action      *AppRecoveryAction    `json:"action,omitempty"`
	Plan        AppRecoveryCreatePlan `json:"plan"`
}

func CreateAppRecovery(ctx context.Context, creator AppRecoveryCreator, options AppRecoveryCreateOptions) (AppRecoveryCreateResult, error) {
	if err := options.Validate(); err != nil {
		return AppRecoveryCreateResult{}, err
	}
	result := AppRecoveryCreateResult{
		PackageName: options.PackageName,
		DryRun:      options.DryRun,
		Plan: AppRecoveryCreatePlan{
			PackageName:      options.PackageName,
			VersionCodes:     append([]int64(nil), options.VersionCodes...),
			VersionCodeStart: options.VersionCodeStart,
			VersionCodeEnd:   options.VersionCodeEnd,
			AllUsers:         options.AllUsers,
			SDKLevels:        append([]int64(nil), options.SDKLevels...),
			RegionCodes:      append([]string(nil), options.RegionCodes...),
			Confirm:          options.Confirm,
			Steps:            appRecoveryCreateSteps(options),
		},
	}
	if options.DryRun {
		return result, nil
	}
	if creator == nil {
		return AppRecoveryCreateResult{}, fmt.Errorf("app recovery creator is required")
	}
	action, err := creator.CreateAppRecovery(ctx, options)
	if err != nil {
		return AppRecoveryCreateResult{}, err
	}
	result.Created = true
	result.Action = &action
	return result, nil
}

func appRecoveryCreateSteps(options AppRecoveryCreateOptions) []string {
	steps := []string{"create draft remote in-app update recovery"}
	if len(options.VersionCodes) > 0 {
		steps = append(steps, "target version codes")
	}
	if options.VersionCodeStart > 0 {
		steps = append(steps, "target version code range")
	}
	if options.AllUsers {
		steps = append(steps, "target all users")
	}
	if len(options.SDKLevels) > 0 {
		steps = append(steps, "target android sdk levels")
	}
	if len(options.RegionCodes) > 0 {
		steps = append(steps, "target regions")
	}
	return steps
}
