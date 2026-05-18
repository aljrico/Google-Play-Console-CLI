package cmd

import (
	"context"
	"fmt"
	"io"

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

func newAuthCommand(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage Google Play API authentication",
	}

	cmd.AddCommand(newAuthLoginCommand(out), newAuthStatusCommand(out), newAuthDoctorCommand(out))
	return cmd
}

func newAuthLoginCommand(out io.Writer) *cobra.Command {
	var (
		name               string
		serviceAccountFile string
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Store a service account profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			if serviceAccountFile == "" {
				return fmt.Errorf("--service-account is required")
			}
			if err := validateServiceAccount(cmd.Context(), serviceAccountFile); err != nil {
				return err
			}

			store, err := config.Load()
			if err != nil {
				return err
			}
			store.ActiveProfile = name
			store.Profiles[name] = config.Profile{
				Name:               name,
				ServiceAccountFile: serviceAccountFile,
			}
			if err := config.Save(store); err != nil {
				return err
			}

			return output.Write(out, opts.output, opts.pretty, authStatus{
				ActiveProfile:  name,
				ServiceAccount: serviceAccountFile,
				ConfigPath:     config.Path(),
			})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Profile name")
	cmd.Flags().StringVar(&serviceAccountFile, "service-account", "", "Path to a Google service account JSON key")
	return cmd
}

func newAuthStatusCommand(out io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show the active auth profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := config.Load()
			if err != nil {
				return err
			}
			profile, ok := store.Profiles[store.ActiveProfile]
			if !ok || store.ActiveProfile == "" {
				return fmt.Errorf("no active auth profile; run gpc auth login")
			}
			return output.Write(out, opts.output, opts.pretty, authStatus{
				ActiveProfile:  store.ActiveProfile,
				ServiceAccount: profile.ServiceAccountFile,
				ConfigPath:     config.Path(),
			})
		},
	}
}

func newAuthDoctorCommand(out io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Validate the active auth profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := config.Load()
			if err != nil {
				return err
			}
			profile, ok := store.Profiles[store.ActiveProfile]
			if !ok || store.ActiveProfile == "" {
				return fmt.Errorf("no active auth profile; run gpc auth login")
			}
			if err := validateServiceAccount(cmd.Context(), profile.ServiceAccountFile); err != nil {
				return err
			}
			return output.Write(out, opts.output, opts.pretty, map[string]string{
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
