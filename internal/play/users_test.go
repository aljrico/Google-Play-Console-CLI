package play

import (
	"context"
	"reflect"
	"testing"
)

func TestNewDeveloperAccountAcceptsIDAndResource(t *testing.T) {
	for _, value := range []string{"1234567890", "developers/1234567890"} {
		account, err := NewDeveloperAccount(value)
		if err != nil {
			t.Fatalf("NewDeveloperAccount(%q) error = %v", value, err)
		}
		if account.ResourceName() != "developers/1234567890" {
			t.Fatalf("ResourceName() = %q, want developers/1234567890", account.ResourceName())
		}
	}
}

func TestListUsersPassesOptionsToLister(t *testing.T) {
	lister := &fakeUserLister{result: UserListResult{
		Developer:     "1234567890",
		NextPageToken: "next",
		Users:         []User{{Email: "user@example.com"}},
	}}
	options := UserListOptions{
		Developer: "1234567890",
		PageSize:  25,
		PageToken: "page",
	}

	result, err := ListUsers(context.Background(), lister, options)
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	if len(result.Users) != 1 {
		t.Fatalf("len(Users) = %d, want 1", len(result.Users))
	}
	if !reflect.DeepEqual(lister.options, options) {
		t.Fatalf("options = %#v, want %#v", lister.options, options)
	}
}

func TestListUsersRejectsInvalidOptions(t *testing.T) {
	tests := []UserListOptions{
		{},
		{Developer: "developers/123/extra"},
		{Developer: "1234567890", PageSize: -2},
	}
	for _, options := range tests {
		_, err := ListUsers(context.Background(), nil, options)
		if err == nil {
			t.Fatalf("ListUsers(%#v) expected validation error", options)
		}
	}
}

