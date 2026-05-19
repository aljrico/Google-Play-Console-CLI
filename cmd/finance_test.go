package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestFinanceReportsSummarizeOutputsTotalsWithoutAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	file := writeRootTestPathContent(t, filepath.Join(t.TempDir(), "earnings.csv"), "Transaction Type,Merchant Currency,Amount (Merchant Currency)\nCharge,USD,9.99\nCharge,USD,1.01\nGoogle fee,USD,-1.50\n")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"finance",
		"reports",
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
		`"reportType":"earnings"`,
		`"rows":3`,
		`"amountColumn":"Amount (Merchant Currency)"`,
		`"transactionType":"Charge","count":2,"total":"11","currency":"USD"`,
		`"transactionType":"Google fee","count":1,"total":"-1.5","currency":"USD"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestFinanceReportsDownloadDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	outputPath := filepath.Join(t.TempDir(), "earnings.zip")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"finance",
		"reports",
		"download",
		"--bucket",
		"pubsite_prod_rev_0123456789",
		"--object",
		"earnings/earnings_202605.zip",
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
	for _, want := range []string{`"dryRun":true`, `"downloaded":false`, `"bucket":"pubsite_prod_rev_0123456789"`, `"object":"earnings/earnings_202605.zip"`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}
