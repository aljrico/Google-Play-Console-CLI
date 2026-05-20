# Google Play Console CLI

[![Release](https://img.shields.io/github/v/release/aljrico/Google-Play-Console-CLI?logo=github)](https://github.com/aljrico/Google-Play-Console-CLI/releases)
[![Build](https://github.com/aljrico/Google-Play-Console-CLI/actions/workflows/release.yml/badge.svg)](https://github.com/aljrico/Google-Play-Console-CLI/actions/workflows/release.yml)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

`gpc` is a Go CLI built from years of mobile games release and revenue-ops pain. Less Play Console clicking; stable JSON for scripts; mutations guarded by `--dry-run` and `--confirm`; commands that survive CI.

Coverage spans the Google Play Developer API surfaces mobile teams operate repeatedly: releases and tracks, store metadata, reviews, monetization catalogs (legacy IAP and the new one-time/subscription APIs), purchases and orders, financial and statistics reports, vitals, app recovery, and real-time developer notifications.

Heavily inspired by [`rorkai/App-Store-Connect-CLI`](https://github.com/rorkai/App-Store-Connect-CLI), which is the north star for the kind of fast, scriptable workflow this CLI tries to make pleasant on the Google Play side.

## Install

```sh
brew install aljrico/tap/gpc
```

Or with the install script (macOS/Linux):

```sh
curl -fsSL https://raw.githubusercontent.com/aljrico/Google-Play-Console-CLI/main/scripts/install.sh | sh
```

Or from source:

```sh
make build
make install PREFIX="$HOME/.local"
```

## Authenticate

```sh
gpc auth login --name "MyApp" --service-account /path/to/service-account.json
gpc auth doctor
gpc account status
```

The service account needs access to the target app in Play Console. `vitals` commands additionally require the Play Developer Reporting API to be enabled and app-level "View app information (read-only)" access on the service account.

## What It Does

A representative tour, grouped by capability. The full command reference is generated into [`docs/COMMANDS.md`](docs/COMMANDS.md); the parity map vs. App Store Connect CLI lives in [`docs/PARITY.md`](docs/PARITY.md).

Remove `--dry-run` to validate a planned mutation against Google Play; add `--confirm` to commit.

**Releases and tracks**

```sh
gpc releases upload --package com.example.app --track internal --aab ./app-release.aab --release-note "en-US=Bug fixes." --dry-run
gpc releases promote --package com.example.app --from internal --to production --version-code 123 --status draft --dry-run
gpc releases resume --package com.example.app --track production --version-code 123 --status inProgress --user-fraction 0.25 --dry-run
gpc status --package com.example.app
gpc validate --package com.example.app
```

**Store metadata**

```sh
gpc metadata apply --package com.example.app --file ./play-metadata.json --dry-run
gpc listings update --package com.example.app --language en-US --title "Example" --dry-run
gpc images upload --package com.example.app --language en-US --type featureGraphic --file ./feature.png --dry-run
gpc details update --package com.example.app --contact-email support@example.com --dry-run
```

**Reviews**

```sh
gpc reviews list --package com.example.app --max-results 25
gpc reviews reply --package com.example.app --review-id review-123 --text "Thanks for the feedback." --dry-run
```

**Monetization catalogs**

```sh
gpc in-app-products list --package com.example.app
gpc one-time-products create --package com.example.app --product-id coins_100 --from-json one-time-product.json --regions-version 2026/05 --dry-run
gpc subscriptions create --package com.example.app --product-id premium_monthly --listing 'en-US,Premium,Full access' --base-plan-id monthly --billing-period P1M --price US:USD:4:990000000 --regions-version 2026/05 --dry-run
gpc subscription-offers create --package com.example.app --product-id premium_monthly --base-plan-id monthly --offer-id intro --free-region US --phase-duration P7D --regions-version 2026/05 --dry-run
gpc subscription-offers batch-patch-phase-prices --package com.example.app --price premium_monthly/monthly/intro/0/US:USD:1:990000000 --regions-version 2026/05 --dry-run
```

**Purchases, orders, and pricing**

```sh
gpc purchases subscription --package com.example.app --token PURCHASE_TOKEN
gpc orders refund --package com.example.app --order-id GPA.1234-5678-9012-34567 --dry-run
gpc pricing convert-region-prices --package com.example.app --currency USD --units 9 --nanos 990000000
gpc pricing build-price-patches --from-json conversion.json --target subscription-base-plan --product-id premium_monthly --base-plan-id monthly
```

**Reports and analytics**

```sh
gpc finance reports download --bucket pubsite_prod_rev_0123456789 --object earnings/earnings_202605.zip --file ./earnings_202605.zip --dry-run
gpc finance reports summarize --file ./earnings_202605.csv
gpc analytics stats summarize --file ./store_performance.csv
gpc insights reports summarize --finance-file ./earnings_202605.csv --stats-file ./store_performance.csv
```

**Vitals and quality**

```sh
gpc vitals metric-set query --package com.example.app --metric-set crash-rate --metric crashRate --dimension versionCode --aggregation DAILY --start-date 2026-05-01 --end-date 2026-05-19
gpc vitals errors issues search --package com.example.app --filter "errorIssueType = CRASH" --start-date 2026-05-01 --end-date 2026-05-19
gpc vitals anomalies list --package com.example.app --filter 'activeBetween("2026-05-01T00:00:00Z", "2026-05-19T00:00:00Z")'
```

**RTDN and notifications**

```sh
gpc notifications pubsub setup --project play-project --topic play-rtdn --subscription play-rtdn-sub --dry-run
gpc notifications pubsub pull --project play-project --subscription play-rtdn-sub --decode-rtdn
gpc notifications rtdn decode --file ./pubsub-rtdn.json
GPC_NOTIFY_WEBHOOK_URL="$SLACK_WEBHOOK_URL" gpc notify slack --title "Release" --message "Internal release staged" --dry-run
```

**Migration from fastlane supply**

```sh
gpc migrate supply convert --directory fastlane/metadata/android --pretty > play-metadata.json
gpc migrate supply changelogs --directory fastlane/metadata/android --version-code 42 --pretty
gpc migrate supply images --directory fastlane/metadata/android --language en-US --type phoneScreenshots --pretty
```

**Discovery and CI**

```sh
gpc capabilities --status tested
gpc schema --resource edits.tracks --method list --pretty
gpc diff json ./metadata.old.json ./metadata.new.json --fail-on-change
gpc workflow run release-internal --dry-run
```

## Metadata Files

`gpc metadata apply` accepts strict JSON: unknown fields and trailing JSON are rejected, languages are canonicalized as BCP-47 tags, and listing text is checked against Play's public limits before any API call.

```json
{
  "details": {
    "defaultLanguage": "en-US",
    "contactWebsite": "https://example.com/support",
    "contactEmail": "support@example.com",
    "contactPhone": "+15555555555"
  },
  "listings": [
    {
      "language": "en-US",
      "title": "Example",
      "shortDescription": "Short store listing copy.",
      "fullDescription": "Long store listing copy.",
      "video": "https://youtu.be/example"
    }
  ]
}
```

Supported listing fields and Play's limits: `title` (30 characters), `shortDescription` (80), `fullDescription` (4000), `video` either an empty string or a YouTube URL. `gpc migrate supply convert` emits this shape from fastlane supply text files.

## Notes On Specific Surfaces

**Legacy IAP vs the new monetization APIs.** `in-app-products` uses Google's legacy `inappproducts` API for managed products and catalog inspection: `create` builds managed products only (asking Google to auto-convert missing regional prices from the default), live patches and deletes reject legacy subscription SKUs, and batch deletes preflight every requested SKU and fail closed unless Google returns managed products for all of them. The newer `one-time-products`, `subscriptions`, and `subscription-offers` commands use Google's monetization resources.

**One-time products.** `one-time-products create` uses Google's patch endpoint with `allowMissing=true`, because that is the actual create surface. `create --from-json` accepts either Google Play API `OneTimeProduct` JSON or `gpc one-time-products get --output json` shape; the `--package` and `--product-id` flags override immutable IDs in the file. For simple buy products, `create` can also build the body from `--listing` and `--price` flags; advanced rent options, new-regions pricing, and compliance settings still belong in JSON. Purchase option batch deletion follows Google's rule that each request targets a different one-time product; use `--force` only when you also intend to delete associated offers.

**Subscriptions and offers.** `subscriptions create` accepts JSON or basic flags for one auto-renewing, prepaid, or installments base plan with `--billing-period`, regional `--price`, optional `--restricted-country`, and compliance flags (EEA withdrawal right type, tokenized digital asset declaration, regional reduced-tax tiers, US streaming tax type). `subscription-offers create` accepts JSON or basic flags for one or two phases across explicit regions (free, paid-price, relative discount, or absolute discount), including paid or free other-regions fallbacks and acquisition/upgrade targeting; Google Play only supports offers under auto-renewing base plans. Phase patch commands fetch current offers and preserve untouched phases and regions.

**Reports and insights.** `finance reports download` and `analytics stats download` fetch report objects from the Google Play reports Cloud Storage bucket. Use the bucket ID shown in Play Console (shaped like `pubsite_prod_rev_0123456789`) and pass the exact object path; finance reports are ZIP, statistics CSV. `insights reports summarize` composes the downloaded finance and statistics CSVs into one local JSON artifact with derived KPIs such as net revenue by report type/currency, store listing acquisitions, acquisition rate, and revenue per acquisition.

**Pricing patches.** `pricing convert-region-prices` calls Google Play's regional price conversion API for a source Money value. `pricing build-price-patches` turns that JSON output into deterministic patch arguments and a suggested dry-run command for in-app product regional prices, one-time product purchase option prices, subscription base-plan prices, or subscription offer phase prices.

**Reviews.** Google Play limits the review API to recent reviews with comments; reply text is capped at 350 characters and live replies require a service account with review-reply access.

**Vitals.** Uses the Play Developer Reporting API, which is separate from Android Publisher and requires the Reporting API to be enabled plus app-level "View app information (read-only)" access. Query and search commands accept explicit metrics, filters, and date ranges so automation never relies on API defaults; error issue and report search intervals use UTC when a timezone is set, and omitting it uses the API's UTC default.

**RTDN.** `notifications pubsub setup` creates the Google Cloud side of Play real-time developer notifications: a topic, a subscription, and the Pub/Sub Publisher IAM binding for `google-play-developer-notifications@system.gserviceaccount.com`. Selecting the app-level RTDN topic in Play Console is still an operator step. `notifications pubsub pull` reads from a pull subscription and can decode each message as a Google Play RTDN payload, acknowledging only after output succeeds and only when both `--ack` and `--confirm` are passed. `notifications rtdn decode` handles wrapped Pub/Sub push payloads or `--unwrapped` developer notification JSON delivered directly.

## Output

`gpc` defaults to JSON for stable script consumption. Override with `--output table` or `--output markdown`; pair JSON with `--pretty` for human reading.

## Status

All commands in [`docs/PARITY.md`](docs/PARITY.md) are covered by tests. Coverage is intentionally broad on the Play workflows mobile teams hit repeatedly: releases, listings, monetization depth across both legacy and new APIs, reviews, finance and statistics reports, vitals, and RTDN.

The `blocked` rows in the parity matrix reflect Android Publisher API limits, not implementation gaps — features like category mutation, video previews, pre-registration, and Play Console browser automation are not exposed by the public API.

## Prior Art

This project is not pretending to be the first Google Play CLI. [`tamtom/play-console-cli`](https://github.com/tamtom/play-console-cli) is an active Go CLI for Play Console automation, [`Vacxe/google-play-cli`](https://github.com/Vacxe/google-play-cli) covers Google Play publishing workflows, and [`matlink/gplaycli`](https://github.com/matlink/gplaycli) is older prior art focused on Play Store downloading rather than publisher operations.

The reason for this project is narrower and more opinionated: an ASC CLI-inspired command shape, stable JSON by default, mutation commands guarded by `--dry-run` and `--confirm`, generated command/capability docs, and broad coverage for monetization, vitals, financial reports, and RTDN — the workflows that matter most once an app is already shipping and revenue ops is the day job.

## Development

Tagged releases are built with GoReleaser for macOS, Linux, and Windows on `amd64` and `arm64`. Pushing a `vX.Y.Z` tag triggers the release workflow, which publishes archives, checksums, and the Homebrew formula to [`aljrico/homebrew-tap`](https://github.com/aljrico/homebrew-tap).

```sh
make release-check
make snapshot
git tag -a vX.Y.Z -m "Release vX.Y.Z"
git push origin vX.Y.Z
```

Publishing the Homebrew formula requires `HOMEBREW_TAP_GITHUB_TOKEN` with write access to `aljrico/homebrew-tap`. Pull requests run a macOS packaging check that validates the GoReleaser config, builds a snapshot release, tests the checksum-verifying install script, audits the generated formula, and installs it locally.

The install script supports `GPC_VERSION`, `GPC_INSTALL_DIR`, `GPC_REPO`, and `GPC_BASE_URL`, and verifies release archives against `checksums.txt` before installing.
