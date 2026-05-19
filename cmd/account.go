package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/aljrico/Google-Play-Console-CLI/internal/config"
	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/spf13/cobra"
)

type accountStatusResult struct {
	ConfigPath               string   `json:"configPath"`
	Configured               bool     `json:"configured"`
	ActiveProfile            string   `json:"activeProfile,omitempty"`
	ServiceAccountFile       string   `json:"serviceAccountFile,omitempty"`
	ServiceAccountFileExists bool     `json:"serviceAccountFileExists"`
	ServiceAccountEmail      string   `json:"serviceAccountEmail,omitempty"`
	ProjectID                string   `json:"projectId,omitempty"`
	ServiceAccountValid      bool     `json:"serviceAccountValid"`
	Problems                 []string `json:"problems,omitempty"`
}

type serviceAccountMetadata struct {
	Type        string `json:"type"`
	ProjectID   string `json:"project_id"`
	PrivateKey  string `json:"private_key"`
	ClientEmail string `json:"client_email"`
}

func newAccountCommand(out io.Writer, options *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "Inspect local Google Play account configuration",
	}
	cmd.AddCommand(newAccountStatusCommand(out, options))
	return cmd
}

func newAccountStatusCommand(out io.Writer, options *globalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Summarize local account and service account health",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			status, err := buildAccountStatus()
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, status)
		},
	}
}

func buildAccountStatus() (accountStatusResult, error) {
	store, err := config.Load()
	if err != nil {
		return accountStatusResult{}, err
	}
	status := accountStatusResult{
		ConfigPath:          config.Path(),
		ActiveProfile:       store.ActiveProfile,
		Configured:          store.ActiveProfile != "" && len(store.Profiles) > 0,
		Problems:            make([]string, 0),
		ServiceAccountValid: false,
	}
	if store.ActiveProfile == "" {
		status.Problems = append(status.Problems, "no active auth profile; run gpc auth login")
		return status, nil
	}
	profile, ok := store.Profiles[store.ActiveProfile]
	if !ok {
		status.Problems = append(status.Problems, fmt.Sprintf("active profile %q is missing", store.ActiveProfile))
		return status, nil
	}
	status.ServiceAccountFile = profile.ServiceAccountFile
	metadata, problems := inspectServiceAccountFile(profile.ServiceAccountFile)
	status.Problems = append(status.Problems, problems...)
	status.ServiceAccountFileExists = metadata != nil
	if metadata != nil {
		status.ServiceAccountEmail = metadata.ClientEmail
		status.ProjectID = metadata.ProjectID
		status.ServiceAccountValid = isValidServiceAccountMetadata(*metadata)
	}
	return status, nil
}

func inspectServiceAccountFile(path string) (*serviceAccountMetadata, []string) {
	if path == "" {
		return nil, []string{"service account file is not configured"}
	}
	info, err := os.Lstat(path)
	switch {
	case err == nil:
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, []string{fmt.Sprintf("service account file cannot be a symlink: %s", path)}
		}
		if info.IsDir() {
			return nil, []string{fmt.Sprintf("service account file is a directory: %s", path)}
		}
	case os.IsNotExist(err):
		return nil, []string{fmt.Sprintf("service account file does not exist: %s", path)}
	default:
		return nil, []string{fmt.Sprintf("inspect service account file %s: %v", path, err)}
	}
	data, err := osReadFile(path)
	if err != nil {
		return nil, []string{fmt.Sprintf("read service account file %s: %v", path, err)}
	}
	var metadata serviceAccountMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, []string{fmt.Sprintf("parse service account file %s: %v", path, err)}
	}
	problems := make([]string, 0)
	if metadata.Type != "service_account" {
		problems = append(problems, "service account JSON type must be service_account")
	}
	if metadata.ClientEmail == "" {
		problems = append(problems, "service account JSON is missing client_email")
	}
	if metadata.ProjectID == "" {
		problems = append(problems, "service account JSON is missing project_id")
	}
	if metadata.PrivateKey == "" {
		problems = append(problems, "service account JSON is missing private_key")
	}
	return &metadata, problems
}

func isValidServiceAccountMetadata(metadata serviceAccountMetadata) bool {
	return metadata.Type == "service_account" &&
		metadata.ClientEmail != "" &&
		metadata.ProjectID != "" &&
		metadata.PrivateKey != ""
}
