package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/spf13/cobra"
)

func newAppRecoveryCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var packageName string

	cmd := &cobra.Command{
		Use:   "app-recovery",
		Short: "Inspect Google Play app recovery actions",
	}
	cmd.PersistentFlags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.AddCommand(newAppRecoveryListCommand(out, options, &packageName))
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
