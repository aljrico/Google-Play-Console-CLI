package insights

import (
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
