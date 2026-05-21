---
name: gpc-iap-setup
description: Build and maintain Google Play in-app product, one-time product, subscription, and subscription offer catalogs with gpc.
---

# gpc IAP Setup

Use this skill for Google Play monetization catalog workflows with `gpc`: managed products, the new one-time products and offers APIs, and subscriptions with base plans and offers.

## First Checks

```sh
gpc auth doctor --output json
gpc capabilities --section "Monetization" --output json
```

Google Play exposes two monetization APIs in parallel: the legacy `inappproducts` API (covered by `gpc in-app-products`) for managed products only, and the newer monetization API (covered by `gpc one-time-products`, `gpc subscriptions`, and `gpc subscription-offers`). Use the legacy commands only for managed products; everything else belongs on the new API.

## Inspect Before Mutating

```sh
gpc in-app-products list --package com.example.app --output json
gpc one-time-products list --package com.example.app --output json
gpc subscriptions list --package com.example.app --output json
```

`batch-get` accepts repeatable IDs:

```sh
gpc one-time-products batch-get --package com.example.app --product-id coins_100 --product-id coins_500 --output json
gpc subscriptions batch-get --package com.example.app --product-id premium_monthly --product-id premium_yearly --output json
```

## Managed Products (Legacy)

```sh
gpc in-app-products create \
  --package com.example.app \
  --sku coins_100 \
  --default-language en-US \
  --default-price USD:1990000 \
  --title "100 coins" \
  --description "A small coin pack." \
  --dry-run \
  --output json

gpc in-app-products patch \
  --package com.example.app \
  --sku coins_100 \
  --regional-price US:USD:2990000 \
  --regional-price BR:BRL:9990000 \
  --dry-run \
  --output json
```

`create` builds managed products only; live patches and deletes reject legacy subscription SKUs. Compliance flags on `patch` cover EEA withdrawal right type, tokenized digital asset declaration, regional reduced-tax tiers, and US streaming tax type.

## One-Time Products (New API)

```sh
gpc one-time-products create \
  --package com.example.app \
  --product-id coins_100 \
  --listing 'en-US,100 coins,Buy coins.' \
  --price US:USD:1:990000000 \
  --offer-tag public \
  --regions-version 2026/05 \
  --dry-run \
  --output json
```

`create` uses Google's patch-with-`allowMissing` surface — that is the actual create endpoint. For richer products (rent options, new-regions pricing, compliance), prefer `--from-json` with a typed `OneTimeProduct` body. Purchase option mutations are separate commands:

```sh
gpc one-time-products purchase-option batch-patch-prices --package com.example.app --price coins_100/buy/US:USD:3:490000000 --regions-version 2026/05 --dry-run --output json
gpc one-time-products purchase-option deactivate --package com.example.app --product-id coins_100 --purchase-option-id buy --dry-run --output json
```

One-time product offers (discounts, pre-orders) live under `gpc one-time-product-offers` with `create` / `batch-patch-availability` / `batch-patch-relative-discounts` / `batch-patch-absolute-discounts`.

## Subscriptions

```sh
gpc subscriptions create \
  --package com.example.app \
  --product-id premium_monthly \
  --listing 'en-US,Premium,Full access' \
  --base-plan-id monthly \
  --billing-period P1M \
  --price US:USD:4:990000000 \
  --regions-version 2026/05 \
  --dry-run \
  --output json
```

Basic flags build one auto-renewing base plan; for prepaid use `--prepaid --time-extension TIME_EXTENSION_ACTIVE`; for installments use `--installments --committed-payments N --renewal-type RENEWAL_TYPE_...`. Compliance flags: `--eea-withdrawal-right-type`, `--tokenized-digital-asset`, repeatable `--regional-tax-tier`, repeatable `--regional-streaming-tax`. Use `--from-json` for richer Subscription bodies.

Base plan state and pricing are separate commands:

```sh
gpc subscriptions base-plan batch-activate --package com.example.app --base-plan premium_monthly/monthly --dry-run --output json
gpc subscriptions base-plan batch-patch-prices --package com.example.app --regions-version 2026/05 --price premium_monthly/monthly/US:USD:4:990000000 --dry-run --output json
```

## Subscription Offers

```sh
gpc subscription-offers create \
  --package com.example.app \
  --product-id premium_monthly \
  --base-plan-id monthly \
  --offer-id intro \
  --free-region US \
  --phase-duration P7D \
  --regions-version 2026/05 \
  --dry-run \
  --output json
```

Basic flags build one phase across explicit regions; add `--phase-2-*` flags for a two-phase offer. Targeting flags cover acquisition (`--targeting-acquisition-scope`) and upgrade (`--targeting-upgrade-scope` plus subscription/billing/once-per-user options). Google Play only supports offers under auto-renewing base plans.

Phase patches are separate:

```sh
gpc subscription-offers batch-patch-phase-prices --package com.example.app --price premium_monthly/monthly/intro/0/US:USD:1:990000000 --regions-version 2026/05 --dry-run --output json
```

`0` is the phase index. Phase patches preserve untouched phases and regions; `batch-patch-phase-free` clears existing regional pricing for the targeted phase.

## Guardrails

- Always run with `--dry-run` first; add `--confirm` only with explicit user intent.
- The legacy `in-app-products` API and the new monetization API are distinct. Never mix calls — managed products use `gpc in-app-products`, one-time products and subscriptions use the new commands.
- Product IDs, base plan IDs, and offer IDs are immutable. Renaming requires delete + recreate.
- `regions-version` must match a published Play regions version (e.g. `2026/05`). It is not a free-form string.
- Subscription and one-time product `create --from-json` accept both the Google API shape and `gpc <resource> get --output json` shape; immutable IDs on the command line override any matching field in the body.
- Pricing for base plans is set on the Play side; do not project pricing from RevenueCat or other downstream catalogs.
