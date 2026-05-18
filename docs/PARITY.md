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
| `install-skills` | `skills` | N/A | planned | Useful later for agent workflows, not core CLI behavior. |
| `init` | `init` | N/A | tested | `gpc init` creates `.gpc/` helper docs and a starter workflow template, with dry-run and force modes. |
| `docs` | `docs` | N/A | tested | `gpc docs parity` prints the embedded parity matrix as JSON-wrapped content or raw markdown. |
| `version` | `version` | N/A | implemented | Prints build metadata. |
| `completion` | `completion` | N/A | implemented | Cobra generates shell completions for supported shells. |

## Apps And Releases

| `asc` family | Closest `gpc` family | Google Play API coverage | Status | Notes |
| --- | --- | --- | --- | --- |
| `apps` | `apps` | Limited; most Play APIs require package name | blocked | Google Play has no broad app list equivalent in the Android Publisher API; `gpc apps list` returns an explicit unsupported-surface error without requiring auth. |
| `versions` | `releases` | `applications.tracks.releases`, `edits.tracks` | implemented | `gpc releases list` reads releases for a track through `edits.tracks`; upload, promote, halt, and resume cover the first mutation workflows. |
| `builds` | `artifacts` / `releases upload` | `edits.apks`, `edits.bundles`, `generatedapks` | tested | `gpc releases upload` uploads APKs or AABs through an edit and assigns the uploaded version code to the requested track. |
| `build-bundles` | `generated-apks` / `releases upload` | `edits.bundles`, `generatedapks` | tested | `gpc releases upload` uploads AABs and `gpc generated-apks list` inspects generated APK download metadata for a version code. Download support remains planned. |
| `release` | `releases` | `edits`, `edits.tracks`, `applications.tracks.releases` | implemented | `gpc releases upload`, `promote`, `halt`, and `resume` insert edits, update tracks, validate, and commit only with `--confirm`; promotion requires an explicit version code and defaults target status to draft. |
| `publish` | `publish` | `edits`, `edits.tracks` | implemented | `gpc publish internal` supports AAB upload planning and live validate/commit flow, appending through the raw Google track model to preserve existing release metadata. |
| `status` | `status` | `applications.tracks.releases`, `edits.tracks` | tested | `gpc status` summarizes non-draft releases by track and can include draft releases on request. |
| `submit` | `publish` / `releases` | `edits.commit`, `edits.validate` | implemented | Release upload and promotion validate by default and commit only with `--confirm`. |
| `validate` | `validate` | `edits.validate` | tested | `gpc validate` creates a temporary edit, validates it, and deletes the edit afterwards. |
| `release-notes` | `releases upload --release-note` | `edits.tracks.releases.releaseNotes` | tested | Upload accepts repeatable localized `language=text` release notes and maps them to Play track releases. |

## Tracks, Testing, And Distribution

| `asc` family | Closest `gpc` family | Google Play API coverage | Status | Notes |
| --- | --- | --- | --- | --- |
| `testflight` | `tracks` / `testers` | `edits.tracks`, `edits.testers` | implemented | `gpc tracks list` lists Play tracks through a temporary edit; tester management is still planned. |
| `sandbox` | N/A | N/A | not applicable | Apple sandbox testers are App Store-specific. |
| `xcode` | N/A | N/A | not applicable | Local Xcode helpers do not belong in a Play Console CLI. |
| `xcode-cloud` | N/A | N/A | not applicable | Apple CI service. |
| `devices` | `device-tier-configs` / `system-apks` | `applications.deviceTierConfigs`, `systemapks.variants` | tested | `gpc device-tier-configs list` and `get` inspect app device tier configs; `gpc system-apks variants list` inspects generated system APK variants for a version code. |
| `apprecovery` | `app-recovery` | `apprecovery` | tested | `gpc app-recovery list` inspects recovery actions for a package/version code. Create, deploy, cancel, and targeting remain planned. |
| `internalappsharingartifacts` | `internal-sharing` | `internalappsharingartifacts` | tested | `gpc internal-sharing upload` uploads APKs or AABs to internal app sharing, with dry-run and local file preflight. |

## Metadata And Store Listing

