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

func ListUsers(ctx context.Context, lister UserLister, options UserListOptions) (UserListResult, error) {
	if err := options.Validate(); err != nil {
		return UserListResult{}, err
	}
	if lister == nil {
		return UserListResult{}, fmt.Errorf("user lister is required")
	}
	return lister.ListUsers(ctx, options)
}
