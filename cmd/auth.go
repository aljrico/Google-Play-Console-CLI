package cmd

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/aljrico/Google-Play-Console-CLI/internal/config"
	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2/google"
)

type authStatus struct {
	ActiveProfile  string `json:"activeProfile"`
	ServiceAccount string `json:"serviceAccount"`
	ConfigPath     string `json:"configPath"`
}

func newAuthCommand(out io.Writer, options *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage Google Play API authentication",
	}

	cmd.AddCommand(newAuthLoginCommand(out, options), newAuthStatusCommand(out, options), newAuthDoctorCommand(out, options))
	return cmd
}

func newAuthLoginCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var (
		name               string
		serviceAccountFile string
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Store a service account profile",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			if serviceAccountFile == "" {
				return fmt.Errorf("--service-account is required")
			}
			absoluteServiceAccountFile, err := filepath.Abs(serviceAccountFile)
			if err != nil {
				return fmt.Errorf("resolve service account path: %w", err)
			}
			if err := validateServiceAccount(cmd.Context(), absoluteServiceAccountFile); err != nil {
				return err
			}

			store, err := config.Load()
			if err != nil {
				return err
			}
			store.ActiveProfile = name
			store.Profiles[name] = config.Profile{
				Name:               name,
				ServiceAccountFile: absoluteServiceAccountFile,
			}
			if err := config.Save(store); err != nil {
				return err
			}

			return output.Write(out, options.output, options.pretty, authStatus{
				ActiveProfile:  name,
				ServiceAccount: absoluteServiceAccountFile,
				ConfigPath:     config.Path(),
			})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Profile name")
	cmd.Flags().StringVar(&serviceAccountFile, "service-account", "", "Path to a Google service account JSON key")
	return cmd
}

func newAuthStatusCommand(out io.Writer, options *globalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show the active auth profile",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := config.Load()
			if err != nil {
				return err
			}
			profile, ok := store.Profiles[store.ActiveProfile]
			if !ok || store.ActiveProfile == "" {
				return fmt.Errorf("no active auth profile; run playpub auth login")
			}
			return output.Write(out, options.output, options.pretty, authStatus{
				ActiveProfile:  store.ActiveProfile,
				ServiceAccount: profile.ServiceAccountFile,
				ConfigPath:     config.Path(),
			})
		},
	}
}

func newAuthDoctorCommand(out io.Writer, options *globalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Validate the active auth profile",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := config.Load()
			if err != nil {
				return err
			}
			profile, ok := store.Profiles[store.ActiveProfile]
			if !ok || store.ActiveProfile == "" {
				return fmt.Errorf("no active auth profile; run playpub auth login")
			}
			if err := validateServiceAccount(cmd.Context(), profile.ServiceAccountFile); err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, map[string]string{
				"status":        "ok",
				"activeProfile": store.ActiveProfile,
			})
		},
	}
}

func validateServiceAccount(ctx context.Context, path string) error {
	data, err := osReadFile(path)
	if err != nil {
		return fmt.Errorf("read service account file: %w", err)
	}
	if _, err := google.CredentialsFromJSON(ctx, data, "https://www.googleapis.com/auth/androidpublisher"); err != nil {
		return fmt.Errorf("parse service account file: %w", err)
	}
	return nil
}
