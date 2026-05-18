package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/spf13/cobra"
)

func newReleasesCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var packageName string

	cmd := &cobra.Command{
		Use:   "releases",
		Short: "Upload and manage Google Play releases",
	}
	cmd.PersistentFlags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.AddCommand(
		newReleasesListCommand(out, options, &packageName),
		newReleasesUploadCommand(out, options, &packageName),
		newReleasesPromoteCommand(out, options, &packageName),
		newReleasesHaltCommand(out, options, &packageName),
		newReleasesResumeCommand(out, options, &packageName),
	)
	return cmd
}

func newReleasesListCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var trackName string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List releases for a track",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
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
	cmd.Flags().StringVar(&trackName, "track", play.TrackInternal.String(), "Track name, for example internal, alpha, beta, or production")
	return cmd
}

func newReleasesUploadCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		trackName    string
		bundlePath   string
		releaseName  string
		releaseNotes []string
		status       string
		userFraction float64
		confirm      bool
		dryRun       bool
	)

	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload an Android App Bundle to a track",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
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
			typedReleaseNotes, err := parseReleaseNotes(releaseNotes)
			if err != nil {
				return err
			}

			publishOptions := play.PublishInternalOptions{
				PackageName:  typedPackageName,
				Track:        typedTrackName,
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

	cmd.Flags().StringVar(&bundlePath, "aab", "", "Path to the Android App Bundle to upload")
	cmd.Flags().StringVar(&trackName, "track", play.TrackInternal.String(), "Track name, for example internal, alpha, beta, or production")
	cmd.Flags().StringVar(&releaseName, "release-name", "", "Release name shown in Play Console")
	cmd.Flags().StringArrayVar(&releaseNotes, "release-note", nil, "Localized release note as language=text, repeatable")
	cmd.Flags().StringVar(&status, "status", play.ReleaseStatusCompleted.String(), "Release status: completed, draft, halted, inProgress")
	cmd.Flags().Float64Var(&userFraction, "user-fraction", 0, "Staged rollout fraction for inProgress or halted releases")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Commit the edit after validation")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned release upload workflow without calling Google Play")
	return cmd
}

func parseReleaseNotes(values []string) ([]play.ReleaseNote, error) {
	notes := make([]play.ReleaseNote, 0, len(values))
	for _, value := range values {
		language, text, ok := strings.Cut(value, "=")
		if !ok {
			return nil, fmt.Errorf("release note must be formatted as language=text")
		}
		typedLanguage, err := play.NewListingLanguage(language)
		if err != nil {
			return nil, err
		}
		if text == "" {
			return nil, fmt.Errorf("release note text is required")
		}
		notes = append(notes, play.ReleaseNote{Language: typedLanguage, Text: text})
	}
	return notes, nil
}

func newReleasesPromoteCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		fromTrack    string
		toTrack      string
		versionCode  int64
		status       string
		userFraction float64
		confirm      bool
		dryRun       bool
	)

	cmd := &cobra.Command{
		Use:   "promote",
		Short: "Promote a release from one track to another",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
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
	cmd.Flags().StringVar(&fromTrack, "from", play.TrackInternal.String(), "Source track name")
	cmd.Flags().StringVar(&toTrack, "to", play.TrackProduction.String(), "Target track name")
	cmd.Flags().Int64Var(&versionCode, "version-code", 0, "Version code to promote")
	cmd.Flags().StringVar(&status, "status", play.ReleaseStatusDraft.String(), "Target release status: completed, draft, halted, inProgress")
	cmd.Flags().Float64Var(&userFraction, "user-fraction", 0, "Staged rollout fraction for inProgress or halted releases")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Commit the edit after validation")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned promotion workflow without calling Google Play")
	return cmd
}

func newReleasesHaltCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		trackName   string
		versionCode int64
		confirm     bool
		dryRun      bool
	)

	cmd := newRolloutCommand(out, options, play.RolloutActionHalt, packageName, &trackName, &versionCode, nil, &confirm, &dryRun)
	cmd.Flags().StringVar(&trackName, "track", play.TrackProduction.String(), "Track name")
	cmd.Flags().Int64Var(&versionCode, "version-code", 0, "Version code to halt")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Commit the edit after validation")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned halt workflow without calling Google Play")
	return cmd
}

func newReleasesResumeCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		trackName    string
		versionCode  int64
		status       string
		userFraction float64
		confirm      bool
		dryRun       bool
	)

	cmd := newRolloutCommand(out, options, play.RolloutActionResume, packageName, &trackName, &versionCode, &userFraction, &confirm, &dryRun)
	cmd.Flags().StringVar(&trackName, "track", play.TrackProduction.String(), "Track name")
	cmd.Flags().Int64Var(&versionCode, "version-code", 0, "Version code to resume")
	cmd.Flags().StringVar(&status, "status", "", "Target release status: completed or inProgress")
	cmd.Flags().Float64Var(&userFraction, "user-fraction", 0, "Staged rollout fraction when status is inProgress")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Commit the edit after validation")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned resume workflow without calling Google Play")
	cmd.RunE = rolloutRunE(out, options, play.RolloutActionResume, packageName, &trackName, &versionCode, &status, &userFraction, &confirm, &dryRun)
	return cmd
}

func newRolloutCommand(out io.Writer, options *globalOptions, action play.RolloutAction, packageName *string, trackName *string, versionCode *int64, userFraction *float64, confirm *bool, dryRun *bool) *cobra.Command {
	status := ""
	return &cobra.Command{
		Use:   action.String(),
		Short: fmt.Sprintf("%s a staged release", action.Title()),
		Args:  cobra.NoArgs,
		RunE:  rolloutRunE(out, options, action, packageName, trackName, versionCode, &status, userFraction, confirm, dryRun),
	}
}

func rolloutRunE(out io.Writer, options *globalOptions, action play.RolloutAction, packageName *string, trackName *string, versionCode *int64, status *string, userFraction *float64, confirm *bool, dryRun *bool) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
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
		if status != nil && *status != "" {
			typedStatus, err := play.NewReleaseStatus(*status)
			if err != nil {
				return err
			}
			rolloutOptions.Status = typedStatus
		}
		if userFraction != nil && cmd.Flags().Changed("user-fraction") {
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
	}
}
