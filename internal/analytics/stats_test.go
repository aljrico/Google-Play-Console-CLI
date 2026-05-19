package analytics

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSummarizeStatsTotalsMetricColumns(t *testing.T) {
	file := filepath.Join(t.TempDir(), "store_performance.csv")
	content := "Date,Package name,Country/region,Store listing visitors,Store listing acquisitions,Conversion rate\n2026-05-01,com.example.app,US,10,2,20.5\n2026-05-02,com.example.app,US,15,3,21.5\n"
	if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	summary, err := SummarizeStats(StatsSummaryOptions{File: file})
	if err != nil {
		t.Fatalf("SummarizeStats() error = %v", err)
	}
	if summary.Rows != 2 || summary.PackageName != "com.example.app" || summary.StartDate != "2026-05-01" || summary.EndDate != "2026-05-02" {
		t.Fatalf("summary metadata = %#v", summary)
	}
	if summary.Metrics[0].Name != "Conversion rate" || summary.Metrics[0].Total != "42" {
		t.Fatalf("metrics = %#v, want deterministic metric totals", summary.Metrics)
	}
}

func TestSummarizeStatsRejectsMissingRequiredHeaders(t *testing.T) {
	file := filepath.Join(t.TempDir(), "bad.csv")
	if err := os.WriteFile(file, []byte("Package name,Installs\ncom.example.app,1\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	_, err := SummarizeStats(StatsSummaryOptions{File: file})
	if err == nil {
		t.Fatal("SummarizeStats() error = nil, want header validation")
	}
}

func TestSummarizeStatsRejectsInvalidMetricValue(t *testing.T) {
	file := filepath.Join(t.TempDir(), "bad.csv")
	if err := os.WriteFile(file, []byte("Date,Package name,Installs\n2026-05-01,com.example.app,nope\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	_, err := SummarizeStats(StatsSummaryOptions{File: file})
	if err == nil {
		t.Fatal("SummarizeStats() error = nil, want metric validation")
	}
}
