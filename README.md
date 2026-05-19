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
gpc account status
gpc init --dry-run
```

The service account needs access to the target app in Play Console. Android Publisher commands require the Google Play Android Developer API; `vitals` also requires the Play Developer Reporting API and app-level "View app information (read-only)" access.

### Command Shape

```sh
gpc tracks list --package com.example.app
gpc capabilities --status tested
gpc docs parity --output markdown
gpc docs commands --output markdown
gpc install-skills list
gpc install-skills --skill gpc-cli-usage --dry-run
gpc account status
gpc schema --resource edits.tracks --method list --pretty
gpc diff json ./metadata.old.json ./metadata.new.json --fail-on-change
gpc workflow list
gpc workflow run release-internal --dry-run
gpc migrate supply inspect --directory fastlane/metadata/android
gpc migrate supply convert --directory fastlane/metadata/android --pretty > play-metadata.json
gpc migrate supply changelogs --directory fastlane/metadata/android --version-code 42 --pretty
gpc migrate supply images --directory fastlane/metadata/android --language en-US --type phoneScreenshots --pretty
GPC_NOTIFY_WEBHOOK_URL="$WEBHOOK_URL" gpc notify send --message "Internal release staged" --dry-run
GPC_NOTIFY_WEBHOOK_URL="$DISCORD_WEBHOOK_URL" gpc notify discord --title "Release" --message "Internal release staged" --field track=internal --dry-run
GPC_NOTIFY_WEBHOOK_URL="$SLACK_WEBHOOK_URL" gpc notify slack --title "Release" --message "Internal release staged" --field track=internal --dry-run
gpc search "release upload" --limit 5
gpc snitch report --title "Confusing release output" --command "gpc releases list --package com.example.app"
gpc notifications pubsub setup --project play-project --topic play-rtdn --subscription play-rtdn-sub --dry-run
gpc notifications pubsub pull --project play-project --subscription play-rtdn-sub --decode-rtdn
gpc notifications rtdn decode --file ./pubsub-rtdn.json
gpc notifications rtdn decode --file ./unwrapped-rtdn.json --unwrapped
gpc insights anomalies summarize --file ./vitals-anomalies.json
gpc finance reports download --bucket pubsite_prod_rev_0123456789 --object earnings/earnings_202605.zip --file ./earnings_202605.zip --dry-run
gpc finance reports summarize --file ./earnings_202605.csv
gpc analytics stats download --bucket pubsite_prod_rev_0123456789 --object stats/store_performance/store_performance_com.example.app_202605_country.csv --file ./store_performance.csv --dry-run
gpc analytics stats summarize --file ./store_performance_com.example.app_202605_country.csv
gpc device-tier-configs list --package com.example.app --page-size 25
gpc device-tier-configs get --package com.example.app --id 7
gpc testers get --package com.example.app --track internal
gpc testers update --package com.example.app --track internal --google-group qa@example.com --dry-run
gpc status --package com.example.app
gpc validate --package com.example.app
gpc releases list --package com.example.app --track internal
gpc releases upload --package com.example.app --track internal --aab ./app-release.aab --release-note "en-US=Bug fixes." --dry-run
gpc releases upload --package com.example.app --track internal --apk ./app-release.apk --dry-run
gpc releases promote --package com.example.app --from internal --to production --version-code 123 --status draft --release-note "en-US=Production rollout." --dry-run
gpc releases halt --package com.example.app --track production --version-code 123 --dry-run
gpc releases resume --package com.example.app --track production --version-code 123 --status inProgress --user-fraction 0.25 --dry-run
gpc internal-sharing upload --package com.example.app --aab ./app-release.aab --dry-run
gpc app-recovery list --package com.example.app --version-code 123
gpc app-recovery create --package com.example.app --version-code 123 --region US --dry-run
gpc app-recovery add-targeting --package com.example.app --id 7 --region US --dry-run
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
gpc metadata apply --package com.example.app --file ./play-metadata.json --dry-run
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
gpc in-app-products batch-get --package com.example.app --sku coins_100 --sku coins_500
gpc in-app-products create --package com.example.app --sku coins_100 --default-language en-US --default-price USD:1990000 --title "100 coins" --description "A small coin pack." --dry-run
gpc in-app-products patch --package com.example.app --sku coins_100 --status inactive --dry-run
gpc in-app-products patch --package com.example.app --sku coins_100 --listing-language en-US --default-price USD:2990000 --title "100 coins" --description "A better coin pack." --dry-run
gpc in-app-products delete --package com.example.app --sku coins_100 --dry-run
gpc in-app-products batch-delete --package com.example.app --sku coins_100 --sku coins_500 --dry-run
gpc one-time-products list --package com.example.app --page-size 50
gpc one-time-products get --package com.example.app --product-id coins_100
gpc one-time-products batch-get --package com.example.app --product-id coins_100 --product-id coins_500
gpc one-time-products delete --package com.example.app --product-id coins_100 --dry-run
gpc one-time-products batch-delete --package com.example.app --product-id coins_100 --product-id coins_500 --dry-run
gpc one-time-products purchase-option batch-delete --package com.example.app --product-id - --purchase-option coins_100/buy --purchase-option coins_500/rent --dry-run
gpc one-time-products purchase-option deactivate --package com.example.app --product-id coins_100 --purchase-option-id buy --dry-run
gpc one-time-product-offers list --package com.example.app --product-id coins_100 --purchase-option-id buy
gpc one-time-product-offers get --package com.example.app --product-id coins_100 --purchase-option-id buy --offer-id intro
gpc one-time-product-offers batch-get --package com.example.app --product-id - --purchase-option-id - --offer coins_100/buy/intro --offer coins_500/buy/preorder
gpc one-time-product-offers batch-delete --package com.example.app --offer coins_100/buy/intro --offer coins_500/rent/preorder --dry-run
gpc one-time-product-offers batch-deactivate --package com.example.app --offer coins_100/buy/intro --offer coins_100/rent/winback --dry-run
gpc one-time-product-offers deactivate --package com.example.app --product-id coins_100 --purchase-option-id buy --offer-id intro --dry-run
gpc one-time-product-offers cancel --package com.example.app --product-id coins_100 --purchase-option-id buy --offer-id preorder --dry-run
gpc subscriptions list --package com.example.app --page-size 50
gpc subscriptions get --package com.example.app --product-id premium_monthly
gpc subscriptions batch-get --package com.example.app --product-id premium_monthly --product-id premium_yearly
gpc subscriptions patch --package com.example.app --product-id premium_monthly --listing-language en-US --title "Premium" --description "Full access" --regions-version 2022/02 --dry-run
gpc subscriptions delete --package com.example.app --product-id premium_draft --dry-run
gpc subscriptions base-plan deactivate --package com.example.app --product-id premium_monthly --base-plan-id monthly --dry-run
gpc subscription-offers list --package com.example.app --product-id premium_monthly --base-plan-id monthly
gpc subscription-offers get --package com.example.app --product-id premium_monthly --base-plan-id monthly --offer-id intro
gpc subscription-offers delete --package com.example.app --product-id premium_monthly --base-plan-id monthly --offer-id draft-intro --dry-run
gpc subscription-offers deactivate --package com.example.app --product-id premium_monthly --base-plan-id monthly --offer-id intro --dry-run
gpc subscription-offers batch-get --package com.example.app --product-id - --base-plan-id - --offer premium_monthly/monthly/intro --offer premium_yearly/annual/winback
gpc purchases product --package com.example.app --token PURCHASE_TOKEN
gpc purchases product acknowledge --package com.example.app --product-id coins_100 --token PURCHASE_TOKEN --dry-run
gpc purchases product consume --package com.example.app --product-id coins_100 --token PURCHASE_TOKEN --dry-run
gpc purchases subscription --package com.example.app --token PURCHASE_TOKEN
gpc purchases subscription acknowledge --package com.example.app --subscription-id premium_monthly --token PURCHASE_TOKEN --dry-run
gpc purchases subscription cancel --package com.example.app --token PURCHASE_TOKEN --cancellation-type userRequestedStopRenewals --dry-run
gpc purchases subscription revoke --package com.example.app --token PURCHASE_TOKEN --refund full --dry-run
gpc purchases subscription revoke --package com.example.app --token PURCHASE_TOKEN --refund item --refund-product-id premium_addon --dry-run
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

