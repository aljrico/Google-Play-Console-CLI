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

type fakeAppRecoveryLister struct {
	options AppRecoveryListOptions
	result  AppRecoveryListResult
}

func (l *fakeAppRecoveryLister) ListAppRecoveries(ctx context.Context, options AppRecoveryListOptions) (AppRecoveryListResult, error) {
	l.options = options
	return l.result, nil
}
