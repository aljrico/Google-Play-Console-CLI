package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/spf13/cobra"
)

func newPublishCommand(out io.Writer, options *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Run high-level Google Play publishing workflows",
	}

	cmd.AddCommand(newPublishInternalCommand(out, options))
	return cmd
}

func newPublishInternalCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var (
		packageName  string
		bundlePath   string
		releaseName  string
		releaseNotes []string
		status       string
		userFraction float64
		confirm      bool
		dryRun       bool
	)

	cmd := &cobra.Command{
		Use:   "internal",
		Short: "Publish an Android App Bundle to the internal track",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(packageName)
			if err != nil {
				return err
			}
			typedStatus, err := play.NewReleaseStatus(status)
			if err != nil {
				return err
			}
			typedReleaseNotes, err := parseReleaseNotes(releaseNotes)
			if err != nil {
				return err
			}

			publishOptions := play.PublishInternalOptions{
				PackageName:  typedPackageName,
				Track:        play.TrackInternal,
				BundlePath:   bundlePath,
				ReleaseName:  releaseName,
				Status:       typedStatus,
				ReleaseNotes: typedReleaseNotes,
				Confirm:      confirm,
				DryRun:       dryRun,
			}
			if cmd.Flags().Changed("user-fraction") {
				publishOptions.UserFraction = &userFraction
			}
			if dryRun {
				result, err := play.PublishInternal(cmd.Context(), nil, publishOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}

			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.PublishInternal(cmd.Context(), publisher, publishOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}

	cmd.Flags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.Flags().StringVar(&bundlePath, "aab", "", "Path to the Android App Bundle to upload")
	cmd.Flags().StringVar(&releaseName, "release-name", "", "Release name shown in Play Console")
	cmd.Flags().StringArrayVar(&releaseNotes, "release-note", nil, "Localized release note as language=text, repeatable")
	cmd.Flags().StringVar(&status, "status", play.ReleaseStatusCompleted.String(), "Release status: completed, draft, halted, inProgress")
	cmd.Flags().Float64Var(&userFraction, "user-fraction", 0, "Staged rollout fraction for inProgress or halted releases")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Commit the edit after validation")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned publishing workflow without calling Google Play")
	return cmd
}
