package analytics

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSummarizeStatsTotalsMetricColumns(t *testing.T) {
	file := filepath.Join(t.TempDir(), "store_performance.csv")
	content := "Date,Package name,Country,Store listing visitors,Store listing acquisitions,Conversion rate\n2026-05-01,com.example.app,US,10,2,20.5\n2026-05-02,com.example.app,US,20,3,21.5\n"
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
	wantMetrics := []MetricSummary{
		{Name: "Conversion rate", Aggregation: "average", Value: "21"},
		{Name: "Store listing acquisitions", Aggregation: "sum", Value: "5"},
		{Name: "Store listing visitors", Aggregation: "sum", Value: "30"},
	}
	if len(summary.Metrics) != len(wantMetrics) {
		t.Fatalf("metrics = %#v, want %#v", summary.Metrics, wantMetrics)
	}
	for index, want := range wantMetrics {
		if summary.Metrics[index] != want {
			t.Fatalf("metrics[%d] = %#v, want %#v", index, summary.Metrics[index], want)
		}
	}
}

func TestSummarizeStatsIgnoresCommonDimensions(t *testing.T) {
	file := filepath.Join(t.TempDir(), "installs_by_device.csv")
	content := "Date,Package name,Device,App Version Code,Android OS Version,Language,Carrier,Install events\n2026-05-01,com.example.app,oriole,42,15,en,Example Wireless,100\n"
	if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	summary, err := SummarizeStats(StatsSummaryOptions{File: file})
	if err != nil {
		t.Fatalf("SummarizeStats() error = %v", err)
	}
	if len(summary.Metrics) != 1 || summary.Metrics[0].Name != "Install events" || summary.Metrics[0].Value != "100" {
		t.Fatalf("metrics = %#v, want only install metric", summary.Metrics)
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

func TestSummarizeStatsRejectsMalformedDecimal(t *testing.T) {
	file := filepath.Join(t.TempDir(), "bad.csv")
	if err := os.WriteFile(file, []byte("Date,Package name,Installs\n2026-05-01,com.example.app,.\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	_, err := SummarizeStats(StatsSummaryOptions{File: file})
	if err == nil {
		t.Fatal("SummarizeStats() error = nil, want metric validation")
	}
}
