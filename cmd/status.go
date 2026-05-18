package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/spf13/cobra"
)

func newStatusCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var (
		packageName  string
		includeDraft bool
	)

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Summarize Google Play release status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(packageName)
			if err != nil {
				return err
			}
			statusOptions := play.ReleaseStatusOptions{
				PackageName:  typedPackageName,
				IncludeDraft: includeDraft,
			}
			if err := statusOptions.Validate(); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			status, err := play.GetReleaseStatus(cmd.Context(), publisher, statusOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, status)
		},
	}
	cmd.Flags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.Flags().BoolVar(&includeDraft, "include-draft", false, "Include draft releases in the status summary")
	return cmd
}
