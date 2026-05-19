package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/spf13/cobra"
)

func newTestersCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var packageName string

	cmd := &cobra.Command{
		Use:   "testers",
		Short: "Manage Google Play track tester groups",
	}
	cmd.PersistentFlags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.AddCommand(
		newTestersGetCommand(out, options, &packageName),
		newTestersUpdateCommand(out, options, &packageName),
	)
	return cmd
}

func newTestersGetCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var trackName string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get Google Groups configured as testers for a track",
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
			testers, err := play.GetTesters(cmd.Context(), publisher, play.TestersGetOptions{
				PackageName: typedPackageName,
				Track:       typedTrackName,
			})
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, testers)
		},
	}
	cmd.Flags().StringVar(&trackName, "track", "internal", "Track name, for example internal, alpha, beta, or production")
	return cmd
}

func newTestersUpdateCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		trackName    string
		googleGroups []string
		clear        bool
		confirm      bool
		dryRun       bool
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Replace Google Groups configured as testers for a track",
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
			typedGoogleGroups, err := parseTesterGoogleGroups(googleGroups)
			if err != nil {
				return err
			}
			updateOptions := play.TestersUpdateOptions{
				PackageName:  typedPackageName,
				Track:        typedTrackName,
				GoogleGroups: typedGoogleGroups,
				Clear:        clear,
				Confirm:      confirm,
				DryRun:       dryRun,
			}
			if dryRun {
				result, err := play.UpdateTesters(cmd.Context(), nil, updateOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			if err := updateOptions.Validate(); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.UpdateTesters(cmd.Context(), publisher, updateOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&trackName, "track", "internal", "Track name, for example internal, alpha, beta, or production")
	cmd.Flags().StringArrayVar(&googleGroups, "google-group", nil, "Testing Google Group email address, repeatable")
	cmd.Flags().BoolVar(&clear, "clear", false, "Remove all testing Google Groups from the track")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Commit the tester update")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned tester update without calling Google Play")
	return cmd
}

func parseTesterGoogleGroups(values []string) ([]play.TesterGoogleGroup, error) {
	googleGroups := make([]play.TesterGoogleGroup, 0, len(values))
	for _, value := range values {
		googleGroup, err := play.NewTesterGoogleGroup(value)
		if err != nil {
			return nil, err
		}
		googleGroups = append(googleGroups, googleGroup)
	}
	return googleGroups, nil
}
