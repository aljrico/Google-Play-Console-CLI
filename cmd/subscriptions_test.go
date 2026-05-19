package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSubscriptionsListRejectsInvalidPageSizeBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
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

func TestSubscriptionsBasePlanActivateDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"base-plan",
		"activate",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
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
		`"action":"activate"`,
		`"basePlanId":"monthly"`,
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

func TestSubscriptionsCreateDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")
	bodyPath := filepath.Join(t.TempDir(), "subscription.json")
	if err := os.WriteFile(bodyPath, []byte(`{
		"packageName":"ignored",
		"productId":"ignored",
		"listings":[{"languageCode":"en-US","title":"Premium","description":"Full access"}],
		"basePlans":[{
			"basePlanId":"monthly",
			"state":"ACTIVE",
			"autoRenewingBasePlanType":{"billingPeriodDuration":"P1M"},
			"offerTags":[{"tag":"public"}],
			"regionalConfigs":[{"regionCode":"US","newSubscriberAvailability":true,"price":{"currencyCode":"USD","units":"4","nanos":990000000}}],
			"otherRegionsConfig":{"newSubscriberAvailability":true,"usdPrice":{"currencyCode":"USD","units":"4"},"eurPrice":{"currencyCode":"EUR","units":"4"}}
		}]
	}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--from-json",
		bodyPath,
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
		`"action":"create"`,
		`"dryRun":true`,
		`"created":false`,
		`"packageName":"com.example.app"`,
		`"productId":"premium"`,
		`"basePlanId":"monthly"`,
		`"regionsVersion":"2026/05"`,
		`"currencyCode":"USD"`,
		`"nanos":990000000`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, `"state":"ACTIVE"`) || strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect state or auth", output)
	}
}

func TestSubscriptionsCreateBasicFlagsDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--listing",
		"en-US,Premium,Full access",
		"--base-plan-id",
		"monthly",
		"--billing-period",
		"P1M",
		"--price",
		"us:USD:4:990000000",
		"--restricted-country",
		"br",
		"--eea-withdrawal-right-type",
		"WITHDRAWAL_RIGHT_SERVICE",
		"--tokenized-digital-asset",
		"false",
		"--regional-tax-tier",
		"FR:TAX_TIER_NEWS_1",
		"--regional-streaming-tax",
		"US:STREAMING_TAX_TYPE_TELCO_VIDEO_SALES",
		"--offer-tag",
		"public",
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
		`"productId":"premium"`,
		`"languageCode":"en-US"`,
		`"title":"Premium"`,
		`"basePlanId":"monthly"`,
		`"type":"autoRenewing"`,
		`"billingPeriodDuration":"P1M"`,
		`"legacyCompatible":true`,
		`"offerTags":["public"]`,
		`"regionCode":"US"`,
		`"newSubscriberAvailability":true`,
		`"currencyCode":"USD"`,
		`"nanos":990000000`,
		`"restrictedCountries":["BR"]`,
		`"eeaWithdrawalRightType":"WITHDRAWAL_RIGHT_SERVICE"`,
		`"isTokenizedDigitalAsset":false`,
		`"taxTier":"TAX_TIER_NEWS_1"`,
		`"streamingTaxType":"STREAMING_TAX_TYPE_TELCO_VIDEO_SALES"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionsCreateBasicFlagsCanDisableLegacyCompatibility(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--listing",
		"en-US,Premium,Full access",
		"--base-plan-id",
		"monthly",
		"--billing-period",
		"P1M",
		"--price",
		"US:USD:4",
		"--legacy-compatible=false",
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
	if strings.Contains(output, `"legacyCompatible":true`) {
		t.Fatalf("output = %s, did not expect legacy-compatible base plan", output)
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionsCreateBasicPrepaidFlagsDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--listing",
		"en-US,Premium,Full access",
		"--base-plan-id",
		"monthly-prepaid",
		"--prepaid",
		"--billing-period",
		"P1M",
		"--time-extension",
		"TIME_EXTENSION_ACTIVE",
		"--price",
		"us:USD:4:990000000",
		"--offer-tag",
		"public",
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
		`"productId":"premium"`,
		`"basePlanId":"monthly-prepaid"`,
		`"type":"prepaid"`,
		`"billingPeriodDuration":"P1M"`,
		`"timeExtension":"TIME_EXTENSION_ACTIVE"`,
		`"offerTags":["public"]`,
		`"regionCode":"US"`,
		`"newSubscriberAvailability":true`,
		`"currencyCode":"USD"`,
		`"nanos":990000000`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, `"legacyCompatible":true`) || strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect legacy compatibility or auth", output)
	}
}

func TestSubscriptionsCreateBasicInstallmentsFlagsDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--listing",
		"en-US,Premium,Full access",
		"--base-plan-id",
		"monthly-installments",
		"--installments",
		"--billing-period",
		"P1M",
		"--committed-payments",
		"12",
		"--renewal-type",
		"RENEWAL_TYPE_RENEWS_WITHOUT_COMMITMENT",
		"--price",
		"us:USD:4:990000000",
		"--offer-tag",
		"public",
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
		`"productId":"premium"`,
		`"basePlanId":"monthly-installments"`,
		`"type":"installments"`,
		`"billingPeriodDuration":"P1M"`,
		`"committedPaymentsCount":12`,
		`"renewalType":"RENEWAL_TYPE_RENEWS_WITHOUT_COMMITMENT"`,
		`"offerTags":["public"]`,
		`"regionCode":"US"`,
		`"newSubscriberAvailability":true`,
		`"currencyCode":"USD"`,
		`"nanos":990000000`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, `"legacyCompatible":true`) || strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect legacy compatibility or auth", output)
	}
}

