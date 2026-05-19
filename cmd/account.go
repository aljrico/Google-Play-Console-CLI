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
	ServiceAccountReadable   bool     `json:"serviceAccountReadable"`
	ServiceAccountJSONParsed bool     `json:"serviceAccountJsonParsed"`
	ServiceAccountEmail      string   `json:"serviceAccountEmail,omitempty"`
	ProjectID                string   `json:"projectId,omitempty"`
	ServiceAccountMetadataOK bool     `json:"serviceAccountMetadataOk"`
	Problems                 []string `json:"problems,omitempty"`
}

type serviceAccountMetadata struct {
	Type        string `json:"type"`
	ProjectID   string `json:"project_id"`
	PrivateKey  string `json:"private_key"`
	ClientEmail string `json:"client_email"`
}

type serviceAccountInspection struct {
	Exists   bool
	Readable bool
	Parsed   bool
	Metadata *serviceAccountMetadata
	Problems []string
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
		Short: "Summarize local account and service account metadata",
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
		ConfigPath: config.Path(),
		Problems:   make([]string, 0),
	}
	if store.ActiveProfile == "" {
		status.Problems = append(status.Problems, "no active auth profile; run playpub auth login")
		return status, nil
	}
	status.ActiveProfile = store.ActiveProfile
	profile, ok := store.Profiles[store.ActiveProfile]
	if !ok {
		status.Problems = append(status.Problems, fmt.Sprintf("active profile %q is missing", store.ActiveProfile))
		return status, nil
	}

	status.Configured = true
	status.ServiceAccountFile = profile.ServiceAccountFile
	inspection := inspectServiceAccountFile(profile.ServiceAccountFile)
	status.Problems = append(status.Problems, inspection.Problems...)
	status.ServiceAccountFileExists = inspection.Exists
	status.ServiceAccountReadable = inspection.Readable
	status.ServiceAccountJSONParsed = inspection.Parsed
	if inspection.Metadata != nil {
		status.ServiceAccountEmail = inspection.Metadata.ClientEmail
		status.ProjectID = inspection.Metadata.ProjectID
		status.ServiceAccountMetadataOK = isCompleteServiceAccountMetadata(*inspection.Metadata)
	}
	return status, nil
}

func inspectServiceAccountFile(path string) serviceAccountInspection {
	inspection := serviceAccountInspection{Problems: make([]string, 0)}
	if path == "" {
		inspection.Problems = append(inspection.Problems, "service account file is not configured")
		return inspection
	}

	info, err := os.Stat(path)
	switch {
	case err == nil:
		inspection.Exists = true
		if info.IsDir() {
			inspection.Problems = append(inspection.Problems, fmt.Sprintf("service account file is a directory: %s", path))
			return inspection
		}
	case os.IsNotExist(err):
		inspection.Problems = append(inspection.Problems, fmt.Sprintf("service account file does not exist: %s", path))
		return inspection
	default:
		inspection.Problems = append(inspection.Problems, fmt.Sprintf("inspect service account file %s: %v", path, err))
		return inspection
	}

	data, err := osReadFile(path)
	if err != nil {
		inspection.Problems = append(inspection.Problems, fmt.Sprintf("read service account file %s: %v", path, err))
		return inspection
	}
	inspection.Readable = true

	var metadata serviceAccountMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		inspection.Problems = append(inspection.Problems, fmt.Sprintf("parse service account file %s: %v", path, err))
		return inspection
	}
	inspection.Parsed = true
	inspection.Metadata = &metadata

	if metadata.Type != "service_account" {
		inspection.Problems = append(inspection.Problems, "service account JSON type must be service_account")
	}
	if metadata.ClientEmail == "" {
		inspection.Problems = append(inspection.Problems, "service account JSON is missing client_email")
	}
	if metadata.ProjectID == "" {
		inspection.Problems = append(inspection.Problems, "service account JSON is missing project_id")
	}
	if metadata.PrivateKey == "" {
		inspection.Problems = append(inspection.Problems, "service account JSON is missing private_key")
	}
	return inspection
}

func isCompleteServiceAccountMetadata(metadata serviceAccountMetadata) bool {
	return metadata.Type == "service_account" &&
		metadata.ClientEmail != "" &&
		metadata.ProjectID != "" &&
		metadata.PrivateKey != ""
}
