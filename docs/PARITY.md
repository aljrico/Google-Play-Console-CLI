# Parity Matrix

This matrix maps the `asc` command families from App Store Connect CLI to the closest `gpc` command shape for Google Play Console automation.

Status values:

- `planned`: the feature should exist, but is not implemented yet.
- `implemented`: the command works.
- `tested`: implemented with automated coverage.
- `documented`: implemented and covered in user docs.
- `blocked`: useful, but blocked by missing or insufficient public Google Play API coverage.
- `not applicable`: Apple-specific behavior with no Google Play equivalent.

Sources:

- Reference command families: https://github.com/rorkai/App-Store-Connect-CLI/blob/main/docs/COMMANDS.md
- Google Play Android Developer API: https://developers.google.com/android-publisher/api-ref/rest

## Getting Started

| `asc` family | Closest `gpc` family | Google Play API coverage | Status | Notes |
| --- | --- | --- | --- | --- |
| `auth` | `auth` | OAuth/service account credentials | implemented | Scaffold supports service account profile storage and validation. |
| `doctor` | `auth doctor` | OAuth/service account credentials | implemented | Validates the configured service account JSON. |
| `install-skills` | `install-skills` | N/A | tested | Lists and installs bundled gpc agent skills with frontmatter validation, dry-run, selection, and force modes. |
| `init` | `init` | N/A | tested | `gpc init` creates `.gpc/` helper docs and a starter workflow template, with dry-run and force modes. |
| `docs` | `docs` | N/A | tested | `gpc docs parity` prints the embedded parity matrix, `gpc docs commands` generates a command reference from the Cobra command tree as JSON or markdown, and `docs/COMMANDS.md` is checked in with `make docs-check`. Release packaging is covered by GoReleaser config, Homebrew formula smoke checks, checksum-verifying install script, and README install docs. |
| `version` | `version` | N/A | implemented | Prints build metadata. |
| `completion` | `completion` | N/A | implemented | Cobra generates shell completions for supported shells. |

## Apps And Releases

| `asc` family | Closest `gpc` family | Google Play API coverage | Status | Notes |
| --- | --- | --- | --- | --- |
| `apps` | `apps` | Limited; most Play APIs require package name | blocked | Google Play has no broad app list equivalent in the Android Publisher API; `gpc apps list` returns an explicit unsupported-surface error without requiring auth. |
| `versions` | `releases` | `applications.tracks.releases`, `edits.tracks` | implemented | `gpc releases list` reads releases for a track through `edits.tracks`; upload, promote, halt, and resume cover the first mutation workflows. |
| `builds` | `artifacts` / `releases upload` | `edits.apks`, `edits.bundles`, `generatedapks` | tested | `gpc releases upload` uploads APKs or AABs through an edit and assigns the uploaded version code to the requested track. |
| `build-bundles` | `generated-apks` / `releases upload` | `edits.bundles`, `generatedapks` | tested | `gpc releases upload` uploads AABs, `gpc generated-apks list` inspects generated APK download metadata, and `gpc generated-apks download` downloads a selected generated APK by download ID with overwrite protection. |
| `release` | `releases` | `edits`, `edits.tracks`, `applications.tracks.releases` | implemented | `gpc releases upload`, `promote`, `halt`, and `resume` insert edits, update tracks, validate, and commit only with `--confirm`; promotion requires an explicit version code and defaults target status to draft. |
| `publish` | `publish` | `edits`, `edits.tracks` | implemented | `gpc publish internal` supports AAB upload planning and live validate/commit flow, appending through the raw Google track model to preserve existing release metadata. |
| `status` | `status` | `applications.tracks.releases`, `edits.tracks` | tested | `gpc status` summarizes non-draft releases by track and can include draft releases on request. |
| `submit` | `publish` / `releases` | `edits.commit`, `edits.validate` | implemented | Release upload and promotion validate by default and commit only with `--confirm`. |
| `validate` | `validate` | `edits.validate` | tested | `gpc validate` creates a temporary edit, validates it, and deletes the edit afterwards. |
| `release-notes` | `releases upload --release-note` | `edits.tracks.releases.releaseNotes` | tested | Upload accepts repeatable localized `language=text` release notes and maps them to Play track releases. |

