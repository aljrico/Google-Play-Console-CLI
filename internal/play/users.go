package play

import (
	"context"
	"fmt"
	"strings"
)

type DeveloperAccount string

func NewDeveloperAccount(value string) (DeveloperAccount, error) {
	if value == "" {
		return "", fmt.Errorf("developer account is required")
	}
	if strings.HasPrefix(value, "developers/") {
		value = strings.TrimPrefix(value, "developers/")
	}
	if value == "" || strings.Contains(value, "/") {
		return "", fmt.Errorf("invalid developer account %q", value)
	}
	return DeveloperAccount(value), nil
}

func (d DeveloperAccount) String() string {
	return string(d)
}

func (d DeveloperAccount) ResourceName() string {
	return "developers/" + d.String()
}

func (d DeveloperAccount) Validate() error {
	_, err := NewDeveloperAccount(d.String())
	return err
}

type UserName string

func NewUserName(value string) (UserName, error) {
	if value == "" {
		return "", fmt.Errorf("user name is required")
	}
	segments := strings.Split(value, "/")
	if len(segments) != 4 || segments[0] != "developers" || segments[2] != "users" {
		return "", fmt.Errorf("invalid user name %q", value)
	}
	if _, err := NewDeveloperAccount(segments[1]); err != nil {
		return "", fmt.Errorf("invalid user developer: %w", err)
	}
	if _, err := NewGrantUserEmail(segments[3]); err != nil {
		return "", fmt.Errorf("invalid user email: %w", err)
	}
	return UserName(value), nil
}

func NewUserNameFromParts(developer DeveloperAccount, email GrantUserEmail) UserName {
	return UserName(developer.ResourceName() + "/users/" + email.String())
}

func (n UserName) String() string {
	return string(n)
}

func (n UserName) Validate() error {
	_, err := NewUserName(n.String())
	return err
}

type UserPermission string

const (
	UserPermissionSeeAllApps                  UserPermission = "CAN_SEE_ALL_APPS"
	UserPermissionViewFinancialDataGlobal     UserPermission = "CAN_VIEW_FINANCIAL_DATA_GLOBAL"
	UserPermissionManagePermissionsGlobal     UserPermission = "CAN_MANAGE_PERMISSIONS_GLOBAL"
	UserPermissionEditGamesGlobal             UserPermission = "CAN_EDIT_GAMES_GLOBAL"
	UserPermissionPublishGamesGlobal          UserPermission = "CAN_PUBLISH_GAMES_GLOBAL"
	UserPermissionReplyToReviewsGlobal        UserPermission = "CAN_REPLY_TO_REVIEWS_GLOBAL"
	UserPermissionManagePublicAPKsGlobal      UserPermission = "CAN_MANAGE_PUBLIC_APKS_GLOBAL"
	UserPermissionManageTrackAPKsGlobal       UserPermission = "CAN_MANAGE_TRACK_APKS_GLOBAL"
	UserPermissionManageTrackUsersGlobal      UserPermission = "CAN_MANAGE_TRACK_USERS_GLOBAL"
	UserPermissionManagePublicListingGlobal   UserPermission = "CAN_MANAGE_PUBLIC_LISTING_GLOBAL"
	UserPermissionManageDraftAppsGlobal       UserPermission = "CAN_MANAGE_DRAFT_APPS_GLOBAL"
	UserPermissionCreateManagedPlayAppsGlobal UserPermission = "CAN_CREATE_MANAGED_PLAY_APPS_GLOBAL"
	UserPermissionChangeManagedPlayGlobal     UserPermission = "CAN_CHANGE_MANAGED_PLAY_SETTING_GLOBAL"
	UserPermissionManageOrdersGlobal          UserPermission = "CAN_MANAGE_ORDERS_GLOBAL"
	UserPermissionManageAppContentGlobal      UserPermission = "CAN_MANAGE_APP_CONTENT_GLOBAL"
	UserPermissionViewNonFinancialDataGlobal  UserPermission = "CAN_VIEW_NON_FINANCIAL_DATA_GLOBAL"
	UserPermissionViewAppQualityGlobal        UserPermission = "CAN_VIEW_APP_QUALITY_GLOBAL"
	UserPermissionManageDeepLinksGlobal       UserPermission = "CAN_MANAGE_DEEPLINKS_GLOBAL"
)

