package insights

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/aljrico/Google-Play-Console-CLI/internal/reporting"
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

func TestSummarizeAnomaliesFlagsPartialPaginatedInput(t *testing.T) {
	file := filepath.Join(t.TempDir(), "anomalies.json")
	content := `{
		"packageName": "com.example.app",
		"nextPageToken": "page-2",
		"anomalies": []
	}`
	if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	summary, err := SummarizeAnomalies(AnomalySummaryOptions{File: file})
	if err != nil {
		t.Fatalf("SummarizeAnomalies() error = %v", err)
	}
	if !summary.Partial || summary.NextPageToken != "page-2" {
		t.Fatalf("summary = %#v, want partial page token", summary)
	}
}

func TestSummarizeAnomaliesRejectsWrongJSONShape(t *testing.T) {
	file := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(file, []byte(`{}`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	_, err := SummarizeAnomalies(AnomalySummaryOptions{File: file})
	if err == nil {
		t.Fatal("SummarizeAnomalies() error = nil, want shape validation")
	}
}

func TestSummarizeAnomaliesAcceptsTypedAnomalyListResultJSON(t *testing.T) {
	file := filepath.Join(t.TempDir(), "anomalies.json")
	content, err := json.Marshal(reporting.AnomalyListResult{
		Options: reporting.AnomalyListOptions{PackageName: play.PackageName("com.example.app")},
		Anomalies: []reporting.Anomaly{
			{Name: "a1", MetricSet: "z"},
			{Name: "a2", MetricSet: "a"},
		},
	})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if err := os.WriteFile(file, content, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	summary, err := SummarizeAnomalies(AnomalySummaryOptions{File: file})
	if err != nil {
		t.Fatalf("SummarizeAnomalies() error = %v", err)
	}
	if summary.PackageName != "com.example.app" {
		t.Fatalf("PackageName = %q, want options package", summary.PackageName)
	}
	if summary.MetricSets[0].Name != "a" || summary.MetricSets[1].Name != "z" {
		t.Fatalf("MetricSets = %#v, want deterministic tie order", summary.MetricSets)
	}
}

func TestSummarizeAnomaliesValidatesFile(t *testing.T) {
	_, err := SummarizeAnomalies(AnomalySummaryOptions{})
	if err == nil {
		t.Fatal("SummarizeAnomalies() error = nil, want file validation")
	}
}
