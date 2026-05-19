package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestAnalyticsStatsSummarizeOutputsTotalsWithoutAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")
	file := writeRootTestPathContent(t, filepath.Join(t.TempDir(), "store_performance.csv"), "Date,Package name,Country/region,Store listing visitors,Store listing acquisitions\n2026-05-01,com.example.app,US,10,2\n2026-05-02,com.example.app,US,15,3\n")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"analytics",
		"stats",
		"summarize",
		"--file",
		file,
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"rows":2`,
		`"packageName":"com.example.app"`,
		`"startDate":"2026-05-01"`,
		`"endDate":"2026-05-02"`,
		`"name":"Store listing visitors","aggregation":"sum","value":"25"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestAnalyticsStatsDownloadDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")
	outputPath := filepath.Join(t.TempDir(), "stats.csv")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"analytics",
		"stats",
		"download",
		"--bucket",
		"pubsite_prod_rev_0123456789",
		"--object",
		"stats/store_performance/store_performance_com.example.app_202605_country.csv",
		"--file",
		outputPath,
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"dryRun":true`, `"downloaded":false`, `"object":"stats/store_performance/store_performance_com.example.app_202605_country.csv"`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}
