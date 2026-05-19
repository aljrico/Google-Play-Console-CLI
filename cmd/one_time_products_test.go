package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOneTimeProductsListRejectsInvalidPageSizeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"list",
		"--package",
		"com.example.app",
		"--page-size",
		"1001",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected page size validation error")
	}
	if !strings.Contains(err.Error(), "page size") {
		t.Fatalf("error = %v, want page size validation", err)
	}
}

func TestOneTimeProductsGetRejectsInvalidProductIDBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"get",
		"--package",
		"com.example.app",
		"--product-id",
		"Coins",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected product ID validation error")
	}
	if !strings.Contains(err.Error(), "one-time product ID") {
		t.Fatalf("error = %v, want product ID validation", err)
	}
}

func TestOneTimeProductsBatchGetRejectsDuplicatesBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"batch-get",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--product-id",
		"coins_100",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected duplicate product ID validation error")
	}
	if !strings.Contains(err.Error(), "duplicated") {
		t.Fatalf("error = %v, want duplicate validation", err)
	}
}

func TestOneTimeProductsBatchGetRejectsMissingProductIDBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"batch-get",
		"--package",
		"com.example.app",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected missing product ID validation error")
	}
	if !strings.Contains(err.Error(), "at least one") {
		t.Fatalf("error = %v, want missing product ID validation", err)
	}
}

func TestOneTimeProductsCreateDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	bodyPath := filepath.Join(t.TempDir(), "one-time-product.json")
	if err := os.WriteFile(bodyPath, []byte(`{
		"packageName":"ignored.by.flags",
		"productId":"ignored_by_flags",
		"listings":[{"languageCode":"en-US","title":"100 coins","description":"Buy coins."}],
		"purchaseOptions":[{
			"purchaseOptionId":"buy",
			"state":"ACTIVE",
			"buyOption":{"legacyCompatible":true},
			"regionalPricingAndAvailabilityConfigs":[{"regionCode":"US","availability":"AVAILABLE","price":{"currencyCode":"USD","units":"1","nanos":990000000}}],
			"newRegionsConfig":{"availability":"AVAILABLE","usdPrice":{"currencyCode":"USD","units":"1"},"eurPrice":{"currencyCode":"EUR","units":"1"}}
		}]
	}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--from-json",
		bodyPath,
		"--regions-version",
		"2026/05",
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
		`"dryRun":true`,
		`"created":false`,
		`"packageName":"com.example.app"`,
		`"productId":"coins_100"`,
		`"purchaseOptionId":"buy"`,
		`"regionsVersion":"2026/05"`,
		`"latencyTolerance":"latencyTolerant"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
	if strings.Contains(output, `"state":"ACTIVE"`) {
		t.Fatalf("output = %s, did not expect output-only state from input JSON", output)
	}
}

func TestOneTimeProductsCreateBasicFlagsDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--listing",
		"en-US,100 coins,Buy coins.",
		"--price",
		"us:USD:1:990000000",
		"--purchase-option-id",
		"buy",
		"--offer-tag",
		"public",
		"--multi-quantity",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"dryRun":true`,
		`"created":false`,
		`"productId":"coins_100"`,
		`"languageCode":"en-US"`,
		`"title":"100 coins"`,
		`"purchaseOptionId":"buy"`,
		`"type":"buy"`,
		`"legacyCompatible":true`,
		`"multiQuantityEnabled":true`,
		`"offerTags":["public"]`,
		`"regionCode":"US"`,
		`"availability":"available"`,
		`"currencyCode":"USD"`,
		`"nanos":990000000`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestOneTimeProductsCreateRejectsJSONWithBasicFlagsBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	bodyPath := filepath.Join(t.TempDir(), "one-time-product.json")
	if err := os.WriteFile(bodyPath, []byte(`{
		"listings":[{"languageCode":"en-US","title":"100 coins","description":"Buy coins."}],
		"purchaseOptions":[{"purchaseOptionId":"buy","buyOption":{"legacyCompatible":true},"regionalPricingAndAvailabilityConfigs":[{"regionCode":"US","availability":"AVAILABLE","price":{"currencyCode":"USD","units":"1"}}]}]
	}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--from-json",
		bodyPath,
		"--listing",
		"en-US,100 coins,Buy coins.",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected from-json and basic flags validation error")
	}
	if !strings.Contains(err.Error(), "cannot be combined") {
		t.Fatalf("error = %v, want combination validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductsCreateBasicFlagsRejectsTooManyOfferTagsBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	args := []string{
		"one-time-products",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--listing",
		"en-US,100 coins,Buy coins.",
		"--price",
		"US:USD:1",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	}
	for index := range 21 {
		args = append(args, "--offer-tag", fmt.Sprintf("tag%d", index))
	}
	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected offer tag limit validation error")
	}
	if !strings.Contains(err.Error(), "at most 20 offer tags") {
		t.Fatalf("error = %v, want offer tag limit validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductsCreateRejectsInvalidBodyBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	bodyPath := filepath.Join(t.TempDir(), "one-time-product.json")
	if err := os.WriteFile(bodyPath, []byte(`{"listings":[{"languageCode":"en-US","title":"100 coins","description":"Buy coins."}]}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--from-json",
		bodyPath,
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected product body validation error")
	}
	if !strings.Contains(err.Error(), "requires at least one") {
		t.Fatalf("error = %v, want body validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductsPatchDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"patch",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--listing-language",
		"en-US",
		"--title",
		"100 coins",
		"--description",
		"Buy a stack of coins.",
		"--regions-version",
		"2026/05",
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
		`"productId":"coins_100"`,
		`"languageCode":"en-US"`,
		`"title":"100 coins"`,
		`"updateMask":"listings"`,
		`"latencyTolerance":"latencyTolerant"`,
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

func TestOneTimeProductsPatchRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"patch",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--listing-language",
		"en-US",
		"--title",
		"100 coins",
		"--regions-version",
		"2026/05",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductsBatchPatchListingsDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"batch-patch-listings",
		"--package",
		"com.example.app",
		"--listing",
		"coins_100,en-US,100 coins,Buy coins.",
		"--listing",
		"coins_500,es-ES,500 monedas,Compra monedas.",
		"--regions-version",
		"2026/05",
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
		`"productId":"coins_100"`,
		`"languageCode":"es-ES"`,
		`"title":"500 monedas"`,
		`"updateMask":"listings"`,
		`"latencyTolerance":"latencyTolerant"`,
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

func TestOneTimeProductsBatchPatchListingsRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"batch-patch-listings",
		"--package",
		"com.example.app",
		"--listing",
		"coins_100,en-US,100 coins,Buy coins.",
		"--regions-version",
		"2026/05",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductsBatchPatchListingsRejectsMultipleCSVRecordsBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"batch-patch-listings",
		"--package",
		"com.example.app",
		"--listing",
		"coins_100,en-US,100 coins,Buy coins.\ncoins_500,en-US,500 coins,Buy more coins.",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected CSV record validation error")
	}
	if !strings.Contains(err.Error(), "exactly one CSV record") {
		t.Fatalf("error = %v, want CSV record validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductsDeleteDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"delete",
		"--package",
		"com.example.app",
		"--product-id",
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
		`"productId":"coins_100"`,
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

func TestOneTimeProductsDeleteRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"delete",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
}

func TestOneTimeProductsBatchDeleteDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"batch-delete",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--product-id",
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
		`"productIds":["coins_100","coins_500"]`,
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

func TestOneTimeProductsBatchDeleteRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"batch-delete",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
}

func TestOneTimeProductsBatchDeleteRejectsDuplicatesBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"batch-delete",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--product-id",
		"coins_100",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected duplicate product ID validation error")
	}
	if !strings.Contains(err.Error(), "duplicated") {
		t.Fatalf("error = %v, want duplicate validation", err)
	}
}

func TestOneTimeProductsPurchaseOptionDeactivateDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"purchase-option",
		"deactivate",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
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
		`"action":"deactivate"`,
		`"purchaseOptionId":"buy"`,
		`"latencyTolerance":"latencyTolerant"`,
		`"dryRun":true`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestOneTimeProductsPurchaseOptionRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"purchase-option",
		"activate",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductsPurchaseOptionBatchDeleteDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"purchase-option",
		"batch-delete",
		"--package",
		"com.example.app",
		"--product-id",
		"-",
		"--purchase-option",
		"coins_100/buy",
		"--purchase-option",
		"coins_500/rent",
		"--latency-tolerance",
		"latencyTolerant",
		"--force",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"parentProductId":"-"`,
		`"purchaseOptionId":"buy"`,
		`"latencyTolerance":"latencyTolerant"`,
		`"force":true`,
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

func TestOneTimeProductsPurchaseOptionBatchDeleteRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"purchase-option",
		"batch-delete",
		"--package",
		"com.example.app",
		"--purchase-option",
		"coins_100/buy",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductsPurchaseOptionBatchDeleteInfersSingleParentBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"purchase-option",
		"batch-delete",
		"--package",
		"com.example.app",
		"--purchase-option",
		"coins_100/buy",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, `"parentProductId":"coins_100"`) {
		t.Fatalf("output = %s, want inferred parent product", output)
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestOneTimeProductsPurchaseOptionBatchDeleteRejectsRepeatedProductBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"purchase-option",
		"batch-delete",
		"--package",
		"com.example.app",
		"--purchase-option",
		"coins_100/buy",
		"--purchase-option",
		"coins_100/rent",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected repeated product validation error")
	}
	if !strings.Contains(err.Error(), "at most one request per one-time product") {
		t.Fatalf("error = %v, want repeated product validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestPurchaseOptionBatchPatchAvailabilityDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"purchase-option",
		"batch-patch-availability",
		"--package",
		"com.example.app",
		"--availability",
		"coins_100/buy/us:noLongerAvailable",
		"--availability",
		"coins_100/buy/FR:availableForOffersOnly",
		"--regions-version",
		"2026/05",
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
		`"productId":"coins_100"`,
		`"purchaseOptionId":"buy"`,
		`"regionCode":"US"`,
		`"availability":"noLongerAvailable"`,
		`"availability":"availableForOffersOnly"`,
		`"updateMask":"purchaseOptions"`,
		`"regionsVersion":"2026/05"`,
		`"latencyTolerance":"latencyTolerant"`,
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

func TestPurchaseOptionBatchPatchAvailabilityRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"purchase-option",
		"batch-patch-availability",
		"--package",
		"com.example.app",
		"--availability",
		"coins_100/buy/US:available",
		"--regions-version",
		"2026/05",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirmation gate")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestPurchaseOptionBatchPatchAvailabilityRejectsMalformedPatchBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"purchase-option",
		"batch-patch-availability",
		"--package",
		"com.example.app",
		"--availability",
		"coins_100/buy:available",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected availability format validation error")
	}
	if !strings.Contains(err.Error(), "productId/purchaseOptionId/REGION") {
		t.Fatalf("error = %v, want availability format validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestPurchaseOptionBatchPatchPricesDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"purchase-option",
		"batch-patch-prices",
		"--package",
		"com.example.app",
		"--price",
		"coins_100/buy/us:USD:3:490000000",
		"--price",
		"coins_100/buy/FR:EUR:2",
		"--regions-version",
		"2026/05",
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
		`"productId":"coins_100"`,
		`"purchaseOptionId":"buy"`,
		`"regionCode":"US"`,
		`"currencyCode":"USD"`,
		`"units":3`,
		`"nanos":490000000`,
		`"updateMask":"purchaseOptions"`,
		`"regionsVersion":"2026/05"`,
		`"latencyTolerance":"latencyTolerant"`,
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

func TestPurchaseOptionBatchPatchPricesRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"purchase-option",
		"batch-patch-prices",
		"--package",
		"com.example.app",
		"--price",
		"coins_100/buy/US:USD:3:490000000",
		"--regions-version",
		"2026/05",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirmation gate")
	}
	if !strings.Contains(err.Error(), "requires --confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation gate", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestPurchaseOptionBatchPatchPricesRejectsMalformedPatchBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-products",
		"purchase-option",
		"batch-patch-prices",
		"--package",
		"com.example.app",
		"--price",
		"coins_100/buy:USD:3",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected price format validation error")
	}
	if !strings.Contains(err.Error(), "productId/purchaseOptionId/REGION") {
		t.Fatalf("error = %v, want price format validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}
