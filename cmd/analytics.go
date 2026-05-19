package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/analytics"
	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/spf13/cobra"
)

func newAnalyticsCommand(out io.Writer, options *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analytics",
		Short: "Summarize Google Play statistics reports",
	}
	cmd.AddCommand(newAnalyticsStatsCommand(out, options))
	return cmd
}

func newAnalyticsStatsCommand(out io.Writer, options *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Summarize downloaded Play statistics CSVs",
	}
	cmd.AddCommand(newAnalyticsStatsSummarizeCommand(out, options))
	return cmd
}

func newAnalyticsStatsSummarizeCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var summaryOptions analytics.StatsSummaryOptions
	cmd := &cobra.Command{
		Use:   "summarize",
		Short: "Summarize a Play statistics report CSV",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			summary, err := analytics.SummarizeStats(summaryOptions)
			if err != nil {
				return err
			}
			if options.output == output.Table || options.output == output.Markdown {
				return output.Write(out, options.output, options.pretty, summary.Metrics)
			}
			return output.Write(out, options.output, options.pretty, summary)
		},
	}
	cmd.Flags().StringVar(&summaryOptions.File, "file", "", "Downloaded Google Play statistics report CSV")
	_ = cmd.MarkFlagRequired("file")
	_ = cmd.MarkFlagFilename("file", "csv")
	return cmd
}
