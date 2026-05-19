package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSubscriptionOffersListRejectsInvalidPageSizeBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"list",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
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

func TestSubscriptionOffersGetRejectsMissingOfferIDBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"get",
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
		t.Fatal("expected offer ID validation error")
	}
	if !strings.Contains(err.Error(), "subscription offer ID") {
		t.Fatalf("error = %v, want offer ID validation", err)
	}
}

func TestSubscriptionOffersDeactivateDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"deactivate",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro",
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
		`"offerId":"intro"`,
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

func TestSubscriptionOffersBatchDeactivateDryRunInfersParentsBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-deactivate",
		"--package",
		"com.example.app",
		"--offer",
		"premium/monthly/intro",
		"--offer",
		"premium/annual/winback",
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
		`"productId":"premium"`,
		`"basePlanId":"-"`,
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

func TestSubscriptionOffersBatchPatchAvailabilityDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-patch-availability",
		"--package",
		"com.example.app",
		"--availability",
		"premium/monthly/intro/us: false",
		"--availability",
		"premium/annual/winback/FR:true",
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
		`"basePlanId":"-"`,
		`"offerId":"intro"`,
		`"regionCode":"US"`,
		`"availability":false`,
		`"newSubscriberAvailability":false`,
		`"updateMask":"regionalConfigs"`,
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

func TestSubscriptionOffersBatchPatchAvailabilityRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-patch-availability",
		"--package",
		"com.example.app",
		"--availability",
		"premium/monthly/intro/US:true",
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

func TestSubscriptionOffersBatchPatchAvailabilityRejectsMalformedPatchBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-patch-availability",
		"--package",
		"com.example.app",
		"--availability",
		"premium/monthly/intro:true",
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
	if !strings.Contains(err.Error(), "productId/basePlanId/offerId/REGION:true|false") {
		t.Fatalf("error = %v, want availability format validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestSubscriptionOffersBatchPatchAvailabilityRejectsInvalidBooleanBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-patch-availability",
		"--package",
		"com.example.app",
		"--availability",
		"premium/monthly/intro/US:notabool",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected availability boolean validation error")
	}
	if !strings.Contains(err.Error(), "productId/basePlanId/offerId/REGION:true|false") {
		t.Fatalf("error = %v, want availability format validation", err)
	}
	if strings.Contains(err.Error(), "strconv.ParseBool") {
		t.Fatalf("error = %v, did not expect raw strconv error", err)
	}
}

func TestSubscriptionOffersBatchPatchPhaseRelativeDiscountsDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-patch-phase-relative-discounts",
		"--package",
		"com.example.app",
		"--relative-discount",
		"premium/monthly/intro/0/us:0.75",
		"--relative-discount",
		"premium/annual/winback/1/FR:0.5",
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
		`"basePlanId":"-"`,
		`"offerId":"intro"`,
		`"phaseIndex":0`,
		`"regionCode":"US"`,
		`"relativeDiscount":0.75`,
		`"updateMask":"phases"`,
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

func TestSubscriptionOffersBatchPatchPhaseRelativeDiscountsRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-patch-phase-relative-discounts",
		"--package",
		"com.example.app",
		"--relative-discount",
		"premium/monthly/intro/0/US:0.75",
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

