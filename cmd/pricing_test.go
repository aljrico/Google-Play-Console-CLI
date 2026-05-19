package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestPricingConvertRejectsInvalidPriceBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

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

func TestPricingBuildPricePatchesOutputsPatchArgsWithoutAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")
	jsonPath := writeTempPricingConversion(t, `{
		"packageName": "com.example.app",
		"sourcePrice": {"currencyCode": "USD", "units": 9, "nanos": 990000000},
		"regionVersion": "2026/05",
		"convertedRegionPrices": {
			"US": {"regionCode": "US", "price": {"currencyCode": "USD", "units": 9, "nanos": 990000000}},
			"BR": {"regionCode": "BR", "price": {"currencyCode": "BRL", "units": 19}}
		}
	}`)

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"pricing",
		"build-price-patches",
		"--from-json",
		jsonPath,
		"--target",
		"one-time-product",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"target":"one-time-product"`,
		`"priceArgs":["coins_100/buy/BR:BRL:19","coins_100/buy/US:USD:9:990000000"]`,
		`"suggestedCommand":["playpub","one-time-products","purchase-option","batch-patch-prices","--package","com.example.app","--regions-version","2026/05","--price","coins_100/buy/BR:BRL:19","--price","coins_100/buy/US:USD:9:990000000","--dry-run"]`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestPricingBuildPricePatchesRejectsInvalidInputBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")
	jsonPath := writeTempPricingConversion(t, `{
		"packageName": "com.example.app",
		"sourcePrice": {"currencyCode": "USD", "units": 1},
		"convertedRegionPrices": {
			"us": {"price": {"currencyCode": "USD", "units": 1}}
		}
	}`)

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"pricing",
		"build-price-patches",
		"--from-json",
		jsonPath,
		"--target",
		"in-app-product",
		"--sku",
		"coins_100",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want validation error")
	}
	if !strings.Contains(err.Error(), "two-letter ISO 3166") {
		t.Fatalf("error = %v, want region validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestPricingBuildPricePatchesRejectsMismatchedPackageFlag(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")
	jsonPath := writeTempPricingConversion(t, `{
		"packageName": "com.example.staging",
		"sourcePrice": {"currencyCode": "USD", "units": 1},
		"convertedRegionPrices": {
			"US": {"price": {"currencyCode": "USD", "units": 1}}
		}
	}`)

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"pricing",
		"--package",
		"com.example.prod",
		"build-price-patches",
		"--from-json",
		jsonPath,
		"--target",
		"in-app-product",
		"--sku",
		"coins_100",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want package mismatch")
	}
	if !strings.Contains(err.Error(), "does not match conversion packageName") {
		t.Fatalf("error = %v, want package mismatch", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestPricingBuildPricePatchesRejectsDuplicateRegionKeyInJSON(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")
	jsonPath := writeTempPricingConversion(t, `{
		"packageName": "com.example.app",
		"sourcePrice": {"currencyCode": "USD", "units": 9, "nanos": 990000000},
		"regionVersion": "2026/05",
		"convertedRegionPrices": {
			"US": {"price": {"currencyCode": "USD", "units": 9, "nanos": 990000000}},
			"US": {"price": {"currencyCode": "USD", "units": 19, "nanos": 990000000}}
		}
	}`)

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"pricing",
		"build-price-patches",
		"--from-json",
		jsonPath,
		"--target",
		"in-app-product",
		"--sku",
		"coins_100",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want duplicate region key error")
	}
	if !strings.Contains(err.Error(), `"US" is duplicated`) {
		t.Fatalf("error = %v, want duplicate region key", err)
	}
}

func writeTempPricingConversion(t *testing.T, content string) string {
	t.Helper()
	path := t.TempDir() + "/conversion.json"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}
