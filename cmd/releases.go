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
		packageName string
		trackName   string
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
	cmd.AddCommand(&cobra.Command{
		Use:   "upload",
		Short: "Upload an Android App Bundle to a track",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if packageName == "" {
				return fmt.Errorf("--package is required")
			}
			return fmt.Errorf("releases upload is not implemented yet")
		},
	})

	return cmd
}
