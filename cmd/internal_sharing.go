package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/spf13/cobra"
)

func newInternalSharingCommand(out io.Writer, options *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "internal-sharing",
		Short: "Upload artifacts to Google Play internal app sharing",
	}
	cmd.AddCommand(newInternalSharingUploadCommand(out, options))
	return cmd
}

func newInternalSharingUploadCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var (
		packageName string
		apkPath     string
		bundlePath  string
		dryRun      bool
	)

	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload an APK or Android App Bundle to internal app sharing",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(packageName)
			if err != nil {
				return err
			}
			uploadOptions := play.InternalSharingUploadOptions{
				PackageName: typedPackageName,
				APKPath:     apkPath,
				BundlePath:  bundlePath,
				DryRun:      dryRun,
			}
			if err := uploadOptions.Validate(); err != nil {
				return err
			}
			if dryRun {
				result, err := play.UploadInternalSharingArtifact(cmd.Context(), nil, uploadOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			if err := play.ValidateReadableFile(uploadOptionsPath(uploadOptions)); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.UploadInternalSharingArtifact(cmd.Context(), publisher, uploadOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.Flags().StringVar(&apkPath, "apk", "", "Path to the APK to upload")
	cmd.Flags().StringVar(&bundlePath, "aab", "", "Path to the Android App Bundle to upload")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned internal sharing upload without calling Google Play")
	return cmd
}

func uploadOptionsPath(options play.InternalSharingUploadOptions) string {
	if options.APKPath != "" {
		return options.APKPath
	}
	return options.BundlePath
}