## Tracks, Testing, And Distribution

| `asc` family | Closest `gpc` family | Google Play API coverage | Status | Notes |
| --- | --- | --- | --- | --- |
| `testflight` | `tracks` / `testers` | `edits.tracks`, `edits.testers` | tested | `gpc tracks list` lists Play tracks through a temporary edit; `gpc testers get` reads track tester Google Groups, and `gpc testers update` replaces them with dry-run/confirm edit gating. |
| `sandbox` | N/A | N/A | not applicable | Apple sandbox testers are App Store-specific. |
| `xcode` | N/A | N/A | not applicable | Local Xcode helpers do not belong in a Play Console CLI. |
| `xcode-cloud` | N/A | N/A | not applicable | Apple CI service. |
| `devices` | `device-tier-configs` / `system-apks` | `applications.deviceTierConfigs`, `systemapks.variants` | tested | `gpc device-tier-configs list` and `get` inspect app device tier configs; `gpc system-apks variants list` inspects generated system APK variants for a version code. |
| `apprecovery` | `app-recovery` | `apprecovery` | tested | `gpc app-recovery list` inspects recovery actions for a package/version code; `deploy` and `cancel` apply guarded recovery mutations only with `--confirm`, with dry-run planning. Create and targeting remain planned. |
| `internalappsharingartifacts` | `internal-sharing` | `internalappsharingartifacts` | tested | `gpc internal-sharing upload` uploads APKs or AABs to internal app sharing, with dry-run and local file preflight. |

## Metadata And Store Listing

| `asc` family | Closest `gpc` family | Google Play API coverage | Status | Notes |
| --- | --- | --- | --- | --- |
| `localizations` | `listings` | `edits.listings` | tested | `gpc listings list`, `get`, partial `update`, `delete`, and `delete-all` cover localized store listing records. |
| `metadata` | `metadata` | `edits.details`, `edits.listings` | tested | Metadata slices are covered through `gpc listings` and `gpc details`; file-based sync is still planned. |
| `screenshots` | `images` | `edits.images` | tested | `gpc images list`, `upload`, `delete`, and `delete-all` manage localized screenshots/images with dry-run/confirm edit gating for mutations. |
| `video-previews` | `videos` | Not clearly exposed in Android Publisher API | blocked | Play video preview management may require Console UI or another API surface. |
| `background-assets` | `images` | `edits.images` | tested | Feature graphics, icons, TV banners, and screenshots are covered by `gpc images list`, `upload`, `delete`, and `delete-all`. |
| `product-pages` | N/A | N/A | not applicable | App Store custom product pages do not have a direct Play equivalent. |
| `routing-coverage` | N/A | N/A | not applicable | Apple Maps routing coverage is Apple-specific. |
| `app-tags` | N/A | N/A | not applicable | Apple-generated discoverability tags are Apple-specific. |
| `categories` | `details` | `edits.details` | blocked | Current Android Publisher `edits.details` exposes default language and contact fields, not category mutation. |
| `age-rating` | `details` / `data-safety` | Partial | blocked | Some declarations are API-backed, but Play Console coverage may be incomplete. |
| `accessibility` | `details` / `data-safety` | Partial | blocked | Needs API verification before command design. |
| `encryption` | N/A | N/A | not applicable | Apple export compliance workflow. |
| `eula` | N/A | N/A | not applicable | No direct Play equivalent. |
| `data-safety` | `data-safety` | `applications.dataSafety` | tested | `gpc data-safety update` uploads a Play data safety CSV only with `--confirm`; dry-run previews are local. |

## Monetization

