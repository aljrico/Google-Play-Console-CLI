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
