package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/insights"
	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/spf13/cobra"
)

func newInsightsCommand(out io.Writer, options *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "insights",
		Short: "Summarize Google Play data exports",
	}
	cmd.AddCommand(newInsightsAnomaliesCommand(out, options))
	return cmd
}

func newInsightsAnomaliesCommand(out io.Writer, options *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "anomalies",
		Short: "Summarize Android vitals anomalies",
	}
	cmd.AddCommand(newInsightsAnomaliesSummarizeCommand(out, options))
	return cmd
}

func newInsightsAnomaliesSummarizeCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var summaryOptions insights.AnomalySummaryOptions
	cmd := &cobra.Command{
		Use:   "summarize",
		Short: "Summarize a vitals anomalies JSON export",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			summary, err := insights.SummarizeAnomalies(summaryOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, summary)
		},
	}
	cmd.Flags().StringVar(&summaryOptions.File, "file", "", "JSON output from gpc vitals anomalies list")
	return cmd
}
