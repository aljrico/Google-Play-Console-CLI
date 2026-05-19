package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
)

func TestOneTimeProductOffersListRejectsInvalidWildcardParentBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"list",
		"--package",
		"com.example.app",
		"--product-id",
		"-",
		"--purchase-option-id",
		"buy",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected wildcard validation error")
	}
	if !strings.Contains(err.Error(), "purchase option ID") {
		t.Fatalf("error = %v, want purchase option validation", err)
	}
}

func TestOneTimeProductOffersGetRejectsInvalidOfferIDBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"get",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"Intro",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected offer ID validation error")
	}
	if !strings.Contains(err.Error(), "one-time product offer ID") {
		t.Fatalf("error = %v, want offer ID validation", err)
	}
}

func TestOneTimeProductOffersBatchGetRejectsMissingOfferBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-get",
		"--package",
		"com.example.app",
		"--product-id",
		"-",
		"--purchase-option-id",
		"-",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected offer validation error")
	}
	if !strings.Contains(err.Error(), "at least one one-time product offer") {
		t.Fatalf("error = %v, want missing offer validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestOneTimeProductOffersBatchGetRejectsParentMismatchBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-get",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer",
		"coins_500/buy/intro",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected parent mismatch validation error")
	}
	if !strings.Contains(err.Error(), "does not match parent product ID") {
		t.Fatalf("error = %v, want parent product validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestOneTimeProductOffersBatchGetRejectsInvalidOfferIDBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-get",
		"--package",
		"com.example.app",
		"--product-id",
		"-",
		"--purchase-option-id",
		"-",
		"--offer",
		"coins_100/buy/Intro",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected offer ID validation error")
	}
	if !strings.Contains(err.Error(), "one-time product offer ID") {
		t.Fatalf("error = %v, want offer ID validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestOneTimeProductOffersCreateDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")
	bodyPath := filepath.Join(t.TempDir(), "offer.json")
	if err := os.WriteFile(bodyPath, []byte(`{
		"packageName":"ignored.by.flags",
		"productId":"ignored_by_flags",
		"purchaseOptionId":"ignored",
		"offerId":"ignored",
		"state":"ACTIVE",
		"discountedOffer":{"startTime":"2026-06-01T00:00:00Z","endTime":"2026-07-01T00:00:00Z","redemptionLimit":"5"},
		"regionalPricingAndAvailabilityConfigs":[{"regionCode":"US","availability":"AVAILABLE","relativeDiscount":0.5}]
	}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"intro",
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
		`"productId":"coins_100"`,
		`"purchaseOptionId":"buy"`,
		`"offerId":"intro"`,
		`"regionsVersion":"2026/05"`,
		`"latencyTolerance":"latencyTolerant"`,
		`"relativeDiscount":0.5`,
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

func TestOneTimeProductOffersCreateBasicRelativeDiscountDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"intro",
		"--offer-tag",
		"public",
		"--start-time",
		"2026-06-01T00:00:00Z",
		"--end-time",
		"2026-07-01T00:00:00Z",
		"--redemption-limit",
		"5",
		"--relative-discount",
		"us:0.5",
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
		`"purchaseOptionId":"buy"`,
		`"offerId":"intro"`,
		`"type":"discounted"`,
		`"offerTags":["public"]`,
		`"startTime":"2026-06-01T00:00:00Z"`,
		`"endTime":"2026-07-01T00:00:00Z"`,
		`"redemptionLimit":5`,
		`"regionCode":"US"`,
		`"availability":"available"`,
		`"relativeDiscount":0.5`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestOneTimeProductOffersCreateBasicAbsoluteDiscountDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"intro",
		"--start-time",
		"2026-06-01T00:00:00Z",
		"--end-time",
		"2026-07-01T00:00:00Z",
		"--absolute-discount",
		"us:USD:1:500000000",
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
		`"type":"discounted"`,
		`"regionCode":"US"`,
		`"availability":"available"`,
		`"absoluteDiscount":{"currencyCode":"USD","units":1,"nanos":500000000}`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestOneTimeProductOffersCreateBasicNoOverrideDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"intro",
		"--start-time",
		"2026-06-01T00:00:00Z",
		"--end-time",
		"2026-07-01T00:00:00Z",
		"--no-override",
		"us",
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
		`"type":"discounted"`,
		`"regionCode":"US"`,
		`"availability":"available"`,
		`"noOverride":true`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestOneTimeProductOffersCreateBasicPreOrderDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"preorder",
		"--pre-order",
		"--start-time",
		"2026-06-01T00:00:00Z",
		"--end-time",
		"2026-07-01T00:00:00Z",
		"--release-time",
		"2026-08-01T00:00:00Z",
		"--price-change-behavior",
		"PRE_ORDER_PRICE_CHANGE_BEHAVIOR_NEW_ORDERS_ONLY",
		"--no-override",
		"us",
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
		`"type":"preOrder"`,
		`"preOrderOffer":{"startTime":"2026-06-01T00:00:00Z","endTime":"2026-07-01T00:00:00Z","releaseTime":"2026-08-01T00:00:00Z","priceChangeBehavior":"PRE_ORDER_PRICE_CHANGE_BEHAVIOR_NEW_ORDERS_ONLY"}`,
		`"regionCode":"US"`,
		`"availability":"available"`,
		`"noOverride":true`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestOneTimeProductOffersCreateBasicMixedDiscountModesDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"intro",
		"--relative-discount",
		"US:0.5",
		"--absolute-discount",
		"JP:JPY:100",
		"--no-override",
		"BR",
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
		`"regionCode":"US"`,
		`"relativeDiscount":0.5`,
		`"regionCode":"JP"`,
		`"absoluteDiscount":{"currencyCode":"JPY","units":100}`,
		`"regionCode":"BR"`,
		`"noOverride":true`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestOneTimeProductOffersCreateBasicFlagsRejectsDuplicateDiscountRegionBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"intro",
		"--relative-discount",
		"US:0.5",
		"--absolute-discount",
		"US:USD:1",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected duplicated region validation error")
	}
	if !strings.Contains(err.Error(), "region US is duplicated") {
		t.Fatalf("error = %v, want duplicate region validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductOffersCreateBasicFlagsRejectsReleaseTimeWithoutPreOrderBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"intro",
		"--relative-discount",
		"US:0.5",
		"--release-time",
		"not-a-time",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected release-time validation error")
	}
	if !strings.Contains(err.Error(), "--release-time requires --pre-order") {
		t.Fatalf("error = %v, want release-time pre-order validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestOneTimeProductOffersCreateBasicFlagsRejectsRedemptionLimitWithPreOrderBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"preorder",
		"--pre-order",
		"--start-time",
		"2026-06-01T00:00:00Z",
		"--end-time",
		"2026-07-01T00:00:00Z",
		"--release-time",
		"2026-08-01T00:00:00Z",
		"--price-change-behavior",
		"PRE_ORDER_PRICE_CHANGE_BEHAVIOR_NEW_ORDERS_ONLY",
		"--no-override",
		"US",
		"--redemption-limit",
		"0",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected redemption-limit validation error")
	}
	if !strings.Contains(err.Error(), "--redemption-limit cannot be used with --pre-order") {
		t.Fatalf("error = %v, want redemption-limit pre-order validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestOneTimeProductOffersCreateBasicFlagsRejectsPriceBehaviorWithoutPreOrderBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"intro",
		"--relative-discount",
		"US:0.5",
		"--price-change-behavior=",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected price-change-behavior validation error")
	}
	if !strings.Contains(err.Error(), "--price-change-behavior requires --pre-order") {
		t.Fatalf("error = %v, want price-change-behavior pre-order validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestOneTimeProductOffersCreateRejectsJSONWithBasicFlagsBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")
	bodyPath := filepath.Join(t.TempDir(), "offer.json")
	if err := os.WriteFile(bodyPath, []byte(`{
		"discountedOffer":{"startTime":"2026-06-01T00:00:00Z","endTime":"2026-07-01T00:00:00Z"},
		"regionalPricingAndAvailabilityConfigs":[{"regionCode":"US","availability":"AVAILABLE","relativeDiscount":0.5}]
	}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"intro",
		"--from-json",
		bodyPath,
		"--relative-discount",
		"US:0.5",
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

func TestOneTimeProductOffersCreateBasicFlagsRejectsInvalidTimeBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	for _, tc := range []struct {
		name string
		flag string
	}{
		{name: "start", flag: "--start-time"},
		{name: "end", flag: "--end-time"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := newRootCommand(&buf)
			cmd.SetArgs([]string{
				"one-time-product-offers",
				"create",
				"--package",
				"com.example.app",
				"--product-id",
				"coins_100",
				"--purchase-option-id",
				"buy",
				"--offer-id",
				"intro",
				"--relative-discount",
				"US:0.5",
				tc.flag,
				"not-a-time",
				"--regions-version",
				"2026/05",
				"--dry-run",
				"--output",
				"json",
			})

			err := cmd.Execute()
			if err == nil {
				t.Fatal("expected RFC3339 validation error")
			}
			if !strings.Contains(err.Error(), "must be RFC3339") {
				t.Fatalf("error = %v, want RFC3339 validation", err)
			}
			if strings.Contains(err.Error(), "no active auth profile") {
				t.Fatalf("error = %v, did not expect auth error", err)
			}
		})
	}
}

func TestOneTimeProductOffersCreateBasicFlagsRejectsInvalidRelativeDiscountBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"intro",
		"--relative-discount",
		"US:0",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected relative discount validation error")
	}
	if !strings.Contains(err.Error(), "relative discount must be greater than 0 and less than 1") {
		t.Fatalf("error = %v, want relative discount range validation", err)
	}
	if strings.Contains(err.Error(), "requires exactly one") {
		t.Fatalf("error = %v, did not expect downstream price mode error", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductOffersCreateBasicFlagsRejectsMalformedAbsoluteDiscountBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"intro",
		"--absolute-discount",
		"US:USD:x",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected absolute discount format validation error")
	}
	if !strings.Contains(err.Error(), "absolute discount must use REGION:CURRENCY:UNITS[:NANOS]") {
		t.Fatalf("error = %v, want absolute discount format validation", err)
	}
	if strings.Contains(err.Error(), "price units") {
		t.Fatalf("error = %v, did not expect generic price units error", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductOffersCreateRejectsInvalidBodyBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")
	bodyPath := filepath.Join(t.TempDir(), "offer.json")
	if err := os.WriteFile(bodyPath, []byte(`{"discountedOffer":{}}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"intro",
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
		t.Fatal("expected offer body validation error")
	}
	if !strings.Contains(err.Error(), "requires at least one regional config") {
		t.Fatalf("error = %v, want regional config validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductOffersBatchDeleteDryRunInfersParentsBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-delete",
		"--package",
		"com.example.app",
		"--offer",
		"coins_100/buy/intro",
		"--offer",
		"coins_100/rent/preorder",
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
		`"purchaseOptionId":"-"`,
		`"offerId":"intro"`,
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

func TestOneTimeProductOffersBatchDeleteRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-delete",
		"--package",
		"com.example.app",
		"--offer",
		"coins_100/buy/intro",
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

func TestOneTimeProductOffersBatchDeleteInfersOmittedPurchaseOptionBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-delete",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--offer",
		"coins_100/buy/intro",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, `"purchaseOptionId":"buy"`) {
		t.Fatalf("output = %s, want inferred purchase option", output)
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestOneTimeProductOffersBatchPatchAvailabilityDryRunInfersParentsBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-patch-availability",
		"--package",
		"com.example.app",
		"--availability",
		"coins_100/buy/intro/us:noLongerAvailable",
		"--availability",
		"coins_100/rent/winback/FR:available",
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
		`"purchaseOptionId":"-"`,
		`"offerId":"intro"`,
		`"regionCode":"US"`,
		`"availability":"noLongerAvailable"`,
		`"updateMask":"regionalPricingAndAvailabilityConfigs"`,
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

func TestOneTimeProductOffersBatchPatchAvailabilityRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-patch-availability",
		"--package",
		"com.example.app",
		"--availability",
		"coins_100/buy/intro/US:available",
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

func TestOneTimeProductOffersBatchPatchAvailabilityRejectsMalformedPatchBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-patch-availability",
		"--package",
		"com.example.app",
		"--availability",
		"coins_100/buy/intro:available",
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
	if !strings.Contains(err.Error(), "productId/purchaseOptionId/offerId/REGION:available|noLongerAvailable") {
		t.Fatalf("error = %v, want availability format validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductOffersBatchPatchRelativeDiscountsDryRunInfersParentsBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-patch-relative-discounts",
		"--package",
		"com.example.app",
		"--relative-discount",
		"coins_100/buy/intro/us:0.5",
		"--relative-discount",
		"coins_100/rent/winback/FR:0.25",
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
		`"purchaseOptionId":"-"`,
		`"offerId":"intro"`,
		`"regionCode":"US"`,
		`"relativeDiscount":0.5`,
		`"updateMask":"regionalPricingAndAvailabilityConfigs"`,
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

func TestOneTimeProductOffersBatchPatchRelativeDiscountsRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-patch-relative-discounts",
		"--package",
		"com.example.app",
		"--relative-discount",
		"coins_100/buy/intro/US:0.5",
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

func TestOneTimeProductOffersBatchPatchRelativeDiscountsRejectsMalformedPatchBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-patch-relative-discounts",
		"--package",
		"com.example.app",
		"--relative-discount",
		"coins_100/buy/intro:0.5",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected relative discount format validation error")
	}
	if !strings.Contains(err.Error(), "productId/purchaseOptionId/offerId/REGION:0.5") {
		t.Fatalf("error = %v, want relative discount format validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductOffersBatchPatchAbsoluteDiscountsDryRunInfersParentsBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-patch-absolute-discounts",
		"--package",
		"com.example.app",
		"--absolute-discount",
		"coins_100/buy/intro/us:USD:1:500000000",
		"--absolute-discount",
		"coins_100/rent/winback/FR:EUR:2",
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
		`"purchaseOptionId":"-"`,
		`"offerId":"intro"`,
		`"regionCode":"US"`,
		`"currencyCode":"USD"`,
		`"units":1`,
		`"nanos":500000000`,
		`"updateMask":"regionalPricingAndAvailabilityConfigs"`,
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

func TestOneTimeProductOffersBatchPatchAbsoluteDiscountsRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-patch-absolute-discounts",
		"--package",
		"com.example.app",
		"--absolute-discount",
		"coins_100/buy/intro/US:USD:1",
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

func TestOneTimeProductOffersBatchPatchAbsoluteDiscountsRejectsMalformedPatchBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-patch-absolute-discounts",
		"--package",
		"com.example.app",
		"--absolute-discount",
		"coins_100/buy/intro:USD:1",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected absolute discount format validation error")
	}
	if !strings.Contains(err.Error(), "productId/purchaseOptionId/offerId/REGION:CURRENCY:UNITS[:NANOS]") {
		t.Fatalf("error = %v, want absolute discount format validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductOffersBatchPatchAbsoluteDiscountsRejectsMalformedMoneyBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-patch-absolute-discounts",
		"--package",
		"com.example.app",
		"--absolute-discount",
		"coins_100/buy/intro/US:USD",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected absolute discount money format validation error")
	}
	if !strings.Contains(err.Error(), "one-time product offer absolute discount must use productId/purchaseOptionId/offerId/REGION:CURRENCY:UNITS[:NANOS]") {
		t.Fatalf("error = %v, want absolute discount format validation", err)
	}
	if strings.Contains(err.Error(), "purchase option price") {
		t.Fatalf("error = %v, did not expect purchase option price message", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestOneTimeProductOffersBatchPatchNoOverridesDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-patch-no-overrides",
		"--package",
		"com.example.app",
		"--no-override",
		"coins_100/buy/intro/us",
		"--no-override",
		"coins_100/rent/winback/FR",
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
		`"purchaseOptionId":"-"`,
		`"offerId":"intro"`,
		`"regionCode":"US"`,
		`"noOverride":true`,
		`"updateMask":"regionalPricingAndAvailabilityConfigs"`,
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

func TestOneTimeProductOffersBatchPatchNoOverridesRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-patch-no-overrides",
		"--package",
		"com.example.app",
		"--no-override",
		"coins_100/buy/intro/US",
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

func TestOneTimeProductOffersBatchDeactivateDryRunInfersParentsBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-deactivate",
		"--package",
		"com.example.app",
		"--offer",
		"coins_100/buy/intro",
		"--offer",
		"coins_100/rent/winback",
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
		`"productId":"coins_100"`,
		`"purchaseOptionId":"-"`,
		`"offerId":"intro"`,
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

func TestOneTimeProductOffersBatchActivateRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-activate",
		"--package",
		"com.example.app",
		"--offer",
		"coins_100/buy/intro",
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

func TestOneTimeProductOffersBatchCancelDryRunCallsOutPendingOrders(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"batch-cancel",
		"--package",
		"com.example.app",
		"--offer",
		"coins_100/buy/preorder",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "pending orders") {
		t.Fatalf("output = %s, want pending orders warning in plan", output)
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth error", output)
	}
}

func TestOneTimeProductOffersDeactivateDryRunPrintsPlanBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"deactivate",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"intro",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	var result play.OneTimeProductOfferStateUpdateResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Unmarshal() error = %v\n%s", err, buf.String())
	}
	if result.Action != play.OneTimeProductOfferStateActionDeactivate {
		t.Fatalf("Action = %q, want deactivate", result.Action)
	}
	if result.Applied {
		t.Fatal("Applied = true, want dry-run plan")
	}
}

func TestOneTimeProductOffersCancelRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"one-time-product-offers",
		"cancel",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--purchase-option-id",
		"buy",
		"--offer-id",
		"preorder",
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
