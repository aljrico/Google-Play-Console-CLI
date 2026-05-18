package play

import (
	"context"
	"testing"
)

func TestUpdateDataSafetyDryRunDoesNotRequireUpdater(t *testing.T) {
	result, err := UpdateDataSafety(context.Background(), nil, DataSafetyUpdateOptions{
		PackageName:  "com.example.app",
		CSVPath:      "data-safety.csv",
		SafetyLabels: "question,answer\n",
		DryRun:       true,
	})
	if err != nil {
		t.Fatalf("UpdateDataSafety() error = %v", err)
	}
	if !result.DryRun || result.Applied {
		t.Fatalf("result = %#v, want dry-run not applied", result)
	}
}

func TestUpdateDataSafetyRejectsInvalidOptions(t *testing.T) {
	tests := []DataSafetyUpdateOptions{
		{},
		{PackageName: "bad", CSVPath: "data-safety.csv", SafetyLabels: "question,answer\n", DryRun: true},
		{PackageName: "com.example.app", SafetyLabels: "question,answer\n", DryRun: true},
		{PackageName: "com.example.app", CSVPath: "data-safety.csv", DryRun: true},
		{PackageName: "com.example.app", CSVPath: "data-safety.csv", SafetyLabels: "question,answer\n"},
	}
	for _, options := range tests {
		_, err := UpdateDataSafety(context.Background(), nil, options)
		if err == nil {
			t.Fatalf("UpdateDataSafety(%#v) expected validation error", options)
		}
	}
}

func TestUpdateDataSafetyAppliesWithConfirm(t *testing.T) {
	updater := &fakeDataSafetyUpdater{}
	result, err := UpdateDataSafety(context.Background(), updater, DataSafetyUpdateOptions{
		PackageName:  "com.example.app",
		CSVPath:      "data-safety.csv",
		SafetyLabels: "question,answer\n",
		Confirm:      true,
	})
	if err != nil {
		t.Fatalf("UpdateDataSafety() error = %v", err)
	}
	if !result.Applied {
		t.Fatal("Applied = false, want true")
	}
	if updater.packageName != "com.example.app" || updater.safetyLabels != "question,answer\n" {
		t.Fatalf("updater = %#v, want package and CSV content", updater)
	}
}

type fakeDataSafetyUpdater struct {
	packageName  PackageName
	safetyLabels string
}

func (u *fakeDataSafetyUpdater) UpdateDataSafety(ctx context.Context, packageName PackageName, safetyLabels string) error {
	u.packageName = packageName
	u.safetyLabels = safetyLabels
	return nil
}
