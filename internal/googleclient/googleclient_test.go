package googleclient

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aljrico/Google-Play-Console-CLI/internal/config"
)

func TestActiveProfileHTTPClientRejectsMissingProfile(t *testing.T) {
	t.Setenv("GPC_CONFIG", filepath.Join(t.TempDir(), "config.json"))

	_, err := ActiveProfileHTTPClient(context.Background(), "scope")
	if err == nil {
		t.Fatal("expected missing active profile error")
	}
	if !strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, want active profile error", err)
	}
}

func TestActiveProfileHTTPClientWrapsInvalidServiceAccount(t *testing.T) {
	tempDir := t.TempDir()
	serviceAccountFile := filepath.Join(tempDir, "service-account.json")
	if err := os.WriteFile(serviceAccountFile, []byte(`{"not":"a service account"}`), 0o600); err != nil {
		t.Fatalf("write service account: %v", err)
	}
	t.Setenv("GPC_CONFIG", filepath.Join(tempDir, "config.json"))
	if err := config.Save(config.Store{
		ActiveProfile: "default",
		Profiles: map[string]config.Profile{
			"default": {Name: "default", ServiceAccountFile: serviceAccountFile},
		},
	}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	_, err := ActiveProfileHTTPClient(context.Background(), "scope")
	if err == nil {
		t.Fatal("expected parse service account error")
	}
	if !strings.Contains(err.Error(), "parse service account file") {
		t.Fatalf("error = %v, want parse service account file", err)
	}
}
