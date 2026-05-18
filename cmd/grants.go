package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/spf13/cobra"
)

func newGrantsCommand(out io.Writer, options *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grants",
		Short: "Manage Google Play app access grants",
	}
	cmd.AddCommand(
		newGrantsCreateCommand(out, options),
		newGrantsPatchCommand(out, options),
		newGrantsDeleteCommand(out, options),
	)
	return cmd
}

func newGrantsCreateCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var (
		developer   string
		userEmail   string
		packageName string
		permissions []string
		confirm     bool
		dryRun      bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create app-level access for a Play Console user",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedDeveloper, err := play.NewDeveloperAccount(developer)
			if err != nil {
				return err
			}
			typedUserEmail, err := play.NewGrantUserEmail(userEmail)
			if err != nil {
				return err
			}
			typedPackageName, err := play.NewPackageName(packageName)
			if err != nil {
				return err
			}
			typedPermissions, err := parseGrantPermissions(permissions)
			if err != nil {
				return err
			}
			createOptions := play.GrantCreateOptions{
				Developer:   typedDeveloper,
				UserEmail:   typedUserEmail,
				PackageName: typedPackageName,
				Permissions: typedPermissions,
				Confirm:     confirm,
				DryRun:      dryRun,
			}
			return runGrantCreate(cmd, out, options, createOptions)
		},
	}
	cmd.Flags().StringVar(&developer, "developer", "", "Developer account ID or resource, for example 1234567890 or developers/1234567890")
	cmd.Flags().StringVar(&userEmail, "user-email", "", "Play Console user email")
	cmd.Flags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.Flags().StringArrayVar(&permissions, "permission", nil, "App-level grant permission, repeatable")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the grant creation")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned grant creation without calling Google Play")
	return cmd
}

func newGrantsPatchCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var (
		name        string
		permissions []string
		confirm     bool
		dryRun      bool
	)

	cmd := &cobra.Command{
		Use:   "patch",
		Short: "Replace app-level permissions for an access grant",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedName, err := play.NewGrantName(name)
			if err != nil {
				return err
			}
			typedPermissions, err := parseGrantPermissions(permissions)
			if err != nil {
				return err
			}
			patchOptions := play.GrantPatchOptions{
				Name:        typedName,
				Permissions: typedPermissions,
				Confirm:     confirm,
				DryRun:      dryRun,
			}
			return runGrantPatch(cmd, out, options, patchOptions)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Grant resource name, developers/{developer}/users/{email}/grants/{package}")
	cmd.Flags().StringArrayVar(&permissions, "permission", nil, "App-level grant permission, repeatable")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the grant patch")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned grant patch without calling Google Play")
	return cmd
}

func newGrantsDeleteCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var (
		name    string
		confirm bool
		dryRun  bool
	)

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an app-level access grant",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedName, err := play.NewGrantName(name)
			if err != nil {
				return err
			}
			deleteOptions := play.GrantDeleteOptions{
				Name:    typedName,
				Confirm: confirm,
				DryRun:  dryRun,
			}
			return runGrantDelete(cmd, out, options, deleteOptions)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Grant resource name, developers/{developer}/users/{email}/grants/{package}")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the grant deletion")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned grant deletion without calling Google Play")
	return cmd
}

func runGrantCreate(cmd *cobra.Command, out io.Writer, options *globalOptions, createOptions play.GrantCreateOptions) error {
	if createOptions.DryRun {
		result, err := play.CreateGrant(cmd.Context(), nil, createOptions)
		if err != nil {
			return err
		}
		return output.Write(out, options.output, options.pretty, result)
	}
	if err := createOptions.Validate(); err != nil {
		return err
	}
	publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
	if err != nil {
		return err
	}
	result, err := play.CreateGrant(cmd.Context(), publisher, createOptions)
	if err != nil {
		return err
	}
	return output.Write(out, options.output, options.pretty, result)
}

func runGrantPatch(cmd *cobra.Command, out io.Writer, options *globalOptions, patchOptions play.GrantPatchOptions) error {
	if patchOptions.DryRun {
		result, err := play.PatchGrant(cmd.Context(), nil, patchOptions)
		if err != nil {
			return err
		}
		return output.Write(out, options.output, options.pretty, result)
	}
	if err := patchOptions.Validate(); err != nil {
		return err
	}
	publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
	if err != nil {
		return err
	}
	result, err := play.PatchGrant(cmd.Context(), publisher, patchOptions)
	if err != nil {
		return err
	}
	return output.Write(out, options.output, options.pretty, result)
}

func runGrantDelete(cmd *cobra.Command, out io.Writer, options *globalOptions, deleteOptions play.GrantDeleteOptions) error {
	if deleteOptions.DryRun {
		result, err := play.DeleteGrant(cmd.Context(), nil, deleteOptions)
		if err != nil {
			return err
		}
		return output.Write(out, options.output, options.pretty, result)
	}
	if err := deleteOptions.Validate(); err != nil {
		return err
	}
	publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
	if err != nil {
		return err
	}
	result, err := play.DeleteGrant(cmd.Context(), publisher, deleteOptions)
	if err != nil {
		return err
	}
	return output.Write(out, options.output, options.pretty, result)
}

func parseGrantPermissions(values []string) ([]play.GrantPermission, error) {
	permissions := make([]play.GrantPermission, 0, len(values))
	for _, value := range values {
		permission, err := play.NewGrantPermission(value)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	return permissions, nil
}
