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

type fakeUserLister struct {
	options UserListOptions
	result  UserListResult
}

func (l *fakeUserLister) ListUsers(ctx context.Context, options UserListOptions) (UserListResult, error) {
	l.options = options
	return l.result, nil
}
