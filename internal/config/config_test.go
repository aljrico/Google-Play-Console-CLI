package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadMissingConfigReturnsEmptyStore(t *testing.T) {
	t.Setenv("GPC_CONFIG", filepath.Join(t.TempDir(), "missing.json"))

	store, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if store.Profiles == nil {
		t.Fatal("Profiles = nil, want empty map")
	}
	if len(store.Profiles) != 0 {
		t.Fatalf("len(Profiles) = %d, want 0", len(store.Profiles))
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "nested", "config.json")
	t.Setenv("GPC_CONFIG", configPath)

	want := Store{
		ActiveProfile: "default",
		Profiles: map[string]Profile{
			"default": {
				Name:               "default",
				ServiceAccountFile: "/tmp/service-account.json",
			},
		},
	}

	if err := Save(want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	got, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.ActiveProfile != want.ActiveProfile {
		t.Fatalf("ActiveProfile = %q, want %q", got.ActiveProfile, want.ActiveProfile)
	}
	if got.Profiles["default"].ServiceAccountFile != "/tmp/service-account.json" {
		t.Fatalf("ServiceAccountFile = %q", got.Profiles["default"].ServiceAccountFile)
	}
}

func TestLoadCorruptConfigIncludesPath(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("GPC_CONFIG", configPath)
	if err := os.WriteFile(configPath, []byte("{"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := Load()
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !strings.Contains(err.Error(), configPath) {
		t.Fatalf("error = %q, want config path", err.Error())
	}
}
