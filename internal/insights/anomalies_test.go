package insights

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSummarizeAnomaliesCountsMetricSetsAndMetrics(t *testing.T) {
	file := filepath.Join(t.TempDir(), "anomalies.json")
	content := `{
		"packageName": "com.example.app",
		"anomalies": [
			{"name":"a1","metricSet":"apps/com.example.app/crashRateMetricSet","metric":{"metric":"crashRate"}},
			{"name":"a2","metricSet":"apps/com.example.app/crashRateMetricSet","metric":{"metric":"crashRate"}},
			{"name":"a3","metricSet":"apps/com.example.app/errorCountMetricSet","metric":{"metric":"errorCount"}}
		]
	}`
	if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	summary, err := SummarizeAnomalies(AnomalySummaryOptions{File: file})
	if err != nil {
		t.Fatalf("SummarizeAnomalies() error = %v", err)
	}
	if summary.Total != 3 {
		t.Fatalf("Total = %d, want 3", summary.Total)
	}
	if summary.PackageName != "com.example.app" {
		t.Fatalf("PackageName = %q", summary.PackageName)
	}
	if summary.MetricSets[0].Count != 2 || summary.Metrics[0].Name != "crashRate" {
		t.Fatalf("summary = %#v, want crashRate first", summary)
	}
}

func TestSummarizeAnomaliesValidatesFile(t *testing.T) {
	_, err := SummarizeAnomalies(AnomalySummaryOptions{})
	if err == nil {
		t.Fatal("SummarizeAnomalies() error = nil, want file validation")
	}
}