| `asc` family | Closest `gpc` family | Google Play API coverage | Status | Notes |
| --- | --- | --- | --- | --- |
| `localizations` | `listings` | `edits.listings` | tested | `gpc listings list`, `get`, partial `update`, `delete`, and `delete-all` cover localized store listing records. |
| `metadata` | `metadata` | `edits.details`, `edits.listings` | tested | Metadata slices are covered through `gpc listings` and `gpc details`; file-based sync is still planned. |
| `screenshots` | `images` | `edits.images` | tested | `gpc images list` inspects localized screenshot/image metadata by language and image type. Upload and delete remain planned. |
| `video-previews` | `videos` | Not clearly exposed in Android Publisher API | blocked | Play video preview management may require Console UI or another API surface. |
| `background-assets` | `images` | `edits.images` | tested | Feature graphics, icons, TV banners, and screenshots are covered by `gpc images list`; uploads remain planned. |
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
| `iap` | `in-app-products` | `inappproducts` | tested | `gpc in-app-products list` and `get` cover read-only legacy catalog inspection. Mutations remain planned. |
| `iap` | `one-time-products` | `monetization.onetimeproducts` | planned | Modern one-time product resources have separate endpoints and pagination semantics. |
| `subscriptions` | `subscriptions` | `monetization.subscriptions`, `basePlans` | tested | `gpc subscriptions list` and `get` cover read-only subscriptions and embedded base plans. Mutations remain planned. |
| `subscriptions` | `subscription-offers` | `monetization.subscriptions.basePlans.offers` | tested | `gpc subscription-offers list` and `get` cover read-only offer inspection, including Google wildcard list parents. Batch-get and mutations remain planned. |
| `pricing` | `pricing` | `monetization.convertRegionPrices`, product/subscription pricing APIs | tested | `gpc pricing convert-region-prices` calculates Play regional prices from an explicit source price. Product and subscription price mutations remain planned. |
| `finance` | `reports` | Separate reporting APIs, not Android Publisher REST | planned | Needs separate API/client research. |
| `analytics` | `reports` | Separate reporting APIs, not Android Publisher REST | planned | Needs separate API/client research. |
| `insights` | `insights` | Built from reports | planned | Derived command, not a direct API mapping. |
| `orders` | `orders` | `orders` | tested | `gpc orders get` and `batch-get` inspect order details by ID. Refunds remain planned and should require explicit confirmation. |
| `purchases` | `purchases` | `purchases.productsv2`, `purchases.subscriptionsv2`, `voidedpurchases` | tested | `gpc purchases product` and `subscription` cover read-only purchase-token status, and `gpc purchases voided list` covers voided purchase reporting. Acknowledge, consume, cancel, and refund remain planned. |

## Review, Quality, And Feedback

| `asc` family | Closest `gpc` family | Google Play API coverage | Status | Notes |
| --- | --- | --- | --- | --- |
| `review` | `review` | No clear public review-submission lifecycle API | blocked | Track/release state is available, but review queue control appears limited. |
| `reviews` | `reviews` | `reviews` | tested | `gpc reviews list`, `get`, and guarded `reply` cover Play review reading and developer replies. Google limits this API to recent reviews with comments; reply text is capped at 350 characters and requires review-reply access. |
| `performance` | `vitals` / `quality` | Separate reporting APIs | planned | Android vitals is outside the core Android Publisher REST surface. |
| `crashes` | `vitals` / `quality` | Separate reporting APIs | planned | Needs separate API/client research. |

## Team And Access

| `asc` family | Closest `gpc` family | Google Play API coverage | Status | Notes |
| --- | --- | --- | --- | --- |
| `account` | `account` | Partial | planned | Account health is likely a mix of API-backed checks and local auth checks. |
| `users` | `users` | `users` | tested | `gpc users list` covers read-only developer account user inspection with pagination. Mutations remain planned. |
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
| `workflow` | `workflow` | N/A | planned | Repo-local automation runner, independent of Google APIs. |
| `webhooks` | `notifications` | Pub/Sub and Real-time developer notifications | planned | Likely separate setup from Android Publisher REST. |
| `notify` | `notify` | N/A | planned | External notifications are CLI utility behavior. |
| `migrate` | `migrate` | N/A | planned | Should support fastlane supply metadata migration. |
| `diff` | `diff` | N/A | planned | Deterministic non-mutating plans are important for CI. |
| `capabilities` | `capabilities` | N/A | tested | `gpc capabilities` exposes this parity matrix as structured CLI output with status and section filters. |
| `schema` | `schema` | Discovery document | planned | Google APIs expose a machine-readable discovery document. |
| `snitch` | `snitch` | N/A | planned | Nice-to-have friction reporter. |
| `web` | `web` | N/A | planned | Experimental Play Console browser workflows only when no API exists. |

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
