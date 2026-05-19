package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadMissingConfigReturnsEmptyStore(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", filepath.Join(t.TempDir(), "missing.json"))

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
	t.Setenv("PLAYPUB_CONFIG", configPath)

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

func TestSaveUsesPrivatePermissions(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "nested", "config.json")
	t.Setenv("PLAYPUB_CONFIG", configPath)

	store := Store{
		ActiveProfile: "default",
		Profiles: map[string]Profile{
			"default": {
				Name:               "default",
				ServiceAccountFile: "/tmp/service-account.json",
			},
		},
	}
	if err := Save(store); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	dirInfo, err := os.Stat(filepath.Dir(configPath))
	if err != nil {
		t.Fatalf("Stat(config dir) error = %v", err)
	}
	if got := dirInfo.Mode().Perm(); got != 0o700 {
		t.Fatalf("config dir perm = %o, want 700", got)
	}

	fileInfo, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Stat(config file) error = %v", err)
	}
	if got := fileInfo.Mode().Perm(); got != 0o600 {
		t.Fatalf("config file perm = %o, want 600", got)
	}
}

func TestLoadCorruptConfigIncludesPath(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("PLAYPUB_CONFIG", configPath)
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
