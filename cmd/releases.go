package cmd

import (
	"fmt"
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
		fromTrack    string
		toTrack      string
		versionCode  int64
		rolloutTrack string
		confirm      bool
		dryRun       bool
	)

	cmd := &cobra.Command{
		Use:   "releases",
		Short: "Upload and manage Google Play releases",
	}

	cmd.PersistentFlags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	listCommand := &cobra.Command{
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
	}
	listCommand.Flags().StringVar(&trackName, "track", play.TrackInternal.String(), "Track name, for example internal, alpha, beta, or production")
	cmd.AddCommand(listCommand)
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
	uploadCommand.Flags().StringVar(&trackName, "track", play.TrackInternal.String(), "Track name, for example internal, alpha, beta, or production")
	uploadCommand.Flags().StringVar(&releaseName, "release-name", "", "Release name shown in Play Console")
	uploadCommand.Flags().StringVar(&status, "status", play.ReleaseStatusCompleted.String(), "Release status: completed, draft, halted, inProgress")
	uploadCommand.Flags().Float64Var(&userFraction, "user-fraction", 0, "Staged rollout fraction for inProgress or halted releases")
	uploadCommand.Flags().BoolVar(&confirm, "confirm", false, "Commit the edit after validation")
	uploadCommand.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned release upload workflow without calling Google Play")
	cmd.AddCommand(uploadCommand)

	promoteCommand := &cobra.Command{
		Use:   "promote",
		Short: "Promote a release from one track to another",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(packageName)
			if err != nil {
				return err
			}
			typedFromTrack, err := play.NewTrackName(fromTrack)
			if err != nil {
				return err
			}
			typedToTrack, err := play.NewTrackName(toTrack)
			if err != nil {
				return err
			}
			typedStatus, err := play.NewReleaseStatus(status)
			if err != nil {
				return err
			}

			promoteOptions := play.PromoteReleaseOptions{
				PackageName: typedPackageName,
				FromTrack:   typedFromTrack,
				ToTrack:     typedToTrack,
				VersionCode: versionCode,
				Status:      typedStatus,
				Confirm:     confirm,
				DryRun:      dryRun,
			}
			if cmd.Flags().Changed("user-fraction") {
				promoteOptions.UserFraction = &userFraction
			}
			if dryRun {
				result, err := play.PromoteRelease(cmd.Context(), nil, promoteOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}

			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.PromoteRelease(cmd.Context(), publisher, promoteOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	promoteCommand.Flags().StringVar(&fromTrack, "from", play.TrackInternal.String(), "Source track name")
	promoteCommand.Flags().StringVar(&toTrack, "to", play.TrackProduction.String(), "Target track name")
	promoteCommand.Flags().Int64Var(&versionCode, "version-code", 0, "Version code to promote")
	promoteCommand.Flags().StringVar(&status, "status", play.ReleaseStatusDraft.String(), "Target release status: completed, draft, halted, inProgress")
	promoteCommand.Flags().Float64Var(&userFraction, "user-fraction", 0, "Staged rollout fraction for inProgress or halted releases")
	promoteCommand.Flags().BoolVar(&confirm, "confirm", false, "Commit the edit after validation")
	promoteCommand.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned promotion workflow without calling Google Play")
	cmd.AddCommand(promoteCommand)

	haltCommand := newRolloutCommand(out, options, play.RolloutActionHalt, &packageName, &rolloutTrack, &versionCode, &userFraction, &confirm, &dryRun)
	haltCommand.Flags().StringVar(&rolloutTrack, "track", play.TrackProduction.String(), "Track name")
	haltCommand.Flags().Int64Var(&versionCode, "version-code", 0, "Version code to halt")
	haltCommand.Flags().BoolVar(&confirm, "confirm", false, "Commit the edit after validation")
	haltCommand.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned halt workflow without calling Google Play")
	cmd.AddCommand(haltCommand)

	resumeCommand := newRolloutCommand(out, options, play.RolloutActionResume, &packageName, &rolloutTrack, &versionCode, &userFraction, &confirm, &dryRun)
	resumeCommand.Flags().StringVar(&rolloutTrack, "track", play.TrackProduction.String(), "Track name")
	resumeCommand.Flags().Int64Var(&versionCode, "version-code", 0, "Version code to resume")
	resumeCommand.Flags().Float64Var(&userFraction, "user-fraction", 0, "Staged rollout fraction for the resumed release")
	resumeCommand.Flags().BoolVar(&confirm, "confirm", false, "Commit the edit after validation")
	resumeCommand.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned resume workflow without calling Google Play")
	cmd.AddCommand(resumeCommand)

	return cmd
}

func newRolloutCommand(out io.Writer, options *globalOptions, action play.RolloutAction, packageName *string, trackName *string, versionCode *int64, userFraction *float64, confirm *bool, dryRun *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   action.String(),
		Short: fmt.Sprintf("%s a staged release", action.Title()),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			typedTrackName, err := play.NewTrackName(*trackName)
			if err != nil {
				return err
			}

			rolloutOptions := play.RolloutOptions{
				PackageName: typedPackageName,
				Track:       typedTrackName,
				VersionCode: *versionCode,
				Action:      action,
				Confirm:     *confirm,
				DryRun:      *dryRun,
			}
			if cmd.Flags().Changed("user-fraction") {
				rolloutOptions.UserFraction = userFraction
			}
			if *dryRun {
				result, err := play.UpdateRollout(cmd.Context(), nil, rolloutOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}

			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.UpdateRollout(cmd.Context(), publisher, rolloutOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	return cmd
}