func TestSubscriptionsCreateBasicFlagsRejectTimeExtensionWithoutPrepaidBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--listing",
		"en-US,Premium,Full access",
		"--base-plan-id",
		"monthly",
		"--billing-period",
		"P1M",
		"--time-extension=",
		"--price",
		"US:USD:4",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected time-extension validation error")
	}
	if !strings.Contains(err.Error(), "--time-extension requires --prepaid") {
		t.Fatalf("error = %v, want time-extension prepaid validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestSubscriptionsCreateBasicFlagsRejectDuplicateRestrictedCountryBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--listing",
		"en-US,Premium,Full access",
		"--base-plan-id",
		"monthly",
		"--billing-period",
		"P1M",
		"--price",
		"US:USD:4",
		"--restricted-country",
		"br",
		"--restricted-country",
		"BR",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected duplicate restricted country validation error")
	}
	if !strings.Contains(err.Error(), "restricted country BR is duplicated") {
		t.Fatalf("error = %v, want duplicate restricted country validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestSubscriptionsCreateBasicFlagsRejectInvalidTokenizedDigitalAssetBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--listing",
		"en-US,Premium,Full access",
		"--base-plan-id",
		"monthly",
		"--billing-period",
		"P1M",
		"--price",
		"US:USD:4",
		"--tokenized-digital-asset",
		"maybe",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected tokenized digital asset validation error")
	}
	if !strings.Contains(err.Error(), "invalid syntax") {
		t.Fatalf("error = %v, want bool parse validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestSubscriptionsCreateBasicFlagsRejectCommittedPaymentsWithoutInstallmentsBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--listing",
		"en-US,Premium,Full access",
		"--base-plan-id",
		"monthly",
		"--billing-period",
		"P1M",
		"--committed-payments",
		"0",
		"--price",
		"US:USD:4",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected committed-payments validation error")
	}
	if !strings.Contains(err.Error(), "--committed-payments requires --installments") {
		t.Fatalf("error = %v, want committed-payments installments validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestSubscriptionsCreateBasicFlagsRejectLegacyCompatibleWithInstallmentsBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--listing",
		"en-US,Premium,Full access",
		"--base-plan-id",
		"monthly-installments",
		"--installments",
		"--billing-period",
		"P1M",
		"--committed-payments",
		"12",
		"--renewal-type",
		"RENEWAL_TYPE_RENEWS_WITHOUT_COMMITMENT",
		"--price",
		"US:USD:4",
		"--legacy-compatible=false",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected legacy-compatible validation error")
	}
	if !strings.Contains(err.Error(), "--legacy-compatible cannot be used with --installments") {
		t.Fatalf("error = %v, want legacy-compatible installments validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestSubscriptionsCreateBasicFlagsRejectPrepaidWithInstallmentsBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--listing",
		"en-US,Premium,Full access",
		"--base-plan-id",
		"monthly-installments",
		"--prepaid",
		"--installments",
		"--billing-period",
		"P1M",
		"--price",
		"US:USD:4",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected mutually exclusive base plan type validation error")
	}
	if !strings.Contains(err.Error(), "--prepaid and --installments cannot be used together") {
		t.Fatalf("error = %v, want mutually exclusive base plan type validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestSubscriptionsCreateBasicFlagsRejectLegacyCompatibleWithPrepaidBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--listing",
		"en-US,Premium,Full access",
		"--base-plan-id",
		"monthly-prepaid",
		"--prepaid",
		"--billing-period",
		"P1M",
		"--price",
		"US:USD:4",
		"--legacy-compatible=false",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected legacy-compatible validation error")
	}
	if !strings.Contains(err.Error(), "--legacy-compatible cannot be used with --prepaid") {
		t.Fatalf("error = %v, want legacy-compatible prepaid validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestSubscriptionsCreateRejectsJSONWithBasicFlagsBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")
	bodyPath := filepath.Join(t.TempDir(), "subscription.json")
	if err := os.WriteFile(bodyPath, []byte(`{
		"listings":[{"languageCode":"en-US","title":"Premium","description":"Full access"}],
		"basePlans":[{"basePlanId":"monthly","autoRenewingBasePlanType":{"billingPeriodDuration":"P1M"},"regionalConfigs":[{"regionCode":"US","newSubscriberAvailability":true,"price":{"currencyCode":"USD","units":"4"}}]}]
	}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--from-json",
		bodyPath,
		"--listing",
		"en-US,Premium,Full access",
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

func TestSubscriptionsCreateRejectsInvalidBodyBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")
	bodyPath := filepath.Join(t.TempDir(), "subscription.json")
	if err := os.WriteFile(bodyPath, []byte(`{"listings":[{"languageCode":"en-US","title":"Premium"}]}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
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
		t.Fatal("expected body validation error")
	}
	if !strings.Contains(err.Error(), "requires at least one base plan") {
		t.Fatalf("error = %v, want base plan validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestSubscriptionsBasePlanRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"base-plan",
		"deactivate",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
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

func TestSubscriptionsBasePlanDeleteDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"base-plan",
		"delete",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"productId":"premium"`,
		`"basePlanId":"monthly"`,
		`"dryRun":true`,
		`"deleted":false`,
		`"steps":["delete base plan"]`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionsBasePlanDeleteRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"base-plan",
		"delete",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
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

func TestSubscriptionsBatchPatchListingsDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"batch-patch-listings",
		"--package",
		"com.example.app",
		"--listing",
		`premium,en-US,"Premium, Plus","Full access"`,
		"--listing",
		"vip,es-ES,VIP,Acceso completo",
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
		`"productId":"premium"`,
		`"languageCode":"en-US"`,
		`"title":"Premium, Plus"`,
		`"productId":"vip"`,
		`"regionsVersion":"2026/05"`,
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

func TestSubscriptionsBatchPatchListingsRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"batch-patch-listings",
		"--package",
		"com.example.app",
		"--listing",
		"premium,en-US,Premium,Full access",
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

func TestSubscriptionsBatchPatchListingsRejectsMultipleCSVRecordsBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"batch-patch-listings",
		"--package",
		"com.example.app",
		"--listing",
		"premium,en-US,Premium,Full access\nvip,en-US,VIP,All access",
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

func TestSubscriptionsBasePlanBatchDeactivateDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"base-plan",
		"batch-deactivate",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--base-plan-id",
		"annual",
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
		`"basePlanId":"monthly"`,
		`"basePlanId":"annual"`,
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

func TestSubscriptionsBasePlanBatchDeactivateDryRunInfersWildcardProductBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"base-plan",
		"batch-deactivate",
		"--package",
		"com.example.app",
		"--base-plan",
		"premium/monthly",
		"--base-plan",
		"vip/annual",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"productId":"-"`,
		`"productId":"premium"`,
		`"productId":"vip"`,
		`"basePlanId":"monthly"`,
		`"basePlanId":"annual"`,
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

func TestSubscriptionsBasePlanBatchActivateRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"base-plan",
		"batch-activate",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
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

func TestSubscriptionsBasePlanBatchActivateRejectsMissingBasePlanBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"base-plan",
		"batch-activate",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected missing base plan validation error")
	}
	if !strings.Contains(err.Error(), "at least one subscription base plan is required") {
		t.Fatalf("error = %v, want missing base plan validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestSubscriptionsBasePlanBatchMigratePricesDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"base-plan",
		"batch-migrate-prices",
		"--package",
		"com.example.app",
		"--regions-version",
		"2026/05",
		"--migration",
		"premium/monthly/US/2026-05-01T00:00:00Z",
		"--migration",
		"premium/monthly/BR/2026-05-01T00:00:00Z",
		"--price-increase-type",
		"optOut",
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
		`"productId":"premium"`,
		`"basePlanId":"monthly"`,
		`"regionCode":"US"`,
		`"regionCode":"BR"`,
		`"priceIncreaseType":"optOut"`,
		`"regionsVersion":"2026/05"`,
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

func TestSubscriptionsBasePlanBatchMigratePricesRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"base-plan",
		"batch-migrate-prices",
		"--package",
		"com.example.app",
		"--regions-version",
		"2026/05",
		"--migration",
		"premium/monthly/US/2026-05-01T00:00:00Z",
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

func TestSubscriptionsBasePlanBatchPatchPricesDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"base-plan",
		"batch-patch-prices",
		"--package",
		"com.example.app",
		"--regions-version",
		"2026/05",
		"--price",
		"premium/monthly/US:USD:4:990000000",
		"--price",
		"premium/monthly/BR:BRL:19:990000000",
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
		`"productId":"premium"`,
		`"basePlanId":"monthly"`,
		`"regionCode":"US"`,
		`"currencyCode":"USD"`,
		`"nanos":990000000`,
		`"regionsVersion":"2026/05"`,
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

func TestSubscriptionsBasePlanBatchPatchPricesRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"base-plan",
		"batch-patch-prices",
		"--package",
		"com.example.app",
		"--regions-version",
		"2026/05",
		"--price",
		"premium/monthly/US:USD:4:990000000",
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

func TestSubscriptionsGetRejectsInvalidProductIDBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"get",
		"--package",
		"com.example.app",
		"--product-id",
		"Premium",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected product ID validation error")
	}
	if !strings.Contains(err.Error(), "subscription product ID") {
		t.Fatalf("error = %v, want product ID validation", err)
	}
}

func TestSubscriptionsBatchGetRejectsMissingProductIDBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"batch-get",
		"--package",
		"com.example.app",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected product ID validation error")
	}
	if !strings.Contains(err.Error(), "at least one subscription product ID") {
		t.Fatalf("error = %v, want missing product ID validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestSubscriptionsPatchDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"patch",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--listing-language",
		"en-US",
		"--title",
		"Premium",
		"--description",
		"Full access",
		"--benefit",
		"Unlimited projects",
		"--regions-version",
		"2022/02",
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
		`"productId":"premium"`,
		`"dryRun":true`,
		`"languageCode":"en-US"`,
		`"title":"Premium"`,
		`"updateMask":"listings"`,
		`"regionsVersion":"2022/02"`,
		`"latencyTolerance":"latencyTolerant"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionsPatchRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"patch",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--listing-language",
		"en-US",
		"--title",
		"Premium",
		"--regions-version",
		"2022/02",
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
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestSubscriptionsDeleteDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"delete",
		"--package",
		"com.example.app",
		"--product-id",
		"premium_monthly",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"productId":"premium_monthly"`,
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

func TestSubscriptionsDeleteRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"delete",
		"--package",
		"com.example.app",
		"--product-id",
		"premium_monthly",
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
		t.Fatalf("error = %v, did not expect auth", err)
	}
}
