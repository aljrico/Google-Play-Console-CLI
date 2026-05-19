package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/workflow"
	"github.com/spf13/cobra"
)

func newWorkflowCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "workflow",
		Short: "Run repo-local playpub workflows",
	}
	cmd.PersistentFlags().StringVar(&file, "file", workflow.DefaultFile, "Workflow JSON file")
	cmd.AddCommand(
		newWorkflowListCommand(out, options, &file),
		newWorkflowRunCommand(out, options, &file),
	)
	return cmd
}

func newWorkflowListCommand(out io.Writer, options *globalOptions, file *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configured workflows",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			workflows, err := workflow.List(workflow.ListOptions{File: *file})
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, workflows)
		},
	}
	return cmd
}

func newWorkflowRunCommand(out io.Writer, options *globalOptions, file *string) *cobra.Command {
	var (
		workDir string
		dryRun  bool
		confirm bool
	)

	cmd := &cobra.Command{
		Use:   "run NAME",
		Short: "Run one configured workflow",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := workflow.Run(cmd.Context(), nil, workflow.RunOptions{
				File:    *file,
				Name:    args[0],
				WorkDir: workDir,
				DryRun:  dryRun,
				Confirm: confirm,
			})
			if result.Name == "" && len(result.Steps) == 0 {
				return err
			}
			if writeErr := output.Write(out, options.output, options.pretty, result); writeErr != nil {
				return writeErr
			}
			return err
		},
	}
	cmd.Flags().StringVar(&workDir, "workdir", "", "Working directory for shell steps; defaults to the workflow root")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned workflow steps without executing them")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Execute the workflow shell steps")
	return cmd
}
