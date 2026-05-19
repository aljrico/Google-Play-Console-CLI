package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/agentskills"
	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/spf13/cobra"
)

func newInstallSkillsCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var (
		directory string
		skills    []string
		force     bool
		dryRun    bool
	)

	cmd := &cobra.Command{
		Use:   "install-skills",
		Short: "Install bundled gpc agent skills",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := agentskills.Install(cmd.Context(), agentskills.InstallOptions{
				Directory: directory,
				Skills:    skills,
				Force:     force,
				DryRun:    dryRun,
			})
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&directory, "directory", "", "Directory for installed agent skills, defaults to ~/.agents/skills")
	cmd.Flags().StringArrayVar(&skills, "skill", nil, "Install one bundled skill by name; repeat to install multiple")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing skill files")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print planned skill installs without writing")
	cmd.AddCommand(newInstallSkillsListCommand(out, options))
	return cmd
}

func newInstallSkillsListCommand(out io.Writer, options *globalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List bundled gpc agent skills",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return output.Write(out, options.output, options.pretty, agentskills.List())
		},
	}
}
