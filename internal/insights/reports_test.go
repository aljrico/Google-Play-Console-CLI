package insights

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSummarizeReportsCombinesFinanceAndStatsReports(t *testing.T) {
	dir := t.TempDir()
	financeFile := filepath.Join(dir, "earnings.csv")
	statsFile := filepath.Join(dir, "stats.csv")
	if err := os.WriteFile(financeFile, []byte("Transaction Type,Merchant Currency,Amount (Merchant Currency)\nCharge,USD,9.99\nCharge,USD,1.01\nGoogle fee,USD,-1.50\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.WriteFile(statsFile, []byte("Date,Package name,Country,Store listing visitors,Store listing acquisitions,Conversion rate\n2026-05-01,com.example.app,US,10,2,20\n2026-05-02,com.example.app,US,20,3,22\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	summary, err := SummarizeReports(ReportInsightsOptions{
		FinanceFiles: []string{financeFile},
		StatsFiles:   []string{statsFile},
	})
	if err != nil {
		t.Fatalf("SummarizeReports() error = %v", err)
	}
	if len(summary.FinanceReports) != 1 || summary.FinanceReports[0].ReportType != "earnings" {
		t.Fatalf("FinanceReports = %#v, want earnings summary", summary.FinanceReports)
	}
	if len(summary.StatsReports) != 1 || summary.StatsReports[0].PackageName != "com.example.app" {
		t.Fatalf("StatsReports = %#v, want app stats summary", summary.StatsReports)
	}
	if len(summary.Highlights) != 2 {
		t.Fatalf("Highlights = %#v, want two highlights", summary.Highlights)
	}
	if !strings.Contains(summary.Highlights[0].Summary, "top transaction by count is Charge") {
		t.Fatalf("finance highlight = %#v, want top charge transaction", summary.Highlights[0])
	}
	if !strings.Contains(summary.Highlights[1].Summary, "Store listing acquisitions") {
		t.Fatalf("stats highlight = %#v, want first sum metric", summary.Highlights[1])
	}
	for _, want := range []ReportInsightKPI{
		{Name: "netRevenue", Value: "9.5", Currency: "USD", ReportType: "earnings"},
		{Name: "storeListingAcquisitions", Value: "5"},
		{Name: "storeListingAcquisitionRate", Value: "0.166667"},
		{Name: "netRevenuePerStoreListingAcquisition", Value: "1.9", Currency: "USD", ReportType: "earnings"},
	} {
		if !hasKPI(summary.KPIs, want) {
			t.Fatalf("KPIs = %#v, want %#v", summary.KPIs, want)
		}
	}
}

func TestSummarizeReportsKeepsFinanceKPIsSeparateByReportType(t *testing.T) {
	dir := t.TempDir()
	earningsFile := filepath.Join(dir, "earnings.csv")
	estimatedSalesFile := filepath.Join(dir, "estimated-sales.csv")
	statsFile := filepath.Join(dir, "stats.csv")
	if err := os.WriteFile(earningsFile, []byte("Transaction Type,Merchant Currency,Amount (Merchant Currency)\nCharge,USD,9.99\nGoogle fee,USD,-1.99\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.WriteFile(estimatedSalesFile, []byte("Financial Status,Currency of Sale,Charged Amount\nCharged,USD,12.00\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.WriteFile(statsFile, []byte("Date,Package name,Store listing acquisitions\n2026-05-01,com.example.app,4\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	summary, err := SummarizeReports(ReportInsightsOptions{
		FinanceFiles: []string{earningsFile, estimatedSalesFile},
		StatsFiles:   []string{statsFile},
	})
	if err != nil {
		t.Fatalf("SummarizeReports() error = %v", err)
	}
	for _, want := range []ReportInsightKPI{
		{Name: "netRevenue", Value: "8", Currency: "USD", ReportType: "earnings"},
		{Name: "netRevenue", Value: "12", Currency: "USD", ReportType: "estimated-sales"},
		{Name: "netRevenuePerStoreListingAcquisition", Value: "2", Currency: "USD", ReportType: "earnings"},
		{Name: "netRevenuePerStoreListingAcquisition", Value: "3", Currency: "USD", ReportType: "estimated-sales"},
	} {
		if !hasKPI(summary.KPIs, want) {
			t.Fatalf("KPIs = %#v, want %#v", summary.KPIs, want)
		}
	}
}

func TestSummarizeReportsRequiresAtLeastOneFile(t *testing.T) {
	_, err := SummarizeReports(ReportInsightsOptions{})
	if err == nil {
		t.Fatal("SummarizeReports() error = nil, want file validation")
	}
}

func TestSummarizeReportsRejectsWhitespaceFile(t *testing.T) {
	_, err := SummarizeReports(ReportInsightsOptions{FinanceFiles: []string{" earnings.csv"}})
	if err == nil {
		t.Fatal("SummarizeReports() error = nil, want whitespace validation")
	}
}

func TestSummarizeReportsKeepsStableJSONArraysWhenOnlyFinanceProvided(t *testing.T) {
	file := filepath.Join(t.TempDir(), "earnings.csv")
	if err := os.WriteFile(file, []byte("Transaction Type,Merchant Currency,Amount (Merchant Currency)\nCharge,USD,9.99\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	summary, err := SummarizeReports(ReportInsightsOptions{FinanceFiles: []string{file}})
	if err != nil {
		t.Fatalf("SummarizeReports() error = %v", err)
	}
	content, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	for _, want := range []string{`"financeReports":[`, `"statsReports":[]`, `"kpis":[`, `1 row`} {
		if !strings.Contains(string(content), want) {
			t.Fatalf("JSON = %s, want %s", content, want)
		}
	}
}

func TestSummarizeReportsKeepsStableJSONArraysWhenOnlyStatsProvided(t *testing.T) {
	file := filepath.Join(t.TempDir(), "stats.csv")
	if err := os.WriteFile(file, []byte("Date,Package name,Store listing visitors\n2026-05-01,com.example.app,10\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	summary, err := SummarizeReports(ReportInsightsOptions{StatsFiles: []string{file}})
	if err != nil {
		t.Fatalf("SummarizeReports() error = %v", err)
	}
	content, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	for _, want := range []string{`"financeReports":[]`, `"statsReports":[`, `"kpis":[]`, `1 row`} {
		if !strings.Contains(string(content), want) {
			t.Fatalf("JSON = %s, want %s", content, want)
		}
	}
}

func hasKPI(kpis []ReportInsightKPI, want ReportInsightKPI) bool {
	for _, kpi := range kpis {
		if kpi.Name == want.Name && kpi.Value == want.Value && kpi.Currency == want.Currency && kpi.ReportType == want.ReportType {
			return true
		}
	}
	return false
}