func NewUserPermission(value string) (UserPermission, error) {
	permission := UserPermission(strings.TrimSpace(value))
	switch permission {
	case UserPermissionSeeAllApps,
		UserPermissionViewFinancialDataGlobal,
		UserPermissionManagePermissionsGlobal,
		UserPermissionEditGamesGlobal,
		UserPermissionPublishGamesGlobal,
		UserPermissionReplyToReviewsGlobal,
		UserPermissionManagePublicAPKsGlobal,
		UserPermissionManageTrackAPKsGlobal,
		UserPermissionManageTrackUsersGlobal,
		UserPermissionManagePublicListingGlobal,
		UserPermissionManageDraftAppsGlobal,
		UserPermissionCreateManagedPlayAppsGlobal,
		UserPermissionChangeManagedPlayGlobal,
		UserPermissionManageOrdersGlobal,
		UserPermissionManageAppContentGlobal,
		UserPermissionViewNonFinancialDataGlobal,
		UserPermissionViewAppQualityGlobal,
		UserPermissionManageDeepLinksGlobal:
		return permission, nil
	default:
		return "", fmt.Errorf("unsupported user permission %q", value)
	}
}

func (p UserPermission) String() string {
	return string(p)
}

func (p UserPermission) Validate() error {
	_, err := NewUserPermission(p.String())
	return err
}

type UserListOptions struct {
	Developer DeveloperAccount `json:"developer"`
	PageSize  int64            `json:"pageSize,omitempty"`
	PageToken string           `json:"pageToken,omitempty"`
}

func (o UserListOptions) Validate() error {
	if err := o.Developer.Validate(); err != nil {
		return err
	}
	if o.PageSize < -1 {
		return fmt.Errorf("page size must be -1 or greater")
	}
	if o.PageSize == 0 {
		return nil
	}
	return nil
}

type UserGrant struct {
	Name                string   `json:"name,omitempty"`
	PackageName         string   `json:"packageName,omitempty"`
	AppLevelPermissions []string `json:"appLevelPermissions,omitempty"`
}

type User struct {
	Name                        string           `json:"name,omitempty"`
	Email                       string           `json:"email,omitempty"`
	AccessState                 string           `json:"accessState,omitempty"`
	ExpirationTime              string           `json:"expirationTime,omitempty"`
	DeveloperAccountPermissions []UserPermission `json:"developerAccountPermissions,omitempty"`
	Partial                     bool             `json:"partial"`
	Grants                      []UserGrant      `json:"grants"`
}

type UserListResult struct {
	Developer     DeveloperAccount `json:"developer"`
	NextPageToken string           `json:"nextPageToken,omitempty"`
	Users         []User           `json:"users"`
}

type UserLister interface {
	ListUsers(ctx context.Context, options UserListOptions) (UserListResult, error)
}

type UserDeleter interface {
	DeleteUser(ctx context.Context, options UserDeleteOptions) error
}

type UserCreator interface {
	CreateUser(ctx context.Context, options UserCreateOptions) (User, error)
}

func ListUsers(ctx context.Context, lister UserLister, options UserListOptions) (UserListResult, error) {
	if err := options.Validate(); err != nil {
		return UserListResult{}, err
	}
	if lister == nil {
		return UserListResult{}, fmt.Errorf("user lister is required")
	}
	return lister.ListUsers(ctx, options)
}

type UserCreateOptions struct {
	Developer      DeveloperAccount `json:"developer"`
	UserEmail      GrantUserEmail   `json:"userEmail"`
	Permissions    []UserPermission `json:"developerAccountPermissions"`
	ExpirationTime string           `json:"expirationTime,omitempty"`
	Confirm        bool             `json:"confirm"`
	DryRun         bool             `json:"dryRun"`
}

func (o UserCreateOptions) Validate() error {
	if err := o.Developer.Validate(); err != nil {
		return err
	}
	if err := o.UserEmail.Validate(); err != nil {
		return err
	}
	if err := validateUserPermissions(o.Permissions); err != nil {
		return err
	}
	if o.Confirm && o.DryRun {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.Confirm && !o.DryRun {
		return fmt.Errorf("user creation requires --confirm or --dry-run")
	}
	return nil
}

func (o UserCreateOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live user creation cannot run with --dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live user creation requires --confirm")
	}
	return nil
}

