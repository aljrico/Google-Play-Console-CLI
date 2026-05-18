package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/spf13/cobra"
)

func newTracksCommand(out io.Writer) *cobra.Command {
	var packageName string

	cmd := &cobra.Command{
		Use:   "tracks",
		Short: "Manage Google Play release tracks",
	}

	cmd.PersistentFlags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List release tracks",
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(packageName)
			if err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			tracks, err := play.ListTracks(cmd.Context(), publisher, typedPackageName)
			if err != nil {
				return err
			}
			return output.Write(out, opts.output, opts.pretty, tracks)
		},
	})

	return cmd
}