func TestSubscriptionOffersBatchPatchPhaseRelativeDiscountsRejectsMalformedPatchBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-patch-phase-relative-discounts",
		"--package",
		"com.example.app",
		"--relative-discount",
		"premium/monthly/intro/US:0.75",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected phase relative discount format validation error")
	}
	if !strings.Contains(err.Error(), "productId/basePlanId/offerId/phaseIndex/REGION:0.75") {
		t.Fatalf("error = %v, want phase relative discount format validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestSubscriptionOffersBatchPatchPhaseAbsoluteDiscountsDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-patch-phase-absolute-discounts",
		"--package",
		"com.example.app",
		"--absolute-discount",
		"premium/monthly/intro/0/us:USD:1:500000000",
		"--absolute-discount",
		"premium/annual/winback/1/FR:EUR:2",
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
		`"basePlanId":"-"`,
		`"offerId":"intro"`,
		`"phaseIndex":0`,
		`"regionCode":"US"`,
		`"currencyCode":"USD"`,
		`"units":1`,
		`"nanos":500000000`,
		`"updateMask":"phases"`,
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

func TestSubscriptionOffersCreateDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")
	bodyPath := filepath.Join(t.TempDir(), "offer.json")
	if err := os.WriteFile(bodyPath, []byte(`{
		"state":"ACTIVE",
		"offerTags":[{"tag":"intro"}],
		"regionalConfigs":[
			{"regionCode":"US","newSubscriberAvailability":true},
			{"regionCode":"FR","newSubscriberAvailability":true}
		],
		"otherRegionsConfig":{"otherRegionsNewSubscriberAvailability":true},
		"phases":[{
			"duration":"P1M",
			"recurrenceCount":1,
			"regionalConfigs":[
				{"regionCode":"US","price":{"currencyCode":"USD","units":"1"}},
				{"regionCode":"FR","price":{"currencyCode":"EUR","nanos":990000000}}
			],
			"otherRegionsConfig":{"free":{}}
		}],
		"targeting":{"acquisitionRule":{"scope":{"thisSubscription":{}}}}
	}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
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

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"action":"create"`,
		`"dryRun":true`,
		`"created":false`,
		`"productId":"premium"`,
		`"basePlanId":"monthly"`,
		`"offerId":"intro"`,
		`"regionsVersion":"2026/05"`,
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
	if strings.Contains(output, `"state":"ACTIVE"`) {
		t.Fatalf("output = %s, did not expect output-only state from input JSON", output)
	}
}

func TestSubscriptionOffersCreateBasicFreePhaseDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro",
		"--offer-tag",
		"trial",
		"--free-region",
		"us",
		"--free-region",
		"FR",
		"--other-regions-free",
		"--targeting-acquisition-scope",
		"this-subscription",
		"--phase-duration",
		"P7D",
		"--phase-recurrence",
		"1",
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
		`"basePlanId":"monthly"`,
		`"offerId":"intro"`,
		`"offerTags":["trial"]`,
		`"otherRegionsConfig":{"newSubscriberAvailability":true}`,
		`"targeting":{"acquisition":{"scope":{"thisSubscription":true}}}`,
		`"regionCode":"US"`,
		`"regionCode":"FR"`,
		`"newSubscriberAvailability":true`,
		`"duration":"P7D"`,
		`"recurrenceCount":1`,
		`"otherRegionsConfig":{"free":true}`,
		`"free":true`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionOffersCreateRejectsInvalidAcquisitionTargetingScope(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro",
		"--free-region",
		"US",
		"--targeting-acquisition-scope",
		"specific-subscription-in-app",
		"--phase-duration",
		"P7D",
		"--regions-version",
		"2026/05",
		"--dry-run",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected acquisition targeting scope validation error")
	}
	if !strings.Contains(err.Error(), "acquisition targeting scope") {
		t.Fatalf("error = %v, want acquisition targeting scope validation", err)
	}
}

func TestSubscriptionOffersCreateBasicUpgradeTargetingDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"upgrade-intro",
		"--free-region",
		"US",
		"--targeting-upgrade-scope",
		"specific-subscription-in-app",
		"--targeting-upgrade-product-id",
		"basic",
		"--targeting-upgrade-billing-period",
		"P1M",
		"--targeting-upgrade-once-per-user",
		"--phase-duration",
		"P7D",
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
		`"offerId":"upgrade-intro"`,
		`"upgrade":{"scope":{"specificSubscriptionInApp":"basic"},"billingPeriodDuration":"P1M","oncePerUser":true}`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionOffersCreateRejectsMixedTargeting(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro",
		"--free-region",
		"US",
		"--targeting-acquisition-scope",
		"this-subscription",
		"--targeting-upgrade-scope",
		"this-subscription",
		"--phase-duration",
		"P7D",
		"--regions-version",
		"2026/05",
		"--dry-run",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected mixed targeting validation error")
	}
	if !strings.Contains(err.Error(), "cannot combine acquisition and upgrade targeting") {
		t.Fatalf("error = %v, want mixed targeting validation", err)
	}
}

func TestSubscriptionOffersCreateBasicPricePhaseDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro",
		"--offer-tag",
		"paid-intro",
		"--price",
		"us:USD:1:990000000",
		"--price",
		"FR:EUR:0:990000000",
		"--phase-duration",
		"P1M",
		"--phase-recurrence",
		"1",
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
		`"basePlanId":"monthly"`,
		`"offerId":"intro"`,
		`"offerTags":["paid-intro"]`,
		`"regionCode":"US"`,
		`"regionCode":"FR"`,
		`"newSubscriberAvailability":true`,
		`"duration":"P1M"`,
		`"recurrenceCount":1`,
		`"price":{"currencyCode":"USD","units":1,"nanos":990000000}`,
		`"price":{"currencyCode":"EUR","nanos":990000000}`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionOffersCreateBasicTwoPhaseDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro",
		"--free-region",
		"US",
		"--free-region",
		"FR",
		"--phase-duration",
		"P7D",
		"--other-regions-free",
		"--phase-2-price",
		"US:USD:1:990000000",
		"--phase-2-price",
		"FR:EUR:1:990000000",
		"--phase-2-duration",
		"P1M",
		"--phase-2-recurrence",
		"2",
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
		`"otherRegionsConfig":{"newSubscriberAvailability":true}`,
		`"duration":"P7D"`,
		`"free":true`,
		`"duration":"P1M"`,
		`"recurrenceCount":2`,
		`"price":{"currencyCode":"USD","units":1,"nanos":990000000}`,
		`"price":{"currencyCode":"EUR","units":1,"nanos":990000000}`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if count := strings.Count(output, `"otherRegionsConfig":{"free":true}`); count != 2 {
		t.Fatalf("output = %s, want two free phase other-regions configs, got %d", output, count)
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionOffersCreateBasicTwoPhasePaidOtherRegionsDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro-paid",
		"--free-region",
		"US",
		"--phase-duration",
		"P7D",
		"--other-regions-usd-price",
		"USD:1",
		"--other-regions-eur-price",
		"EUR:1",
		"--phase-2-price",
		"US:USD:1:990000000",
		"--phase-2-duration",
		"P1M",
		"--phase-2-other-regions-usd-price",
		"USD:1:990000000",
		"--phase-2-other-regions-eur-price",
		"EUR:1:990000000",
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
		`"otherRegionsConfig":{"newSubscriberAvailability":true}`,
		`"otherRegionsPrices":{"usdPrice":{"currencyCode":"USD","units":1},"eurPrice":{"currencyCode":"EUR","units":1}}`,
		`"otherRegionsPrices":{"usdPrice":{"currencyCode":"USD","units":1,"nanos":990000000},"eurPrice":{"currencyCode":"EUR","units":1,"nanos":990000000}}`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionOffersCreateBasicRelativeOtherRegionsDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro-relative",
		"--relative-discount",
		"US:0.5",
		"--other-regions-relative-discount",
		"0.5",
		"--phase-duration",
		"P1M",
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
		`"otherRegionsConfig":{"newSubscriberAvailability":true}`,
		`"otherRegionsConfig":{"relativeDiscount":0.5}`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionOffersCreateBasicAbsoluteOtherRegionsDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro-absolute",
		"--absolute-discount",
		"US:USD:1:990000000",
		"--other-regions-absolute-usd-discount",
		"USD:1:990000000",
		"--other-regions-absolute-eur-discount",
		"EUR:1:990000000",
		"--phase-duration",
		"P1M",
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
		`"otherRegionsConfig":{"newSubscriberAvailability":true}`,
		`"absoluteDiscounts":{"usdPrice":{"currencyCode":"USD","units":1,"nanos":990000000},"eurPrice":{"currencyCode":"EUR","units":1,"nanos":990000000}}`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionOffersCreateRejectsIncompletePaidOtherRegions(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro",
		"--free-region",
		"US",
		"--phase-duration",
		"P7D",
		"--other-regions-usd-price",
		"USD:1",
		"--regions-version",
		"2026/05",
		"--dry-run",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected paid other-regions validation error")
	}
	if !strings.Contains(err.Error(), "other-regions prices require USD and EUR prices") {
		t.Fatalf("error = %v, want paid other-regions validation", err)
	}
}

func TestSubscriptionOffersCreateRejectsMissingSecondPhasePaidOtherRegions(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro",
		"--free-region",
		"US",
		"--phase-duration",
		"P7D",
		"--other-regions-usd-price",
		"USD:1",
		"--other-regions-eur-price",
		"EUR:1",
		"--phase-2-price",
		"US:USD:1:990000000",
		"--phase-2-duration",
		"P1M",
		"--regions-version",
		"2026/05",
		"--dry-run",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected second phase paid other-regions validation error")
	}
	if !strings.Contains(err.Error(), "second phase other-regions price mode is required") {
		t.Fatalf("error = %v, want second phase paid other-regions validation", err)
	}
}

