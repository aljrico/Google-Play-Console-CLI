package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/spf13/cobra"
)

func newSystemAPKsCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var packageName string

	cmd := &cobra.Command{
		Use:   "system-apks",
		Short: "Inspect Google Play system APK variants",
	}
	cmd.PersistentFlags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.AddCommand(newSystemAPKVariantsCommand(out, options, &packageName))
	return cmd
}

func newSystemAPKVariantsCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "variants",
		Short: "Inspect generated system APK variants",
	}
	cmd.AddCommand(newSystemAPKVariantsListCommand(out, options, packageName))
	return cmd
}

func newSystemAPKVariantsListCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var versionCode int64

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List system APK variants for an App Bundle version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			listOptions := play.SystemAPKVariantListOptions{
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
			result, err := play.ListSystemAPKVariants(cmd.Context(), publisher, listOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().Int64Var(&versionCode, "version-code", 0, "App Bundle version code")
	return cmd
}
