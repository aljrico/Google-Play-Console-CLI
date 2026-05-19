package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/project"
	"github.com/spf13/cobra"
)

func newInitCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var (
		directory string
		force     bool
		dryRun    bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create a local gpc workspace",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			plan, err := project.Init(cmd.Context(), project.InitOptions{
				Directory: directory,
				Force:     force,
				DryRun:    dryRun,
			})
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, plan)
		},
	}
	cmd.Flags().StringVar(&directory, "directory", ".gpc", "Directory for gpc helper files")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing gpc helper files")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned init files without writing")
	return cmd
}
