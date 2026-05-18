package play

import (
	"context"
	"testing"
)

func TestCreateGrantDryRunDoesNotRequireManager(t *testing.T) {
	result, err := CreateGrant(context.Background(), nil, GrantCreateOptions{
		Developer:   "1234567890",
		UserEmail:   "user@example.com",
		PackageName: "com.example.app",
		Permissions: []GrantPermission{GrantPermissionViewNonFinancialData},
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("CreateGrant() error = %v", err)
	}
	if !result.DryRun || result.Applied {
		t.Fatalf("result = %#v, want dry-run not applied", result)
	}
	if result.Plan.Target != "developers/1234567890/users/user@example.com/grants/com.example.app" {
		t.Fatalf("target = %q, want grant resource", result.Plan.Target)
	}
	if result.Desired == nil || result.Desired.Name != "developers/1234567890/users/user@example.com/grants/com.example.app" {
		t.Fatalf("desired grant = %#v, want full grant resource", result.Desired)
	}
	if len(result.Plan.Permissions) != 1 || result.Plan.Permissions[0] != GrantPermissionViewNonFinancialData {
		t.Fatalf("plan permissions = %#v, want requested permission", result.Plan.Permissions)
	}
}

func TestPatchGrantAppliesWithConfirm(t *testing.T) {
	manager := &fakeGrantManager{grant: Grant{Name: "developers/123/users/user@example.com/grants/com.example.app"}}
	result, err := PatchGrant(context.Background(), manager, GrantPatchOptions{
		Name:        "developers/123/users/user@example.com/grants/com.example.app",
		Permissions: []GrantPermission{GrantPermissionReplyToReviews},
		Confirm:     true,
	})
	if err != nil {
		t.Fatalf("PatchGrant() error = %v", err)
	}
	if !result.Applied || result.Grant == nil {
		t.Fatalf("result = %#v, want applied grant", result)
	}
}

func TestDeleteGrantAppliesWithConfirm(t *testing.T) {
	manager := &fakeGrantManager{}
	result, err := DeleteGrant(context.Background(), manager, GrantDeleteOptions{
		Name:    "developers/123/users/user@example.com/grants/com.example.app",
		Confirm: true,
	})
	if err != nil {
		t.Fatalf("DeleteGrant() error = %v", err)
	}
	if !result.Applied || !manager.deleted {
		t.Fatalf("result = %#v deleted = %v, want applied deletion", result, manager.deleted)
	}
}

func TestGrantOptionsRejectInvalidInputs(t *testing.T) {
	createTests := []GrantCreateOptions{
		{},
		{Developer: "123", UserEmail: "user@example.com", PackageName: "bad", Permissions: []GrantPermission{GrantPermissionViewNonFinancialData}, DryRun: true},
		{Developer: "123", UserEmail: "not-email", PackageName: "com.example.app", Permissions: []GrantPermission{GrantPermissionViewNonFinancialData}, DryRun: true},
		{Developer: "123", UserEmail: "user@example.com", PackageName: "com.example.app", DryRun: true},
		{Developer: "123", UserEmail: "user@example.com", PackageName: "com.example.app", Permissions: []GrantPermission{"NOPE"}, DryRun: true},
		{Developer: "123", UserEmail: "user@example.com", PackageName: "com.example.app", Permissions: []GrantPermission{GrantPermissionViewNonFinancialData}},
	}
	for _, options := range createTests {
		if _, err := CreateGrant(context.Background(), nil, options); err == nil {
			t.Fatalf("CreateGrant(%#v) expected validation error", options)
		}
	}

	if _, err := PatchGrant(context.Background(), nil, GrantPatchOptions{Name: "bad", Permissions: []GrantPermission{GrantPermissionViewNonFinancialData}, DryRun: true}); err == nil {
		t.Fatal("PatchGrant() expected invalid name error")
	}
	for _, name := range []GrantName{
		"developers/123/users//grants/com.example.app",
		"developers/123/users/not-email/grants/com.example.app",
		"developers/123/users/user@example.com/grants/com.example.app/extra",
	} {
		if _, err := PatchGrant(context.Background(), nil, GrantPatchOptions{Name: name, Permissions: []GrantPermission{GrantPermissionViewNonFinancialData}, DryRun: true}); err == nil {
			t.Fatalf("PatchGrant(%q) expected invalid name error", name)
		}
	}
	if _, err := DeleteGrant(context.Background(), nil, GrantDeleteOptions{Name: "developers/123/users/user@example.com/grants/com.example.app"}); err == nil {
		t.Fatal("DeleteGrant() expected confirmation error")
	}
}

type fakeGrantManager struct {
	grant   Grant
	deleted bool
}

func (m *fakeGrantManager) CreateGrant(ctx context.Context, options GrantCreateOptions) (Grant, error) {
	return m.grant, nil
}

func (m *fakeGrantManager) PatchGrant(ctx context.Context, options GrantPatchOptions) (Grant, error) {
	return m.grant, nil
}

func (m *fakeGrantManager) DeleteGrant(ctx context.Context, options GrantDeleteOptions) error {
	m.deleted = true
	return nil
}
