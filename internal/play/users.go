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
	Name                        string      `json:"name,omitempty"`
	Email                       string      `json:"email,omitempty"`
	AccessState                 string      `json:"accessState,omitempty"`
	ExpirationTime              string      `json:"expirationTime,omitempty"`
	DeveloperAccountPermissions []string    `json:"developerAccountPermissions,omitempty"`
	Partial                     bool        `json:"partial"`
	Grants                      []UserGrant `json:"grants"`
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

func ListUsers(ctx context.Context, lister UserLister, options UserListOptions) (UserListResult, error) {
	if err := options.Validate(); err != nil {
		return UserListResult{}, err
	}
	if lister == nil {
		return UserListResult{}, fmt.Errorf("user lister is required")
	}
	return lister.ListUsers(ctx, options)
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
