package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/finance"
	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/spf13/cobra"
)

func newFinanceCommand(out io.Writer, options *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "finance",
		Short: "Summarize Google Play financial reports",
	}
	cmd.AddCommand(newFinanceReportsCommand(out, options))
	return cmd
}

func newFinanceReportsCommand(out io.Writer, options *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reports",
		Short: "Summarize downloaded Play financial report CSVs",
	}
	cmd.AddCommand(newFinanceReportsSummarizeCommand(out, options))
	return cmd
}

func newFinanceReportsSummarizeCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var summaryOptions finance.ReportSummaryOptions
	cmd := &cobra.Command{
		Use:   "summarize",
		Short: "Summarize a Play financial report CSV",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			summary, err := finance.SummarizeReport(summaryOptions)
			if err != nil {
				return err
			}
			if options.output == output.Table || options.output == output.Markdown {
				return output.Write(out, options.output, options.pretty, summary.TransactionTypes)
			}
			return output.Write(out, options.output, options.pretty, summary)
		},
	}
	cmd.Flags().StringVar(&summaryOptions.File, "file", "", "Downloaded Google Play earnings or estimated-sales CSV")
	_ = cmd.MarkFlagRequired("file")
	_ = cmd.MarkFlagFilename("file", "csv")
	return cmd
}
