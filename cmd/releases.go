package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/spf13/cobra"
)

func newReleasesCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var (
		packageName  string
		trackName    string
		bundlePath   string
		releaseName  string
		status       string
		userFraction float64
		confirm      bool
		dryRun       bool
	)

	cmd := &cobra.Command{
		Use:   "releases",
		Short: "Upload and manage Google Play releases",
	}

	cmd.PersistentFlags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.PersistentFlags().StringVar(&trackName, "track", play.TrackInternal.String(), "Track name, for example internal, alpha, beta, or production")
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List releases for a track",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(packageName)
			if err != nil {
				return err
			}
			typedTrackName, err := play.NewTrackName(trackName)
			if err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			releases, err := play.ListTrackReleases(cmd.Context(), publisher, typedPackageName, typedTrackName)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, releases)
		},
	})
	uploadCommand := &cobra.Command{
		Use:   "upload",
		Short: "Upload an Android App Bundle to a track",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(packageName)
			if err != nil {
				return err
			}
			typedTrackName, err := play.NewTrackName(trackName)
			if err != nil {
				return err
			}
			typedStatus, err := play.NewReleaseStatus(status)
			if err != nil {
				return err
			}

			publishOptions := play.PublishInternalOptions{
				PackageName: typedPackageName,
				Track:       typedTrackName,
				BundlePath:  bundlePath,
				ReleaseName: releaseName,
				Status:      typedStatus,
				Confirm:     confirm,
				DryRun:      dryRun,
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

	uploadCommand.Flags().StringVar(&bundlePath, "aab", "", "Path to the Android App Bundle to upload")
	uploadCommand.Flags().StringVar(&releaseName, "release-name", "", "Release name shown in Play Console")
	uploadCommand.Flags().StringVar(&status, "status", play.ReleaseStatusCompleted.String(), "Release status: completed, draft, halted, inProgress")
	uploadCommand.Flags().Float64Var(&userFraction, "user-fraction", 0, "Staged rollout fraction for inProgress or halted releases")
	uploadCommand.Flags().BoolVar(&confirm, "confirm", false, "Commit the edit after validation")
	uploadCommand.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned release upload workflow without calling Google Play")
	cmd.AddCommand(uploadCommand)

	return cmd
}
