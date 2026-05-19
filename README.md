# Google Play Console CLI

`gpc` is a fast, scriptable CLI for the Google Play Developer API.

The goal is the Android-side sibling to `asc`: predictable commands, JSON-friendly output, CI-first behavior, and no interactive prompts unless a command explicitly opts in.

## Quick Start

```sh
gpc version
gpc --help
```

From this repo:

```sh
make build
make install PREFIX="$HOME/.local"
```

From a Homebrew tap after a tagged release:

```sh
brew tap aljrico/tap
brew install gpc
```

Or with the install script on macOS/Linux:

```sh
curl -fsSL https://raw.githubusercontent.com/aljrico/Google-Play-Console-CLI/main/scripts/install.sh | sh
```

### Authenticate

```sh
gpc auth login \
  --name "MyApp" \
  --service-account /path/to/service-account.json

gpc auth status
gpc auth doctor
gpc init --dry-run
```

The service account needs access to the target app in Play Console. Android Publisher commands require the Google Play Android Developer API; `vitals` also requires the Play Developer Reporting API and app-level "View app information (read-only)" access.

### Command Shape

```sh
gpc tracks list --package com.example.app
gpc capabilities --status tested
gpc docs parity --output markdown
gpc docs commands --output markdown
gpc schema --resource edits.tracks --method list --pretty
gpc diff json ./metadata.old.json ./metadata.new.json --fail-on-change
gpc workflow list
gpc workflow run release-internal --dry-run
gpc device-tier-configs list --package com.example.app --page-size 25
gpc device-tier-configs get --package com.example.app --id 7
gpc testers get --package com.example.app --track internal
gpc testers update --package com.example.app --track internal --google-group qa@example.com --dry-run
gpc status --package com.example.app
gpc validate --package com.example.app
gpc releases list --package com.example.app --track internal
gpc releases upload --package com.example.app --track internal --aab ./app-release.aab --release-note "en-US=Bug fixes." --dry-run
gpc releases upload --package com.example.app --track internal --apk ./app-release.apk --dry-run
gpc releases promote --package com.example.app --from internal --to production --version-code 123 --status draft --dry-run
gpc releases halt --package com.example.app --track production --version-code 123 --dry-run
gpc releases resume --package com.example.app --track production --version-code 123 --status inProgress --user-fraction 0.25 --dry-run
gpc internal-sharing upload --package com.example.app --aab ./app-release.aab --dry-run
gpc app-recovery list --package com.example.app --version-code 123
gpc app-recovery deploy --package com.example.app --id 7 --dry-run
gpc app-recovery cancel --package com.example.app --id 7 --dry-run
gpc generated-apks list --package com.example.app --version-code 123
gpc generated-apks download --package com.example.app --version-code 123 --download-id split-download --file ./split.apk --dry-run
gpc system-apks variants list --package com.example.app --version-code 123
gpc images list --package com.example.app --language en-US --type phoneScreenshots
gpc images upload --package com.example.app --language en-US --type featureGraphic --file ./feature.png --dry-run
gpc images delete --package com.example.app --language en-US --type phoneScreenshots --image-id image-1 --dry-run
gpc images delete-all --package com.example.app --language en-US --type phoneScreenshots --dry-run
gpc listings update --package com.example.app --language en-US --title "Example" --dry-run
gpc listings delete --package com.example.app --language en-US --dry-run
gpc details update --package com.example.app --contact-email support@example.com --dry-run
gpc data-safety update --package com.example.app --csv ./data-safety.csv --dry-run
gpc reviews list --package com.example.app --max-results 25
gpc reviews get --package com.example.app --review-id review-123
gpc reviews reply --package com.example.app --review-id review-123 --text "Thanks for the feedback." --dry-run
gpc vitals metric-set get --package com.example.app --metric-set crash-rate
gpc vitals metric-set query --package com.example.app --metric-set crash-rate --metric crashRate --dimension versionCode --aggregation DAILY --start-date 2026-05-01 --end-date 2026-05-19
gpc vitals errors issues search --package com.example.app --filter "errorIssueType = CRASH" --start-date 2026-05-01 --end-date 2026-05-19 --order-by "errorReportCount desc"
gpc vitals errors reports search --package com.example.app --filter "errorIssueId = issue-123" --start-date 2026-05-01 --end-date 2026-05-19 --time-zone UTC
gpc vitals anomalies list --package com.example.app --filter 'activeBetween("2026-05-01T00:00:00Z", "2026-05-19T00:00:00Z")'
gpc in-app-products list --package com.example.app
gpc in-app-products get --package com.example.app --sku coins_100
gpc one-time-products list --package com.example.app --page-size 50
gpc one-time-products get --package com.example.app --product-id coins_100
gpc one-time-product-offers list --package com.example.app --product-id coins_100 --purchase-option-id buy
gpc one-time-product-offers get --package com.example.app --product-id coins_100 --purchase-option-id buy --offer-id intro
gpc subscriptions list --package com.example.app --page-size 50
gpc subscriptions get --package com.example.app --product-id premium_monthly
gpc subscription-offers list --package com.example.app --product-id premium_monthly --base-plan-id monthly
gpc subscription-offers get --package com.example.app --product-id premium_monthly --base-plan-id monthly --offer-id intro
gpc purchases product --package com.example.app --token PURCHASE_TOKEN
gpc purchases product acknowledge --package com.example.app --product-id coins_100 --token PURCHASE_TOKEN --dry-run
gpc purchases product consume --package com.example.app --product-id coins_100 --token PURCHASE_TOKEN --dry-run
gpc purchases subscription --package com.example.app --token PURCHASE_TOKEN
gpc purchases subscription revoke --package com.example.app --token PURCHASE_TOKEN --refund full --dry-run
gpc purchases voided list --package com.example.app --max-results 25
gpc orders get --package com.example.app --order-id GPA.1234-5678-9012-34567
gpc orders batch-get --package com.example.app --order-id GPA.1234 --order-id GPA.5678
gpc orders refund --package com.example.app --order-id GPA.1234-5678-9012-34567 --dry-run
gpc orders refund --package com.example.app --order-id GPA.1234-5678-9012-34567 --revoke --confirm
gpc pricing convert-region-prices --package com.example.app --currency USD --units 9 --nanos 990000000
gpc users list --developer 1234567890
gpc users create --developer 1234567890 --user-email user@example.com --permission CAN_VIEW_NON_FINANCIAL_DATA_GLOBAL --dry-run
gpc users patch --developer 1234567890 --user-email user@example.com --permission CAN_REPLY_TO_REVIEWS_GLOBAL --dry-run
gpc users delete --developer 1234567890 --user-email user@example.com --dry-run
gpc grants create --developer 1234567890 --user-email user@example.com --package com.example.app --permission CAN_VIEW_NON_FINANCIAL_DATA --dry-run
```

