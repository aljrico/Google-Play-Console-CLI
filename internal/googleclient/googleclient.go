package googleclient

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2/google"

	"github.com/aljrico/Google-Play-Console-CLI/internal/config"
)

func ActiveProfileHTTPClient(ctx context.Context, scopes ...string) (*http.Client, error) {
	store, err := config.Load()
	if err != nil {
		return nil, err
	}
	profile, ok := store.Profiles[store.ActiveProfile]
	if !ok || store.ActiveProfile == "" {
		return nil, fmt.Errorf("no active auth profile; run gpc auth login")
	}
	credentialsJSON, err := os.ReadFile(profile.ServiceAccountFile)
	if err != nil {
		return nil, fmt.Errorf("read service account file %s: %w", profile.ServiceAccountFile, err)
	}
	jwtConfig, err := google.JWTConfigFromJSON(credentialsJSON, scopes...)
	if err != nil {
		return nil, fmt.Errorf("parse service account file %s: %w", profile.ServiceAccountFile, err)
	}
	return jwtConfig.Client(ctx), nil
}