| `asc` family | Closest `gpc` family | Google Play API coverage | Status | Notes |
| --- | --- | --- | --- | --- |
| `iap` | `in-app-products` | `inappproducts` | tested | `gpc in-app-products list` and `get` cover legacy catalog inspection, while `patch --status` applies guarded active/inactive status changes for managed products. Legacy subscription patches are rejected; broader catalog mutations remain planned. |
| `iap` | `one-time-products` / `one-time-product-offers` | `monetization.onetimeproducts` | tested | `gpc one-time-products list` and `get` cover read-only modern one-time product inspection; `gpc one-time-product-offers list` and `get` inspect discounted and pre-order offers. Mutations remain planned. |
| `subscriptions` | `subscriptions` | `monetization.subscriptions`, `basePlans` | tested | `gpc subscriptions list` and `get` cover read-only subscriptions and embedded base plans. Mutations remain planned. |
| `subscriptions` | `subscription-offers` | `monetization.subscriptions.basePlans.offers` | tested | `gpc subscription-offers list` and `get` cover read-only offer inspection, including Google wildcard list parents. Batch-get and mutations remain planned. |
| `pricing` | `pricing` | `monetization.convertRegionPrices`, product/subscription pricing APIs | tested | `gpc pricing convert-region-prices` calculates Play regional prices from an explicit source price. Product and subscription price mutations remain planned. |
| `finance` | `finance reports summarize` | Play financial reports via downloaded CSV/GCS exports | tested | `gpc finance reports summarize` summarizes downloaded Play earnings and estimated-sales CSVs by status/type and currency amount. Programmatic GCS download remains planned. |
| `analytics` | `analytics stats summarize` | Play statistics reports via downloaded CSV/GCS exports | tested | `gpc analytics stats summarize` summarizes downloaded Play statistics CSVs by numeric metric columns, summing additive metrics and averaging rate-like metrics. Programmatic GCS download remains planned. |
| `insights` | `insights anomalies summarize` | Built from reports | tested | `gpc insights anomalies summarize` derives local summaries from `gpc vitals anomalies list` JSON output and marks paginated inputs as partial. Broader analytics and finance insights remain planned. |
| `orders` | `orders` | `orders` | tested | `gpc orders get` and `batch-get` inspect order details by ID; `refund` applies guarded refunds only with `--confirm`, with optional revoke and dry-run planning. |
| `purchases` | `purchases` | `purchases.products`, `purchases.productsv2`, `purchases.subscriptionsv2`, `voidedpurchases` | tested | `gpc purchases product` and `subscription` cover read-only purchase-token status, `gpc purchases product acknowledge` and `consume` apply guarded product purchase mutations, `gpc purchases subscription revoke` handles guarded full/prorated revocation, and `gpc purchases voided list` covers voided purchase reporting. Legacy subscription acknowledge/cancel and item-based revocation remain planned. |

## Review, Quality, And Feedback

| `asc` family | Closest `gpc` family | Google Play API coverage | Status | Notes |
| --- | --- | --- | --- | --- |
| `review` | `review` | No clear public review-submission lifecycle API | blocked | Track/release state is available, but review queue control appears limited. |
| `reviews` | `reviews` | `reviews` | tested | `gpc reviews list`, `get`, and guarded `reply` cover Play review reading and developer replies. Google limits this API to recent reviews with comments; reply text is capped at 350 characters and requires review-reply access. |
| `performance` | `vitals` | Play Developer Reporting API | tested | `gpc vitals metric-set get` inspects Android vitals metric-set metadata and freshness windows; `gpc vitals metric-set query` fetches metric rows with explicit metrics, dimensions, filters, timeline, cohort, and pagination; `gpc vitals anomalies list` lists detected metric anomalies. |
| `crashes` | `vitals` | Play Developer Reporting API | tested | `gpc vitals metric-set get --metric-set crash-rate` and `error-count` cover crash/error metric-set metadata, `gpc vitals metric-set query` can query crash/error metric rows, `gpc vitals errors issues search` searches grouped issues, and `gpc vitals errors reports search` searches individual crash/ANR/non-fatal reports. |

## Team And Access

| `asc` family | Closest `gpc` family | Google Play API coverage | Status | Notes |
| --- | --- | --- | --- | --- |
| `account` | `account` | Local auth/config inspection | tested | `gpc account status` summarizes active profile and service-account metadata without making live Google Play calls or exposing private keys. |
| `users` | `users` | `users` | tested | `gpc users list` covers developer account user inspection with pagination, `create` grants account-level access, `patch` replaces account-level fields, and `delete` removes all account access with dry-run/confirm gating. App-level access is handled through `grants`. |
| `actors` | `users` / `grants` | `users`, `grants` | tested | `gpc users list` exposes account users and embedded per-app grants; `gpc grants` manages app-level grants with dry-run/confirm gating. |
| `agreements` | N/A | N/A | blocked | Play agreement state may not be exposed through public APIs. |
| `grants` | `grants` | `grants` | tested | `gpc grants create`, `patch`, and `delete` manage app-level user grants with dry-run/confirm gating. Listing remains covered through `users list`. |

