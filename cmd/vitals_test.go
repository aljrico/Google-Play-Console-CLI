package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestVitalsMetricSetGetRejectsUnsupportedMetricSetBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"vitals",
		"metric-set",
		"get",
		"--package",
		"com.example.app",
		"--metric-set",
		"crashes",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected metric set validation error")
	}
	if !strings.Contains(err.Error(), "unsupported vitals metric set") {
		t.Fatalf("error = %v, want metric set validation", err)
	}
}

func TestVitalsMetricSetGetRejectsMissingMetricSetBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"vitals",
		"metric-set",
		"get",
		"--package",
		"com.example.app",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected metric set validation error")
	}
	if !strings.Contains(err.Error(), "vitals metric set is required") {
		t.Fatalf("error = %v, want required metric set validation", err)
	}
}

func TestVitalsMetricSetQueryRejectsMissingMetricBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"vitals",
		"metric-set",
		"query",
		"--package",
		"com.example.app",
		"--metric-set",
		"crash-rate",
		"--aggregation",
		"DAILY",
		"--start-date",
		"2026-05-01",
		"--end-date",
		"2026-05-19",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected metric validation error")
	}
	if !strings.Contains(err.Error(), "at least one metric is required") {
		t.Fatalf("error = %v, want metric validation", err)
	}
}

func TestVitalsMetricSetQueryRejectsInvalidStartDateBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"vitals",
		"metric-set",
		"query",
		"--package",
		"com.example.app",
		"--metric-set",
		"crash-rate",
		"--metric",
		"crashRate",
		"--aggregation",
		"DAILY",
		"--start-date",
		"2026/05/01",
		"--end-date",
		"2026-05-19",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected start date validation error")
	}
	if !strings.Contains(err.Error(), "must use YYYY-MM-DD") {
		t.Fatalf("error = %v, want start date validation", err)
	}
}

func TestVitalsMetricSetQueryRejectsInvalidRangeBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"vitals",
		"metric-set",
		"query",
		"--package",
		"com.example.app",
		"--metric-set",
		"crash-rate",
		"--metric",
		"crashRate",
		"--aggregation",
		"DAILY",
		"--start-date",
		"2026-05-19",
		"--end-date",
		"2026-05-01",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected date range validation error")
	}
	if !strings.Contains(err.Error(), "start date must be before end date") {
		t.Fatalf("error = %v, want date range validation", err)
	}
}

func TestVitalsMetricSetQueryRejectsUnsupportedAggregationBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"vitals",
		"metric-set",
		"query",
		"--package",
		"com.example.app",
		"--metric-set",
		"error-count",
		"--metric",
		"errorReportCount",
		"--aggregation",
		"HOURLY",
		"--start-date",
		"2026-05-01",
		"--end-date",
		"2026-05-19",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected aggregation validation error")
	}
	if !strings.Contains(err.Error(), "aggregation period HOURLY is not supported") {
		t.Fatalf("error = %v, want aggregation validation", err)
	}
}

func TestVitalsErrorsIssuesSearchRejectsInvalidRangeBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"vitals",
		"errors",
		"issues",
		"search",
		"--package",
		"com.example.app",
		"--start-date",
		"2026-05-19",
		"--end-date",
		"2026-05-01",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected date range validation error")
	}
	if !strings.Contains(err.Error(), "start date must be before end date") {
		t.Fatalf("error = %v, want date range validation", err)
	}
}

func TestVitalsErrorsReportsSearchRejectsUnsupportedTimeZoneBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"vitals",
		"errors",
		"reports",
		"search",
		"--package",
		"com.example.app",
		"--start-date",
		"2026-05-01",
		"--end-date",
		"2026-05-19",
		"--time-zone",
		"America/Los_Angeles",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected timezone validation error")
	}
	if !strings.Contains(err.Error(), "only support UTC") {
		t.Fatalf("error = %v, want timezone validation", err)
	}
}

func TestVitalsAnomaliesListRejectsNegativePageSizeBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"vitals",
		"anomalies",
		"list",
		"--package",
		"com.example.app",
		"--page-size",
		"-1",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected page size validation error")
	}
	if !strings.Contains(err.Error(), "page size cannot be negative") {
		t.Fatalf("error = %v, want page size validation", err)
	}
}
