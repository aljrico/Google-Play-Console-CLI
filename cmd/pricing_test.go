package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestPricingConvertRejectsInvalidPriceBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"pricing",
		"convert-region-prices",
		"--package",
		"com.example.app",
		"--currency",
		"USD",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected price validation error")
	}
	if !strings.Contains(err.Error(), "price must be greater than 0") {
		t.Fatalf("error = %v, want price validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}