## Signing And Platform-Specific Apple Features

| `asc` family | Closest `gpc` family | Google Play API coverage | Status | Notes |
| --- | --- | --- | --- | --- |
| `signing` | N/A | N/A | not applicable | Android signing belongs to local build tooling and Play App Signing, not ASC-style cert/profile APIs. |
| `bundle-ids` | N/A | N/A | not applicable | Android package names are app identifiers but not managed like Apple bundle IDs. |
| `certificates` | N/A | N/A | not applicable | Apple-specific signing assets. |
| `profiles` | N/A | N/A | not applicable | Apple provisioning profiles. |
| `merchant-ids` | N/A | N/A | not applicable | Apple Pay merchant IDs. |
| `pass-type-ids` | N/A | N/A | not applicable | Apple Wallet pass identifiers. |
| `notarization` | N/A | N/A | not applicable | macOS notarization. |
| `app-clips` | N/A | N/A | not applicable | Apple App Clips. |
| `android-ios-mapping` | N/A | N/A | not applicable | Reference-specific bridge feature, not core Play automation. |
| `marketplace` | N/A | N/A | not applicable | Apple alternative marketplace resources. |
| `alternative-distribution` | N/A | N/A | not applicable | Apple alternative distribution resources. |
| `pre-orders` | N/A | N/A | blocked | Play pre-registration may not be exposed through Android Publisher REST. |
| `nominations` | N/A | N/A | not applicable | Apple featuring nominations. |
| `game-center` | N/A | N/A | not applicable | Apple Game Center. |

## Automation And Utility

| `asc` family | Closest `gpc` family | Google Play API coverage | Status | Notes |
| --- | --- | --- | --- | --- |
| `workflow` | `workflow` | N/A | tested | `gpc workflow list` reads `.gpc/workflow.json`; `gpc workflow run NAME` executes steps sequentially or prints the plan with `--dry-run`. |
| `webhooks` | `notifications` | Pub/Sub and Real-time developer notifications | tested | `gpc notifications rtdn decode` decodes wrapped Pub/Sub push payloads and unwrapped push payloads containing Google Play real-time developer notifications. Pub/Sub setup automation and pull variants remain planned. |
| `notify` | `notify` | N/A | tested | `gpc notify send` posts generic JSON webhook notifications with dry-run/confirm gating and redacted endpoint output. Service-specific adapters remain planned. |
| `migrate` | `migrate` | N/A | tested | `gpc migrate supply inspect` inventories fastlane supply metadata directories by locale, listing files, changelogs, image sets, and unknown files as stable JSON. Conversion/import remains planned. |
| `diff` | `diff` | N/A | tested | `gpc diff json FROM TO` compares local JSON payloads with deterministic JSON Pointer paths and optional `--fail-on-change` CI behavior. |
| `capabilities` | `capabilities` | N/A | tested | `gpc capabilities` exposes this parity matrix as structured CLI output with status and section filters. |
| `search` | `search` | N/A | tested | `gpc search QUERY` searches command paths, descriptions, and flag names for agent-oriented command discovery. |
| `schema` | `schema` | Discovery document | tested | `gpc schema` fetches the Android Publisher discovery document and emits a stable resource/method summary, with resource and method filters. |
| `snitch` | `snitch` | N/A | tested | `gpc snitch report` generates a deterministic GitHub issue URL for CLI friction without network or auth side effects. |
| `web` | `web status` | N/A | blocked | Boundary command is tested; Play Console browser automation remains blocked from the Go CLI until there is a stable, testable automation contract. Console-only workflows should use explicit operator-driven browser automation outside the CLI. |

## First Vertical Slice

The first API-backed slice should be:

```sh
gpc publish internal \
  --package com.example.app \
  --aab ./app-release.aab \
  --release-name "1.2.3" \
  --dry-run
```

This should exercise:

- service account auth
- edit insert
- bundle upload
- track update
- edit validate
- optional edit commit behind `--confirm`
- stable JSON output for CI
