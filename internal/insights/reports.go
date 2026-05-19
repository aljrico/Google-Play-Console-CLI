package insights

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/aljrico/Google-Play-Console-CLI/internal/analytics"
	"github.com/aljrico/Google-Play-Console-CLI/internal/decimal"
	"github.com/aljrico/Google-Play-Console-CLI/internal/finance"
)

type ReportInsightsOptions struct {
	FinanceFiles []string `json:"financeFiles,omitempty"`
	StatsFiles   []string `json:"statsFiles,omitempty"`
}

type ReportInsights struct {
	FinanceReports []finance.ReportSummary  `json:"financeReports"`
	StatsReports   []analytics.StatsSummary `json:"statsReports"`
	Highlights     []ReportInsightHighlight `json:"highlights,omitempty"`
	KPIs           []ReportInsightKPI       `json:"kpis"`
}

type ReportInsightHighlight struct {
	Kind    string `json:"kind"`
	File    string `json:"file"`
	Summary string `json:"summary"`
}

type ReportInsightKPI struct {
	Name       string `json:"name"`
	Value      string `json:"value"`
	Currency   string `json:"currency,omitempty"`
	ReportType string `json:"reportType,omitempty"`
	Summary    string `json:"summary"`
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
	result.KPIs = reportInsightKPIs(result.FinanceReports, result.StatsReports)
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
	message := fmt.Sprintf("%s report has %s", summary.ReportType, formatRowCount(summary.Rows))
	if top.TransactionType != "" {
		message = fmt.Sprintf("%s; top transaction by count is %s (%s, total %s", message, top.TransactionType, formatRowCount(top.Count), top.Total)
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
	message := fmt.Sprintf("stats report has %s", formatRowCount(summary.Rows))
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

func reportInsightKPIs(financeReports []finance.ReportSummary, statsReports []analytics.StatsSummary) []ReportInsightKPI {
	kpis := []ReportInsightKPI{}
	netRevenueByBasis := financeTotalsByBasis(financeReports)
	for _, basis := range sortedFinanceBasisKeys(netRevenueByBasis) {
		total := netRevenueByBasis[basis]
		kpis = append(kpis, ReportInsightKPI{
			Name:       "netRevenue",
			Value:      total.String(),
			Currency:   basis.Currency,
			ReportType: basis.ReportType,
			Summary:    fmt.Sprintf("%s net finance total is %s %s", basis.ReportType, total.String(), basis.Currency),
		})
	}
	acquisitions := sumStatsMetric(statsReports, "Store listing acquisitions")
	if acquisitions != "" {
		kpis = append(kpis, ReportInsightKPI{
			Name:    "storeListingAcquisitions",
			Value:   acquisitions,
			Summary: fmt.Sprintf("store listing acquisitions total %s", acquisitions),
		})
	}
	visitors := sumStatsMetric(statsReports, "Store listing visitors")
	if acquisitions != "" && visitors != "" {
		if visitorCount, ok := positiveInt(visitors); ok {
			acquisitionAmount, _ := decimal.Parse(acquisitions)
			rate := acquisitionAmount.Average(visitorCount)
			kpis = append(kpis, ReportInsightKPI{
				Name:    "storeListingAcquisitionRate",
				Value:   rate,
				Summary: fmt.Sprintf("store listing acquisitions per visitor %s", rate),
			})
		}
	}
	if acquisitions != "" {
		if acquisitionCount, ok := positiveInt(acquisitions); ok {
			for _, basis := range sortedFinanceBasisKeys(netRevenueByBasis) {
				total := netRevenueByBasis[basis]
				value := total.Average(acquisitionCount)
				kpis = append(kpis, ReportInsightKPI{
					Name:       "netRevenuePerStoreListingAcquisition",
					Value:      value,
					Currency:   basis.Currency,
					ReportType: basis.ReportType,
					Summary:    fmt.Sprintf("%s net revenue per store listing acquisition is %s %s", basis.ReportType, value, basis.Currency),
				})
			}
		}
	}
	return kpis
}

type financeKPIBasis struct {
	ReportType string
	Currency   string
}

func financeTotalsByBasis(reports []finance.ReportSummary) map[financeKPIBasis]decimal.Amount {
	totals := map[financeKPIBasis]decimal.Amount{}
	for _, report := range reports {
		for _, transactionType := range report.TransactionTypes {
			if transactionType.Currency == "" {
				continue
			}
			total, err := decimal.Parse(transactionType.Total)
			if err != nil {
				continue
			}
			basis := financeKPIBasis{ReportType: report.ReportType, Currency: transactionType.Currency}
			totals[basis] = totals[basis].Add(total)
		}
	}
	return totals
}

func sumStatsMetric(reports []analytics.StatsSummary, metricName string) string {
	var total decimal.Amount
	found := false
	for _, report := range reports {
		for _, metric := range report.Metrics {
			if metric.Name != metricName || metric.Aggregation != "sum" {
				continue
			}
			value, err := decimal.Parse(metric.Value)
			if err != nil {
				continue
			}
			total = total.Add(value)
			found = true
		}
	}
	if !found {
		return ""
	}
	return total.String()
}

func positiveInt(value string) (int, bool) {
	if strings.Contains(value, ".") {
		return 0, false
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, false
	}
	return parsed, true
}

func sortedFinanceBasisKeys(values map[financeKPIBasis]decimal.Amount) []financeKPIBasis {
	keys := make([]financeKPIBasis, 0, len(values))
	for basis := range values {
		keys = append(keys, basis)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].ReportType == keys[j].ReportType {
			return keys[i].Currency < keys[j].Currency
		}
		return keys[i].ReportType < keys[j].ReportType
	})
	return keys
}

func formatRowCount(rows int) string {
	if rows == 1 {
		return "1 row"
	}
	return fmt.Sprintf("%d rows", rows)
}