func TestSubscriptionOffersCreateRejectsMixedOtherRegionsPriceModes(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro",
		"--price",
		"US:USD:1:990000000",
		"--phase-duration",
		"P1M",
		"--other-regions-usd-price",
		"USD:1",
		"--other-regions-eur-price",
		"EUR:1",
		"--other-regions-relative-discount",
		"0.5",
		"--regions-version",
		"2026/05",
		"--dry-run",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected mixed other-regions price mode validation error")
	}
	if !strings.Contains(err.Error(), "requires exactly one of prices, relative discount, or absolute discounts") {
		t.Fatalf("error = %v, want mixed other-regions price mode validation", err)
	}
}

func TestSubscriptionOffersCreateRejectsNaNOtherRegionsRelativeDiscount(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro",
		"--relative-discount",
		"US:0.5",
		"--phase-duration",
		"P1M",
		"--other-regions-relative-discount",
		"NaN",
		"--regions-version",
		"2026/05",
		"--dry-run",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected NaN other-regions relative discount validation error")
	}
	if !strings.Contains(err.Error(), "other-regions relative discount must be greater than 0 and less than 1") {
		t.Fatalf("error = %v, want NaN other-regions relative discount validation", err)
	}
}

func TestSubscriptionOffersCreateRejectsFreeAndPaidOtherRegions(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro",
		"--free-region",
		"US",
		"--phase-duration",
		"P7D",
		"--other-regions-free",
		"--other-regions-usd-price",
		"USD:1",
		"--other-regions-eur-price",
		"EUR:1",
		"--regions-version",
		"2026/05",
		"--dry-run",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected free and paid other-regions validation error")
	}
	if !strings.Contains(err.Error(), "--other-regions-free cannot be combined with other-regions price mode flags") {
		t.Fatalf("error = %v, want free and paid other-regions validation", err)
	}
}