Metadata files accepted by `gpc metadata apply` are strict JSON: unknown fields and trailing JSON are rejected, languages are canonicalized as BCP-47 tags, and listing text is checked against Play's public limits before any API call.

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

Supported listing fields are `title` (30 characters), `shortDescription` (80), `fullDescription` (4000), and `video` as an empty string or a YouTube URL. `gpc migrate supply convert` emits this shape from fastlane supply text files. `gpc migrate supply changelogs` groups fastlane `changelogs/VERSION_CODE.txt` files into Play release-note payloads and raw `language=text` arguments for `--release-note`. `gpc migrate supply images` converts fastlane image directories into validated `gpc images upload` argument lists.

Review APIs follow Google Play's limits: list responses are recent reviews with comments, reply text is capped at 350 characters, and live replies require a service account with review-reply access.

`vitals` uses the Play Developer Reporting API, which is separate from Android Publisher and requires the Play Developer Reporting API to be enabled plus app-level "View app information (read-only)" access. Query and search commands accept explicit metrics, filters, and date ranges so automation never relies on API defaults; error issue and report search intervals use UTC when a timezone is set, and omitting it uses the API's UTC default.

`notifications rtdn decode` expects a wrapped Pub/Sub push JSON body with `message.data` containing the base64-encoded Google Play `DeveloperNotification`; add `--unwrapped` when Pub/Sub push delivery sends the developer notification JSON directly. Pass exactly one of `--file` or `--data`.

