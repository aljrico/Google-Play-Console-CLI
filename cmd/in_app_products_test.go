package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestInAppProductsGetRejectsMissingSKUBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"get",
		"--package",
		"com.example.app",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected SKU validation error")
	}
	if !strings.Contains(err.Error(), "SKU") {
		t.Fatalf("error = %v, want SKU validation", err)
	}
}

func TestInAppProductsBatchGetRejectsMissingSKUBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"batch-get",
		"--package",
		"com.example.app",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected SKU validation error")
	}
	if !strings.Contains(err.Error(), "at least one in-app product SKU") {
		t.Fatalf("error = %v, want SKU validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestInAppProductsBatchGetRejectsDuplicateSKUBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"batch-get",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--sku",
		"coins_100",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected duplicate SKU validation error")
	}
	if !strings.Contains(err.Error(), "duplicated") {
		t.Fatalf("error = %v, want duplicate validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestInAppProductsDeleteDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"delete",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--latency-tolerance",
		"latencyTolerant",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"sku":"coins_100"`,
		`"latencyTolerance":"latencyTolerant"`,
		`"dryRun":true`,
		`"deleted":false`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestInAppProductsDeleteRequiresDryRunOrConfirmBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"delete",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirmation validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestInAppProductsBatchDeleteDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"batch-delete",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--sku",
		"coins_500",
		"--latency-tolerance",
		"latencyTolerant",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"skus":["coins_100","coins_500"]`,
		`"latencyTolerance":"latencyTolerant"`,
		`"dryRun":true`,
		`"deleted":false`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestInAppProductsBatchDeleteRequiresDryRunOrConfirmBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"batch-delete",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirmation validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestInAppProductsBatchDeleteRejectsDuplicateSKUBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"batch-delete",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--sku",
		"coins_100",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected duplicate SKU validation error")
	}
	if !strings.Contains(err.Error(), "duplicated") {
		t.Fatalf("error = %v, want duplicate validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestInAppProductsCreateDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"create",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--default-language",
		"en-US",
		"--default-price",
		"USD:1990000",
		"--title",
		"100 coins",
		"--description",
		"A small coin pack.",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"action":"create"`,
		`"dryRun":true`,
		`"purchaseType":"managedUser"`,
		`"priceMicros":"1990000"`,
		`"autoConvertMissingPrices":true`,
		`"created":false`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestInAppProductsCreateRequiresDryRunOrConfirmBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"create",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--default-language",
		"en-US",
		"--default-price",
		"USD:1990000",
		"--title",
		"100 coins",
		"--description",
		"A small coin pack.",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirmation validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation validation", err)
	}
}

func TestInAppProductsCreateRejectsBadPriceBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"create",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--default-language",
		"en-US",
		"--default-price",
		"USD:free",
		"--title",
		"100 coins",
		"--description",
		"A small coin pack.",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected price validation error")
	}
	if !strings.Contains(err.Error(), "price micros") {
		t.Fatalf("error = %v, want price micros validation", err)
	}
}

func TestInAppProductsPatchDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"patch",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--status",
		"inactive",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"action":"patch"`,
		`"dryRun":true`,
		`"status":"inactive"`,
		`"applied":false`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestInAppProductsPatchPriceAndListingDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"patch",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--listing-language",
		"en-US",
		"--default-price",
		"USD:2990000",
		"--title",
		"100 coins",
		"--description",
		"A better coin pack.",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"action":"patch"`,
		`"dryRun":true`,
		`"priceMicros":"2990000"`,
		`"autoConvertMissingPrices":true`,
		`"description":"A better coin pack."`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestInAppProductsPatchRegionalPricesDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"patch",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--regional-price",
		"US:USD:2990000",
		"--regional-price",
		"BR:BRL:9990000",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"regionCode":"US"`,
		`"currency":"BRL"`,
		`"priceMicros":"9990000"`,
		`"autoConvertMissingPrices":true`,
		`"applied":false`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestInAppProductsPatchTaxComplianceDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"patch",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--eea-withdrawal-right-type",
		"WITHDRAWAL_RIGHT_DIGITAL_CONTENT",
		"--tokenized-digital-asset",
		"false",
		"--regional-tax-tier",
		"FR:TAX_TIER_NEWS_1",
		"--regional-streaming-tax",
		"US:STREAMING_TAX_TYPE_TELCO_VIDEO_SALES",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"eeaWithdrawalRightType":"WITHDRAWAL_RIGHT_DIGITAL_CONTENT"`,
		`"isTokenizedDigitalAsset":false`,
		`"taxTier":"TAX_TIER_NEWS_1"`,
		`"streamingTaxType":"STREAMING_TAX_TYPE_TELCO_VIDEO_SALES"`,
		`"dryRun":true`,
		`"applied":false`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestInAppProductsPatchRejectsBadTokenizedDigitalAssetBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"patch",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--tokenized-digital-asset",
		"maybe",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected tokenized digital asset validation error")
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestInAppProductsPatchRequiresDryRunOrConfirmBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"patch",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--status",
		"active",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirmation validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation validation", err)
	}
}

func TestInAppProductsPatchRejectsMissingMutationBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"patch",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected mutation validation error")
	}
	if !strings.Contains(err.Error(), "requires at least one") {
		t.Fatalf("error = %v, want mutation validation", err)
	}
}

func TestInAppProductsPatchRejectsListingWithoutLanguageBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"patch",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--title",
		"100 coins",
		"--description",
		"A better coin pack.",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected listing language validation error")
	}
	if !strings.Contains(err.Error(), "requires --listing-language") {
		t.Fatalf("error = %v, want default language validation", err)
	}
}

func TestInAppProductsPatchRejectsPartialListingBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"patch",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--listing-language",
		"en-US",
		"--title",
		"100 coins",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected partial listing validation error")
	}
	if !strings.Contains(err.Error(), "listing description is required") {
		t.Fatalf("error = %v, want listing description validation", err)
	}
}

func TestInAppProductsPatchRejectsConfirmAndDryRunBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"patch",
		"--package",
		"com.example.app",
		"--sku",
		"coins_100",
		"--status",
		"active",
		"--confirm",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected mutually exclusive flag validation error")
	}
	if !strings.Contains(err.Error(), "cannot be used together") {
		t.Fatalf("error = %v, want mutually exclusive validation", err)
	}
}