func (o UserCreateOptions) UserName() UserName {
	return NewUserNameFromParts(o.Developer, o.UserEmail)
}

type UserMutationPlan struct {
	Action         string           `json:"action"`
	Name           UserName         `json:"name"`
	Permissions    []UserPermission `json:"developerAccountPermissions,omitempty"`
	ExpirationTime string           `json:"expirationTime,omitempty"`
	Confirm        bool             `json:"confirm"`
	Steps          []string         `json:"steps"`
}

type UserMutationResult struct {
	Action  string           `json:"action"`
	DryRun  bool             `json:"dryRun"`
	Applied bool             `json:"applied"`
	User    *User            `json:"user,omitempty"`
	Desired *User            `json:"desiredUser,omitempty"`
	Plan    UserMutationPlan `json:"plan"`
}

func CreateUser(ctx context.Context, creator UserCreator, options UserCreateOptions) (UserMutationResult, error) {
	if err := options.Validate(); err != nil {
		return UserMutationResult{}, err
	}
	desired := User{
		Name:                        options.UserName().String(),
		Email:                       options.UserEmail.String(),
		DeveloperAccountPermissions: append([]UserPermission(nil), options.Permissions...),
		ExpirationTime:              options.ExpirationTime,
		Grants:                      []UserGrant{},
	}
	result := UserMutationResult{
		Action:  "create",
		DryRun:  options.DryRun,
		Desired: &desired,
		Plan: UserMutationPlan{
			Action:         "create",
			Name:           options.UserName(),
			Permissions:    append([]UserPermission(nil), options.Permissions...),
			ExpirationTime: options.ExpirationTime,
			Confirm:        options.Confirm,
			Steps:          []string{"create user access"},
		},
	}
	if options.DryRun {
		return result, nil
	}
	if creator == nil {
		return UserMutationResult{}, fmt.Errorf("user creator is required")
	}
	user, err := creator.CreateUser(ctx, options)
	if err != nil {
		return UserMutationResult{}, err
	}
	result.Applied = true
	result.User = &user
	return result, nil
}

type UserDeleteOptions struct {
	Name    UserName `json:"name"`
	Confirm bool     `json:"confirm"`
	DryRun  bool     `json:"dryRun"`
}

func (o UserDeleteOptions) Validate() error {
	if err := o.Name.Validate(); err != nil {
		return err
	}
	if o.Confirm && o.DryRun {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.Confirm && !o.DryRun {
		return fmt.Errorf("user deletion requires --confirm or --dry-run")
	}
	return nil
}

func (o UserDeleteOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live user deletion cannot run with --dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live user deletion requires --confirm")
	}
	return nil
}

type UserDeletePlan struct {
	Name    UserName `json:"name"`
	Confirm bool     `json:"confirm"`
	Steps   []string `json:"steps"`
}

type UserDeleteResult struct {
	Name    UserName       `json:"name"`
	DryRun  bool           `json:"dryRun"`
	Deleted bool           `json:"deleted"`
	Plan    UserDeletePlan `json:"plan"`
}

func DeleteUser(ctx context.Context, deleter UserDeleter, options UserDeleteOptions) (UserDeleteResult, error) {
	if err := options.Validate(); err != nil {
		return UserDeleteResult{}, err
	}
	result := UserDeleteResult{
		Name:   options.Name,
		DryRun: options.DryRun,
		Plan: UserDeletePlan{
			Name:    options.Name,
			Confirm: options.Confirm,
			Steps:   []string{"delete user access"},
		},
	}
	if options.DryRun {
		return result, nil
	}
	if deleter == nil {
		return UserDeleteResult{}, fmt.Errorf("user deleter is required")
	}
	if err := deleter.DeleteUser(ctx, options); err != nil {
		return UserDeleteResult{}, err
	}
	result.Deleted = true
	return result, nil
}

func validateUserPermissions(permissions []UserPermission) error {
	if len(permissions) == 0 {
		return fmt.Errorf("at least one user permission is required")
	}
	seen := make(map[UserPermission]struct{}, len(permissions))
	for _, permission := range permissions {
		if err := permission.Validate(); err != nil {
			return err
		}
		if _, ok := seen[permission]; ok {
			return fmt.Errorf("user permission %q is duplicated", permission)
		}
		seen[permission] = struct{}{}
	}
	return nil
}