Review APIs follow Google Play's limits: list responses are recent reviews with comments, reply text is capped at 350 characters, and live replies require a service account with review-reply access.

`vitals` uses the Play Developer Reporting API, which is separate from Android Publisher and requires the Play Developer Reporting API to be enabled plus app-level "View app information (read-only)" access. Query and search commands accept explicit metrics, filters, and date ranges so automation never relies on API defaults; error issue and report search intervals use UTC when a timezone is set, and omitting it uses the API's UTC default.

`in-app-products` uses Google's legacy `inappproducts` API. Use it for managed products and catalog inspection; `one-time-products`, `subscriptions`, and `subscription-offers` use the newer monetization resources.

### First Publish Flow

```sh
gpc publish internal \
  --package com.example.app \
  --aab ./app-release.aab \
  --release-name "1.2.3" \
  --dry-run \
  --output json \
  --pretty
```

Remove `--dry-run` to validate the edit against Google Play. Add `--confirm` to commit the edit after validation.

## Output

`gpc` defaults to JSON so scripts get stable output. Override it with:

```sh
gpc version --output json --pretty
gpc version --output table
gpc version --output markdown
```

## Releases

Tagged releases are built with GoReleaser v2.15.4 for macOS, Linux, and Windows on `amd64` and `arm64`. The release workflow publishes archives, checksums, and a Homebrew formula from `.goreleaser.yaml`.

```sh
make release-check
make snapshot
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0
```

Publishing the Homebrew formula requires `HOMEBREW_TAP_GITHUB_TOKEN` with write access to `aljrico/homebrew-tap`.

Pull requests run a macOS packaging check that validates the GoReleaser config, builds a snapshot release, tests the checksum-verifying install script, audits the generated formula, installs it locally, runs the formula test, and smokes `gpc version`.

The install script supports `GPC_VERSION`, `GPC_INSTALL_DIR`, `GPC_REPO`, and `GPC_BASE_URL`, and verifies release archives against `checksums.txt` before installing.

## Status

Early but functional. Auth/profile storage, the command taxonomy, generated command/schema docs, core release workflows, localized listings, app-level details, review reading/replies, read-only monetization catalog commands, purchase checks, order lookup, and guarded order refunds are in place.

See [docs/PARITY.md](docs/PARITY.md) for the working parity map against App Store Connect CLI.
