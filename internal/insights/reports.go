package insights

import (
	"fmt"
	"strings"

	"github.com/aljrico/Google-Play-Console-CLI/internal/analytics"
	"github.com/aljrico/Google-Play-Console-CLI/internal/finance"
)

type ReportInsightsOptions struct {
	FinanceFiles []string `json:"financeFiles,omitempty"`
	StatsFiles   []string `json:"statsFiles,omitempty"`
}

type ReportInsights struct {
	FinanceReports []finance.ReportSummary  `json:"financeReports,omitempty"`
	StatsReports   []analytics.StatsSummary `json:"statsReports,omitempty"`
	Highlights     []ReportInsightHighlight `json:"highlights,omitempty"`
}

type ReportInsightHighlight struct {
	Kind    string `json:"kind"`
	File    string `json:"file"`
	Summary string `json:"summary"`
}

func SummarizeReports(options ReportInsightsOptions) (ReportInsights, error) {
	if err := options.Validate(); err != nil {
		return ReportInsights{}, err
	}
	result := ReportInsights{
		FinanceReports: []finance.ReportSummary{},
		StatsReports:   []analytics.StatsSummary{},
		Highlights:     []ReportInsightHighlight{},
	}
	for _, file := range options.FinanceFiles {
		summary, err := finance.SummarizeReport(finance.ReportSummaryOptions{File: file})
		if err != nil {
			return ReportInsights{}, err
		}
		result.FinanceReports = append(result.FinanceReports, summary)
		result.Highlights = append(result.Highlights, financeReportHighlight(summary))
	}
	for _, file := range options.StatsFiles {
		summary, err := analytics.SummarizeStats(analytics.StatsSummaryOptions{File: file})
		if err != nil {
			return ReportInsights{}, err
		}
		result.StatsReports = append(result.StatsReports, summary)
		result.Highlights = append(result.Highlights, statsReportHighlight(summary))
	}
	return result, nil
}

func (o ReportInsightsOptions) Validate() error {
	if len(o.FinanceFiles) == 0 && len(o.StatsFiles) == 0 {
		return fmt.Errorf("at least one finance or stats report file is required")
	}
	for _, file := range o.FinanceFiles {
		if strings.TrimSpace(file) == "" {
			return fmt.Errorf("finance report file is required")
		}
		if strings.TrimSpace(file) != file {
			return fmt.Errorf("finance report file cannot have leading or trailing whitespace")
		}
	}
	for _, file := range o.StatsFiles {
		if strings.TrimSpace(file) == "" {
			return fmt.Errorf("stats report file is required")
		}
		if strings.TrimSpace(file) != file {
			return fmt.Errorf("stats report file cannot have leading or trailing whitespace")
		}
	}
	return nil
}

func financeReportHighlight(summary finance.ReportSummary) ReportInsightHighlight {
	top := topTransactionTotal(summary.TransactionTypes)
	message := fmt.Sprintf("%s report has %d rows", summary.ReportType, summary.Rows)
	if top.TransactionType != "" {
		message = fmt.Sprintf("%s; top transaction by count is %s (%d rows, total %s", message, top.TransactionType, top.Count, top.Total)
		if top.Currency != "" {
			message += " " + top.Currency
		}
		message += ")"
	}
	return ReportInsightHighlight{
		Kind:    "finance",
		File:    summary.File,
		Summary: message,
	}
}

func topTransactionTotal(totals []finance.TransactionTotal) finance.TransactionTotal {
	var top finance.TransactionTotal
	for _, total := range totals {
		if top.TransactionType == "" || total.Count > top.Count || (total.Count == top.Count && total.TransactionType < top.TransactionType) {
			top = total
		}
	}
	return top
}

func statsReportHighlight(summary analytics.StatsSummary) ReportInsightHighlight {
	top := topMetricSummary(summary.Metrics)
	message := fmt.Sprintf("stats report has %d rows", summary.Rows)
	if summary.PackageName != "" {
		message = fmt.Sprintf("%s for %s", message, summary.PackageName)
	}
	if summary.StartDate != "" && summary.EndDate != "" {
		message = fmt.Sprintf("%s from %s to %s", message, summary.StartDate, summary.EndDate)
	}
	if top.Name != "" {
		message = fmt.Sprintf("%s; first headline metric is %s (%s %s)", message, top.Name, top.Aggregation, top.Value)
	}
	return ReportInsightHighlight{
		Kind:    "stats",
		File:    summary.File,
		Summary: message,
	}
}

func topMetricSummary(metrics []analytics.MetricSummary) analytics.MetricSummary {
	for _, metric := range metrics {
		if metric.Aggregation == "sum" {
			return metric
		}
	}
	if len(metrics) == 0 {
		return analytics.MetricSummary{}
	}
	return metrics[0]
}