`notifications pubsub setup` creates the Google Cloud side of Play real-time developer notifications: a topic, a subscription, and the Pub/Sub Publisher IAM binding for `google-play-developer-notifications@system.gserviceaccount.com`. You still need to select the topic in Play Console and send a test notification there.

`notifications pubsub pull` reads messages from a pull subscription and can decode each message as a Google Play RTDN payload. It only acknowledges messages after output succeeds and only when both `--ack` and `--confirm` are passed.

`finance reports download` and `analytics stats download` fetch report objects from the Google Play reports Cloud Storage bucket. Use the bucket ID shown in Play Console, usually shaped like `pubsite_prod_rev_0123456789`, and pass the exact object path for the report you want. Financial reports are ZIP files; statistics reports are CSV files.

`in-app-products` uses Google's legacy `inappproducts` API. Use it for managed products and catalog inspection; `create` builds managed products only and asks Google to auto-convert missing regional prices from the default price, while live patches and deletes reject legacy subscription SKUs. Batch deletes preflight every requested SKU and fail closed unless Google returns managed products for all of them. Price patches also request regional auto-conversion. `one-time-products`, `subscriptions`, and `subscription-offers` use the newer monetization resources. One-time product purchase option batch deletion follows Google's current rule that each request targets a different one-time product; use `--force` only when you also intend to delete associated offers.

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

Early but functional. Auth/profile storage, the command taxonomy, generated command/schema docs, core release workflows, localized listings, app-level details, file-based metadata apply, fastlane supply metadata conversion, review reading/replies, monetization catalog inspection and guarded state updates, purchase checks, order lookup, and guarded order refunds are in place.

See [docs/PARITY.md](docs/PARITY.md) for the working parity map against App Store Connect CLI.
See [docs/COMMANDS.md](docs/COMMANDS.md) for the generated command reference.
