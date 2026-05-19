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
	cmd.AddCommand(newInsightsAnomaliesCommand(out, options), newInsightsReportsCommand(out, options))
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

func newInsightsReportsCommand(out io.Writer, options *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reports",
		Short: "Summarize finance and analytics report insights",
	}
	cmd.AddCommand(newInsightsReportsSummarizeCommand(out, options))
	return cmd
}

func newInsightsReportsSummarizeCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var summaryOptions insights.ReportInsightsOptions
	cmd := &cobra.Command{
		Use:   "summarize",
		Short: "Summarize Play finance and statistics report insights",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			summary, err := insights.SummarizeReports(summaryOptions)
			if err != nil {
				return err
			}
			if options.output == output.Table || options.output == output.Markdown {
				return output.Write(out, options.output, options.pretty, summary.Highlights)
			}
			return output.Write(out, options.output, options.pretty, summary)
		},
	}
	cmd.Flags().StringArrayVar(&summaryOptions.FinanceFiles, "finance-file", nil, "Downloaded Google Play earnings or estimated-sales CSV; repeatable")
	cmd.Flags().StringArrayVar(&summaryOptions.StatsFiles, "stats-file", nil, "Downloaded Google Play statistics CSV; repeatable")
	_ = cmd.MarkFlagFilename("finance-file", "csv")
	_ = cmd.MarkFlagFilename("stats-file", "csv")
	return cmd
}