func TestNewUserNameFromParts(t *testing.T) {
	name := NewUserNameFromParts("1234567890", "user@example.com")
	if name.String() != "developers/1234567890/users/user@example.com" {
		t.Fatalf("name = %q, want user resource", name)
	}
	if err := name.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestDeleteUserDryRunDoesNotCallDeleter(t *testing.T) {
	result, err := DeleteUser(context.Background(), nil, UserDeleteOptions{
		Name:   "developers/1234567890/users/user@example.com",
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("DeleteUser() error = %v", err)
	}
	if result.Deleted {
		t.Fatalf("Deleted = true, want false")
	}
	if result.Plan.Name != "developers/1234567890/users/user@example.com" {
		t.Fatalf("plan = %#v, want name", result.Plan)
	}
}

func TestCreateUserDryRunDoesNotCallCreator(t *testing.T) {
	result, err := CreateUser(context.Background(), nil, UserCreateOptions{
		Developer:   "1234567890",
		UserEmail:   "user@example.com",
		Permissions: []UserPermission{UserPermissionViewNonFinancialDataGlobal},
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if result.Applied {
		t.Fatalf("Applied = true, want false")
	}
	if result.Desired == nil || result.Desired.Email != "user@example.com" {
		t.Fatalf("desired = %#v, want user@example.com", result.Desired)
	}
}

func TestCreateUserPassesOptionsToCreator(t *testing.T) {
	creator := &fakeUserCreator{user: User{Email: "user@example.com"}}
	options := UserCreateOptions{
		Developer:   "1234567890",
		UserEmail:   "user@example.com",
		Permissions: []UserPermission{UserPermissionViewNonFinancialDataGlobal},
		Confirm:     true,
	}

	result, err := CreateUser(context.Background(), creator, options)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if !result.Applied {
		t.Fatalf("Applied = false, want true")
	}
	if !reflect.DeepEqual(creator.options, options) {
		t.Fatalf("options = %#v, want %#v", creator.options, options)
	}
}

func TestCreateUserRejectsInvalidOptions(t *testing.T) {
	tests := []UserCreateOptions{
		{},
		{Developer: "bad/account", UserEmail: "user@example.com", Permissions: []UserPermission{UserPermissionViewNonFinancialDataGlobal}, DryRun: true},
		{Developer: "1234567890", Permissions: []UserPermission{UserPermissionViewNonFinancialDataGlobal}, DryRun: true},
		{Developer: "1234567890", UserEmail: "user@example.com", DryRun: true},
		{Developer: "1234567890", UserEmail: "user@example.com", Permissions: []UserPermission{"NOPE"}, DryRun: true},
		{Developer: "1234567890", UserEmail: "user@example.com", Permissions: []UserPermission{UserPermissionViewNonFinancialDataGlobal}},
		{Developer: "1234567890", UserEmail: "user@example.com", Permissions: []UserPermission{UserPermissionViewNonFinancialDataGlobal}, Confirm: true, DryRun: true},
	}
	for _, options := range tests {
		if _, err := CreateUser(context.Background(), nil, options); err == nil {
			t.Fatalf("CreateUser(%#v) expected validation error", options)
		}
	}
}

func TestPatchUserDryRunDoesNotCallPatcher(t *testing.T) {
	result, err := PatchUser(context.Background(), nil, UserPatchOptions{
		Name:        "developers/1234567890/users/user@example.com",
		Permissions: []UserPermission{UserPermissionReplyToReviewsGlobal},
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("PatchUser() error = %v", err)
	}
	if result.Applied {
		t.Fatalf("Applied = true, want false")
	}
	if result.Plan.Name != "developers/1234567890/users/user@example.com" || result.Plan.Action != "patch" {
		t.Fatalf("plan = %#v, want patch plan", result.Plan)
	}
}

func TestPatchUserPassesOptionsToPatcher(t *testing.T) {
	patcher := &fakeUserPatcher{user: User{Email: "user@example.com"}}
	options := UserPatchOptions{
		Name:           "developers/1234567890/users/user@example.com",
		ExpirationTime: "2027-01-02T03:04:05Z",
		Confirm:        true,
	}

	result, err := PatchUser(context.Background(), patcher, options)
	if err != nil {
		t.Fatalf("PatchUser() error = %v", err)
	}
	if !result.Applied {
		t.Fatalf("Applied = false, want true")
	}
	if !reflect.DeepEqual(patcher.options, options) {
		t.Fatalf("options = %#v, want %#v", patcher.options, options)
	}
}

func TestPatchUserRejectsInvalidOptions(t *testing.T) {
	tests := []UserPatchOptions{
		{},
		{Name: "developers/1234567890/users/user@example.com", DryRun: true},
		{Name: "developers/1234567890/users/user@example.com", Permissions: []UserPermission{"NOPE"}, DryRun: true},
		{Name: "developers/1234567890/users/user@example.com", Permissions: []UserPermission{UserPermissionReplyToReviewsGlobal}},
		{Name: "developers/1234567890/users/user@example.com", Permissions: []UserPermission{UserPermissionReplyToReviewsGlobal}, Confirm: true, DryRun: true},
	}
	for _, options := range tests {
		if _, err := PatchUser(context.Background(), nil, options); err == nil {
			t.Fatalf("PatchUser(%#v) expected validation error", options)
		}
	}
}

func TestUserPatchOptionsBuildsUpdateMask(t *testing.T) {
	options := UserPatchOptions{
		Name:           "developers/1234567890/users/user@example.com",
		Permissions:    []UserPermission{UserPermissionReplyToReviewsGlobal},
		ExpirationTime: "2027-01-02T03:04:05Z",
		DryRun:         true,
	}
	if options.UpdateMask() != "developerAccountPermissions,expirationTime" {
		t.Fatalf("UpdateMask() = %q, want developerAccountPermissions,expirationTime", options.UpdateMask())
	}
}

func TestDeleteUserPassesOptionsToDeleter(t *testing.T) {
	deleter := &fakeUserDeleter{}
	options := UserDeleteOptions{
		Name:    "developers/1234567890/users/user@example.com",
		Confirm: true,
	}

	result, err := DeleteUser(context.Background(), deleter, options)
	if err != nil {
		t.Fatalf("DeleteUser() error = %v", err)
	}
	if !result.Deleted {
		t.Fatalf("Deleted = false, want true")
	}
	if deleter.options != options {
		t.Fatalf("options = %#v, want %#v", deleter.options, options)
	}
}

func TestDeleteUserRejectsInvalidOptions(t *testing.T) {
	tests := []UserDeleteOptions{
		{},
		{Name: "developers/123/users/user@example.com/extra", DryRun: true},
		{Name: "developers/1234567890/users/user@example.com"},
		{Name: "developers/1234567890/users/user@example.com", Confirm: true, DryRun: true},
	}
	for _, options := range tests {
		if _, err := DeleteUser(context.Background(), nil, options); err == nil {
			t.Fatalf("DeleteUser(%#v) expected validation error", options)
		}
	}
}

type fakeUserLister struct {
	options UserListOptions
	result  UserListResult
}

func (l *fakeUserLister) ListUsers(ctx context.Context, options UserListOptions) (UserListResult, error) {
	l.options = options
	return l.result, nil
}

type fakeUserDeleter struct {
	options UserDeleteOptions
}

func (d *fakeUserDeleter) DeleteUser(ctx context.Context, options UserDeleteOptions) error {
	d.options = options
	return nil
}

type fakeUserCreator struct {
	options UserCreateOptions
	user    User
}

func (c *fakeUserCreator) CreateUser(ctx context.Context, options UserCreateOptions) (User, error) {
	c.options = options
	return c.user, nil
}

type fakeUserPatcher struct {
	options UserPatchOptions
	user    User
}

func (p *fakeUserPatcher) PatchUser(ctx context.Context, options UserPatchOptions) (User, error) {
	p.options = options
	return p.user, nil
}
