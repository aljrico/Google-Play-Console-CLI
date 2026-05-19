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

func TestAddAppRecoveryTargetingDryRunDoesNotCallAdder(t *testing.T) {
	result, err := AddAppRecoveryTargeting(context.Background(), nil, AppRecoveryTargetingUpdateOptions{
		PackageName:   "com.example.app",
		AppRecoveryID: "7",
		AllUsers:      true,
		SDKLevels:     []int64{26, 35},
		RegionCodes:   []string{"US", "BR"},
		DryRun:        true,
	})
	if err != nil {
		t.Fatalf("AddAppRecoveryTargeting() error = %v", err)
	}
	if result.Applied {
		t.Fatalf("Applied = true, want false")
	}
	if !reflect.DeepEqual(result.Plan.Steps, []string{"add app recovery targeting", "target all users", "target android sdk levels", "target regions"}) {
		t.Fatalf("Steps = %#v", result.Plan.Steps)
	}
}

func TestAddAppRecoveryTargetingPassesOptionsToAdder(t *testing.T) {
	adder := &fakeAppRecoveryMutator{}
	options := AppRecoveryTargetingUpdateOptions{
		PackageName:   "com.example.app",
		AppRecoveryID: "7",
		RegionCodes:   []string{"US"},
		Confirm:       true,
	}

	result, err := AddAppRecoveryTargeting(context.Background(), adder, options)
	if err != nil {
		t.Fatalf("AddAppRecoveryTargeting() error = %v", err)
	}
	if !result.Applied {
		t.Fatalf("Applied = false, want true")
	}
	if !reflect.DeepEqual(adder.targetingOptions, options) {
		t.Fatalf("targetingOptions = %#v, want %#v", adder.targetingOptions, options)
	}
}

func TestAddAppRecoveryTargetingRejectsInvalidOptions(t *testing.T) {
	tests := []AppRecoveryTargetingUpdateOptions{
		{},
		{PackageName: "bad", AppRecoveryID: "7", AllUsers: true, DryRun: true},
		{PackageName: "com.example.app", AllUsers: true, DryRun: true},
		{PackageName: "com.example.app", AppRecoveryID: "7", DryRun: true},
		{PackageName: "com.example.app", AppRecoveryID: "7", SDKLevels: []int64{0}, DryRun: true},
		{PackageName: "com.example.app", AppRecoveryID: "7", RegionCodes: []string{"us"}, DryRun: true},
		{PackageName: "com.example.app", AppRecoveryID: "7", AllUsers: true},
		{PackageName: "com.example.app", AppRecoveryID: "7", AllUsers: true, Confirm: true, DryRun: true},
	}
	for _, options := range tests {
		if _, err := AddAppRecoveryTargeting(context.Background(), nil, options); err == nil {
			t.Fatalf("AddAppRecoveryTargeting(%#v) expected validation error", options)
		}
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
	deployOptions    AppRecoveryMutationOptions
	cancelOptions    AppRecoveryMutationOptions
	targetingOptions AppRecoveryTargetingUpdateOptions
}

func (m *fakeAppRecoveryMutator) DeployAppRecovery(ctx context.Context, options AppRecoveryMutationOptions) error {
	m.deployOptions = options
	return nil
}

func (m *fakeAppRecoveryMutator) CancelAppRecovery(ctx context.Context, options AppRecoveryMutationOptions) error {
	m.cancelOptions = options
	return nil
}

func (m *fakeAppRecoveryMutator) AddAppRecoveryTargeting(ctx context.Context, options AppRecoveryTargetingUpdateOptions) error {
	m.targetingOptions = options
	return nil
}
