package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestInsightsAnomaliesSummarizeOutputsCountsWithoutAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")
	file := writeRootTestPathContent(t, filepath.Join(t.TempDir(), "anomalies.json"), `{
		"packageName": "com.example.app",
		"anomalies": [
			{"name":"a1","metricSet":"apps/com.example.app/crashRateMetricSet","metric":{"metric":"crashRate"}},
			{"name":"a2","metricSet":"apps/com.example.app/crashRateMetricSet","metric":{"metric":"crashRate"}}
		]
	}`)

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"insights",
		"anomalies",
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
		`"total":2`,
		`"packageName":"com.example.app"`,
		`"name":"apps/com.example.app/crashRateMetricSet","count":2`,
		`"top metric: crashRate (2)"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestInsightsReportsSummarizeOutputsFinanceAndStatsWithoutAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")
	dir := t.TempDir()
	financeFile := writeRootTestPathContent(t, filepath.Join(dir, "earnings.csv"), "Transaction Type,Merchant Currency,Amount (Merchant Currency)\nCharge,USD,9.99\nCharge,USD,1.01\nGoogle fee,USD,-1.50\n")
	statsFile := writeRootTestPathContent(t, filepath.Join(dir, "stats.csv"), "Date,Package name,Country,Store listing visitors,Store listing acquisitions,Conversion rate\n2026-05-01,com.example.app,US,10,2,20\n2026-05-02,com.example.app,US,20,3,22\n")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"insights",
		"reports",
		"summarize",
		"--finance-file",
		financeFile,
		"--stats-file",
		statsFile,
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"reportType":"earnings"`,
		`"transactionType":"Charge"`,
		`"packageName":"com.example.app"`,
		`"name":"Store listing acquisitions"`,
		`"kind":"finance"`,
		`"kind":"stats"`,
		`"name":"netRevenuePerStoreListingAcquisition"`,
		`"value":"1.9"`,
		`top transaction by count is Charge`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}
