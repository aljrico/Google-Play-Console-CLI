package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/spf13/cobra"
)

func newGeneratedAPKsCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var packageName string

	cmd := &cobra.Command{
		Use:   "generated-apks",
		Short: "Inspect generated APK metadata for an App Bundle version",
	}
	cmd.PersistentFlags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.AddCommand(
		newGeneratedAPKsListCommand(out, options, &packageName),
		newGeneratedAPKsDownloadCommand(out, options, &packageName),
	)
	return cmd
}

func newGeneratedAPKsListCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var versionCode int64

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List generated APKs for a version code",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			listOptions := play.GeneratedAPKListOptions{
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
			result, err := play.ListGeneratedAPKs(cmd.Context(), publisher, listOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().Int64Var(&versionCode, "version-code", 0, "App Bundle version code")
	return cmd
}

func newGeneratedAPKsDownloadCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		versionCode int64
		downloadID  string
		outputPath  string
		force       bool
		dryRun      bool
	)

	cmd := &cobra.Command{
		Use:   "download",
		Short: "Download one generated APK by download ID",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			typedDownloadID, err := play.NewGeneratedAPKDownloadID(downloadID)
			if err != nil {
				return err
			}
			downloadOptions := play.GeneratedAPKDownloadOptions{
				PackageName: typedPackageName,
				VersionCode: versionCode,
				DownloadID:  typedDownloadID,
				OutputPath:  outputPath,
				Force:       force,
				DryRun:      dryRun,
			}
			if err := downloadOptions.Validate(); err != nil {
				return err
			}
			if dryRun {
				result, err := play.DownloadGeneratedAPK(cmd.Context(), nil, downloadOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.DownloadGeneratedAPK(cmd.Context(), publisher, downloadOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().Int64Var(&versionCode, "version-code", 0, "App Bundle version code")
	cmd.Flags().StringVar(&downloadID, "download-id", "", "Generated APK download ID from generated-apks list")
	cmd.Flags().StringVar(&outputPath, "file", "", "Destination .apk path")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite the destination file")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned download without calling Google Play")
	return cmd
}
