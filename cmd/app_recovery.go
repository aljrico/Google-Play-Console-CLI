package cmd

import (
	"fmt"
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/spf13/cobra"
)

func newAppRecoveryCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var packageName string

	cmd := &cobra.Command{
		Use:   "app-recovery",
		Short: "Inspect and manage Google Play app recovery actions",
	}
	cmd.PersistentFlags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.AddCommand(
		newAppRecoveryListCommand(out, options, &packageName),
		newAppRecoveryAddTargetingCommand(out, options, &packageName),
		newAppRecoveryDeployCommand(out, options, &packageName),
		newAppRecoveryCancelCommand(out, options, &packageName),
	)
	return cmd
}

func newAppRecoveryAddTargetingCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		appRecoveryID string
		allUsers      bool
		sdkLevels     []int64
		regionCodes   []string
		confirm       bool
		dryRun        bool
	)

	cmd := &cobra.Command{
		Use:   "add-targeting",
		Short: "Add targeting to an app recovery action",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			typedAppRecoveryID, err := play.NewAppRecoveryID(appRecoveryID)
			if err != nil {
				return err
			}
			targetingOptions := play.AppRecoveryTargetingUpdateOptions{
				PackageName:   typedPackageName,
				AppRecoveryID: typedAppRecoveryID,
				AllUsers:      allUsers,
				SDKLevels:     sdkLevels,
				RegionCodes:   regionCodes,
				Confirm:       confirm,
				DryRun:        dryRun,
			}
			if err := targetingOptions.Validate(); err != nil {
				return err
			}
			if dryRun {
				result, err := play.AddAppRecoveryTargeting(cmd.Context(), nil, targetingOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.AddAppRecoveryTargeting(cmd.Context(), publisher, targetingOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&appRecoveryID, "id", "", "App recovery action ID")
	cmd.Flags().BoolVar(&allUsers, "all-users", false, "Target all users")
	cmd.Flags().Int64SliceVar(&sdkLevels, "sdk-level", nil, "Android SDK level to add, repeatable")
	cmd.Flags().StringSliceVar(&regionCodes, "region", nil, "ISO 3166-1 alpha-2 region code to add, repeatable")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the app recovery targeting update")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned app recovery targeting update without calling Google Play")
	return cmd
}

func newAppRecoveryListCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var versionCode int64

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List app recovery actions for a version code",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			listOptions := play.AppRecoveryListOptions{
				PackageName: typedPackageName,
				VersionCode: versionCode,
			}
			if err := listOptions.Validate(); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.ListAppRecoveries(cmd.Context(), publisher, listOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().Int64Var(&versionCode, "version-code", 0, "Version code targeted by recovery actions")
	return cmd
}

func newAppRecoveryDeployCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	return newAppRecoveryMutationCommand(out, options, packageName, "deploy", "Deploy a draft app recovery action")
}

func newAppRecoveryCancelCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	return newAppRecoveryMutationCommand(out, options, packageName, "cancel", "Cancel an app recovery action")
}

func newAppRecoveryMutationCommand(out io.Writer, options *globalOptions, packageName *string, action string, short string) *cobra.Command {
	var (
		appRecoveryID string
		confirm       bool
		dryRun        bool
	)

	cmd := &cobra.Command{
		Use:   action,
		Short: short,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			typedAppRecoveryID, err := play.NewAppRecoveryID(appRecoveryID)
			if err != nil {
				return err
			}
			mutationOptions := play.AppRecoveryMutationOptions{
				PackageName:   typedPackageName,
				AppRecoveryID: typedAppRecoveryID,
				Confirm:       confirm,
				DryRun:        dryRun,
			}
			if err := mutationOptions.Validate(); err != nil {
				return err
			}
			if dryRun {
				return runAppRecoveryMutationDryRun(cmd, out, options, mutationOptions, action)
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			return runAppRecoveryMutation(cmd, out, options, publisher, mutationOptions, action)
		},
	}
	cmd.Flags().StringVar(&appRecoveryID, "id", "", "App recovery action ID")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the app recovery mutation")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned app recovery mutation without calling Google Play")
	return cmd
}

func runAppRecoveryMutationDryRun(cmd *cobra.Command, out io.Writer, options *globalOptions, mutationOptions play.AppRecoveryMutationOptions, action string) error {
	return runAppRecoveryMutation(cmd, out, options, nil, mutationOptions, action)
}

func runAppRecoveryMutation(cmd *cobra.Command, out io.Writer, options *globalOptions, mutator play.AppRecoveryMutator, mutationOptions play.AppRecoveryMutationOptions, action string) error {
	var (
		result play.AppRecoveryMutationResult
		err    error
	)
	switch action {
	case "deploy":
		result, err = play.DeployAppRecovery(cmd.Context(), mutator, mutationOptions)
	case "cancel":
		result, err = play.CancelAppRecovery(cmd.Context(), mutator, mutationOptions)
	default:
		err = fmt.Errorf("unsupported app recovery action %q", action)
	}
	if err != nil {
		return err
	}
	return output.Write(out, options.output, options.pretty, result)
}
