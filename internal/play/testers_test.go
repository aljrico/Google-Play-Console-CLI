package play

import (
	"context"
	"reflect"
	"testing"
)

func TestNewTesterGoogleGroupRejectsDisplayName(t *testing.T) {
	if _, err := NewTesterGoogleGroup("QA <qa@example.com>"); err == nil {
		t.Fatal("NewTesterGoogleGroup() error = nil, want display-name rejection")
	}
}

func TestUpdateTestersDryRunSortsGroupsWithoutPublisher(t *testing.T) {
	packageName := mustPackageName(t, "com.example.app")
	result, err := UpdateTesters(context.Background(), nil, TestersUpdateOptions{
		PackageName: packageName,
		Track:       TrackInternal,
		GoogleGroups: []TesterGoogleGroup{
			"Beta@Example.com",
			"alpha@example.com",
		},
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("UpdateTesters() error = %v", err)
	}

	want := []TesterGoogleGroup{"alpha@example.com", "beta@example.com"}
	if !reflect.DeepEqual(result.Desired.GoogleGroups, want) {
		t.Fatalf("Desired.GoogleGroups = %#v, want %#v", result.Desired.GoogleGroups, want)
	}
	if result.Committed {
		t.Fatal("Committed = true, want false")
	}
}

func TestUpdateTestersClearCommitsEmptyGoogleGroups(t *testing.T) {
	packageName := mustPackageName(t, "com.example.app")
	fake := &fakeTestersPublisher{editID: "edit-1"}

	result, err := UpdateTesters(context.Background(), fake, TestersUpdateOptions{
		PackageName: packageName,
		Track:       TrackBeta,
		Clear:       true,
		Confirm:     true,
	})
	if err != nil {
		t.Fatalf("UpdateTesters() error = %v", err)
	}

	if !result.Committed {
		t.Fatal("Committed = false, want true")
	}
	if len(fake.updatedGoogleGroups) != 0 {
		t.Fatalf("updatedGoogleGroups = %#v, want empty", fake.updatedGoogleGroups)
	}
	wantCalls := []string{"insert", "update", "validate", "commit"}
	if !reflect.DeepEqual(fake.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", fake.calls, wantCalls)
	}
}

func TestGetTestersDeletesTemporaryEdit(t *testing.T) {
	packageName := mustPackageName(t, "com.example.app")
	fake := &fakeTestersPublisher{
		editID: "edit-1",
		testers: TrackTesters{
			PackageName:  packageName,
			Track:        TrackInternal,
			GoogleGroups: []TesterGoogleGroup{"qa@example.com"},
		},
	}

	testers, err := GetTesters(context.Background(), fake, TestersGetOptions{
		PackageName: packageName,
		Track:       TrackInternal,
	})
	if err != nil {
		t.Fatalf("GetTesters() error = %v", err)
	}

	if testers.GoogleGroups[0] != "qa@example.com" {
		t.Fatalf("GoogleGroups = %#v, want qa@example.com", testers.GoogleGroups)
	}
	wantCalls := []string{"insert", "get", "delete"}
	if !reflect.DeepEqual(fake.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", fake.calls, wantCalls)
	}
}

func TestUpdateTestersRequiresConfirmOrDryRun(t *testing.T) {
	packageName := mustPackageName(t, "com.example.app")
	_, err := UpdateTesters(context.Background(), nil, TestersUpdateOptions{
		PackageName:  packageName,
		Track:        TrackInternal,
		GoogleGroups: []TesterGoogleGroup{"qa@example.com"},
	})
	if err == nil {
		t.Fatal("UpdateTesters() error = nil, want confirm or dry-run error")
	}
}

func TestUpdateTestersRejectsConfirmAndDryRunTogether(t *testing.T) {
	packageName := mustPackageName(t, "com.example.app")
	_, err := UpdateTesters(context.Background(), nil, TestersUpdateOptions{
		PackageName:  packageName,
		Track:        TrackInternal,
		GoogleGroups: []TesterGoogleGroup{"qa@example.com"},
		Confirm:      true,
		DryRun:       true,
	})
	if err == nil {
		t.Fatal("UpdateTesters() error = nil, want confirm and dry-run conflict")
	}
}

func TestUpdateTestersRejectsCanonicalDuplicateGroups(t *testing.T) {
	packageName := mustPackageName(t, "com.example.app")
	_, err := UpdateTesters(context.Background(), nil, TestersUpdateOptions{
		PackageName: packageName,
		Track:       TrackInternal,
		GoogleGroups: []TesterGoogleGroup{
			" qa@example.com ",
			"QA@example.com",
		},
		DryRun: true,
	})
	if err == nil {
		t.Fatal("UpdateTesters() error = nil, want duplicate group error")
	}
}

func mustPackageName(t *testing.T, value string) PackageName {
	t.Helper()
	packageName, err := NewPackageName(value)
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	return packageName
}

type fakeTestersPublisher struct {
	editID              string
	testers             TrackTesters
	updatedGoogleGroups []TesterGoogleGroup
	calls               []string
}

func (p *fakeTestersPublisher) InsertEdit(ctx context.Context, packageName PackageName) (Edit, error) {
	p.calls = append(p.calls, "insert")
	return Edit{ID: p.editID}, nil
}

func (p *fakeTestersPublisher) GetTesters(ctx context.Context, packageName PackageName, editID string, track TrackName) (TrackTesters, error) {
	p.calls = append(p.calls, "get")
	return p.testers, nil
}

func (p *fakeTestersPublisher) UpdateTesters(ctx context.Context, packageName PackageName, editID string, track TrackName, googleGroups []TesterGoogleGroup) (TrackTesters, error) {
	p.calls = append(p.calls, "update")
	p.updatedGoogleGroups = append([]TesterGoogleGroup(nil), googleGroups...)
	return TrackTesters{PackageName: packageName, Track: track, GoogleGroups: googleGroups}, nil
}

func (p *fakeTestersPublisher) ValidateEdit(ctx context.Context, packageName PackageName, editID string) error {
	p.calls = append(p.calls, "validate")
	return nil
}

func (p *fakeTestersPublisher) CommitEdit(ctx context.Context, packageName PackageName, editID string) (Edit, error) {
	p.calls = append(p.calls, "commit")
	return Edit{ID: editID}, nil
}

func (p *fakeTestersPublisher) DeleteEdit(ctx context.Context, packageName PackageName, editID string) error {
	p.calls = append(p.calls, "delete")
	return nil
}
