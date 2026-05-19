package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/snitch"
	"github.com/spf13/cobra"
)

func newSnitchCommand(out io.Writer, options *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snitch",
		Short: "Report gpc friction",
	}
	cmd.AddCommand(newSnitchReportCommand(out, options))
	return cmd
}

func newSnitchReportCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var reportOptions snitch.ReportOptions
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate a GitHub issue URL for CLI friction",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			report, err := snitch.BuildReport(reportOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, report)
		},
	}
	cmd.Flags().StringVar(&reportOptions.Repository, "repo", snitch.DefaultRepository, "GitHub repository as owner/name")
	cmd.Flags().StringVar(&reportOptions.Title, "title", "", "Short issue title")
	cmd.Flags().StringVar(&reportOptions.Body, "body", "", "Issue body")
	cmd.Flags().StringVar(&reportOptions.Command, "command", "", "gpc command or workflow that caused friction")
	cmd.Flags().StringArrayVar(&reportOptions.Labels, "label", nil, "GitHub issue label; repeatable")
	return cmd
}
