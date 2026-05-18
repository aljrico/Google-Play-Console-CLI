package play

import (
	"context"
	"reflect"
	"testing"
)

func TestUpdateRolloutHaltDryRunDoesNotRequireController(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := UpdateRollout(context.Background(), nil, RolloutOptions{
		PackageName: packageName,
		Track:       TrackProduction,
		VersionCode: 42,
		Action:      RolloutActionHalt,
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("UpdateRollout() error = %v", err)
	}
	if result.Plan.Status != ReleaseStatusHalted {
		t.Fatalf("Status = %q, want %q", result.Plan.Status, ReleaseStatusHalted)
	}
}

func TestUpdateRolloutResumeRequiresStatus(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = NewRolloutPlan(RolloutOptions{
		PackageName: packageName,
		Track:       TrackProduction,
		VersionCode: 42,
		Action:      RolloutActionResume,
	})
	if err == nil {
		t.Fatal("expected status error")
	}
}

func TestUpdateRolloutResumeRequiresUserFractionForInProgress(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = NewRolloutPlan(RolloutOptions{
		PackageName: packageName,
		Track:       TrackProduction,
		VersionCode: 42,
		Action:      RolloutActionResume,
		Status:      ReleaseStatusInProgress,
	})
	if err == nil {
		t.Fatal("expected user-fraction error")
	}
}

func TestUpdateRolloutValidatesAndCleansUpWithoutConfirm(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	controller := &fakeRolloutController{}

	result, err := UpdateRollout(context.Background(), controller, RolloutOptions{
		PackageName: packageName,
		Track:       TrackProduction,
		VersionCode: 42,
		Action:      RolloutActionHalt,
	})
	if err != nil {
		t.Fatalf("UpdateRollout() error = %v", err)
	}
	if result.Committed {
		t.Fatal("Committed = true, want false")
	}
	wantCalls := []string{"insert", "update-status", "validate", "delete"}
	if !reflect.DeepEqual(controller.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", controller.calls, wantCalls)
	}
}

func TestUpdateRolloutCommitsWithConfirm(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	controller := &fakeRolloutController{}
	userFraction := 0.25

	result, err := UpdateRollout(context.Background(), controller, RolloutOptions{
		PackageName:  packageName,
		Track:        TrackProduction,
		VersionCode:  42,
		Action:       RolloutActionResume,
		Status:       ReleaseStatusInProgress,
		UserFraction: &userFraction,
		Confirm:      true,
	})
	if err != nil {
		t.Fatalf("UpdateRollout() error = %v", err)
	}
	if !result.Committed {
		t.Fatal("Committed = false, want true")
	}
	if controller.status != ReleaseStatusInProgress {
		t.Fatalf("status = %q, want %q", controller.status, ReleaseStatusInProgress)
	}
	wantCalls := []string{"insert", "update-status", "validate", "commit"}
	if !reflect.DeepEqual(controller.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", controller.calls, wantCalls)
	}
}

func TestUpdateRolloutCanResumeToCompleted(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	plan, err := NewRolloutPlan(RolloutOptions{
		PackageName: packageName,
		Track:       TrackProduction,
		VersionCode: 42,
		Action:      RolloutActionResume,
		Status:      ReleaseStatusCompleted,
	})
	if err != nil {
		t.Fatalf("NewRolloutPlan() error = %v", err)
	}
	if plan.Status != ReleaseStatusCompleted {
		t.Fatalf("Status = %q, want %q", plan.Status, ReleaseStatusCompleted)
	}
}

type fakeRolloutController struct {
	calls  []string
	status ReleaseStatus
}

func (c *fakeRolloutController) InsertEdit(ctx context.Context, packageName PackageName) (Edit, error) {
	c.calls = append(c.calls, "insert")
	return Edit{ID: "edit-123"}, nil
}

func (c *fakeRolloutController) UpdateTrackReleaseStatus(ctx context.Context, packageName PackageName, editID string, track TrackName, versionCode int64, status ReleaseStatus, userFraction *float64) (TrackRelease, error) {
	c.calls = append(c.calls, "update-status")
	c.status = status
	return TrackRelease{Status: status, UserFraction: userFraction, VersionCodes: []int64{versionCode}}, nil
}

func (c *fakeRolloutController) ValidateEdit(ctx context.Context, packageName PackageName, editID string) error {
	c.calls = append(c.calls, "validate")
	return nil
}

func (c *fakeRolloutController) CommitEdit(ctx context.Context, packageName PackageName, editID string) (Edit, error) {
	c.calls = append(c.calls, "commit")
	return Edit{ID: editID}, nil
}

func (c *fakeRolloutController) DeleteEdit(ctx context.Context, packageName PackageName, editID string) error {
	c.calls = append(c.calls, "delete")
	return nil
}