func TestSubscriptionOffersCreateRejectsTwoPhaseRegionMismatch(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro",
		"--free-region",
		"US",
		"--phase-duration",
		"P7D",
		"--phase-2-price",
		"FR:EUR:1:990000000",
		"--phase-2-duration",
		"P1M",
		"--regions-version",
		"2026/05",
		"--dry-run",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected second phase region validation error")
	}
	if !strings.Contains(err.Error(), "second phase region FR is not configured in the first phase") {
		t.Fatalf("error = %v, want second phase region validation", err)
	}
}

func TestSubscriptionOffersCreateBasicRelativeDiscountPhaseDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro",
		"--offer-tag",
		"half-off",
		"--relative-discount",
		"us:0.5",
		"--relative-discount",
		"FR:0.25",
		"--phase-duration",
		"P1M",
		"--phase-recurrence",
		"1",
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
		`"basePlanId":"monthly"`,
		`"offerId":"intro"`,
		`"offerTags":["half-off"]`,
		`"regionCode":"US"`,
		`"regionCode":"FR"`,
		`"newSubscriberAvailability":true`,
		`"duration":"P1M"`,
		`"recurrenceCount":1`,
		`"relativeDiscount":0.5`,
		`"relativeDiscount":0.25`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionOffersCreateBasicAbsoluteDiscountPhaseDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro",
		"--offer-tag",
		"absolute-intro",
		"--absolute-discount",
		"us:USD:1:990000000",
		"--absolute-discount",
		"FR:EUR:0:990000000",
		"--phase-duration",
		"P1M",
		"--phase-recurrence",
		"1",
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
		`"basePlanId":"monthly"`,
		`"offerId":"intro"`,
		`"offerTags":["absolute-intro"]`,
		`"regionCode":"US"`,
		`"regionCode":"FR"`,
		`"newSubscriberAvailability":true`,
		`"duration":"P1M"`,
		`"recurrenceCount":1`,
		`"absoluteDiscount":{"currencyCode":"USD","units":1,"nanos":990000000}`,
		`"absoluteDiscount":{"currencyCode":"EUR","nanos":990000000}`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSubscriptionOffersCreateBasicFlagsRejectDuplicatePhaseRegionBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro",
		"--free-region",
		"US",
		"--price",
		"us:USD:1",
		"--phase-duration",
		"P1M",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected duplicate region validation error")
	}
	if !strings.Contains(err.Error(), "subscription offer create region US is duplicated") {
		t.Fatalf("error = %v, want duplicate region validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestSubscriptionOffersCreateBasicFlagsRejectInvalidRelativeDiscountBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro",
		"--relative-discount",
		"US:not-a-number",
		"--phase-duration",
		"P1M",
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
	if !strings.Contains(err.Error(), "subscription offer create relative discount must use REGION:0.5") {
		t.Fatalf("error = %v, want relative discount format validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestSubscriptionOffersCreateBasicFlagsRejectMalformedAbsoluteDiscountBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro",
		"--absolute-discount",
		"US:USD:not-a-number",
		"--phase-duration",
		"P1M",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected absolute discount validation error")
	}
	if !strings.Contains(err.Error(), "subscription offer create absolute discount must use REGION:CURRENCY:UNITS[:NANOS]") {
		t.Fatalf("error = %v, want absolute discount format validation", err)
	}
	if strings.Contains(err.Error(), "price units") {
		t.Fatalf("error = %v, did not expect generic money parse error", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestSubscriptionOffersCreateRejectsJSONWithBasicFlagsBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")
	bodyPath := filepath.Join(t.TempDir(), "offer.json")
	if err := os.WriteFile(bodyPath, []byte(`{
		"regionalConfigs":[{"regionCode":"US","newSubscriberAvailability":true}],
		"phases":[{"duration":"P7D","recurrenceCount":1,"regionalConfigs":[{"regionCode":"US","free":{}}]}]
	}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro",
		"--from-json",
		bodyPath,
		"--free-region",
		"US",
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

func TestSubscriptionOffersCreateRejectsInvalidBodyBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")
	bodyPath := filepath.Join(t.TempDir(), "offer.json")
	if err := os.WriteFile(bodyPath, []byte(`{"regionalConfigs":[{"regionCode":"US","newSubscriberAvailability":true}]}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"create",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
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
	if !strings.Contains(err.Error(), "requires one or two phases") {
		t.Fatalf("error = %v, want phase validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestSubscriptionOffersBatchPatchPhaseAbsoluteDiscountsRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-patch-phase-absolute-discounts",
		"--package",
		"com.example.app",
		"--absolute-discount",
		"premium/monthly/intro/0/US:USD:1",
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

func TestSubscriptionOffersBatchPatchPhaseAbsoluteDiscountsRejectsMalformedPatchBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-patch-phase-absolute-discounts",
		"--package",
		"com.example.app",
		"--absolute-discount",
		"premium/monthly/intro/US:USD:1",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected phase absolute discount format validation error")
	}
	if !strings.Contains(err.Error(), "productId/basePlanId/offerId/phaseIndex/REGION:CURRENCY:UNITS[:NANOS]") {
		t.Fatalf("error = %v, want phase absolute discount format validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestSubscriptionOffersBatchPatchPhasePricesDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-patch-phase-prices",
		"--package",
		"com.example.app",
		"--price",
		"premium/monthly/intro/0/us:USD:1:990000000",
		"--price",
		"premium/annual/winback/1/FR:EUR:2",
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
		`"basePlanId":"-"`,
		`"offerId":"intro"`,
		`"phaseIndex":0`,
		`"regionCode":"US"`,
		`"currencyCode":"USD"`,
		`"units":1`,
		`"nanos":990000000`,
		`"updateMask":"phases"`,
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

func TestSubscriptionOffersBatchPatchPhasePricesRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-patch-phase-prices",
		"--package",
		"com.example.app",
		"--price",
		"premium/monthly/intro/0/US:USD:1",
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

func TestSubscriptionOffersBatchPatchPhasePricesRejectsMalformedPatchBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-patch-phase-prices",
		"--package",
		"com.example.app",
		"--price",
		"premium/monthly/intro/US:USD:1",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected phase price format validation error")
	}
	if !strings.Contains(err.Error(), "productId/basePlanId/offerId/phaseIndex/REGION:CURRENCY:UNITS[:NANOS]") {
		t.Fatalf("error = %v, want phase price format validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestSubscriptionOffersBatchPatchPhaseFreeDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-patch-phase-free",
		"--package",
		"com.example.app",
		"--free",
		"premium/monthly/intro/0/us",
		"--free",
		"premium/annual/winback/1/FR",
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
		`"basePlanId":"-"`,
		`"offerId":"intro"`,
		`"phaseIndex":0`,
		`"regionCode":"US"`,
		`"free":true`,
		`"updateMask":"phases"`,
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

func TestSubscriptionOffersBatchPatchPhaseFreeRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-patch-phase-free",
		"--package",
		"com.example.app",
		"--free",
		"premium/monthly/intro/0/US",
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

func TestSubscriptionOffersBatchPatchPhaseFreeRejectsMalformedPatchBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-patch-phase-free",
		"--package",
		"com.example.app",
		"--free",
		"premium/monthly/intro/US",
		"--regions-version",
		"2026/05",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected phase free format validation error")
	}
	if !strings.Contains(err.Error(), "productId/basePlanId/offerId/phaseIndex/REGION") {
		t.Fatalf("error = %v, want phase free format validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestSubscriptionOffersBatchActivateRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-activate",
		"--package",
		"com.example.app",
		"--offer",
		"premium/monthly/intro",
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

func TestSubscriptionOffersBatchActivateRejectsMissingOfferBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-activate",
		"--package",
		"com.example.app",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected missing offer validation error")
	}
	if !strings.Contains(err.Error(), "at least one subscription offer is required") {
		t.Fatalf("error = %v, want missing offer validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestSubscriptionOffersDeleteDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"delete",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro",
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
		`"offerId":"intro"`,
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

func TestSubscriptionOffersDeleteRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"delete",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro",
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

func TestSubscriptionOffersRequiresConfirmOrDryRunBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"activate",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer-id",
		"intro",
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

func TestSubscriptionOffersListAcceptsWildcardsBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"list",
		"--package",
		"com.example.app",
		"--product-id",
		"-",
		"--base-plan-id",
		"-",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected auth error after wildcard validation succeeds")
	}
	if strings.Contains(err.Error(), "invalid subscription product ID") || strings.Contains(err.Error(), "base plan") {
		t.Fatalf("error = %v, want auth error after wildcard validation", err)
	}
}

func TestSubscriptionOffersGetRejectsWildcardBasePlanBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"get",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"-",
		"--offer-id",
		"intro",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected base plan validation error")
	}
	if !strings.Contains(err.Error(), "subscription base plan ID") {
		t.Fatalf("error = %v, want base plan validation", err)
	}
}

func TestSubscriptionOffersBatchGetRejectsMissingOfferBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-get",
		"--package",
		"com.example.app",
		"--product-id",
		"-",
		"--base-plan-id",
		"-",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected offer validation error")
	}
	if !strings.Contains(err.Error(), "at least one subscription offer") {
		t.Fatalf("error = %v, want missing offer validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestSubscriptionOffersBatchGetRejectsParentMismatchBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-get",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--offer",
		"other/monthly/intro",
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

func TestSubscriptionOffersBatchGetRejectsInvalidOfferIDBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-get",
		"--package",
		"com.example.app",
		"--product-id",
		"-",
		"--base-plan-id",
		"-",
		"--offer",
		"premium/monthly/Intro",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected offer ID validation error")
	}
	if !strings.Contains(err.Error(), "subscription offer ID") {
		t.Fatalf("error = %v, want offer ID validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}

func TestSubscriptionOffersBatchGetRejectsOverlongOfferIDBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"batch-get",
		"--package",
		"com.example.app",
		"--product-id",
		"-",
		"--base-plan-id",
		"-",
		"--offer",
		"premium/monthly/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected offer ID validation error")
	}
	if !strings.Contains(err.Error(), "cannot exceed 63 characters") {
		t.Fatalf("error = %v, want offer ID length validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth", err)
	}
}
