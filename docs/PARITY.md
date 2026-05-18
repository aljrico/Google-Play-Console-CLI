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
| `init` | `init` | N/A | planned | Should create `.gpc/` helper docs and workflow templates. |
| `docs` | `docs` | N/A | planned | Should expose embedded command and workflow docs. |
| `version` | `version` | N/A | implemented | Prints build metadata. |
| `completion` | `completion` | N/A | implemented | Cobra generates shell completions for supported shells. |

## Apps And Releases

| `asc` family | Closest `gpc` family | Google Play API coverage | Status | Notes |
| --- | --- | --- | --- | --- |
| `apps` | `apps` | Limited; most Play APIs require package name | planned | Google Play has no broad app list equivalent in the Android Publisher API. |
| `versions` | `releases` | `applications.tracks.releases`, `edits.tracks` | implemented | `gpc releases list` reads releases for a track through `edits.tracks`; upload, promote, halt, and resume cover the first mutation workflows. |
| `builds` | `artifacts` / `releases upload` | `edits.apks`, `edits.bundles`, `generatedapks` | implemented | `gpc releases upload` uploads AABs through an edit; APK support is still planned. |
| `build-bundles` | `bundles` | `edits.bundles`, `generatedapks` | planned | Android App Bundles are first-class in Play. |
| `release` | `release` | `edits`, `edits.tracks`, `applications.tracks.releases` | implemented | `gpc releases upload`, `promote`, `halt`, and `resume` insert edits, update tracks, validate, and commit only with `--confirm`; promotion requires an explicit version code and defaults target status to draft. |
| `publish` | `publish` | `edits`, `edits.tracks` | implemented | `gpc publish internal` supports AAB upload planning and live validate/commit flow, appending through the raw Google track model to preserve existing release metadata. |
| `status` | `status` | `applications.tracks.releases`, `edits.tracks` | planned | Should summarize active releases by track. |
| `submit` | `publish` / `release` | `edits.commit`, `edits.validate` | implemented | Release upload and promotion validate by default and commit only with `--confirm`. |
| `validate` | `validate` | `edits.validate` | planned | Should dry-run an edit before commit. |
| `release-notes` | `release-notes` | `edits.tracks.releases.releaseNotes` | planned | Release notes are attached to track releases. |

## Tracks, Testing, And Distribution

| `asc` family | Closest `gpc` family | Google Play API coverage | Status | Notes |
| --- | --- | --- | --- | --- |
| `testflight` | `tracks` / `testers` | `edits.tracks`, `edits.testers` | implemented | `gpc tracks list` lists Play tracks through a temporary edit; tester management is still planned. |
| `sandbox` | N/A | N/A | not applicable | Apple sandbox testers are App Store-specific. |
| `xcode` | N/A | N/A | not applicable | Local Xcode helpers do not belong in a Play Console CLI. |
| `xcode-cloud` | N/A | N/A | not applicable | Apple CI service. |
| `devices` | `device-tier-configs` / `system-apks` | `applications.deviceTierConfigs`, `systemapks.variants` | planned | Android-specific device targeting deserves its own command family. |
| `apprecovery` | `app-recovery` | `apprecovery` | planned | Google Play has direct app recovery APIs. |
| `internalappsharingartifacts` | `internal-sharing` | `internalappsharingartifacts` | planned | Useful for quick APK/AAB sharing outside normal tracks. |

## Metadata And Store Listing

| `asc` family | Closest `gpc` family | Google Play API coverage | Status | Notes |
| --- | --- | --- | --- | --- |
| `localizations` | `listings` | `edits.listings` | tested | `gpc listings list`, `get`, partial `update`, `delete`, and `delete-all` cover localized store listing records. |
| `metadata` | `metadata` | `edits.details`, `edits.listings` | tested | Metadata slices are covered through `gpc listings` and `gpc details`; file-based sync is still planned. |
| `screenshots` | `images` | `edits.images` | planned | Screenshots and graphic assets are image uploads by language and type. |
| `video-previews` | `videos` | Not clearly exposed in Android Publisher API | blocked | Play video preview management may require Console UI or another API surface. |
| `background-assets` | `images` | `edits.images` | planned | Some asset types map to Play listing images. |
| `product-pages` | N/A | N/A | not applicable | App Store custom product pages do not have a direct Play equivalent. |
| `routing-coverage` | N/A | N/A | not applicable | Apple Maps routing coverage is Apple-specific. |
| `app-tags` | N/A | N/A | not applicable | Apple-generated discoverability tags are Apple-specific. |
| `categories` | `details` | `edits.details` | blocked | Current Android Publisher `edits.details` exposes default language and contact fields, not category mutation. |
| `age-rating` | `details` / `data-safety` | Partial | blocked | Some declarations are API-backed, but Play Console coverage may be incomplete. |
| `accessibility` | `details` / `data-safety` | Partial | blocked | Needs API verification before command design. |
| `encryption` | N/A | N/A | not applicable | Apple export compliance workflow. |
| `eula` | N/A | N/A | not applicable | No direct Play equivalent. |
| `data-safety` | `data-safety` | `applications.dataSafety` | planned | Google-specific declaration workflow. |

## Monetization

| `asc` family | Closest `gpc` family | Google Play API coverage | Status | Notes |
| --- | --- | --- | --- | --- |
| `iap` | `in-app-products` | `inappproducts` | tested | `gpc in-app-products list` and `get` cover read-only legacy catalog inspection. Mutations remain planned. |
| `iap` | `one-time-products` | `monetization.onetimeproducts` | planned | Modern one-time product resources have separate endpoints and pagination semantics. |
| `subscriptions` | `subscriptions` | `monetization.subscriptions`, `basePlans` | tested | `gpc subscriptions list` and `get` cover read-only subscriptions and embedded base plans. Mutations remain planned. |
| `subscriptions` | `subscription-offers` | `monetization.subscriptions.basePlans.offers` | planned | Offers have a separate API surface and are not covered by `gpc subscriptions` yet. |
| `pricing` | `pricing` | `monetization.convertRegionPrices`, product/subscription pricing APIs | planned | App pricing and IAP pricing may need separate commands. |
| `finance` | `reports` | Separate reporting APIs, not Android Publisher REST | planned | Needs separate API/client research. |
| `analytics` | `reports` | Separate reporting APIs, not Android Publisher REST | planned | Needs separate API/client research. |
| `insights` | `insights` | Built from reports | planned | Derived command, not a direct API mapping. |
| `orders` | `orders` | `orders` | planned | Useful for order lookup and refunds. |
| `purchases` | `purchases` | `purchases.products`, `purchases.subscriptions`, `voidedpurchases` | planned | Useful for server-side entitlement support. |

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
| `users` | `users` | `users` | planned | Play supports user management. |
| `actors` | `users` / `grants` | `users`, `grants` | planned | API keys and users differ from App Store Connect actors. |
| `agreements` | N/A | N/A | blocked | Play agreement state may not be exposed through public APIs. |
| `grants` | `grants` | `grants` | planned | Google-specific access grants. |

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
| `capabilities` | `capabilities` | N/A | planned | Should expose this matrix from the CLI later. |
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
