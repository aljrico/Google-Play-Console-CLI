package finance

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSummarizeReportTotalsByTransactionTypeAndCurrency(t *testing.T) {
	file := filepath.Join(t.TempDir(), "earnings.csv")
	content := "Transaction Type,Merchant Currency,Amount (Merchant Currency)\nCharge,USD,9.99\nGoogle fee,USD,-1.50\nCharge refund,USD,-9.99\nCharge,USD,1.01\n"
	if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	summary, err := SummarizeReport(ReportSummaryOptions{File: file})
	if err != nil {
		t.Fatalf("SummarizeReport() error = %v", err)
	}
	if summary.Rows != 4 {
		t.Fatalf("Rows = %d, want 4", summary.Rows)
	}
	if summary.ReportType != "earnings" || summary.AmountColumn != "Amount (Merchant Currency)" || summary.CurrencyColumn != "Merchant Currency" {
		t.Fatalf("summary columns = %#v", summary)
	}
	if len(summary.TransactionTypes) != 3 {
		t.Fatalf("TransactionTypes = %#v, want 3", summary.TransactionTypes)
	}
	charge := summary.TransactionTypes[0]
	if charge.TransactionType != "Charge" || charge.Count != 2 || charge.Total != "11" || charge.Currency != "USD" {
		t.Fatalf("Charge total = %#v", charge)
	}
}

func TestSummarizeReportSupportsBOMHeader(t *testing.T) {
	file := filepath.Join(t.TempDir(), "earnings.csv")
	content := "\ufeffTransaction Type,Merchant Currency,Amount (Merchant Currency)\nCharge,USD,9.99\n"
	if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	summary, err := SummarizeReport(ReportSummaryOptions{File: file})
	if err != nil {
		t.Fatalf("SummarizeReport() error = %v", err)
	}
	if summary.TransactionTypes[0].Total != "9.99" {
		t.Fatalf("summary = %#v", summary)
	}
}

func TestSummarizeReportSupportsEstimatedSalesHeaders(t *testing.T) {
	file := filepath.Join(t.TempDir(), "salesreport.csv")
	content := "Financial Status,Currency of Sale,Charged Amount\nCharged,EUR,12.3456789\nRefunded,EUR,-2.34\n"
	if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	summary, err := SummarizeReport(ReportSummaryOptions{File: file})
	if err != nil {
		t.Fatalf("SummarizeReport() error = %v", err)
	}
	if summary.ReportType != "estimated-sales" {
		t.Fatalf("ReportType = %q, want estimated-sales", summary.ReportType)
	}
	if summary.TransactionTypes[0].Total != "12.3456789" {
		t.Fatalf("summary = %#v, want exact charged amount", summary)
	}
}

func TestSummarizeReportRejectsFractionSyntax(t *testing.T) {
	file := filepath.Join(t.TempDir(), "earnings.csv")
	content := "Transaction Type,Merchant Currency,Amount (Merchant Currency)\nCharge,USD,1/3\n"
	if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	_, err := SummarizeReport(ReportSummaryOptions{File: file})
	if err == nil {
		t.Fatal("SummarizeReport() error = nil, want decimal validation")
	}
}

func TestSummarizeReportRequiresKnownHeaders(t *testing.T) {
	file := filepath.Join(t.TempDir(), "sales.csv")
	if err := os.WriteFile(file, []byte("Type,Amount\nCharge,9.99\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	_, err := SummarizeReport(ReportSummaryOptions{File: file})
	if err == nil {
		t.Fatal("SummarizeReport() error = nil, want header validation")
	}
}
