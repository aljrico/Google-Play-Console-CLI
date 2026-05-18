package play

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestListAppRecoveriesPassesOptionsToLister(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	lister := &fakeAppRecoveryLister{result: AppRecoveryListResult{
		PackageName: packageName,
		VersionCode: 42,
		Actions:     []AppRecoveryAction{{AppRecoveryID: "7", Status: "RECOVERY_STATUS_ACTIVE"}},
	}}
	options := AppRecoveryListOptions{PackageName: packageName, VersionCode: 42}

	result, err := ListAppRecoveries(context.Background(), lister, options)
	if err != nil {
		t.Fatalf("ListAppRecoveries() error = %v", err)
	}
	if len(result.Actions) != 1 {
		t.Fatalf("len(Actions) = %d, want 1", len(result.Actions))
	}
	if !reflect.DeepEqual(lister.options, options) {
		t.Fatalf("options = %#v, want %#v", lister.options, options)
	}
}

func TestListAppRecoveriesRejectsInvalidOptions(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	tests := []AppRecoveryListOptions{
		{PackageName: "bad", VersionCode: 42},
		{PackageName: packageName},
		{PackageName: packageName, VersionCode: -1},
	}
	for _, options := range tests {
		if _, err := ListAppRecoveries(context.Background(), nil, options); err == nil {
			t.Fatalf("ListAppRecoveries(%#v) expected validation error", options)
		}
	}
}

func TestAppRecoveryActionJSONUsesGoogleFieldName(t *testing.T) {
	payload, err := json.Marshal(AppRecoveryAction{AppRecoveryID: "7"})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	output := string(payload)
	if !strings.Contains(output, `"appRecoveryId":"7"`) {
		t.Fatalf("output = %s, want appRecoveryId", output)
	}
	if strings.Contains(output, `"id"`) {
		t.Fatalf("output = %s, did not expect legacy id field", output)
	}
}

func TestDeployAppRecoveryDryRunDoesNotCallMutator(t *testing.T) {
	result, err := DeployAppRecovery(context.Background(), nil, AppRecoveryMutationOptions{
		PackageName:   "com.example.app",
		AppRecoveryID: "7",
		DryRun:        true,
	})
	if err != nil {
		t.Fatalf("DeployAppRecovery() error = %v", err)
	}
	if result.Applied {
		t.Fatalf("Applied = true, want false")
	}
	if result.Action != "deploy" || result.Plan.Action != "deploy" {
		t.Fatalf("result = %#v, want deploy action", result)
	}
}

func TestCancelAppRecoveryPassesOptionsToMutator(t *testing.T) {
	mutator := &fakeAppRecoveryMutator{}
	options := AppRecoveryMutationOptions{
		PackageName:   "com.example.app",
		AppRecoveryID: "7",
		Confirm:       true,
	}

	result, err := CancelAppRecovery(context.Background(), mutator, options)
	if err != nil {
		t.Fatalf("CancelAppRecovery() error = %v", err)
	}
	if !result.Applied {
		t.Fatalf("Applied = false, want true")
	}
	if mutator.cancelOptions != options {
		t.Fatalf("cancelOptions = %#v, want %#v", mutator.cancelOptions, options)
	}
}

func TestAppRecoveryMutationRejectsInvalidOptions(t *testing.T) {
	tests := []AppRecoveryMutationOptions{
		{},
		{PackageName: "bad", AppRecoveryID: "7", DryRun: true},
		{PackageName: "com.example.app", DryRun: true},
		{PackageName: "com.example.app", AppRecoveryID: "0", DryRun: true},
		{PackageName: "com.example.app", AppRecoveryID: "abc", DryRun: true},
		{PackageName: "com.example.app", AppRecoveryID: "7"},
		{PackageName: "com.example.app", AppRecoveryID: "7", Confirm: true, DryRun: true},
	}
	for _, options := range tests {
		if _, err := DeployAppRecovery(context.Background(), nil, options); err == nil {
			t.Fatalf("DeployAppRecovery(%#v) expected validation error", options)
		}
	}
}

type fakeAppRecoveryLister struct {
	options AppRecoveryListOptions
	result  AppRecoveryListResult
}

func (l *fakeAppRecoveryLister) ListAppRecoveries(ctx context.Context, options AppRecoveryListOptions) (AppRecoveryListResult, error) {
	l.options = options
	return l.result, nil
}

type fakeAppRecoveryMutator struct {
	deployOptions AppRecoveryMutationOptions
	cancelOptions AppRecoveryMutationOptions
}

func (m *fakeAppRecoveryMutator) DeployAppRecovery(ctx context.Context, options AppRecoveryMutationOptions) error {
	m.deployOptions = options
	return nil
}

func (m *fakeAppRecoveryMutator) CancelAppRecovery(ctx context.Context, options AppRecoveryMutationOptions) error {
	m.cancelOptions = options
	return nil
}
