package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/spf13/cobra"
)

func newPublishCommand(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Run high-level Google Play publishing workflows",
	}

	cmd.AddCommand(newPublishInternalCommand(out))
	return cmd
}

func newPublishInternalCommand(out io.Writer) *cobra.Command {
	var (
		packageName string
		bundlePath  string
		releaseName string
		status      string
		confirm     bool
		dryRun      bool
	)

	cmd := &cobra.Command{
		Use:   "internal",
		Short: "Publish an Android App Bundle to the internal track",
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(packageName)
			if err != nil {
				return err
			}
			typedStatus, err := play.NewReleaseStatus(status)
			if err != nil {
				return err
			}

			options := play.PublishInternalOptions{
				PackageName: typedPackageName,
				BundlePath:  bundlePath,
				ReleaseName: releaseName,
				Status:      typedStatus,
				Confirm:     confirm,
				DryRun:      dryRun,
			}
			if dryRun {
				result, err := play.PublishInternal(cmd.Context(), nil, options)
				if err != nil {
					return err
				}
				return output.Write(out, opts.output, opts.pretty, result)
			}

			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.PublishInternal(cmd.Context(), publisher, options)
			if err != nil {
				return err
			}
			return output.Write(out, opts.output, opts.pretty, result)
		},
	}

	cmd.Flags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.Flags().StringVar(&bundlePath, "aab", "", "Path to the Android App Bundle to upload")
	cmd.Flags().StringVar(&releaseName, "release-name", "", "Release name shown in Play Console")
	cmd.Flags().StringVar(&status, "status", play.ReleaseStatusCompleted.String(), "Release status: completed, draft, halted, inProgress")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Commit the edit after validation")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned publishing workflow without calling Google Play")
	return cmd
}
