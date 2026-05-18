package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Store struct {
	ActiveProfile string             `json:"activeProfile"`
	Profiles      map[string]Profile `json:"profiles"`
}

type Profile struct {
	Name               string `json:"name"`
	ServiceAccountFile string `json:"serviceAccountFile"`
}

func Path() string {
	if override := os.Getenv("GPC_CONFIG"); override != "" {
		return override
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".gpc/config.json"
	}
	return filepath.Join(home, ".gpc", "config.json")
}

func Load() (Store, error) {
	path := Path()
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Store{Profiles: map[string]Profile{}}, nil
	}
	if err != nil {
		return Store{}, fmt.Errorf("read config: %w", err)
	}

	var store Store
	if err := json.Unmarshal(data, &store); err != nil {
		return Store{}, fmt.Errorf("parse config: %w", err)
	}
	if store.Profiles == nil {
		store.Profiles = map[string]Profile{}
	}
	return store, nil
}

func Save(store Store) error {
	path := Path()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}
